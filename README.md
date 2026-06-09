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
| `middleware/` | `application/` | HTTP concerns (auth, CORS, logging) |
| `cmd/api/` | everything | Composition root, DI wiring |

## Tech Stack

- **Framework:** [Gin](https://github.com/gin-gonic/gin)
- **Database:** PostgreSQL via [pgx v5](https://github.com/jackc/pgx/v5)
- **Auth:** JWT (HS256) via [golang-jwt](https://github.com/golang-jwt/jwt)
- **Logging:** [zerolog](https://github.com/rs/zerolog)
- **Config:** Environment-based via [godotenv](https://github.com/joho/godotenv)

## Project Structure

```
├── cmd/api/              # Entry point — dependency injection
├── internal/
│   ├── domain/           # Entities + port interfaces (pure Go)
│   ├── application/      # Use cases, DTOs, JWT helpers
│   ├── handler/          # HTTP handlers (adapters)
│   ├── middleware/        # Gin middleware (auth, CORS, logging)
|   ├── repository/       # PostgreSQL adapter (implements domain ports)
|   └── router/           # Route definitions
├── migrations/           # SQL migrations (golang-migrate format)
├── pkg/
|   ├── database/         # Database connection pool
|   ├── migrator/         # Auto-run migrations on startup
|   └── response/         # Standard API response helpers
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
