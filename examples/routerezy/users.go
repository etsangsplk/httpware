package main

// To run and play with the server:
//   go run examples/routerezy/users.go
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
	"github.com/nstogner/ctxware/adp/routeradp"
	"github.com/nstogner/ctxware/ezy/routerezy"
	"github.com/nstogner/ctxware/mdl/contentmdl"
	"github.com/nstogner/ctxware/mdl/entitymdl"

	"golang.org/x/net/context"
)

type User struct {
	Id   string `json:"id" xml:"id"`
	Name string `json:"name" xml:"name"`
}

func main() {
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8080"
	}

	r := httprouter.New()

	userDef := entitymdl.Definition{
		Entity: User{},
		Validate: func(u interface{}) error {
			usr := u.(*User)
			if len(usr.Id) < 5 {
				return errors.New("user id must be at least 5 characters")
			}
			return nil
		},
	}

	r.GET("/users/:id", routerezy.Get(handleGet))
	r.POST("/users", routerezy.Post(handlePost, userDef))

	logrus.WithField("port", port).Info("starting server...")
	logrus.Fatal(http.ListenAndServe(":"+port, r))

}

func handlePost(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	u := entitymdl.EntityFromCtx(ctx).(*User)

	w.WriteHeader(http.StatusCreated)
	rct := contentmdl.ResponseTypeFromCtx(ctx)
	rct.MarshalWrite(w, u)
	return nil
}

func handleGet(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	params := routeradp.ParamsFromCtx(ctx)

	usrId := params["id"]
	u := &User{Id: usrId, Name: "sammy"}

	ct := contentmdl.ResponseTypeFromCtx(ctx)
	ct.MarshalWrite(w, u)
	return nil
}
