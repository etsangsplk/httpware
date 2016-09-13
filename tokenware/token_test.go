package tokenware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/jriquelme/httpware"
)

func TestWare(t *testing.T) {
	secret := []byte("shh")
	m := httpware.Compose(
		httpware.DefaultErrHandler,
		New(Config{secret}),
	)
	s := httptest.NewServer(m.ThenFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		return nil
	}))

	// Generate token.
	tkn := jwt.New(jwt.SigningMethodHS256)
	tknStr, err := tkn.SignedString(secret)
	if err != nil {
		t.Fatal(err)
	}

	// Unauthorized
	unauthResp, err := http.Get(s.URL)
	if err != nil {
		t.Fatal(err)
	}
	if unauthResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status code %v, got %v", http.StatusUnauthorized, unauthResp.StatusCode)
	}

	// Authorized
	req, err := http.NewRequest("GET", s.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+tknStr)
	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status code %v, got %v", http.StatusOK, resp.StatusCode)
	}
}
