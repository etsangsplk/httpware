package httpware

import "github.com/nstogner/httpware/httpctx"

// Anonymous is an implementation of the Middleware interface that has no
// dependencies and returns an empty Contains() method.
type Anonymous struct {
	h func(httpctx.Handler) httpctx.Handler
}

// Anon adapts a httpctx.Handler to an Anonymous form of the Middleware
// interface.
func Anon(h func(httpctx.Handler) httpctx.Handler) Anonymous {
	return Anonymous{
		h: h,
	}
}

// Contains returns nothing.
func (a Anonymous) Contains() []string { return []string{} }

// Requires returns nothing.
func (a Anonymous) Requires() []string { return []string{} }

// Handle returns the original handler.
func (a Anonymous) Handle(h httpctx.Handler) httpctx.Handler { return a.h(h) }
