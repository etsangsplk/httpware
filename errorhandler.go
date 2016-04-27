package httpware

import (
	"encoding/json"
	"encoding/xml"
	"net/http"

	"golang.org/x/net/context"
)

var (
	// DefaultErrHandlerConfig is a reasonable configuration.
	DefaultErrHandlerConfig = ErrHandlerConfig{
		Suppress500Messages: false,
	}
	// DefaultErrHandler uses DefaultErrHandlerConfig.
	DefaultErrHandler = NewErrHandler(DefaultErrHandlerConfig)
)

// ErrHandlerConfig is used in NewErrHandler.
type ErrHandlerConfig struct {
	// Suppress500Messages hides >500 specific error messages from responses to
	// clients and shows a predefined message.
	// To allow >500 code responses to contain errors, set this to false.
	Suppress500Messages bool
}

// ErrHandler is an implementation of Errware. It handles any errors that are
// returned by upstream middleware by generating the appropriate http response.
// If the returned error is not nil and of type httpware.Err
// the specified status code is returned. Any other errors are treated as
// a 500 - Internal Server Error.
type ErrHandler struct {
	conf ErrHandlerConfig
}

// NewErrHandler returns a new instance of ErrHandler.
func NewErrHandler(conf ErrHandlerConfig) *ErrHandler {
	return &ErrHandler{
		conf: conf,
	}
}

// HandleErr handles non-nil error return values by upstream middleware.
func (h *ErrHandler) HandleErr(next Handler) Handler {
	return HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		err := next.ServeHTTPCtx(ctx, w, r)
		if err != nil {
			w.Header().Set("X-Content-Type-Options", "nosniff")

			respErr := Err{}

			if e, ok := err.(Err); ok {
				w.WriteHeader(e.StatusCode)
				respErr.StatusCode = e.StatusCode
				respErr.Fields = e.Fields
				if e.StatusCode >= 500 {
					if h.conf.Suppress500Messages {
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
				if h.conf.Suppress500Messages {
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
