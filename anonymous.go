package httpware

// Anonymous is an implementation of the Middleware interface that has no
// dependencies and returns an empty Contains() method.
type Anonymous struct {
	h func(Handler) Handler
}

// Anon adapts a Handler to an Anonymous form of the Middleware
// interface.
func Anon(h func(Handler) Handler) Anonymous {
	return Anonymous{
		h: h,
	}
}

func (a Anonymous) Contains() []string       { return []string{} }
func (a Anonymous) Requires() []string       { return []string{} }
func (a Anonymous) Handle(h Handler) Handler { return a.h(h) }
