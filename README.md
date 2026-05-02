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

## Auth (Register / Login / Me)

Register:

```bash
curl -sS -X POST http://localhost:8080/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"me@example.com","password":"password123","name":"Me"}'
```

Login:

```bash
curl -sS -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"me@example.com","password":"password123"}'
```

Me (replace `$TOKEN` with `access_token` from login):

```bash
TOKEN="..."
curl -sS http://localhost:8080/me -H "Authorization: Bearer $TOKEN"
```

## Workspaces, projects, and repositories

These routes require `Authorization: Bearer <token>`. JSON bodies and responses use **snake_case** field names (aligned with sqlc-generated types). Only **workspace members** may create or list projects in that workspace, or manage repositories on projects in that workspace; otherwise the API returns **403 Forbidden**. Missing or invalid JWT returns **401 Unauthorized**.

Create a workspace and capture its `id` (see your client or prior workspace docs), then:

```bash
WS_ID="<workspace-uuid>"

curl -sS -X POST "http://localhost:8080/workspaces/${WS_ID}/projects" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"name":"mobile-app","description":"optional"}'

curl -sS "http://localhost:8080/workspaces/${WS_ID}/projects" \
  -H "Authorization: Bearer ${TOKEN}"
```

Link a VCS repository record to a project (for example before GitHub sync):

```bash
PROJECT_ID="<project-uuid>"

curl -sS -X POST "http://localhost:8080/projects/${PROJECT_ID}/repositories" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"provider":"github","owner":"acme","name":"api","full_name":"acme/api","external_id":"123456789","default_branch":"main"}'

curl -sS "http://localhost:8080/projects/${PROJECT_ID}/repositories" \
  -H "Authorization: Bearer ${TOKEN}"
```

## Database Migrations

Migrations live in `migrations/` and are applied in order.

Migration **`000004`** replaces the legacy user-scoped `projects` / `project_members` tables with **workspace-scoped** `projects` and `repositories` (see `migrations/000004_projects_repositories.up.sql`). Applying it will **drop** the old tables if they still exist.

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
- `sql/queries/workspaces.sql`
- `sql/queries/projects.sql`
- `sql/queries/repositories.sql`

Generate Go code:

```bash
make sqlc
```

This runs `sqlc generate` into `internal/db/generated/` (generated only; do not hand-edit). Feature code wraps `generated.Queries` in packages such as `internal/repository/` (users), `internal/workspace/`, `internal/project/`, and `internal/repos/` (linked VCS repositories).

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

