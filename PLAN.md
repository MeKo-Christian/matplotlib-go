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

**Current status (2026-06-14):** Phase 2 is closed against the current
Matplotlib 3.10.9 reference set. W1-W5 closed the old high-residual families:
mplot3d, projections, axisartist/projection-toolkit leftovers, text
wrapping/rotated layout, MathText and annotation tails, legend/offsetbox layout,
fills/collections, contour/mesh, arrays, widgets, and mixed raster/vector. The
June 14 full optional `TestReferenceCompare` sweep has no catalog row above
`RMSE 5`; after the 2026-06-18 concise-date offset placement fix, the
highest rows are `imshow_transformed` 4.97, `geo_mollweide_axes` 4.92,
`legend_layout_matrix` 4.86, `spectrum_variants` 4.85, `boxplot_basic` 4.84,
`formatter_engineering_labels` 4.80, and `mathtext_basic` 4.77. The
AGG-native reference sweep is also below `RMSE 5`;
its highest row is `clip_path_batch` at 4.58.

The June 13 regression queue is retained below as closure history. The catalog
RMSE tolerances have been ratcheted to the refreshed metrics plus small headroom,
with all reference-comparable rows capped at or below `RMSE 5`.

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
      selection, category inset placement, category/unit conversion, custom unit
      formatter behavior, function-log minor tick defaults, and W3 category grid
      z-order. The refreshed full sweep is below `RMSE 5`.
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
      `axes_option_breadth`.
      `line2d_markers` was tightened again on 2026-06-18 by routing Line2D
      legend marker samples through the backend marker-batch path and matching
      Matplotlib's handler center semantics; current committed metric: RMSE
      2.77, PSNR 59.56 dB, MeanAbs 0.04.

## Current Regressions

Measured on 2026-06-13 with:

```bash
RUN_OPTIONAL_VISUAL_TESTS=true rtk proxy go test ./test -run 'TestReferenceCompare/(stat_variants|errorbar_basic|date_concise_intraday_labels|units_categories|axes_grid1_showcase|scale_function_defaults|ticks_scales_formatters_gallery|named_colors|formatter_log_mathtext_labels)$' -count=1 -v
```

- [x] `stat_variants`: fixed on 2026-06-14. The mismatch was grid z-order over
      filled statistical patches; the Go showcase now mirrors Matplotlib's
      `set_axisbelow(True)` before adding each y-grid. Updated committed
      golden/reference metric: RMSE 4.49, PSNR 53.24 dB, MeanAbs 0.24.
- [x] `errorbar_basic`: fixed on 2026-06-14 and tightened again on
      2026-06-14. Go `Axes.ErrorBar` now matches Matplotlib's public
      `capsize` semantics (`Line2D` cap marker length is `2*capsize`), default
      `errorbar.capsize=0`, default errorbar linewidth delegation, and limit
      caret drawing when caps are disabled; the legend errorbar sample now uses
      the converted cap marker length without doubling it again. The old Python
      reference workaround that halved `capsize` was removed. Targeted
      `TestGolden` and `TestReferenceCompare` pass for `errorbar_basic`,
      `specialty_depth`, `legend_layout_matrix`, and
      `axes_option_breadth`.
- [x] `date_concise_intraday_labels`: tightened on 2026-06-18 by porting
      Matplotlib's X-axis offset-text placement rule: the offset anchor is
      derived from the bottom tick-label bounding-box union plus the fixed
      3 pt `OFFSETTEXTPAD`, rather than an approximate tick-pad/font-size gap.
      Current committed golden/reference metric: RMSE 0.05, PSNR 74.01 dB,
      MeanAbs 0.00.
- [x] `units_categories`: fixed on 2026-06-13. The remaining mismatch was grid
      z-order, not category mapping: Matplotlib's `set_axisbelow(True)` places
      grids at z=0.5 below default bar patches. Go now exposes `SetAxisBelow`
      and the fixture applies it. Updated committed golden/reference metric:
      RMSE 4.44, PSNR 59.72 dB, MeanAbs 0.03.
