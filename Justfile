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

build-skia:
    CGO_ENABLED=1 go build -tags skia ./...

test:
    go test ./...

test-skia:
    CGO_ENABLED=1 go test -tags skia ./...

backend-info:
    @go run ./examples/backends/info/main.go 2>/dev/null || echo "Backend info example not yet available"

cli:
    go run ./main.go --help

examples:
    @echo "Running examples..."
    @for dir in examples/*/; do \
        if [ -f "$$dir/main.go" ]; then \
            echo "Running $$dir"; \
            cd "$$dir" && go run main.go; \
            cd - > /dev/null; \
        elif [ -f "$$dir/basic.go" ]; then \
            echo "Running $$dir/basic.go"; \
            cd "$$dir" && go run basic.go; \
            cd - > /dev/null; \
        fi; \
    done
    @for subdir in examples/*/*/; do \
        if [ -f "$$subdir/main.go" ]; then \
            echo "Running $$subdir"; \
            cd "$$subdir" && go run main.go; \
            cd - > /dev/null; \
        fi; \
    done

clean-examples:
    @echo "Cleaning PNG files from examples..."
    find examples/ -name "*.png" -type f -delete
    @echo "PNG files removed."
