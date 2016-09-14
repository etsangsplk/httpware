package httpware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestErrorHandler(t *testing.T) {
	hdlr := Compose(DefaultErrHandler).Then(
		HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			var err Err
			switch r.URL.Path {
			case "/bad-req":
				err = Err{
					StatusCode: http.StatusBadRequest,
					Message:    "better luck next time",
				}
			case "/panic":
				panic("ahhh")
			default:
				t.Fatal("this point should not be reached")
			}
			return err
		}),
	)

	// Test returning an Err.
	rec := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "http://testing/bad-req", nil)
	if err != nil {
		t.Fatal(err)
	}
	hdlr.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status code: %v, got: %v", http.StatusBadRequest, rec.Code)
	}

	// Test panicing.
	rec = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "http://testing/panic", nil)
	if err != nil {
		t.Fatal(err)
	}
	hdlr.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status code: %v, got: %v", http.StatusInternalServerError, rec.Code)
	}
	expected := `{"message":"` + http.StatusText(http.StatusInternalServerError) + `"}` + "\n"
	if got := rec.Body.String(); got != expected {
		t.Fatalf("expected response body: %s, got: %s", expected, got)
	}
}
