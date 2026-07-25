# Repository Guidelines

## Project Structure & Module Organization

- `core/`, `transform/`, `render/`, `style/`, `color/`, `cycler/`: plotting primitives and systems.
- `geom/`: public geometry types (points, rects, paths, affine).
- `optional/`: `optional.Value[T]`, the tri-state optional field used by option structs (see `docs/plans/phase2-options-model.md`).
- `backends/`: renderer implementations (`agg` with native cgo FreeType, `skia`, …).
- `canvas/`, `animation/`, `pyplot/`: canvas surface, animation support, pyplot-style convenience API.
- `tri/`: triangulation (faithful Qhull-build-order Delaunay port).
- `internal/`: unexported helpers, incl. `internal/examplecatalog` — the parity-case source of truth.
- `cmd/`: Cobra CLI (`root.go`, `version.go`, `backends.go`) plus subcommands `example/`, `parityviewer/`, `paritystatusdoc/`, `wasm/`, `webdemoexport/`, `vectorshowcase/`.
- `examples/`: runnable samples mirroring matplotlib galleries (also consumed by the parity harness).
- `test/`: catalog-driven parity tests (see below); unit tests live next to code as `*_test.go`. Repo-root `*_test.go` files are meta-tests (API freeze, doc coverage, file-size audit, PLAN.md coupling).
- `testdata/`: golden and reference images at the **repo root** (`golden/`, `matplotlib_ref/`, `pdf_golden/`, `svg_golden/`, `usetex_golden/`).
- `third_party/`: vendored matplotlib 3.10.9, FreeType 2.6.1, qhull.
- `benchmarks/`, `docs/`, `web/`: benchmarks, documentation, web/wasm demo.
- `main.go`: CLI entry.
- **Build prerequisite:** `go.mod` has `replace github.com/cwbudde/qhull-go => ../qhull-go` — a sibling `qhull-go` checkout is required to build.

## Build, Test, and Development Commands

- `just build`: compile all packages (`CGO_ENABLED=1 go build -tags freetype ./...`; auto-builds the vendored FreeType first).
- `just test`: run unit tests (`CGO_ENABLED=1 go test -tags freetype ./...`).
- `just lint`: `golangci-lint`, diff-scoped via `--new-from-merge-base=origin/main`. `just lint-full` checks the whole tree; `just lint-fix` applies fixes.
- `just fmt`: format via `treefmt` (uses `gofumpt` + `gci`; `prettier` for markdown).
- `just cli`: run the CLI (`go run ./main.go --help`).
- `just golden-update [TEST]`: regenerate golden images (all cases, or a `-run` pattern).
- `just parity-viewer`: visual golden-vs-reference diff viewer.
- Many more workflow targets exist (skia, web, text-parity, benchmarks, profiling): see `just --list`.
- Native Skia backend builds use `-tags "skia skiacgo"` + `SKIA_ROOT` and deliberately omit the `freetype` tag (duplicate FreeType symbols) — see the Justfile comments.

## Coding Style & Naming Conventions

- Go 1.25 (`go.mod`). Keep code idiomatic Go: short, lower-case package names; exported API uses `PascalCase`; unexported uses `camelCase`.
- No hidden global state; prefer explicit values and options.
- Formatting: run `treefmt --allow-missing-formatter` (configured for `gofumpt` and `gci`).
- Linting: `golangci-lint run --timeout=5m`; fix with `--fix`.

## Testing Guidelines

- Place tests beside code: `render/render_test.go`, `geom/geom_test.go` patterns.
- Name tests `TestXxx(t *testing.T)`; prefer table-driven tests for variations.
- Aim for deterministic behavior (no randomness without fixed seeds); avoid timing-based assertions.
- Run all packages: `just test`. For verbose: `CGO_ENABLED=1 go test -tags freetype -v ./...`.
- **Frozen public API:** changing the exported surface requires regenerating the frozen audit (`UPDATE_PUBLIC_API_AUDIT=1 go test -tags freetype -run TestStablePublicAPIMatchesFrozenAudit .`; golden in `test/testdata/public_api/`) and the parity-status doc (`go run ./cmd/paritystatusdoc`).

## Parity test catalog

The `test/` directory is flat and catalog-driven — do not add per-case test functions.

