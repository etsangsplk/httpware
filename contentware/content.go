/*
Package contentware provides http middleware for parsing content-types in http
requests (using the 'Accept' & 'Content-Type' headers).
*/
package contentware

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"

	"github.com/jriquelme/httpware"
)

var (
	// JSON content type
	contentTypes = []*ContentType{
		&ContentType{
			SearchText: "json",
			Value:      "application/json",
			Key:        httpware.JSON,
			Unmarshal:  json.Unmarshal,
			Decode:     DecodeFunc(func(r io.Reader, e interface{}) error { return json.NewDecoder(r).Decode(e) }),
			Marshal:    json.Marshal,
			Encode:     EncodeFunc(func(w io.Writer, bs interface{}) error { return json.NewEncoder(w).Encode(bs) }),
		},
		// XML content type
		&ContentType{
			SearchText: "xml",
			Value:      "application/xml",
			Key:        httpware.XML,
			Unmarshal:  xml.Unmarshal,
			Decode:     DecodeFunc(func(r io.Reader, e interface{}) error { return xml.NewDecoder(r).Decode(e) }),
			Marshal:    xml.Marshal,
			Encode:     EncodeFunc(func(w io.Writer, bs interface{}) error { return xml.NewEncoder(w).Encode(bs) }),
		},
	}

	// Defaults is a placeholder.
	Defaults = Config{}
)

// Config is used to define content type preferences.
type Config struct {
}

// DecodeFunc calls reads and unmarshals from a io.Reader.
type DecodeFunc func(io.Reader, interface{}) error

// UnmarshalFunc calls an unmarshaller (ie: JSON/XML).
type UnmarshalFunc func([]byte, interface{}) error

// EncodeFunc marshals and writes all in one go.
type EncodeFunc func(io.Writer, interface{}) error

// MarshalFunc marshals an interface to the correct content type.
type MarshalFunc func(interface{}) ([]byte, error)

// ContentType is a struct which makes serializing and deserializing http
// requests/responses easier.
type ContentType struct {
	// Text which a header should contain to result in this content type
	SearchText string
	// The header value that will be set for this content type
	Value string
	// Identifies the content type
	Key httpware.ContentType
	// Function which is used to read and unmarshal from a io.Reader
	Decode DecodeFunc
	// Function which is used to unmarshal the http request body
	Unmarshal UnmarshalFunc
	// Function is used to write an entity to an io.Writer
	// (usually http.ResponseWriter)
	Encode EncodeFunc
	// Function is used to marshal to a byte array
	Marshal MarshalFunc
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

// New creates a new instance of the middleware.
func New(conf Config) *Middle {
	middle := Middle{
		conf: conf,
	}
	return &middle
}

// Handle takes the next handler as an argument and wraps it in this middleware.
func (m *Middle) Handle(next httpware.Handler) httpware.Handler {
	return httpware.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		ctx = context.WithValue(ctx, httpware.RequestContentTypeKey, GetContentMatch(r.Header.Get("Content-Type")))
		ct := GetContentMatch(r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, httpware.ResponseContentTypeKey, ct)
		w.Header().Set("Content-Type", ct.Value)

		return next.ServeHTTPCtx(ctx, w, r)
	})
}

// GetContentMatch parses a http header and returns the determined
// ContentType struct. If multiple types match, it will choose with priority
// given to the first elements in the given ContentType array.
func GetContentMatch(header string) *ContentType {
	return contentTypes[httpware.ContentTypeFromHeader(header)]
}