- [x] `axes_grid1_showcase`: fixed on 2026-06-14. The remaining mismatch was
      axes/image chrome around the small RGB panels: equal-aspect axes now use
      Matplotlib's figure-fraction anchoring, AGG nearest image placement keeps
      half-pixel top anchors on the Matplotlib side of the rounding boundary,
      and Y spine snapping preserves half-pixel centers. Updated committed
      golden/reference metric: RMSE 1.85, PSNR 52.30 dB, MeanAbs 0.13.
- [x] `scale_function_defaults`: fixed on 2026-06-13. Go now installs
      Matplotlib-style default auto minor `LogLocator` ticks for `functionlog`
      scales and uses the unsnapped Matplotlib y-label bbox extent. Updated
      committed golden/reference metric: RMSE 0.56, PSNR 62.53 dB, MeanAbs
      0.02.
- [x] `ticks_scales_formatters_gallery`: below target in the refreshed
      2026-06-14 sweep after the focused W3 rows moved. Current committed
      golden/reference metric: RMSE 4.76, PSNR 56.77 dB, MeanAbs 0.11.
- [x] `named_colors`: below target in the refreshed 2026-06-14 sweep. Current
      committed golden/reference metric: RMSE 4.26, PSNR 49.74 dB, MeanAbs
      0.47.
- [x] `formatter_log_mathtext_labels`: below target in the refreshed
      2026-06-14 sweep. Current committed golden/reference metric: RMSE 4.53,
      PSNR 59.05 dB, MeanAbs 0.06.

## Remaining Work

- [x] **R1 — Reproduce and classify each regression visually.** Regenerate the
      focused artifacts, inspect golden/reference/diff side-by-side, and classify
      each case as geometry, locator/formatter, text placement, color conversion,
      collection/path stroke, image/layout, or backend antialiasing. Add only
      env-gated diagnostics in `test/diagnostics_test.go` when pixel inspection
      is not enough.
- [x] **R2 — Fix failing catalog rows first.** Prioritize
      `scale_function_defaults` and `units_categories` because they currently
      fail their ratcheted tolerances, then handle the above-target cases that
      still pass only due to broad tolerances.
- [x] **R3 — Re-open W3 where needed.** The date/unit/scale/formatter regressions
      likely share W3 code paths. Compare against
      `third_party/matplotlib/lib/matplotlib/{ticker,scale,dates,category,units}.py`
      and fix the shared computation rather than patching individual fixtures.
- [x] **R4 — Re-check showcase/layout families.** For `axes_grid1_showcase`, use
      `third_party/matplotlib/lib/mpl_toolkits/axes_grid1` as the source of
      truth. For `stat_variants` and `errorbar_basic`, compare against
      `axes/_axes.py`, `collections.py`, `lines.py`, and `legend_handler.py`.
      For `named_colors`, compare against `matplotlib.colors` data and
      conversion behavior.
- [x] **R5 — Full reference sweep and tolerance ratchet.** After fixes, run the
      full optional `TestReferenceCompare` catalog, regold only cases whose core
      behavior changed, and ratchet `internal/examplecatalog.Case` tolerances to
      actual metrics plus small headroom. Remove broad tolerances where defaults
      are enough; keep exceptions only when documented as frozen renderer-level
      differences.

Closed on 2026-06-14 with:

```bash
RUN_OPTIONAL_VISUAL_TESTS=true rtk proxy go test ./test -run '^TestReferenceCompare$' -count=1 -v
RUN_OPTIONAL_VISUAL_TESTS=true rtk proxy go test ./test -run '^TestAGGNativeReferenceCompare$' -count=1 -v
rtk proxy go test ./core ./backends/agg -count=1
```

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

- [x] No `TestReferenceCompare` catalog case is above `RMSE 5` except documented,
      frozen renderer-level exceptions.
- [x] The current regression list above is either below `RMSE 5` or has a clear,
      reviewed exception with the responsible validation cluster recorded.
- [x] Catalog tolerances are tight enough that a regression of any closed case
      fails CI.

