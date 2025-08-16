# Repository Guidelines

## Project Structure & Module Organization
- `cmd/`: Cobra CLI (`root.go`, `version.go`).
- `core/`, `transform/`, `render/`, `style/`, `color/`: plotting primitives and systems.
- `internal/geom/`: geometry types (points, rects, paths, affine).
- `examples/`: runnable samples (when added).
- `test/`: testing docs/assets; unit tests live next to code as `*_test.go`.
- `main.go`: CLI entry.

## Build, Test, and Development Commands
- `make build` / `just build`: compile all packages (`go build ./...`).
- `make test` / `just test`: run unit tests (`go test ./...`).
- `make lint` / `just lint`: run `golangci-lint` checks.
- `make fmt` / `just fmt`: format via `treefmt` (uses `gofumpt` + `gci`).
- `make cli` / `just cli`: run the CLI (`go run ./main.go --help`).

## Coding Style & Naming Conventions
- Go 1.22+ (target 1.24). Keep code idiomatic Go: short, lower-case package names; exported API uses `PascalCase`; unexported uses `camelCase`.
- No hidden global state; prefer explicit values and options.
- Formatting: run `treefmt --allow-missing-formatter` (configured for `gofumpt` and `gci`).
- Linting: `golangci-lint run --timeout=5m`; fix with `--fix`.

## Testing Guidelines
- Place tests beside code: `render/render_test.go`, `internal/geom/geom_test.go` patterns.
- Name tests `TestXxx(t *testing.T)`; prefer table-driven tests for variations.
- Aim for deterministic behavior (no randomness without fixed seeds); avoid timing-based assertions.
- Run all packages: `go test ./...`. For verbose: `go test -v ./...`.

## Commit & Pull Request Guidelines
- Commits: imperative mood, concise scope (e.g., `render: add NullRenderer stack checks`). Group mechanical changes separately from logic.
- PRs: include a clear description, linked issue (if any), and before/after screenshots for rendering/visual changes.
- Requirements: 
  - All checks pass locally: `make fmt && make lint && make test`.
  - Add/adjust tests when changing behavior. Update `README.md`/docs when user-facing APIs change.
  - Keep changes focused; avoid drive-by refactors.

## Tasks & Planning
- Use `TASKS.md` for the living roadmap, priorities, and open questions.
- Always try to go phase by phase and milestone by milestone.
- When opening a PR, reference the relevant `TASKS.md` item/section.
- Update `TASKS.md` status if you complete or reshape a task.

## Architecture Notes
- Core concepts mirror Matplotlib: `Figure → Axes → Artists`; rendering is backend-agnostic. Current focus areas include geometry (`internal/geom`), transforms (`transform`), and a no-op renderer (`render.NullRenderer`) for traversal and testing.
