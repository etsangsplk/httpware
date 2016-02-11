package main

// To run and play with the server:
//   go run examples/boltezy/users.go
//   curl localhost:8080/users -v -H "Content-Type: application/json" -d '{"id": "bdog", "name": "bob"}'
//   curl localhost:8080/users -v -H "Content-Type: application/xml"  -d '<user><id>bdog</id><name>bob</name></user>'
//   curl localhost:8080/users -v -H "Content-Type: application/xml" -H "Accept: application/xml" -d '<user><id>bdog</id><name>bob</name></user>'
//   curl localhost:8080/users/sam -v

import (
	"errors"
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/julienschmidt/httprouter"
	"github.com/nstogner/ctxware/ezy/boltezy"
	"github.com/nstogner/ctxware/mdl/boltmdl"
	"github.com/nstogner/ctxware/mdl/entitymdl"
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

	// Perform db setup.
	db, err := bolt.Open("users-tmp.db", 0600, nil)
	if err != nil {
		panic(err)
	}

	def := boltmdl.Definition{
		DB:               db,
		EntityBucketPath: "/users/",
		IdParam:          "id",
		IdField:          "Id",
		EntityDef: entitymdl.Definition{
			Entity: User{},
			Validate: func(u interface{}) error {
				usr := u.(*User)
				if len(usr.Id) < 5 {
					return errors.New("user id must be at least 5 characters")
				}
				return nil
			},
		},
	}

	r := httprouter.New()
	r.GET("/users/:id", boltezy.Get(nil, def))
	r.POST("/users", boltezy.Post(nil, def))

	logrus.WithField("port", port).Info("starting server...")
	logrus.Fatal(http.ListenAndServe(":"+port, r))

}
