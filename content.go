package httpware

import "strings"

var contentTypeStrings = []string{"json", "xml"}

const (
	// JSON Content-Type
	JSON = iota
	// XML Content-Type
	XML
)

// ContentType is a key which represents different http content-types.
type ContentType uint32

// ContentTypeFromHeader parses an http header looking for keywords such as
// 'json' or 'xml' and returns the appropriate type JSON or XML.
func ContentTypeFromHeader(header string) ContentType {
	for i, s := range contentTypeStrings {
		if strings.Contains(header, s) {
			return ContentType(i)
		}
	}
	// Default to JSON.
	return JSON
}
