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
	conf := Config{
		RemoteLimit: 3,
		TotalLimit:  10,
	}
	m := httpware.MustCompose(
		errorware.New(),
		New(conf),
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

	for i := 1; i <= conf.RemoteLimit+1; i++ {
		if i > conf.RemoteLimit {
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

func TestTotalLimit(t *testing.T) {
	// Define multilayer chaining.
	conf := Config{
		RemoteLimit: 60,
		TotalLimit:  10,
	}
	mid := New(conf)
	c0 := httpware.MustCompose(errorware.New(), mid)
	c1 := c0.With(testWare{})

	// Start test servers.
	s := make([]*httptest.Server, 2)
	s[0] = httptest.NewServer(c0.ThenFunc(testHandler))
	s[1] = httptest.NewServer(c1.ThenFunc(testHandler))

	for i := 1; uint64(i) <= conf.TotalLimit+2; i++ {
		si := i % 2
		if uint64(i) > conf.TotalLimit {
			// Sleep for 100ms to make sure the other requests have reached the middleware.
			time.Sleep(100 * time.Millisecond)
			resp, err := http.Get(s[si].URL)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != 429 {
				t.Fatalf("expected status code %v, got %v", 429, resp.StatusCode)
			}
		} else {
			go http.Get(s[si].URL + "/delay")
		}
	}

}

type testWare struct{}

func (m testWare) Contains() []string                         { return []string{} }
func (m testWare) Requires() []string                         { return []string{} }
func (m testWare) Handle(h httpware.Handler) httpware.Handler { return h }

func testHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if r.URL.Path == "/delay" {
		time.Sleep(10 * time.Second)
	}
	return nil
}

// TestCompositeLimit intends to ensure that a top-level limit is imposed
// on any further chaining of middleware.
func TestCompositeLimit(t *testing.T) {
	// Define multilayer chaining.
	conf := Config{
		RemoteLimit: 6,
		TotalLimit:  10,
	}
	mid := New(conf)
	c0 := httpware.MustCompose(errorware.New(), mid)
	c1 := c0.With(testWare{})

	// Start test servers.
	s := make([]*httptest.Server, 2)
	s[0] = httptest.NewServer(c0.ThenFunc(testHandler))
	s[1] = httptest.NewServer(c1.ThenFunc(testHandler))

	for i := 1; i <= conf.RemoteLimit+2; i++ {
		si := i % 2
		if i > conf.RemoteLimit {
			// Sleep for 100ms to make sure the other requests have reached the middleware.
			time.Sleep(100 * time.Millisecond)
			resp, err := http.Get(s[si].URL)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != 429 {
				t.Fatalf("expected status code %v, got %v", 429, resp.StatusCode)
			}
		} else {
			go http.Get(s[si].URL + "/delay")
		}
	}

}
