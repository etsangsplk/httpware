package httperr

type Err struct {
	StatusCode int                    `json:"-"`
	Message    string                 `json:"message"`
	Fields     map[string]interface{} `json:"fields"`
}

func (err Err) Error() string {
	return err.Message
}
