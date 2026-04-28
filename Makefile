.PHONY: dev test docker-up docker-down

dev:
	go run ./cmd/api

test:
	go test ./...

docker-up:
	docker compose -f deploy/docker-compose.yml up -d

docker-down:
	docker compose -f deploy/docker-compose.yml down -v

