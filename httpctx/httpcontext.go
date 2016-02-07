package httpctx

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"strings"

	"golang.org/x/net/context"
)

const (
	TokenKey               = 0
	ParamsKey              = 1
	JsonEntityKey          = 2
	RequestContentTypeKey  = 3
	ResponseContentTypeKey = 4
)

type Handler interface {
	ServeHTTPContext(context.Context, http.ResponseWriter, *http.Request) error
}

type HandlerFunc func(context.Context, http.ResponseWriter, *http.Request) error

func (h HandlerFunc) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return h(ctx, w, r)
}

func Adapt(h Handler) Adapter {
	return Adapter{
		ctx:     context.Background(),
		handler: h,
	}
}

type Adapter struct {
	ctx     context.Context
	handler Handler
}

func (ca Adapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := ca.handler.ServeHTTPContext(ca.ctx, w, r)
	if err == nil {
		return
	}

	if httpErr, ok := err.(Err); ok {
		var body []byte

		// TODO: Look into whether this is the right way to decide content type
		// Maybe: Content-Negotiation middleware
		contentType := w.Header().Get("Content-Type")
		if strings.Contains(contentType, "json") {
			body, _ = json.MarshalIndent(httpErr, "", "  ")
		} else if strings.Contains(contentType, "xml") {
			body, _ = xml.MarshalIndent(httpErr, "", "  ")
		} else {
			body = []byte(httpErr.Message)
		}

		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(httpErr.StatusCode)
		w.Write(body)
		return
	}

	http.Error(w, err.Error(), http.StatusInternalServerError)
	return
}

type Err struct {
	StatusCode int                    `json:"-"`
	Message    string                 `json:"message"`
	Fields     map[string]interface{} `json:"fields"`
}

func (err Err) Error() string {
	return err.Message
}
