// Package storage has the limiter.Store implementations.
package storage

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Atomic: blocked -> deny; otherwise count the hit in a 1s window and block on excess.
var hitScript = redis.NewScript(`
if redis.call('EXISTS', KEYS[2]) == 1 then
  return 0
end
local count = redis.call('INCR', KEYS[1])
if count == 1 then
  redis.call('PEXPIRE', KEYS[1], 1000)
end
if count > tonumber(ARGV[1]) then
  redis.call('SET', KEYS[2], 1, 'PX', ARGV[2])
  return 0
end
return 1
`)

type Redis struct {
	client *redis.Client
}

func NewRedis(client *redis.Client) *Redis {
	return &Redis{client: client}
}

func (r *Redis) Hit(ctx context.Context, key string, limit int, blockFor time.Duration) (bool, error) {
	res, err := hitScript.Run(ctx, r.client,
		[]string{"rl:count:" + key, "rl:block:" + key},
		limit, blockFor.Milliseconds()).Int()
	if err != nil {
		return false, err
	}
	return res == 1, nil
}
