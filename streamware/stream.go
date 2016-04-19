/*
Package streamware provides middleware for Server Sent Events (SSE).
*/
package streamware

import (
	"fmt"
	"net/http"

	"github.com/nstogner/httpware"
	"golang.org/x/net/context"
)

var (
	// Defaults is a reasonable configuration.
	Defaults = Config{}
)

// Config is used to initialize a new instance of Middle.
type Config struct {
}

// Middle is middleware that limits http requests.
type Middle struct {
}

// New creates a new Middle instance.
func New(conf Config) *Middle {
	m := Middle{}
	return &m
}

// Sender is used to send events to the client.
type Sender struct {
	flusher     http.Flusher
	writer      http.ResponseWriter
	CloseNotify <-chan bool
}

// Send sends a single message to the client.
func (s *Sender) Send(msg string) error {
	_, err := fmt.Fprintf(s.writer, "data: %s\n\n", msg)
	if err != nil {
		return err
	}
	s.flusher.Flush()
	return nil
}

// SenderFromCtx retrieves the current Streamer instance.
func SenderFromCtx(ctx context.Context) Sender {
	return ctx.Value(httpware.SenderKey).(Sender)
}

// Handle takes the next handler as an argument and wraps it in this middleware.
func (m *Middle) Handle(next httpware.Handler) httpware.Handler {
	return httpware.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		flusher, ok := w.(http.Flusher)
		if !ok {
			return httpware.NewErr("streaming not supported", http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		sender := Sender{
			flusher:     flusher,
			writer:      w,
			CloseNotify: w.(http.CloseNotifier).CloseNotify(),
		}

		return next.ServeHTTPCtx(context.WithValue(ctx, httpware.SenderKey, sender), w, r)
	})
}
