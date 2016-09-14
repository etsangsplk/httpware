/*
Package routeradapt provides a set of functions which enable using
httpware.Handler implementations with the httprouter package.
*/
package routeradapt

import (
	"context"
	"net/http"

	"github.com/nstogner/httpware"
	"github.com/julienschmidt/httprouter"
)

// Adapt calls the AdaptFunc function.
func Adapt(h httpware.Handler) httprouter.Handle {
	return AdaptFunc(h.ServeHTTPCtx)
}

// AdaptFunc can be the starting point for httpware.Handler implementations. It
// creates a new background context and invokes the ServeHTTPCtx function.
func AdaptFunc(hf httpware.HandlerFunc) httprouter.Handle {
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
