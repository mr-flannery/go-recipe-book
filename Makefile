.PHONY: dev run build test test-unit test-integration test-coverage test-browser test-browser-full test-browser-medium test-browser-minimal test-browser-ui test-browser-full-ui clean db-start migrate ci-local qr help

help:
	@echo "Available commands:"
	@echo "  make dev                  - Start development server with hot reload"
	@echo "  make run                  - Run the server directly (no hot reload)"
	@echo "  make build                - Build the binary to bin/app"
	@echo "  make test                 - Run all tests"
	@echo "  make test-unit            - Run unit tests only (fast, no Docker)"
	@echo "  make test-integration     - Run integration tests (requires Docker)"
	@echo "  make test-coverage        - Run all tests with coverage report"
	@echo "  make test-browser         - Run browser tests headless, minimal mode (requires server running)"
	@echo "  make test-browser-full    - Run browser tests: all browsers + mobile viewports"
	@echo "  make test-browser-medium  - Run browser tests: chromium desktop + mobile viewports"
	@echo "  make test-browser-minimal - Run browser tests: chromium desktop only"
	@echo "  make test-browser-ui      - Run browser tests with UI (requires server running)"
	@echo "  make test-browser-full-ui    - Run browser tests: all browsers + mobile viewports with UI"
	@echo "  make clean                - Remove build artifacts"
	@echo "  make migrate              - Run database migrations"
	@echo "  make ci-local             - Run GitHub Actions CI pipeline locally using act"
	@echo "  make qr                   - Show QR code to access server from mobile"

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

db-start:
	docker compose --profile dev up

clean:
	rm -rf bin/ src/tmp/

migrate:
	go run ./src/main.go migrate

test-browser: test-browser-minimal

test-browser-full:
	cd browser-tests && BROWSER_TEST_MODE=full npx playwright test

test-browser-medium:
	cd browser-tests && BROWSER_TEST_MODE=medium npx playwright test

test-browser-minimal:
	cd browser-tests && BROWSER_TEST_MODE=minimal npx playwright test

test-browser-ui:
	cd browser-tests && npx playwright test --ui

test-browser-full-ui:
	cd browser-tests && BROWSER_TEST_MODE=full npx playwright test --ui

ci-local:
	@if ! command -v act >/dev/null 2>&1; then \
		echo "Error: 'act' is not installed. Install it with:"; \
		echo "  brew install act    (macOS)"; \
		echo "  or see https://github.com/nektos/act#installation"; \
		exit 1; \
	fi
	@if [ ! -f .secrets ]; then \
		echo "Error: .secrets file not found. Copy .secrets.example to .secrets and fill in values."; \
		exit 1; \
	fi
	@mkdir -p .act-artifacts
	act --secret-file .secrets --artifact-server-path .act-artifacts

qr:
	@go run ./cmd/qr
