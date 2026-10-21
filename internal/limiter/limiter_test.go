package limiter_test

import (
	"context"
	"testing"
	"time"

	"github.com/rafael/ratelimiter/internal/limiter"
	"github.com/rafael/ratelimiter/internal/storage"
)

func newLimiter() *limiter.Limiter {
	return limiter.New(storage.NewMemory(), limiter.Config{
		IPLimit:       3,
		DefaultToken:  5,
		TokenLimits:   map[string]int{"vip": 10},
		BlockDuration: 200 * time.Millisecond,
	})
}

func hits(l *limiter.Limiter, ip, token string, n int) (allowed int) {
	for i := 0; i < n; i++ {
		if ok, _ := l.Allow(context.Background(), ip, token); ok {
			allowed++
		}
	}
	return
}

func TestIPLimit(t *testing.T) {
	if got := hits(newLimiter(), "1.1.1.1", "", 10); got != 3 {
		t.Fatalf("allowed = %d, want 3", got)
	}
}

func TestTokenOverridesIP(t *testing.T) {
	l := newLimiter()
	// Same IP, but token limit (10) must apply instead of IP limit (3).
	if got := hits(l, "1.1.1.1", "vip", 20); got != 10 {
		t.Fatalf("allowed = %d, want 10", got)
	}
}

func TestDefaultTokenLimit(t *testing.T) {
	if got := hits(newLimiter(), "1.1.1.1", "other", 20); got != 5 {
		t.Fatalf("allowed = %d, want 5", got)
	}
}

func TestBlockPersistsThenExpires(t *testing.T) {
	l := newLimiter()
	hits(l, "2.2.2.2", "", 4) // exceeds limit -> blocked
	time.Sleep(1100 * time.Millisecond)
	// The 1s window has passed but the block (200ms) too; new window allowed.
	if got := hits(l, "2.2.2.2", "", 1); got != 1 {
		t.Fatalf("expected request allowed after block expired")
	}

	l2 := limiter.New(storage.NewMemory(), limiter.Config{IPLimit: 1, BlockDuration: time.Hour})
	hits(l2, "3.3.3.3", "", 2)
	time.Sleep(1100 * time.Millisecond)
	if got := hits(l2, "3.3.3.3", "", 1); got != 0 {
		t.Fatalf("expected IP still blocked after window reset")
	}
}
