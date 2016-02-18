/*
Package httperr provides a standardized error that can be returned by handlers.
It includes all info that would be included in a call to http.Error with the
optional addition of arbitrary fields.
*/
package httperr

type Err struct {
	StatusCode int                    `json:"-" xml:"-"`
	Message    string                 `json:"message" xml:"message"`
	Fields     map[string]interface{} `json:"fields" xml:"fields"`
}

func New(msg string, status int) Err {
	return Err{
		StatusCode: status,
		Message:    msg,
		Fields:     make(map[string]interface{}),
	}
}

func (err Err) WithField(name string, value interface{}) Err {
	err.Fields[name] = value
	return Err{
		StatusCode: err.StatusCode,
		Message:    err.Message,
		Fields:     err.Fields,
	}
}

func (err Err) Error() string {
	return err.Message
}
