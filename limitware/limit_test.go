package limitware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nstogner/httpware"
	"github.com/nstogner/httpware/errorware"

	"golang.org/x/net/context"
)

func TestRemoteLimit(t *testing.T) {
	def := Def{
		RemoteLimit: 3,
		TotalLimit:  10,
	}
	m := httpware.MustCompose(
		errorware.New(),
		NewReqLimit(def),
	)
	s := httptest.NewServer(m.ThenFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		if r.URL.Path == "/delay" {
			time.Sleep(1 * time.Second)
		}
		return nil
	}))

	resp, err := http.Get(s.URL)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected status code %v, got %v", 200, resp.StatusCode)
	}

	for i := 1; i <= def.RemoteLimit+1; i++ {
		if i == def.RemoteLimit+1 {
			// Sleep for 100ms to make sure the other requests have reached the middleware.
			time.Sleep(100 * time.Millisecond)
			resp, err := http.Get(s.URL)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != 429 {
				t.Fatalf("expected status code %v, got %v", 429, resp.StatusCode)
			}
		} else {
			go http.Get(s.URL + "/delay")
		}
	}
}

// TODO: TestTotalLimit
