package tokenctx

import (
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/nstogner/contextware/httpctx"
	"github.com/nstogner/contextware/httperr"
	"golang.org/x/net/context"
)

func TokenFromContext(ctx context.Context) *jwt.Token {
	return ctx.Value(httpctx.TokenKey).(*jwt.Token)
}

func Auth(next httpctx.Handler, secret interface{}) httpctx.Handler {
	return httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		token, err := jwt.ParseFromRequest(
			r,
			func(token *jwt.Token) (interface{}, error) {
				return secret, nil
			},
		)

		if err == nil && token.Valid {
			newCtx := context.WithValue(ctx, httpctx.TokenKey, token)
			return next.ServeHTTPContext(newCtx, w, r)
		} else {
			return httperr.Err{
				StatusCode: http.StatusUnauthorized,
				Message:    "invalid token",
			}
		}
	})
}
