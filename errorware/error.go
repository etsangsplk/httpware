/*
Package errorware provides middleware for handling errors in other
httpware.Handler implementations.
*/
package errorware

import (
	"net/http"

	"github.com/nstogner/httpware"
	"github.com/nstogner/httpware/contentware"
	"github.com/nstogner/httpware/httperr"

	"golang.org/x/net/context"
)

type Ware struct {
}

func New() Ware {
	return Ware{}
}

func (w Ware) Contains() []string { return []string{"errorware.Ware"} }
func (w Ware) Requires() []string { return []string{} }

func (w Ware) Handle(next httpware.Handler) httpware.Handler {
	return httpware.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		if err := next.ServeHTTPContext(ctx, w, r); err != nil {
			w.Header().Set("X-Content-Type-Options", "nosniff")

			respErr := httperr.Err{}

			if e, ok := err.(httperr.Err); ok {
				w.WriteHeader(e.StatusCode)
				respErr.StatusCode = e.StatusCode
				if e.StatusCode >= 500 {
					respErr.Message = http.StatusText(respErr.StatusCode)
				}
				respErr.Fields = e.Fields
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				respErr.StatusCode = http.StatusInternalServerError
				respErr.Message = err.Error()
			}

			rct := contentware.ResponseTypeFromCtx(ctx)
			// If the response content type has not already been parsed by upstream middleware
			// then it must be parsed now.
			if rct == nil {
				rct = contentware.GetContentMatch(w.Header().Get("Accept"), contentware.JsonOverXml)
			}
			rct.MarshalWrite(w, respErr)
		}
		return nil
	})
}
