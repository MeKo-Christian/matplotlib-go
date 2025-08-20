.PHONY: all fmt lint lint-fix build build-skia test test-skia backend-info cli

all: build

fmt:
	@which treefmt >/dev/null 2>&1 && treefmt --allow-missing-formatter || echo "treefmt not installed; skipping"

lint:
	@which golangci-lint >/dev/null 2>&1 && golangci-lint run --timeout=5m || (echo "Install golangci-lint" && exit 1)

lint-fix:
	@which golangci-lint >/dev/null 2>&1 && golangci-lint run --fix --timeout=5m || (echo "Install golangci-lint" && exit 1)

build:
	go build ./...

build-skia:
	CGO_ENABLED=1 go build -tags skia ./...

test:
	go test ./...

test-skia:
	CGO_ENABLED=1 go test -tags skia ./...

backend-info:
	@go run -c 'import "matplotlib-go/backends"; import "fmt"; fmt.Print(backends.CapabilityMatrix())' || echo "Run 'go run ./examples/backends/info/main.go' for backend information"

cli:
	go run ./main.go --help
