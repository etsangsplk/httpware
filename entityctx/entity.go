package entityctx

import (
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/nstogner/contextware/contentctx"
	"github.com/nstogner/contextware/httpctx"
	"github.com/nstogner/contextware/httperr"
	"golang.org/x/net/context"
)

var maxInt64 = int64(^uint64(0) >> 1)

type Definition struct {
	Entity      interface{}
	Validate    Validator
	MaxByteSize int64
	Identify    Identifier

	reflectedType reflect.Type
}

type Validator func(interface{}) error
type Identifier func(interface{}) []byte

func (d *Definition) inspect() {
	d.reflectedType = reflect.TypeOf(d.Entity)
	if d.MaxByteSize == 0 {
		d.MaxByteSize = maxInt64
	}
}

func EntityFromCtx(ctx context.Context) interface{} {
	return ctx.Value(httpctx.JsonEntityKey)
}

func Unmarshal(next httpctx.Handler, def *Definition) httpctx.Handler {
	def.inspect()
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

		entity := reflect.New(def.reflectedType).Interface()

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

		newCtx := context.WithValue(ctx, httpctx.JsonEntityKey, entity)
		next.ServeHTTPContext(newCtx, w, r)
	})
}

func Validate(next httpctx.Handler, def *Definition) httpctx.Handler {
	def.inspect()
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
