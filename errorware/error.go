/*
Package errorware provides middleware for handling errors returned by other
httpctx.Handler functions.
*/
package errorware

import (
	"net/http"

	"github.com/nstogner/httpware/contentware"
	"github.com/nstogner/httpware/httpctx"
	"github.com/nstogner/httpware/httperr"

	"golang.org/x/net/context"
)

var (
	// Defaults is a reasonable configuration.
	Defaults = Config{
		// It is generally a good thing to hide >500 error messages from
		// clients, and show a predefined message, while logging the
		// interesting stuff.
		Suppress500Messages: true,
	}
)

// Config is used for initiating a new instance of the middleware.
type Config struct {
	// When true, the error messages for >500 http errors will not be sent in
	// responses.
	Suppress500Messages bool
}

// Middle inspects the error returned by downstream handlers. If it
// is of the type httperr.Err, it will return the status code that is contained
// within the Err struct, otherwise it will default to a
// 500 - "internal server error" for all non-nil error returns.
type Middle struct {
	conf Config
}

// New returns an instance of the middleware.
func New(conf Config) *Middle {
	return &Middle{conf}
}

// Contains indentifies this middleware for compositions.
func (m *Middle) Contains() []string { return []string{"github.com/nstogner/errorware"} }

// Requires indentifies what this middleware depends on (nothing).
func (m *Middle) Requires() []string { return []string{} }

// Handle takes the next handler as an argument and wraps it in this middleware.
func (m *Middle) Handle(next httpctx.Handler) httpctx.Handler {
	return httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		err := next.ServeHTTPCtx(ctx, w, r)
		if err != nil {
			w.Header().Set("X-Content-Type-Options", "nosniff")

			respErr := httperr.Err{}

			if e, ok := err.(httperr.Err); ok {
				w.WriteHeader(e.StatusCode)
				respErr.StatusCode = e.StatusCode
				if e.StatusCode >= 500 {
					if m.conf.Suppress500Messages {
						respErr.Message = http.StatusText(respErr.StatusCode)
					} else {
						respErr.Message = e.Message
					}
				} else {
					respErr.Message = e.Message
				}
				respErr.Fields = e.Fields
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				respErr.StatusCode = http.StatusInternalServerError
				if m.conf.Suppress500Messages {
					respErr.Message = http.StatusText(http.StatusInternalServerError)
				} else {
					respErr.Message = err.Error()
				}
			}

			rct := contentware.ResponseTypeFromCtx(ctx)
			// If the response content type has not already been parsed by upstream middleware
			// then it must be parsed now.
			if rct == nil {
				rct = contentware.GetContentMatch(w.Header().Get("Accept"), contentware.JSONOverXML)
			}
			rct.Encode(w, respErr)
		}
		return err
	})
}
