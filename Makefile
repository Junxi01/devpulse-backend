.PHONY: run dev test fmt lint-placeholder docker-up docker-down migrate-up migrate-down sqlc db-reset

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
	docker compose -f deploy/docker-compose.yml up -d

docker-down:
	docker compose -f deploy/docker-compose.yml down -v

migrate-up:
	$(MIGRATE) -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	$(MIGRATE) -path migrations -database "$(DATABASE_URL)" down 1

sqlc:
	$(SQLC) generate

db-reset:
	$(MIGRATE) -path migrations -database "$(DATABASE_URL)" down -all
	$(MIGRATE) -path migrations -database "$(DATABASE_URL)" up

