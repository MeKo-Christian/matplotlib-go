# Matplotlib-Go Development Plan

This plan tracks the **remaining** work to bring `matplotlib-go` to a stable
v1.0 release. The foundation is built; what follows is the focused path to the
finish. The roadmap is cross-checked against the local upstream Matplotlib
snapshot in `third_party/matplotlib` so uncovered areas stay explicit instead of
sliding into a vague "future work" bucket.

---

# Plan Tracking

- `✅` = done and stable
- `🧪` = implemented but under hardening
- `⚪` = in progress
- `⚠️` = deferred / design decision required
- `[ ]` = not started

---

# Already Shipped (Foundation Complete)

The project is well past proof-of-concept. The following are implemented,
tested, and stable:

- **Core model & transforms** — `Figure → Axes → Artists` with stale/callback
  propagation and draw scheduling; explicit transform graph (`transData`,
  `transAxes`, `transFigure`, blended/offset transforms, projection layer).
- **Backends (4)** — AGG (primary anti-aliased raster, native batches, buffer
  regions, filters), GoBasic (pure-Go correctness fallback), SVG (deterministic
  vector with structural diff harness), Skia (opt-in CPU raster). Capability
  matrix and save dispatch are registry-driven.
- **Publication vectors** — PDF, PS/EPS, and PGF via the shared save pipeline.
- **Plot vocabulary** — full 2D basics, statistical, color-mapped, vector-field,
  specialty, patches/collections, and 3D (`mplot3d`) families, plus non-Cartesian
  projections (polar, radar, skewx, geo).
- **Layout / style / API** — subplots, GridSpec, mosaic, subfigures, twin/secondary
  axes; `tight_layout` / `constrained_layout`; insets, dividers, image grids,
  axisartist; `rcParams` / themes / `.mplstyle`; OO core API plus a `pyplot` layer.
- **MathText & usetex** — first-class across the active backends.
- **Renderer effects** — patterns, gradients, and path effects (renderer-neutral
  pipeline with backend-native fast paths).
- **Interactive runtime** — headless event model, draw-idle/blit redraw, Gio
  desktop and WebAgg/WASM frontends, widgets and selectors.
- **Animation** — `FuncAnimation` / `ArtistAnimation` with GIF/APNG/HTML writers
  (dependency-free) and optional ffmpeg MP4/WebM.
- **Parity infrastructure** — catalog-driven `TestGolden` / `TestMatplotlibRef` /
  `TestReferenceCompare` with per-case tolerances; public-surface inventory and
  classification map; catalog-backed CLI and browser galleries.

*Former Phases 1–6, 8A, 9–19 are complete; see git history for the detailed
per-phase implementation logs.*

---

# Phase 1: Backend Deepening (Skia Native + GPU)

**Goal:** finish the backend-specific Skia work. The historical blocker (no Skia
C-ABI binding) is **resolved**: a real Skia library is built locally and a narrow
C-ABI wrapper links it under `-tags "skia skiacgo"`. The remaining work is wiring
the rest of the native primitives through that wrapper and standing up real GPU
mode — no longer external-access-blocked, just unbuilt.

## Done

- [x] **AGG parity diagnostics for non-text residuals** (the one self-contained,
      never-Skia-blocked item): `TestNonTextResidualDiagnostics`
      (`test/diagnostics_test.go`, env-gated by `MPL_GO_RESIDUAL_DIAG`) logs
      per-case residual metrics across dense path collections, translucent
      overlaps, image-interpolation modes, hatch clipping, and mixed
      raster/vector, dumping diff PNGs under
      `testdata/_artifacts/non_text_residuals/`.
- [x] **C-ABI Skia wrapper + cgo bridge** (`backends/skia/skia_cwrap.{h,cpp}`,
      `native_cgo.go`; tag `skiacgo`). Links Skia milestone 151. Native today:
      gradient path fills (`SkShaders` gradients), marker batches
      (`SkCanvas`/`SkPath`), Gouraud triangles (`SkVertices`). Verified by
      `native_cgo_test.go`. `IsCapabilityBridged` flips `MarkerBatch` to native
      (`✓`) when the native surface is linked; default `-tags skia` stays the
      pure-Go CPU bridge.

## Build / test (carry-on entry points)

- Build the library once (outside the repo): a Skia checkout with `include/` and
  a built `out/Shared/libskia.so`. For combined use with text, build with
  `skia_use_freetype=false` (see caveat).
- `just build-skia-native` / `just test-skia-native` (override `SKIA_ROOT`, default
  `/mnt/projekte/Code/skia`). They pass `CGO_CXXFLAGS`/`CGO_LDFLAGS` and add an
  rpath so the test binary finds `libskia.so` without `LD_LIBRARY_PATH`.
