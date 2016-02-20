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
	Defaults = Config{
		RemoteLimit: 100,
		TotalLimit:  1000000,
		RetryAfter:  3600,
	}
)

type Config struct {
	// The number of active requests a single remote address can have
	RemoteLimit int
	// The limit of total active requests
	TotalLimit uint64
	// Sets the header 'Retry-After'. Units are in seconds.
	RetryAfter int
}

type Middle struct {
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
func New(conf Config) Middle {
	middle := Middle{
		remoteLimit: conf.RemoteLimit,
		addrs:       make(map[string]int),
		totalLimit:  conf.TotalLimit,
	}
	if conf.RetryAfter == 0 {
		headerValue := strconv.Itoa(conf.RetryAfter)
		middle.retryHeader = func(w http.ResponseWriter) { w.Header().Set("Retry-After", headerValue) }
	} else {
		middle.retryHeader = func(w http.ResponseWriter) {}
	}
	return middle
}

func (m Middle) Contains() []string { return []string{"limitware", "limitware/requests"} }
func (m Middle) Requires() []string { return []string{"errorware"} }

func (m Middle) Handle(next httpware.Handler) httpware.Handler {
	return httpware.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		remote := strings.Split(r.RemoteAddr, ":")
		if len(remote) != 2 {
			return next.ServeHTTPContext(ctx, w, r)
		}

		if m.increment(remote[0]) {
			defer m.decrement(remote[0])
			return next.ServeHTTPContext(ctx, w, r)
		}

		// Send a 429 response (Too Many Requests).
		m.retryHeader(w)
		return httperr.New("exceeded request rate limit", 429)
	})
}

// TotalRate gets the total number of active requests.
func (m *Middle) TotalRate() uint64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.total
}

func (m *Middle) increment(addr string) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.addrs[addr] < m.remoteLimit && m.total < m.totalLimit {
		m.addrs[addr]++
		m.total++
		return true
	}
	return false
}
func (m *Middle) decrement(addr string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.addrs[addr] <= 1 {
		delete(m.addrs, addr)
	} else {
		m.addrs[addr]--
	}
	m.total--
}
