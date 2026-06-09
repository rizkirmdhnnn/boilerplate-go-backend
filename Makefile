.PHONY: help run build test lint clean swag docker-up docker-down migrate generate watch

APP_NAME := boilerplate
BINARY := api
BUILD_DIR := build

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

run: ## Run the API server locally
	go run ./cmd/api

build: ## Build the binary
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY) ./cmd/api

test: ## Run tests
	go test -v -race -coverprofile=coverage.out ./...

test-short: ## Run short tests
	go test -v -short -race ./...

lint: ## Run linter
	golangci-lint run ./...

swag: ## Regenerate Swagger docs
	swag init -g cmd/api/main.go --output docs --quiet

generate: swag ## Run go generate (alias for make swag)

clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR) coverage.out

docker-up: ## Start services with Docker Compose
	docker compose up -d --build

docker-down: ## Stop services
	docker compose down

docker-logs: ## Follow logs
	docker compose logs -f

migrate-create: ## Create a new migration (usage: make migrate-create NAME=create_todos)
	migrate create -ext sql -dir migrations -seq $(NAME)

migrate-up: ## Run all pending migrations
	migrate -path migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" up

migrate-down: ## Rollback last migration
	migrate -path migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" down 1

coverage: ## Show test coverage
	go tool cover -html=coverage.out

tidy: ## Tidy go modules
	go mod tidy
