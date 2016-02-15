package tokenware

import (
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/nstogner/ctxware"
	"github.com/nstogner/ctxware/lib/httperr"
	"golang.org/x/net/context"
)

type Ware struct {
	secret interface{}
}

func New(secret interface{}) Ware {
	return Ware{
		secret: secret,
	}
}

func (w Ware) Contains() []string {
	return []string{"tokenware.Ware"}
}

func (w Ware) Requires() []string {
	return []string{"errorware.Ware"}
}

func TokenFromCtx(ctx context.Context) *jwt.Token {
	return ctx.Value(ctxware.TokenKey).(*jwt.Token)
}

func (ware Ware) Handle(next ctxware.Handler) ctxware.Handler {
	return ctxware.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		token, err := jwt.ParseFromRequest(
			r,
			func(token *jwt.Token) (interface{}, error) {
				return ware.secret, nil
			},
		)

		if err == nil && token.Valid {
			newCtx := context.WithValue(ctx, ctxware.TokenKey, token)
			return next.ServeHTTPContext(newCtx, w, r)
		} else {
			return httperr.New("invalid token", http.StatusUnauthorized)
		}
	})
}
