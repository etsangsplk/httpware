package boltctx

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/julienschmidt/httprouter"
	"github.com/nstogner/ctxware/contentctx"
	"github.com/nstogner/ctxware/entityctx"
	"github.com/nstogner/ctxware/errorctx"
	"github.com/nstogner/ctxware/httpctx"
	"github.com/nstogner/ctxware/logctx"
	"github.com/nstogner/ctxware/routerctx"
)

type user struct {
	Id   string
	Name string
}

func TestPostGet(t *testing.T) {
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
		IdParam:    "id",
		Identify: func(u interface{}) []byte {
			usr := u.(*user)
			return []byte(usr.Id)
		},
		EntityDef: entityctx.Definition{
			Entity: user{},
		},
	}

	ps := httptest.NewServer(
		httpctx.Adapt(
			errorctx.Handle(
				contentctx.Request(
					contentctx.Response(
						entityctx.Unmarshal(
							Post(def),
							def.EntityDef,
						),
						contentctx.JsonAndXml,
					),
					contentctx.JsonAndXml,
				),
				false,
			),
		))

	resp, err := http.Post(ps.URL, "application/json", bytes.NewReader([]byte(`{"id":"abc", "name":"somebody"}`)))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatal("expected status code:", http.StatusCreated)
	}

	// Test Get...
	r := httprouter.New()
	r.GET("/users/:id",
		routerctx.Adapt(
			errorctx.Handle(
				logctx.Errors(
					contentctx.Response(
						Get(def),
						contentctx.JsonAndXml,
					),
				),
				false,
			),
		))
	gs := httptest.NewServer(r)

	resp, err = http.Get(gs.URL + "/users/abc")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status code %v, got %v", http.StatusOK, resp.StatusCode)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	if !strings.Contains(string(body), "somebody") {
		t.Fatal("unexpected entity from GET request")
	}
}
