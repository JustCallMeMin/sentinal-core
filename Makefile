# Build configuration
BINARY_NAME=sentinal-api
PKG=github.com/sentinal/core
VERSION=$(shell cat pkg/version/version.go | grep "Version =" | cut -d '"' -f 2)
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

# LDFLAGS for version injection
LDFLAGS=-ldflags "-X $(PKG)/pkg/version.Version=$(VERSION) -X $(PKG)/pkg/version.Commit=$(COMMIT) -X $(PKG)/pkg/version.BuildTime=$(BUILD_TIME)"

# Docker configuration
DOCKER_IMAGE=sentinal-core

.PHONY: all build run test clean lint migrate-up migrate-down docker-build docker-run compose-up compose-down logs help

all: help

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@CGO_ENABLED=0 go build $(LDFLAGS) -o bin/$(BINARY_NAME) cmd/api/main.go

## run: Run the application locally
run: build
	@./bin/$(BINARY_NAME)

## test: Run all unit tests
test:
	@echo "Running tests..."
	@go test -v ./...

## lint: Run golangci-lint (requires installation)
lint:
	@echo "Running linter..."
	@golangci-lint run

## clean: Remove build artifacts
clean:
	@echo "Cleaning up..."
	@rm -rf bin/

## migrate-up: Run database up migrations
migrate-up:
	@echo "Running migrations up..."
	@go run cmd/migrate/main.go -up

## migrate-down: Rollback last database migration
migrate-down:
	@echo "Rolling back migrations..."
	@go run cmd/migrate/main.go -down

## migrate-reset: Reset database and run all migrations
migrate-reset:
	@echo "Resetting database..."
	@go run cmd/migrate/main.go -reset

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE):latest .

## docker-run: Run Docker container
docker-run:
	@docker run -p 8080:8080 --env-file .env $(DOCKER_IMAGE):latest

## compose-up: Spin up the entire stack (API, DB, Redis)
compose-up:
	@echo "Spinning up the stack..."
	@docker compose up --build -d

## compose-down: Stop and remove the stack
compose-down:
	@docker compose down

## logs: View logs from all services
logs:
	@docker compose logs -f

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^##' Makefile | sed -e 's/## //'
