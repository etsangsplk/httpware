package corsware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nstogner/httpware"

	"golang.org/x/net/context"
)

func TestWare(t *testing.T) {
	m := httpware.Compose(
		httpware.DefaultErrHandler,
		New(Defaults),
	)
	s := httptest.NewServer(m.ThenFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}))

	// Send Request.
	resp, err := http.Get(s.URL)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status code %v, got %v", http.StatusNoContent, resp.StatusCode)
	}
	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Fatal("expected Access-Control-Allow-Origin header to be set to '*'")
	}
}
