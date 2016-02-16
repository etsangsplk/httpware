# httpware

#### DESCRIPTION
This repository contains a collection of stackable middleware packages which aid in writing http handlers in Go. It also includes a set of functions (inspired by [alice](https://github.com/justinas/alice)) which make composing middleware as simple as possible. All handlers implement the following function:
```Go
ServeHTTPContext(context.Context, http.ResponseWriter, *http.Request) error
```
This type of http handler was inspired by several Go blog posts: [net/context](https://blog.golang.org/context) and [error-handling](https://blog.golang.org/error-handling-and-go).

**NOTE: Currently under development.**
**Stay tuned for stable release.**

#### MIDDLEWARE PACKAGES
| Functionality | Package |
|:--------------|:-------:|
| Parsing request & response content types | contentware |
| Enabling CORS | corsware |
| Unmarshalling (json/xml) and validating entities | entityware |
| Handling errors | errorware |
| Limiting requests | limitware |
| Logging ([logrus](https://github.com/Sirupsen/logrus)) errors and requests | logware |
| JWT authentication ([jwt-go](https://github.com/dgrijalva/jwt-go)) | tokenware |

#### OTHER PACKAGES
| Functionality | Package |
|:--------------|:-------:|
| Compatibility with [httprouter](https://github.com/julienschmidt/httprouter) | routeradapt |
| Standardized http errors | httperr |

#### EXAMPLE
To fetch this repository run:
```sh
go get github.com/nstogner/httpware
```
Consider the following example where several middleware packages are composed together:
```go
func main() {
    // MustCompose chains together middleware. It will panic if middleware
    // dependencies are not met.
    m := httpware.MustCompose(
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
    
    // middleware passes data via the context variable.
    t := contentware.ResponseTypeFromCtx(ctx)
    
    // t is the content type that was set by the contentware package. In this case
    // it will be either JSON, or XML as defined above. The middleware took care of
    // determining the type by inspecting the 'Accept' header.
    t.MarshalWrite(w, resp)
    return nil
}
```

#### COMPOSITIONS
Middleware can be chained into composites:
```go
    m1 := httpware.MustCompose(
        errorware.New(),
        logware.NewErrLogger(logware.Defaults),
    )
```
Composites can be further chained:
```go
    m2 := httpware.MustCompose(m1, contentware.NewRespType(contentware.JsonAndXml))
```
... which is equivalent to:
```go
    m2 := m1.With(contentware.NewRespType(contentware.JsonAndXml))
```

#### ADAPTORS
Middleware can be adapted for use with different routers. For example, httprouter:
```go
main() {
    m := httpware.MustCompose(
        errorware.New(),
        logware.NewErrLogger(logware.Defaults),
    )
    r := httprouter.New()
    r.GET("/users/:id", routeradapt.Adapt(m.Then(handle))
    ...
}

func handle(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
    ps := routeradapt.ParamsFromCtx(ctx)
    id := ps.ByName("id")
    ...
}
```
#### USING NON-NATIVE MIDDLEWARE
Non-native middleware (that which does not implement the httpware.Middleware interface) can be used in compositions if they adhere to the httpware.Handler interface:
```go
func someMiddleware(next httpware.Handler) httpware.Handler {
    return httpware.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
        ...
        return next.ServeHTTPContext(ctx, w, r)
    })
}

func main() {
    m := MustCompose(
        ...,
        // Bring someMiddleware in as an anonymous implementation of Middleware (no dependencies).
        httpware.Anon(someMiddleware),
    )
}
```
