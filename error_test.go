package httpware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nstogner/httpware/httpctx"
	"github.com/nstogner/httpware/httperr"

	"golang.org/x/net/context"
)

func TestHandle(t *testing.T) {
	c := Compose()
	s := httptest.NewServer(
		c.Then(
			httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
				err := httperr.Err{
					StatusCode: http.StatusBadRequest,
					Message:    "better luck next time",
				}
				return err
			}),
		),
	)

	resp, err := http.Get(s.URL)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status code: %v, got: %v", http.StatusBadRequest, resp.StatusCode)
	}
}
