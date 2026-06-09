# Middleware

## Hot Reload

Development mode uses [air](https://github.com/air-verse/air) — automatically restarts the server when `.go` files change.

```bash
# Install (one-time)
go install github.com/air-verse/air@latest

# Start with hot reload
make watch
```

Config in `.air.toml` — watches `cmd/api/`, `internal/`, and `pkg/`; ignores tests and `tmp/`.

---

## Error Handler

Centralized error handling via `middleware.ErrorHandler()`. Maps domain errors to HTTP status codes:

| Domain Error | HTTP Status |
|-------------|-------------|
| `ErrNotFound` | 404 |
| `ErrUnauthorized` | 401 |
| `ErrForbidden` | 403 |
| `ErrConflict` | 409 |
| Other | 500 |

Handlers just return domain errors — the middleware converts them to consistent JSON responses automatically.

### Usage

```go
// In your handler — just return domain errors
if err != nil {
    if errors.Is(err, domain.ErrNotFound) {
        _ = c.Error(err)
        return
    }
}
```

No manual status code mapping needed in handlers.

---

## Rate Limiter

Configurable per-IP sliding window rate limiter. Protects API from abuse.

### Env Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `RATE_LIMIT_ENABLED` | `true` | Enable/disable |
| `RATE_LIMIT_REQUESTS_PER_MIN` | `100` | Max requests per minute per IP |
| `RATE_LIMIT_BURST` | `20` | Max burst capacity |

```bash
# Disable rate limiter (e.g. for local dev)
RATE_LIMIT_ENABLED=false make run
```

When exceeded, returns `429 Too Many Requests` with `Retry-After` header.

### How It Works

Uses a sliding window algorithm — tracks request timestamps per IP in-memory. Thread-safe via `sync.RWMutex`. Burst allows short spikes above the per-minute limit.

---

## Middleware Stack (order applied)

1. **Recovery** — panic recovery, returns 500
2. **RequestID** — injects/generates request ID header
3. **Logger** — structured request logging via zerolog
4. **CORS** — configurable allowed origins
5. **SecurityHeaders** — common security headers (X-Frame-Options, HSTS, etc.)
6. **ErrorHandler** — centralized error → HTTP status mapping
7. **RateLimiter** — per-IP rate limiting (conditional)
8. **Auth** — JWT Bearer token validation (applied to protected routes only)
