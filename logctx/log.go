package logctx

import (
	"log"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/nstogner/contextware/httpctx"
	"github.com/nstogner/contextware/httperr"
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
				log.Println("httperr")
				logrus.WithFields(logrus.Fields{
					"method": r.Method,
					"error":  httpErr,
				}).Info("request failed")
				// Propogate the http error along.
				httperr.Return(httpErr)
			} else {
				logrus.WithFields(logrus.Fields{
					"method": r.Method,
					"error": httperr.Err{
						StatusCode: http.StatusInternalServerError,
						Message:    "internal server error",
					},
				}).Info("request failed")
				// Propogate the error along.
				panic(err)
			}
		}()

		next.ServeHTTPContext(ctx, w, r)
	})
}
