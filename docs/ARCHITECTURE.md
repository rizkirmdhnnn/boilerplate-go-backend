# Architecture

Clean Architecture with dependency inversion — **dependencies flow inward**.

## Layer Diagram

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

**Rule:** Nothing in an inner circle imports from an outer circle.

## Layer Responsibilities

| Layer | Depends On | Purpose |
|-------|-----------|---------|
| `domain/` | nothing | Entities + repository interfaces (ports) |
| `application/` | `domain/` only | Use cases, business logic, DTOs |
| `handler/` | `application/` | HTTP adapter — converts requests/responses |
| `repository/` | `domain/` | Implements repository ports (PostgreSQL) |
| `middleware/` | `application/` | HTTP concerns — recovery, logging, CORS, security headers, auth, error handler, rate limiter |
| `cmd/api/` | everything | Composition root, DI wiring |

## Dependency Injection

All dependencies are wired in `cmd/api/main.go`. Interfaces are defined in `domain/` and implemented in outer layers. Handlers never instantiate services — they receive them via constructor injection.
