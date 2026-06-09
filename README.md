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
│   ├── repository/       # PostgreSQL adapter (implements domain ports)
│   └── router/           # Route definitions
├── migrations/           # SQL migrations
├── pkg/
│   ├── database/         # Database connection pool
│   └── response/         # Standard API response helpers
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
make test         # Run tests
make docker-up    # Docker Compose up
make docker-down  # Docker Compose down
```
