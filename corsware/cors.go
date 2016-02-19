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
	Defaults = Def{
		AllowOrigin:      "*",
		AllowCredentials: false,
		ExposeHeaders:    []string{},
	}
)

// Def (Definition) defines the settings used by the CORS middlware handler.
// The corsware.Default definition should work in most cases.
type Def struct {
	// Header: Access-Control-Allow-Origin (needed for basic cors support)
	AllowOrigin string
	// Header: Access-Control-Allow-Credentials (for allowing cookies)
	AllowCredentials bool
	// Header: Access-Control-Expose-Headers (for allowing cookies)
	ExposeHeaders []string
}

type Ware struct {
	allowOrigin         string
	allowCredentials    string
	exposeHeaders       string
	shouldExposeHeaders bool
}

func New(def Def) Ware {
	return Ware{
		allowOrigin:         def.AllowOrigin,
		allowCredentials:    strconv.FormatBool(def.AllowCredentials),
		exposeHeaders:       strings.Join(def.ExposeHeaders, ", "),
		shouldExposeHeaders: (len(def.ExposeHeaders) > 0),
	}
}

func (w Ware) Contains() []string { return []string{"corsware.Ware"} }
func (w Ware) Requires() []string { return []string{} }

func (ware Ware) Handle(next httpware.Handler) httpware.Handler {
	return httpware.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		w.Header().Set("Access-Control-Allow-Origin", ware.allowOrigin)
		w.Header().Set("Access-Control-Allow-Credentials", ware.allowCredentials)
		if ware.shouldExposeHeaders {
			w.Header().Set("Access-Control-Expose-Headers", ware.exposeHeaders)
		}
		return next.ServeHTTPContext(ctx, w, r)
	})
}
