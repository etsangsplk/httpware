/*
Package limitware provides middleware for limiting the number of requests a
single client can have open at one time. It implements the httpware.Middleware
interface for easy composition with other middleware.
*/
package limitware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/nstogner/httpware"
	"golang.org/x/net/context"
)

var (
	// Defaults is a reasonable configuration.
	Defaults = Config{
		RemoteLimit: 100,
		TotalLimit:  1000000,
		RetryAfter:  3600,
	}
)

// Config is used to initialize a new instance of Middle.
type Config struct {
	// The number of active requests a single remote address can have
	RemoteLimit int
	// The limit of total active requests
	TotalLimit uint64
	// Sets the header 'Retry-After'. Units are in seconds.
	RetryAfter int
}

// Middle is middleware that limits http requests.
type Middle struct {
	remoteLimit int
	totalLimit  uint64

	retryHeader func(w http.ResponseWriter)

	mutex sync.Mutex
	total uint64
	addrs map[string]int
}

// New creates a new limitware.Middle instance. It can limit the requests per
// remote and the total requests.
func New(conf Config) *Middle {
	middle := Middle{
		remoteLimit: conf.RemoteLimit,
		totalLimit:  conf.TotalLimit,
		total:       0,
		addrs:       make(map[string]int),
	}
	if conf.RetryAfter == 0 {
		headerValue := strconv.Itoa(conf.RetryAfter)
		middle.retryHeader = func(w http.ResponseWriter) { w.Header().Set("Retry-After", headerValue) }
	} else {
		middle.retryHeader = func(w http.ResponseWriter) {}
	}
	return &middle
}

// Handle takes the next handler as an argument and wraps it in this middleware.
func (m *Middle) Handle(next httpware.Handler) httpware.Handler {
	return httpware.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		remote := strings.Split(r.RemoteAddr, ":")
		if len(remote) != 2 {
			return next.ServeHTTPCtx(ctx, w, r)
		}

		if m.increment(remote[0]) {
			defer m.decrement(remote[0])
			return next.ServeHTTPCtx(ctx, w, r)
		}

		// Send a 429 response (Too Many Requests).
		m.retryHeader(w)
		return httpware.NewErr("exceeded request rate limit", 429)
	})
}

// TotalRate gets the total number of active requests.
func (m *Middle) TotalRate() uint64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	fmt.Println("totalRate", m.total)
	return m.total
}

// RemoteRate gets the number of active requests for a given remote.
func (m *Middle) RemoteRate(addr string) int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.addrs[addr]
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
