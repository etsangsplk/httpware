package contentware

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"strings"

	"github.com/nstogner/ctxware"

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
	ct := ctx.Value(ctxware.RequestContentTypeKey)
	if ct == nil {
		return nil
	}
	return ct.(*ContentType)
}

func ResponseTypeFromCtx(ctx context.Context) *ContentType {
	ct := ctx.Value(ctxware.ResponseContentTypeKey)
	if ct == nil {
		return nil
	}
	return ct.(*ContentType)
}

type ReqType struct {
	types []*ContentType
}

func NewReqType(types []*ContentType) ReqType {
	if len(types) == 0 {
		panic("content types slice must not be empty")
	}
	return ReqType{
		types: types,
	}
}

func (rq ReqType) Name() string {
	return "contentware.ReqType"
}

func (rq ReqType) Dependencies() []string {
	return []string{}
}

func (rq ReqType) Handle(next ctxware.Handler) ctxware.Handler {
	return ctxware.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		reqContType := GetContentMatch(r.Header.Get("Content-Type"), rq.types)

		newCtx := context.WithValue(ctx, ctxware.RequestContentTypeKey, reqContType)
		return next.ServeHTTPContext(newCtx, w, r)
	})
}

type RespType struct {
	types []*ContentType
}

func NewRespType(types []*ContentType) RespType {
	if len(types) == 0 {
		panic("content types slice must not be empty")
	}
	return RespType{
		types: types,
	}
}

func (rp RespType) Name() string {
	return "contentware.RespType"
}

func (rp RespType) Dependencies() []string {
	return []string{}
}

func (rp RespType) Handle(next ctxware.Handler) ctxware.Handler {
	return ctxware.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		resContType := GetContentMatch(r.Header.Get("accept"), rp.types)

		w.Header().Set("Content-Type", resContType.Value)
		newCtx := context.WithValue(ctx, ctxware.ResponseContentTypeKey, resContType)
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
