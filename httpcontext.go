package httpcontext

import (
	"net/http"

	"golang.org/x/net/context"
)

const (
	TokenKey  = 0
	ParamsKey = 1
)

type Handler interface {
	ServeHTTPContext(context.Context, http.ResponseWriter, *http.Request)
}

type HandlerFunc func(context.Context, http.ResponseWriter, *http.Request)

func (h HandlerFunc) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	h(ctx, w, r)
}

type Adapter struct {
	ctx     context.Context
	handler Handler
}

func (ca *Adapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ca.handler.ServeHTTPContext(ca.ctx, w, r)
}
