package errorctx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nstogner/ctxware/httpctx"
	"github.com/nstogner/ctxware/httperr"
	"golang.org/x/net/context"
)

func TestHandle(t *testing.T) {
	s := httptest.NewServer(
		httpctx.Adapt(
			Handle(
				httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
					httperr.Return(httperr.Err{
						StatusCode: http.StatusBadRequest,
						Message:    "better luck next time",
					})
				}),
				false,
			)))

	resp, err := http.Get(s.URL)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatal("expected status code:", http.StatusBadRequest)
	}
}
