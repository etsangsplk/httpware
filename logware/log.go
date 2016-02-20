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
	Defaults = Config{
		Logger:     logrus.New(),
		Headers:    []string{},
		Referer:    false,
		RemoteAddr: false,
	}
)

type Config struct {
	Logger     *logrus.Logger
	Headers    []string
	Referer    bool
	RemoteAddr bool
}

// logware.Middle logs each error that is returned by the downstream handler.
type Middle struct {
	conf Config
}

func New(conf Config) Middle {
	return Middle{conf}
}

func (m Middle) Contains() []string { return []string{"logware"} }
func (m Middle) Requires() []string { return []string{} }

func (m Middle) Handle(next httpware.Handler) httpware.Handler {
	return httpware.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		// Call downstream handlers.
		err := next.ServeHTTPContext(ctx, w, r)

		// Always log the method and path.
		entry := m.conf.Logger.WithFields(logrus.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
		})

		// Conditional logging...
		if m.conf.Referer {
			entry = entry.WithField("referer", r.Referer())
		}
		if m.conf.RemoteAddr {
			entry = entry.WithField("remoteAddr", r.RemoteAddr)
		}
		for _, h := range m.conf.Headers {
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
