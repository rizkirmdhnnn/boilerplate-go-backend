package cache

import (
	"context"
	"sync"
	"time"
)

// ensure NoopCache implements Cache.
var _ Cache = (*NoopCache)(nil)

// NoopCache is a no-op fallback that does nothing.
// Used when Redis is disabled — handlers can safely call it without nil checks.
type NoopCache struct{}

func NewNoop() *NoopCache { return &NoopCache{} }

func (n *NoopCache) Get(_ context.Context, _ string) (string, error) { return "", ErrMiss }
func (n *NoopCache) Set(_ context.Context, _, _ string, _ time.Duration) error { return nil }
func (n *NoopCache) Del(_ context.Context, _ ...string) error                  { return nil }
func (n *NoopCache) Exists(_ context.Context, _ string) (bool, error)          { return false, nil }
func (n *NoopCache) Close() error                                             { return nil }

// InMemoryCache is a simple sync.Map-backed cache for local dev / testing.
// Not for production use — use Redis instead.
type InMemoryCache struct {
	mu   sync.RWMutex
	data map[string]inMemItem
}

type inMemItem struct {
	value   string
	expires time.Time
}

func NewInMemory() *InMemoryCache {
	return &InMemoryCache{data: make(map[string]inMemItem)}
}

func (m *InMemoryCache) Get(_ context.Context, key string) (string, error) {
	m.mu.RLock()
	item, ok := m.data[key]
	m.mu.RUnlock()
	if !ok || (!item.expires.IsZero() && time.Now().After(item.expires)) {
		if ok {
			m.Del(context.Background(), key)
		}
		return "", ErrMiss
	}
	return item.value, nil
}

func (m *InMemoryCache) Set(_ context.Context, key, value string, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	var expires time.Time
	if ttl > 0 {
		expires = time.Now().Add(ttl)
	}
	m.data[key] = inMemItem{value: value, expires: expires}
	return nil
}

func (m *InMemoryCache) Del(_ context.Context, keys ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, k := range keys {
		delete(m.data, k)
	}
	return nil
}

func (m *InMemoryCache) Exists(_ context.Context, key string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	item, ok := m.data[key]
	if !ok {
		return false, nil
	}
	if !item.expires.IsZero() && time.Now().After(item.expires) {
		return false, nil
	}
	return true, nil
}

func (m *InMemoryCache) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = nil
	return nil
}
