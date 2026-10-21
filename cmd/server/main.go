package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/rafael/ratelimiter/internal/config"
	"github.com/rafael/ratelimiter/internal/limiter"
	"github.com/rafael/ratelimiter/internal/middleware"
	"github.com/rafael/ratelimiter/internal/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr, Password: cfg.RedisPassword, DB: cfg.RedisDB})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis: %v", err)
	}

	// To change the persistence, swap this line for another limiter.Store.
	var store limiter.Store = storage.NewRedis(rdb)

	lim := limiter.New(store, limiter.Config{
		IPLimit:       cfg.IPLimit,
		DefaultToken:  cfg.DefaultToken,
		TokenLimits:   cfg.TokenLimits,
		BlockDuration: cfg.BlockDuration,
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, "ok")
	})

	log.Printf("listening on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, middleware.RateLimit(lim)(mux)))
}
