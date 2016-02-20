package routeradapt

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	"golang.org/x/net/context"
)

func TestAdapt(t *testing.T) {
	r := httprouter.New()
	r.GET("/test/:id", AdaptFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		ps := ParamsFromCtx(ctx)
		if ps.ByName("id") != "abc" {
			t.Fatal("expected id param to equal 'abc'")
		}
		w.WriteHeader(http.StatusNoContent)
		return nil
	}))
	s := httptest.NewServer(r)
	resp, err := http.Get(s.URL + "/test/abc")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status code %v, got %v", http.StatusNoContent, resp.StatusCode)
	}
}