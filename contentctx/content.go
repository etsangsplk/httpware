package contentctx

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"strings"

	"github.com/nstogner/contextware/httpctx"

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

func RequestTypeFromCtx(ctx context.Context) *ContentType {
	ct := ctx.Value(httpctx.RequestContentTypeKey)
	if ct == nil {
		return nil
	}
	return ct.(*ContentType)
}

func ResponseTypeFromCtx(ctx context.Context) *ContentType {
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

	return httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		reqContType := GetContentMatch(r.Header.Get("Content-Type"), types)

		newCtx := context.WithValue(ctx, httpctx.RequestContentTypeKey, reqContType)
		next.ServeHTTPContext(newCtx, w, r)
	})
}

func Response(next httpctx.Handler, types []*ContentType) httpctx.Handler {
	if len(types) == 0 {
		panic("content types slice must not be empty")
	}

	return httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		resContType := GetContentMatch(r.Header.Get("accept"), types)

		w.Header().Set("Content-Type", resContType.Value)
		newCtx := context.WithValue(ctx, httpctx.ResponseContentTypeKey, resContType)
		next.ServeHTTPContext(newCtx, w, r)
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
