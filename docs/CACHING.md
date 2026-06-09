# Caching

Optional Redis-backed caching layer with automatic noop fallback.

## Env Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `REDIS_ENABLED` | `false` | Enable Redis cache |
| `REDIS_HOST` | `localhost` | Redis host |
| `REDIS_PORT` | `6379` | Redis port |
| `REDIS_PASSWORD` | | Redis password |
| `REDIS_DB` | `0` | Redis database index |

```bash
# Enable Redis in .env
REDIS_ENABLED=true
```

## Backend Selection

When Redis is disabled (default), a **NoopCache** is used — handlers work normally without caching. Perfect for local dev without Redis running.

When Redis is enabled, the app connects on startup and verifies with PING. If Redis is unreachable, the app **fails fast** to prevent silent data loss.

## Cache Interface

```go
type Cache interface {
    Get(ctx context.Context, key string) (string, error)
    Set(ctx context.Context, key, value string, ttl time.Duration) error
    Del(ctx context.Context, keys ...string) error
    Exists(ctx context.Context, key string) (bool, error)
    Close() error
}
```

Any backend (Redis, in-memory, custom) that implements this interface can be injected via the `pkg/cache.Cache` port interface.

## Example: User List Caching

`GET /api/v1/users` is already wired — results are cached for **30 seconds**.

```go
cacheKey := fmt.Sprintf("users:list:p%d:pp%d", page, perPage)

// Try cache first
if cached, err := h.cache.Get(ctx, cacheKey); err == nil {
    c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(cached))
    return
}

// Fetch from DB (cache miss)
users, total, err := h.svc.List(ctx, page, perPage)

// Build response
resp := response.APIResponse{Success: true, Data: users, Meta: &response.Meta{...}}

// Cache for next time (best-effort)
data, _ := json.Marshal(resp)
h.cache.Set(ctx, cacheKey, string(data), 30*time.Second)

c.JSON(http.StatusOK, resp)
```

Cache key format: `users:list:p{page}:pp{perPage}`

## Adding Caching to Another Handler

```go
// In your handler
cacheKey := "my:cache:key"
if cached, err := h.cache.Get(ctx, cacheKey); err == nil {
    c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(cached))
    return
}

// ... fetch fresh data ...

data, _ := json.Marshal(result)
h.cache.Set(ctx, cacheKey, string(data), 30*time.Second)
c.JSON(http.StatusOK, result)
```
