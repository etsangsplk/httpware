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
		Logger: logrus.New(),
	}
)

// Config is used to initialize a new instance of this middleware.
type Config struct {
	Logger         *logrus.Logger
	Headers        []string
	Referer        bool
	RemoteAddr     bool
	IgnoreUnder400 bool
	Ignore4XX      bool
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
		// Log panics.
		defer func() {
			if rcv := recover(); rcv != nil {
				m.conf.Logger.WithField("error", rcv).Error("handler panic detected")
				// Pass on the panic.
				panic(rcv)
			}
		}()

		// Call downstream handlers.
		err := next.ServeHTTPCtx(ctx, w, r)

		// Always log the method and path.
		entry := m.conf.Logger.WithFields(logrus.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
		})

		// Conditional logging...
		if m.conf.Referer {
			entry = entry.WithField("referrer", r.Referer())
		}
		if m.conf.RemoteAddr {
			entry = entry.WithField("remoteAddr", r.RemoteAddr)
		}
		for _, h := range m.conf.Headers {
			entry = entry.WithField(h, r.Header.Get(h))
		}

		// Add any errors to the log entry.
		statusCode := 0
		if err != nil {
			if httpErr, ok := err.(httpware.Err); ok {
				entry = entry.WithFields(logrus.Fields{
					"statusCode": httpErr.StatusCode,
					"message":    httpErr.Message,
				})
				entry = entry.WithFields(httpErr.Fields)
				statusCode = httpErr.StatusCode
			} else {
				entry = entry.WithField("error",
					map[string]interface{}{
						"statusCode": http.StatusInternalServerError,
						"message":    err,
					},
				)
				statusCode = http.StatusInternalServerError
			}
		}

		// Log with the right level and pass on the error.
		if statusCode >= 500 {
			entry.Error("server error")
		} else {
			if statusCode >= 400 {
				if !m.conf.Ignore4XX {
					entry.Info("client error")
				}
			} else if !m.conf.IgnoreUnder400 {
				entry.Info("successful request")
			}
		}
		return err
	})
}
