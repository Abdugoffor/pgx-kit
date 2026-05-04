# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Run the application
go run main.go

# Create a new migration file
make migrate name=<migration_name>
# or directly:
go run . migrate:create <migration_name>
```

## Architecture

This is a Go REST API kit using `julienschmidt/httprouter` for routing and `jackc/pgx/v5` (pgxpool) for PostgreSQL access.

**Startup flow:** `main.go` loads `.env` via `helper.LoadEnv()`, then calls `config.DBConnect()` which connects the pool, runs any pending migrations automatically, and returns the pool.

**Migration system** (`config/migration.go`): Migrations are SQL files embedded at compile time via `//go:embed migrations/*.sql`. On startup, `RunMigrations()` creates a `schema_migrations` table if absent, then applies any unapplied `.sql` files in sorted filename order. `MigrateCreate(name)` scaffolds the next numbered file (e.g. `004_<name>.sql`) in `config/migrations/`.

**Module layout** — each domain lives under `module/<name>/`:
- `dto/` — request and response structs with `validate` tags (`go-playground/validator`)
- `service/` — interface + concrete struct that takes `*pgxpool.Pool`; SQL written inline
- `cmd.go` — router wiring (currently empty stubs, to be filled)

**Pointer fields for partial updates:** `Update` DTOs use `*string`/`*bool` so `COALESCE($n, column)` in SQL leaves unchanged fields untouched.

**Environment** — all config is read from `.env` via `godotenv`. Required keys: `DB_DRIVER`, `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`, `DB_TIMEZONE`.

**Connection pool defaults:** MaxConns=50, MinConns=10, MaxConnLifetime=1h, MaxConnIdleTime=30m, HealthCheckPeriod=1m.
