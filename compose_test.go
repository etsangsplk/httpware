package httpware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nstogner/httpware/httpctx"

	"golang.org/x/net/context"
)

type testMiddle1 struct {
}

func newTM1() testMiddle1 {
	return testMiddle1{}
}

func (tm1 testMiddle1) Contains() []string {
	return []string{"testmiddle1.Ware"}
}

func (tm1 testMiddle1) Requires() []string {
	return []string{}
}

func (tm1 testMiddle1) Handle(h httpctx.Handler) httpctx.Handler {
	return httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		w.Header().Set("middle1", "true")
		return h.ServeHTTPCtx(ctx, w, r)
	})
}

type testMiddle2 struct {
}

func newTM2() testMiddle2 {
	return testMiddle2{}
}

func (tm2 testMiddle2) Contains() []string {
	return []string{"testmiddle2.Ware"}
}

func (tm2 testMiddle2) Requires() []string {
	return []string{"testmiddle1.Ware"}
}

func (tm2 testMiddle2) Handle(h httpctx.Handler) httpctx.Handler {
	return httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		w.Header().Set("middle2", "true")
		return h.ServeHTTPCtx(ctx, w, r)
	})
}

func testAdapt(h httpctx.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTPCtx(context.Background(), w, r)
	})
}

func TestComposeMissingDep(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic from MustCompose with missing dependencies")
		}
	}()
	MustCompose(
		newTM2(),
		newTM1(),
	)
}

func TestDoubleCompose(t *testing.T) {
	c1 := MustCompose(
		newTM1(),
	)
	MustCompose(
		c1,
		newTM2(),
	)
}

func TestCompose(t *testing.T) {
	c := MustCompose(
		newTM1(),
		newTM2(),
	)

	s := httptest.NewServer(c.ThenFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}))

	resp, err := http.Get(s.URL)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Header.Get("middle1") != "true" {
		t.Fatal("expected 'middle1' header to be set")
	}
	if resp.Header.Get("middle2") != "true" {
		t.Fatal("expected 'middle2' header to be set")
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status code %v, got %v", http.StatusNoContent, resp.StatusCode)
	}
}
