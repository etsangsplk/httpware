package ctxware

import (
	"net/http"

	"golang.org/x/net/context"
)

type Handler interface {
	ServeHTTPContext(context.Context, http.ResponseWriter, *http.Request) error
}

type HandlerFunc func(context.Context, http.ResponseWriter, *http.Request) error

func (h HandlerFunc) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return h(ctx, w, r)
}
