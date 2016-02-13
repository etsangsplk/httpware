package errormdl

import (
	"net/http"

	"github.com/nstogner/ctxware/lib/httpctx"
	"github.com/nstogner/ctxware/lib/httperr"
	"github.com/nstogner/ctxware/mdl/contentmdl"

	"golang.org/x/net/context"
)

func Handle(next httpctx.Handler, catchAll bool) httpctx.Handler {
	return httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
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

			rct := contentmdl.ResponseTypeFromCtx(ctx)
			// If the response content type has not already been parsed by upstream middleware
			// then it must be parsed now.
			if rct == nil {
				rct = contentmdl.GetContentMatch(w.Header().Get("Accept"), contentmdl.JsonAndXml)
			}
			rct.MarshalWrite(w, respErr)
		}
		return nil
	})
}
