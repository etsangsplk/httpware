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

	"golang.org/x/net/context"
)

const (
	TypeJson = 0
	TypeXml  = 1
)

var (
	ContentTypeJson = &ContentType{
		SearchText: "json",
		Value:      "application/json",
		Key:        TypeJson,
		Unmarshal:  json.Unmarshal,
		Marshal:    MarshalFunc(func(w io.Writer, bs interface{}) error { return json.NewEncoder(w).Encode(bs) }),
	}
	ContentTypeXml = &ContentType{
		SearchText: "xml",
		Value:      "application/xml",
		Key:        TypeXml,
		Unmarshal:  xml.Unmarshal,
		Marshal:    MarshalFunc(func(w io.Writer, bs interface{}) error { return xml.NewEncoder(w).Encode(bs) }),
	}
)

type UnmarshalFunc func([]byte, interface{}) error
type MarshalFunc func(io.Writer, interface{}) error

type ContentType struct {
	SearchText string
	Value      string
	Key        int32
	Unmarshal  UnmarshalFunc
	Marshal    MarshalFunc
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

func Negotiate(next httpctx.Handler, requestTypes []*ContentType, responseTypes []*ContentType) httpctx.Handler {
	if len(responseTypes) == 0 {
		panic("preference slice must not be empty")
	}

	return httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		resContType := getContentMatch(r.Header.Get("accept"), requestTypes)
		reqContType := getContentMatch(r.Header.Get("Content-Type"), responseTypes)

		w.Header().Set("Content-Type", resContType.Value)
		newCtx := context.WithValue(ctx, httpctx.ResponseContentTypeKey, resContType)
		newerCtx := context.WithValue(newCtx, httpctx.RequestContentTypeKey, reqContType)
		return next.ServeHTTPContext(newerCtx, w, r)
	})
}

func getContentMatch(header string, cts []*ContentType) *ContentType {
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
			return httpctx.Err{
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
			return httpctx.Err{
				StatusCode: http.StatusBadRequest,
				Message:    "unable to parse body: " + err.Error(),
			}
		}

		newCtx := context.WithValue(ctx, httpctx.JsonEntityKey, entity)
		return next.ServeHTTPContext(newCtx, w, r)
	})
}