---

# Phase 3: Parity Status Reporting

**Goal:** finish and keep `docs/matplotlib-parity-status.md` as the single
human-readable parity surface, generated from the machine inventories
(`internal/examplecatalog.PublicSurfaceParityRows` over the committed
`test/testdata/parity_surface/upstream_public_surface.json`, plus
`BrowserDemoCoverageRows` / `FeatureCoverageMatrix`).

The doc already exists with Feature Coverage, Browser Demo Coverage, Public
Surface Summary, Closure Owner Summary, Open Public Surface Rows, and now
**Per-Family Status** sections, and the browser-side CI gates are in place
(a `Showcase: true` row without a browser accounting row, or a browser demo
referencing a non-catalog family, both fail CI).

- [x] One table per upstream feature family with columns: upstream API / registry
      item, Go status (`direct-equivalent` / `idiomatic-equivalent` / `partial` /
      `not-started` / `intentional-omission`), local API, parity fixture, user
      example, browser demo, and remaining work — generated by
      `cmd/paritystatusdoc/` via `internal/examplecatalog.MatplotlibParityStatusMarkdown`.
- [x] CI fails when an upstream public row or enumerable registry item is tracked
      but unclassified (`TestAllUpstreamPublicRowsAreClassified`).
- [x] Every `partial`, `not-started`, and `intentional-omission` row has a
      rationale and a next action (`TestPartialAndOmissionRowsHaveNotes`).

**Exit criterion:**

- [x] A developer can open `docs/matplotlib-parity-status.md` and see, per tracked
      upstream feature, whether it is ported and whether it has examples / a
      browser demo, with CI guarding completeness.

---

# Phase 4: Large File Decomposition

**Goal:** reduce the highest-maintenance source and test files into focused
units without changing behavior, public API, golden fixtures, or catalog
semantics. This phase is deliberately mechanical: move code by responsibility,
keep package boundaries stable, and verify after every batch so parity work does
not get mixed with structural churn.

## Scope

Target files are the tracked Go files at or above roughly 1k lines and the
large generated/catalog artifacts that affect review quality:

- Production hotspots: `core/contour.go`, `core/axis.go`, `core/text.go`,
  `core/axes3d_contour_surface.go`, `core/legend.go`, `core/colorbar.go`,
  `core/plot.go`, `core/scatter.go`, `core/arrow_patch.go`,
  `style/mplstyle.go`, `pyplot/pyplot.go`, `cmd/parityviewer/main.go`,
  `backends/agg/{agg_paths.go,freetype_native.go}`, `backends/gobasic`,
  `backends/ps`, and `backends/pgf`.
- Test hotspots: `core/axes3d_test.go`, `core/text_test.go`,
  `core/axis_test.go`, `core/mesh_contour_test.go`, `core/legend_test.go`,
  `core/patch_test.go`, backend test files, `pyplot/pyplot_test.go`,
  `canvas/widget_interaction_test.go`, and `test/diagnostics_test.go`.
- Data/catalog files: `internal/examplecatalog/public_surface_parity.go`,
  `color/named_colors_data.go`, `docs/matplotlib-parity-status.md`, and the
  large JSON/SVG/PNG fixtures. Treat these as generated or fixture-like unless
  review pain justifies a generator or sharded loader.

## Work Plan

- [x] **L1 — Add a repeatable large-file audit.** Add a small documented command
      or `just` target that reports tracked Go files above 1k lines and tracked
      non-Go artifacts above 256 KiB. Record the initial inventory in
      `docs/large-file-decomposition.md` so future splits are measured against a
      stable baseline.
