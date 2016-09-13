/*
Package tokenware provides middleware for decoding & verifying Json Web Tokens
(JWT's) from http requests. It implements the httpware.Middleware interface for
easy composition with other middleware.
*/
package tokenware

import (
	"context"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/jriquelme/httpware"
)

// Config is used to initialize a new instance of this middleware.
type Config struct {
	// The secret should be the same that was used to sign the token.
	Secret interface{}
}

// TokenFromCtx retrieves the decoded JWT.
func TokenFromCtx(ctx context.Context) *jwt.Token {
	return ctx.Value(httpware.TokenKey).(*jwt.Token)
}

// Middle parses the JWT in the 'Authorization' header. It will
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

// Handle takes the next handler as an argument and wraps it in this middleware.
func (m *Middle) Handle(next httpware.Handler) httpware.Handler {
	return httpware.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		token, err := request.ParseFromRequest(
			r,
			request.AuthorizationHeaderExtractor,
			func(token *jwt.Token) (interface{}, error) {
				return m.conf.Secret, nil
			},
		)

		if err == nil && token.Valid {
			newCtx := context.WithValue(ctx, httpware.TokenKey, token)
			return next.ServeHTTPCtx(newCtx, w, r)
		}

		// No soup for you.
		return httpware.NewErr("invalid token", http.StatusUnauthorized)
	})
}
