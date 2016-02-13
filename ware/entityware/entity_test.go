package entityware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nstogner/ctxware"
	"github.com/nstogner/ctxware/ware/contentware"

	"golang.org/x/net/context"
)

type user struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func TestParser(t *testing.T) {
	c := ctxware.MustCompose(
		contentware.NewReqType(contentware.JsonAndXml),
		NewParser(user{}, Maximum),
	)

	s := httptest.NewServer(
		c.Then(
			ctxware.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
				u := EntityFromCtx(ctx).(*user)
				if u.Id != 123 {
					t.Fatal("expected user id to equal 123")
				}
				if u.Name != "abc" {
					t.Fatal("expected user name to equal 'abc'")
				}
				return nil
			}),
		),
	)

	b := bytes.NewReader([]byte(`{"id": 123, "name": "abc"}`))
	_, err := http.Post(s.URL, "application/json", b)
	if err != nil {
		t.Fatal(err)
	}
}