.PHONY: run dev test fmt lint-placeholder docker-up docker-down docker-logs migrate-up migrate-down migrate-status sqlc db-reset project-context-handoff

# Load local env vars (DATABASE_URL, etc.) if present.
ifneq (,$(wildcard .env))
include .env
export
endif

# Prefer installed binaries (brew/go install), but allow overriding:
#   make MIGRATE=/path/to/migrate SQLC=/path/to/sqlc migrate-up
MIGRATE ?= migrate
SQLC ?= sqlc

run:
	go run ./cmd/api

dev:
	go run ./cmd/api

test:
	go test ./...

fmt:
	go fmt ./...

lint-placeholder:
	@echo "lint placeholder: add golangci-lint or staticcheck later"

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down -v

docker-logs:
	docker compose logs -f --tail=200

migrate-up:
	$(MIGRATE) -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	$(MIGRATE) -path migrations -database "$(DATABASE_URL)" down 1

migrate-status:
	$(MIGRATE) -path migrations -database "$(DATABASE_URL)" version

sqlc:
	$(SQLC) generate

db-reset:
	$(MIGRATE) -path migrations -database "$(DATABASE_URL)" down -all
	$(MIGRATE) -path migrations -database "$(DATABASE_URL)" up

# Append a dated line to PROJECT_CONTEXT.md session log (§9).
# Usage: make project-context-handoff MSG='one-line summary, no secrets'
project-context-handoff:
	@if [ -z "$(MSG)" ]; then \
		echo "usage: make project-context-handoff MSG='short summary without secrets'" >&2; \
		exit 1; \
	fi
	chmod +x scripts/project-context-handoff.sh
	./scripts/project-context-handoff.sh "$(MSG)"
