package routeradp

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/nstogner/ctxware"
	"golang.org/x/net/context"
)

func Adapt(h ctxware.Handler) httprouter.Handle {
	return AdaptFunc(h.ServeHTTPContext)
}

func AdaptFunc(hf ctxware.HandlerFunc) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := context.Background()
		paramsCtx := context.WithValue(ctx, ctxware.RouterParamsKey, ps)
		hf.ServeHTTPContext(paramsCtx, w, r)
	}
}

func ParamsFromCtx(ctx context.Context) httprouter.Params {
	return ctx.Value(ctxware.RouterParamsKey).(httprouter.Params)
}
