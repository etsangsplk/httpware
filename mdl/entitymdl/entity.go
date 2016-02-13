package entitymdl

import (
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/nstogner/ctxware/lib/httpctx"
	"github.com/nstogner/ctxware/lib/httperr"
	"github.com/nstogner/ctxware/mdl/contentmdl"

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
}

func (d *Definition) NewEntity() interface{} {
	return reflect.New(d.reflectedType).Interface()
}

func EntityFromCtx(ctx context.Context) interface{} {
	return ctx.Value(httpctx.EntityKey)
}

func Unmarshal(next httpctx.Handler, def Definition) httpctx.Handler {
	if def.MaxByteSize == 0 {
		def.MaxByteSize = maxInt64
	}
	def.Inspect()
	return httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		body, err := ioutil.ReadAll(http.MaxBytesReader(w, r.Body, def.MaxByteSize))
		if err != nil {
			return httperr.Err{
				StatusCode: http.StatusRequestEntityTooLarge,
				Message:    "request size exceeded limit: " + err.Error(),
				Fields: map[string]interface{}{
					"byteLimit": def.MaxByteSize,
				},
			}
		}

		entity := def.NewEntity()

		ct := contentmdl.RequestTypeFromCtx(ctx)
		if ct == nil {
			panic("missing required middleware: contentmdl.Request")
		}
		if err := ct.Unmarshal(body, entity); err != nil {
			return httperr.Err{
				StatusCode: http.StatusBadRequest,
				Message:    "unable to parse body: " + err.Error(),
			}
		}

		newCtx := context.WithValue(ctx, httpctx.EntityKey, entity)
		return next.ServeHTTPContext(newCtx, w, r)
	})
}

func Validate(next httpctx.Handler, def Definition) httpctx.Handler {
	if def.Validate == nil {
		panic("Validate field must be defined")
	}
	def.Inspect()
	return httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		e := EntityFromCtx(ctx)
		if e == nil {
			panic("missing required middleware: entitymdl.Unmarshal")
		}

		if err := def.Validate(e); err != nil {
			return httperr.Err{
				StatusCode: http.StatusBadRequest,
				Message:    err.Error(),
			}
		}

		return next.ServeHTTPContext(ctx, w, r)
	})
}
