/*
Package limitware provides middleware for limiting the number of requests a
single client can have open at one time. It implements the httpware.Middleware
interface for easy composition with other middleware.
*/
package limitware

import (
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/nstogner/httpware"
	"github.com/nstogner/httpware/httperr"
	"golang.org/x/net/context"
)

var (
	Defaults = Def{
		RemoteLimit: 100,
		TotalLimit:  1000000,
		RetryAfter:  3600,
	}
)

type Def struct {
	// The number of active requests a single remote address can have
	RemoteLimit int
	// The limit of total active requests
	TotalLimit uint64
	// Sets the header 'Retry-After'. Units are in seconds.
	RetryAfter int
}

type ReqLimit struct {
	remoteLimit int
	totalLimit  uint64

	retryHeader func(w http.ResponseWriter)

	mutex sync.Mutex
	total uint64
	addrs map[string]int
}

// New creates a new Rate httpware.Middleware instance. The limit is the max
// number of requests that a single remote address can have open. It applies
// to any handlers downstream from this middleware.
func NewReqLimit(def Def) ReqLimit {
	rl := ReqLimit{
		remoteLimit: def.RemoteLimit,
		addrs:       make(map[string]int),
		totalLimit:  def.TotalLimit,
	}
	if def.RetryAfter == 0 {
		headerValue := strconv.Itoa(def.RetryAfter)
		rl.retryHeader = func(w http.ResponseWriter) { w.Header().Set("Retry-After", headerValue) }
	} else {
		rl.retryHeader = func(w http.ResponseWriter) {}
	}
	return rl
}

func (w ReqLimit) Contains() []string { return []string{"limitware.Rate"} }
func (w ReqLimit) Requires() []string { return []string{"errorware.Ware"} }

func (rl ReqLimit) Handle(next httpware.Handler) httpware.Handler {
	return httpware.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		remote := strings.Split(r.RemoteAddr, ":")
		if len(remote) != 2 {
			return next.ServeHTTPContext(ctx, w, r)
		}

		if rl.increment(remote[0]) {
			defer rl.decrement(remote[0])
			return next.ServeHTTPContext(ctx, w, r)
		}

		// Send a 429 response (Too Many Requests).
		rl.retryHeader(w)
		return httperr.New("exceeded request rate limit", 429)
	})
}

// TotalRate gets the total number of active requests.
func (rl *ReqLimit) TotalRate() uint64 {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	return rl.total
}

func (rl *ReqLimit) increment(addr string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	if rl.addrs[addr] < rl.remoteLimit && rl.total < rl.totalLimit {
		rl.addrs[addr]++
		rl.total++
		return true
	}
	return false
}
func (rl *ReqLimit) decrement(addr string) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	if rl.addrs[addr] <= 1 {
		delete(rl.addrs, addr)
	} else {
		rl.addrs[addr]--
	}
	rl.total--
}
