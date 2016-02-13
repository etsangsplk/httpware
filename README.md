# ctxware

#### Golang HTTP middleware using net/context

**NOTE: CURRENTLY UNDER HEAVY DEVELOPMENT !!!**

This repository contains a collection of stackable middleware packages which aid in writing http handlers in Go. All middleware is of the form:
```Go
func(context.Context, http.ResponseWriter, *http.Request) error
```
This repository contains two groups of "ware":
* Middleware (mdl):
  * Add singular functionality
  * May depend on other middleware
* Adaptorware (adp)
  * Adapt other forms of handlers to the above form

#### Middleware
| Functionality | Package |
|:--------------|:--------:|
| Parsing Request & Response Content Types | contentmdl |
| Unmarshalling (json/xml) and validating entities | entitymdl |
| Handling errors | errormdl |
| JWT authentication ([jwt-go](https://github.com/dgrijalva/jwt-go)) | tokenmdl |
| Logging ([logrus](https://github.com/Sirupsen/logrus)) | logmdl |
#### Adaptorware
| Functionality | Package |
|:--------------|:--------:|
| Compatibility with [httprouter](https://github.com/julienschmidt/httprouter) | routeradp |