- [x] **L2 — Split tests first.** Split large `*_test.go` files by behavior
      family while keeping package names and helper visibility unchanged. Start
      with `core/axes3d_test.go`, then `core/text_test.go`,
      `core/axis_test.go`, `core/mesh_contour_test.go`, and the backend test
      files. Run the affected package tests after each file family.
  - [x] `core/axes3d_test.go`: split into projection/view/limits,
        scatter-mappables-colorbar, wire/surface/trisurf/voxels, frame/ticks/text,
        contour, and shared `axes3d_test_helpers.go`.
  - [x] `core/text_test.go`: split into draw/layout, multiline layout, bbox,
        annotation, MathText/TeX, and shared text recorder/helper files.
  - [x] `core/axis_test.go`: split into limits/scales, ticks/formatters,
        label positioning, grid/spines/frame, polar, and shared axis test
        helpers.
  - [x] `core/mesh_contour_test.go`: split mesh/pcolor/hist2d tests from
        contour/clabel/tri-contour tests.
  - [x] `core/legend_test.go`: split entry collection, layout/draw, sample
        drawing, best placement, and helper recorders.
  - [x] `core/patch_test.go`: split rectangle/fancybox, connection styles,
        arrow styles, connection patch, and miscellaneous patch classes.
  - [x] Backend tests: split `backends/svg/svg_test.go`,
        `backends/agg/agg_test.go`, and `backends/pdf/pdf_test.go` by lifecycle,
        paths/hatches, markers/batches, images, text/fonts, and clipping or
        serialization.
  - [x] Stateful/runtime tests: split `pyplot/pyplot_test.go` by registry/layout,
        wrappers, backend/show/save, and rc/style; split
        `canvas/widget_interaction_test.go` by buttons/sliders, check/radio/textbox,
        selectors, cursor, and picking.
  - [x] Diagnostics: split or annotate `test/diagnostics_test.go` by diagnostic
        family if it keeps growing, while preserving env-gated behavior.
- [ ] **L3 — Split algorithm-heavy core files.** Split `core/contour.go` into
      API/set construction, level selection, line extraction, filled-band
      geometry, label placement, and geometry helpers. Then split `core/axis.go`
      into types/defaults, spine/frame drawing, ticks, tick-label layout, and
      polar helpers. These are the highest-value production splits because they
      currently mix public API, drawing, geometry algorithms, and layout logic.
  - [x] `core/contour.go`: extract public API and `ContourSet` construction into
        `contour_api.go`.
  - [x] `core/contour.go`: extract coordinate normalization, triangulation, and
        level/locator helpers into `contour_levels.go`.
  - [x] `core/contour.go`: extract line segment generation, stitching, structured
        grid ordering, and boundary orientation into `contour_lines.go`.
  - [x] `core/contour.go`: extract filled band polygon clipping, saddle handling,
        triangle bands, and color mapping into `contour_filled.go`.
  - [x] `core/contour.go`: extract clabel placement, inline erasing, label angle,
        formatter, and text-width helpers into `contour_labels.go`.
  - [x] `core/axis.go`: extract axis side/type definitions and constructors into
        `axis_types.go`.
  - [x] `core/axis.go`: extract spine/frame drawing, snapping, and spine position
        helpers into `axis_spine.go`.
  - [x] `core/axis.go`: extract major/minor tick drawing, tick target counts, and
        tick direction/style helpers into `axis_ticks.go`.
  - [x] `core/axis.go`: extract tick-label drawing, offset text, label bounds, and
        label-origin math into `axis_ticklabels.go`.
  - [x] `core/axis.go`: extract polar spine/tick/tick-label behavior into
        `axis_polar.go`.
