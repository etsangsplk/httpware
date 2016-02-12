package boltmdl

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/nstogner/ctxware/adp/routeradp"
	"github.com/nstogner/ctxware/lib/httpctx"
	"github.com/nstogner/ctxware/lib/httperr"
	"github.com/nstogner/ctxware/mdl/contentmdl"
	"github.com/nstogner/ctxware/mdl/entitymdl"
	"github.com/spaolacci/murmur3"
	"golang.org/x/net/context"
)

var (
	ErrAlreadyExists     = errors.New("entity already exists")
	ErrNotFound          = errors.New("entity not found")
	ErrETagNoMatch       = errors.New("e-tag does not match")
	ErrMissingETagHeader = errors.New("missing e-tag header")

	errorMap = map[error]httperr.Err{
		ErrAlreadyExists: httperr.Err{
			StatusCode: http.StatusConflict,
			Message:    ErrAlreadyExists.Error(),
		},
		ErrNotFound: httperr.Err{
			StatusCode: http.StatusNotFound,
			Message:    ErrNotFound.Error(),
		},
		ErrETagNoMatch: httperr.Err{
			StatusCode: 412, // Precondition failed
			Message:    ErrETagNoMatch.Error(),
		},
		ErrMissingETagHeader: httperr.Err{
			StatusCode: http.StatusBadRequest,
			Message:    ErrMissingETagHeader.Error(),
		},
	}
)

type Definition struct {
	DB               *bolt.DB
	EntityBucketPath string
	ETagBucketPath   string
	IdField          string
	IdParam          string
	EntityDef        entitymdl.Definition
}

func (d *Definition) IdOf(e interface{}) string {
	return reflect.ValueOf(e).Elem().FieldByName(d.IdField).String()
}

func Post(next httpctx.Handler, def Definition) httpctx.Handler {
	return update(next, def, "POST")
}
func Put(next httpctx.Handler, def Definition) httpctx.Handler {
	return update(next, def, "PUT")
}

func bucketSlice(path string) [][]byte {
	bs := make([][]byte, 0)
	for _, lvl := range strings.Split(path, "/") {
		if lvl != "" {
			bs = append(bs, []byte(lvl))
		}
	}
	return bs
}

func nestedBucket(tx *bolt.Tx, path [][]byte) *bolt.Bucket {
	bkt, _ := tx.CreateBucketIfNotExists(path[0])
	for i := 1; i < len(path); i++ {
		bkt, _ = bkt.CreateBucketIfNotExists(path[i])
	}
	return bkt
}

func generateETag(entity []byte) uint64 {
	etag := murmur3.Sum64(entity)
	return etag
}

// itob returns an 8-byte big endian representation of v.
func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

// Requires contentmdl.Unmarshal
func update(next httpctx.Handler, def Definition, method string) httpctx.Handler {
	entityBktPath := bucketSlice(def.EntityBucketPath)

	etagsEnabled := (def.ETagBucketPath != "")
	etagBktPath := bucketSlice(def.ETagBucketPath)

	// Predefine method-dependant actions so that conditionals are not needed
	// while handling requests.
	var updateFunc func(*http.Request, *int, *string, *bolt.Tx, []byte, []byte) error
	switch method {
	case "POST":
		if etagsEnabled {
			updateFunc = func(r *http.Request, wc *int, etg *string, tx *bolt.Tx, id []byte, dbGob []byte) error {
				etagBkt := nestedBucket(tx, etagBktPath)
				if etagBkt.Get(id) != nil {
					return ErrAlreadyExists
				}
				etagInt := generateETag(dbGob)
				*etg = strconv.FormatUint(etagInt, 10)
				if err := etagBkt.Put(id, itob(etagInt)); err != nil {
					return err
				}
				entBkt := nestedBucket(tx, entityBktPath)
				if err := entBkt.Put(id, dbGob); err != nil {
					return err
				}
				*wc = http.StatusCreated
				return nil
			}
		} else {
			updateFunc = func(r *http.Request, wc *int, etg *string, tx *bolt.Tx, id []byte, dbGob []byte) error {
				b := nestedBucket(tx, entityBktPath)
				if b.Get(id) != nil {
					return ErrAlreadyExists
				}
				if err := b.Put(id, dbGob); err != nil {
					return err
				}
				*wc = http.StatusCreated
				return nil
			}
		}
	case "PUT":
		if etagsEnabled {
			updateFunc = func(r *http.Request, wc *int, etg *string, tx *bolt.Tx, id []byte, dbGob []byte) error {
				*wc = http.StatusCreated
				etagBkt := nestedBucket(tx, etagBktPath)
				etagB := etagBkt.Get(id)
				if etagB != nil {
					etagInt := binary.BigEndian.Uint64(etagB)
					currentEtag := strconv.FormatUint(etagInt, 10)
					if inm := r.Header.Get("If-None-Match"); inm != currentEtag {
						if inm == "" {
							return ErrMissingETagHeader
						} else {
							return ErrETagNoMatch
						}
					}
					*wc = http.StatusOK
				}
				entityBkt := nestedBucket(tx, entityBktPath)
				if err := entityBkt.Put(id, dbGob); err != nil {
					return err
				}
				newEtagInt := generateETag(dbGob)
				*etg = strconv.FormatUint(newEtagInt, 10)
				if err := etagBkt.Put(id, itob(newEtagInt)); err != nil {
					return err
				}
				return nil
			}
		} else {
			updateFunc = func(r *http.Request, wc *int, etg *string, tx *bolt.Tx, id []byte, dbGob []byte) error {
				*wc = http.StatusCreated
				b := nestedBucket(tx, entityBktPath)
				if b.Get(id) != nil {
					*wc = http.StatusOK
				}
				if err := b.Put(id, dbGob); err != nil {
					return err
				}
				return nil
			}
		}
	default:
		panic("only POST and PUT methods are supported")
	}

	return httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		entity := entitymdl.EntityFromCtx(ctx)
		if entity == nil {
			panic("missing required middleware: entitymdl.Unmarshal")
		}

		id := []byte(def.IdOf(entity))
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

		var respCode int
		var entityTag string
		err = def.DB.Update(func(tx *bolt.Tx) error {
			return updateFunc(r, &respCode, &entityTag, tx, id, dbGob)
		})
		if err != nil {
			httpErr, match := errorMap[err]
			if !match {
				httperr.Return(httperr.Err{
					StatusCode: http.StatusInternalServerError,
					Message:    err.Error(),
				})
			}
			httperr.Return(httpErr)
		}

		rct := contentmdl.ResponseTypeFromCtx(ctx)
		if rct == nil {
			panic("missing required middleware: contentmdl.Response")
		}
		w.Header().Set("ETag", entityTag)
		w.WriteHeader(respCode)
		rct.MarshalWrite(w, entity)
	})
}

// requires params to be set
func Get(next httpctx.Handler, def Definition) httpctx.Handler {
	bktsPath := make([][]byte, 0)
	for _, lvl := range strings.Split(def.EntityBucketPath, "/") {
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
