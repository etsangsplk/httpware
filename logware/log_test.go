package logware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nstogner/httpware"
	"github.com/nstogner/httpware/httperr"
	"golang.org/x/net/context"
)

func TestLog(t *testing.T) {
	var buffer bytes.Buffer
	conf := Defaults
	conf.Logger.Out = &buffer
	m := httpware.MustCompose(
		New(conf),
	)
	s := httptest.NewServer(m.ThenFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		if r.URL.Path == "/400" {
			return httperr.New("didnt like your request", http.StatusBadRequest)
		}
		if r.URL.Path == "/500" {
			return httperr.New("ahhhh it blew up", http.StatusInternalServerError)
		}
		return nil
	}))

	cases := []struct {
		Path     string
		Expected string
	}{
		{
			Path:     "/",
			Expected: "served request",
		},
		{
			Path:     "/400",
			Expected: "didnt like your request",
		},
		{
			Path:     "/500",
			Expected: "ahhhh it blew up",
		},
	}
	for _, c := range cases {
		http.Get(s.URL + c.Path)
		got := buffer.String()
		if !strings.Contains(got, c.Expected) {
			t.Fatalf("expected log output to contain: '%s', got: \n%s", c.Expected, got)
		}
	}
}
