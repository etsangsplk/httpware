package errorctx

import (
	"net/http"

	"github.com/nstogner/netmiddle/contentctx"
	"github.com/nstogner/netmiddle/httpctx"
	"github.com/nstogner/netmiddle/httperr"
	"golang.org/x/net/context"
)

func Handle(next httpctx.Handler, secret interface{}) httpctx.Handler {
	return httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		err := next.ServeHTTPContext(ctx, w, r)
		if err == nil {
			return nil
		}

		if httpErr, ok := err.(httperr.Err); ok {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.WriteHeader(httpErr.StatusCode)

			rct := contentctx.RespContentTypeFromContext(ctx)
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
