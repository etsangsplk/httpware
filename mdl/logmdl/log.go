package logmdl

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/nstogner/ctxware/lib/httpctx"
	"github.com/nstogner/ctxware/lib/httperr"
	"golang.org/x/net/context"
)

func Requests(next httpctx.Handler) httpctx.Handler {
	return httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		logrus.WithFields(logrus.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
		}).Info("serving request...")
		return next.ServeHTTPContext(ctx, w, r)
	})
}

func Errors(next httpctx.Handler) httpctx.Handler {
	return httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
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
