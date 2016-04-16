package httpware

import (
	"encoding/json"
	"encoding/xml"
	"net/http"

	"golang.org/x/net/context"
)

// Suppress500Messages hides >500 specific error messages from responses to
// clients and shows a predefined message. This option defaults to true.
// To allow >500 code responses to contain errors, set this to false.
var Suppress500Messages = true

// Err is a struct which carries the information of an error which occurs in
// a http handler.
type Err struct {
	StatusCode int                    `json:"-" xml:"-"`
	Message    string                 `json:"message" xml:"message"`
	Fields     map[string]interface{} `json:"fields,omitempty" xml:"fields,omitempty"`
}

// NewErr creates an bare minimum http error.
func NewErr(msg string, status int) Err {
	return Err{
		StatusCode: status,
		Message:    msg,
		Fields:     make(map[string]interface{}),
	}
}

// WithField returns a new Err with the given key-value pair included
// in the 'Fields' field.
func (err Err) WithField(name string, value interface{}) Err {
	err.Fields[name] = value
	return Err{
		StatusCode: err.StatusCode,
		Message:    err.Message,
		Fields:     err.Fields,
	}
}

// The Error() method allows the Err struct to satisfy the standard error
// interface.
func (err Err) Error() string {
	return err.Message
}

// handleHTTPErrors returns the appropriate http response code based on the
// returned error. If the returned error is not nil and of type httpware.Err
// the specified status code is returned. Any other errors are treated as
// a 500 - Internal Server Error.
func handleHTTPErrors(next Handler) Handler {
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
