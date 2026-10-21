package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rafael/ratelimiter/internal/limiter"
	"github.com/rafael/ratelimiter/internal/middleware"
	"github.com/rafael/ratelimiter/internal/storage"
)

func do(h http.Handler, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "9.9.9.9:1234"
	if token != "" {
		req.Header.Set("API_KEY", token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestMiddleware429(t *testing.T) {
	l := limiter.New(storage.NewMemory(), limiter.Config{
		IPLimit: 2, TokenLimits: map[string]int{"vip": 5}, BlockDuration: time.Minute,
	})
	h := middleware.RateLimit(l)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))

	for i := 0; i < 2; i++ {
		if c := do(h, "").Code; c != 200 {
			t.Fatalf("req %d: code %d", i, c)
		}
	}
	rec := do(h, "")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("code = %d, want 429", rec.Code)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != middleware.LimitMsg {
		t.Fatalf("body = %q", got)
	}

	// Same IP is blocked, but a token with its own limit still goes through.
	for i := 0; i < 5; i++ {
		if c := do(h, "vip").Code; c != 200 {
			t.Fatalf("token req %d: code %d", i, c)
		}
	}
	if c := do(h, "vip").Code; c != 429 {
		t.Fatalf("token over limit: code %d", c)
	}
}
