package httpadp

import (
	"net/http"

	"github.com/nstogner/ctxware/lib/httpctx"

	"golang.org/x/net/context"
)

// Adapter is used to call a httpctx.Handler where a http.handler is expected.
type Adapter struct {
	ctx     context.Context
	handler httpctx.Handler
}

func Adapt(h httpctx.Handler) Adapter {
	return Adapter{
		ctx:     context.Background(),
		handler: h,
	}
}

func (ca Adapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ca.handler.ServeHTTPContext(ca.ctx, w, r)
}
