package entitymdl

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nstogner/ctxware/adp/httpadp"
	"github.com/nstogner/ctxware/lib/httpctx"
	"github.com/nstogner/ctxware/mdl/contentmdl"

	"golang.org/x/net/context"
)

type user struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func TestUnmarshal(t *testing.T) {
	userDef := Definition{
		Entity: user{},
	}

	s := httptest.NewServer(
		httpadp.Adapt(
			contentmdl.Request(
				Unmarshal(
					httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
						u := EntityFromCtx(ctx).(*user)
						if u.Id != 123 {
							t.Fatal("expected user id to equal 123")
						}
						if u.Name != "abc" {
							t.Fatal("expected user name to equal 'abc'")
						}
						return nil
					}),
					userDef),
				contentmdl.JsonAndXml,
			),
		),
	)

	b := bytes.NewReader([]byte(`{"id": 123, "name": "abc"}`))
	_, err := http.Post(s.URL, "application/json", b)
	if err != nil {
		t.Fatal(err)
	}
}
