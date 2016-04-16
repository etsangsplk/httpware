package httpware

import (
	"encoding/json"
	"encoding/xml"
	"net/http"

	"github.com/nstogner/httpware/httpctx"
	"github.com/nstogner/httpware/httperr"

	"golang.org/x/net/context"
)

// It is generally a good thing to hide >500 error messages from
// clients, and show a predefined message, while logging the
// interesting stuff. This option defaults to true, to allow 500+
// code responses to contain errors, set this to false.
var Suppress500Messages = true

// handleHttpErrors returns the appropriate http response code based on the
// returned error. If the returned error is not nil and of type httperr.Err
// the specified status code is returned. Any other errors are treated as
// a 500 - Internal Server Error.
func handleHttpErrors(next httpctx.Handler) httpctx.Handler {
	return httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		err := next.ServeHTTPCtx(ctx, w, r)
		if err != nil {
			w.Header().Set("X-Content-Type-Options", "nosniff")

			respErr := httperr.Err{}

			if e, ok := err.(httperr.Err); ok {
				w.WriteHeader(e.StatusCode)
				respErr.StatusCode = e.StatusCode
				respErr.Fields = e.Fields
				if e.StatusCode >= 500 {
					if Suppress500Messages {
						respErr.Message = http.StatusText(respErr.StatusCode)
						respErr.Fields = nil
					} else {
						respErr.Message = e.Message
					}
				} else {
					respErr.Message = e.Message
				}
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				respErr.StatusCode = http.StatusInternalServerError
				if Suppress500Messages {
					respErr.Message = http.StatusText(http.StatusInternalServerError)
				} else {
					respErr.Message = err.Error()
				}
			}

			switch ContentTypeFromHeader(w.Header().Get("Accept")) {
			case JSON:
				json.NewEncoder(w).Encode(respErr)
			case XML:
				xml.NewEncoder(w).Encode(respErr)
			}
		}
		return err
	})
}
