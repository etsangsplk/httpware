package boltezy

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
	"github.com/nstogner/ctxware/mdl/boltmdl"
	"github.com/nstogner/ctxware/mdl/entitymdl"
)

type user struct {
	Id   string
	Name string
}

func TestPostGet(t *testing.T) {
	// Perform db setup.
	db, err := bolt.Open("test.db", 0600, nil)
	defer func() {
		os.Remove("test.db")
	}()
	if err != nil {
		t.Fatal(err)
	}

	// Define bolt user entity.
	def := boltmdl.Definition{
		DB:               db,
		EntityBucketPath: "/users/",
		IdParam:          "id",
		IdField:          "Id",
		EntityDef: entitymdl.Definition{
			Entity:   user{},
			Validate: func(u interface{}) error { return nil },
		},
	}

	// Define routes.
	r := httprouter.New()
	r.POST("/users", Post(nil, def))
	r.GET("/users/:id", Get(nil, def))
	s := httptest.NewServer(r)

	// Test POST endpoint.
	resp, err := http.Post(s.URL+"/users", "application/json", bytes.NewReader([]byte(`{"id":"abc", "name":"somebody"}`)))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatal("expected status code:", http.StatusCreated)
	}

	// Test GET endpoint.
	resp, err = http.Get(s.URL + "/users/abc")
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
