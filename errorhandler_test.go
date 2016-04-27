package httpware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/net/context"
)

func TestHandle(t *testing.T) {
	c := Compose(DefaultErrHandler)
	s := httptest.NewServer(
		c.Then(
			HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
				err := Err{
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
