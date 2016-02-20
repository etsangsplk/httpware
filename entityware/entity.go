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

type Config struct {
	// MaxBodySize is the maximum request body (bytes) that will be accepted.
	MaxBodySize int64
	// Entity is a non-pointer instance of the expected entity.
	Entity interface{}
	// If left nil, Validate will not be called.
	Validate ValidateFunc
}

type ValidateFunc func(interface{}) error

type Middle struct {
	conf          Config
	reflectedType reflect.Type
}

func New(conf Config) *Middle {
	return &Middle{
		conf:          conf,
		reflectedType: reflect.TypeOf(conf.Entity),
	}
}

func (m *Middle) Contains() []string { return []string{"github.com/nstogner/entityware"} }
func (m *Middle) Requires() []string { return []string{"github.com/nstogner/contentware"} }

// NewEnitity returns a pointer to the new instance of an entity.
func (m *Middle) NewEntity() interface{} {
	return reflect.New(m.reflectedType).Interface()
}

func (m *Middle) Handle(next httpware.Handler) httpware.Handler {
	return httpware.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		body, err := ioutil.ReadAll(http.MaxBytesReader(w, r.Body, m.conf.MaxBodySize))
		if err != nil {
			return httperr.New("request size exceeded limit", http.StatusRequestEntityTooLarge).WithField("byteLimit", m.conf.MaxBodySize)
		}

		// Pointer to a new instance of the entity
		entity := m.NewEntity()

		// Unmarshal the body based on the content type that was determined.
		ct := contentware.RequestTypeFromCtx(ctx)
		if err := ct.Unmarshal(body, entity); err != nil {
			httperr.New("unable to parse body: "+err.Error(), http.StatusBadRequest)
		}

		// Validate the entity if the validate function was defined.
		if m.conf.Validate != nil {
			if err := m.conf.Validate(entity); err != nil {
				return httperr.New(err.Error(), http.StatusBadRequest)
			}
		}

		newCtx := context.WithValue(ctx, httpware.EntityKey, entity)
		return next.ServeHTTPContext(newCtx, w, r)
	})
}
