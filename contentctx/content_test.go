package contentctx

import (
	"bytes"
	"encoding/json"
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

func TestUnmarshal(t *testing.T) {
	u := user{}

	s := httptest.NewServer(
		httpctx.Adapt(
			Unmarshal(
				httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
					u := EntityFromContext(ctx).(*user)
					if u.Id != 123 {
						t.Fatal("expected user id to equal 123")
					}
					if u.Name != "abc" {
						t.Fatal("expected user name to equal 'abc'")
					}
					return nil
				}),
				u, 100000, json.Unmarshal)))

	b := bytes.NewReader([]byte(`{"id": 123, "name": "abc"}`))
	_, err := http.Post(s.URL, "application/json", b)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRequest(t *testing.T) {
	s := httptest.NewServer(
		httpctx.Adapt(
			Request(
				httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
					ct := ReqContentTypeFromContext(ctx)
					switch r.URL.Path {
					case "/test-json":
						if ct.Key != KeyJson {
							t.Fatal("expected json type")
						}
						return nil
					case "/test-xml":
						if ct.Key != KeyXml {
							t.Fatal("expected xml type")
						}
						return nil
					}
					t.Fatal("this point should never have been reached")
					return nil
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
				httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
					ct := RespContentTypeFromContext(ctx)
					switch r.URL.Path {
					case "/test-json":
						if ct.Key != KeyJson {
							t.Fatal("expected json type")
						}
						return nil
					case "/test-xml":
						if ct.Key != KeyXml {
							t.Fatal("expected xml type")
						}
						return nil
					}
					t.Fatal("this point should never have been reached")
					return nil
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
