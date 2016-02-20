/*
Package corsware provides middleware for setting the necessary http headers to
allow CORS requests. It implements the httpware.Middleware interface for
easy composition with other middleware.
*/
package corsware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/nstogner/httpware"
	"golang.org/x/net/context"
)

var (
	Defaults = Config{
		AllowOrigin:      "*",
		AllowCredentials: false,
		ExposeHeaders:    []string{},
	}
)

// Config defines the settings used by the CORS middlware handler.
// The corsware.Defaults definition should work in most cases.
type Config struct {
	// Header: Access-Control-Allow-Origin (needed for basic cors support)
	AllowOrigin string
	// Header: Access-Control-Allow-Credentials (for allowing cookies)
	AllowCredentials bool
	// Header: Access-Control-Expose-Headers (for allowing cookies)
	ExposeHeaders []string
}

type Middle struct {
	allowOrigin         string
	allowCredentials    string
	exposeHeaders       string
	shouldExposeHeaders bool
}

func New(conf Config) *Middle {
	return &Middle{
		allowOrigin:         conf.AllowOrigin,
		allowCredentials:    strconv.FormatBool(conf.AllowCredentials),
		exposeHeaders:       strings.Join(conf.ExposeHeaders, ", "),
		shouldExposeHeaders: (len(conf.ExposeHeaders) > 0),
	}
}

func (m *Middle) Contains() []string { return []string{"github.com/nstogner/corsware"} }
func (m *Middle) Requires() []string { return []string{} }

func (m *Middle) Handle(next httpware.Handler) httpware.Handler {
	return httpware.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		w.Header().Set("Access-Control-Allow-Origin", m.allowOrigin)
		w.Header().Set("Access-Control-Allow-Credentials", m.allowCredentials)
		if m.shouldExposeHeaders {
			w.Header().Set("Access-Control-Expose-Headers", m.exposeHeaders)
		}
		return next.ServeHTTPContext(ctx, w, r)
	})
}
