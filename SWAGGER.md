# Swagger API Docs

This project uses [swaggo/swag](https://github.com/swaggo/swag) to auto-generate OpenAPI/Swagger documentation from Go annotations — no manual YAML/JSON editing needed.

---

## Access Swagger UI

Start the server, then open:

```
http://localhost:8080/swagger/index.html
```

Every endpoint is listed with request/response schemas. Click **Authorize** and paste your JWT token (from `POST /api/v1/auth/login`) to test protected endpoints directly from the browser.

---

## Adding Annotations to Endpoints

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

### Key Annotations

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

### Parameter Types

- `body` — JSON body (references a Go struct)
- `path` — URL path param (e.g. `/:id`)
- `query` — Query string param (e.g. `?page=1`)

---

## Global API Info

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

---

## Regenerate Docs

After adding or changing annotations:

```bash
make swag
```

This runs `swag init -g cmd/api/main.go --output docs --quiet` and regenerates `docs/`. The Swagger UI reflects changes immediately on refresh.

---

## How It Works

1. `swag init` scans all `.go` files in the project
2. Picks up `@Router`, `@Summary`, `@Param` etc. from handler files
3. Reads Go structs (DTOs, request/response models) and converts them to JSON Schema
4. Generates `docs/docs.go`, `docs/swagger.json`, `docs/swagger.yaml`
5. Gin serves these at `GET /swagger/*any` via `gin-swagger`

---

## DTO Structs Auto-Documented

Request and response structs in `internal/application/dto.go` (or wherever you define them) appear automatically in Swagger's **Models** section, including field names, types, and validation rules from struct tags:

```go
type RegisterRequest struct {
    Email    string `json:"email" binding:"required,email"`         // email, required
    Name     string `json:"name"   binding:"required,min=2,max=100"` // string, 2-100 chars
    Password string `json:"password" binding:"required,min=8"`       // string, min 8 chars
}
```
