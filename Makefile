.DEFAULT_GOAL := dev
SHELL := /bin/bash

DB_HOST := localhost
DB_PORT := 5432

DB_NAME := ${RAIDMATE_DATABASE_NAME}
DB_USER := ${RAIDMATE_DATABASE_USER}
DB_PASSWORD := ${RAIDMATE_DATABASE_PASSWORD}

.PHONY: dev
dev:
	@go build -tags=viper_bind_struct -o .tmp/bin/raid-mate ./cmd/app/main.go
	@.tmp/bin/raid-mate --config .tmp/config.yaml

.PHONY: database
database:
	@docker-compose -f docker-compose.yaml up -d
	@sleep 2 # wait for database to start

.PHONY: db-up
db-up: database migrate-up
	@echo "Database is up and running"

.PHONY: db-down
db-down: migrate-down
	@docker-compose -f docker-compose.yaml down

.PHONY: new-migration
new-migration:
	@migrate create -ext sql -dir ./internal/database/migrations -seq $(name)

.PHONY: migrate-up
migrate-up:
	@migrate -path internal/database/migrations -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" -verbose up

.PHONY: migrate-down
migrate-down:
	@migrate -path internal/database/migrations -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" -verbose down -all

.PHONY: generate
generate:
	@sqlc generate

.PHONY: lint
lint:
	@pre-commit run -a