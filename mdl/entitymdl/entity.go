package entitymdl

import (
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/nstogner/ctxware"
	"github.com/nstogner/ctxware/lib/httperr"
	"github.com/nstogner/ctxware/mdl/contentmdl"

	"golang.org/x/net/context"
)

var Maximum = int64(^uint64(0) >> 1)

func EntityFromCtx(ctx context.Context) interface{} {
	return ctx.Value(ctxware.EntityKey)
}

// Parser

type Parser struct {
	entity        interface{}
	reflectedType reflect.Type
	maxSize       int64
}

func NewParser(entity interface{}, maxSize int64) Parser {
	return Parser{
		entity:        entity,
		reflectedType: reflect.TypeOf(entity),
		maxSize:       maxSize,
	}
}

func (p Parser) Name() string {
	return "entitymdl.Parser"
}

func (p Parser) Dependencies() []string {
	return []string{"contentmdl.ReqType"}
}

func (p Parser) NewEntity() interface{} {
	return reflect.New(p.reflectedType).Interface()
}

func (p Parser) Handle(next ctxware.Handler) ctxware.Handler {
	return ctxware.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		body, err := ioutil.ReadAll(http.MaxBytesReader(w, r.Body, p.maxSize))
		if err != nil {
			return httperr.Err{
				StatusCode: http.StatusRequestEntityTooLarge,
				Message:    "request size exceeded limit: " + err.Error(),
				Fields: map[string]interface{}{
					"byteLimit": p.maxSize,
				},
			}
		}

		entity := p.NewEntity()

		ct := contentmdl.RequestTypeFromCtx(ctx)
		if err := ct.Unmarshal(body, entity); err != nil {
			return httperr.Err{
				StatusCode: http.StatusBadRequest,
				Message:    "unable to parse body: " + err.Error(),
			}
		}

		newCtx := context.WithValue(ctx, ctxware.EntityKey, entity)
		return next.ServeHTTPContext(newCtx, w, r)
	})
}

// Validator

type Validator struct {
	validate ValidateFunc
}

type ValidateFunc func(interface{}) error

func NewValidator(vf ValidateFunc) Validator {
	if vf == nil {
		panic("validate func must not be nil")
	}
	return Validator{
		validate: vf,
	}
}

func (v Validator) Name() string {
	return "entitymdl.Validator"
}

func (v Validator) Dependencies() []string {
	return []string{"entitymdl.Parser"}
}

func (v Validator) Handle(next ctxware.Handler) ctxware.Handler {
	return ctxware.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		e := EntityFromCtx(ctx)

		if err := v.validate(e); err != nil {
			return httperr.Err{
				StatusCode: http.StatusBadRequest,
				Message:    err.Error(),
			}
		}

		return next.ServeHTTPContext(ctx, w, r)
	})
}
