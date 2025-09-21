## Kart (Order Food Online)

Go backend implementing an OpenAPI 3 spec using chi, oapi-codegen, PostgreSQL, and sqlc.

### Features
- OpenAPI-first: request validation middleware and generated server/types
- PostgreSQL with goose migrations (docker-compose) and sqlc generated queries

### Requirements
- Go 1.21+
- Docker and Docker Compose

### Quick Start (Docker)
```bash
# build and start: postgres -> migrate -> app
make compose-up

# logs
docker compose logs -f app | cat

# stop
make compose-down
```

### Local Dev (Go directly)
```bash
# start only postgres
make db-up

# set local DB URL (or put this in .env)
export DATABASE_URL=postgres://postgres:postgres@localhost:5432/kart?sslmode=disable

# apply schema + dev seed (from your host)
make migrate
make seed

# run server locally
make run-local
```

### Endpoints (default API key: `apitest`)
```bash
# List products
curl -sS http://localhost:8080/product

# Get product by ID (OpenAPI expects path int64; this server uses string IDs internally)
curl -sS http://localhost:8080/product/10

# Place order
curl -sS http://localhost:8080/order \
  -H 'Content-Type: application/json' \
  -H 'api_key: apitest' \
  -d '{
    "couponCode": "",
    "items": [
      {"productId": "10", "quantity": 1}
    ]
  }'
```

### Project Layout
- `api/openapi.yaml`: API spec (3.1). Server generated into `internal/openapi`.
- `internal/server`: router, handlers, middleware
- `internal/service`: business logic
- `internal/repo`: repositories using `internal/sqlc`
- `db/migrations`: schema; `db/migrations_dev`: dev seed
- `db/queries`: sqlc SQL

### Code Generation
```bash
# regenerate OpenAPI server/types
go generate ./internal/openapi

# regenerate sqlc models/queries
make generate
```

### Testing
```bash
go test ./...
```

### Docker Compose Reference
- `postgres`: PostgreSQL 16
- `migrate`: applies `db/migrations` (schema) before app
- `app`: Go server (distroless); spec baked into image at `/api/openapi.yaml`


### Configuration
Environment variables (loaded from `.env` if present):
- `APP_ENV` (default: `dev`)
- `HTTP_ADDR` (default: `:8080`)
- `API_KEY` (default: `apitest`)
- `DATABASE_URL` (required for local run; docker-compose sets it automatically)

### Notes
- Spec includes `servers: /`; validator is configured with host checks silenced and API key authentication.
- Coupon validation requires presence mask to have at least two bits set.