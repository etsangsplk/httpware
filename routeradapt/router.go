/*
Package routeradapt provides a set of functions which enable using
httpctx.Handler implementations with the httprouter package.
*/
package routeradapt

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/nstogner/httpware"
	"github.com/nstogner/httpware/httpctx"
	"golang.org/x/net/context"
)

// Adapt calls the AdaptFunc function.
func Adapt(h httpctx.Handler) httprouter.Handle {
	return AdaptFunc(h.ServeHTTPCtx)
}

// AdaptFunc can be the starting point for httpctx.Handler implementations. It
// creates a new background context and invokes the ServeHTTPCtx function.
func AdaptFunc(hf httpctx.HandlerFunc) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := context.Background()
		paramsCtx := context.WithValue(ctx, httpware.RouterParamsKey, ps)
		hf.ServeHTTPCtx(paramsCtx, w, r)
	}
}

// ParamsFromCtx retrieves the httprouter.Params that are set by httprouter.
func ParamsFromCtx(ctx context.Context) httprouter.Params {
	return ctx.Value(httpware.RouterParamsKey).(httprouter.Params)
}
