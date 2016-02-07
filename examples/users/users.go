package main

// To run and play with the server:
//   go run examples/users/users.go
//   curl localhost:8080 -v -H "Content-Type: application/json" -d '{"id": "bdog", "name": "bob"}'
//   curl localhost:8080 -v -H "Content-Type: application/xml"  -d '<user><id>bdog</id><name>bob</name></user>'
//   curl localhost:8080 -v -H "Content-Type: application/xml" -H "Accept: application/xml" -d '<user><id>bdog</id><name>bob</name></user>'
//   curl localhost:8080/sman -v

import (
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"github.com/nstogner/netmiddle/contentctx"
	"github.com/nstogner/netmiddle/httpctx"
	"github.com/nstogner/netmiddle/routerctx"

	"golang.org/x/net/context"
)

type user struct {
	Id   string `json:"id" xml:"id"`
	Name string `json:"name" xml:"name"`
}

func main() {
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8080"
	}

	contentTypes := []*contentctx.ContentType{
		contentctx.ContentTypeJson,
		contentctx.ContentTypeXml,
	}

	r := httprouter.New()

	r.GET("/:id", routerctx.Adapt(
		contentctx.Negotiate(
			httpctx.HandlerFunc(handleGet), contentTypes, contentTypes)))
	r.POST("/", routerctx.Adapt(
		contentctx.Negotiate(
			contentctx.Unmarshal(httpctx.HandlerFunc(handlePost), user{}, 10000, nil), contentTypes, contentTypes)))

	logrus.WithField("port", port).Info("starting server...")
	logrus.Fatal(http.ListenAndServe(":"+port, r))

}

func handlePost(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	u := contentctx.EntityFromContext(ctx).(*user)

	w.WriteHeader(http.StatusCreated)
	rct := contentctx.RespContentTypeFromContext(ctx)
	rct.Marshal(w, u)
	return nil
}

func handleGet(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ps := routerctx.ParamsFromContext(ctx)

	usrId := ps.ByName("id")
	u := &user{Id: usrId, Name: "sammy"}

	ct := contentctx.RespContentTypeFromContext(ctx)
	ct.Marshal(w, u)
	return nil
}
