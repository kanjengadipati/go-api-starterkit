package middleware

import (
	"sync"
	"time"
)

type RateLimitStore interface {
	Allow(key string, limit int, window time.Duration, now time.Time) (bool, time.Time)
}

// InMemoryRateLimitStore is suitable for a single app instance ONLY.
// WARNING: This implementation is NOT suitable for multi-instance/distributed deployments
// as rate limits are not shared across instances. With REDIS_URL set, appsetup.newRateLimitStore
// wires a Redis-backed implementation instead; see internal/appsetup/rate_limit_store.go.
type InMemoryRateLimitStore struct {
	mu      sync.Mutex
	entries map[string]rateLimitEntry
}

func NewInMemoryRateLimitStore() *InMemoryRateLimitStore {
	return &InMemoryRateLimitStore{
		entries: make(map[string]rateLimitEntry),
	}
}

func (s *InMemoryRateLimitStore) Allow(key string, limit int, window time.Duration, now time.Time) (bool, time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cleanupExpired(now)

	entry, ok := s.entries[key]
	if !ok || now.After(entry.expiresAt) {
		s.entries[key] = rateLimitEntry{
			count:     1,
			expiresAt: now.Add(window),
		}
		return true, s.entries[key].expiresAt
	}

	if entry.count >= limit {
		return false, entry.expiresAt
	}

	entry.count++
	s.entries[key] = entry
	return true, entry.expiresAt
}

func (s *InMemoryRateLimitStore) cleanupExpired(now time.Time) {
	for key, entry := range s.entries {
		if now.After(entry.expiresAt) {
			delete(s.entries, key)
		}
	}
}