- [x] **L4 — Split text, 3D, and presentation helpers.** Split `core/text.go`
      into text/annotation API, multiline layout, bbox drawing, annotation
      arrows, and coordinate helpers. Split `core/axes3d_contour_surface.go`
      into 3D contour lines, filled contours, surface/trisurf generation, and
      compound contour path helpers. Split `core/legend.go` and
      `core/colorbar.go` by layout, mapping/configuration, drawing, and helper
      responsibilities.
  - [x] `core/text.go`: extract text/annotation option structs, constructors, and
        public methods into `text_api.go`.
  - [x] `core/text.go`: extract single-line/multiline measurement, rotated
        layout, wrapping, and alignment helpers into `text_layout.go`.
  - [x] `core/text.go`: extract bbox rectangle/path calculation and bbox drawing
        into `text_bbox.go`.
  - [x] `core/text.go`: extract annotation arrow drawing, arrow clipping, and
        coordinate conversion helpers into `annotation.go`.
  - [x] `core/axes3d_contour_surface.go`: extract 3D contour line and
        tri-contour line projection helpers into `axes3d_contour.go`.
  - [x] `core/axes3d_contour_surface.go`: extract filled contour, tri-filled
        contour, compound path, and band-polygon helpers into
        `axes3d_contourf.go`.
  - [x] `core/axes3d_contour_surface.go`: extract surface, trisurf, sampling, and
        surface polygon projection helpers into `axes3d_surface.go`.
  - [x] `core/legend.go`: extract layout calculation and best-placement avoidance
        into `legend_layout.go`.
  - [x] `core/legend.go`: extract entry collection, handler overrides, stem
        detection, and option-to-entry conversion into `legend_entries.go`.
  - [x] `core/legend.go`: extract line/scatter/errorbar/patch sample drawing into
        `legend_samples.go`.
  - [x] `core/colorbar.go`: extract placement and parent-axes geometry into
        `colorbar_layout.go`.
  - [x] `core/colorbar.go`: extract scale/tick/boundary mapping and extension
        value helpers into `colorbar_scale.go`.
  - [x] `core/colorbar.go`: extract draw, overlay, outline, extension path, and
        divider rendering into `colorbar_draw.go`.
- [ ] **L5 — Split facade and tool entrypoints.** Split `pyplot/pyplot.go` into
      registry/current-state management, figure/axes helpers, plotting wrappers,
      rc/style helpers, and backend/show/save management. Split
      `cmd/parityviewer/main.go` into CLI flags, case loading, image comparison,
      rerender commands, and HTML rendering.
  - [ ] `pyplot/pyplot.go`: extract figure registry, current figure/axes,
        manager cache, close/clear, and reset-for-tests behavior into
        `state.go`.
  - [ ] `pyplot/pyplot.go`: extract figure, axes, subplot, mosaic, divider,
        parasite, and 3D axes helpers into `layout.go`.
  - [ ] `pyplot/pyplot.go`: extract stateful plotting wrappers by family:
        lines/scatter/3D, bars/fills/stats/errorbar, images/mesh/signal,
        contour/triangulation, vector fields, widgets/specialty.
  - [ ] `pyplot/pyplot.go`: extract labels, limits, ticks, scales, grid,
        colorbar, legend, and figure text wrappers into `axes_wrappers.go`.
  - [ ] `pyplot/pyplot.go`: extract rc/style helpers into `rc.go`.
  - [ ] `pyplot/pyplot.go`: extract backend selection, savefig, show/draw/pause,
        event connect/disconnect, and show-handler canvas into `backend.go`.
  - [ ] `cmd/parityviewer/main.go`: extract flag parsing and repo/path option
        resolution into `cli.go`.
  - [ ] `cmd/parityviewer/main.go`: extract directory/parity case loading and
        printed case summaries into `cases.go`.
  - [ ] `cmd/parityviewer/main.go`: extract rerender command construction and
        execution into `rerender.go`.
  - [ ] `cmd/parityviewer/main.go`: extract image loading, compositing, metrics,
        raw diff, amplified diff, and PNG encoding into `images.go`.
  - [ ] `cmd/parityviewer/main.go`: extract HTML page/card rendering and CSS
        constants into `html.go`.
