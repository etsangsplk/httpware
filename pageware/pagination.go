/*
Package pageware provides middleware for pagination use cases. It handles
inspecting http query parameters and setting default values to start/limit
variables.
*/
package pageware

import (
	"net/http"
	"strconv"

	"github.com/nstogner/httpware"
	"golang.org/x/net/context"
)

var (
	// Defaults is a reasonable configuration.
	Defaults = Config{
		StartQuery:   "start",
		LimitQuery:   "limit",
		LimitDefault: 10,
	}
)

// Config is used to initialize a new instance of Middle.
type Config struct {
	// The string pulled from the URL for determining the record offset.
	StartQuery string
	// The string pulled from teh URL for determining the record limit.
	LimitQuery string
	// The limit that will be returned when none was provided in the request.
	LimitDefault int
}

// Middle is middleware that limits http requests.
type Middle struct {
	startQuery   string
	limitQuery   string
	limitDefault int
}

// New creates a new Middle instance.
func New(conf Config) *Middle {
	middle := Middle{
		startQuery:   conf.StartQuery,
		limitQuery:   conf.LimitQuery,
		limitDefault: conf.LimitDefault,
	}
	return &middle
}

// Page contains parsed pagination info.
type Page struct {
	Start int
	Limit int
}

// PageFromCtx retrieves the generated Page struct.
func PageFromCtx(ctx context.Context) Page {
	return ctx.Value(httpware.PageKey).(Page)
}

// Handle takes the next handler as an argument and wraps it in this middleware.
func (m *Middle) Handle(next httpware.Handler) httpware.Handler {
	return httpware.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		q := r.URL.Query()
		s := q.Get(m.startQuery)
		l := q.Get(m.limitQuery)
		page := Page{}
		var err error
		if s == "" {
			page.Start = 0
		} else {
			page.Start, err = strconv.Atoi(s)
			if err != nil {
				return httpware.NewErr("param 'start' must be an integer", http.StatusBadRequest)
			}
			if page.Start < 0 {
				return httpware.NewErr("param 'start' must not be negative", http.StatusBadRequest)
			}
		}
		if l == "" {
			page.Limit = m.limitDefault
		} else {
			page.Limit, err = strconv.Atoi(l)
			if err != nil {
				return httpware.NewErr("invalid query parameter", http.StatusBadRequest).WithField(m.limitQuery, "must not be an integer")
			}
			if page.Limit <= 0 {
				return httpware.NewErr("invalid query parameter", http.StatusBadRequest).WithField(m.limitQuery, "must not be greater than zero")
			}
		}

		return next.ServeHTTPCtx(context.WithValue(ctx, httpware.PageKey, page), w, r)
	})
}
