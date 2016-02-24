package httpware

import (
	"fmt"
	"net/http"

	"github.com/nstogner/httpware/httpctx"

	"golang.org/x/net/context"
)

// The Middleware interface serves as the building block for composing
// http middleware.
type Middleware interface {
	Contains() []string
	Requires() []string
	Handle(httpctx.Handler) httpctx.Handler
}

// Composite is a collection of Middleware instances.
type Composite struct {
	middle   []Middleware
	contains []string
}

// MustCompose takes multiple Middleware instances and checks for declared
// dependencies, returning an instance of Composite. It will panic if any
// dependencies are not met.
func MustCompose(mdlw ...Middleware) *Composite {
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
	for c := range contains {
		containsSlice = append(containsSlice, c)
	}
	return &Composite{
		middle:   append(([]Middleware)(nil), mdlw...),
		contains: containsSlice,
	}
}

// Contains indentifies all of the Middleware instances included in the given
// Composite instance.
func (c *Composite) Contains() []string { return c.contains }

// Requires returns an empty array because a composite should have all
// requirements fulfilled.
func (c *Composite) Requires() []string { return []string{} }

// Handle takes the next handler as an argument and wraps it in each instance
// of Middleware contained in the Composite.
func (c *Composite) Handle(h httpctx.Handler) httpctx.Handler {
	for i := len(c.middle) - 1; i >= 0; i-- {
		h = c.middle[i].Handle(h)
	}
	return CompositeHandler{
		h: h,
	}
}

// With is a convenience method which in turn calls MustCompose, prepending
// the current Composite as the first Middleware instance.
func (c *Composite) With(mdlw ...Middleware) *Composite {
	m := make([]Middleware, 1)
	m[0] = Middleware(c)
	return MustCompose(append(m, mdlw...)...)
}

// Then is used to call the final handler than will terminate the chain of
// middleware.
func (c *Composite) Then(hf httpctx.HandlerFunc) CompositeHandler {
	return c.Handle(hf).(CompositeHandler)
}

// ThenFunc is a convenience method which calls the Then method.
func (c *Composite) ThenFunc(hf httpctx.HandlerFunc) CompositeHandler {
	return c.Handle(hf).(CompositeHandler)
}

// CompositeHandler implements the http.Handler interface, allowing it to be
// used by functions such as http.ListenAndServe. It also implements the
// httpctx.Handler interface.
type CompositeHandler struct {
	h httpctx.Handler
}

// ServeHTTP fulfills the http.Handler interface.
func (ch CompositeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ch.h.ServeHTTPCtx(context.Background(), w, r)
}

// ServeHTTPCtx fulfills the httpctx.Handler interface.
func (ch CompositeHandler) ServeHTTPCtx(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return ch.h.ServeHTTPCtx(ctx, w, r)
}
