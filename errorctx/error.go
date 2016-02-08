package errorctx

import (
	"net/http"

	"github.com/nstogner/contextware/contentctx"
	"github.com/nstogner/contextware/httpctx"
	"github.com/nstogner/contextware/httperr"
	"golang.org/x/net/context"
)

func Handle(next httpctx.Handler) httpctx.Handler {
	return httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		err := next.ServeHTTPContext(ctx, w, r)
		if err == nil {
			return nil
		}

		if httpErr, ok := err.(httperr.Err); ok {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.WriteHeader(httpErr.StatusCode)

			rct := contentctx.ResponseTypeFromCtx(ctx)
			// If the response content type has not already been parsed by upstream middleware
			// then it must be parsed now.
			if rct == nil {
				rct = contentctx.GetContentMatch(w.Header().Get("Accept"), contentctx.JsonAndXml)
			}

			rct.MarshalWrite(w, httpErr)
			return nil
		}

		// If a regular error was returned, resort to internal server error.
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	})
}