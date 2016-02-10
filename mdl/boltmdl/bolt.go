package boltmdl

import (
	"bytes"
	"encoding/gob"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/nstogner/ctxware/adp/routeradp"
	"github.com/nstogner/ctxware/lib/httpctx"
	"github.com/nstogner/ctxware/lib/httperr"
	"github.com/nstogner/ctxware/mdl/contentmdl"
	"github.com/nstogner/ctxware/mdl/entitymdl"
	"golang.org/x/net/context"
)

var (
	ErrAlreadyExists = errors.New("entity already exists")
	ErrNotFound      = errors.New("entity not found")

	errorMap = map[error]httperr.Err{
		ErrAlreadyExists: httperr.Err{
			StatusCode: http.StatusConflict,
			Message:    ErrAlreadyExists.Error(),
		},
		ErrNotFound: httperr.Err{
			StatusCode: http.StatusNotFound,
			Message:    ErrNotFound.Error(),
		},
	}
)

type Definition struct {
	DB         *bolt.DB
	BucketPath string
	Identify   Identifier
	IdParam    string
	EntityDef  entitymdl.Definition
}

type Identifier func(interface{}) string

// Requires contentmdl.Unmarshal
func Post(next httpctx.Handler, def Definition) httpctx.Handler {
	bktsPath := make([][]byte, 0)
	for _, lvl := range strings.Split(def.BucketPath, "/") {
		if lvl != "" {
			bktsPath = append(bktsPath, []byte(lvl))
		}
	}
	bktDepth := len(bktsPath)

	return httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		entity := entitymdl.EntityFromCtx(ctx)
		if entity == nil {
			panic("missing required middleware: entitymdl.Unmarshal")
		}

		id := []byte(def.Identify(entity))
		buf := &bytes.Buffer{}
		err := gob.NewEncoder(buf).Encode(entity)
		if err != nil {
			httperr.Return(httperr.Err{
				StatusCode: http.StatusInternalServerError,
				Message:    err.Error(),
			})
		}
		dbGob, err := ioutil.ReadAll(buf)
		if err != nil {
			httperr.Return(httperr.Err{
				StatusCode: http.StatusInternalServerError,
				Message:    err.Error(),
			})
		}

		err = def.DB.Update(func(tx *bolt.Tx) error {
			bkt, _ := tx.CreateBucketIfNotExists(bktsPath[0])
			for i := 1; i < bktDepth; i++ {
				bkt, _ = bkt.CreateBucketIfNotExists(bktsPath[i])
			}
			if bkt.Get(id) != nil {
				return ErrAlreadyExists
			}
			bkt.Put(id, dbGob)
			return nil
		})
		if err != nil {
			httperr.Return(errorMap[err])
		}

		rct := contentmdl.RequestTypeFromCtx(ctx)
		if rct == nil {
			panic("missing required middleware: contentmdl.Response")
		}
		w.WriteHeader(http.StatusCreated)
		rct.MarshalWrite(w, entity)
	})
}

// requires params to be set
func Get(next httpctx.Handler, def Definition) httpctx.Handler {
	bktsPath := make([][]byte, 0)
	for _, lvl := range strings.Split(def.BucketPath, "/") {
		if lvl != "" {
			bktsPath = append(bktsPath, []byte(lvl))
		}
	}
	bktDepth := len(bktsPath)

	def.EntityDef.Inspect()

	return httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		ps := routeradp.ParamsFromCtx(ctx)
		if ps == nil {
			panic("missing required middleware: routeradp.Adapt or the like")
		}
		id := ps[def.IdParam]
		if id == "" {
			httperr.Return(errorMap[ErrAlreadyExists])
		}

		var dbGob []byte
		err := def.DB.View(func(tx *bolt.Tx) error {
			bkt := tx.Bucket(bktsPath[0])
			if bkt == nil {
				return ErrNotFound
			}
			for i := 1; i < bktDepth; i++ {
				bkt = bkt.Bucket(bktsPath[i])
				if bkt == nil {
					return ErrNotFound
				}
			}
			dbGob = bkt.Get([]byte(id))
			return nil
		})
		if err != nil {
			httperr.Return(errorMap[err])
		}

		rct := contentmdl.ResponseTypeFromCtx(ctx)
		if rct == nil {
			panic("missing required middleware: contentmdl.Response")
		}

		entity := def.EntityDef.NewEntity()
		buf := bytes.NewReader(dbGob)
		err = gob.NewDecoder(buf).Decode(entity)
		if err != nil {
			httperr.Return(httperr.Err{
				StatusCode: http.StatusInternalServerError,
				Message:    err.Error(),
			})
		}

		rct.MarshalWrite(w, entity)
	})
}
