# Repository Guidelines

## Project Structure & Module Organization

- `cmd/`: Cobra CLI (`root.go`, `version.go`).
- `core/`, `transform/`, `render/`, `style/`, `color/`: plotting primitives and systems.
- `internal/geom/`: geometry types (points, rects, paths, affine).
- `examples/`: runnable samples (when added).
- `test/`: testing docs/assets; unit tests live next to code as `*_test.go`.
- `main.go`: CLI entry.

## Build, Test, and Development Commands

- `just build`: compile all packages (`go build ./...`).
- `just test`: run unit tests (`go test ./...`).
- `just lint`: run `golangci-lint` checks.
- `just fmt`: format via `treefmt` (uses `gofumpt` + `gci`).
- `just cli`: run the CLI (`go run ./main.go --help`).

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

## Parity test catalog

The `test/` directory is flat and catalog-driven — do not add per-case test functions.

- **Source of truth:** `internal/examplecatalog.Case` carries the ID, factory, and per-case tolerances (`MinPSNR`, `MaxMeanAbs`, `MaxRMSE`; zero = defaults).
- **New parity case:** add a catalog row, drop `test/testdata/golden/{id}.png`, optionally drop `test/testdata/matplotlib_ref/{id}.png` + `test/matplotlib_ref/plots/{id}.py`. `TestGolden`, `TestMatplotlibRef`, and `TestReferenceCompare` discover it automatically. No test code edits needed.
- **Run one case:** `go test ./test/ -run TestGolden/{id}` (or `TestMatplotlibRef/{id}`, `TestReferenceCompare/{id}`); regex works too.
- **File responsibilities:** `helpers_test.go` (shared helpers + optional-visual gating maps), `golden_test.go` / `matplotlib_ref_test.go` / `reference_compare_test.go` (single subtest-loop each), `diagnostics_test.go` (env-gated dev probes — add new diagnostics here, not as new files), `contour_compare_test.go` and `artifact_test.go` (kept separate; bespoke).
- **Optional-visual gating** (e.g., FreeType-only fixtures): add the case ID to `optionalVisualGoldenIDs` / `optionalVisualMplRefIDs` in `helpers_test.go`.
- **Backend-conditional tests:** prefer a runtime skip (e.g., `agg.NativeFreetypeVersion() == ""`) over a `//go:build` constraint so the file compiles in all configurations. See `TestBarTextDiagnostic`.

## Parity testing

- Matplotlib parity is the primary goal for this port: prefer changing core library behavior over changing examples or fixtures when rendered output diverges.
- When trying to achieve parity with Matplotlib, always compare with the original code of matplotlib at ./third_party. **`./third_party/matplotlib` and the locally installed matplotlib are both 3.10.9.** You can regenerate/instrument a reference directly: `PYTHONPATH=. python3 test/matplotlib_ref/plots/<id>.py --output-dir /tmp/x` (or `test/matplotlib_ref/generate.py --output-dir <dir>` for all).
  - **3.10.9 migration (2026-05-29/30):** system + third_party were upgraded 3.8.4 → 3.10.9 and `testdata/matplotlib_ref/*.png` were regenerated under 3.10.9. The Go port's text layout, hinting (`force_autohint`, `hinting_factor=8`), font sizing, and title positioning already match 3.10.9. **Correction:** an earlier note claimed 3.10.9 "rewrote mathtext internals" with an "axis-height based `_genfrac` and per-font TeX-table sub/superscript constants" — this is **wrong**. The vendored 3.10.9 `_mathtext.py` has no `axis_height` symbol, `_genfrac` is still `"="`-centered, `DejaVuSansFontConstants` is `pass`, and every FontConstantsBase sub/superscript constant is byte-identical to 3.8.4. So the mathtext port did **not** need an algorithm retarget. The remaining residual text divergence comes from a **FreeType version gap**: matplotlib refs use FreeType **2.6.1** (its pinned version) while the AGG backend links system FreeType (e.g. 2.13.2). True pixel parity requires building/linking FreeType 2.6.1 — see the `freetype261` build tag and `third_party/freetype/build.sh`.
- If possible, try to inspect the rendered output visually and compare it with the original matplotlib output.
- Try to keep the examples as close to the original matplotlib examples as possible.
- The aim is to have core library parity with matplotlib, so don't tweak the examples to achieve parity, but rather tweak the core library to achieve parity with the original matplotlib output.

## Commit & Pull Request Guidelines

- Commits: imperative mood, concise scope (e.g., `render: add NullRenderer stack checks`). Group mechanical changes separately from logic.
- PRs: include a clear description, linked issue (if any), and before/after screenshots for rendering/visual changes.
- Requirements:
  - All checks pass locally: `just fmt && just lint && just test`.
  - Add/adjust tests when changing behavior. Update `README.md`/docs when user-facing APIs change.
  - Keep changes focused; avoid drive-by refactors.

## Tasks & Planning

- Use `PLAN.md` for the living roadmap, priorities, and open questions.
- Always try to go phase by phase and milestone by milestone.
- When opening a PR, reference the relevant `PLAN.md` item/section.
- Update `PLAN.md` status if you complete or reshape a task.

## Architecture Notes

- Core concepts mirror Matplotlib: `Figure → Axes → Artists`; rendering is backend-agnostic.
- The shared renderer contract lives in `render/`, with optional capability interfaces for backend-specific features like text drawing, transformed images, DPI-aware text metrics, and PNG export.
- Current focus areas include geometry (`internal/geom`), transforms (`transform`), and keeping backend capabilities/documentation aligned with what each renderer actually implements.
