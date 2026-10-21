// Package limiter contains the rate limiting business rules, independent of
// HTTP and of any persistence technology.
package limiter

import (
	"context"
	"time"
)

// Store is the persistence Strategy. Any backend (Redis, memory, ...) that
// implements it can be plugged into the Limiter.
type Store interface {
	// Hit registers one request for key within a one second window.
	// If the key is blocked, or the request exceeds limit, it returns
	// allowed=false and (re)starts a block of blockFor duration.
	Hit(ctx context.Context, key string, limit int, blockFor time.Duration) (allowed bool, err error)
}

// Config holds the limiter rules.
type Config struct {
	IPLimit       int            // max requests/second per IP
	DefaultToken  int            // max requests/second for tokens without a specific limit
	TokenLimits   map[string]int // specific max requests/second per token
	BlockDuration time.Duration  // how long an offender stays blocked
}

// Limiter decides whether a request is allowed.
type Limiter struct {
	store Store
	cfg   Config
}

func New(store Store, cfg Config) *Limiter {
	return &Limiter{store: store, cfg: cfg}
}

// Allow applies the rules. When a token is present it takes precedence over
// the IP: only the token limit is evaluated.
func (l *Limiter) Allow(ctx context.Context, ip, token string) (bool, error) {
	if token != "" {
		limit, ok := l.cfg.TokenLimits[token]
		if !ok {
			limit = l.cfg.DefaultToken
		}
		return l.store.Hit(ctx, "token:"+token, limit, l.cfg.BlockDuration)
	}
	return l.store.Hit(ctx, "ip:"+ip, l.cfg.IPLimit, l.cfg.BlockDuration)
}
