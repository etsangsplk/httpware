package entityctx

import (
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/nstogner/ctxware/contentctx"
	"github.com/nstogner/ctxware/httpctx"
	"github.com/nstogner/ctxware/httperr"
	"golang.org/x/net/context"
)

var maxInt64 = int64(^uint64(0) >> 1)

type Definition struct {
	Entity      interface{}
	Validate    Validator
	MaxByteSize int64

	reflectedType reflect.Type
}

type Validator func(interface{}) error

func (d *Definition) Inspect() {
	d.reflectedType = reflect.TypeOf(d.Entity)
	if d.MaxByteSize == 0 {
		d.MaxByteSize = maxInt64
	}
}

func (d *Definition) NewEntity() interface{} {
	return reflect.New(d.reflectedType).Interface()
}

func EntityFromCtx(ctx context.Context) interface{} {
	return ctx.Value(httpctx.EntityKey)
}

func Unmarshal(next httpctx.Handler, def *Definition) httpctx.Handler {
	def.Inspect()
	return httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(http.MaxBytesReader(w, r.Body, def.MaxByteSize))
		if err != nil {
			httperr.Return(httperr.Err{
				StatusCode: http.StatusRequestEntityTooLarge,
				Message:    "request size exceeded limit: " + err.Error(),
				Fields: map[string]interface{}{
					"byteLimit": def.MaxByteSize,
				},
			})
		}

		entity := def.NewEntity()

		ct := contentctx.RequestTypeFromCtx(ctx)
		if ct == nil {
			panic("missing required middleware: contentctx.Request")
		}
		if err := ct.Unmarshal(body, entity); err != nil {
			httperr.Return(httperr.Err{
				StatusCode: http.StatusBadRequest,
				Message:    "unable to parse body: " + err.Error(),
			})
		}

		newCtx := context.WithValue(ctx, httpctx.EntityKey, entity)
		next.ServeHTTPContext(newCtx, w, r)
	})
}

func Validate(next httpctx.Handler, def *Definition) httpctx.Handler {
	def.Inspect()
	return httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		e := EntityFromCtx(ctx)
		if e == nil {
			panic("missing required middleware: entityctx.Unmarshal")
		}

		if err := def.Validate(e); err != nil {
			httperr.Return(httperr.Err{
				StatusCode: http.StatusBadRequest,
				Message:    err.Error(),
			})
		}

		next.ServeHTTPContext(ctx, w, r)
	})
}
