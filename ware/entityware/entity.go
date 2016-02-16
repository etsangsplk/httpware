/*
Package entityware provides http middleware for parsing serialized entities
into their Go datastructures. It uses the ctxware.Middleware interface.
*/
package entityware

import (
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/nstogner/ctxware"
	"github.com/nstogner/ctxware/lib/httperr"
	"github.com/nstogner/ctxware/ware/contentware"

	"golang.org/x/net/context"
)

const (
	KB  = 1024
	MB  = 1024 * 1024
	GB  = 1024 * 1024 * 1024
	MAX = int64(^uint64(0) >> 1)
)

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

func (p Parser) Contains() []string {
	return []string{"entityware.Parser"}
}

func (p Parser) Requires() []string {
	return []string{"contentware.ReqType"}
}

func (p Parser) NewEntity() interface{} {
	return reflect.New(p.reflectedType).Interface()
}

func (p Parser) Handle(next ctxware.Handler) ctxware.Handler {
	return ctxware.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		body, err := ioutil.ReadAll(http.MaxBytesReader(w, r.Body, p.maxSize))
		if err != nil {
			return httperr.New("request size exceeded limit", http.StatusRequestEntityTooLarge).WithField("byteLimit", p.maxSize)
		}

		entity := p.NewEntity()

		ct := contentware.RequestTypeFromCtx(ctx)
		if err := ct.Unmarshal(body, entity); err != nil {
			httperr.New("unable to parse body: "+err.Error(), http.StatusBadRequest)
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

func (v Validator) Contains() []string {
	return []string{"entityware.Validator"}
}

func (v Validator) Requires() []string {
	return []string{"entityware.Parser"}
}

func (v Validator) Handle(next ctxware.Handler) ctxware.Handler {
	return ctxware.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		e := EntityFromCtx(ctx)

		if err := v.validate(e); err != nil {
			return httperr.New(err.Error(), http.StatusBadRequest)
		}

		return next.ServeHTTPContext(ctx, w, r)
	})
}
