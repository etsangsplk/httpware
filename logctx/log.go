package logctx

import (
	"errors"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/nstogner/ctxware/httpctx"
	"github.com/nstogner/ctxware/httperr"
	"golang.org/x/net/context"
)

func Requests(next httpctx.Handler) httpctx.Handler {
	return httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		logrus.WithFields(logrus.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
		}).Info("serving request...")
		next.ServeHTTPContext(ctx, w, r)
	})
}

func Errors(next httpctx.Handler) httpctx.Handler {
	return httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err == nil {
				return
			}

			if httpErr, ok := err.(httperr.Err); ok {
				logrus.WithFields(logrus.Fields{
					"method": r.Method,
					"error":  httpErr,
				}).Info("request failed")
				// Propogate the http error along.
				httperr.Return(httpErr)
			} else {
				msg, ok := err.(error)
				if !ok {
					msg = errors.New("unidentified error")
				}
				logrus.WithFields(logrus.Fields{
					"method": r.Method,
					"error": map[string]interface{}{
						"statusCode": http.StatusInternalServerError,
						"message":    msg,
					},
				}).Info("request failed")
				// Propogate the error along.
				panic(err)
			}
		}()

		next.ServeHTTPContext(ctx, w, r)
	})
}
