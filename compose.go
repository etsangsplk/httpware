package ctxware

import (
	"fmt"
	"net/http"

	"golang.org/x/net/context"
)

type Middleware interface {
	Contains() []string
	Requires() []string
	Handle(Handler) Handler
}

type Composite struct {
	middle   []Middleware
	contains []string
}

func MustCompose(mdlw ...Middleware) Composite {
	contains := make(map[string]bool)
	for _, m := range mdlw {
		for _, is := range m.Contains() {
			contains[is] = true
		}
		for _, dep := range m.Requires() {
			if !contains[dep] {
				panic(fmt.Errorf("missing dependency '%s'", dep))
			}
		}
	}
	containsSlice := make([]string, 0)
	for c, _ := range contains {
		containsSlice = append(containsSlice, c)
	}
	return Composite{
		middle:   append(([]Middleware)(nil), mdlw...),
		contains: containsSlice,
	}
}

func (c Composite) Contains() []string { return c.contains }
func (c Composite) Requires() []string { return []string{} }
func (c Composite) Handle(h Handler) Handler {
	for i := len(c.middle) - 1; i >= 0; i-- {
		h = c.middle[i].Handle(h)
	}
	return CompositeHandler{
		h: h,
	}
}

func (c Composite) With(mdlw ...Middleware) Composite {
	return MustCompose(append([]Middleware{Middleware(c)}, mdlw...)...)
}

func (c Composite) Then(hf HandlerFunc) CompositeHandler {
	return c.Handle(hf).(CompositeHandler)
}
func (c Composite) ThenFunc(hf HandlerFunc) CompositeHandler {
	return c.Handle(hf).(CompositeHandler)
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
