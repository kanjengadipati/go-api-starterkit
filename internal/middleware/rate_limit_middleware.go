package middleware

import (
	"fmt"
	"time"

	"pleco-api/internal/httpx"

	"github.com/gin-gonic/gin"
)

type rateLimitEntry struct {
	count     int
	expiresAt time.Time
}

type RateLimiter struct {
	limit  int
	window time.Duration
	store  RateLimitStore
	now    func() time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return NewRateLimiterWithStore(limit, window, NewInMemoryRateLimitStore())
}

func NewRateLimiterWithStore(limit int, window time.Duration, store RateLimitStore) *RateLimiter {
	if limit < 1 {
		limit = 1
	}
	if window <= 0 {
		window = time.Minute
	}
	if store == nil {
		store = NewInMemoryRateLimitStore()
	}

	return &RateLimiter{
		limit:  limit,
		window: window,
		store:  store,
		now:    time.Now,
	}
}

func (r *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := fmt.Sprintf("%s:%s", c.FullPath(), c.ClientIP())
		if !r.allow(key) {
			httpx.Error(c, 429, "too many requests")
			c.Abort()
			return
		}

		c.Next()
	}
}

func (r *RateLimiter) allow(key string) bool {
	allowed, _ := r.allowAt(key)
	return allowed
}

func (r *RateLimiter) allowAt(key string) (bool, time.Time) {
	return r.store.Allow(key, r.limit, r.window, r.now())
}
