/*
Package logware provides http middleware for logging requests and errors. It
is based on the httpware.Middleware interface.
*/
package logware

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/nstogner/httpware"
	"github.com/nstogner/httpware/httperr"
	"golang.org/x/net/context"
)

var (
	Defaults = Def{
		Logger:     logrus.New(),
		Headers:    []string{},
		Referer:    false,
		RemoteAddr: false,
	}
)

type Def struct {
	Logger     *logrus.Logger
	Headers    []string
	Referer    bool
	RemoteAddr bool
}

// ErrLogger logs each error that is returned by the downstream handler.
type Logger struct {
	def Def
}

func New(def Def) Logger {
	return Logger{def}
}

func (l Logger) Contains() []string { return []string{"logware.ErrLogger"} }
func (l Logger) Requires() []string { return []string{} }

func (l Logger) Handle(next httpware.Handler) httpware.Handler {
	return httpware.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		// Call downstream handlers.
		err := next.ServeHTTPContext(ctx, w, r)

		// Always log the method and path.
		entry := l.def.Logger.WithFields(logrus.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
		})

		// Conditional logging...
		if l.def.Referer {
			entry = entry.WithField("referer", r.Referer())
		}
		if l.def.RemoteAddr {
			entry = entry.WithField("remoteAddr", r.RemoteAddr)
		}
		for _, h := range l.def.Headers {
			entry = entry.WithField(h, r.Header.Get(h))
		}

		// Add any errors to the log entry.
		logAsError := false
		if err != nil {
			if httpErr, ok := err.(httperr.Err); ok {
				entry.WithField("error", httpErr)
				if httpErr.StatusCode >= 500 {
					logAsError = true
				}
			} else {
				entry.WithField("error",
					map[string]interface{}{
						"statusCode": http.StatusInternalServerError,
						"message":    err,
					},
				)
				logAsError = true
			}
		}

		// Log with the right level and pass on the error.
		if logAsError {
			entry.Error("failed to serve request")
		} else {
			entry.Info("served request")
		}
		return err
	})
}
