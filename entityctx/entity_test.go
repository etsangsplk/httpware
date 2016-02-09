package entityctx

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nstogner/ctxware/contentctx"
	"github.com/nstogner/ctxware/httpctx"
	"golang.org/x/net/context"
)

type user struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func TestUnmarshal(t *testing.T) {
	userDef := &Definition{
		Entity: user{},
	}

	s := httptest.NewServer(
		httpctx.Adapt(
			contentctx.Request(
				Unmarshal(
					httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
						u := EntityFromCtx(ctx).(*user)
						if u.Id != 123 {
							t.Fatal("expected user id to equal 123")
						}
						if u.Name != "abc" {
							t.Fatal("expected user name to equal 'abc'")
						}
					}),
					userDef),
				contentctx.JsonAndXml,
			),
		),
	)

	b := bytes.NewReader([]byte(`{"id": 123, "name": "abc"}`))
	_, err := http.Post(s.URL, "application/json", b)
	if err != nil {
		t.Fatal(err)
	}
}
