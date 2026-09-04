# Productivity OS — developer entry points.
# Requires: Go, Node/pnpm, Docker Compose (for Postgres). golang-migrate CLI on PATH.

SHELL := /bin/bash
-include .env
export

DATABASE_URL ?= postgres://productivity_os:productivity_os@localhost:5433/productivity_os?sslmode=disable
MIGRATIONS_DIR := db/migrations
SQLC := go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.27.0

.PHONY: help
help: ## List targets
	@grep -hE '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  %-16s %s\n", $$1, $$2}'

## --- database ---
.PHONY: db-up
db-up: ## Start the local Postgres container
	docker compose up -d --wait postgres

.PHONY: db-down
db-down: ## Stop the local Postgres container
	docker compose down

.PHONY: db-reset
db-reset: ## Recreate the local Postgres volume from scratch
	docker compose down -v && docker compose up -d --wait postgres

.PHONY: migrate
migrate: ## Apply all pending migrations (embedded)
	go run ./cmd/migrate up

.PHONY: migrate-down
migrate-down: ## Roll every migration back (dev only)
	go run ./cmd/migrate down

.PHONY: migrate-drop
migrate-drop: ## Drop the entire database including migration history (dev only)
	go run ./cmd/migrate drop

.PHONY: migrate-create
migrate-create: ## Scaffold a migration pair: make migrate-create name=add_widgets
	@test -n "$(name)" || (echo "usage: make migrate-create name=<snake_case>"; exit 1)
	@next=$$(printf '%06d' $$(( $$(ls $(MIGRATIONS_DIR)/*.up.sql 2>/dev/null | wc -l) + 1 ))); \
	touch $(MIGRATIONS_DIR)/$${next}_$(name).up.sql $(MIGRATIONS_DIR)/$${next}_$(name).down.sql; \
	echo "created $(MIGRATIONS_DIR)/$${next}_$(name).{up,down}.sql"

## --- code ---
.PHONY: sqlc
sqlc: ## Regenerate sqlc query code
	$(SQLC) generate

.PHONY: sqlc-diff
sqlc-diff: ## Fail if generated sqlc code is stale
	$(SQLC) diff

.PHONY: run
run: ## Run the server
	go run ./cmd/server

.PHONY: build
build: ## Build the server binary into ./bin
	go build -o bin/server ./cmd/server

.PHONY: test
test: ## Run the Go test suite
	go test ./...

.PHONY: lint
lint: ## Run go vet and golangci-lint
	go vet ./...
	golangci-lint run

## --- frontend ---
.PHONY: web-dev
web-dev: ## Run the Vite dev server
	cd web && pnpm install && pnpm dev

.PHONY: web-build
web-build: ## Build the frontend into web/dist
	cd web && pnpm install && pnpm build
