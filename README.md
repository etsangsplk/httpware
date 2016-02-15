# ctxware

#### Golang HTTP middleware using net/context

**NOTE: CURRENTLY UNDER HEAVY DEVELOPMENT !!!**

This repository contains a collection of stackable middleware packages which aid in writing http handlers in Go. All middleware is of the form:
```Go
func(context.Context, http.ResponseWriter, *http.Request) error
```
It also includes a set of functions (inspired by [alice](https://github.com/justinas/alice)) which make composing middleware as simple as possible.

#### Middleware
| Functionality | Package |
|:--------------|:-------:|
| Handling errors | errorware |
| Logging ([logrus](https://github.com/Sirupsen/logrus)) | logware |
| Parsing Request & Response Content Types | contentware |
| Unmarshalling (json/xml) and validating entities | entityware |
| JWT authentication ([jwt-go](https://github.com/dgrijalva/jwt-go)) | tokenware |

#### Adaptors
| Functionality | Package |
|:--------------|:-------:|
| Compatibility with [httprouter](https://github.com/julienschmidt/httprouter) | routeradp |

#### Middleware Composition
Middleware can be composed together for easy use:
```go
func main() {
    // MustCompose chains together middleware. It will panic if middleware
    // dependencies are not met.
    m := ctxware.MustCompose(
        errorware.New(),
        logware.NewErrLogger(nil),
        logware.NewReqLogger(nil),
        contentware.NewRespType(contentware.JsonAndXml),
    )

    http.ListenAndServe("localhost:8080", m.ThenFunc(handle))
}

func handle(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
    resp := &struct {
        Greeting string `json:"greeting", "xml": greeting`
    }{"Hi there!"}
    // Use the content type that was specified by the 'Accept' header.
    t := contentware.ResponseTypeFromCtx(ctx)
    t.MarshalWrite(w, resp)
    return nil
}
```
