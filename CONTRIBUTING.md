# Contributing

Thanks for your interest in contributing! This project aims to build a Matplotlib-like plotting library in Go.

## Getting started

- Requirements: Go 1.22+ (preferably 1.24), `golangci-lint`, and `treefmt` (optional).
- Clone and bootstrap:
  - `go mod download`
  - `just fmt` to format, `just lint` to lint, `just build` to compile.

## Commands

- `just build`: compile all packages.
- `just test`: run tests.
- `just lint`: run `golangci-lint`.
- `just fmt`: run formatters via `treefmt`.

## Style

- Keep packages cohesive: `core`, `transform`, `render`, `style`, `color`, and `geom`.
- Avoid global state; prefer explicit values and options.
- Ensure determinism and clear, stable APIs.

## CI

GitHub Actions run unit tests, formatting checks, and linters on pushes and PRs.

## License

Unless stated otherwise, contributions are under the repository’s LICENSE.
