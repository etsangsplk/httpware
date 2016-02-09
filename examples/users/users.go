package main

// To run and play with the server:
//   go run examples/users/users.go
//   curl localhost:8080/users -v -H "Content-Type: application/json" -d '{"id": "bdog", "name": "bob"}'
//   curl localhost:8080/users -v -H "Content-Type: application/xml"  -d '<user><id>bdog</id><name>bob</name></user>'
//   curl localhost:8080/users -v -H "Content-Type: application/xml" -H "Accept: application/xml" -d '<user><id>bdog</id><name>bob</name></user>'
//   curl localhost:8080/users/sam -v

import (
	"errors"
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"github.com/nstogner/ctxware/contentctx"
	"github.com/nstogner/ctxware/easyctx"
	"github.com/nstogner/ctxware/entityctx"
	"github.com/nstogner/ctxware/routerctx"

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

	userDef := entityctx.Definition{
		Entity: user{},
		Validate: func(u interface{}) error {
			usr := u.(*user)
			if len(usr.Id) < 5 {
				return errors.New("user id must be at least 5 characters")
			}
			return nil
		},
	}

	r.GET("/users/:id", easyctx.Get(handleGet))
	r.POST("/users", easyctx.Post(handlePost, userDef))

	logrus.WithField("port", port).Info("starting server...")
	logrus.Fatal(http.ListenAndServe(":"+port, r))

}

func handlePost(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	u := entityctx.EntityFromCtx(ctx).(*user)

	w.WriteHeader(http.StatusCreated)
	rct := contentctx.ResponseTypeFromCtx(ctx)
	rct.MarshalWrite(w, u)
	return
}

func handleGet(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	params := routerctx.ParamsFromCtx(ctx)

	usrId := params["id"]
	u := &user{Id: usrId, Name: "sammy"}

	ct := contentctx.ResponseTypeFromCtx(ctx)
	ct.MarshalWrite(w, u)
	return
}
