# Boilerplate Go Backend

Modern Go backend template using **Gin** framework with clean architecture pattern.

## Tech Stack

- **Framework:** [Gin](https://github.com/gin-gonic/gin)
- **Database:** PostgreSQL via [pgx v5](https://github.com/jackc/pgx/v5)
- **Auth:** JWT (HS256) via [golang-jwt](https://github.com/golang-jwt/jwt)
- **Logging:** [zerolog](https://github.com/rs/zerolog)
- **Config:** Environment-based via [godotenv](https://github.com/joho/godotenv)
- **Migrations:** Raw SQL files

## Project Structure

```
├── cmd/api/           # Entry point
├── internal/
│   ├── config/        # Environment config
│   ├── handler/       # HTTP handlers
│   ├── middleware/     # Gin middleware
│   ├── model/         # Domain models
│   ├── repository/    # Data access layer
│   ├── service/       # Business logic
│   └── router/        # Route setup
├── migrations/        # SQL migrations
└── pkg/
    ├── database/      # DB connection pool
    └── response/      # Standard API response
```

## Quick Start

### Prerequisites

- Go 1.23+
- Docker & Docker Compose (optional)
- PostgreSQL 16 (if running locally)

### Local Development

```bash
# Copy env
cp .env.example .env
# Edit .env with your config

# Run
make run
```

### With Docker

```bash
make docker-up
```

API is available at `http://localhost:8080`.

## API Endpoints

### Public

| Method | Path | Description |
|--------|------|-------------|
| GET | /health | Health check |
| POST | /api/v1/auth/register | Register |
| POST | /api/v1/auth/login | Login |

### Protected (Bearer Token)

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/v1/users/me | Current profile |
| GET | /api/v1/users | List users |
| GET | /api/v1/users/:id | Get user |
| PUT | /api/v1/users/:id | Update user |
| DELETE | /api/v1/users/:id | Delete user |

## Makefile

```bash
make run          # Start server
make build        # Build binary
make test         # Run tests
make lint         # Run linter
make docker-up    # Docker Compose up
make docker-down  # Docker Compose down
make tidy         # Tidy modules
```
