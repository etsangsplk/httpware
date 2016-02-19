/*
Package entityware provides http middleware for parsing serialized entities
into their Go datastructures. It uses the httpware.Middleware interface.
*/
package entityware

import (
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/nstogner/httpware"
	"github.com/nstogner/httpware/contentware"
	"github.com/nstogner/httpware/httperr"

	"golang.org/x/net/context"
)

const (
	KB  = 1024
	MB  = 1024 * 1024
	GB  = 1024 * 1024 * 1024
	MAX = int64(^uint64(0) >> 1)
)

func EntityFromCtx(ctx context.Context) interface{} {
	return ctx.Value(httpware.EntityKey)
}

type Def struct {
	// MaxBodySize is the maximum request body (bytes) that will be accepted.
	MaxBodySize int64
	// Entity is a non-pointer instance of the expected entity.
	Entity interface{}
	// If left nil, Validate will not be called.
	Validate ValidateFunc
}

type ValidateFunc func(interface{}) error

type Ware struct {
	def           Def
	reflectedType reflect.Type
}

func New(def Def) Ware {
	return Ware{
		def:           def,
		reflectedType: reflect.TypeOf(def.Entity),
	}
}

func (ware Ware) Contains() []string { return []string{"entityware.Ware"} }
func (ware Ware) Requires() []string { return []string{"contentware.ReqType"} }

// NewEnitity returns a pointer to the new instance of an entity.
func (ware Ware) NewEntity() interface{} {
	return reflect.New(ware.reflectedType).Interface()
}

func (ware Ware) Handle(next httpware.Handler) httpware.Handler {
	return httpware.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		body, err := ioutil.ReadAll(http.MaxBytesReader(w, r.Body, ware.def.MaxBodySize))
		if err != nil {
			return httperr.New("request size exceeded limit", http.StatusRequestEntityTooLarge).WithField("byteLimit", ware.def.MaxBodySize)
		}

		// Pointer to a new instance of the entity
		entity := ware.NewEntity()

		// Unmarshal the body based on the content type that was determined.
		ct := contentware.RequestTypeFromCtx(ctx)
		if err := ct.Unmarshal(body, entity); err != nil {
			httperr.New("unable to parse body: "+err.Error(), http.StatusBadRequest)
		}

		// Validate the entity if the validate function was defined.
		if ware.def.Validate != nil {
			if err := ware.def.Validate(entity); err != nil {
				return httperr.New(err.Error(), http.StatusBadRequest)
			}
		}

		newCtx := context.WithValue(ctx, httpware.EntityKey, entity)
		return next.ServeHTTPContext(newCtx, w, r)
	})
}
