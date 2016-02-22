package httpware

import (
	"fmt"
	"net/http"

	"golang.org/x/net/context"
)

// The Middleware interface serves as the building block for composing
// http middleware.
type Middleware interface {
	Contains() []string
	Requires() []string
	Handle(Handler) Handler
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
func (c *Composite) Handle(h Handler) Handler {
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
func (c *Composite) Then(hf HandlerFunc) CompositeHandler {
	return c.Handle(hf).(CompositeHandler)
}

// ThenFunc is a convenience method which calls the Then method.
func (c *Composite) ThenFunc(hf HandlerFunc) CompositeHandler {
	return c.Handle(hf).(CompositeHandler)
}

// CompositeHandler implements the http.Handler interface, allowing it to be
// used by functions such as http.ListenAndServe. It also implements the
// httpware.Handler interface.
type CompositeHandler struct {
	h Handler
}

// ServeHTTP fulfills the http.Handler interface.
func (ch CompositeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ch.h.ServeHTTPContext(context.Background(), w, r)
}

// ServeHTTPContext fulfills the httpware.Handler interface.
func (ch CompositeHandler) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return ch.h.ServeHTTPContext(ctx, w, r)
}
