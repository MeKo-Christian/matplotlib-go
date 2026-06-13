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

- [x] **Native path collections.** Add a `drawPathCollectionNative` to the
      `nativeBatchBridge` interface + dispatch in `Renderer.DrawPathCollection`
      (`skia_native.go`); render all items into one native surface (loop
      `mgsk_draw_path`, or add a batched multi-path C entrypoint). Flip
      `pathcollectionbatch` in `IsCapabilityBridged`.
- [x] **Native quad meshes.** Add `drawQuadMeshNative`; emit two `SkVertices`
      triangles per cell with the face color (reuse `mgsk_draw_vertices`) or one
      `mgsk_draw_path` per cell. Flip `quadmeshbatch`.
- [x] **Native transformed images.** Add `mgsk_draw_image` (SkImage raster copy +
      `drawImageRect` with sampling) to the wrapper; implement
      `render.ImageTransformer` on the native path.
- [x] **Native hatching via tiled `SkShader`.** Add a tiled-shader hatch
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

# Phase 2: Visual Parity Closure via Code Parity (RMSE ≤ 5)

**Goal:** every catalog case renders within `RMSE 5` of its Matplotlib 3.10.9
reference, or carries a documented, frozen tolerance exception. The route remains
**code parity**: for each visual mismatch, find the responsible upstream path in
`third_party/matplotlib` and port the computation faithfully into Go. Do not use
example-source workarounds, fixture-specific core branches, catalog-ID
conditionals, or unexplained empirical constants.

**Current status (2026-06-13):** the original Phase 2 structural work is mostly
complete. W1-W5 closed the old high-residual families: mplot3d, projections,
axisartist/projection-toolkit leftovers, text wrapping/rotated layout, MathText
and annotation tails, legend/offsetbox layout, fills/collections, contour/mesh,
arrays, widgets, and mixed raster/vector. The latest full W5 scoreboard is below
`RMSE 5`; highest W5 cases were `legend_layout_matrix` 4.98,
`line2d_markers` 4.80, `mathtext_basic` 4.77, `specialty_artists` 4.62,
`widgets_gallery` 4.55, `annotation_legend_offsetbox_gallery` 4.48, and
`specialty_depth` 4.33.

Phase 2 is still open because a fresh focused sweep shows regressions outside the
closed W5 queue. Some currently pass only because their catalog tolerances are
still loose, so treat the list below as the active Phase 2 work queue before any
final tolerance ratchet.

## Completed Work

- [x] **W1 — mplot3d structural parity.** Closed the mplot3d family by porting
      contour alpha behavior, 3D tick-space / `MaxNLocator` handoff, shaded 3D
      fills, and filled-contour antialias semantics. The main gallery and all
      targeted 3D cases are now below `RMSE 5`.
- [x] **W2 — geo, polar, radar, and projection-toolkit parity.** Closed the
      projection family by matching polar radial-label transforms, geo theta
      formatting, `rlabel_position`, frame/spine semantics, AGG-style snap-auto
      gridline rendering, Lambert text half-tie rounding, Skew-T grid/spine
      placement, and axisartist/reference-line dash behavior.
- [x] **W3 — ticks, scales, dates, units, and inset placement.** Ported the
      large structural pieces: `MultipleLocator`, `AutoMinorLocator`, log locator
      and formatter handoff, date locator/formatter behavior, concise date label
      selection, category inset placement, category/unit conversion, and custom
      unit formatter behavior. This work must now be revisited because the
      current regression sweep shows several W3-family cases back above `RMSE 5`.
- [x] **W4 — text wrapping and rotated multiline layout.** Closed
      `text_layout_gallery` by porting Matplotlib-style wrap width calculation,
      literal-space wrapping, multiline `_get_layout` metrics, rotation-mode
      anchor bbox semantics, and unclipped default `Axes.Text` behavior.
- [x] **W5 — 5-7.3 cleanup band.** Closed all original W5 cases below `RMSE 5`:
      `layout_bbox_helpers`, `plot_variants`, `axes_convenience_helpers`,
      `legend_layout_matrix`, `line2d_markers`, `fill_stacked`,
      `specialty_depth`, `mesh_contour_tri`, `mixed_raster_vector`,
      `arrays_showcase`, `widgets_gallery`, `fill_basic`,
      `annotation_legend_offsetbox_gallery`, `clip_path_batch`,
      `mathtext_inline_labels`, `annotation_composition`, `fill_variants`,
      `mathtext_basic`, `specialty_artists`, and
      `axes_option_breadth_17_75_3`.

## Current Regressions

Measured on 2026-06-13 with:

```bash
RUN_OPTIONAL_VISUAL_TESTS=true rtk proxy go test ./test -run 'TestReferenceCompare/(stat_variants|errorbar_basic|date_concise_intraday_labels|units_categories|axes_grid1_showcase|scale_function_defaults|ticks_scales_formatters_gallery|named_colors|formatter_log_mathtext_labels)$' -count=1 -v
```

- [ ] `stat_variants`: RMSE 7.40, PSNR 50.28 dB, MeanAbs 0.52. Current catalog
      tolerance is loose enough that the subtest passes, but it is above the
      Phase 2 target. Start from statistical helper defaults and stack/violin/
      boxplot collection rendering.
- [ ] `errorbar_basic`: RMSE 6.98, PSNR 55.61 dB, MeanAbs 0.13. Subtest passes
      under existing tolerance but regressed above target. Start from data
      errorbar cap, marker-edge, limit-caret, legend, and line snap semantics.
