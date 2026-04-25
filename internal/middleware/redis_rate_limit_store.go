package middleware

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisRateLimitStore implements a fixed-window counter per key using Redis INCR + TTL.
type RedisRateLimitStore struct {
	rdb    *redis.Client
	prefix string
}

func NewRedisRateLimitStore(rdb *redis.Client) *RedisRateLimitStore {
	if rdb == nil {
		return nil
	}
	return &RedisRateLimitStore{rdb: rdb, prefix: "ratelimit:"}
}

func (s *RedisRateLimitStore) redisKey(key string) string {
	return s.prefix + key
}

func (s *RedisRateLimitStore) Allow(key string, limit int, window time.Duration, now time.Time) (bool, time.Time) {
	ctx := context.Background()
	rkey := s.redisKey(key)

	count, err := s.rdb.Incr(ctx, rkey).Result()
	if err != nil {
		// Fail open so a Redis outage does not hard-block the API.
		return true, now.Add(window)
	}
	if count == 1 {
		_ = s.rdb.Expire(ctx, rkey, window).Err()
	}

	ttl, err := s.rdb.TTL(ctx, rkey).Result()
	expiresAt := now.Add(window)
	if err == nil && ttl > 0 {
		expiresAt = now.Add(ttl)
	}

	return int(count) <= limit, expiresAt
}
