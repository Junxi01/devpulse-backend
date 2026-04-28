# devpulse-backend

DevPulse is a developer productivity platform. This backend is responsible for receiving GitHub events, managing projects, and generating AI engineering reports.

This repository currently provides a **Day 1 skeleton**: configuration loading, database connectivity, an HTTP server, and health/readiness endpoints for local development.

## Tech Stack

- Go 1.22+
- HTTP router: `chi`
- Logging: standard library `log/slog` (JSON structured logging)
- Database: PostgreSQL (`pgxpool` + `sqlc`)
- Config: environment variables + `.env` (`godotenv`)
- Migrations: `golang-migrate`

## Quick Start (Local)

Create a local `.env` file:

```bash
cp .env.example .env
```

Start PostgreSQL:

```bash
make docker-up
```

Start the API:

```bash
make dev
```

Validate the endpoints:

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz
```

Expected responses:

- `GET /healthz` -> `{"status":"ok"}`
- `GET /readyz` -> `{"status":"ready"}` (returns 503 if the database is not reachable)

## Database Migrations

Migrations live in `migrations/` and are applied in order.

Run migrations up:

```bash
make migrate-up
```

Rollback the latest migration:

```bash
make migrate-down
```

Reset the database schema (dangerous, drops all migrations then re-applies):

```bash
make db-reset
```

## sqlc (Type-safe Queries)

SQL query sources live in:

- `sql/queries/users.sql`
- `sql/queries/projects.sql`

Generate Go code:

```bash
make sqlc
```

Generated code output:

- `internal/db/generated`

## Environment Variables

- `APP_ENV`: environment name (default: `development`)
- `HTTP_ADDR`: HTTP bind address (default: `:8080`)
- `DATABASE_URL`: PostgreSQL connection string (required)
- `JWT_SECRET`: JWT signing secret (required; placeholder for Day 1)

## Project Layout

```
cmd/api/main.go               # entrypoint: config, logger, db, server
internal/config/config.go     # config loading + validation (.env supported)
internal/db/db.go             # database connection (pgxpool)
internal/db/generated         # sqlc generated code
internal/server/server.go     # chi router, middleware, routes
internal/health/handler.go    # /healthz and /readyz handlers
deploy/docker-compose.yml     # local PostgreSQL
migrations/                   # schema migrations (golang-migrate)
sql/queries/                  # sqlc query sources
```

