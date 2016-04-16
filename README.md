# httpware

#### DESCRIPTION
This repository contains a collection of stackable middleware packages which aid in writing http handlers in Go. It also includes a set of functions (inspired by [alice](https://github.com/justinas/alice)) which make composing middleware as simple as possible. All handlers implement the following function:
```Go
ServeHTTPCtx(context.Context, http.ResponseWriter, *http.Request) error
```
This type of http handler was inspired by several Go blog posts: [net/context](https://blog.golang.org/context) and [error-handling](https://blog.golang.org/error-handling-and-go).

**NOTE: Currently under development.**
**Stay tuned for a stable release.**

#### MIDDLEWARE PACKAGES
| Functionality | Package |
|:--------------|:-------:|
| Parsing request & response content types | contentware |
| Enabling CORS | corsware |
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
	// Compose chains together middleware.
	m := httpware.Compose(
		contentware.New(contentware.Defaults),
		logware.New(logware.Defaults),
	)

	http.ListenAndServe("localhost:8080", m.ThenFunc(handle))
}

// handle is meant to demonstrate a POST or PUT endpoint.
func handle(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	u := &user{}
	rqt := contentware.ResponseTypeFromCtx(ctx)
	// Decode from JSON or XML based on the 'Content-Type' header.
	if err := rqt.Decode(r.Body, u); err != nil {
		return httperr.New("could not parse body: "+err.Error(), http.StatusBadRequest)
	}

	if err := u.validate(); err != nil {
		return httperr.New("invalid entity", http.StatusBadRequest).WithField("invalid", err.Error())
	}

	// Store user to db here.

	rst := contentware.ResponseTypeFromCtx(ctx)
	// Write the user back in the response as JSON or XML based on the
	// 'Accept' header.
	rst.Encode(w, u)
	return nil
}

type user struct {
	ID    string `json:"id" xml:"id"`
	Email string `json:"email" xml:"email"`
}

func (u *user) validate() error {
	if u.ID == "" {
		return errors.New("field 'id' must not be empty")
	}
	return nil
}
```

#### COMPOSITIONS
Middleware can be chained into composites:
```go
    m1 := httpware.MustCompose(
        errorware.New(errorware.Defaults),
        logware.New(logware.Defaults),
    )
```
Composites can be further chained:
```go
    m2 := httpware.Compose(m1, contentware.New(contentware.Defaults))
```
... which is equivalent to:
```go
    m2 := m1.With(contentware.New(contentware.Defaults))
```

#### ADAPTORS
Middleware can be adapted for use with different routers. For example, httprouter:
```go
main() {
    m := httpware.Compose(
        errorware.New(errorware.Defaults),
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
