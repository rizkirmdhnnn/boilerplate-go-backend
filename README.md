# Boilerplate Go Backend

Modern Go backend with **Clean Architecture** using the **Gin** framework.

## Architecture

```
┌──────────────────────────────────────────────┐
│           cmd/api/main.go                    │
│         (composition root / DI)              │
├──────────────────────────────────────────────┤
│               router/                        │
│            (routes + wiring)                 │
├───────────────┬──────────────────┬───────────┤
│   handler/    │   middleware/    │ repository│
│ (HTTP adapter)│  (HTTP concern)  │ (DB impl) │
├───────────────┴──────────────────┴───────────┤
│              application/                    │
│           (use cases / service)              │
├──────────────────────────────────────────────┤
│                domain/                       │
│    (entities + port interfaces)  ← 0 deps   │
└──────────────────────────────────────────────┘
```

**Dependency rule:** Dependencies flow **inward**. Nothing in an inner circle imports from an outer circle.

| Layer | Depends On | Purpose |
|-------|-----------|---------|
| `domain/` | nothing | Entities + repository interfaces (ports) |
| `application/` | `domain/` only | Use cases, business logic, DTOs |
| `handler/` | `application/` | HTTP adapter — converts requests/responses |
| `repository/` | `domain/` | Implements repository ports (PostgreSQL) |
| `middleware/` | `application/` | HTTP concerns — recovery, logging, CORS, security headers, auth, error handler, rate limiter |
| `cmd/api/` | everything | Composition root, DI wiring |

## Tech Stack

