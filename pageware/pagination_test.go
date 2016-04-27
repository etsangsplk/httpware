package pageware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nstogner/httpware"
	"golang.org/x/net/context"
)

func TestPagination(t *testing.T) {
	m := httpware.Compose(
		httpware.DefaultErrHandler,
		New(Defaults),
	)
	s := httptest.NewServer(m.ThenFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		page := PageFromCtx(ctx)
		switch r.URL.RawQuery {
		case "":
			if page.Start != 0 {
				t.Fatal("expected page.Start == 0")
			}
			if page.Limit != Defaults.LimitDefault {
				t.Fatal("expected page.Limit == Defaults.LimitDefault")
			}
		case "start=10&limit=5":
			if page.Start != 10 {
				t.Fatal("expected page.Start == 10")
			}
			if page.Limit != 5 {
				t.Fatal("expected page.Limit == 5")
			}
		default:
			t.Fatal("this point should not be reached")
		}
		return nil
	}))

	// Valid test cases:
	http.Get(s.URL)
	http.Get(s.URL + "?start=10&limit=5")
	// Invalid test cases:
	r, _ := http.Get(s.URL + "?start=-10&limit=5")
	if r.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status code %v, got: %v, while testing negative start param", http.StatusBadRequest, r.StatusCode)
	}
	r, _ = http.Get(s.URL + "?start=10&limit=0")
	if r.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status code %v, got: %v, while testing zero limit param", http.StatusBadRequest, r.StatusCode)
	}
}
