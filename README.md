# netmiddle
##### Golang HTTP middleware using net context

This repository uses middleware of the form:
```Go
func(context.Context, http.ResponseWriter, *http.Request) error
```

##### Middleware Functionality
|  Package   | Functionality |
|:----------:|:-------------:|
| contentctx | Parsing Request Content-Type |
| contentctx | Setting Response Content-Type from Accept header |
| contentctx | Marshalling/Unmarshalling responses (JSON & XML) |
| errorctx   | Handling http errors |
| routerctx  | Adapting [httprouter](https://github.com/julienschmidt/httprouter) parameters to the above form |
| tokenctx   | JWT authentication |

##### Reasonable Middleware Compositions (simplectx)
The simplectx package includes a set of functions where the middleware has already been composed into reasonable combinations.
For example, using [httprouter](https://github.com/julienschmidt/httprouter):
```Go
type User struct {
    Id string `json:"id" xml:"id"`
    Name string `json:"name" xml:"name"`
}

func main() {
    r := httprouter.New()
    r.GET("/:id", simplectx.Get(handleGet))
    r.POST("/", simplectx.Post(handlePost, User{}))
    http.ListenAndServe(":8080", r)
}
```
The use of such middleware allows for less cluttered handler functions:
```Go
func handleGet(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
    ps := routerctx.ParamsFromContext(ctx)

    usrId := ps.ByName("id")
    // Usually this would be a db call...
    u := &User{Id: usrId, Name: "sammy"}

    ct := contentctx.RespContentTypeFromContext(ctx)
    ct.MarshalWrite(w, u)
    return nil
}

func handlePost(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
    u := contentctx.EntityFromContext(ctx).(*User)

    // Store u in a database here.
    
    w.WriteHeader(http.StatusCreated)
    rct := contentctx.RespContentTypeFromContext(ctx)
    rct.MarshalWrite(w, u)
    return nil
}
```
