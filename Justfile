# Justfile for common tasks

set shell := ["bash", "-cu"]

default: build

fmt:
    if command -v treefmt >/dev/null 2>&1; then \
      treefmt --allow-missing-formatter; \
    else \
      echo "treefmt not installed; skipping"; \
    fi

lint:
    if command -v golangci-lint >/dev/null 2>&1; then \
      golangci-lint run --timeout=5m; \
    else \
      echo "golangci-lint not installed; run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
      exit 1; \
    fi

lint-fix:
    if command -v golangci-lint >/dev/null 2>&1; then \
      golangci-lint run --fix --timeout=5m; \
    else \
      echo "golangci-lint not installed; run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
      exit 1; \
    fi

build:
    go build ./...

test:
    go test ./...

cli:
    go run ./main.go --help
