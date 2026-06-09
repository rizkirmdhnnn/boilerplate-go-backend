package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// ErrMiss is returned when a key is not found in cache. Useful in handlers
// to distinguish "not cached" from other errors without importing redis.
var ErrMiss = errors.New("cache: key not found")

// ensure RedisCache implements Cache.
var _ Cache = (*RedisCache)(nil)

// RedisCache implements Cache backed by Redis.
type RedisCache struct {
	client *redis.Client
}

// Config holds Redis connection parameters.
type Config struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// NewRedis creates a RedisCache and verifies the connection (PING).
// When cfg.Host is empty it returns nil, nil — so callers can gracefully
// skip caching when Redis is not configured.
func NewRedis(cfg Config) (*RedisCache, error) {
	if cfg.Host == "" {
		return nil, nil
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("cache: redis ping failed: %w", err)
	}

	log.Info().Str("addr", addr).Int("db", cfg.DB).Msg("cache: connected to Redis")
	return &RedisCache{client: client}, nil
}

// Get returns the string value for key, or ErrMiss if missing.
func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrMiss
	}
	if err != nil {
		return "", fmt.Errorf("cache: get %s: %w", key, err)
	}
	return val, nil
}

// Set stores a string value with the given TTL.
func (r *RedisCache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	if err := r.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("cache: set %s: %w", key, err)
	}
	return nil
}

// Del removes one or more keys.
func (r *RedisCache) Del(ctx context.Context, keys ...string) error {
	if err := r.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("cache: del %v: %w", keys, err)
	}
	return nil
}

// Exists checks if a key exists.
func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("cache: exists %s: %w", key, err)
	}
	return n > 0, nil
}

// Close closes the Redis connection.
func (r *RedisCache) Close() error {
	if r.client == nil {
		return nil
	}
	return r.client.Close()
}
