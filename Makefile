.PHONY: dev run build test clean migrate help

help:
	@echo "Available commands:"
	@echo "  make dev      - Start development server with hot reload"
	@echo "  make run      - Run the server directly (no hot reload)"
	@echo "  make build    - Build the binary to bin/app"
	@echo "  make test     - Run all tests"
	@echo "  make clean    - Remove build artifacts"
	@echo "  make migrate  - Run database migrations"

dev:
	air

run:
	go run ./src/main.go

build:
	@mkdir -p bin
	go build -o bin/app ./src/main.go

test:
	go test ./...

clean:
	rm -rf bin/ src/tmp/

migrate:
	go run ./src/main.go migrate
