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

type Middle struct {
	secret interface{}
}

func New(secret interface{}) Middle {
	return Middle{
		secret: secret,
	}
}

func (m Middle) Contains() []string { return []string{"github.com/nstogner/tokenware"} }
func (m Middle) Requires() []string { return []string{"github.com/nstogner/errorware"} }

func TokenFromCtx(ctx context.Context) *jwt.Token {
	return ctx.Value(httpware.TokenKey).(*jwt.Token)
}

func (m Middle) Handle(next httpware.Handler) httpware.Handler {
	return httpware.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		token, err := jwt.ParseFromRequest(
			r,
			func(token *jwt.Token) (interface{}, error) {
				return m.secret, nil
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
