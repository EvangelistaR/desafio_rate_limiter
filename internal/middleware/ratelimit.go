// Package middleware adapts the limiter to net/http.
package middleware

import (
	"log"
	"net"
	"net/http"

	"github.com/rafael/ratelimiter/internal/limiter"
)

const (
	TokenHeader = "API_KEY"
	LimitMsg    = "you have reached the maximum number of requests or actions allowed within a certain time frame"
)

func RateLimit(l *limiter.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			allowed, err := l.Allow(r.Context(), clientIP(r), r.Header.Get(TokenHeader))
			if err != nil {
				log.Printf("rate limiter error: %v", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			if !allowed {
				http.Error(w, LimitMsg, http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
