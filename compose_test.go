package ctxware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/net/context"
)

type testMiddle1 struct {
}

func NewTM1() testMiddle1 {
	return testMiddle1{}
}

func (tm1 testMiddle1) Name() string {
	return "testmiddle1.Ware"
}

func (tm1 testMiddle1) Dependencies() []string {
	return []string{}
}

func (tm1 testMiddle1) Handle(h Handler) Handler {
	return HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		w.Header().Set("middle1", "true")
		return h.ServeHTTPContext(ctx, w, r)
	})
}

type testMiddle2 struct {
}

func NewTM2() testMiddle2 {
	return testMiddle2{}
}

func (tm2 testMiddle2) Name() string {
	return "testmiddle2.Ware"
}

func (tm2 testMiddle2) Dependencies() []string {
	return []string{"testmiddle1.Ware"}
}

func (tm2 testMiddle2) Handle(h Handler) Handler {
	return HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		w.Header().Set("middle2", "true")
		return h.ServeHTTPContext(ctx, w, r)
	})
}

func testAdapt(h Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTPContext(context.Background(), w, r)
	})
}

func TestComposeMissingDep(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic from MustCompose with missing dependencies")
		}
	}()
	MustCompose(
		NewTM2(),
		NewTM1(),
	)
}

func TestCompose(t *testing.T) {
	c := MustCompose(
		NewTM1(),
		NewTM2(),
	)

	s := httptest.NewServer(c.Then(HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		w.WriteHeader(http.StatusNoContent)
		return nil
	})))

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
		t.Fatal("expected status code %v, got %v", http.StatusNoContent, resp.StatusCode)
	}
}
