DB_URL := postgres://meilanasapta:postgres@localhost:5432/go_auth?sslmode=disable

.PHONY: migrate-up migrate-down migrate-down-all migrate-status migrate-force migrate-create migrate-drop seed db-setup

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