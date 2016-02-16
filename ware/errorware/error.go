/*
Package errorware provides middleware for handling errors in other
ctxware.Handler implementations.
*/
package errorware

import (
	"net/http"

	"github.com/nstogner/ctxware"
	"github.com/nstogner/ctxware/lib/httperr"
	"github.com/nstogner/ctxware/ware/contentware"

	"golang.org/x/net/context"
)

type Ware struct {
}

func New() Ware {
	return Ware{}
}

func (w Ware) Contains() []string {
	return []string{"errorware.Ware"}
}

func (w Ware) Requires() []string {
	return []string{}
}

func (w Ware) Handle(next ctxware.Handler) ctxware.Handler {
	return ctxware.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
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
				rct = contentware.GetContentMatch(w.Header().Get("Accept"), contentware.JsonAndXml)
			}
			rct.MarshalWrite(w, respErr)
		}
		return nil
	})
}
