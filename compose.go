package httpware

import (
	"net/http"

	"github.com/nstogner/httpware/httpctx"

	"golang.org/x/net/context"
)

// The Middleware interface serves as the building block for composing
// http middleware.
type Middleware interface {
	Handle(httpctx.Handler) httpctx.Handler
}

// Composite is a collection of Middleware instances.
type Composite struct {
	middle []Middleware
}

// Compose takes multiple Middleware instances and returns a Composite
// instance.
func Compose(mdlw ...Middleware) *Composite {
	return &Composite{
		middle: append(([]Middleware)(nil), mdlw...),
	}
}

// Handle takes the next handler as an argument and wraps it in each instance
// of Middleware contained in the Composite.
func (c *Composite) Handle(h httpctx.Handler) httpctx.Handler {
	for i := len(c.middle) - 1; i >= 0; i-- {
		h = c.middle[i].Handle(h)
	}
	return CompositeHandler{
		// Wrap with error-handling middleware.
		h: handleHttpErrors(h),
	}
}

// With is a convenience method which in turn calls Compose, prepending
// the current Composite as the first Middleware instance.
func (c *Composite) With(mdlw ...Middleware) *Composite {
	m := make([]Middleware, 1)
	m[0] = Middleware(c)
	return Compose(append(m, mdlw...)...)
}

// Then is used to call the final handler than will terminate the chain of
// middleware.
func (c *Composite) Then(hf httpctx.Handler) CompositeHandler {
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
