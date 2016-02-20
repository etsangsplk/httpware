/*
Package contentware provides http middleware for parsing content-types in http
requests (using the 'Accept' & 'Content-Type' headers).
*/
package contentware

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"strings"

	"github.com/nstogner/httpware"

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

	JsonOverXml = []*ContentType{Json, Xml}
	XmlOverJson = []*ContentType{Xml, Json}

	// Defaults is a reasonable confguration that should work 90% of the time.
	Defaults = Config{
		RequestTypes:  JsonOverXml,
		ResponseTypes: JsonOverXml,
	}
)

type Config struct {
	// RequestTypes is a list of parsable types in order of preference.
	RequestTypes []*ContentType
	// ResponseTypes is a list of serializable types in order of preference.
	ResponseTypes []*ContentType
}

type UnmarshalFunc func([]byte, interface{}) error

// MarshalWriteFunc marshals and writes all in one go.
type MarshalWriteFunc func(io.Writer, interface{}) error

type ContentType struct {
	// Text which a header should contain to result in this content type
	SearchText string
	// The header value that will be set for this content type
	Value string
	// Identifies the content type
	Key int32
	// Function which is used to unmarshal the http request body
	Unmarshal UnmarshalFunc
	// Function with is used to write an entity to an io.Writer
	// (usually http.ResponseWriter)
	MarshalWrite MarshalWriteFunc
}

// RequestTypeFromCtx gives the content type that was parsed from the
// 'Content-Type' header.
func RequestTypeFromCtx(ctx context.Context) *ContentType {
	ct := ctx.Value(httpware.RequestContentTypeKey)
	if ct == nil {
		return nil
	}
	return ct.(*ContentType)
}

// ResponseTypeFromCtx gives the content type that was parsed from the
// 'Accept' header.
func ResponseTypeFromCtx(ctx context.Context) *ContentType {
	ct := ctx.Value(httpware.ResponseContentTypeKey)
	if ct == nil {
		return nil
	}
	return ct.(*ContentType)
}

// contentware.Middle is middleware that parses content types. The 'Content-Type'
// header is inspected for determining the request content type. The 'Accept'
// header is parsed for determining the appropriate response content type.
type Middle struct {
	conf Config
}

// New creates a new instance of the middleware. It panics if an invalid
// configuration is passed.
func New(conf Config) *Middle {
	middle := Middle{
		conf: conf,
	}
	if len(conf.RequestTypes) == 0 {
		panic("conf.RequestTypes must not be empty")
	}
	if len(conf.ResponseTypes) == 0 {
		panic("conf.ResponseTypes must not be empty")
	}
	return &middle
}

func (m *Middle) Contains() []string { return []string{"github.com/nstogner/contentware"} }
func (m *Middle) Requires() []string { return []string{} }

func (m *Middle) Handle(next httpware.Handler) httpware.Handler {
	return httpware.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		ctx = context.WithValue(ctx, httpware.RequestContentTypeKey, GetContentMatch(r.Header.Get("Content-Type"), m.conf.RequestTypes))
		ct := GetContentMatch(r.Header.Get("Accept"), m.conf.ResponseTypes)
		ctx = context.WithValue(ctx, httpware.ResponseContentTypeKey, ct)
		w.Header().Set("Content-Type", ct.Value)

		return next.ServeHTTPContext(ctx, w, r)
	})
}

// GetContentMatch parses a http header and returns the determined
// ContentType struct. If multiple types match, it will choose with priority
// given to the first elements in the given ContentType array.
func GetContentMatch(header string, cts []*ContentType) *ContentType {
	for _, c := range cts {
		if strings.Contains(header, c.SearchText) {
			return c
		}
	}
	return cts[0]
}
