package httpware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nstogner/httpware/httpctx"

	"golang.org/x/net/context"
)

func someTestingMiddlware(next httpctx.Handler) httpctx.Handler {
	return httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		w.Header().Set("anon", "true")
		return next.ServeHTTPCtx(ctx, w, r)
	})
}

func TestAnon(t *testing.T) {
	a := Anon(someTestingMiddlware)
	m := MustCompose(a)
	s := httptest.NewServer(m.ThenFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error { return nil }))
	resp, err := http.Get(s.URL)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Header.Get("anon") != "true" {
		t.Fatal("expected 'anon' header to be set to 'true'")
	}
}
