package routerezy

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	"golang.org/x/net/context"
)

func BenchmarkGet(b *testing.B) {
	r := httprouter.New()
	r.GET("/test", Get(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "handled")
	}))
	req, err := http.NewRequest("GET", "http://example.com/test", nil)
	if err != nil {
		b.Fatal(err)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(w, req)
	}
}
