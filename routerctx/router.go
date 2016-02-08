package routerctx

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/nstogner/contextware/httpctx"
	"golang.org/x/net/context"
)

func Adapt(h httpctx.Handler) httprouter.Handle {
	return AdaptFunc(h.ServeHTTPContext)
}

func AdaptFunc(hf httpctx.HandlerFunc) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := context.Background()
		paramsCtx := context.WithValue(ctx, httpctx.ParamsKey, ps)
		hf.ServeHTTPContext(paramsCtx, w, r)
	}
}

func ParamsFromContext(ctx context.Context) httprouter.Params {
	return ctx.Value(httpctx.ParamsKey).(httprouter.Params)
}