- **FreeType caveat:** a Skia built without system FreeType statically bundles its
  own. Combining `skiacgo` with the `freetype` tag (agg's vendored FreeType 2.6.1)
  in one binary produces duplicate `FT_*` symbols and a runtime crash, so the
  native recipes deliberately omit `freetype`. Rebuild Skia with
  `skia_use_freetype=false` to run native Skia and agg text in one binary.

## Remaining work (each unblocked; wire the wrapper entrypoint)

- [ ] **Native path collections.** Add a `drawPathCollectionNative` to the
      `nativeBatchBridge` interface + dispatch in `Renderer.DrawPathCollection`
      (`skia_native.go`); render all items into one native surface (loop
      `mgsk_draw_path`, or add a batched multi-path C entrypoint). Flip
      `pathcollectionbatch` in `IsCapabilityBridged`.
- [ ] **Native quad meshes.** Add `drawQuadMeshNative`; emit two `SkVertices`
      triangles per cell with the face color (reuse `mgsk_draw_vertices`) or one
      `mgsk_draw_path` per cell. Flip `quadmeshbatch`.
- [ ] **Native transformed images.** Add `mgsk_draw_image` (SkImage raster copy +
      `drawImageRect` with sampling) to the wrapper; implement
      `render.ImageTransformer` on the native path.
- [ ] **Native hatching via tiled `SkShader`.** Add a tiled-shader hatch
      entrypoint to the wrapper and route `NativeHatcher` through it instead of
      `render.DrawHatchFallback`; flip `nativehatcher`.
- [ ] **Real GPU mode** (`SkSurface::MakeRenderTarget`) behind `skiagpu` +
      `skiacgo`: add `mgsk_surface_new_gpu` (Ganesh `GrDirectContext` over GL or
      Vulkan) + flush; wire `FlushGPU`; keep deterministic CPU readback for
      goldens. Requires a platform GPU context (GL/Vulkan dev libs). The
      CPU-readback scaffold (`gpu_scaffold_test.go`) already exists.
- [ ] **Per-mode reporting divergence.** Once GPU is real, split
      `ModeCapabilities` (`strategy.go`) so CPU and GPU capability sets differ and
      the live `BackendComparisonReport` shows per-mode native/fallback/unavailable.
- [ ] **Parity hardening.** Add `skiacgo` tests comparing native output to the AGG
      goldens within tolerance (markers/gradients/Gouraud first), and update each
      `NativePathRequirements` row in `strategy.go` from `StatusDeferred` to
      `StatusImplemented` as its primitive lands.

**Exit criterion:**

- [ ] Skia is a viable secondary raster backend: native batch primitives + real
      GPU acceleration available under `skiacgo`/`skiagpu`, with parity-checked
      output and truthful per-mode capability reporting. *(Wrapper + first native
      primitives done; remaining items above.)*

---

# Phase 2: Visual Parity Closure (RMSE > 5)

**Goal:** every catalog case renders within `RMSE 5` of its Matplotlib
reference, or carries a documented, frozen tolerance exception. Core-library
fixes are preferred over fixture tweaks (no example-source workarounds,
fixture-specific core branches, catalog-ID conditionals, or unexplained
empirical constants — enforced by `internal/examplecatalog.ValidationClusters`).

**Status today:** the committed catalog tolerances all pass
(`TestReferenceCompare`). The dominant *unfinished* residual is the **MathText
family**, which remains well above target:

- `mathtext_fractions` (~RMSE 27), `mathtext_matrices` (~RMSE 25),
  `mathtext_inline_labels` (~RMSE 12), `mathtext_basic` (~RMSE 12).
- Likely core areas: `github.com/cwbudde/mathtext` layout — fraction-axis
  alignment, stretchy / `\genfrac` delimiter sizing, square-root geometry, matrix /
  stack ink bounds, sub/superscript placement — plus `core/mathtext.go`, AGG text
  bounds, and anchored-box / annotation layout.

**Workflow:**

- [ ] Regenerate the current over-threshold list before each work session:
      `just parity-viewer-print` (or `go test ./test -run TestReferenceCompare`).
      Visual artifacts land under `testdata/_artifacts/reference_compare/`
      (`*_golden.png`, `*_matplotlib_ref.png`, `*_…_diff.png`).
- [ ] Close the MathText family by following upstream `_mathtext.py`, not
      fixture-local offsets.
- [ ] For any case that cannot reach `RMSE 5`, record a documented frozen
      tolerance on the catalog row with an owner and rationale.

**Exit criteria:**

- [ ] `TestReferenceCompare` records no catalog case above `RMSE 5` except those
      with a documented, frozen tolerance exception.
- [ ] Each fix is validated against source parity and visual artifacts, not just
      the metric delta.

---

# Phase 3: Parity Status Reporting

**Goal:** finish and keep `docs/matplotlib-parity-status.md` as the single
human-readable parity surface, generated from the machine inventories
(`internal/examplecatalog.PublicSurfaceParityRows` over the committed
`test/testdata/parity_surface/upstream_public_surface.json`, plus
`BrowserDemoCoverageRows` / `FeatureCoverageMatrix`).

The doc already exists with Feature Coverage, Browser Demo Coverage, Public
Surface Summary, Closure Owner Summary, and Open Public Surface Rows sections,
and the browser-side CI gates are in place (a `Showcase: true` row without a
browser accounting row, or a browser demo referencing a non-catalog family, both
fail CI). Remaining work is the upstream-family detail and its guard:

- [ ] One table per upstream feature family with columns: upstream API / registry
      item, Go status (`direct-equivalent` / `idiomatic-equivalent` / `partial` /
      `not-started` / `intentional-omission`), local API, parity fixture, user
      example, browser demo, and remaining work — generated, not hand-written.
- [ ] CI fails when an upstream public row or enumerable registry item is tracked
      but unclassified.
- [ ] Every `partial`, `not-started`, and `intentional-omission` row has a
      rationale and a next action.

**Exit criterion:**

- [ ] A developer can open `docs/matplotlib-parity-status.md` and see, per tracked
      upstream feature, whether it is ported and whether it has examples / a
      browser demo, with CI guarding completeness.

---

# Phase 4: Documentation, Performance, and v1.0 Release

**Goal:** make the project consumable by users who have not followed the
development thread, establish performance baselines, and tag a stable v1.0.
*(The Matplotlib migration guide, the backend-selection guide
`docs/backend-selection.md`, the showcase caption/snippet review, the
intentional-divergence "anti-gallery", and the README browser-gallery entry
point are already done.)*

### 4.1 API Documentation

- [ ] Package-level GoDoc passes for every public package, with a worked example
      per package.
- [ ] Hosted documentation site (pkg.go.dev plus a curated landing page on the
      existing GitHub Pages deployment).

### 4.2 Performance Pass

- [ ] Profiling sweep across the catalog: find hotspots that exceed the
      100k-point smoothness goal and the sub-second typical-plot goal.
- [ ] Reusable benchmark suite under `benchmarks/` with CI regression tracking.
- [ ] Documented memory-usage targets and a tuning guide for long-running apps.

### 4.3 Release Readiness

- [ ] Semantic-version policy decision and `CHANGELOG.md` baseline.
- [ ] Final golden / reference regeneration pass with per-case tolerances frozen
      for v1.0.
- [ ] Public API stability audit: rename or hide any symbol not intended for the
      v1.0 surface.
- [ ] CI gate: `just fmt && just lint && just test` plus catalog-driven parity
      checks all pass on the release branch.
- [ ] Tag v1.0.

**Exit criteria:**

- [ ] A new user can install the module, follow the docs, and reproduce every
      showcase plot.
- [ ] The public API surface is documented, audited, and frozen for v1.0.
- [ ] Performance and parity baselines are tracked in CI.

---

# Development Guidelines

## Backend Strategy

- **Primary raster backend:** AGG (`backends/agg/`) — anti-aliased, sub-pixel
  accurate, reference for parity fixtures.
- **AGG port ownership:** if a parity failure is caused by a fundamental
  rasterization, text, path, transform, or blending issue in the Go AGG port,
  fix `../agg_go` rather than adding compensating behavior in this repository.
- **Pure-Go fallback:** GoBasic (`backends/gobasic/`) — dependency-light
  correctness fallback.
- **Primary vector backend:** SVG (`backends/svg/`) — deterministic,
  browser-readable, structurally tested.
- **Publication vector backends:** PDF / PS / PGF.
- **Accelerated raster backend:** Skia (`backends/skia/`) — opt-in CPU and future
  GPU paths.

## Testing Strategy

- Catalog-driven parity tests (`internal/examplecatalog.Case` + `test/`).
- Golden image tests for raster backends, structural diff for vector backends.
- Property-based tests for data ranges and transforms.
- Visual regression against Matplotlib references with documented per-case
  tolerances.
- `go test ./...` runs the full suite; `go test ./test/ -run <id>` runs one
  parity case.

## API Design Principles

- Follow Matplotlib conventions where sensible; document and explain divergences.
- Use functional options for configuration; keep zero-value defaults useful.
- Keep the object-oriented core API first-class; offer `pyplot` as a
  migration-friendly convenience layer.
- Provide escape hatches (renderer access, raw paths) for advanced cases.

## Performance Goals

- Handle datasets up to 100k points smoothly.
- Sub-second rendering for typical plots.
- Memory-efficient for long-running applications and animation.

## Examples-Driven Development

- Every feature gets a working example tied to the catalog.
- Examples serve as integration tests and gallery content.
- Showcase examples appear in the WASM browser gallery and the README.
- Examples demonstrate real-world usage rather than minimal API smoke tests.

---

This roadmap reflects the work remaining to bring matplotlib-go to a stable,
documented v1.0 release. **Phase 1** hardens the backend matrix (mostly the
deferred external Skia binding); **Phase 2** closes the remaining visual parity
gap (chiefly MathText); **Phase 3** finishes and guards the parity status
report; **Phase 4** delivers documentation, performance baselines, and the v1.0
release.
