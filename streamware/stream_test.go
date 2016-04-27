package streamware

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nstogner/httpware"
	"golang.org/x/net/context"
)

func TestStreaming(t *testing.T) {
	m := httpware.Compose(
		httpware.DefaultErrHandler,
		New(Defaults),
	)
	s := httptest.NewServer(m.ThenFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		sender := SenderFromCtx(ctx)
		for {
			if err := sender.Send("hello"); err != nil {
				t.Fatal("error sending message: ", err)
			}
		}
		return nil
	}))

	resp, err := http.Get(s.URL)
	if err != nil {
		t.Fatal(err)
	}
	scanner := bufio.NewScanner(resp.Body)
	i := 1
	var total string
	for scanner.Scan() {
		// Scanner.Text() will pull out everything up to the new-line
		// character. In this case we add it back.
		total = total + scanner.Text() + "\n"
		// There should be exactly 4 new-line characters sent in the
		// for 2 messages.
		if i == 4 {
			if total != "data: hello\n\ndata: hello\n\n" {
				t.Fatalf("unexpected messages: %s", total)
			} else {
				break
			}
		}
		i++
	}
}
