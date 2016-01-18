package routercontext

import (
	"bluesky/httpcontext"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"golang.org/x/net/context"
)

func AdaptHandler(h httpcontext.Handler) httprouter.Handle {
	return Adapt(h.ServeHTTPContext)
}

func Adapt(hf httpcontext.HandlerFunc) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := context.Background()
		paramsCtx := context.WithValue(ctx, httpcontext.ParamsKey, ps)
		hf.ServeHTTPContext(paramsCtx, w, r)
	}
}

func ParamsFromContext(ctx context.Context) httprouter.Params {
	return ctx.Value(httpcontext.ParamsKey).(httprouter.Params)
}
