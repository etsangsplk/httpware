package httperr

type Err struct {
	StatusCode int                    `json:"-" xml:"-"`
	Message    string                 `json:"message" xml:"message"`
	Fields     map[string]interface{} `json:"fields" xml:"fields"`
}

func (err Err) Error() string {
	return err.Message
}
