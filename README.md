# httpware

#### UPDATE
This project was started before net/context was integrated into the standard library and made a first class citizen in go 1.7. The standard library now includes the context as a part of the http.Request struct, making the first parameter to the ServeHTTPCtx function this repo uses unnecessary. The repo should be refactored to grab the request context from the Request struct. This leaves the question of whether or not to leave the error return value in the middleware signature (it breaks compatibility with http.Handler, but it is nice b/c it avoids the pitfall of forgetting early returns).

[Read: context has arrived...](https://medium.com/@matryer/context-has-arrived-per-request-state-in-go-1-7-4d095be83bd8)

#### DESCRIPTION
This repository contains a collection of middleware packages which aid in writing http handlers in Go. It also includes a set of functions (inspired by [alice](https://github.com/justinas/alice)) which make composing middleware as simple as possible. All handlers implement the following function:
```Go
ServeHTTPCtx(context.Context, http.ResponseWriter, *http.Request) error
```
This type of http handler was inspired by several Go blog posts: [net/context](https://blog.golang.org/context) and [error-handling](https://blog.golang.org/error-handling-and-go).

**Requires: Go 1.7**

#### MIDDLEWARE PACKAGES
| Functionality | Package |
|:--------------|:-------:|
| Parsing request & response content types | contentware |
| Enabling CORS | corsware |
| Limiting requests | limitware |
| Logging ([logrus](https://github.com/Sirupsen/logrus)) | logware |
| Server Sent Events | streamware |
| JWT authentication ([jwt-go](https://github.com/dgrijalva/jwt-go)) | tokenware |
| Pagination | pageware |

#### ROUTER (ADAPTOR) PACKAGES
| Functionality | Package |
|:--------------|:-------:|
| Compatibility with [httprouter](https://github.com/julienschmidt/httprouter) | routeradapt |

#### EXAMPLE
To fetch this repository run:
```sh
go get github.com/nstogner/httpware
```
Consider the following example where several middleware packages are composed together:
```go
func main() {
	// Compose chains together middleware.
	m := httpware.Compose(
		httpware.DefaultErrHandler,
		contentware.New(contentware.Defaults),
		logware.New(logware.Defaults),
	)

	http.ListenAndServe("localhost:8080", m.ThenFunc(handle))
}

// handle is meant to demonstrate a POST or PUT endpoint.
func handle(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	u := &User{}
	rqt := contentware.ResponseTypeFromCtx(ctx)
	// Decode from JSON or XML based on the 'Content-Type' header.
	if err := rqt.Decode(r.Body, u); err != nil {
		return httpware.NewErr("could not parse body: "+err.Error(), http.StatusBadRequest)
	}

	if err := u.validate(); err != nil {
		return httpware.NewErr("invalid entity", http.StatusBadRequest).WithField("invalid", err.Error())
	}

	// Store user to db here.

	rst := contentware.ResponseTypeFromCtx(ctx)
	// Write the user back in the response as JSON or XML based on the
	// 'Accept' header.
	rst.Encode(w, u)
	return nil
}

type User struct {
	ID    string `json:"id" xml:"id"`
	Email string `json:"email" xml:"email"`
}

func (u *User) validate() error {
	if u.ID == "" {
		return errors.New("field 'id' must not be empty")
	}
	return nil
}
```

#### COMPOSITIONS
Middleware can be chained into composites:
```go
    m1 := httpware.Compose(
        httpware.DefaultErrHandler,
        logware.New(logware.Defaults),
        limitware.New(limitware.Defaults),
    )
```
Composites can be further chained:
```go
    m2 := m1.With(contentware.New(contentware.Defaults))
```

#### ADAPTORS
Middleware can be adapted for use with different routers. For example, httprouter:
```go
main() {
    m := httpware.Compose(
        httpware.DefaultErrHandler,
        logware.New(logware.Defaults),
    )
    r := httprouter.New()
    r.GET("/users/:id", routeradapt.Adapt(m.ThenFunc(handle))
    ...
}

func handle(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
    ps := routeradapt.ParamsFromCtx(ctx)
    id := ps.ByName("id")
    ...
}
```
