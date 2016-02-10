# ctxware

#### Golang HTTP middleware using net/context
**NOTE: CURRENTLY UNDER HEAVY DEVELOPMENT**

This repository contains a collection of stackable middleware packages which aid in writing http handlers in Go. All middleware is of the form:
```Go
func(context.Context, http.ResponseWriter, *http.Request) 
```
This repository contains three groups of "ware":
* Middleware (mdl):
  * Add singular functionality
  * May depend on other middleware
* Adaptorware (adp)
  * Adapt other forms of handlers to the above form
* Easyware (ezy)
  * Consist of reasonable combinations of middleware and adaptorware

#### Middleware
| Functionality | Package |
|:--------------|:--------:|
| Parsing Request & Response Content Types | contentmdl |
| Unmarshalling (json/xml) and validating entities | entitymdl |
| Handling errors | errormdl |
| JWT authentication ([jwt-go](https://github.com/dgrijalva/jwt-go)) | tokenmdl |
| Logging ([logrus](https://github.com/Sirupsen/logrus)) | logmdl |
| [BoltDB](https://github.com/boltdb/bolt) persistence | boltmdl |
#### Adaptorware
| Functionality | Package |
|:--------------|:--------:|
| Compatibility with net/http | httpadp |
| Compatibility with [httprouter](https://github.com/julienschmidt/httprouter) | routeradp |
#### Easyware
| Functionality | Package |
|:--------------|:--------:|
| Reasonable middleware adapted for [httprouter](https://github.com/julienschmidt/httprouter) | routerezy |
| Reasonable middleware adapted for [boltdb](https://github.com/boltdb/bolt) | routerezy |

#### Example: Easyware (routerezy)
The routerezy package includes an opinionated set of functions (middleware compositions) which should cover most use cases. These functions are adapted for use with the [httprouter](https://github.com/julienschmidt/httprouter) package:
```Go
type User struct {
    Id string `json:"id" xml:"id"`
    Name string `json:"name" xml:"name"`
}

func main() {

    // Define the user entity.
    userDef := entitymdl.Definition{
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
    r.GET("/:id", routerezy.Get(handleGet))
    r.POST("/", routerezy.Post(handlePost, userDef))
    http.ListenAndServe(":8080", r)
}
```
The use of such middleware allows for less cluttered handler functions:
```Go
func handleGet(ctx context.Context, w http.ResponseWriter, r *http.Request) {
    params := routermdl.ParamsFromCtx(ctx)
    usrId := params["id"]

    // Usually this would be a db call...
    u := &User{Id: usrId, Name: "sammy"}

    ct := contentmdl.ResponseTypeFromCtx(ctx)
    ct.MarshalWrite(w, u)
}

func handlePost(ctx context.Context, w http.ResponseWriter, r *http.Request) {
    u := contentmdl.EntityFromCtx(ctx).(*User)

    // Store u in a database here.
    
    w.WriteHeader(http.StatusCreated)
    rt := contentmdl.ResponseTypeFromCtx(ctx)
    rt.MarshalWrite(w, u)
}
```
