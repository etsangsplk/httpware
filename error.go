package httpware

// Err is a struct which carries the information of an error which occurs in
// a http handler.
type Err struct {
	StatusCode int                    `json:"-" xml:"-"`
	Message    string                 `json:"message" xml:"message"`
	Fields     map[string]interface{} `json:"fields,omitempty" xml:"fields,omitempty"`
}

// NewErr creates an bare minimum http error.
func NewErr(msg string, status int) Err {
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
