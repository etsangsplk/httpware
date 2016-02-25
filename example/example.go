/*
This example is intended to show a very simple use case for middleware in the
httpware repository. Once the server is running try POSTing a user:

curl -v localhost:8080 -d '{"id":"bob", "email":"bob@email.com"}'
*/
package main

import (
	"net/http"

	"golang.org/x/net/context"

	"github.com/nstogner/httpware"
	"github.com/nstogner/httpware/contentware"
	"github.com/nstogner/httpware/errorware"
	"github.com/nstogner/httpware/httperr"
	"github.com/nstogner/httpware/logware"
)

func main() {
	// MustCompose chains together middleware. It will panic if middleware
	// dependencies are not met.
	m := httpware.MustCompose(
		contentware.New(contentware.Defaults),
		errorware.New(errorware.Defaults),
		logware.New(logware.Defaults),
	)

	http.ListenAndServe("localhost:8080", m.ThenFunc(handle))
}

type user struct {
	ID    string `json:"id" xml:"id"`
	Email string `json:"email" xml:"email"`
}

// handle is meant to demonstrate a POST or PUT endpoint.
func handle(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	u := &user{}
	rqt := contentware.ResponseTypeFromCtx(ctx)
	// Decode from JSON or XML based on the 'Content-Type' header.
	if err := rqt.Decode(r.Body, u); err != nil {
		return httperr.New("could not parse body: "+err.Error(), http.StatusBadRequest)
	}

	// Store u in DB here... //

	rst := contentware.ResponseTypeFromCtx(ctx)
	// Encode to JSON or XML based on the 'Accept' header.
	rst.Encode(w, u)
	return nil
}
