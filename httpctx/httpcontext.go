package httpctx

import (
	"net/http"

	"golang.org/x/net/context"
)

// Registry of context keys
const (
	TokenKey               = 0
	ParamsKey              = 1
	JsonEntityKey          = 2
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

func Adapt(h Handler) Adapter {
	return Adapter{
		ctx:     context.Background(),
		handler: h,
	}
}

type Adapter struct {
	ctx     context.Context
	handler Handler
}

func (ca Adapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ca.handler.ServeHTTPContext(ca.ctx, w, r)
}
