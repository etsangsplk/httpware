/*
This example is intended to show a very simple use case for middleware in the
httpware repository. Once the server is running try POSTing a user:

curl -v localhost:8080 -d '{"id":"bob", "email":"bob@email.com"}'
*/
package main

import (
	"errors"
	"net/http"

	"golang.org/x/net/context"

	"github.com/nstogner/httpware"
	"github.com/nstogner/httpware/contentware"
	"github.com/nstogner/httpware/logware"
)

func main() {
	// Compose chains together middleware.
	m := httpware.Compose(
		httpware.DefaultErrHandler,
		contentware.New(contentware.Defaults),
		logware.New(logware.Defaults),
	)

	http.ListenAndServe("localhost:8080", m.ThenFunc(handle))
}

// handle is meant to demonstrate a POST or PUT endpoint.
func handle(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	u := &User{}
	rqt := contentware.ResponseTypeFromCtx(ctx)
	// Decode from JSON or XML based on the 'Content-Type' header.
	if err := rqt.Decode(r.Body, u); err != nil {
		return httpware.NewErr("could not parse body: "+err.Error(), http.StatusBadRequest)
	}

	if err := u.validate(); err != nil {
		return httpware.NewErr("invalid entity", http.StatusBadRequest).WithField("invalid", err.Error())
	}

	// Store user to db here.

	rst := contentware.ResponseTypeFromCtx(ctx)
	// Write the user back in the response as JSON or XML based on the
	// 'Accept' header.
	rst.Encode(w, u)
	return nil
}

type User struct {
	ID    string `json:"id" xml:"id"`
	Email string `json:"email" xml:"email"`
}

func (u *User) validate() error {
	if u.ID == "" {
		return errors.New("field 'id' must not be empty")
	}
	return nil
}
