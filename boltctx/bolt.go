package boltctx

import (
	"bytes"
	"encoding/gob"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/nstogner/contextware/contentctx"
	"github.com/nstogner/contextware/entityctx"
	"github.com/nstogner/contextware/httpctx"
	"github.com/nstogner/contextware/httperr"
	"golang.org/x/net/context"
)

var (
	ErrAlreadyExists = errors.New("entity already exists")

	errorMap = map[error]httperr.Err{
		ErrAlreadyExists: httperr.Err{
			StatusCode: http.StatusConflict,
			Message:    ErrAlreadyExists.Error(),
		},
	}
)

type Definition struct {
	DB         *bolt.DB
	BucketPath string
	Identify   Identifier
}

type Identifier func(interface{}) []byte

// Requires contentctx.Unmarshal
func Post(def Definition) httpctx.Handler {
	bktsPath := make([][]byte, 0)
	for _, lvl := range strings.Split(def.BucketPath, "/") {
		if lvl != "" {
			bktsPath = append(bktsPath, []byte(lvl))
		}
	}
	bktDepth := len(bktsPath)

	return httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		entity := entityctx.EntityFromCtx(ctx)
		if entity == nil {
			panic("missing required middleware: entityctx.Unmarshal")
		}

		id := def.Identify(entity)
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
			bkt.Put(id, dbGob)
			return nil
		})
		if err != nil {
			httperr.Return(errorMap[err])
		}

		rct := contentctx.RequestTypeFromCtx(ctx)
		if rct == nil {
			panic("missing required middleware: contentctx.Response")
		}
		w.WriteHeader(http.StatusCreated)
		rct.MarshalWrite(w, entity)
	})
}
