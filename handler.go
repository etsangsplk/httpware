package httpware

import (
	"net/http"

	"golang.org/x/net/context"
)

// The Handler interface is intended to be an improvement on the http.Handler
// interface. It uses the net/context package to enable the sharing of data
// betweeen middleware. It also returns an error value to reduce the risk of
// sneaky bugs that can be caused by a call to http.Error, while forgeting to
// return early in a standard http.Handler function.
type Handler interface {
	ServeHTTPContext(context.Context, http.ResponseWriter, *http.Request) error
}

// HandlerFunc is an adapter to allow the use of ordinary functions as
// handers.
type HandlerFunc func(context.Context, http.ResponseWriter, *http.Request) error

// ServerHTTPContext calls h(ctx, w, r).
func (h HandlerFunc) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return h(ctx, w, r)
}
