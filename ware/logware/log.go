package logware

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/nstogner/ctxware"
	"github.com/nstogner/ctxware/lib/httperr"
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

	headers    logFunc
	referer    logFunc
	remoteaddr logFunc
}

type logFunc func(e *logrus.Entry, r *http.Request) *logrus.Entry

func (def *Def) predefine() {
	if def.Referer {
		def.referer = func(e *logrus.Entry, r *http.Request) *logrus.Entry { return e.WithField("referer", r.Referer()) }
	} else {
		def.referer = func(e *logrus.Entry, r *http.Request) *logrus.Entry { return e }
	}
	if def.RemoteAddr {
		def.remoteaddr = func(e *logrus.Entry, r *http.Request) *logrus.Entry { return e.WithField("remoteAddr", r.RemoteAddr) }
	} else {
		def.remoteaddr = func(e *logrus.Entry, r *http.Request) *logrus.Entry { return e }
	}
	def.headers = func(e *logrus.Entry, r *http.Request) *logrus.Entry {
		for _, h := range def.Headers {
			e = e.WithField(h, r.Header.Get(h))
		}
		return e
	}
}

type ReqLogger struct {
	def Def
}

func NewReqLogger(def Def) ReqLogger {
	// Predefine functions so that if-then logic is not needed in handler.
	def.predefine()
	return ReqLogger{def}
}

func (rl ReqLogger) Contains() []string {
	return []string{"logware.ReqLogger"}
}

func (rl ReqLogger) Requires() []string {
	return []string{}
}

func (rl ReqLogger) Handle(next ctxware.Handler) ctxware.Handler {
	return ctxware.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		wf := rl.def.Logger.WithFields(logrus.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
		})
		wf = rl.def.referer(wf, r)
		wf = rl.def.remoteaddr(wf, r)
		wf = rl.def.headers(wf, r)
		wf.Info("serving request...")
		return next.ServeHTTPContext(ctx, w, r)
	})
}

type ErrLogger struct {
	def Def
}

func NewErrLogger(def Def) ErrLogger {
	// Predefine functions so that if-then logic is not needed in handler.
	def.predefine()
	return ErrLogger{def}
}

func (el ErrLogger) Contains() []string {
	return []string{"logware.ErrLogger"}
}

func (el ErrLogger) Requires() []string {
	return []string{}
}

func (el ErrLogger) Handle(next ctxware.Handler) ctxware.Handler {
	return ctxware.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		if err := next.ServeHTTPContext(ctx, w, r); err != nil {
			wf := el.def.Logger.WithFields(logrus.Fields{
				"method": r.Method,
				"path":   r.URL.Path,
			})
			wf = el.def.referer(wf, r)
			wf = el.def.remoteaddr(wf, r)
			wf = el.def.headers(wf, r)
			if httpErr, ok := err.(httperr.Err); ok {
				wf.WithField("error", httpErr).Info("request failed")
				// Pass the http error along.
				return httpErr
			} else {
				wf.WithField("error",
					map[string]interface{}{
						"statusCode": http.StatusInternalServerError,
						"message":    err,
					},
				).Info("request failed")
				// Pass the error along.
				return err
			}
		}
		return nil
	})
}
