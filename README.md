# devpulse-backend

DevPulse is a developer productivity platform. This backend service receives GitHub events, manages project data, and provides a foundation for AI engineering reports.

This repository currently provides a runnable backend skeleton only (no complex business logic yet).

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
make run
```

Validate the endpoints:

```bash
curl http://localhost:8080/healthz
```

Expected responses:

- `GET /healthz` -> `{"status":"ok","service":"devpulse-api"}`

## Docker Compose (API + PostgreSQL + Redis)

Start all local services in one command:

```bash
docker compose up --build
```

Or via Makefile shortcut:

```bash
make docker-up
```

Services started:

- `api` on `http://localhost:8080`
- `postgres` on `localhost:5432`
- `redis` on `localhost:6379`

View logs:

```bash
make docker-logs
```

Stop and remove containers/volumes:

```bash
make docker-down
```

Verify API from host:

```bash
curl http://localhost:8080/healthz
```

## Database Migrations

Migrations live in `migrations/` and are applied in order.

Note on `DATABASE_URL`:

- When running the API **inside Docker Compose**, use `postgres:5432` (service DNS).
- When running `make migrate-*` **from your host machine**, use `localhost:5432` (published port).

Run migrations up:

```bash
make migrate-up
```

Rollback the latest migration:

```bash
make migrate-down
```

Show current migration version:

```bash
make migrate-status
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

This runs `sqlc generate` and writes code into:
\n+- `internal/db/generated/` (generated only; do not hand-edit)\n+\n+Hand-written domain logic should live outside the generated package (e.g. `internal/repository/`), and can wrap `generated.Queries` behind small interfaces.\n+
Generated code output:

- `internal/db/generated`

## Environment Variables

- `APP_ENV`: environment name (default: `development`)
- `APP_MODE`: runtime mode (default: `api`)
- `HTTP_ADDR`: HTTP bind address (default: `:8080`)
- `DATABASE_URL`: PostgreSQL connection string (required)
- `REDIS_ADDR`: Redis endpoint (default: `localhost:6379`)
- `JWT_SECRET`: JWT signing secret (required; placeholder for Day 1)
- `AI_PROVIDER`: AI provider name (default: `openai`)
- `GITHUB_MODE`: GitHub integration mode (default: `mock`)

## Project Layout

```
cmd/api/main.go               # entrypoint: config, logger, db, server
internal/config/config.go     # config loading + validation (.env supported)
internal/logger/              # slog logger setup
internal/middleware/          # shared HTTP middleware (request logging, etc.)
internal/db/db.go             # database connection (pgxpool)
internal/db/generated         # sqlc generated code
internal/server/server.go     # chi router, middleware, routes
internal/health/handler.go    # /healthz and /readyz handlers
docs/                         # engineering docs and runbooks
api/                          # OpenAPI / API contracts
deploy/docker-compose.yml     # local API + PostgreSQL + Redis
Dockerfile                    # container build for API service
migrations/                   # schema migrations (golang-migrate)
sql/queries/                  # sqlc query sources
```

