/*
Package logware provides http middleware for logging requests and errors. It
is based on the httpware.Middleware interface.
*/
package logware

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/nstogner/httpware"
	"golang.org/x/net/context"
)

var (
	// Defaults is a reasonable configuration that should work for most cases.
	Defaults = Config{
		Logger:     logrus.New(),
		Headers:    []string{},
		Referer:    false,
		RemoteAddr: false,
		Successes:  true,
		Failures:   true,
		Panics:     true,
	}
)

// Config is used to initialize a new instance of this middleware.
type Config struct {
	Logger     *logrus.Logger
	Headers    []string
	Referer    bool
	RemoteAddr bool
	// Should <500 http responses be logged?
	Successes bool
	// Should 500+ http responses be logged?
	Failures bool
	Panics   bool
}

// Middle logs http responses and any errors returned by the downstream
// handler.
type Middle struct {
	conf Config
}

// New returns a new logware.Middle instance.
func New(conf Config) *Middle {
	return &Middle{conf}
}

// Handle takes the next handler as an argument and wraps it in this middleware.
func (m *Middle) Handle(next httpware.Handler) httpware.Handler {
	return httpware.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		// Should panics be logged?
		if m.conf.Panics {
			defer func() {
				if rcv := recover(); rcv != nil {
					m.conf.Logger.WithField("error", rcv).Error("handler panic detected")
					// Pass on the panic.
					panic(rcv)
				}
			}()
		}

		// Call downstream handlers.
		err := next.ServeHTTPCtx(ctx, w, r)

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
			if httpErr, ok := err.(httpware.Err); ok {
				entry = entry.WithFields(logrus.Fields{
					"statusCode": httpErr.StatusCode,
					"message":    httpErr.Message,
				})
				entry = entry.WithFields(httpErr.Fields)
				if httpErr.StatusCode >= 500 {
					logAsError = true
				}
			} else {
				entry = entry.WithField("error",
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
			if m.conf.Failures {
				entry.Error("failed to serve request")
			}
		} else {
			if m.conf.Successes {
				entry.Info("served request")
			}
		}
		return err
	})
}
