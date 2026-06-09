# Boilerplate Go Backend

Modern Go backend with **Clean Architecture** using the **Gin** framework.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/rizkirmdhnnn/boilerplate-go-backend)](https://goreportcard.com/report/github.com/rizkirmdhnnn/boilerplate-go-backend)
[![Go Version](https://img.shields.io/badge/Go-1.25-blue)](https://go.dev/)

## Quick Start

```bash
cp .env.example .env
# Edit .env with your DB config

# Install module locally
go get github.com/rizkirmdhnnn/boilerplate-go-backend

make run        # Local
make docker-up  # Docker
```

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
│   ├── middleware/        # Gin middleware
│   ├── repository/       # PostgreSQL adapter (implements domain ports)
│   └── router/           # Route definitions
├── migrations/           # SQL migrations (golang-migrate format)
├── pkg/
│   ├── cache/            # Cache interface + Redis / noop implementations
│   ├── database/         # Database connection pool
│   ├── migrator/         # Auto-run migrations on startup
│   └── response/         # Standard API response helpers
├── docs/                 # Detailed documentation
├── .air.toml             # Hot reload config
├── .env.example
├── Dockerfile
├── docker-compose.yml
└── Makefile
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
| GET | /api/v1/users | List (paginated, cached 30s) |
| GET | /api/v1/users/:id | Get by ID |
| PUT | /api/v1/users/:id | Update |
| DELETE | /api/v1/users/:id | Delete |

## Documentation

| Doc | Description |
|-----|-------------|
| [ARCHITECTURE.md](docs/ARCHITECTURE.md) | Clean Architecture layers & DI |
| [DATABASE.md](docs/DATABASE.md) | Migrations with golang-migrate |
| [CACHING.md](docs/CACHING.md) | Redis caching interface & usage |
| [MIDDLEWARE.md](docs/MIDDLEWARE.md) | Hot reload, error handler, rate limiter, middleware stack |
| [SWAGGER.md](SWAGGER.md) | Swagger/OpenAPI annotation guide |

**Swagger UI:** `http://localhost:8080/swagger/index.html` (after starting the server)

## Commands

```bash
make run          # Start server
make build        # Build binary
make test         # Run tests (+ race detector, coverage)
make lint         # Run golangci-lint
make swag         # Regenerate Swagger docs
make generate     # Alias for make swag
make coverage     # Open HTML coverage report
make migrate-create NAME=xxx  # Create new migration file
make migrate-up   # Run pending migrations
make migrate-down # Rollback last migration
make watch        # Hot reload (auto-restart on file change)
make docker-up    # Docker Compose up (api + postgres + redis)
make docker-down  # Docker Compose down
```

## Kubernetes

Deploy to any K8s cluster with [Kustomize](https://kustomize.io/):

```bash
make k8s-apply       # Apply all manifests
make k8s-status      # Pods + Services + HPA
make k8s-logs        # Follow API logs
make k8s-restart     # Rollout restart
make k8s-secret      # Sync secrets from .env
make k8s-delete      # Tear down
```

### Manifests (`k8s/`)

| File | Description |
|------|-------------|
| `namespace.yaml` | `boilerplate` namespace |
| `configmap.yaml` | Non-sensitive config (DB host, rate limit, log level, etc.) |
| `secret.yaml` | Template — use `kubectl create secret` or External Secrets |
| `deployment.yaml` | 2 replicas, rolling update, health probes, resource limits |
| `service.yaml` | ClusterIP on port 80 |
| `ingress.yaml` | nginx ingress with TLS — update host before using |
| `hpa.yaml` | Autoscale 2→10 pods at 70% CPU / 80% memory |
| `kustomization.yaml` | Kustomize root — add `overlays/` for env-specific config |

### Secrets

Secrets are **not committed** with real values. Use one of:

```bash
# Makefile helper (reads from .env)
make k8s-secret

# Or manually
kubectl create secret generic boilerplate-secret -n boilerplate \
  --from-literal=DB_PASSWORD='your-password' \
  --from-literal=JWT_SECRET='your-secret'

# Or GitOps: SealedSecrets / External Secrets
```

### Image

Update the image in `k8s/deployment.yaml` to your container registry:

```yaml
image: ghcr.io/rizkirmdhnnn/boilerplate-go-backend:latest
```

Build & push:

```bash
docker build -t ghcr.io/rizkirmdhnnn/boilerplate-go-backend:latest .
docker push ghcr.io/rizkirmdhnnn/boilerplate-go-backend:latest
```
