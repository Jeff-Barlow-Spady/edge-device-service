REGISTRY ?= ghcr.io
USERNAME ?= your-username
VERSION ?= latest

PLATFORMS ?= linux/amd64,linux/arm64

.PHONY: build build-multiarch run run-dev test clean lint format help

help: ## Show this help message
    @echo 'Usage: make [target]'
    @echo ''
    @echo 'Targets:'
    @awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build all services for local architecture
    docker-compose build

build-multiarch: ## Build all services for multiple architectures
    docker buildx build --platform=$(PLATFORMS) \
        -t $(REGISTRY)/$(USERNAME)/gpio:$(VERSION) services/gpio \
        -t $(REGISTRY)/$(USERNAME)/auth:$(VERSION) services/auth \
        -t $(REGISTRY)/$(USERNAME)/metrics:$(VERSION) services/metrics

run: ## Run services in production mode
    docker-compose up -d

run-dev: ## Run services in development mode with debug logs
    docker-compose up

test: ## Run tests for all services
    docker-compose -f docker-compose.yml -f docker-compose.test.yml up --build --exit-code-from tests

clean: ## Clean up containers, volumes, and build cache
    docker-compose down -v
    docker system prune -f

lint: ## Run golangci-lint
golangci-lint run ./...

format: ## Format Go code
go fmt ./...

test: ## Run tests with coverage
go test -v -race -cover ./...

build-gpiosvc: ## Build GPIO service
CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/gpiosvc ./cmd/gpiosvc

build-authsvc: ## Build Auth service  
CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/authsvc ./cmd/authsvc

build-metricsvc: ## Build Metrics service
CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/metricsvc ./cmd/metricsvc

build-arm: ## Build for Raspberry Pi
GOOS=linux GOARCH=arm64 make build-gpiosvc build-authsvc build-metricsvc

proto: ## Generate protobuf code
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/proto/*.proto

