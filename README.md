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

## Demo Mode (local)

- **`APP_MODE`**: `demo` (default) for local development, or `live` for a production-like setting. The API does **not** embed demo accounts in request handlers; demo data is loaded only when you run the seed command.
- **Seed** (idempotent: safe to run multiple times) after PostgreSQL is up and migrations are applied:

```bash
make migrate-up
make seed-demo
```

`make seed-demo` runs `go run ./cmd/seed`, which stores the demo password with **bcrypt** (same cost as user registration). It refuses to run when `APP_MODE=live` unless you set **`SEED_DEMO=1`** for a deliberate one-off load.

**Demo login** (local only; never use these credentials outside local/dev):

| Field | Value |
|-------|-------|
| Email | `demo@devpulse.local` |
| Password | `demo123456` |

After seeding, log in with the same **`POST /auth/login`** request as above using those credentials. The seed also creates a demo workspace (**Demo Workspace**), project (**Demo Project**), and repository metadata (**github** / `devpulse/demo-backend`, external id `1001`).

See also `seed/demo_data.sql` (documentation pointer only).

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

## GitHub activity (read model)

For each linked **repository** `id` (from `GET /projects/{projectID}/repositories`), workspace members can read ingested activity (webhook/sync data is stored in PostgreSQL; empty lists are normal until a sync job populates rows):

```bash
REPO_ID="<repository-uuid>"

curl -sS "http://localhost:8080/repositories/${REPO_ID}/events" \
  -H "Authorization: Bearer ${TOKEN}"

curl -sS "http://localhost:8080/repositories/${REPO_ID}/pull-requests" \
  -H "Authorization: Bearer ${TOKEN}"

curl -sS "http://localhost:8080/repositories/${REPO_ID}/issues" \
  -H "Authorization: Bearer ${TOKEN}"
```

Optional query params: `limit`, `offset` (same defaults as other list endpoints).

You do **not** need a real GitHub webhook or public URL to try this flow locally: use the mock importers below.

### Mock GitHub event import (local JSON)

Sample webhook-style files live in `seed/github_events/`:

- `push.json` — includes one commit to populate **`commits`**
- `pull_request_opened.json` — populates **`pull_requests`**
- `issues_opened.json` — populates **`issues`**

Each file also creates a row in **`repository_events`** (keyed by `delivery_id` as `external_id`).

From the repository root, after the database is migrated and **`make seed-demo`** has created the demo user and demo repository (`github` / `devpulse/demo-backend`):

```bash
make seed-events
```

Imports are **idempotent**: re-running `make seed-events` skips rows that already exist (same `delivery_id`, pull request `number`, issue `number`, or commit `sha` per repository).

- Override target repo: `go run ./cmd/seed-events -repo <repository-uuid>`
- Override input directory: `go run ./cmd/seed-events -dir /path/to/json`
- Default directory is `./seed/github_events` (resolved from the process working directory)

Same safety gate as `make seed-demo`: `APP_MODE=demo` (default) or `SEED_DEMO=1`.

## Database Migrations

Migrations live in `migrations/` and are applied in order.

Migration **`000004`** replaces the legacy user-scoped `projects` / `project_members` tables with **workspace-scoped** `projects` and `repositories` (see `migrations/000004_projects_repositories.up.sql`). Applying it will **drop** the old tables if they still exist.

Migration **`000005`** adds **`repository_events`**, **`pull_requests`**, **`issues`**, and **`commits`** keyed by `repositories.id` (FK `ON DELETE CASCADE`). Inserts are intended for sync/webhook ingestion; listing APIs cover events, PRs, and issues first.

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
- `sql/queries/github_activity.sql`

Generate Go code:

```bash
make sqlc
```

This runs `sqlc generate` into `internal/db/generated/` (generated only; do not hand-edit). Feature code wraps `generated.Queries` in packages such as `internal/repository/` (users), `internal/workspace/`, `internal/project/`, and `internal/repos/` (linked VCS repositories).

## Environment Variables

- `APP_ENV`: environment name (default: `development`)
- `APP_MODE`: `demo` (default) or `live` — used for local demo/seeding policy; the HTTP server does not switch API behavior on this value yet, but `make seed-demo` requires `demo` unless `SEED_DEMO=1`
- `HTTP_ADDR`: HTTP bind address (default: `:8080`)
- `DATABASE_URL`: PostgreSQL connection string (required)
- `REDIS_ADDR`: Redis endpoint (default: `localhost:6379`)
- `JWT_SECRET`: JWT signing secret (required; placeholder for Day 1)
- `AI_PROVIDER`: AI provider name (default: `openai`)
- `GITHUB_MODE`: GitHub integration mode (default: `mock`)

## Project Layout

```
cmd/api/main.go               # entrypoint: config, logger, db, server
cmd/seed/main.go              # optional: idempotent local demo data (see make seed-demo)
cmd/seed-events/main.go      # optional: import mock GitHub JSON (see make seed-events)
internal/github/mock/         # local JSON → repository_events, PRs, issues, commits
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

