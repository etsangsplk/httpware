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
| Logging ([logrus](https://github.com/Sirupsen/logrus)) | logware |
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
        contentware.New(contentware.Defaults),
        errorware.New(errorware.Defaults),
        logware.New(logware.Defaults),
    )

    http.ListenAndServe("localhost:8080", m.ThenFunc(handle))
}

func handle(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
    resp := &struct {
        Greeting string `json:"greeting" xml:"greeting"`
    }{"Hi there!"}
    
    // middleware passes data via the context variable.
    t := contentware.ResponseTypeFromCtx(ctx)
    
    // t is the content type that was set by the contentware package. In this case
    // The middleware took care of determining the type by inspecting the 'Accept'
    // header.
    t.MarshalWrite(w, resp)
    return nil
}
```

#### COMPOSITIONS
Middleware can be chained into composites:
```go
    m1 := httpware.MustCompose(
        errorware.New(),
        logware.New(logware.Defaults),
    )
```
Composites can be further chained:
```go
    m2 := httpware.MustCompose(m1, contentware.New(contentware.Defaults))
```
... which is equivalent to:
```go
    m2 := m1.With(contentware.New(contentware.Defaults))
```

#### ADAPTORS
Middleware can be adapted for use with different routers. For example, httprouter:
```go
main() {
    m := httpware.MustCompose(
        errorware.New(),
        logware.New(logware.Defaults),
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
#### CREATING MIDDLEWARE
Any middleware that follows the httpware.Middleware interface can be composed using the above methods.
```go
type Middleware interface {
    Contains() []string
    Requires() []string
    Handle(Handler) Handler
}
```
The Contains() method is used to identify the middleware for dependency management. This should contain the fully qualified package name:
```go
[]string{"github.com/nstogner/contentware"}
```
The Requires() method is used to define what upstream middleware is relied on. This is enforced when the composition functions are called.
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
