package errormdl

import (
	"net/http"

	"github.com/nstogner/ctxware/lib/httpctx"
	"github.com/nstogner/ctxware/lib/httperr"
	"github.com/nstogner/ctxware/mdl/contentmdl"

	"golang.org/x/net/context"
)

func Handle(next httpctx.Handler, catchAll bool) httpctx.Handler {
	return httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()

			if err == nil {
				return
			}

			if httpErr, ok := err.(httperr.Err); ok {
				w.Header().Set("X-Content-Type-Options", "nosniff")
				w.WriteHeader(httpErr.StatusCode)

				if httpErr.StatusCode >= 500 {
					httpErr.Message = http.StatusText(httpErr.StatusCode)
				}
				rct := contentmdl.ResponseTypeFromCtx(ctx)
				// If the response content type has not already been parsed by upstream middleware
				// then it must be parsed now.
				if rct == nil {
					rct = contentmdl.GetContentMatch(w.Header().Get("Accept"), contentmdl.JsonAndXml)
				}

				rct.MarshalWrite(w, httpErr)
			} else {
				if catchAll {
					// If a regular error was returned, resort to internal server error.
					http.Error(w, "internal server error", http.StatusInternalServerError)
				} else {
					// Explode.
					panic(err)
				}
			}
		}()

		next.ServeHTTPContext(ctx, w, r)
	})
}
