/*
Package httpware defines a http handler interface that allows for data-sharing
through the use of net/context.
*/
package httpware

import (
	"context"
	"net/http"
)

// The Handler interface is intended to be an improvement on the http.Handler
// interface. It uses the net/context package to enable the sharing of data
// betweeen middleware. It also returns an error value to reduce the risk of
// sneaky bugs that can be caused by a call to http.Error, while forgeting to
// return early in a standard http.Handler function.
type Handler interface {
	ServeHTTPCtx(context.Context, http.ResponseWriter, *http.Request) error
}

// HandlerFunc is an adapter to allow the use of ordinary functions as
// handers.
type HandlerFunc func(context.Context, http.ResponseWriter, *http.Request) error

// ServeHTTPCtx calls h(ctx, w, r).
func (h HandlerFunc) ServeHTTPCtx(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return h(ctx, w, r)
}
