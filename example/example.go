package main

import (
	"net/http"

	"golang.org/x/net/context"

	"github.com/nstogner/ctxware"
	"github.com/nstogner/ctxware/ware/contentware"
	"github.com/nstogner/ctxware/ware/errorware"
	"github.com/nstogner/ctxware/ware/logware"
)

func main() {
	// MustCompose chains together middleware. It will panic if middleware
	// dependencies are not met.
	midware := ctxware.MustCompose(
		errorware.New(),
		logware.NewErrLogger(),
		logware.NewReqLogger(),
		contentware.NewRespType(contentware.JsonAndXml),
	)

	http.ListenAndServe("localhost:8080", midware.ThenFunc(handle))
}

func handle(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	resp := &struct {
		Greeting string `json:"greeting", "xml": greeting`
	}{"Hi there!"}
	// Use the content type that was specified by the 'Accept' header.
	t := contentware.ResponseTypeFromCtx(ctx)
	t.MarshalWrite(w, resp)
	return nil
}
