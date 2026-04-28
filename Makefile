ENV_FILE ?= .env
DOCKER_ENV_FILE ?= .env.docker
GO ?= go

-include $(ENV_FILE)

ifeq ($(strip $(DATABASE_URL)),)
DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(if $(DB_SSLMODE),$(DB_SSLMODE),disable)
else
DB_URL := $(DATABASE_URL)
endif

.DEFAULT_GOAL := help

.PHONY: help fmt test check postman-test migrate-up migrate-down migrate-down-all migrate-status migrate-force migrate-create migrate-drop seed db-setup docker-up docker-down docker-logs docker-rebuild

help: ## Show available Makefile commands
	@awk 'BEGIN {FS = ":.*## "}; /^[a-zA-Z0-9_.-]+:.*## / {printf "\033[36m%-18s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

fmt: ## Format Go source files
	$(GO) fmt ./...

test: ## Run the full Go test suite
	$(GO) test ./...

check: fmt test ## Format code and run tests

postman-test: ## Run the Postman collection with Newman against the local environment
	npm run postman:local

migrate-up: ## Apply all pending migrations using the migrate CLI
	migrate -path migrations -database "$(DB_URL)" up

migrate-down: ## Roll back the latest migration using the migrate CLI
	migrate -path migrations -database "$(DB_URL)" down 1

migrate-down-all: ## Roll back all migrations using the migrate CLI
	migrate -path migrations -database "$(DB_URL)" down

migrate-status: ## Show migration status using the migrate CLI
	migrate -path migrations -database "$(DB_URL)" status

migrate-force: ## Force the migration version using VERSION=<number>
ifndef VERSION
	$(error VERSION is undefined. Usage: make migrate-force VERSION=<version>)
endif
	migrate -path migrations -database "$(DB_URL)" force $(VERSION)

migrate-create: ## Create a new SQL migration with NAME=<migration_name>
ifndef NAME
	$(error NAME is undefined. Usage: make migrate-create NAME=<migration_name>)
endif
	migrate create -ext sql -dir migrations -seq "$(NAME)"

migrate-drop: ## Drop all database objects using the migrate CLI (requires CONFIRM=1)
ifndef CONFIRM
	$(error CONFIRM is undefined. Usage: make migrate-drop CONFIRM=1)
endif
	migrate -path migrations -database "$(DB_URL)" drop -f

seed: ## Run seed data using the Go seed command
	$(GO) run ./cmd/seed

db-setup: migrate-up seed ## Apply migrations and run seed data

docker-up: ## Start the Docker stack with build using DOCKER_ENV_FILE=<file>
	docker-compose --env-file $(DOCKER_ENV_FILE) up --build

docker-down: ## Stop the Docker stack using DOCKER_ENV_FILE=<file>
	docker-compose --env-file $(DOCKER_ENV_FILE) down

docker-logs: ## Follow Docker logs using DOCKER_ENV_FILE=<file>
	docker-compose --env-file $(DOCKER_ENV_FILE) logs -f

docker-rebuild: ## Rebuild Docker images without cache using DOCKER_ENV_FILE=<file>
	docker-compose --env-file $(DOCKER_ENV_FILE) build --no-cache
