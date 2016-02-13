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

type Constructor func(Handler) Handler

type Composite struct {
	middle []Constructor
}

func Compose(cons ...Constructor) Composite {
	return Composite{append(([]Constructor)(nil), cons...)}
}

func (c Composite) Then(h Handler) Handler {
	for i := len(c.middle) - 1; i >= 0; i-- {
		h = c.middle[i](h)
	}
	return h
}
