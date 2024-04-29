.DEFAULT_GOAL := dev
SHELL := /bin/bash

.PHONY: dev
dev:
	@go build -tags=viper_bind_struct -o .tmp/bin/raid-mate ./cmd/app/main.go
	@.tmp/bin/raid-mate

.PHONY: lint
lint:
	@pre-commit run --hook-stage pre-push -a
	@pre-commit run --hook-stage pre-commit -a