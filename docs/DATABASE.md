# Database Migrations

Uses [golang-migrate](https://github.com/golang-migrate/migrate) and runs **automatically on startup**.

## Migration Files

SQL files live in `migrations/` with naming format:

```
{sequence}_{name}.up.sql
{sequence}_{name}.down.sql
```

Example:

```
migrations/
├── 000001_create_users_table.up.sql
├── 000001_create_users_table.down.sql
├── 000002_create_todos_table.up.sql
└── 000002_create_todos_table.down.sql
```

## Create a New Migration

```bash
make migrate-create NAME=create_todos
```

This generates:

```
migrations/
├── 000002_create_todos_table.up.sql   ← write your CREATE TABLE here
└── 000002_create_todos_table.down.sql ← write DROP TABLE here
```

## Manual Commands

```bash
make migrate-up      # Run all pending migrations
make migrate-down    # Rollback last migration
```

## Auto-run on Startup

When the server starts (`make run`), `pkg/migrator/migrator.go` automatically runs any pending migrations using golang-migrate's Go library. Tracked in the database via the `schema_migrations` table.

## How It Works

1. Server starts → `migrator.Run()` is called in `main.go`
2. golang-migrate checks the `schema_migrations` table for applied migrations
3. Any unapplied `.up.sql` files are executed in sequence
4. If no new migrations exist, returns `ErrNoChange` — no error, just nothing to do
