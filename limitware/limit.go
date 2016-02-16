/*
Package limitware provides middleware for limiting the number of requests a
single client can have open at one time. It implements the httpware.Middleware
interface for easy composition with other middleware.
*/
package limitware

import (
	"net/http"
	"sync"

	"github.com/nstogner/httpware"
	"github.com/nstogner/httpware/httperr"
	"golang.org/x/net/context"
)

type Rate struct {
	limit int

	addrsMutex sync.Mutex
	addrs      map[string]int
}

// New creates a new Rate httpware.Middleware instance. The limit is the max
// number of requests that a single remote address can have open. It applies
// to any handlers downstream from this middleware.
func NewRate(limit int) Rate {
	return Rate{
		limit: limit,
		addrs: make(map[string]int),
	}
}

func (w Rate) Contains() []string { return []string{"limitware.Rate"} }
func (w Rate) Requires() []string { return []string{"errorware.Ware"} }

func (ware Rate) Handle(next httpware.Handler) httpware.Handler {
	return httpware.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		defer ware.decrement(r.RemoteAddr)
		if !ware.increment(r.RemoteAddr) {
			return httperr.New("exceeded request rate limit", 429)
		}
		return next.ServeHTTPContext(ctx, w, r)
	})
}

func (ware *Rate) increment(addr string) bool {
	ware.addrsMutex.Lock()
	defer ware.addrsMutex.Unlock()
	return (ware.addrs[addr] >= ware.limit)
}
func (ware *Rate) decrement(addr string) {
	ware.addrsMutex.Lock()
	defer ware.addrsMutex.Unlock()
	if ware.addrs[addr] <= 1 {
		delete(ware.addrs, addr)
	} else {
		ware.addrs[addr]--
	}
}
