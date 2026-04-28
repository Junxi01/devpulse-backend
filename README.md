# devpulse-backend

DevPulse is a developer productivity platform. This backend is responsible for receiving GitHub events, managing projects, and generating AI engineering reports.

This repository currently provides a **Day 1 skeleton**: configuration loading, database connectivity, an HTTP server, and health/readiness endpoints for local development.

## Tech Stack

- Go 1.22+
- HTTP router: `chi`
- Logging: standard library `log/slog` (JSON structured logging)
- Database: PostgreSQL (`database/sql` + `pgx` driver)
- Config: environment variables + `.env` (`godotenv`)

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

## Environment Variables

- `APP_ENV`: environment name (default: `development`)
- `HTTP_ADDR`: HTTP bind address (default: `:8080`)
- `DATABASE_URL`: PostgreSQL connection string (required)
- `JWT_SECRET`: JWT signing secret (required; placeholder for Day 1)

## Project Layout

```
cmd/api/main.go               # entrypoint: config, logger, db, server
internal/config/config.go     # config loading + validation (.env supported)
internal/db/db.go             # database connection (sql.DB)
internal/server/server.go     # chi router, middleware, routes
internal/health/handler.go    # /healthz and /readyz handlers
deploy/docker-compose.yml     # local PostgreSQL
```

