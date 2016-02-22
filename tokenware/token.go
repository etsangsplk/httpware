/*
Package tokenware provides middleware for decoding & verifying Json Web Tokens
(JWT's) from http requests. It implements the httpware.Middleware interface for
easy composition with other middleware.
*/
package tokenware

import (
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/nstogner/httpware"
	"github.com/nstogner/httpware/httperr"
	"golang.org/x/net/context"
)

type Config struct {
	// The secret should be the same that was used to sign the token.
	Secret interface{}
}

// TokenFromCtx retrieves the decoded JWT.
func TokenFromCtx(ctx context.Context) *jwt.Token {
	return ctx.Value(httpware.TokenKey).(*jwt.Token)
}

// tokenware.Middle parses the JWT in the 'Authorization' header. It will
// return an 'Unauthorized' response if the token is missing or invalid.
type Middle struct {
	conf Config
}

// New returns a new instance of the middleware.
func New(conf Config) *Middle {
	return &Middle{
		// Note: A config struct is used here so that backwards compatibility
		// can be maintained in the API if new fields need to be added later.
		conf: conf,
	}
}

func (m *Middle) Contains() []string { return []string{"github.com/nstogner/tokenware"} }
func (m *Middle) Requires() []string { return []string{"github.com/nstogner/errorware"} }

func (m *Middle) Handle(next httpware.Handler) httpware.Handler {
	return httpware.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		token, err := jwt.ParseFromRequest(
			r,
			func(token *jwt.Token) (interface{}, error) {
				return m.conf.Secret, nil
			},
		)

		if err == nil && token.Valid {
			newCtx := context.WithValue(ctx, httpware.TokenKey, token)
			return next.ServeHTTPContext(newCtx, w, r)
		} else {
			return httperr.New("invalid token", http.StatusUnauthorized)
		}
	})
}
