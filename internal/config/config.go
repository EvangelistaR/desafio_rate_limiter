// Package config loads settings from environment variables (or a .env file).
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	IPLimit       int
	DefaultToken  int
	TokenLimits   map[string]int
	BlockDuration time.Duration
}

// Load reads the environment. A .env file in the working dir is optional and
// never overrides variables already set.
func Load() (Config, error) {
	_ = godotenv.Load()

	var err error
	c := Config{
		Port:          getenv("PORT", "8080"),
		RedisAddr:     getenv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
	}
	if c.RedisDB, err = getint("REDIS_DB", 0); err != nil {
		return c, err
	}
	if c.IPLimit, err = getint("RATE_LIMIT_IP", 10); err != nil {
		return c, err
	}
	if c.DefaultToken, err = getint("RATE_LIMIT_TOKEN", 100); err != nil {
		return c, err
	}
	if c.BlockDuration, err = time.ParseDuration(getenv("BLOCK_DURATION", "5m")); err != nil {
		return c, fmt.Errorf("BLOCK_DURATION: %w", err)
	}
	if c.TokenLimits, err = parseTokenLimits(os.Getenv("TOKEN_LIMITS")); err != nil {
		return c, err
	}
	return c, nil
}

// parseTokenLimits parses "tokenA:100,tokenB:50".
func parseTokenLimits(s string) (map[string]int, error) {
	out := map[string]int{}
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		name, val, ok := strings.Cut(part, ":")
		n, err := strconv.Atoi(strings.TrimSpace(val))
		if !ok || err != nil || strings.TrimSpace(name) == "" {
			return nil, fmt.Errorf("TOKEN_LIMITS: invalid entry %q (expected token:limit)", part)
		}
		out[strings.TrimSpace(name)] = n
	}
	return out, nil
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func getint(k string, def int) (int, error) {
	v := os.Getenv(k)
	if v == "" {
		return def, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", k, err)
	}
	return n, nil
}
