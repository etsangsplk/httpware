package routeradp

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/nstogner/ctxware/lib/httpctx"
	"golang.org/x/net/context"
)

func Adapt(h httpctx.Handler) httprouter.Handle {
	return AdaptFunc(h.ServeHTTPContext)
}

func AdaptFunc(hf httpctx.HandlerFunc) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := context.Background()
		// Standardize params to be a map so that the same can be done w/ other routers.
		params := make(map[string]string)
		for _, p := range ps {
			params[p.Key] = p.Value
		}
		paramsCtx := context.WithValue(ctx, httpctx.ParamsKey, params)
		hf.ServeHTTPContext(paramsCtx, w, r)
	}
}

func ParamsFromCtx(ctx context.Context) map[string]string {
	params := ctx.Value(httpctx.ParamsKey)
	if params == nil {
		return nil
	}
	return params.(map[string]string)
}
