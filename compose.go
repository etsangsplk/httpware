package ctxware

import (
	"fmt"
	"net/http"

	"golang.org/x/net/context"
)

type Middleware interface {
	Name() string
	Dependencies() []string
	Handle(Handler) Handler
}

type Composite struct {
	middle []Middleware
}

func MustCompose(mdlw ...Middleware) Composite {
	names := make(map[string]bool, 0)
	for _, m := range mdlw {
		for _, dep := range m.Dependencies() {
			if !names[dep] {
				panic(fmt.Errorf("missing dependency '%s' required by middleware '%s'", dep, m.Name()))
			}
		}
		names[m.Name()] = true
	}
	return Composite{append(([]Middleware)(nil), mdlw...)}
}

type CompositeHandler struct {
	h Handler
}

func (ch CompositeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ch.h.ServeHTTPContext(context.Background(), w, r)
}

func (ch CompositeHandler) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return ch.h.ServeHTTPContext(ctx, w, r)
}

func (c Composite) Then(h Handler) CompositeHandler {
	for i := len(c.middle) - 1; i >= 0; i-- {
		h = c.middle[i].Handle(h)
	}
	return CompositeHandler{
		h: h,
	}
}
