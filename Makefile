SHELL := /bin/bash

# Default local database URL (override by passing DATABASE_URL=...)
DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/kart?sslmode=disable

.PHONY: generate tidy build run migrate seed seed-status seed-redo seed-reset compose-up compose-down db-up db-down db-logs

generate:
	go generate ./...

tidy:
	go mod tidy

build: generate tidy
	go build ./...

run: build
	go run ./cmd/server


# Run goose migrations against DATABASE_URL (env var)
migrate:
	go run github.com/pressly/goose/v3/cmd/goose@latest -dir db/migrations postgres "$(DATABASE_URL)" up

# Run dev seed migrations against DATABASE_URL (env var)
seed:
	go run github.com/pressly/goose/v3/cmd/goose@latest -dir db/migrations_dev postgres "$(DATABASE_URL)" up

# Show status of dev seed migrations
seed-status:
	go run github.com/pressly/goose/v3/cmd/goose@latest -dir db/migrations_dev postgres "$(DATABASE_URL)" status | cat

# Re-run the last dev seed migration (useful after editing the file)
seed-redo:
	go run github.com/pressly/goose/v3/cmd/goose@latest -dir db/migrations_dev postgres "$(DATABASE_URL)" redo

# WARNING: Roll back all dev seed migrations, then re-apply
seed-reset:
	go run github.com/pressly/goose/v3/cmd/goose@latest -dir db/migrations_dev postgres "$(DATABASE_URL)" reset && \
	go run github.com/pressly/goose/v3/cmd/goose@latest -dir db/migrations_dev postgres "$(DATABASE_URL)" up

# Run server with Makefile DATABASE_URL
run-local:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/server

# Bring up full stack with compose (postgres -> migrate -> app via depends_on)
compose-up:
	docker compose up --build -d

compose-down:
	docker compose down

# Start only Postgres
db-up:
	docker compose up -d postgres

# Stop Postgres
db-down:
	docker compose stop postgres

# Tail Postgres logs
db-logs:
	docker compose logs -f postgres | cat


