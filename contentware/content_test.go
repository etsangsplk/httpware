package contentware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nstogner/httpware"

	"golang.org/x/net/context"
)

type user struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func TestRequest(t *testing.T) {
	c := httpware.MustCompose(
		NewReqType(JsonAndXml),
	)
	s := httptest.NewServer(
		c.Then(
			httpware.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
				ct := RequestTypeFromCtx(ctx)
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
		),
	)

	hc := http.Client{}

	req, err := http.NewRequest("GET", s.URL+"/test-json", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	_, err = hc.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	req, err = http.NewRequest("GET", s.URL+"/test-xml", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/xml")
	_, err = hc.Do(req)
	if err != nil {
		t.Fatal(err)
	}
}

func TestResponse(t *testing.T) {
	c := httpware.MustCompose(
		NewRespType(JsonAndXml),
	)
	s := httptest.NewServer(
		c.Then(
			httpware.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
				ct := ResponseTypeFromCtx(ctx)
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
		),
	)

	hc := http.Client{}

	req, err := http.NewRequest("GET", s.URL+"/test-json", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Accept", "application/json")
	_, err = hc.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	req, err = http.NewRequest("GET", s.URL+"/test-xml", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Accept", "application/xml")
	_, err = hc.Do(req)
	if err != nil {
		t.Fatal(err)
	}
}
