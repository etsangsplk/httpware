package boltctx

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/nstogner/contextware/contentctx"
	"github.com/nstogner/contextware/entityctx"
	"github.com/nstogner/contextware/errorctx"
	"github.com/nstogner/contextware/httpctx"
)

type user struct {
	Id   string
	Name string
}

func TestPost(t *testing.T) {
	db, err := bolt.Open("test.db", 0600, nil)
	defer func() {
		os.Remove("test.db")
	}()
	if err != nil {
		t.Fatal(err)
	}

	def := Definition{
		DB:         db,
		BucketPath: "/users/",
		Identify: func(u interface{}) []byte {
			usr := u.(*user)
			return []byte(usr.Id)
		},
	}

	userDef := &entityctx.Definition{
		Entity: user{},
	}

	s := httptest.NewServer(
		httpctx.Adapt(
			errorctx.Handle(
				contentctx.Request(
					contentctx.Response(
						entityctx.Unmarshal(
							Post(def),
							userDef,
						),
						contentctx.JsonAndXml,
					),
					contentctx.JsonAndXml,
				),
				false,
			),
		))

	resp, err := http.Post(s.URL, "application/json", bytes.NewReader([]byte(`{"id":"abc", "name":"somebody"}`)))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatal("expected status code:", http.StatusCreated)
	}
}
