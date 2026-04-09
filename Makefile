ENV_FILE ?= .env
-include $(ENV_FILE)

DB_HOST ?= $(DB_HOST)
DB_PORT ?= $(DB_PORT)
DB_USER ?= $(DB_USER)
DB_PASSWORD ?= $(DB_PASSWORD)
DB_NAME ?= $(DB_NAME)

DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

.PHONY: migrate-up migrate-down migrate-down-all migrate-status migrate-force migrate-create migrate-drop seed db-setup docker-up docker-down docker-logs docker-rebuild

migrate-up:
	migrate -path migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path migrations -database "$(DB_URL)" down 1

migrate-down-all:
	migrate -path migrations -database "$(DB_URL)" down

migrate-status:
	migrate -path migrations -database "$(DB_URL)" status

migrate-force:
ifndef VERSION
	$(error VERSION is undefined. Usage: make migrate-force VERSION=<version>)
endif
	migrate -path migrations -database "$(DB_URL)" force $(VERSION)

migrate-create:
ifndef NAME
	$(error NAME is undefined. Usage: make migrate-create NAME=<migration_name>)
endif
	migrate create -ext sql -dir migrations -seq "$(NAME)"

migrate-drop:
	migrate -path migrations -database "$(DB_URL)" drop -f

seed:
	go run cmd/seed/seed.go

db-setup: migrate-up seed

docker-up:
	docker-compose --env-file .env.docker up --build

docker-down:
	docker-compose --env-file .env.docker down

docker-logs:
	docker-compose --env-file .env.docker logs -f

docker-rebuild:
	docker-compose --env-file .env.docker build --no-cache
