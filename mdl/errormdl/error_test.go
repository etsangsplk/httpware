package errormdl

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nstogner/ctxware/adp/httpadp"
	"github.com/nstogner/ctxware/lib/httpctx"
	"github.com/nstogner/ctxware/lib/httperr"

	"golang.org/x/net/context"
)

func TestHandle(t *testing.T) {
	s := httptest.NewServer(
		httpadp.Adapt(
			Handle(
				httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
					err := httperr.Err{
						StatusCode: http.StatusBadRequest,
						Message:    "better luck next time",
					}
					return err
				}),
				false,
			)))

	resp, err := http.Get(s.URL)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status code: %v, got: %v", http.StatusBadRequest, resp.StatusCode)
	}
}
