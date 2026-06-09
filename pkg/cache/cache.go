package cache

import (
	"context"
	"time"
)

// Cache is the port interface for caching.
// Any cache backend (Redis, in-memory, etc.) must implement this.
type Cache interface {
	// Get returns the value for key, or ErrMiss if not found.
	Get(ctx context.Context, key string) (string, error)

	// Set stores value for key with TTL.
	Set(ctx context.Context, key string, value string, ttl time.Duration) error

	// Del removes one or more keys.
	Del(ctx context.Context, keys ...string) error

	// Exists returns true if key exists.
	Exists(ctx context.Context, key string) (bool, error)

	// Close cleans up the cache backend.
	Close() error
}
