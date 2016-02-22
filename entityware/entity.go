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
	// KB = Kilobytes
	KB = 1024
	// MB = Megabytes
	MB = 1024 * 1024
	// GB = Gigabytes
	GB = 1024 * 1024 * 1024
	// MAX is the maximum value of an int64
	MAX = int64(^uint64(0) >> 1)
)

// EntityFromCtx returns an interface{} which contains a pointer to an entity
// of the type passed in the config, and parsed from the request body. It can
// be converted to the configured type, but don't forget to use a pointer,
// ie: EntityFromCtx(ctx).(*User)
func EntityFromCtx(ctx context.Context) interface{} {
	return ctx.Value(httpware.EntityKey)
}

// Config is passed to the New function to initiate an instance of Middle.
type Config struct {
	// MaxBodySize is the maximum request body (bytes) that will be accepted.
	MaxBodySize int64
	// Entity is a non-pointer instance of the expected entity.
	Entity interface{}
	// If left nil, Validate will not be called.
	Validate ValidateFunc
}

// ValidateFunc will be passed an interface{} which can be converted to a
// pointer to an entity.
type ValidateFunc func(interface{}) error

// Middle parses and optionally validates an entity in an http
// request body.
type Middle struct {
	conf          Config
	reflectedType reflect.Type
}

// New returns a new instance of the middleware.
func New(conf Config) *Middle {
	return &Middle{
		conf:          conf,
		reflectedType: reflect.TypeOf(conf.Entity),
	}
}

// Contains indentifies this middleware for compositions.
func (m *Middle) Contains() []string { return []string{"github.com/nstogner/entityware"} }

// Requires indentifies what this middleware depends on. In this case,
// it depends on github.com/nstogner/contentware.
func (m *Middle) Requires() []string { return []string{"github.com/nstogner/contentware"} }

// NewEntity returns a pointer to a new instance of the entity provided in
// the configuration.
func (m *Middle) NewEntity() interface{} {
	return reflect.New(m.reflectedType).Interface()
}

// Handle takes the next handler as an argument and wraps it in this middleware.
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

		ctx = context.WithValue(ctx, httpware.EntityKey, entity)
		return next.ServeHTTPCtx(ctx, w, r)
	})
}
