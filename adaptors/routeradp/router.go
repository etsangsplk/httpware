package routeradp

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/nstogner/ctxware"
	"golang.org/x/net/context"
)

// Adapt calls the AdaptFunc function.
func Adapt(h ctxware.Handler) httprouter.Handle {
	return AdaptFunc(h.ServeHTTPContext)
}

// AdaptFunc can be the starting point for ctxware.Handler implementations. It
// creates a new background context and invokes the ServeHTTPContext function.
func AdaptFunc(hf ctxware.HandlerFunc) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := context.Background()
		paramsCtx := context.WithValue(ctx, ctxware.RouterParamsKey, ps)
		hf.ServeHTTPContext(paramsCtx, w, r)
	}
}

// ParamsFromCtx retreives the httprouter.Params that are set by httprouter.
func ParamsFromCtx(ctx context.Context) httprouter.Params {
	return ctx.Value(ctxware.RouterParamsKey).(httprouter.Params)
}
