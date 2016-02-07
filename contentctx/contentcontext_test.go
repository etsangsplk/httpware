package contentctx

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nstogner/netmiddle/httpctx"

	"golang.org/x/net/context"
)

type user struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func TestMarshal(t *testing.T) {
	u := user{}

	s := httptest.NewServer(
		httpctx.Adapt(
			Unmarshal(
				httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
					u := EntityFromContext(ctx).(*user)
					if u.Id != 123 {
						t.Fatal("expected user id to equal 123")
					}
					if u.Name != "abc" {
						t.Fatal("expected user name to equal 'abc'")
					}
					return nil
				}),
				u, 100000, json.Unmarshal)))

	b := bytes.NewReader([]byte(`{"id": 123, "name": "abc"}`))
	_, err := http.Post(s.URL, "application/json", b)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNegotiate(t *testing.T) {
	u := user{
		Id:   456,
		Name: "XYZ",
	}

	js := httptest.NewServer(
		httpctx.Adapt(
			Negotiate(
				httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
					reqContentType := ReqContentTypeFromContext(ctx)
					reqContentType.Marshal(w, u)
					return nil
				}),
				[]*ContentType{ContentTypeJson, ContentTypeXml}, []*ContentType{ContentTypeJson, ContentTypeXml})))

	c := http.Client{}

	req, err := http.NewRequest("GET", js.URL+"/test-json", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Header.Get("Content-Type") != ContentTypeJson.Value {
		t.Fatal("expected content-type: " + ContentTypeJson.Value)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &u); err != nil {
		t.Fatal(err)
	}

	req.Header.Set("accept", "")
	resp, err = c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Header.Get("Content-Type") != ContentTypeJson.Value {
		t.Fatal("expected content-type: " + ContentTypeJson.Value)
	}

	req, err = http.NewRequest("GET", js.URL+"/test-xml", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("accept", "applicaiton/xml")
	req.Header.Set("Content-Type", "application/xml")
	resp, err = c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Header.Get("Content-Type") != ContentTypeXml.Value {
		t.Fatal("expected content-type: " + ContentTypeXml.Value)
	}
	body, _ = ioutil.ReadAll(resp.Body)
	if err := xml.Unmarshal(body, &u); err != nil {
		t.Fatal(err)
	}
}
