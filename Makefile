.PHONY: help run test build docker-build docker-run migrate lint fmt

BINARY_NAME=nanoflux
DOCKER_IMAGE=nanoflux
DOCKER_TAG=latest

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

run: ## Run the application locally
	go run ./cmd/api

test: ## Run all tests
	go test -v -race ./...

test-coverage: ## Run tests with coverage
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

build: ## Build binary
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o $(BINARY_NAME) ./cmd/api

docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run: ## Run Docker container
	docker run --rm -p 8080:8080 --env-file .env $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-push: ## Push Docker image (requires tag)
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):$(TAG)
	docker push $(DOCKER_IMAGE):$(TAG)

migrate: ## Run database migrations
	docker compose up -d postgres
	sleep 3
	docker compose exec -T postgres psql -U nanoflux -d nanoflux -f /docker-entrypoint-initdb.d/001_create_urls.sql

migrate-down: ## Drop all tables (dev only)
	docker compose exec -T postgres psql -U nanoflux -d nanoflux -c "DROP TABLE IF EXISTS urls;"

db-shell: ## Open PostgreSQL shell
	docker compose exec postgres psql -U nanoflux -d nanoflux

lint: ## Run linter (requires golangci-lint)
	golangci-lint run ./...

fmt: ## Format code
	go fmt ./...
	goimports -w .

clean: ## Clean build artifacts
	rm -f $(BINARY_NAME)
	rm -f coverage.out
	docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) 2>/dev/null || true

deps: ## Download and tidy dependencies
	go mod download
	go mod tidy

dev: ## Start development environment (postgres + app)
	docker compose up -d postgres
	sleep 3
	go run ./cmd/api

.DEFAULT_GOAL := help