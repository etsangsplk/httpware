package contentctx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nstogner/contextware/httpctx"

	"golang.org/x/net/context"
)

type user struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func TestRequest(t *testing.T) {
	s := httptest.NewServer(
		httpctx.Adapt(
			Request(
				httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
					ct := RequestTypeFromCtx(ctx)
					switch r.URL.Path {
					case "/test-json":
						if ct.Key != KeyJson {
							t.Fatal("expected json type")
						}
						return
					case "/test-xml":
						if ct.Key != KeyXml {
							t.Fatal("expected xml type")
						}
						return
					}
					t.Fatal("this point should never have been reached")
				}),
				JsonAndXml)))

	c := http.Client{}

	req, err := http.NewRequest("GET", s.URL+"/test-json", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	_, err = c.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	req, err = http.NewRequest("GET", s.URL+"/test-xml", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/xml")
	_, err = c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
}

func TestResponse(t *testing.T) {
	s := httptest.NewServer(
		httpctx.Adapt(
			Response(
				httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
					ct := ResponseTypeFromCtx(ctx)
					switch r.URL.Path {
					case "/test-json":
						if ct.Key != KeyJson {
							t.Fatal("expected json type")
						}
						return
					case "/test-xml":
						if ct.Key != KeyXml {
							t.Fatal("expected xml type")
						}
						return
					}
					t.Fatal("this point should never have been reached")
				}),
				JsonAndXml)))

	c := http.Client{}

	req, err := http.NewRequest("GET", s.URL+"/test-json", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Accept", "application/json")
	_, err = c.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	req, err = http.NewRequest("GET", s.URL+"/test-xml", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Accept", "application/xml")
	_, err = c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
}
