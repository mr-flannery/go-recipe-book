.PHONY: dev run build test test-unit test-integration test-coverage clean migrate help

help:
	@echo "Available commands:"
	@echo "  make dev              - Start development server with hot reload"
	@echo "  make run              - Run the server directly (no hot reload)"
	@echo "  make build            - Build the binary to bin/app"
	@echo "  make test             - Run all tests"
	@echo "  make test-unit        - Run unit tests only (fast, no Docker)"
	@echo "  make test-integration - Run integration tests (requires Docker)"
	@echo "  make test-coverage    - Run all tests with coverage report"
	@echo "  make clean            - Remove build artifacts"
	@echo "  make migrate          - Run database migrations"

dev:
	air

run:
	go run ./src/main.go

build:
	@mkdir -p bin
	go build -o bin/app ./src/main.go

test:
	go test ./...

test-unit:
	go test -short ./...

test-integration:
	go test -v -run "Integration" ./...

test-coverage:
	go test -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean:
	rm -rf bin/ src/tmp/

migrate:
	go run ./src/main.go migrate
