/*
Package httperr provides a standardized error that can be returned by handlers.
It includes all info that would be included in a call to http.Error with the
optional addition of arbitrary fields.
*/
package httperr

// Err is a struct which carries the information of an error which occurs in
// a http handler.
type Err struct {
	StatusCode int                    `json:"-" xml:"-"`
	Message    string                 `json:"message" xml:"message"`
	Fields     map[string]interface{} `json:"fields,omitempty" xml:"fields,omitempty"`
}

// New creates an bare minimum http error.
func New(msg string, status int) Err {
	return Err{
		StatusCode: status,
		Message:    msg,
		Fields:     make(map[string]interface{}),
	}
}

// WithField returns a new Err with the given key-value pair included
// in the 'Fields' field.
func (err Err) WithField(name string, value interface{}) Err {
	err.Fields[name] = value
	return Err{
		StatusCode: err.StatusCode,
		Message:    err.Message,
		Fields:     err.Fields,
	}
}

// The Error() method allows the Err struct to satisfy the standard error
// interface.
func (err Err) Error() string {
	return err.Message
}
