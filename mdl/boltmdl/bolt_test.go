package boltmdl

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
	"github.com/nstogner/ctxware/adp/httpadp"
	"github.com/nstogner/ctxware/adp/routeradp"
	"github.com/nstogner/ctxware/mdl/contentmdl"
	"github.com/nstogner/ctxware/mdl/entitymdl"
	"github.com/nstogner/ctxware/mdl/errormdl"
	"github.com/nstogner/ctxware/mdl/logmdl"
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
		Identify: func(u interface{}) string {
			return u.(*user).Id
		},
		EntityDef: entitymdl.Definition{
			Entity: user{},
		},
	}

	ps := httptest.NewServer(
		httpadp.Adapt(
			errormdl.Handle(
				contentmdl.Request(
					contentmdl.Response(
						entitymdl.Unmarshal(
							Post(nil, def),
							def.EntityDef,
						),
						contentmdl.JsonAndXml,
					),
					contentmdl.JsonAndXml,
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
		routeradp.Adapt(
			errormdl.Handle(
				logmdl.Errors(
					contentmdl.Response(
						Get(nil, def),
						contentmdl.JsonAndXml,
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