- **Framework:** [Gin](https://github.com/gin-gonic/gin)
- **Database:** PostgreSQL via [pgx v5](https://github.com/jackc/pgx/v5)
- **Migrations:** [golang-migrate](https://github.com/golang-migrate/migrate)
- **Auth:** JWT (HS256) via [golang-jwt](https://github.com/golang-jwt/jwt)
- **Logging:** [zerolog](https://github.com/rs/zerolog)
- **Config:** Environment-based via [godotenv](https://github.com/joho/godotenv)
- **Cache:** [Redis](https://github.com/redis/go-redis) with noop/in-memory fallback
- **Hot Reload:** [air](https://github.com/air-verse/air)
- **Rate Limiter:** Custom sliding-window per IP

## Project Structure

```
├── cmd/api/              # Entry point — dependency injection
├── internal/
│   ├── domain/           # Entities + port interfaces (pure Go)
│   ├── application/      # Use cases, DTOs, JWT helpers
│   ├── handler/          # HTTP handlers (adapters)
│   ├── middleware/        # Gin middleware (auth, CORS, logging, error handler, rate limiter, security headers)
│   ├── repository/       # PostgreSQL adapter (implements domain ports)
│   └── router/           # Route definitions
├── migrations/           # SQL migrations (golang-migrate format)
├── pkg/
│   ├── cache/            # Cache interface + Redis / noop / in-memory implementations
│   ├── database/         # Database connection pool
│   ├── migrator/         # Auto-run migrations on startup
│   └── response/         # Standard API response helpers
├── .air.toml             # Hot reload config
├── .env.example
├── Dockerfile
├── docker-compose.yml
└── Makefile
```

## Quick Start

```bash
# Copy env
cp .env.example .env
# Edit .env with your DB config

# Local
make run

# Docker
make docker-up
```

## API Documentation

This project uses **swaggo/swag** to auto-generate OpenAPI docs from Go annotations.

```bash
make swag              # regenerate docs after adding annotations
```

👉 See **[SWAGGER.md](SWAGGER.md)** for full guide — annotations reference, parameter types, DTO auto-docs, and how it works.

**Swagger UI:** `http://localhost:8080/swagger/index.html` (after starting the server)

## Database Migrations

Migrations use [golang-migrate](https://github.com/golang-migrate/migrate) and run **automatically on startup** — no manual `migrate up` needed.

SQL files live in `migrations/` with naming format `{sequence}_{name}.up.sql` / `{sequence}_{name}.down.sql`.

### Create a New Migration

```bash
make migrate-create NAME=create_todos
```

Generates:
```
migrations/
├── 000002_create_todos_table.up.sql   ← write your CREATE TABLE here
└── 000002_create_todos_table.down.sql ← write DROP TABLE here
```

### Manual Commands

```bash
make migrate-up      # Run all pending migrations
make migrate-down    # Rollback last migration
```

### Auto-run on Startup

When the server starts (`make run`), `pkg/migrator/migrator.go` automatically runs any pending migrations using golang-migrate's Go library. Tracked in DB via `schema_migrations` table.

## Hot Reload

Development mode uses [air](https://github.com/air-verse/air) — automatically restarts the server when `.go` files change.

```bash
# Install (one-time)
go install github.com/air-verse/air@latest

# Start with hot reload
make watch
```

Config in `.air.toml` — watches `cmd/api/`, `internal/`, and `pkg/`; ignores tests and `tmp/`.

## Error Handler

Centralized error handling via `middleware.ErrorHandler()`. Maps domain errors to appropriate HTTP status codes:

| Domain Error | HTTP Status |
|-------------|-------------|
| `ErrNotFound` | 404 |
| `ErrUnauthorized` | 401 |
| `ErrForbidden` | 403 |
| `ErrConflict` | 409 |
| Other | 500 |

Handlers just return domain errors — middleware converts them to consistent JSON responses automatically.

## Rate Limiter

Configurable per-IP sliding window rate limiter. Protects API from abuse.

| Env Variable | Default | Description |
|--------------|---------|-------------|
| `RATE_LIMIT_ENABLED` | `true` | Enable/disable |
| `RATE_LIMIT_REQUESTS_PER_MIN` | `100` | Max requests per minute per IP |
| `RATE_LIMIT_BURST` | `20` | Max burst capacity |

When limit is exceeded, returns `429 Too Many Requests` with `Retry-After` header.

```bash
# Disable rate limiter (e.g. for local dev)
RATE_LIMIT_ENABLED=false make run
```

## Caching

Optional Redis-backed caching layer with automatic noop fallback.

| Env Variable | Default | Description |
|--------------|---------|-------------|
| `REDIS_ENABLED` | `false` | Enable Redis cache |
| `REDIS_HOST` | `localhost` | Redis host |
| `REDIS_PORT` | `6379` | Redis port |
| `REDIS_PASSWORD` | | Redis password |
| `REDIS_DB` | `0` | Redis database index |

```bash
# Enable Redis in .env
REDIS_ENABLED=true
```

When Redis is disabled, a **noop cache** is used — handlers work normally without caching. Perfect for local dev without Redis running.

### Cache Interface

```go
type Cache interface {
    Get(ctx context.Context, key string) (string, error)
    Set(ctx context.Context, key, value string, ttl time.Duration) error
    Del(ctx context.Context, keys ...string) error
    Exists(ctx context.Context, key string) (bool, error)
    Close() error
}
```

Any cache backend (Redis, in-memory, custom) that implements this interface can be injected.

### Demo: User List Caching

`GET /api/v1/users` is already wired — results are cached for **30 seconds**. Cache key format: `users:list:p{page}:pp{perPage}`.

To add caching to another handler:

```go
// In your handler
cacheKey := "my:cache:key"
if cached, err := h.cache.Get(ctx, cacheKey); err == nil {
    c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(cached))
    return
}

// ... fetch data ...

data, _ := json.Marshal(result)
h.cache.Set(ctx, cacheKey, string(data), 30*time.Second)
c.JSON(http.StatusOK, result)
```

## API Endpoints

**Public:**
| Method | Path | Description |
|--------|------|-------------|
| GET | /health | Health check |
| POST | /api/v1/auth/register | Register |
| POST | /api/v1/auth/login | Login |

**Protected (Bearer token):**
| Method | Path | Description |
|--------|------|-------------|
| GET | /api/v1/users/me | Current profile |
| GET | /api/v1/users | List (paginated) |
| GET | /api/v1/users/:id | Get by ID |
| PUT | /api/v1/users/:id | Update |
| DELETE | /api/v1/users/:id | Delete |

## Commands

```bash
make run          # Start server
make build        # Build binary
make test         # Run tests (+ race detector, coverage)
make lint         # Run golangci-lint
make swag         # Regenerate Swagger docs (after adding annotations)
make generate     # Alias for make swag
make coverage     # Open HTML coverage report
make migrate-create NAME=xxx  # Create new migration file
make migrate-up   # Run pending migrations
make migrate-down # Rollback last migration
make watch        # Hot reload (auto-restart on file change)
make docker-up    # Docker Compose up (api + postgres)
make docker-down  # Docker Compose down
```
