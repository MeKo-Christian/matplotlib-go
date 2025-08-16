# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Matplotlib-Go is a Go-native plotting library inspired by Matplotlib, designed to be renderer-agnostic with a `Figure → Axes → Artists` hierarchy. The project emphasizes deterministic outputs, cross-platform consistency, and high-quality rendering.

## Development Commands

**Building and Testing:**

- `make build` or `just build` - compile all packages
- `make test` or `just test` - run all tests
- `go test ./...` - run tests directly

**Code Quality:**

- `make lint` or `just lint` - run golangci-lint (required for CI)
- `make lint-fix` or `just lint-fix` - auto-fix linting issues
- `make fmt` or `just fmt` - format code using treefmt

**CLI Testing:**

- `make cli` or `just cli` - run the CLI help command
- `go run ./main.go --help` - test CLI directly

## Architecture

The codebase follows a modular, backend-agnostic design:

**Core Packages:**

- `core/` - Artist tree (Figure, Axes, Artist interface) and rendering traversal
- `render/` - Renderer interface abstraction for multiple backends
- `style/` - Global styling defaults and configuration (RC system)
- `transform/` - Coordinate transforms and scales (Linear, Log, Affine)
- `color/` - Color handling and colormaps
- `internal/geom/` - Geometric primitives and utilities

**Key Concepts:**

- **Artist Interface**: All drawable elements implement `Draw(renderer, context)`, `Z()`, and `Bounds()`
- **Figure/Axes Hierarchy**: Figure contains Axes, Axes contain Artists
- **Transform Chain**: Data coordinates → Axes coordinates → Pixel coordinates
- **DrawContext**: Carries per-draw state (transforms, styling, clipping)
- **Z-ordering**: Artists are automatically sorted by Z-value for consistent layering

**Development Phase**: Currently in Phase A (scaffolding). Many packages contain doc.go stubs with implementation planned for later phases.

## Requirements

- Go 1.24+ (uses Go 1.24.0 in go.mod)
- golangci-lint (for linting)
- treefmt (optional, for formatting)

## Testing Strategy

Run `go test ./...` to execute the full test suite. The project emphasizes deterministic testing and will include golden image tests for visual output validation.
