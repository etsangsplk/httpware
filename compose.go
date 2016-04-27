package httpware

import (
	"net/http"

	"golang.org/x/net/context"
)

// The Middleware interface serves as the building block for composing
// http middleware.
type Middleware interface {
	Handle(Handler) Handler
}

// The Errware interface serves as a check point for the non-nil
// error return values from downstream middleware.
type Errware interface {
	HandleErr(Handler) Handler
}

// Composite is a collection of Middleware instances.
type Composite struct {
	errWare Errware
	middle  []Middleware
}

// Compose takes multiple Middleware instances and returns a Composite
// instance.
func Compose(errWare Errware, mdlw ...Middleware) *Composite {
	return &Composite{
		errWare: errWare,
		middle:  append(([]Middleware)(nil), mdlw...),
	}
}

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

// With is a convenience method which in turn calls Compose, prepending
// the current Composite as the first Middleware instance.
func (c *Composite) With(mdlw ...Middleware) *Composite {
	m := make([]Middleware, 1)
	m[0] = Middleware(c)
	return Compose(c.errWare, append(m, mdlw...)...)
}

// Then is used to call the final handler than will terminate the chain of
// middleware.
func (c *Composite) Then(h Handler) CompositeHandler {
	return CompositeHandler{c.errWare.HandleErr(c.Handle(h))}
}

// ThenFunc is a convenience method which calls the Then method.
func (c *Composite) ThenFunc(hf HandlerFunc) CompositeHandler {
	return c.Then(hf)
}

// CompositeHandler implements the http.Handler interface, allowing it to be
// used by functions such as http.ListenAndServe. It also implements the
// Handler interface.
type CompositeHandler struct {
	h Handler
}

// ServeHTTP fulfills the http.Handler interface.
func (ch CompositeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ch.h.ServeHTTPCtx(context.Background(), w, r)
}

// ServeHTTPCtx fulfills the Handler interface.
func (ch CompositeHandler) ServeHTTPCtx(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return ch.h.ServeHTTPCtx(ctx, w, r)
}
