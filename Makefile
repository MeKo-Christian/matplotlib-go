.PHONY: all fmt lint lint-fix build test cli

all: build

fmt:
	@which treefmt >/dev/null 2>&1 && treefmt --allow-missing-formatter || echo "treefmt not installed; skipping"

lint:
	@which golangci-lint >/dev/null 2>&1 && golangci-lint run --timeout=5m || (echo "Install golangci-lint" && exit 1)

lint-fix:
	@which golangci-lint >/dev/null 2>&1 && golangci-lint run --fix --timeout=5m || (echo "Install golangci-lint" && exit 1)

build:
	go build ./...

test:
	go test ./...

cli:
	go run ./main.go --help