- [ ] **L6 — Split backend implementation files.** Split AGG path preparation,
      snapping/simplification/chunking, hatching, FreeType measurement, text
      drawing, and MathText rasterization into focused files. Apply the same
      pattern to GoBasic, PS, and PGF: renderer lifecycle, paths, images, text,
      document serialization, and shared utility helpers.
  - [ ] `backends/agg/agg_paths.go`: extract path preparation, finite filtering,
        bounds/culling, snapping, simplification, chunking, Gouraud color math,
        and hatching into focused files.
  - [ ] `backends/agg/freetype_native.go`: split native text drawing,
        measurement/cache, FreeType run/C interop, MathText image rendering, and
        alpha-mask utilities.
  - [ ] `backends/gobasic/gobasic.go`: split renderer lifecycle/state,
        path rasterization/clipping masks, image transforms/scaling, text
        rendering, blending, and utility conversion helpers.
  - [ ] `backends/ps/ps.go`: split renderer lifecycle/rasterization, path and
        hatch serialization, marker/path procedure reuse, images, text/font
        paths, document output, and PostScript formatting utilities.
  - [ ] `backends/pgf/pgf.go`: split renderer lifecycle/rasterization, path and
        hatch serialization, marker/path macro reuse, images, text/TeX escaping,
        document options, and PGF formatting utilities.
  - [ ] Keep backend split batches backend-local and verify with targeted package
        tests such as `go test ./backends/agg`, `go test ./backends/gobasic`,
        `go test ./backends/ps`, and `go test ./backends/pgf`.
- [ ] **L7 — Decide generated-data strategy.** For
      `internal/examplecatalog/public_surface_parity.go` and
      `color/named_colors_data.go`, either document them as generated/catalog
      data that intentionally stays large, or introduce source data plus
      generation checks. For goldens and binary fixtures, keep files intact and
      avoid artificial splitting.
- [ ] **L8 — Verify and commit in small batches.** Each batch should be mostly
      move-only and end with `just fmt` plus targeted `go test` commands. Before
      closing the phase, run `just fmt && just lint && just test` and a focused
      optional parity sweep for files that touched rendering behavior.

**Exit criteria:**

- [ ] No non-generated Go source file above 1k lines remains without an explicit
      reason recorded in `docs/large-file-decomposition.md`.
- [ ] Large tests are grouped by behavior family, with shared helpers moved into
      dedicated helper files.
- [ ] Generated/catalog/fixture files have an explicit keep-large, generate, or
      shard decision documented.
- [ ] Full formatting, linting, unit tests, and relevant parity checks pass after
      the decomposition.

---

# Phase 5: Documentation, Performance, and v1.0 Release

**Goal:** make the project consumable by users who have not followed the
development thread, establish performance baselines, and tag a stable v1.0.
*(The Matplotlib migration guide, the backend-selection guide
`docs/backend-selection.md`, the showcase caption/snippet review, the
intentional-divergence "anti-gallery", and the README browser-gallery entry
point are already done.)*

### 5.1 API Documentation

- [x] Package-level GoDoc passes for every stable public package, with a worked
      example per package. Guarded by
      `TestStablePublicPackagesHaveGoDocAndExamples`; the stable surface is the
      root command, public plotting packages, and public backend packages
      (excluding `cmd/*`, `internal/*`, `test/*`, and individual gallery
      `examples/*` packages).
- [x] Hosted documentation site (pkg.go.dev plus a curated landing page on the
      existing GitHub Pages deployment). The WASM gallery landing page now links
      to pkg.go.dev, README, examples gallery, backend selection, migration
      notes, and parity status.

### 5.2 Performance Pass

- [x] **P0 — Keep profiling repeatable and visible.** Initial harness lives under
      `benchmarks/`; first representative catalog + 100k scatter profiles are
      recorded in `docs/performance-profiling.md`. Remaining work:
  - [x] Add a `just bench-render` target that runs the representative catalog
        and 100k scatter benchmarks with `-benchmem`.
  - [x] Add a `just profile-render` target that writes CPU/heap profiles under
        `testdata/_artifacts/perf/` without requiring developers to remember
        the long `go test` invocations.
  - [x] Add CI regression tracking for the benchmark suite. Start as a
        report-only artifact, then promote to guarded thresholds once baseline
        variance is known. Implemented by
        `.github/workflows/benchmark-report.yml`.