- [ ] `date_concise_intraday_labels`: RMSE 6.67, PSNR 58.18 dB, MeanAbs 0.10.
      Subtest passes under the documented `MaxRMSE=7.0` exception, but it is no
      longer under the Phase 2 target. Re-check `ConciseDateFormatter` tick-level
      selection, offset suppression, and label baseline/rotation placement.
- [ ] `units_categories`: RMSE 6.36, PSNR 56.54 dB, MeanAbs 0.21. This currently
      fails its ratcheted catalog tolerance (`MinPSNR=57.0`, `MaxMeanAbs=0.15`,
      `MaxRMSE=5.0`). Re-check category unit mapping, tick placement, bar/text
      layout, and any shared W3 category inset assumptions.
- [ ] `axes_grid1_showcase`: RMSE 6.25, PSNR 50.22 dB, MeanAbs 0.22. Subtest
      passes under older broad tolerance but is above target. Start from divider,
      image-grid, inset, anchored artist, and colorbar layout against
      `mpl_toolkits.axes_grid1`.
- [ ] `scale_function_defaults`: RMSE 6.06, PSNR 59.03 dB, MeanAbs 0.08. This
      currently fails its ratcheted catalog tolerance (`MinPSNR=63.0`,
      `MaxMeanAbs=0.05`, `MaxRMSE=3.5`). Re-check function-scale transform,
      inverse transform sampling, autoscale, and default locator/formatter
      installation.
- [ ] `ticks_scales_formatters_gallery`: RMSE 5.80, PSNR 56.05 dB, MeanAbs 0.13.
      Subtest passes under the documented `MaxRMSE=7.0` exception, but remains
      above the Phase 2 target. Re-check the W3 gallery panels after fixing the
      focused `scale_function_defaults`, `units_categories`, and formatter cases.
- [ ] `named_colors`: RMSE 5.39, PSNR 49.57 dB, MeanAbs 0.50. Subtest passes
      under existing broad color tolerance but is above target. Re-check named
      color table values, swatch geometry, text placement, and color conversion
      edge cases.
- [ ] `formatter_log_mathtext_labels`: RMSE 5.28, PSNR 60.65 dB, MeanAbs 0.05.
      Subtest passes under existing broad formatter tolerance but is above
      target. Re-check `LogFormatterMathtext` exponent formatting, label-only-base
      behavior, minor-threshold sparsity, and MathText baseline placement.

## Remaining Work

- [ ] **R1 — Reproduce and classify each regression visually.** Regenerate the
      focused artifacts, inspect golden/reference/diff side-by-side, and classify
      each case as geometry, locator/formatter, text placement, color conversion,
      collection/path stroke, image/layout, or backend antialiasing. Add only
      env-gated diagnostics in `test/diagnostics_test.go` when pixel inspection
      is not enough.
- [ ] **R2 — Fix failing catalog rows first.** Prioritize
      `scale_function_defaults` and `units_categories` because they currently
      fail their ratcheted tolerances, then handle the above-target cases that
      still pass only due to broad tolerances.
- [ ] **R3 — Re-open W3 where needed.** The date/unit/scale/formatter regressions
      likely share W3 code paths. Compare against
      `third_party/matplotlib/lib/matplotlib/{ticker,scale,dates,category,units}.py`
      and fix the shared computation rather than patching individual fixtures.
- [ ] **R4 — Re-check showcase/layout families.** For `axes_grid1_showcase`, use
      `third_party/matplotlib/lib/mpl_toolkits/axes_grid1` as the source of
      truth. For `stat_variants` and `errorbar_basic`, compare against
      `axes/_axes.py`, `collections.py`, `lines.py`, and `legend_handler.py`.
      For `named_colors`, compare against `matplotlib.colors` data and
      conversion behavior.
- [ ] **R5 — Full reference sweep and tolerance ratchet.** After fixes, run the
      full optional `TestReferenceCompare` catalog, regold only cases whose core
      behavior changed, and ratchet `internal/examplecatalog.Case` tolerances to
      actual metrics plus small headroom. Remove broad tolerances where defaults
      are enough; keep exceptions only when documented as frozen renderer-level
      differences.

## Method

1. Inspect the committed diff artifact
   (`testdata/_artifacts/reference_compare/{id}_golden_vs_matplotlib_ref_diff.png`)
   or regenerate the focused `TestReferenceCompare` output.
2. Locate the upstream Matplotlib code path and instrument both sides to find the
   first diverging intermediate value. Use temporary Python prints with
   `PYTHONPATH=.` and env-gated Go diagnostics; do not commit probes unless they
   are generally useful diagnostics.
3. Port the upstream computation faithfully and idiomatically into the Go core.
   Prefer shared behavior fixes over fixture changes.
4. Regold only after the core fix is in place, verify the metric and the visual
   diff, and run neighboring catalog cases for regressions.

**Exit criteria:**

- [ ] No `TestReferenceCompare` catalog case is above `RMSE 5` except documented,
      frozen renderer-level exceptions.
- [ ] The current regression list above is either below `RMSE 5` or has a clear,
      reviewed exception with the responsible validation cluster recorded.
- [ ] Catalog tolerances are tight enough that a regression of any closed case
      fails CI.

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
gap via code parity with upstream (chiefly mplot3d, projections, ticks/scales,
and text layout — MathText is closed); **Phase 3** finishes and guards the
parity status report; **Phase 4** delivers documentation, performance
baselines, and the v1.0 release.
