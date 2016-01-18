package jwtctx

import (
	"bluesky/httpcontext"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/net/context"
)

func TokenFromContext(ctx context.Context) jwt.Token {
	return ctx.Value(httpcontext.TokenKey).(jwt.Token)
}

func Auth(next httpcontext.Handler, secret interface{}) httpcontext.Handler {
	return httpcontext.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		token, err := jwt.ParseFromRequest(
			r,
			func(token *jwt.Token) (interface{}, error) {
				return secret, nil
			},
		)

		if err == nil && token.Valid {
			newCtx := context.WithValue(ctx, httpcontext.TokenKey, token)
			next.ServeHTTPContext(newCtx, w, r)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	})
}
