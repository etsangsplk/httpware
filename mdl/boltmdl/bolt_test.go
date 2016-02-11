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
		DB:               db,
		EntityBucketPath: "/users/",
		IdParam:          "id",
		IdField:          "Id",
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

func TestPutGetETags(t *testing.T) {
	db, err := bolt.Open("test.db", 0600, nil)
	defer func() {
		os.Remove("test.db")
	}()
	if err != nil {
		t.Fatal(err)
	}

	def := Definition{
		DB:               db,
		EntityBucketPath: "/users/",
		ETagBucketPath:   "/etags/users/",
		IdParam:          "id",
		IdField:          "Id",
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
							Put(nil, def),
							def.EntityDef,
						),
						contentmdl.JsonAndXml,
					),
					contentmdl.JsonAndXml,
				),
				false,
			),
		))

	client := &http.Client{}

	// Initial PUT
	req, err := http.NewRequest("PUT", ps.URL, bytes.NewReader([]byte(`{"id":"abc", "name":"somebody"}`)))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatal("expected status code:", http.StatusCreated)
	}
	etag := resp.Header.Get("ETag")
	if etag == "" {
		t.Fatal("expected ETag header to be set with successfull PUT")
	}

	// Second PUT with bad ETag
	req, err = http.NewRequest("PUT", ps.URL, bytes.NewReader([]byte(`{"id":"abc", "name":"somebody"}`)))
	req.Header.Set("If-None-Match", "12830198091fjakslfdj0")
	if err != nil {
		t.Fatal(err)
	}
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 412 {
		t.Fatalf("expected status code: %v, got: %v", http.StatusConflict, 412)
	}

	// Third PUT with ETag
	req, err = http.NewRequest("PUT", ps.URL, bytes.NewReader([]byte(`{"id":"abc", "name":"somebody"}`)))
	req.Header.Set("If-None-Match", etag)
	if err != nil {
		t.Fatal(err)
	}
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status code: %v, got: %v", http.StatusOK, resp.StatusCode)
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
