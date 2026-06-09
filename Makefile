.PHONY: help run watch build test lint clean swag docker-up docker-down migrate generate

APP_NAME := boilerplate
BINARY := api
BUILD_DIR := build

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

run: ## Run the API server locally
	go run ./cmd/api

watch: ## Run with hot reload (air)
	@command -v air >/dev/null 2>&1 || { echo "air not installed. Run: go install github.com/air-verse/air@latest"; exit 1; }
	air

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

# ──────────────────────────────────────────────
# Kubernetes
# ──────────────────────────────────────────────

K8S_NS := boilerplate

k8s-apply: ## Apply all K8s manifests via Kustomize
	kubectl apply -k k8s/

k8s-delete: ## Delete all K8s manifests via Kustomize
	kubectl delete -k k8s/

k8s-status: ## Show K8s pods and services
	@echo "=== Pods ===" && kubectl get pods -n $(K8S_NS)
	@echo "=== Services ===" && kubectl get svc -n $(K8S_NS)
	@echo "=== HPA ===" && kubectl get hpa -n $(K8S_NS)

k8s-logs: ## Follow API logs
	kubectl logs -n $(K8S_NS) -l app.kubernetes.io/component=api --tail=50 -f

k8s-restart: ## Rollout restart API deployment
	kubectl rollout restart deployment/boilerplate-api -n $(K8S_NS)

k8s-secret: ## Create secret from .env file (usage: make k8s-secret)
	@test -f .env || { echo ".env file not found"; exit 1; }
	kubectl create secret generic boilerplate-secret -n $(K8S_NS) \
		--from-literal=DB_PASSWORD="$(shell grep DB_PASSWORD .env | cut -d= -f2)" \
		--from-literal=JWT_SECRET="$(shell grep JWT_SECRET .env | cut -d= -f2)" \
		--from-literal=REDIS_PASSWORD="$(shell grep REDIS_PASSWORD .env | cut -d= -f2)" \
		--dry-run=client -o yaml | kubectl apply -f -

tidy: ## Tidy go modules
	go mod tidy
