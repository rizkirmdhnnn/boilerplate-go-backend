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

## Swagger API Docs

This project uses [swaggo/swag](https://github.com/swaggo/swag) to auto-generate OpenAPI/Swagger documentation from Go annotations — no manual YAML/JSON editing needed.

### Access Swagger UI

Start the server, then open:

```
http://localhost:8080/swagger/index.html
```

Every endpoint is listed with request/response schemas. Click **Authorize** and paste your JWT token (from `POST /api/v1/auth/login`) to test protected endpoints directly from the browser.

### Adding Annotations to Endpoints

Add Swagger annotations **above every handler function** in `internal/handler/`:

```go
// @Summary      List all todos
// @Description  Returns paginated list of todos for current user
// @Tags         todos
// @Security     BearerAuth
// @Produce      json
// @Param        page     query  int  false  "Page number"  default(1)
// @Param        per_page query  int  false  "Items per page"  default(10)
// @Success      200  {object}  response.APIResponse
// @Failure      401  {object}  response.APIResponse
// @Router       /api/v1/todos [get]
func (h *TodoHandler) List(c *gin.Context) {
```

**Key annotations:**

| Annotation | Purpose | Example |
|-----------|---------|---------|
| `@Summary` | Short endpoint title | `List all todos` |
| `@Description` | Longer explanation | `Returns paginated list...` |
| `@Tags` | Grouping in Swagger UI | `todos` |
| `@Accept` | Request content type | `json` |
| `@Produce` | Response content type | `json` |
| `@Param` | Request parameter | `body body dto.Request true "payload"` |
| `@Success` | Success response | `200 {object} response.APIResponse` |
| `@Failure` | Error response | `400 {object} response.APIResponse` |
| `@Security` | Auth requirement | `BearerAuth` |
| `@Router` | Path + method | `/api/v1/todos [get]` |

**Parameter types:**
- `body` — JSON body (references a Go struct)
- `path` — URL path param (e.g. `/:id`)
- `query` — Query string param (e.g. `?page=1`)

### Regenerate Docs

After adding or changing annotations:

```bash
make swag
```

This runs `swag init -g cmd/api/main.go --output docs --quiet` and regenerates `docs/`. The Swagger UI reflects changes immediately on refresh.

### Global API Info

API-level metadata lives in `cmd/api/main.go` (set once):

```go
// @title           Boilerplate API
// @version         1.0.0
// @description     Go backend template with Clean Architecture
// @contact.name    Rizkirmdhn
// @host            localhost:8080
// @BasePath        /api/v1
//
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
```

### How it works

1. `swag init` scans all `.go` files in the project
2. Picks up `@Router`, `@Summary`, `@Param` etc. from handler files
3. Reads Go structs (DTOs, request/response models) and converts them to JSON Schema
4. Generates `docs/docs.go`, `docs/swagger.json`, `docs/swagger.yaml`
5. Gin serves these at `GET /swagger/*any` via `gin-swagger`

### DTO Structs Auto-Documented

Request and response structs in `internal/application/dto.go` (or wherever you define them) appear automatically in Swagger's **Models** section, including field names, types, and validation rules from struct tags:

```go
type RegisterRequest struct {
    Email    string `json:"email" binding:"required,email"`         // email, required
    Name     string `json:"name"   binding:"required,min=2,max=100"` // string, 2-100 chars
    Password string `json:"password" binding:"required,min=8"`       // string, min 8 chars
}
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
make docker-up    # Docker Compose up (api + postgres)
make docker-down  # Docker Compose down
```
