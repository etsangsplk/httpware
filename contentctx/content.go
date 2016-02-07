package contentctx

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	"github.com/nstogner/netmiddle/httpctx"
	"github.com/nstogner/netmiddle/httperr"

	"golang.org/x/net/context"
)

const (
	KeyJson = 0
	KeyXml  = 1
)

var (
	Json = &ContentType{
		SearchText:   "json",
		Value:        "application/json",
		Key:          KeyJson,
		Unmarshal:    json.Unmarshal,
		MarshalWrite: MarshalWriteFunc(func(w io.Writer, bs interface{}) error { return json.NewEncoder(w).Encode(bs) }),
	}
	Xml = &ContentType{
		SearchText:   "xml",
		Value:        "application/xml",
		Key:          KeyXml,
		Unmarshal:    xml.Unmarshal,
		MarshalWrite: MarshalWriteFunc(func(w io.Writer, bs interface{}) error { return xml.NewEncoder(w).Encode(bs) }),
	}
	JsonAndXml = []*ContentType{
		Json,
		Xml,
	}
)

type UnmarshalFunc func([]byte, interface{}) error
type MarshalWriteFunc func(io.Writer, interface{}) error

type ContentType struct {
	SearchText   string
	Value        string
	Key          int32
	Unmarshal    UnmarshalFunc
	MarshalWrite MarshalWriteFunc
}

func EntityFromContext(ctx context.Context) interface{} {
	return ctx.Value(httpctx.JsonEntityKey)
}

func ReqContentTypeFromContext(ctx context.Context) *ContentType {
	ct := ctx.Value(httpctx.RequestContentTypeKey)
	if ct == nil {
		return nil
	}
	return ct.(*ContentType)
}

func RespContentTypeFromContext(ctx context.Context) *ContentType {
	ct := ctx.Value(httpctx.ResponseContentTypeKey)
	if ct == nil {
		return nil
	}
	return ct.(*ContentType)
}

func Request(next httpctx.Handler, types []*ContentType) httpctx.Handler {
	if len(types) == 0 {
		panic("content types slice must not be empty")
	}

	return httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		reqContType := GetContentMatch(r.Header.Get("Content-Type"), types)

		newCtx := context.WithValue(ctx, httpctx.RequestContentTypeKey, reqContType)
		return next.ServeHTTPContext(newCtx, w, r)
	})
}

func Response(next httpctx.Handler, types []*ContentType) httpctx.Handler {
	if len(types) == 0 {
		panic("content types slice must not be empty")
	}

	return httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		resContType := GetContentMatch(r.Header.Get("accept"), types)

		w.Header().Set("Content-Type", resContType.Value)
		newCtx := context.WithValue(ctx, httpctx.ResponseContentTypeKey, resContType)
		return next.ServeHTTPContext(newCtx, w, r)
	})
}

func GetContentMatch(header string, cts []*ContentType) *ContentType {
	for _, c := range cts {
		if strings.Contains(header, c.SearchText) {
			return c
		}
	}
	return cts[0]
}

func Unmarshal(next httpctx.Handler, v interface{}, maxBytesSize int64, unmarshaller UnmarshalFunc) httpctx.Handler {
	t := reflect.TypeOf(v)
	return httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		body, err := ioutil.ReadAll(http.MaxBytesReader(w, r.Body, maxBytesSize))
		if err != nil {
			return httperr.Err{
				StatusCode: http.StatusRequestEntityTooLarge,
				Message:    "request size exceeded limit: " + err.Error(),
				Fields: map[string]interface{}{
					"byteLimit": maxBytesSize,
				},
			}
		}

		entity := reflect.New(t).Interface()

		ct := ReqContentTypeFromContext(ctx)
		var uf UnmarshalFunc
		if unmarshaller != nil {
			uf = unmarshaller
		} else {
			uf = ct.Unmarshal
		}
		if err := uf(body, entity); err != nil {
			return httperr.Err{
				StatusCode: http.StatusBadRequest,
				Message:    "unable to parse body: " + err.Error(),
			}
		}

		newCtx := context.WithValue(ctx, httpctx.JsonEntityKey, entity)
		return next.ServeHTTPContext(newCtx, w, r)
	})
}
