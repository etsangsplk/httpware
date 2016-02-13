package httpctx

import (
	"net/http"

	"golang.org/x/net/context"
)

// Registry of context keys
const (
	TokenKey               = 0
	ParamsKey              = 1
	EntityKey              = 2
	RequestContentTypeKey  = 3
	ResponseContentTypeKey = 4
)

type Handler interface {
	ServeHTTPContext(context.Context, http.ResponseWriter, *http.Request) error
}

type HandlerFunc func(context.Context, http.ResponseWriter, *http.Request) error

func (h HandlerFunc) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return h(ctx, w, r)
}
