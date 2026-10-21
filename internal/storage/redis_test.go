package storage_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"github.com/rafael/ratelimiter/internal/storage"
)

func TestRedisHitAndBlock(t *testing.T) {
	mr := miniredis.RunT(t)
	s := storage.NewRedis(redis.NewClient(&redis.Options{Addr: mr.Addr()}))
	ctx := context.Background()

	for i := 0; i < 2; i++ {
		if ok, err := s.Hit(ctx, "ip:x", 2, time.Minute); err != nil || !ok {
			t.Fatalf("hit %d: ok=%v err=%v", i, ok, err)
		}
	}
	if ok, _ := s.Hit(ctx, "ip:x", 2, time.Minute); ok {
		t.Fatal("third hit should be denied")
	}

	mr.FastForward(2 * time.Second) // window expired, block still active
	if ok, _ := s.Hit(ctx, "ip:x", 2, time.Minute); ok {
		t.Fatal("should still be blocked")
	}
	mr.FastForward(time.Minute)
	if ok, _ := s.Hit(ctx, "ip:x", 2, time.Minute); !ok {
		t.Fatal("block should have expired")
	}
}
