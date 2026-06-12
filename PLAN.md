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

# Phase 2: Visual Parity Closure via Code Parity (RMSE ≤ 5)

**Goal:** every catalog case renders within `RMSE 5` of its Matplotlib
reference, or carries a documented, frozen tolerance exception — and the route
there is **code parity**: when a case diverges, find the upstream code path in
`third_party/matplotlib` (3.10.9) responsible for the difference and make the
Go implementation a faithful, idiomatic translation of it. Visual parity then
follows by construction instead of by tuning. The existing bans stand and are
the enforcement half of this principle: no example-source workarounds,
fixture-specific core branches, catalog-ID conditionals, or unexplained
empirical constants (`internal/examplecatalog.ValidationClusters`).

**Status (2026-06-12, measured `testdata/golden/` vs `testdata/matplotlib_ref/`
at HEAD, harness metric):** 165 paired cases; **131 at RMSE ≤ 5**, 52 below
RMSE 1, strict-text cases at exactly 0. The **MathText family — formerly the
headline residual — is closed**: `mathtext_integrals` 0.00,
`mathtext_matrices` 0.24, `mathtext_gallery` 2.28, `mathtext_fractions` 4.01;
only `mathtext_basic` (5.33) and `mathtext_inline_labels` (5.88) sit marginally
above target, and their diffs are general text placement, not mathtext layout.
**34 cases remain above RMSE 5**, clustering into five workstreams.

## Workstreams (ordered by residual; each names the upstream source to translate)

- [ ] **W1 — mplot3d structural parity.** `mplot3d_gallery` 22.7,
      `mplot3d_tricontourf3d` 13.0, `mplot3d_contourf3d` 6.1,
      `mplot3d_fill_between3d` 5.3, `mplot3d_errorbar3d` 5.1, plus a tail just
      under target (`bar3d` 4.9, `terrain` 4.9, `surface3d` 4.7). The diff
      images show whole-axes divergence in *every* panel — pane/grid/tick
      placement and projected geometry, not one bad artist. Audit the Go port
      line-by-line against `mpl_toolkits/mplot3d/{proj3d,axis3d,axes3d,art3d}.py`
      (view/projection matrix, `_axinfo` constants, tick & label offset math,
      depth sorting), porting each computation faithfully.
- [ ] **W2 — geo/polar/radar projections.** `projection_toolkit_gallery` 20.2,
      `radar_basic` 6.9, `geo_mollweide_axes` 5.9, `geo_lambert_axes` 5.3
      (aitoff/hammer pass at 3.9 — use them as the behavioral baseline). Diffs
      concentrate on gridline paths and tick-label anchoring around the
      projection boundary. Port targets:
      `lib/matplotlib/projections/{geo,polar}.py` (gridline path generation,
      tick label positioning) and the axisartist twin-axes panel
      (`axisartist_showcase` 5.7).
- [ ] **W3 — ticks, scales, and inset placement.**
      `ticks_scales_formatters_gallery` 17.6, `date_concise_intraday_labels`
      5.8. The gallery diff isolates three distinct defects: (a) major/minor
      tick *positions* differ in the MultipleLocator and log panels, (b) the
      embedded "Categories" inset axes is placed wholly wrong, (c) custom-unit
      tick labels render doubled/offset. Port targets:
      `lib/matplotlib/{ticker,scale,dates,category,units}.py` for (a)/(c) and
      the inset-axes placement path for (b).
- [ ] **W4 — text layout: wrapping and rotated multiline.**
      `text_layout_gallery` 14.6. The diff isolates the residual to: the wrap
      point of `wrap=True` text (display-width logic in
      `Text._get_wrapped_text`), the rotated multiline block, and the
      `rotation_mode="anchor"` box. Port `lib/matplotlib/text.py`
      (`_get_layout`, `_get_wrapped_text`, rotation/anchor handling)
      faithfully; the unrotated alignment grid already matches.
- [ ] **W5 — the 5–7.3 band (≈ 21 cases).** `layout_bbox_helpers` 7.3,
      `plot_variants` 7.2, `axes_convenience_helpers` 7.2, `mesh_contour_tri`
      7.1, `legend_layout_matrix` 7.1, `line2d_markers` 7.0, `fill_stacked`
      6.9, `specialty_depth` 6.9, `mixed_raster_vector` 6.4, `arrays_showcase`
      6.3, `widgets_gallery` 6.2, `fill_basic` 6.1,
      `annotation_legend_offsetbox_gallery` 6.0, `clip_path_batch` 6.0,
      `mathtext_inline_labels` 5.9, `annotation_composition` 5.8,
      `fill_variants` 5.5, `unstructured_showcase` 5.5, `mathtext_basic` 5.3,
      `specialty_artists` 5.3, `axes_option_breadth_17_75_3` 5.3. Expect a few
      shared root causes (legend/offsetbox layout, fill-edge handling, label
      placement) rather than 21 independent bugs: diagnose each diff first,
      group by root cause, then fix each group as an upstream port
      (`legend.py`, `offsetbox.py`, collection/fill paths).

## Method (code parity, per failing case)

1. Inspect the committed diff artifact
   (`testdata/_artifacts/reference_compare/{id}_golden_vs_matplotlib_ref_diff.png`;
   regenerate with `just parity-viewer-print` or
   `go test ./test -run TestReferenceCompare`). Classify the residual:
   geometry, placement, text, or anti-aliasing.
2. Locate the upstream code path in `third_party/matplotlib`. Instrument both
   sides to find the *first diverging intermediate value* — Python via
   `PYTHONPATH=. python3 test/matplotlib_ref/plots/<id>.py` with temporary
   prints, Go via an env-gated probe in `test/diagnostics_test.go`.
3. Make the Go side a faithful idiomatic translation of the upstream
   computation. Cite the upstream file/function in the commit message so every
   fix carries its provenance.
4. Regold, confirm the metric *and* the visual diff, and check neighboring
   cases for regressions (`TestReferenceCompare` full run).

## Tolerance ratchet (new)

The committed per-case tolerances are far looser than today's actuals (e.g.
`mathtext_gallery` MaxRMSE 55 vs actual 2.3, `widgets_gallery` 120 vs 6.2,
`text_annotation_matrix` 42 vs 4.7), so the suite currently cannot catch large
regressions on already-good cases.

- [ ] After each workstream lands, ratchet the affected rows down to
      ≈ actual + small headroom.
- [ ] End state: closed cases use the package defaults (no per-row override);
      overrides remain only on documented, frozen exceptions.

**Exit criteria:**

- [ ] `TestReferenceCompare` records no catalog case above `RMSE 5` except those
      with a documented, frozen tolerance exception.
- [ ] Every parity fix names the upstream matplotlib code it translates
      (commit-message provenance), and is validated against the visual diff,
      not just the metric delta.
- [ ] Catalog tolerances are ratcheted so a regression of any closed case
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
