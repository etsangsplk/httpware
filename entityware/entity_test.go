package entityware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nstogner/httpware"
	"github.com/nstogner/httpware/contentware"
	"github.com/nstogner/httpware/httpctx"

	"golang.org/x/net/context"
)

type user struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestParsing(t *testing.T) {
	c := httpware.MustCompose(
		contentware.New(contentware.Defaults),
		New(Config{
			Entity:      user{},
			MaxByteSize: MAX,
		}),
	)

	s := httptest.NewServer(
		c.Then(
			httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
				u := EntityFromCtx(ctx).(*user)
				if u.ID != 123 {
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

// TODO: TestValidate
