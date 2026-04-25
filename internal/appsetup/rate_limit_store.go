package appsetup

import (
	"context"
	"log"
	"strings"

	"go-api-starterkit/internal/middleware"

	"github.com/redis/go-redis/v9"
)

// newRateLimitStore returns a shared rate-limit backend: Redis when redisURL is set
// and reachable, otherwise the default in-memory store.
func newRateLimitStore(redisURL string) middleware.RateLimitStore {
	url := strings.TrimSpace(redisURL)
	if url == "" {
		return middleware.NewInMemoryRateLimitStore()
	}

	opt, err := redis.ParseURL(url)
	if err != nil {
		log.Printf("rate limit: invalid REDIS_URL, using in-memory store: %v", err)
		return middleware.NewInMemoryRateLimitStore()
	}

	rdb := redis.NewClient(opt)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Printf("rate limit: redis unavailable (%v), using in-memory store", err)
		_ = rdb.Close()
		return middleware.NewInMemoryRateLimitStore()
	}

	log.Println("rate limit: using Redis store (REDIS_URL)")
	return middleware.NewRedisRateLimitStore(rdb)
}
