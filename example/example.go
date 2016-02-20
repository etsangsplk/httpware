package main

import (
	"net/http"

	"golang.org/x/net/context"

	"github.com/nstogner/httpware"
	"github.com/nstogner/httpware/contentware"
	"github.com/nstogner/httpware/errorware"
	"github.com/nstogner/httpware/logware"
)

func main() {
	// MustCompose chains together middleware. It will panic if middleware
	// dependencies are not met.
	m := httpware.MustCompose(
		errorware.New(),
		logware.New(logware.Defaults),
		contentware.New(contentware.Defaults),
	)
	//logrus.SetFormatter(&logrus.JSONFormatter{})

	http.ListenAndServe("localhost:8080", m.ThenFunc(handle))
}

func handle(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	resp := &struct {
		Greeting string `json:"greeting", "xml": greeting`
	}{"Hi there!"}

	// middleware passes data via the context variable.
	t := contentware.ResponseTypeFromCtx(ctx)
	// t is the content type that was set by the contentware package. In this case
	// The middleware took care of determining the type by inspecting the 'Accept'
	// header.
	t.MarshalWrite(w, resp)
	return nil
}
