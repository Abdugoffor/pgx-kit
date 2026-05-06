# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Run the server (listens on :8081)
make run
# or
go run main.go

# Create a new database migration
make migrate name=<migration_name>

# Generate a full module scaffold
make gen name=<name> table=<table> fields="fieldName:goType:validateTag,..."
# Example:
make gen name=brand table=brands fields="name:string:required max=255,slug:string:required max=255"

# Run tests
go test ./...

# Run a single test
go test ./module/product_service/... -run TestName
```

## Environment

Requires a `.env` file in the project root:

```
DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=<user>
DB_PASSWORD=<password>
DB_NAME=<dbname>
DB_SSLMODE=disable
DB_TIMEZONE=Asia/Tashkent
JWT_KEY=<secret>
```

Migrations run automatically on startup via `config/migration.go` (embedded SQL files, tracked in `schema_migrations` table). The database must exist before running; tables are created via `config/migrations/`.

## Architecture

This is a modular REST API using `httprouter` + `pgx/v5`. Each domain feature is a self-contained module under `module/`:

```
module/<feature>_service/
    cmd.go          # registers routes on the router
    handler/        # HTTP handlers — parse request, call service, write response
    service/        # business logic + raw SQL queries via pgxpool
    dto/            # request and response structs with json/validate tags
```

`main.go` initializes the DB pool, runs migrations, instantiates all modules, and starts the server.

`config/database.go` configures pgxpool (10–50 connections, 1h max lifetime, 30m idle timeout).

`middleware/middleware.go` provides `CheckRole(next, roles...)` — validates the JWT from `Authorization: Bearer <token>`, checks the role claim, and sets `ContextUserID`, `ContextRole`, and `ContextCompanyID` on the request context. Helper functions `middleware.UserID(r)`, `middleware.UserRole(r)`, and `middleware.CompanyID(r)` extract these values.

`helper/helper.go` provides JSON response helpers (`helper.JSON(w, status, data)`) and struct validation (`helper.Validate(v)` returns `map[string]string` of field→failing tag). The shared `validator` instance uses JSON field names in error maps.

## Code Generator

`tools/gen/main.go` scaffolds a full module from CLI flags. Field definitions use the format `fieldName:goType:validateTag` (use `*goType` for optional pointer fields). The generator creates all files under `module/<name>_service/` and outputs the `main.go` registration snippet to stdout. After generating, register the module in `main.go`.

## Key Patterns

**Adding a new module:** Create `module/<name>_service/` with `cmd.go`, `handler/`, `service/`, `dto/`. Register `NewXHandler(router, db)` in `main.go`. Prefer `make gen` to scaffold the boilerplate.

**Routes** are registered under `/api/v1/`. Protect them with `middleware.CheckRole(handler.Method, "admin", "user")`.

**Database queries** use parameterized SQL directly on `*pgxpool.Pool`. Use `pgx.ErrNoRows` to detect 404s.

**Company isolation:** Products and categories are scoped by `company_id` from the JWT. Handlers call `requireCompany()` (returns 403 if `CompanyID == 0`) and pass `companyID` into every service method and SQL query.

**Pagination:** use offset-based for admin list endpoints (`page`, `page_size`); use cursor/keyset for public list endpoints (`cursor`, `limit`). Fetch `limit+1` rows to determine `has_next`.

**Filtering/sorting:** whitelist allowed sort columns in the handler before interpolating into SQL. Use `ILIKE '%' || $n || '%'` for case-insensitive search.

**Partial updates:** use `COALESCE($n, column)` so omitted fields retain their current value.

**Validation errors** return HTTP 422 with `{"errors": {"field": "tag"}}`. Other errors return `{"error": "message"}`.

**Soft deletes** use a `deleted_at TIMESTAMPTZ` column (currently users only); always add `AND deleted_at IS NULL` to queries on those tables.

**JWT claims:** `user_id` (float64), `role` (string), `company_id` (float64), 24-hour expiry, signed with HMAC-SHA256 using `JWT_KEY`.

**Sentinel errors:** define domain errors (e.g., `ErrNoCompany`, `ErrCategoryInvalid`, `ErrInvalidCredentials`) in the service layer and map them to HTTP status codes in the handler.