- [x] **P1 — Low-risk text hot-path wins.** Catalog CPU is dominated by AGG
      native FreeType text measurement (`withNativeFreetypeRun` /
      `FT_Load_Glyph` / `FT_New_Face`) and allocation profiles show expensive
      text shaping/fallback churn. Do these before invasive renderer changes:
  - [x] Cache native FreeType face setup or measured text-run metrics by
        font-path, size, DPI, hinting factor, and text. Target: reduce
        catalog CPU share for `withNativeFreetypeRun` and repeated
        `FT_New_Face` calls without changing strict text parity.
  - [x] Memoize `fontFaceSupportsRune(face, rune)` and avoid repeated
        `fontFaceCacheKey` / `filepath.Clean` work inside text shaping loops.
        Target: reduce `ShapeTextRuns` / fallback allocation volume on
        `text_layout_gallery`, `mathtext_gallery`, and dense tick-label cases.
  - [x] Reuse short-lived shaping buffers where safe (`sfnt.Buffer`, feature
        slices, glyph slices). Keep the API immutable to callers.

- [x] **P1 — Reduce 100k scatter allocation pressure.** The 100k scatter
      benchmark is under one second on the profiling machine but allocates
      ~366 MB and ~3.7M objects per draw. The dominant source is per-marker
      path cloning/transformation in `PathCollection.drawPathCollection`,
      `applyAffinePath`, and AGG `devPath`.
  - [x] Add a focused benchmark threshold row for `BenchmarkLargeScatter100KDraw`
        so regressions are visible before optimizing.
  - [x] Combine display-space marker scale+translate into one affine path
        application, removing one full `geom.Path` allocation from the
        `PathInDisplay` scatter path.
  - [x] Add a renderer/backend fast path for repeated marker prototypes that
        transforms marker vertices into backend scratch storage instead of
        allocating a full `geom.Path` per point.
  - [x] Remove or reuse the extra AGG y-flip clone (`devPath`) for batched
        path-collection markers. Target: materially reduce B/op and allocs/op
        before chasing rasterizer CPU.

- [x] **P2 — Surface and image-copy reuse for repeated renders.** Catalog
      allocations show fresh AGG surfaces and `GetImage` copies are a major
      source of bytes allocated. This is acceptable for one-shot export but
      not ideal for interactive or animation redraw loops.
  - [x] Add a benchmark that redraws the same figure into a reused renderer.
  - [x] Document or expose a supported renderer-reuse path for long-running
        apps. Target: avoid allocating a new surface for every redraw.
  - [x] Avoid `GetImage` copies in benchmark/save paths that do not need an
        owned Go image.

- [x] **P2 — Cache scalar mapping setup.** `ScalarMapInfo.Color` /
      `color.GetColormap` is smaller than text and path batching but visible in
      image/collection-heavy catalog rows.
  - [x] Cache resolved colormap and norm state on scalar-mapped artists for
        draw-time reuse.
  - [x] Add focused benchmarks for scalar-mapped image, scatter, and mesh rows.

- [x] **Memory targets and tuning guide.** Convert the measured baselines into
      user-facing guidance:
  - [x] Define v1.0 memory targets for typical catalog plots, 100k scatter, and
        repeated interactive redraw.
  - [x] Document practical tuning advice: renderer reuse, avoiding unnecessary
        `GetImage`, batching markers, text-heavy tick-label costs, and backend
        selection tradeoffs.

### 5.3 Release Readiness

- [ ] Semantic-version policy decision and `CHANGELOG.md` baseline.
- [ ] Final golden / reference regeneration pass with per-case tolerances frozen
      for v1.0.
- [x] Public API stability audit: rename or hide any symbol not intended for the
      v1.0 surface. Closed on 2026-06-14 by freezing the stable exported Go API
      in `test/testdata/public_api/stable_public_api.json`, adding a CI-style
      audit test for accidental exported-surface drift, and promoting geometry
      primitives from `internal/geom` to the public `geom` package so renderer,
      transform, canvas, backend, and pyplot signatures do not expose
      unimportable internal types.
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
parity status report; **Phase 4** decomposes the largest source, test, and
catalog files into focused units; **Phase 5** delivers documentation,
performance baselines, and the v1.0 release.
