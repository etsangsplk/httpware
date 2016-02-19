package httpwarebenchmarks

import (
	"net/http"
	"testing"
)

type respWriter struct {
}

func (w *respWriter) Header() http.Header {
	return http.Header{}
}
func (w *respWriter) Write(b []byte) (int, error) {
	return 1, nil
}
func (w *respWriter) WriteHeader(h int) {
	return
}

func do(s string) {
	i := len(s)
	i++
	i--
}

// What if I just always call the function?
func BenchmarkClosure(b *testing.B) {
	w := &respWriter{}
	s := "hey"
	var someFunc func(http.ResponseWriter)
	// In middleware, I could define closures to be called in handlers...
	if s == "hey" {
		someFunc = func(rw http.ResponseWriter) { do(s) }
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		someFunc(w)
	}
}

// What if I run string comparisons in the handlers?
func BenchmarkIfStringMatch(b *testing.B) {
	_ = &respWriter{}
	s := "hey there"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if s == "hey there" {
			do(s)
		}
	}
}

// Should I run string comparisons before the handlers?
func BenchmarkIfBool(b *testing.B) {
	_ = &respWriter{}
	s := "hey"
	boolVal := (s == "hey")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if boolVal {
			do(s)
		}
	}
}