- **Source of truth:** `internal/examplecatalog.Case` carries the ID, factory, and per-case tolerances (`MinPSNR`, `MaxMeanAbs`, `MaxRMSE`; zero = defaults).
- **New parity case:** add a catalog row, drop `testdata/golden/{id}.png`, optionally drop `testdata/matplotlib_ref/{id}.png` + `test/matplotlib_ref/plots/{id}.py`. `TestGolden`, `TestMatplotlibRef`, and `TestReferenceCompare` discover it automatically. No test code edits needed.
- **Run one case:** `CGO_ENABLED=1 go test -tags freetype ./test/ -run TestGolden/{id}` (or `TestMatplotlibRef/{id}`, `TestReferenceCompare/{id}`); regex works too.
- **File responsibilities:** `helpers_test.go` (shared helpers), `golden_test.go` / `matplotlib_ref_test.go` / `reference_compare_test.go` (single subtest-loop each), `diagnostics_<family>_test.go` (env-gated dev probes, one file per diagnostic family — currently `alpha`, `bar_text`, `histogram_profile`, `nontext`, `rng`; add to the matching family file or start a new family), `contour_compare_test.go` and `artifact_test.go` (kept separate; bespoke).
- **No optional-visual gating:** every golden/reference case runs unconditionally (the historical `RUN_OPTIONAL_VISUAL_TESTS` gate was removed in Phase 18; renders cost ~0.05 s each). The strict text cases route through `strictMplRefIDs` in `helpers_test.go` for tighter thresholds, not for skipping.
- **Backend-conditional tests:** prefer a runtime skip (e.g., `agg.NativeFreetypeVersion() == ""`) over a `//go:build` constraint so the file compiles in all configurations. See `TestBarBasicTextPlacementDiagnostic`.
- **Side-by-side harness:** `test/parity/<id>/` holds per-case Go/Python plot pairs driving the `examples/` packages (`test/parity/run.py` renders the Python side); this is separate from the `test/matplotlib_ref/plots/` reference scripts.

## Parity testing

- Matplotlib parity is the primary goal for this port: prefer changing core library behavior over changing examples or fixtures when rendered output diverges.
- When trying to achieve parity with Matplotlib, always compare with the original code of matplotlib at ./third_party. **`./third_party/matplotlib` and the locally installed matplotlib are both 3.10.9.** You can regenerate/instrument a reference directly: `PYTHONPATH=. python3 test/matplotlib_ref/plots/<id>.py --output-dir /tmp/x` (or `test/matplotlib_ref/generate.py --output-dir <dir>` for all).
  - **3.10.9 migration (2026-05-29/30):** system + third_party were upgraded 3.8.4 → 3.10.9 and `testdata/matplotlib_ref/*.png` were regenerated under 3.10.9. The Go port's text layout, hinting (`force_autohint`, `hinting_factor=8`), font sizing, and title positioning already match 3.10.9. **Correction:** an earlier note claimed 3.10.9 "rewrote mathtext internals" with an "axis-height based `_genfrac` and per-font TeX-table sub/superscript constants" — this is **wrong**. The vendored 3.10.9 `_mathtext.py` has no `axis_height` symbol, `_genfrac` is still `"="`-centered, `DejaVuSansFontConstants` is `pass`, and every FontConstantsBase sub/superscript constant is byte-identical to 3.8.4. So the mathtext port did **not** need an algorithm retarget. Text pixel parity is achieved by linking matplotlib's pinned **FreeType 2.6.1** — which the AGG backend now does **by default** (see below); `title_strict` and the strict-text cases match the references at RMSE ~0.
- If possible, try to inspect the rendered output visually and compare it with the original matplotlib output.
- Try to keep the examples as close to the original matplotlib examples as possible.
- The aim is to have core library parity with matplotlib, so don't tweak the examples to achieve parity, but rather tweak the core library to achieve parity with the original matplotlib output.

### FreeType 2.6.1 is the default (text parity)

matplotlib generates every reference image with **FreeType 2.6.1** (its pinned version) and **DejaVu Sans 2.35** (bundled). The AGG backend renders text via native FreeType bitmaps (autohinted, `hinting_factor=8`) using that same DejaVu 2.35 TTF, and **statically links the vendored FreeType 2.6.1 by default** — so glyph rasterization byte-matches the references (the autohinter changed between 2.6.1 and current system FreeType, costing ~20 RMSE on dense text; `title_strict` goes from RMSE ~20 to **0**). There is a single canonical golden set in `testdata/golden/`; no per-FreeType-version split.

- **Prerequisite:** `just freetype261-build` builds a static FreeType 2.6.1 into `third_party/freetype/prefix/` (gitignored; pinned tarball/sha mirror `third_party/matplotlib/subprojects/freetype-2.6.1.wrap`). Every cgo `just` target (`build`, `test`, `lint`, `golden-update`, `parity-viewer`, …) depends on it, so the prefix is built automatically. The cgo flags live in `backends/agg/freetype_native.go` (unconditional, `${SRCDIR}`-relative).
- **Compile fallback:** `-tags systemfreetype` links the system FreeType via pkg-config for environments without the vendored prefix (IDE/gopls, quick `go vet`). It is **not** parity-exact — golden/reference tests are expected to diverge — and exists only so the cgo packages compile.
- The guard test `TestNativeFreetypeIsPinned261` (agg, default cgo build) asserts the linked version is 2.6.1; it is skipped under `-tags systemfreetype`.

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
- Current focus areas include geometry (`geom`), transforms (`transform`), and keeping backend capabilities/documentation aligned with what each renderer actually implements.
