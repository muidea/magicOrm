export PATH := $(GOPATH)/bin:$(PATH)
export GO111MODULE=on

.PHONY: all build test lint vet fmt fmt-check clean deps

all: fmt vet test build

build:
	@echo "Building project..."
	go build ./...

test:
	@echo "Running tests..."
	go test ./... --count=1

lint:
	@echo "Running code quality checks..."
	@echo "1. Running go vet..."
	@go vet ./... 2>&1 | (grep -v "^vendor/" || true)
	@echo "2. Checking code format (excluding vendor)..."
	@if find . -name "*.go" -not -path "./vendor/*" -exec gofmt -d {} + 2>/dev/null | grep -q '^'; then \
		echo "Code is not formatted correctly. Run 'make fmt' to fix."; \
		find . -name "*.go" -not -path "./vendor/*" -exec gofmt -d {} + 2>/dev/null | head -50; \
		exit 1; \
	else \
		echo "Code is properly formatted."; \
	fi

vet:
	@echo "Running go vet..."
	go vet ./...

fmt:
	@echo "Formatting code..."
	@find . -name "*.go" -not -path "./vendor/*" -exec gofmt -w {} +
	@if command -v goimports >/dev/null 2>&1; then \
		find . -name "*.go" -not -path "./vendor/*" -exec goimports -w {} +; \
		echo "goimports completed."; \
	else \
		echo "goimports not found, skipping..."; \
		echo "To install: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi

fmt-check:
	@echo "Checking code format (excluding vendor)..."
	@if find . -name "*.go" -not -path "./vendor/*" -exec gofmt -d {} + 2>/dev/null | grep -q '^'; then \
		echo "Code is not formatted correctly. Run 'make fmt' to fix."; \
		find . -name "*.go" -not -path "./vendor/*" -exec gofmt -d {} + 2>/dev/null | head -50; \
		exit 1; \
	else \
		echo "Code is properly formatted."; \
	fi

clean:
	@echo "Cleaning up..."
	go clean -cache
	rm -f coverage.out coverage.html

deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download
