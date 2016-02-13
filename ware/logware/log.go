package logware

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/nstogner/ctxware"
	"github.com/nstogner/ctxware/lib/httperr"
	"golang.org/x/net/context"
)

type ReqLogger struct {
}

func NewReqLogger() ReqLogger {
	return ReqLogger{}
}

func (rl ReqLogger) Name() string {
	return "logware.ReqLogger"
}

func (rl ReqLogger) Dependencies() []string {
	return []string{}
}

func (rl ReqLogger) Handle(next ctxware.Handler) ctxware.Handler {
	return ctxware.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		logrus.WithFields(logrus.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
		}).Info("serving request...")
		return next.ServeHTTPContext(ctx, w, r)
	})
}

type ErrLogger struct {
}

func NewErrLogger() ErrLogger {
	return ErrLogger{}
}

func (el ErrLogger) Name() string {
	return "logware.ErrLogger"
}

func (el ErrLogger) Dependencies() []string {
	return []string{}
}

func (el ErrLogger) Handle(next ctxware.Handler) ctxware.Handler {
	return ctxware.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		if err := next.ServeHTTPContext(ctx, w, r); err != nil {
			if httpErr, ok := err.(httperr.Err); ok {
				logrus.WithFields(logrus.Fields{
					"method": r.Method,
					"error":  httpErr,
				}).Info("request failed")
				// Pass the http error along.
				return httpErr
			} else {
				logrus.WithFields(logrus.Fields{
					"method": r.Method,
					"error": map[string]interface{}{
						"statusCode": http.StatusInternalServerError,
						"message":    err,
					},
				}).Info("request failed")
				// Pass the error along.
				return err
			}
		}
		return nil
	})
}
