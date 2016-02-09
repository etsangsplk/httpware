# ctxware

#### Golang HTTP middleware using net/context
**NOTE: CURRENTLY UNDER HEAVY DEVELOPMENT**

This repository is a collection of middleware packages which aid in writing http handlers in Go. All middleware is of the form:
```Go
func(context.Context, http.ResponseWriter, *http.Request) 
```

#### Package Descriptions
| Functionality | Package |
|:--------------|:--------:|
| Parsing Request & Response Content Types | contentctx |
| Unmarshalling (json/xml) and validating entities | entityctx |
| Handling errors | errorctx |
| Compatibility with [httprouter](https://github.com/julienschmidt/httprouter) | routerctx |
| JWT authentication | tokenctx |
| Logging | logctx |
| [BoltDB](https://github.com/boltdb/bolt) persistence | boltctx |
| Reasonable compositions of the above middleware | easyctx |

#### Using the Reasonable Middleware Compositions (easyctx)
The easyctx package includes an opinionated set of functions (middleware compositions) which should cover most use cases. These functions are adapted for use with the [httprouter](https://github.com/julienschmidt/httprouter) package:
```Go
type User struct {
    Id string `json:"id" xml:"id"`
    Name string `json:"name" xml:"name"`
}

func main() {

    // Define the user entity.
    userDef := entityctx.Definition{
        Entity: User{},
        Validate: func(u interface{}) error {
            usr := u.(*User)
            if len(usr.Id) < 5 {
                return errors.New("user id must be at least 5 characters")
            }
            return nil
        },
    }

    r := httprouter.New()
    r.GET("/:id", easyctx.Get(handleGet))
    r.POST("/", easyctx.Post(handlePost, userDef))
    http.ListenAndServe(":8080", r)
}
```
The use of such middleware allows for less cluttered handler functions:
```Go
func handleGet(ctx context.Context, w http.ResponseWriter, r *http.Request) {
    params := routerctx.ParamsFromCtx(ctx)
    usrId := params["id"]

    // Usually this would be a db call...
    u := &User{Id: usrId, Name: "sammy"}

    ct := contentctx.ResponseTypeFromCtx(ctx)
    ct.MarshalWrite(w, u)
}

func handlePost(ctx context.Context, w http.ResponseWriter, r *http.Request) {
    u := contentctx.EntityFromCtx(ctx).(*User)

    // Store u in a database here.
    
    w.WriteHeader(http.StatusCreated)
    rt := contentctx.ResponseTypeFromCtx(ctx)
    rt.MarshalWrite(w, u)
}
```
