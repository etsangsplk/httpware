# ctxware
**!!! Currently under heavy development !!!**

#### DESCRIPTION
This repository contains a collection of stackable middleware packages which aid in writing http handlers in Go. All handlers implement the following function:
```Go
ServeHTTPContext(context.Context, http.ResponseWriter, *http.Request) error
```
It also includes a set of functions (inspired by [alice](https://github.com/justinas/alice)) which make composing middleware as simple as possible.

#### MIDDLEWARE PACKAGES
| Functionality | Package |
|:--------------|:-------:|
| Handling errors | errorware |
| Logging ([logrus](https://github.com/Sirupsen/logrus)) errors and requests | logware |
| Parsing request & response content types | contentware |
| Unmarshalling (json/xml) and validating entities | entityware |
| JWT authentication ([jwt-go](https://github.com/dgrijalva/jwt-go)) | tokenware |

#### ADAPTOR PACKAGES
| Functionality | Package |
|:--------------|:-------:|
| Compatibility with [httprouter](https://github.com/julienschmidt/httprouter) | routeradp |

#### EXAMPLE
To fetch this repository run:
```sh
go get github.com/nstogner/ctxware
```
Consider the following example where several middleware packages are composed together:
```go
func main() {
    // MustCompose chains together middleware. It will panic if middleware
    // dependencies are not met.
    m := ctxware.MustCompose(
        errorware.New(),
        logware.NewErrLogger(logware.Defaults),
        logware.NewReqLogger(logware.Defaults),
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

#### COMPOSITIONS
Middleware can be chained into composites:
```go
    m1 := ctxware.MustCompose(
        errorware.New(),
        logware.NewErrLogger(logware.Defaults),
    )
```
Composites can be further chained:
```go
    m2 := ctxware.MustCompose(m1, contentware.NewRespType(contentware.JsonAndXml))
```
... which is equivalent to:
```go
    m2 := m1.With(contentware.NewRespType(contentware.JsonAndXml))
```

#### ADAPTORS
Middleware can be adapted for use with different routers. For example, httprouter:
```go
main() {
    m := ctxware.MustCompose(
        errorware.New(),
        logware.NewErrLogger(logware.Defaults),
    )
    r := httprouter.New()
    r.GET("/users/:id", routeradp.Adapt(m.Then(handle))
    ...
}

func handle(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
    ps := routeradp.ParamsFromCtx(ctx)
    id := ps.ByName("id")
    ...
}
```
#### USING NON-NATIVE MIDDLEWARE
Non-native middleware (that which does not implement the ctxware.Middleware interface) can be used in compositions if they adhere to the ctxware.Handler interface:
```go
func someMiddlware(next ctxware.Handler) ctxware.Handler {
    return ctxware.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
        ...
        return next.ServeHTTPContext(ctx, w, r)
    })
}

func main() {
    m := MustCompose(
        ...,
        // Bring someMiddleware in as an anonymous implementation of Middleware (no dependencies).
        ctxware.Anon(someMiddleware),
    )
}
```
