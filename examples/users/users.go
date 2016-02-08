package main

// To run and play with the server:
//   go run examples/users/users.go
//   curl localhost:8080 -v -H "Content-Type: application/json" -d '{"id": "bdog", "name": "bob"}'
//   curl localhost:8080 -v -H "Content-Type: application/xml"  -d '<user><id>bdog</id><name>bob</name></user>'
//   curl localhost:8080 -v -H "Content-Type: application/xml" -H "Accept: application/xml" -d '<user><id>bdog</id><name>bob</name></user>'
//   curl localhost:8080/sam -v

import (
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"github.com/nstogner/contextware/contentctx"
	"github.com/nstogner/contextware/easyctx"
	"github.com/nstogner/contextware/routerctx"

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

	r := httprouter.New()

	r.GET("/:id", easyctx.Get(handleGet))
	r.POST("/", easyctx.Post(handlePost, user{}))

	logrus.WithField("port", port).Info("starting server...")
	logrus.Fatal(http.ListenAndServe(":"+port, r))

}

func handlePost(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	u := contentctx.EntityFromCtx(ctx).(*user)

	w.WriteHeader(http.StatusCreated)
	rct := contentctx.ResponseTypeFromCtx(ctx)
	rct.MarshalWrite(w, u)
	return nil
}

func handleGet(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ps := routerctx.ParamsFromCtx(ctx)

	usrId := ps.ByName("id")
	u := &user{Id: usrId, Name: "sammy"}

	ct := contentctx.ResponseTypeFromCtx(ctx)
	ct.MarshalWrite(w, u)
	return nil
}
