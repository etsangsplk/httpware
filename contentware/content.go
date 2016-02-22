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
	// KeyJSON identifies the JSON content type.
	KeyJSON = 0
	// KeyXML identifies the XML content type.
	KeyXML = 1
)

var (
	// JSON content type
	JSON = &ContentType{
		SearchText:   "json",
		Value:        "application/json",
		Key:          KeyJSON,
		Unmarshal:    json.Unmarshal,
		MarshalWrite: MarshalWriteFunc(func(w io.Writer, bs interface{}) error { return json.NewEncoder(w).Encode(bs) }),
	}
	// XML content type
	XML = &ContentType{
		SearchText:   "xml",
		Value:        "application/xml",
		Key:          KeyXML,
		Unmarshal:    xml.Unmarshal,
		MarshalWrite: MarshalWriteFunc(func(w io.Writer, bs interface{}) error { return xml.NewEncoder(w).Encode(bs) }),
	}

	// JSONOverXML is the preference of using JSON over XML.
	JSONOverXML = []*ContentType{JSON, XML}
	// XMLOverJSON is the preference of using XML over JSON.
	XMLOverJSON = []*ContentType{XML, JSON}

	// Defaults is a reasonable confguration that should work 90% of the time.
	Defaults = Config{
		RequestTypes:  JSONOverXML,
		ResponseTypes: JSONOverXML,
	}
)

// Config is used to define content type preferences.
type Config struct {
	// RequestTypes is a list of parsable types in order of preference.
	RequestTypes []*ContentType
	// ResponseTypes is a list of serializable types in order of preference.
	ResponseTypes []*ContentType
}

// UnmarshalFunc calls an unmarshaller (ie: JSON/XML).
type UnmarshalFunc func([]byte, interface{}) error

// MarshalWriteFunc marshals and writes all in one go.
type MarshalWriteFunc func(io.Writer, interface{}) error

// ContentType is a struct which makes serializing and deserializing http
// requests/responses easier.
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

// Middle is middleware that parses content types. The 'Content-Type'
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

// Contains indentifies this middleware for compositions.
func (m *Middle) Contains() []string { return []string{"github.com/nstogner/contentware"} }

// Requires indentifies what this middleware depends on (nothing).
func (m *Middle) Requires() []string { return []string{} }

// Handle takes the next handler as an argument and wraps it in this middleware.
func (m *Middle) Handle(next httpware.Handler) httpware.Handler {
	return httpware.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		ctx = context.WithValue(ctx, httpware.RequestContentTypeKey, GetContentMatch(r.Header.Get("Content-Type"), m.conf.RequestTypes))
		ct := GetContentMatch(r.Header.Get("Accept"), m.conf.ResponseTypes)
		ctx = context.WithValue(ctx, httpware.ResponseContentTypeKey, ct)
		w.Header().Set("Content-Type", ct.Value)

		return next.ServeHTTPCtx(ctx, w, r)
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
