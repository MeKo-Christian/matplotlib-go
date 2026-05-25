# Matplotlib-Go Development Plan

This plan tracks the remaining work to bring `matplotlib-go` to a v1.0 release. The
roadmap is cross-checked against the local upstream Matplotlib snapshot in
`third_party/matplotlib` so uncovered areas are tracked explicitly instead of
# Phase 12: Artist, Line2D, and Marker Semantics

**Goal:** close foundational `artist.py`, `lines.py`, and `markers.py` gaps that
affect visible parity and migration.

Status: completed and compacted.

- [x] Shared artist metadata, clipping, alpha, visibility, stale/in-layout, and
      centralized traversal behavior are implemented and covered by tests.
- [x] Marker catalogue, fillstyles, `markevery`, line cap/join/dash behavior,
      and focused marker parity matrices are implemented or explicitly
      classified.
- [x] Audit and public-surface rows for `artist.py`, `lines.py`, and
      `markers.py` are closed to precise statuses.

---

# Phase 13: Axis, Ticker, Formatter, Scale, and Transform Breadth

**Goal:** close axis/ticker/scale/transform parity gaps that drive persistent
tick and coordinate residuals.

Status: completed and compacted.

- [x] Tick lifecycle, offset/scientific text behavior, locator/formatter
      breadth, and shared/twin-axis interactions are covered.
- [x] Scale/transform breadth includes log/symlog/logit and inversion/caching
      edge behavior with renderer-neutral tests.
- [x] Audit and public-surface rows for `axis.py`, `ticker.py`, `scale.py`, and
      `transforms.py` are closed to precise statuses.

---

# Phase 14: Collections, Scalar Mapping, Meshes, and Colorbars

**Goal:** close high-value collection/scalar-mapping/mesh/colorbar gaps that
affect visible output and migration.

Status: completed and compacted.

- [x] Mutable collection/scalar-mappable state updates propagate deterministically.
- [x] Mesh shading/shape rules and colorbar orientation/boundary/tick behavior
      are source-backed and fixture-covered.
- [x] Advanced norm/color-helper scope is implemented or explicitly classified.
- [x] Audit and public-surface rows for `collections.py`, `cm.py`, `colors.py`,
      `colorbar.py`, and `colorizer.py` are closed to precise statuses.

---

# Phase 15: Patches, Text, Annotation, Legend, and Offset Boxes (Core Closure)

**Goal:** complete the non-coordinate-boundary portion of 12.4 (former 12.4A-F)
with a compact status view.

Status: completed and compacted.

- [x] Patch/hatch coverage includes box styles, arrow/connection styles,
      hatch-density semantics, and focused fixture coverage.
- [x] Text/font, annotation/offset-box, and legend handler/layout breadth are
      implemented with renderer-neutral tests and focused matrix fixtures.
- [x] Audit and public-surface rows for `patches.py`, `hatch.py`, `text.py`,
      `font_manager.py`, `textpath.py`, `legend.py`, `legend_handler.py`, and
      `offsetbox.py` are closed to precise statuses.

---

# Phase 16: Display Coordinate and Backend Boundary Parity (Dedicated Former 12.4G)

**Goal:** remove y-orientation and renderer-boundary mismatches so signed
display-space geometry follows upstream Matplotlib semantics.

**Reference sources:** `third_party/matplotlib/lib/matplotlib/transforms.py`,
`patches.py`, `text.py`, `backend_bases.py`, `_backend_agg.cpp`, local
`core/artist.go`, `core/arrow_patch.go`, `core/text.go`, `backends/agg/`, and
when needed `../agg_go` vs `../agg_2.4`.

- [ ] Define and document the coordinate contract between core display-space
      geometry and renderer backends.
- [ ] Add renderer-neutral regressions for signed display-space geometry:
      `ConnectionStyle("arc3", rad=...)`, arrow shrink/clip on curves,
      arrow-head normals, rotated-text bbox orientation, annotation arrow start
      after bbox clipping.
- [ ] Audit and remove non-source-backed y-sign compensations in core,
      transforms, and fixtures.
- [ ] Reconcile `text_annotation_matrix` bbox-arrow behavior against upstream
      `Annotation.update_positions` and
      `FancyArrowPatch._get_path_in_displaycoord`.
- [ ] Validate boundary fixes across `patch_style_matrix`,
      `annotation_composition`, and `transform_coordinates`.
- [ ] If residuals are AGG-port specific, fix `../agg_go` directly instead of
      compensating in this repository.
- [ ] Track before/after metrics; immediate target:
      `text_annotation_matrix` `RMSE < 10` against Matplotlib reference with
      1:1 example sources.

Execution track (kept from former 12.4G):

Status: [x] done · [~] in progress · [ ] todo.

- [x] G1 Contract & core pivot.
- [x] G2 AGG backend owns device flip.
- [~] G3 Core positioning/text helpers y-up conversion.
- [ ] G4 AGG parity validation.
- [ ] G5 Example 1:1 port sweep.
- [ ] G6 Vector/other backend inversion ownership.
- [ ] G7 Full-suite regen and revalidation.
- [ ] G8 Renderer-neutral signed-geometry regression set.

Exit criteria:

- [ ] Signed display-space paths/annotations/text bboxes/arrow geometry are
      source-backed under the documented coordinate contract.
- [ ] No `text_annotation_matrix`-specific sign hacks exist in example or core.
- [ ] `TestMatplotlibRef/text_annotation_matrix` reports `RMSE < 10` without
      regressions in related fixtures.
- [ ] Remaining mismatch is classified with evidence as core, renderer
      boundary, AGG-port, or upstream limitation.

---

# Phase 17: Images, Pyplot, Backends, Widgets, and Animation

**Goal:** close remaining image/stateful-wrapper/backend/widget/animation
parity decisions and keep unsupported scope explicit.

Status: largely completed and compacted; final closure still open.

- [x] Image class/resampling scope, interpolation registry coverage,
      transformed-image behavior, and typed omissions/errors are implemented.
- [x] Pyplot/stateful wrapper coverage is broad and explicitly classified,
      with object-oriented APIs remaining source of truth.
- [x] Backend lifecycle/tool semantics, widget interaction scope, and animation
      playback behavior are implemented and test-backed.
- [x] Public-surface parity rows for `image.py`, `pyplot.py`,
      `_pylab_helpers.py`, `backend_bases.py`, `backend_tools.py`,
      `widgets.py`, and `animation.py` are split into precise statuses.

Exit criteria (remaining open):

- [ ] Every `GapDecisionImplement` row in `FoundationAPIGapAudit` is either
      implemented with catalog coverage or deliberately reclassified with
      rationale.
- [ ] Every `partial` core feature row in `FeatureCoverageMatrix` is moved to
      `implemented`, `intentional-omission`, or a smaller precise partial row.
- [ ] `go test ./test/...` parity failures from newly changed behavior are
      resolved by source-backed core fixes or fixture updates.

---

# Phase 18: User-Facing Example Breadth
- Remaining renderer nudges must carry owner, rationale, and removal path before acceptance.

---

# Phase 9: Matplotlib API Coverage Audit

**Goal:** close the _existence_ gap, not just the _quality_ gap. Phase 8's RMSE
audit can only flag a case that already has a catalog example, so a feature
that was never implemented produces no failing parity test and stays invisible.
This phase enumerates Matplotlib's public catalogs (colormaps, markers, named
colors, interpolation modes, arrow/connection styles, patch classes, hatch
styles) and, for each missing entry, either implements it or records it as a
documented intentional divergence — then adds a catalog case so it becomes
visible to `TestReferenceCompare`.

Logically this precedes Phase 8: a feature must exist before its render
quality can be RMSE-audited.

**Reference sources:** `third_party/matplotlib/lib/matplotlib/_cm.py`,
`_cm_listed.py`, `markers.py`, `_color_data.py`, `colors.py`, `patches.py`,
`hatch.py`, `image.py`; current `color/`, `core/patch*.go`, marker drawing in
`core/`.

### 9.1 Coverage inventory

- [ ] Generate a machine-checked diff of the Matplotlib public surface against
      the Go surface for each enumerable catalog below, committed under
      `internal/parityutil` or `test/` so it can run in CI.
- [ ] For every gap, record a decision: implement, or document as an
      intentional divergence with a reason.
- [ ] Add an `AGENTS.md` / `PLAN.md` note that any new enumerable catalog must
      ship with a coverage check so this gap cannot silently reopen.

### 9.2 Colormaps

**Gap:** ~8 of Matplotlib's ~87 base colormaps are registered (`viridis`,
`plasma`, `inferno`, `magma`, `cividis`, `gray`, `binary`, `blues`).

- [x] Add the remaining perceptually-uniform sequential, sequential,
      diverging (`RdBu`, `coolwarm`, `seismic`, `bwr`, `PiYG`, …), cyclic
      (`twilight`, `twilight_shifted`, `hsv`), qualitative (`tab10`, `tab20`,
      `tab20b/c`, `Set1-3`, `Pastel1/2`, `Paired`, `Accent`, `Dark2`), and
      miscellaneous (`jet`, `rainbow`, `terrain`, `gist_*`, `ocean`, `cubehelix`,
      `nipy_spectral`, …) colormaps.
- [x] Support reversed `_r` variants and `Colormap.resampled` / `reversed`.
- [x] Confirm `ListedColormap` and `LinearSegmentedColormap` construction and
      registration parity.
- [x] Add catalog cases exercising a diverging, a qualitative, and a cyclic
      colormap end to end.

### 9.3 Markers

**Gap:** ~7 of Matplotlib's ~26 marker styles exist (circle, cross, diamond,
plus, square, triangle, path).

- [x] Add the missing built-in markers: `,` pixel, `.` point, `v^<>` triangle
      directions, `1234` tri markers, `8` octagon, `p` pentagon, `*` star,
      `hH` hexagons, `X` filled-x, `P` filled-plus, `d` thin diamond,
      `|_` vline/hline, and the caret markers (`TICKLEFT/RIGHT/UP/DOWN`,
      `CARETLEFT/...`).
- [ ] Support `MarkerStyle` with `fillstyle` (`full/left/right/bottom/top/none`).
      `MarkerStyle` exists and `full` / `none` are routed through scatter;
      half-fill (`left/right/bottom/top`) still needs alternate-path drawing
      rather than the current single-path collection model.
- [x] Support mathtext markers and `(numsides, style, angle)` tuple markers.
- [x] Add catalog cases exercising the full marker grid.

### 9.4 Named colors

**Gap:** no CSS4/X11, xkcd, or tableau (`tab:`) color databases found.

- [x] Add the CSS4/X11 named-color table and base single-letter colors.
- [x] Add the `tab:` tableau cycle names and the `xkcd:` color survey table.
- [x] Provide a `to_rgba`-equivalent that resolves names, hex, shorthand,
      grayscale strings, and `(r,g,b[,a])` tuples uniformly.

### 9.5 Image interpolation modes

**Gap:** ~5 of ~16 interpolation modes exist (nearest, bilinear, bicubic,
hamming, hanning).

- [ ] Add `lanczos`, `spline16`, `spline36`, `kaiser`, `quadric`, `catrom`,
      `gaussian`, `bessel`, `mitchell`, `sinc`, `blackman`, `hermite`, and the
      `antialiased`/`auto` resampling policy.
- [ ] Route the modes through AGG and the shared image resampler with documented
      fallbacks where a backend cannot match a filter.

### 9.6 Arrow, connection, and patch classes

**Gap:** only `FancyArrow` exists; the `ArrowStyle` / `ConnectionStyle`
registries, `FancyArrowPatch`, and `ConnectionPatch` are absent. Patch classes
`Shadow`, `RegularPolygon`, `CirclePolygon`, `Arc`, `Annulus`, and `StepPatch`
are missing.

- [x] Add the `ArrowStyle` registry (`-`, `->`, `-[`, `]-[`, `|-|`, `<-`,
      `<->`, `fancy`, `simple`, `wedge`, …) and the `ConnectionStyle` registry
      (`arc3`, `arc`, `angle`, `angle3`, `bar`).
- [x] Add `FancyArrowPatch` and `ConnectionPatch` wired into `Annotate`.
- [x] Add the missing patch classes (`Shadow`, `RegularPolygon`,
      `CirclePolygon`, `Arc`, `Annulus`, `StepPatch`).
- [ ] Audit `FancyBboxPatch` box-style coverage against `BoxStyle._style_list`.

### 9.7 Hatch styles and miscellaneous

- [ ] Verify the full hatch character set (`/ \ | - + x o O . *`) and
      repeat-density semantics against `hatch.py`.
- [ ] Add `set_sketch_params` / `pyplot.xkcd()` sketch-style support, or
      document the omission.
- [ ] Decide on `figimage`: implement or document as an intentional omission.
      (`pcolorfast` now maps to the typed `PColorFast` / `PColorMesh` path.)
- [ ] Audit `rcParams` keys against upstream and record which keys are
      unsupported.

**Exit criteria:**

- [ ] The committed coverage check reports every enumerable catalog as either
      implemented or holding a documented intentional-divergence entry.
- [ ] Newly implemented features each have at least one catalog case so they
      surface in `TestGolden` / `TestReferenceCompare`.
- [ ] A user calling `cmap="coolwarm"`, `marker="*"`, `color="xkcd:teal"`, or
      `interpolation="lanczos"` gets correct output or a clear, documented
      error — never a silent wrong default.

---

# Phase 10: Feature and Demo Coverage Audit

✅ **Completed.** The coarse feature/demo audit now makes missing Matplotlib
parity coverage visible and testable before v1.0.

Completed scope:

- `FeatureCoverageMatrix` classifies foundational Matplotlib areas by
  implementation, parity fixture, showcase, browser-demo, and breadth status.
- `FoundationAPIGapAudit` records decisions for thin or missing fundamental
  APIs across artists, axes, ticks, transforms, lines, collections, patches,
  text, images, colorbars, colors, pyplot, and backends.
- `DemoBreadthGaps` tracks fixture-heavy or thin user-facing examples and links
  high-priority gaps to target feature families.
- `BrowserDemoCoverageRows` reconciles inactive web reference modules and
  CLI-only showcases against active, planned, or reference-only browser demo
  status.
- Reference-consistency tests enforce Go/Python parity source pairs and visible
  Matplotlib reference modules for catalog cases.
- `docs/phase-9a-coverage-audit.md` explains the audit inventories and
  clarifies that implementation follow-up continues in Phases 9B-9E.

---

# Phase 11: Exhaustive Public Surface Parity Map

**Goal:** turn the coarse Phase 9A audit into the detailed answer originally
needed: for each relevant upstream Matplotlib public API, enumerable registry,
and gallery family, state whether the Go port has a direct equivalent, an
idiomatic equivalent, an intentional omission, or no implementation yet.

**Reference sources:** `third_party/matplotlib/lib/matplotlib/`,
`third_party/matplotlib/galleries/examples/`, `internal/examplecatalog/`,
`core/`, `transform/`, `render/`, `color/`, `style/`, `pyplot/`, `canvas/`,
`backends/`, `examples/`, and `test/parity/`.

### 11.1 Public API Inventory Generator

- [x] Add a small internal tool that scans upstream Python modules for public
      classes, functions, constants, and registries in the modules tracked by
      `FoundationAPIGapAudit`: `artist.py`, `axis.py`, `ticker.py`,
      `scale.py`, `transforms.py`, `lines.py`, `markers.py`,
      `collections.py`, `patches.py`, `text.py`, `legend.py`,
      `offsetbox.py`, `image.py`, `colorbar.py`, `cm.py`, `colors.py`,
      `pyplot.py`, `backend_bases.py`, `backend_tools.py`, `widgets.py`, and
      `animation.py`.
- [x] Store the normalized inventory under `internal/examplecatalog` or
      `test/testdata/parity_surface/` so CI can diff upstream-visible
      additions.
- [x] Treat enumerable registries specially: markers, line styles, draw styles,
      cap/join styles, colormaps, named colors, norms, locators, formatters,
      scales, patch classes, box styles, arrow styles, connection styles,
      hatch patterns, projections, backends, toolbar tools, widgets, and image
      interpolation modes.

Current slice landed:

- `internal/examplecatalog/extract_public_surface.py` uses Python `ast` to
  extract a stable upstream inventory for the Phase 9B tracked modules.
- `test/testdata/parity_surface/upstream_public_surface.json` stores the
  committed inventory: 21 modules and 591 public-surface rows covering public
  classes, functions, constants, and selected registries such as markers,
  line styles, patch styles, scales, toolbar tools, and image interpolation
  modes.
- Catalog tests now verify landmark upstream rows and fail when the committed
  artifact differs from the extractor output.
- `internal/examplecatalog.PublicSurfaceParityRows` seeds the Phase 9B.2
  mapping with first-pass classifications for landmark rows including
  `Artist`, `Line2D`, the `*` marker, `lanczos` interpolation, `pyplot.plot`,
  `Button`, and `FuncAnimation`.

Implementation notes:

- Use Python `ast` for the upstream scan rather than regex so class/function
  extraction is stable.
- Keep private names out by default, but allow explicit include lists for
  upstream registries whose public API is stored in underscored module data
  such as `_cm.py`, `_cm_listed.py`, and `_color_data.py`.
- Start with a generated JSON artifact and a Go test that verifies every
  public upstream row has a local classification.

### 11.2 Go Equivalent Mapping

- [x] Add a `PublicSurfaceParityRows` inventory that maps each upstream row to
      one of:
      `direct-equivalent`, `idiomatic-equivalent`, `partial`, `not-started`,
      `intentional-omission`.
- [x] For every row, record the local Go package/file, catalog IDs, demo IDs,
      and implementation note.
- [x] Fail tests when a new upstream row appears without a classification.

Current slice landed:

- `internal/examplecatalog.PublicSurfaceParityRowsForSurface` now classifies
  the committed upstream inventory row-by-row, with exact overrides for
  landmark APIs and conservative module/family rules for the remaining public
  classes, functions, constants, and registries.
- The committed 591-row upstream surface is covered by tests that require one
  classification per row, stable status values, an existing `FeatureCoverage`
  row, at least one local Go file reference, and valid catalog/showcase
  references when present.
- Exact overrides now capture high-signal parity answers such as `Artist`,
  `Line2D`, `pyplot.plot`, `Button`, `FuncAnimation`, marker `*`, and AGG's
  direct `lanczos` interpolation support.

Implementation notes:

- Seed the mapping from `FeatureCoverageMatrix` and `FoundationAPIGapAudit`,
  then refine it row-by-row.
- Keep the first pass conservative: if the Go port supports a concept but not
  the full upstream behavior, mark it `partial`.
- Prefer documenting an intentional omission over leaving behavior ambiguous.

### 11.3 Human Parity Status Report

- [ ] Generate or maintain `docs/matplotlib-parity-status.md` from the
      machine-readable inventories.
- [ ] Include one table per upstream feature family with columns:
      upstream API / registry item, Go status, local API, parity fixture,
      user-facing example, browser demo, and remaining work.
- [ ] Add a summary table that answers directly:
      "ported", "partially ported", "not ported", "intentionally omitted",
      "has parity fixture", "has user example", and "has browser demo".

Implementation notes:

- Do not hand-write status that duplicates data without a check. Either
  generate the report or add tests that ensure the doc references every
  inventory row.
- Link every "missing" or "partial" row to a Phase 9C, 9D, or 9E task.

**Exit criteria:**

- [ ] A developer can open `docs/matplotlib-parity-status.md` and see a
      detailed answer to whether each tracked upstream feature is ported and
      whether it has examples.
- [ ] CI fails when an upstream public row or enumerable registry item is
      tracked but unclassified.
- [ ] Every `partial`, `not-started`, and `intentional-omission` row has a
      rationale and a next action.

---

# Phase 12: Foundational API Parity Closure

**Goal:** implement or explicitly omit the missing fundamental APIs surfaced by
`FoundationAPIGapAudit`, prioritizing behavior that affects visual parity,
Matplotlib migration examples, or broad public use.

### 12.1 Artist, Line2D, and Marker Semantics

**Goal:** close the foundational `artist.py`, `lines.py`, and `markers.py`
gaps that affect static rendering, legends, migration examples, and parity
fixtures. Keep the Go API explicit and typed, but match Matplotlib behavior
where the output is visible.

#### 12.1A Landed Baseline

- [x] Common artists that embed `ArtistRasterization` now get shared
      Matplotlib-style metadata for visibility, artist-level alpha, in-layout,
      and stale state. The zero value remains visible, alpha=1, and
      in-layout=true.
- [x] Artist traversal skips invisible artists for both normal and overlay
      draws.
- [x] `Line2D` and collection-derived artists now combine artist-level alpha
      with their existing stroke/fill/collection alpha behavior, preserving the
      existing zero-value "alpha omitted" semantics.
- [x] `Line2D` now supports optional data markers with marker style/path,
      marker size, marker face color, marker edge color, marker edge width,
      every-N `MarkEvery`, and combined line+marker legend samples.

#### 12.1B Shared Artist Metadata Remainder

- [x] Add shared label accessors for artists that currently store labels in
      concrete fields (`Line2D.Label`, `Scatter2D.Label`, `Collection.Label`,
      `Patch.Label`, etc.) so legend and inspection code can use one path.
- [x] Add per-artist clip metadata:
      clip-on flag, clip rectangle, clip path, and clip transform where needed
      for static rendering parity.
- [x] Wire per-artist clipping into traversal without breaking existing axes
      clip behavior. Required proof: an artist can render clipped differently
      from the axes default, and overlay artists still respect their intended
      unclipped behavior.
- [x] Add custom per-artist transform support for the common static cases:
      data transform override, axes/figure coordinate transform, and explicit
      display transform for path-like artists.
- [x] Decide whether stale state needs parent propagation now. If yes, wire it
      through `Axes`/`Figure`; if no, document the intentional v1.0 scope as
      local artist state only.
- [x] Add focused catalog/parity coverage for visibility, alpha, custom clip,
      and custom transform behavior.

Current slice landed:

- `core.ArtistLabel`, `core.SetArtistLabel`, `ArtistLabeler`, and
  `ArtistLabelSetter` now provide a shared inspection/update path for common
  labeled artists.
- Legend collection now uses the shared label path and follows Matplotlib's
  convention of skipping empty labels and labels beginning with `_`.
- `ArtistRasterization` now carries explicit clip metadata: clip-on, clip rect,
  clip path, and optional affine clip-path transform with clone-returning path
  access.
- Normal and overlay draw traversal now nests explicit artist clips around the
  artist draw call, while overlay artists without explicit clip metadata remain
  unclipped by this artist-level path.
- `ArtistRasterization` now carries optional transform metadata:
  coordinate-system overrides via `SetTransformCoords` and explicit
  coordinate-to-display transforms via `SetTransform`.
- Lines, scatter/path collections, line and patch collections, patch artists,
  wedges, shadows, and fancy-arrow patches now resolve their display paths
  through the shared artist transform helper.
- Stale state is intentionally scoped to local artist metadata for v1.0.
  Parent `Axes`/`Figure` propagation is deferred until those types own explicit
  stale lifecycle semantics; see
  `docs/adr/0002-artist-stale-state-scope.md`.
- Added the fixture-only `artist_metadata` parity case with Go golden and
  Matplotlib reference output covering hidden artists, artist-level alpha,
  explicit clip boxes, and artist transform coordinate overrides.

#### 12.1C Line2D Data and Stroke Semantics

- [x] Add typed data getters/setters for `Line2D`: clone-returning `Data`,
      `SetData`, `SetXData`, and `SetYData`, with stale invalidation.
- [x] Define and implement NaN / Inf segment splitting to match Matplotlib's
      line-break behavior instead of drawing through invalid points.
- [x] Add `gapcolor` support for dashed lines where Matplotlib paints dash gaps
      with an alternate color.
- [x] Expand `MarkEvery` beyond every-N integers:
      start/step tuple, explicit index list, and slice-like range where that
      maps cleanly to Go.
- [x] Add catalog/parity cases for data mutation, invalid-point line breaks,
      dashed `gapcolor`, and at least two nontrivial `markevery` forms.

Current slice landed:

- `Line2D` now exposes clone-returning `Data` plus `SetData`, `SetXData`, and
  `SetYData`; setters replace internal point storage and mark the artist stale.
- Line drawing and bounds now ignore NaN / Inf coordinates and split display
  paths at invalid points instead of connecting across them.
- Dashed `Line2D` artists can paint gap segments with `SetGapColor`; core
  extracts the dash-gap display path so the behavior is not dependent on
  backend dash-offset support.
- `MarkEverySpec` and helpers (`EveryNMarkers`, `StartStepMarkers`,
  `IndexedMarkers`, `SliceMarkers`) cover richer Matplotlib-style marker
  subsampling while preserving the existing integer `MarkEvery` field.
- Added fixture-only `line2d_semantics` parity coverage, with Go golden and
  Matplotlib reference output for data mutation, invalid-point line breaks,
  dashed gapcolor, and two nontrivial markevery forms.

#### 12.1D Line2D Marker Completion

- [x] Verify current `Line2D` marker size conversion, marker face/edge fallback,
      marker edge width, marker alpha, and legend samples against upstream
      `lines.py` and `markers.py`.
- [x] Support marker face color `"none"` semantics through an explicit Go option
      or sentinel that does not rely on ambiguous zero-alpha fallback.
- [x] Support marker edge color `"auto"` / face-color fallback semantics where
      Matplotlib does so for common markers.
- [x] Ensure line-only markers (`+`, `x`, ticks, carets, etc.) use stroke-only
      marker rendering in both plot output and legends.
- [x] Add catalog/parity cases for Line2D markers in plots and legends,
      including filled, unfilled, line-only, custom path, tuple marker, and
      mathtext marker variants.

Current slice landed:

- `Line2D` marker edge widths now resolve from points to display pixels, matching
  upstream `set_markeredgewidth`; the existing direct field remains the public
  typed knob.
- `MarkerColorSpec` plus `ExplicitMarkerColor`, `AutoMarkerColor`, and
  `NoMarkerColor` provide explicit Go sentinels for Matplotlib-style marker
  face / edge color `"auto"` and `"none"` behavior.
- `Line2D` marker edge `"auto"` follows upstream line-RGB behavior and inherits
  face alpha when the face is filled; `MarkerFillNone` and `NoMarkerColor`
  produce stroke-only markers without relying on zero-alpha ambiguity.
- Line-only markers now draw stroke-only in both plot output and legend samples.
- Added fixture-only `line2d_markers` parity coverage with plot and legend
  samples for filled, unfilled, line-only, custom path, tuple, mathtext, and
  half-filled markers.

#### 12.1E Half-Filled Marker Paths

- [x] Replace the current single filled marker path behavior for
      `MarkerFillLeft`, `MarkerFillRight`, `MarkerFillTop`, and
      `MarkerFillBottom` with split marker drawing: primary half uses
      markerfacecolor, alternate half uses markerfacecoloralt / transparent
      fallback, and edge drawing remains whole-marker.
- [x] Implement split paths for circle, square, diamond, triangle, and polygon
      markers first; explicitly document or implement behavior for line-only,
      tuple, custom path, and mathtext markers.
- [x] Add renderer-neutral tests that inspect the two fill paths and one edge
      path for a half-filled marker.
- [x] Add catalog/parity fixture coverage for all four half-fill directions.

Current slice landed:

- The shared marker split helper now produces primary / alternate half-fill
  paths for circles and polygonal markers, including square, diamond, triangle,
  regular polygon, tuple polygon/star, and polygonal custom paths. Line-only
  markers stay stroke-only; non-polygon custom and mathtext paths remain whole
  paths until a clipping-path marker fill implementation is warranted.
- `Line2D` draws half-filled markers as primary fill, alternate fill, and a
  separate whole-marker edge pass. `Scatter2D` uses the same split paths with a
  transparent alternate fallback and whole-marker edge pass.
- Renderer-neutral tests cover the Line2D two-fill / one-edge contract and the
  Scatter2D split-fill / whole-edge fallback; the `line2d_markers` fixture
  covers all four half-fill directions against Matplotlib.

#### 12.1F Exit Criteria

- [x] `FoundationAPIGapAudit` row `artist-clipping-transform` is either closed
      or split into smaller remaining rows with exact implementation scope.
- [x] `FoundationAPIGapAudit` row `line2d-marker-data-semantics` is either
      closed or split into smaller remaining rows with exact implementation
      scope.
- [x] Public-surface parity rows for `Artist`, `Line2D`, `MarkerStyle`, marker
      registries, and fillstyle registries are updated from `partial` to
      precise final statuses or linked to the remaining 9C.1 subtask.
- [x] 9C.1 has at least one catalog/parity case each for visibility, alpha,
      clipping, Line2D markers in legends, invalid line data, `markevery`,
      `gapcolor`, and half-filled markers.
- [x] `go test ./core -count=1` and the relevant catalog cases under
      `go test ./test/ -run ...` pass after the behavior changes.

Implementation notes:

- Compare against `third_party/matplotlib/lib/matplotlib/artist.py`,
  `lines.py`, and `markers.py` before changing behavior.
- Add focused catalog cases for marker fill styles, Line2D markers in legends,
  clipping, alpha, and visibility.
- Prefer adding shared Go option structs/mixins over copying Python's dynamic
  setter model wholesale.

### 12.2 Axis, Ticker, Formatter, Scale, and Transform Breadth

**Goal:** close the user-visible axis, ticker, formatter, scale, and transform
gaps surfaced by Phases 10/11 without cloning Matplotlib's full dynamic class
hierarchy. Prefer small Go option structs and focused helpers, with one
catalog/parity fixture per behavior family.

#### 12.2A Landed Baseline

- [x] `Axis` owns major/minor locators, formatters, tick labels, mirrored
      top/right axes, tick direction, tick line styling, label styling, and
      extra tick-label levels.
- [x] Common locator coverage exists for auto/max-N, linear, fixed, null,
      multiple-with-offset, log major/minor, auto minor, date/day, and
      categorical axes.
- [x] Common formatter coverage exists for scalar, fixed, null, function,
      printf-style, str-method subset, engineering, percent, log powers, date,
      auto-date, and category labels.
- [x] Scale construction exists for linear, log, symlog, asinh, logit,
      function, custom registry entries, non-positive handling, and shared
      primary/secondary axes.
- [x] Transform infrastructure already covers affine, separable, blended,
      chained, offset, axes/figure/data coordinate transforms, display-rect
      transforms, and opt-in transform-node invalidation/caching.

#### 12.2B Locator Catalog Closure

- [x] Audit upstream `ticker.py` locators row by row against the current Go
      surface, and split each missing locator into `implement`,
      `idiomatic-equivalent`, or `intentional-omission` in the Phase 11 public
      surface notes.
- [x] Tighten `AutoLocator` / `MaxNLocator` edge semantics against upstream:
      `nbins`, `steps`, integer-only mode, pruning, symmetric behavior,
      degenerate ranges, negative ranges, and very small / very large spans.
- [x] Compare `MaxNLocator.view_limits` / `nonsingular` behavior against
      upstream for degenerate, negative, tiny-span, and large-offset domains;
      add focused unit cases for any remaining mismatches.
- [x] Add a `locator_maxn_edge_labels` catalog/parity fixture showing
      degenerate-range expansion, pruning, and large-offset tick labels.
- [x] Add `MaxNLocator` option coverage for custom `steps`, integer-only
      relaxation via `MinTicks`, symmetric ranges, pruning, and degenerate
      range expansion.
- [x] Preserve distinct `MaxNLocator` ticks for tiny spans and large-offset
      spans by deduping generated ticks relative to the selected step.
- [x] Add `IndexLocator`, `LinearLocator` exact-count / preset behavior, and
      `FixedLocator` subsampling behavior.
- [x] Audit `OldAutoLocator`: it is not present in the current vendored
      upstream `ticker.py`; no Go compatibility surface is required unless an
      older Matplotlib target is explicitly added.
- [x] Complete log-family locator behavior: dense minor ticks, `subs="auto"` /
      `subs="all"` equivalents, minor threshold behavior, base changes, and
      safe behavior for non-positive or inverted domains.
- [x] Compare remaining `LogLocator` stride/offset decisions against upstream
      for dense ranges and non-decimal bases; add unit cases for the exact
      ranges that still differ.
- [x] Add a `locator_log_minor_threshold_labels` catalog/parity fixture for
      visible major/minor log-family locator behavior.
- [x] Add `LogLocator` `SubsMode` support for `auto` / `all` equivalents and
      dense automatic minor-tick suppression.
- [x] Add `LogLocator` numticks-driven major thinning and explicit minor
      suppression when dense ranges require a major stride above one.
- [x] Add a scale-specific `SymLogLocator` and install it as the default
      major/minor locator for `symlog` axes.
- [x] Add scale-specific `AsinhLocator` and `LogitLocator` implementations and
      install them as default major/minor locators for `asinh` and `logit`
      axes.
- [x] Add renderer-neutral unit tests for each new locator and a catalog case
      named for each visible locator family rather than one combined fixture.
- [x] Add `locator_linear_labels` covering `LinearLocator`, `MultipleLocator`,
      and default linear tick labels.
- [x] Add `locator_fixed_index_labels` covering `FixedLocator` subsampling and
      `IndexLocator` base/offset placement.
- [x] Close the locator catalog row only after the named linear, fixed/index,
      log, symlog, asinh, and logit fixture coverage is present and listed in
      `PublicSurfaceParityRows`.

#### 12.2C Formatter Catalog Closure

- [x] Audit upstream `ticker.py` formatters row by row and update public-surface
      parity notes for direct equivalents, Go-style equivalents, and deliberate
      omissions.
- [x] Expand `ScalarFormatter` parity: offset text decision, scientific limits,
      power limits, math-text mode, fixed-minus behavior, locale/no-locale
      decision, and step-aware precision.
- [x] Decide and document ScalarFormatter offset text policy: implement
      axis-level offset text, or record a deliberate omission with a migration
      note and parity-surface update.
- [x] Add locale/no-locale ScalarFormatter coverage matching the chosen Go
      policy; keep labels deterministic under the default test locale.
- [x] Add `ScalarFormatter` option coverage for inclusive power limits,
      MathText-style scientific labels, scientific suppression, and
      step-aware precision.
- [x] Add `formatter_scalar_scientific_labels` catalog/parity fixture for
      visible scalar scientific MathText labels.
- [x] Split log formatting into explicit Go formatter types for the existing
      simple `LogFormatter`, exponent-only labels, MathText labels, and
      scientific-notation MathText labels.
- [x] Add sparse minor-label behavior for log-family formatters where it is
      needed for visible parity.
- [x] Add `formatter_log_mathtext_labels` catalog/parity fixture for visible
      log MathText tick labels.
- [x] Tighten `EngFormatter` behavior for separator defaults, unicode micro,
      places, sign handling, unit spacing, and extreme prefixes.
- [x] Add `EngFormatter` coverage for unicode micro output, minus-sign fixing,
      engineering-prefix rollover after rounding, and extreme SI prefixes.
- [x] Add `EngFormatter` MathText-style number wrapping, `FormatEng` aliasing,
      and zero-without-suffix handling.
- [x] Align `EngFormatter` zero-value defaults with upstream separator,
      auto-place formatting, unicode micro, and explicit zero-place escape
      hatches.
- [x] Add `formatter_engineering_labels` catalog/parity fixture for visible
      engineering tick labels, including zero-with-unit spacing.
- [x] Tighten `PercentFormatter` behavior for xmax defaults, decimal auto
      selection, symbol escaping/no-escaping decision, and negative values.
- [x] Add `PercentFormatter` auto-decimal support for configured display ranges
      and fixed-minus handling for negative percentages.
- [x] Add `PercentFormatter` no-symbol output and TeX / raw-LaTeX symbol
      handling.
- [x] Align `PercentFormatter` zero-value defaults with upstream `xmax=100`
      and auto-decimal behavior while preserving explicit zero decimals.
- [x] Add `formatter_percent_labels` catalog/parity fixture for visible
      percent tick labels.
- [x] Audit `IndexFormatter`: it is not present in the current vendored
      upstream `ticker.py`; no Go compatibility surface is required unless an
      older Matplotlib target is explicitly added.
- [x] Add `LogitFormatter` and install it as the default logit-axis formatter
      for major ticks, with minor ticks suppressed by default.
- [x] Add or explicitly omit `FixedFormatter` mismatch warnings,
      `NullFormatter`, `FuncFormatter`, `FormatStrFormatter`, and
      `StrMethodFormatter` edge behavior in the audit rows.
- [x] Add `formatter_fixed_null_labels` catalog/parity fixture for visible
      fixed tick labels and null-label suppression.
- [x] Add one catalog/parity case each for scalar offset/scientific labels, log
      math-text labels, engineering labels, percent labels, and index/fixed/null
      labels.
- [x] Close the formatter catalog row after scalar offset policy is resolved;
      current fixture coverage already includes scientific, log MathText,
      engineering, percent, and fixed/null labels.

#### 12.2D Date, Category, and Unit Tick Breadth

- [x] Audit upstream `dates.py` locator/formatter families separately from
      generic `ticker.py`, keeping the source-of-truth list in the Phase 12
      coverage notes unless those modules are later added to the Phase 11
      public-surface inventory.
- [x] Expand date locators in small slices: year/month, weekday, day-of-month,
      and hour/minute/second locators.
- [x] Add microsecond locator support if practical, and finish interval
      selection gaps for compact ranges.
- [x] Add `AutoDateLocator`-style interval selection and `ConciseDateFormatter`
      style multi-level label suppression where it affects visible parity.
- [x] Tighten timezone handling and Matplotlib date-number conversion so date
      ticks remain stable across UTC and non-UTC locations.
- [x] Preserve explicit user locators/formatters when unit converters refresh
      axis info; add regression tests for date and categorical axes.
- [x] Add `date_concise_intraday_labels` catalog/parity fixture for intraday
      `HourLocator` ticks with concise date formatting.
- [x] Add `date_month_year_labels` catalog/parity fixture for monthly/yearly
      date locator and formatter behavior.
- [x] Existing `units_dates` and `units_categories` catalog cases cover daily
      date labels and categorical tick labels for this phase.
- [x] Add catalog/parity cases for daily, monthly/yearly, intraday, concise
      date, and categorical tick labels.

#### 12.2E Scale-Specific Axis Defaults

- [x] Compare `scale.py` default locator/formatter setup for linear, log,
      symlog, asinh, logit, function, and functionlog against
      `configureScaleAxis`.
- [x] Add named `functionlog` scale support through the Go scale registry.
- [x] Ensure `SetXScale` / `SetYScale`, semilog helpers, colorbar axes, twin
      axes, secondary axes, and shared axes all install the same scale-specific
      locator/formatter defaults.
- [x] Wire `SetXScale` / `SetYScale` symlog defaults to `SymLogLocator` for
      major and minor ticks.
- [x] Wire `SetXScale` / `SetYScale` asinh and logit defaults to
      `AsinhLocator` / `LogitLocator` for major and minor ticks.
- [x] Wire `SetXScale` / `SetYScale` functionlog defaults to the log
      locator/formatter path.
- [x] Verify non-positive handling for log-like scales: clip, mask/drop,
      autoscale interaction, and readable error behavior for invalid domains.
- [x] Add catalog/parity cases for symlog ticks, logit ticks, asinh ticks, and
      function/functionlog scale defaults.

#### 12.2F Tick Styling and Axis Control Surface

- [x] Map Matplotlib `tick_params` behavior to Go axis/tick option structs:
      major/minor/both selection, axis selection, length, width, color, pad,
      label size, label color, rotation, direction, and reset behavior.
- [x] Add side visibility controls for tick marks and labels:
      top/bottom/left/right and labeltop/labelbottom/labelleft/labelright,
      including mirrored and secondary axes.
- [x] Add per-major/minor grid styling where upstream tick params affect grid
      lines: color, alpha, width, style, and visibility.
- [x] Keep ticks axis-owned for v1.0, but document the explicit non-goal of a
      Python-style `Tick` artist clone unless a migration example requires it.
- [x] Add unit tests for option propagation and a catalog/parity case covering
      major/minor styling, side visibility, grid styling, and rotated labels.

#### 12.2G Transform, BBox, and Path Helper Closure

- [x] Audit upstream `transforms.py`, `path.py`, and `bezier.py` rows against
      Go transform/geometry helpers and classify missing rows as implemented,
      Go-style equivalent, or intentional omission.
- [x] Add frozen transform helpers where callers need immutable snapshots of
      dynamic axes/figure/data transforms for annotations, clipping, images, or
      layout.
- [x] Add transformed-path helpers for path plus transform caching, invalidation
      hooks, affine/non-affine split decisions, and clone-safe access.
- [x] Audit affine/non-affine split behavior against the current Go transform
      model; either add the split API or document why all current transforms
      can use the Go-style cached `TransformedPath`.
- [x] Add a Go-style `TransformedPath` helper with clone-safe source/output
      paths and transform-node invalidation-backed cache refresh.
- [x] Expand BBox/rect helpers needed by layout and annotation parity:
      union, intersection, expanded/padded, anchored, transformed, inverse
      transformed, empty/null handling, and point containment.
- [x] Add shared `geom.Rect` anchored child-rectangle helper with compass and
      location-string anchors.
- [x] Add shared `geom.Rect` null sentinel and accumulation helpers for
      BBox-style extent building.
- [x] Add path/bezier helpers only when visible behavior needs them:
      interpolation, clipping against BBoxes, extents under transforms,
      simplification decisions, and curve splitting.
- [x] Add or explicitly omit path simplification thresholds based on a visible
      rendering case; do not add a generic simplifier without a parity target.
- [x] Add curve-splitting helpers only if a clipped curved-path fixture exposes
      a mismatch after the current flattening/interpolation helpers are used.
- [x] Add shared `geom.Path` clone, affine transform, bounds, and transformed
      bounds helpers for visible path extent/layout use cases.
- [x] Add shared `geom.Path` interpolation helper for line, quadratic, and
      cubic segments.
- [x] Add shared `geom.Path` clipping helper against `geom.Rect` for flattened
      path segments.
- [x] Add tests that exercise transform invalidation propagation and transformed
      path cache invalidation, plus catalog/parity cases for annotation
      coordinate modes, clipped transformed paths, and layout BBox behavior.
- [x] Add `transform_annotation_modes` or extend `transform_coordinates` to
      explicitly cover data, axes, figure, and offset annotation coordinates.
- [x] Add `path_clipped_transformed` covering a transformed path clipped by a
      BBox/axes rectangle.
- [x] Add `layout_bbox_helpers` covering anchored/padded/union BBox behavior in
      visible layout output.
- [x] Add renderer-neutral unit tests for frozen transform snapshots and
      transformed-path cache invalidation.

#### 12.2H Exit Criteria

- [x] `FoundationAPIGapAudit` rows `ticker-formatter-catalog`,
      `tick-artist-model`, and `transform-bbox-paths` are closed or split into
      precise remaining rows with exact scope.
- [x] Public-surface parity rows for the currently tracked `axis.py`,
      `ticker.py`, `scale.py`, and `transforms.py` modules are updated from
      broad `partial` notes to either implemented, precise partial, Go-style
      equivalent, or intentional omission status; `dates.py`, `category.py`,
      `path.py`, and `bezier.py` gaps are either added to that inventory or
      documented as supporting coverage notes.
- [x] `FeatureCoverageMatrix` rows `axis-ticker-scale` and `transforms` no
      longer rely on broad "catalog incomplete" notes; each remaining gap links
      to a specific 12.2 subtask.
- [x] 12.2 has catalog/parity coverage for locator, formatter, date/category,
      scale-default, tick styling, and transform/BBox helper behavior.
- [x] `go test ./core ./transform ./internal/examplecatalog -count=1` and the
      relevant `go test ./test/ -run ...` catalog cases pass.

Implementation notes:

- Compare against upstream `axis.py`, `ticker.py`, `scale.py`,
  `dates.py`, `category.py`, `transforms.py`, `path.py`, and `bezier.py`
  before changing behavior.
- Add one catalog case per formatter, locator, scale, or transform helper family
  instead of one enormous fixture.
- Keep the transform graph lean; add only helpers needed by rendering,
  annotations, layout, and clipping parity.
- Prefer Go option structs and typed helpers over Python-style dynamic setter
  surfaces. If a Python surface is intentionally not modeled, record the
  rationale in Phase 11 parity rows.

### 12.3 Collections, Scalar Mapping, Meshes, and Colorbars

**Goal:** close the highest-value scalar-mapping, collection, mesh, and
colorbar gaps surfaced by image/mesh parity without cloning Matplotlib's full
callback-heavy colorizer stack. Prefer explicit Go setters/options that keep
artist color state and colorbar state synchronized.

#### 12.3A Landed Baseline

- [x] Collection artist families exist for path, line, patch, polygon, quad
      mesh, and fill-between collections.
- [x] Scalar map metadata exists through `ScalarMappable` / `ScalarMapInfo` and
      is consumed by images, meshes, contours, vector fields, hexbin, and
      colorbars.
- [x] Common normalization coverage exists for linear, no-norm, log, symlog,
      power, two-slope, centered, and boundary norms.
- [x] Rectilinear mesh support covers flat, nearest, and Gouraud shading, masked
      values, bad/under/over colormap colors, and native mesh renderer batches.
- [x] Vertical colorbars exist for scalar mappables, including log,
      boundary-norm, nonlinear-function scale setup, extension patches, and
      constrained-layout synchronization.

#### 12.3B Collection Scalar-Mappable State

- [x] Audit upstream `collections.py`, `cm.py`, and `colorizer.py` for
      `Collection` / `ScalarMappable` array, cmap, norm, clim, and changed-state
      behavior; record Go-style equivalents and intentional omissions in Phase
      11 public-surface notes.
- [x] Add mutable scalar arrays to collection-style mappables so callers can
      update data values after artist creation without reconstructing the
      artist.
- [x] Add the first Go-style mutable scalar-mapping slice for path, patch, and
      polygon collections: `SetArray`, `GetArray`, `SetColormap`, `SetNorm`,
      `SetCLim`, and face-edge tracking.
- [x] Add Go-style setters for colormap, norm, and clim updates that refresh
      stored mapping metadata and recompute mapped face colors where the artist
      owns scalar-derived colors.
- [x] Preserve explicit face/edge colors when no scalar array is active, and
      define the precedence between scalar-derived face colors, explicit face
      colors, and edge colors matching Matplotlib where visible.
- [x] Support Matplotlib-like "edgecolors='face'" semantics for scalar-mapped
      collections so edge colors can track mapped face colors after scalar,
      norm, cmap, or clim updates.
- [x] Add collection offset-transform support needed by scatter/path
      collections without changing examples to compensate for transform gaps.
- [x] Ensure colorbars derived from a mutable mappable observe updated mapping
      state, either through explicit synchronization or a documented Go-style
      refresh path.
- [x] Add renderer-neutral unit tests for mutable array updates, cmap/norm/clim
      changes, face/edge precedence, offset transforms, and colorbar mapping
      synchronization.
- [x] Add a `collection_mutable_scalarmap` catalog/parity fixture that updates
      a collection's scalar data and colormap before rendering.

#### 12.3C Mesh, PColor, and Scalar Grid Behavior

- [x] Compare `pcolor`, `pcolormesh`, `QuadMesh`, and `PolyQuadMesh` behavior
      against upstream `collections.py` and `axes/_axes.py` for dimensionality,
      shading, masking, edge handling, and scalar array updates.
- [x] Tighten flat/nearest/Gouraud shape validation and edge inference only
      where current behavior diverges visibly from Matplotlib fixtures.
- [x] Make `QuadMesh` scalar updates recompute flat cell colors and Gouraud
      corner colors consistently with its stored shading mode.
- [x] Verify masked and non-finite mesh values continue to route through
      colormap bad/under/over colors after mutable mapping changes.
- [x] Add small catalog/parity cases for mutable pcolormesh scalar data and any
      shape/shading mismatch discovered during the audit.

#### 12.3D Colorbar Orientation, Ticks, and Layout Breadth

- [x] Audit upstream `colorbar.py` for orientation, location, anchor, shrink,
      aspect, fraction, pad, ticklocation, boundaries, values, spacing,
      drawedges, extend, extendfrac, extendrect, and multi-axes behavior.
- [x] Add horizontal colorbar placement with bottom/top tick and label defaults,
      parent-axes shrinking, extension geometry, and constrained-layout sync.
- [x] Add location and anchor support for left/right/top/bottom colorbars using
      Go option fields rather than Python-style overloaded kwargs.
- [x] Add custom tick locator/formatter or explicit tick-list support for
      colorbars, preserving log, boundary, and nonlinear defaults when custom
      ticks are not supplied.
- [x] Expand boundary colorbar rendering for proportional/uniform spacing,
      explicit boundaries/values, drawedges, extension shape variants, and
      visible outline behavior.
- [x] Support colorbars attached to multiple parent axes where layout can be
      represented by the current figure/axes model; document any gridspec-only
      behavior intentionally omitted.
- [x] Add renderer-neutral unit tests for horizontal placement, location/anchor
      geometry, custom ticks, boundary spacing, drawedges, extensions, and
      document multi-axes layout as intentionally omitted for the current
      figure/axes model.
- [x] Add catalog/parity fixtures for horizontal colorbars and boundary
      colorbars with explicit ticks/boundaries.

#### 12.3E Advanced Norms and Color Machinery

- [x] Audit upstream `colors.py` for missing norm families and classify each as
      implement, Go-style equivalent, deferred, or intentional omission.
- [x] Add `AsinhNorm` to the scalar-normalizer catalog if it improves image,
      mesh, or colorbar parity beyond existing asinh axis-scale support.
- [x] Add `FuncNorm` only if a concrete fixture needs user-defined forward /
      inverse color normalization; otherwise document the Go-style alternative
      through custom `ScalarNormalizer` implementations.
- [x] Decide whether multivar/bivar colormaps belong in the v1.0 surface; add a
      narrow implementation only with a visible parity target.
- [x] Decide whether `LightSource` belongs with scalar color machinery or 3D
      surface shading; implement only the subset needed by surface/image parity
      or record an intentional omission.
- [x] Add unit tests for each implemented advanced norm/color helper and update
      public-surface parity rows for omitted helpers with rationale.

#### 12.3F Exit Criteria

- [x] `FoundationAPIGapAudit` and public-surface parity rows for
      `collections.py`, `cm.py`, `colors.py`, `colorbar.py`, and `colorizer.py`
      are updated from broad partial notes to exact implemented, partial,
      Go-style equivalent, deferred, or intentional omission status.
- [x] Mutable scalar-mapped collections can update array, cmap, norm, and clim
      state without reconstructing the artist, and colorbars reflect the
      resulting mapping state through the supported Go API.
- [x] Mesh and colorbar catalog/parity coverage includes mutable scalar mapping,
      horizontal colorbars, and explicit boundary/tick colorbars.
- [x] Remaining advanced norm/color gaps are implemented with tests or recorded
      as intentional omissions with migration guidance.
- [x] `go test ./core ./color ./internal/examplecatalog -count=1` and the
      relevant `go test ./test/ -run ...` catalog cases pass.

Implementation notes:

- Compare against upstream `collections.py`, `cm.py`, `colors.py`,
  `colorbar.py`, and `colorizer.py`.
- Phase 12.3C mesh audit: upstream `_pcolorargs` keeps flat meshes on
  edge-shaped coordinates, expands nearest center coordinates to flat edges, and
  requires Gouraud coordinate grids to match scalar grid shape. The Go
  rectilinear mesh surface matches those visible 1-D coordinate rules. `PColor`
  intentionally aliases the `QuadMesh`-backed `PColorMesh` path; Matplotlib's
  distinct `PolyQuadMesh` return type, masked-coordinate polygon dropping, and
  per-cell hatch/linestyle flexibility remain documented omissions rather than
  cloned API surface.
- Core behavior should be fixed in `core/collection.go`,
  `core/scalar_mappable.go`, `core/norm.go`, and `core/colorbar.go`, not by
  tweaking examples.
- Add Matplotlib reference cases for mutable scalar-mapping and horizontal /
  boundary colorbars.

### 12.4 Patches, Text, Annotation, Legend, and Offset Boxes

**Goal:** close the remaining visible patch, hatch, text, annotation, legend,
and offset-box gaps that still affect static parity output. Prefer focused
typed option surfaces over Python-style dynamic handler registries, but keep the
rendered layout, paths, clipping, and legend samples source-backed by
Matplotlib.

#### 12.4A Landed Baseline

- [x] Common patch artists exist for rectangles, circles, ellipses, polygons,
      paths, wedges, arrows, shadows, regular polygons, arcs, annuli, step
      patches, fancy arrows, and connection patches.
- [x] Hatch rendering is routed through renderer capabilities and covered by
      the existing patch/hatch showcase path.
- [x] Text artists support rotation, alignment, figure/axes/data coordinates,
      bounding boxes, MathText routing, fallback shaping, and renderer text
      metrics.
- [x] Annotation and connection-patch basics exist, including arrow styles,
      connection styles, offset coordinates used by current fixtures, and
      annotation composition coverage.
- [x] Legends exist for line, marker, patch, collection, and figure-level
      samples, with shared artist-label discovery and underscore-label
      filtering.
- [x] Anchored text and axes-grid anchored layout helpers cover the main static
      offset-box use case currently exercised by catalog fixtures.

#### 12.4B Patch and Hatch Catalog Closure

- [x] Audit upstream `patches.py` patch classes, `BoxStyle._style_list`,
      `ArrowStyle._style_list`, and `ConnectionStyle._style_list` against the
      Go patch surface; update Phase 11 public-surface rows for direct
      equivalents, Go-style equivalents, partial rows, and intentional
      omissions.
- [x] Complete `FancyBboxPatch` box-style behavior for the common visible
      styles: square, round, round4, sawtooth, roundtooth, circle, larrow,
      rarrow, darrow, ellipse, and the documented mutation-size /
      mutation-aspect effects.
- [ ] Verify `FancyArrowPatch` and `ConnectionPatch` geometry against upstream
      for shrink points, mutation scale, cap/join style, clipping to patch
      endpoints, and the visible arrow/connection styles already registered.
- [x] Verify hatch character coverage (`/`, `\`, `|`, `-`, `+`, `x`, `o`, `O`,
      `.`, `*`) and repeat-density semantics against `hatch.py`, including
      backend-native vector hatches and AGG raster hatches.
- [x] Add renderer-neutral path tests for the implemented box-style behavior,
      with exact path bounds or segment-count assertions where pixel output
      would be brittle.
- [x] Add focused catalog/parity cases for box styles, connection styles, and
      hatch-density variants instead of expanding `patch_showcase` into an
      overloaded fixture.

Current slice landed:

- `FancyBboxPatch` now supports every upstream `BoxStyle._style_list` entry
  through Go constants: square, round, round4, sawtooth, roundtooth, circle,
  ellipse, larrow, rarrow, and darrow.
- Box-style path generation now applies Matplotlib-style `mutation_size` and
  `mutation_aspect` scaling for padding, rounding, teeth, and arrow geometry.
- Renderer-neutral core tests cover quadratic round corners, mutation
  scale/aspect bounds, circle/ellipse/round4/arrow style bounds, sawtooth and
  roundtooth command structure, and padded left-arrow reflection.
- Renderer-neutral hatch fallback now recognizes the full upstream hatch
  character set, including `o`, `O`, `.`, and `*`, applies repeat-density
  semantics to shape glyphs, and uses the upstream shape-size ratios from
  `hatch.py`.
- AGG native hatching now draws the shape glyphs through the shared hatch
  fallback so raster hatches cover the same character set. Native vector
  pattern definitions for SVG, PDF, PS, and PGF now emit the same shape glyph
  families as vector paths; focused backend tests cover cubic circle geometry,
  filled dot/star geometry, and clipped shape hatch emission.
- `FancyArrowPatch` now applies Matplotlib's default `shrinkA=2` /
  `shrinkB=2` behavior for endpoint-defined arrows, with shrink distances
  converted from points to display pixels through the active figure DPI;
  `ConnectionPatch` keeps its upstream zero-shrink default and explicit
  shrink values use the same point-unit conversion.
- Renderer-neutral patch tests cover DPI-correct default FancyArrowPatch
  endpoint shrinking, explicit ConnectionPatch shrink values, and
  independent-coordinate ConnectionPatch endpoints.
- `FancyArrowPatch` / `ConnectionPatch` now expose `PatchA` / `PatchB` and
  clip endpoint-defined connection paths to those patch boundaries before
  applying shrink distances, matching upstream `patchA` / `patchB` ordering
  for common static patch shapes.
- Renderer-neutral patch tests cover data-space rectangle endpoint clipping
  without backend pixels.
- `ConnectionStyle("arc3", rad=0)` now preserves Matplotlib's quadratic
  Bezier path form, with the control point on the midpoint instead of
  collapsing to a line segment.
- Renderer-neutral patch tests cover zero-rad Arc3 command structure and
  default FancyArrowPatch shrink on the quadratic path.
- `FancyArrowPatch` and `ConnectionPatch` draw paths now use Matplotlib's
  default round cap / round join styling when no explicit patch cap/join style
  is set.
- Renderer-neutral patch tests cover default FancyArrowPatch cap/join paint
  routing.
- `FancyArrowPatch.MutationAspect` now follows upstream ArrowStyle mutation
  semantics by squeezing display-space Y before arrow transmutation and
  restoring it afterward.
- Renderer-neutral patch tests cover mutation-aspect scaling without moving
  the arrow tip.
- `ArrowStyle("wedge")` now has source-backed `shrink_factor` parsing and a
  wedge-specific tapered path, instead of reusing the generic filled-arrow
  mutation.
- Wedge arrow mutation now follows quadratic connection control points for
  curved `arc3`-style paths instead of collapsing the connection to a straight
  start/end segment.
- Renderer-neutral patch tests cover Wedge tail width, midpoint shrink,
  endpoint tapering, and quadratic-connection geometry.
- `ArrowStyle("simple")` and `ArrowStyle("fancy")` now also follow quadratic
  connection control points for curved paths instead of collapsing filled
  arrows to a straight start/end segment.
- Renderer-neutral patch tests cover simple and fancy filled-arrow geometry on
  quadratic connections.
- Curve-style arrow mutations now shorten the stroked connection line under
  begin/end arrow heads while keeping the arrow head anchored at the original
  endpoint, matching Matplotlib's `_Curve` line/head split.
- Renderer-neutral patch tests cover line shortening for `->` without moving
  the arrow-head tip.
- Curve-style arrow-head line shortening now uses Matplotlib's linewidth
  projection pad instead of the full arrow-head length, matching `_Curve`'s
  `_get_arrow_wedge` geometry.
- Renderer-neutral patch tests cover source-derived projected line/head tips.
- `ConnectionStyle("arc")` now uses Matplotlib's style-specific defaults
  (`angleA=0`, `angleB=0`) instead of inheriting the `Angle` / `Angle3`
  `angleA=90` default.
- `ConnectionStyle("arc", rad=...)` now rounds at arm/elbow vertices with the
  upstream vertex sequence instead of rounding at the final endpoint.
- Renderer-neutral patch tests cover the default horizontal start arm for
  `arc,armA=...,armB=...` and rounded arm-end geometry for
  `arc,armA=...,rad=...`.
- `ConnectionStyle("bar", angle=...)` now projects the intermediate endpoint
  onto the requested connecting angle before constructing the second arm,
  matching upstream `Bar.connect` behavior while preserving the original final
  endpoint.
- Renderer-neutral patch tests cover angled bar projection against the
  source-derived horizontal-angle geometry.
- `ArrowStyle("|-|")` now uses the upstream zero-length bar-bracket defaults
  instead of inheriting the square-bracket protrusion length from the generic
  curve style.
- Renderer-neutral patch tests cover `|-|` parser defaults and non-protruding
  endpoint bar geometry.
- Curve-style bracket arrows now parse and apply Matplotlib's `scaleA` /
  `scaleB` overrides, so bracket width/length can be scaled independently from
  the arrow mutation size.
- Renderer-neutral patch tests cover source-backed bracket scaling.
- Added the focused `patch_style_matrix` parity fixture for Phase 12.4 patch
  coverage. It separates box-style, hatch-density, ArrowStyle, and
  ConnectionStyle visual coverage from the broader `patch_showcase` fixture.

#### 12.4C Text and Font Property Breadth

- [x] Audit upstream `text.py`, `font_manager.py`, and `textpath.py` against
      `core.Text`, `render.TextShapingOptions`, and the renderer text
      capability interfaces; classify every public-surface gap as implement,
      Go-style equivalent, deferred, or intentional omission.
- [x] Expand per-text font options for family, style, weight, stretch, variant,
      math font selection, parse-math behavior, and font-feature hooks where
      the current shaping/font-manager layer can support them deterministically.
- [x] Define the deterministic Go policy for Matplotlib font fallback behavior:
      which upstream fontconfig/font-manager features are implemented directly,
      which are approximated by bundled/default fonts, and which require a
      user-supplied font path or family registration.
- [x] Tighten text bounding-box, baseline, rotation-mode, multiline, wrapping,
      and `bbox` patch behavior against upstream where current layout/text
      validation fixtures still carry visible residuals.
- [x] Ensure text alpha, clipping, z-order, path effects, and rasterization
      follow the shared artist metadata path rather than bespoke text-only
      routing.
- [x] Add renderer-neutral tests for font option resolution and text bounds,
      plus catalog/parity fixtures for font variants, multiline layout, rotated
      anchored text, and text-with-bbox output.

Current slice landed:

- `TextOptions`, `Figure.Text`, and `AnnotationOptions` now carry an explicit
  `FontKey` override that takes precedence over `ctx.RC.FontKey` for
  measurement, drawing, multiline text, rotated text, path effects, and
  annotations.
- `TextOptions` and `AnnotationOptions` also accept structured
  `render.FontProperties` for family, style, weight, and direct font-file
  requests. The renderer-facing `render.FontPropertiesKey` preserves those
  fields across existing font-key interfaces.
- Renderer-neutral text tests cover per-text and per-annotation font-key
  override routing, structured font-property round-tripping, and
  font-property routing through font-aware renderer interfaces.
- Structured `render.FontProperties` now also preserves stretch, variant,
  language, and OpenType feature toggles across renderer font-key interfaces.
  The shared shaping layer merges encoded feature toggles into
  `TextShapingOptions`, so per-text font properties can deterministically
  disable features such as `liga`.
- Renderer-neutral font/shaping tests cover extended FontProperties
  round-tripping, structured feature routing through `Text` artists, and
  `liga=0` shaping from an encoded font-properties key.
- Structured `render.FontProperties` now includes `MathFontFamily`, matching
  Matplotlib's high-value `Text.set_math_fontfamily` / `FontProperties`
  control for MathText. The MathText resolver maps the requested font family
  into the existing deterministic font manager path while preserving explicit
  MathText style switches such as `\mathrm` and `\mathsf`.
- Renderer-neutral MathText tests cover `MathFontFamily` round-tripping and
  default MathText run routing through a requested DejaVu Serif math family.
- `TextOptions` and `AnnotationOptions` now expose `ParseMath *bool`, matching
  Matplotlib's high-value per-artist `parse_math=False` control. When disabled,
  dollar-delimited text bypasses MathText preprocessing and is drawn as plain
  text while TeX routing remains controlled by `RC.UseTeX`.
- Renderer-neutral text tests cover `ParseMath=false` for both text artists and
  annotations.
- Artist-level `SetAlpha` now participates in text rendering: text artists
  multiply local text color alpha by shared artist metadata, and annotations
  apply the same multiplier to both annotation text and arrow colors.
- Renderer-neutral text tests cover artist-alpha routing for standalone text
  and annotation text/arrow output.
- `TextOptions.WrapWidth` now provides explicit display-pixel word wrapping
  for text artists, reusing the multiline layout and text-bbox drawing path for
  wrapped output.
- `TextOptions.Wrap` now computes a Matplotlib-style wrap width from the figure
  box when `WrapWidth` is unset, using alignment and rotation to choose the
  available edge distance.
- Renderer-neutral text tests cover explicit wrapped line emission, wrapped
  text-bbox width constraints, and figure-box automatic wrapping.
- `TextOptions.MultiAlignment` now mirrors Matplotlib's high-value
  `multialignment` behavior for multiline and wrapped text: nil follows
  `HAlign`, while an explicit value controls per-line placement inside the
  already aligned text block.
- Renderer-neutral text tests cover left-aligned line placement inside a
  right-aligned multiline block.
- `TextOptions` and `AnnotationOptions` now expose `Linespacing`, matching the
  high-value numeric form of Matplotlib's multiline `linespacing` control while
  keeping the zero value on normal 1.2 spacing.
- Renderer-neutral text tests cover explicit baseline advance for multiline
  text.
- Multiline text now routes per-line glyph paths through `PathEffects` before
  falling back to normal text draws, matching the single-line text path-effect
  behavior.
- Renderer-neutral text tests cover path-effect glyph-path routing for
  multiline text.
- Text bbox patches now rotate with rotated text around the same renderer
  anchor as the text draw call instead of remaining axis-aligned; multiline
  text block bboxes follow the rotated text path instead of staying axis-aligned.
- Renderer-neutral text tests cover rotated single-line and multiline bbox path
  geometry.
- `TextOptions.RotationMode` now supports a Go-style
  `TextRotationModeAnchor` equivalent for Matplotlib's `rotation_mode='anchor'`,
  aligning the unrotated text box before rotation.
- `TextRotationModeXTick` and `TextRotationModeYTick` now apply Matplotlib's
  source-backed angle bands for tick-style horizontal / vertical alignment.
- Renderer-neutral text tests cover anchor-mode rotated text anchors and
  xtick / ytick alignment separately from the default rotated-bbox alignment
  behavior.
- `TextVerticalAlign` now exposes `TextVAlignCenterBaseline`, matching
  Matplotlib's `verticalalignment="center_baseline"` for text and annotation
  layout.
- Renderer-neutral text tests cover single-line center-baseline origin
  placement against measured ascent.
- Multiline text and annotation block layout now use renderer font metrics
  (`MinAscent`, `MinDescent`, and `LineGap`) for normal line spacing and
  numeric `Linespacing`, matching upstream `text.py`'s font-metric line-box
  policy instead of deriving baseline advances from raw point size.
- Renderer-neutral text tests cover normal multiline baseline advance from
  font height plus line gap and numeric multiline baseline advance from
  `linespacing * font height`.

#### 12.4D Annotation and Offset-Box Behavior

- [x] Audit upstream annotation coordinate systems against `Annotate` and the
      transform helpers: data, axes fraction, figure fraction, offset points,
      offset pixels, blended coordinates, artist-relative coordinates, and
      callable/custom coordinate providers.
- [x] Implement or explicitly omit missing annotation clipping semantics:
      `annotation_clip`, clipping to the annotated point, clipping to text /
      arrow patch paths, and interaction with axes clipping.
- [x] Add `AnnotationBbox` or a Go-style equivalent for image/text/box content
      anchored to data or display coordinates, with arrow connection support.
- [x] Add the offset-box families needed for static parity:
      anchored offset box, text area, drawing area, offset image, horizontal /
      vertical packers, and anchored size-bar style layouts.
- [x] Keep draggable, GUI-only, and callback-heavy offset-box behavior out of
      v1.0 unless a concrete interactive fixture requires it; record those
      omissions in Phase 11 rows.
- [x] Add catalog/parity fixtures for annotation clipping, annotation box
      content, packed offset boxes, and anchored size bars.

Current slice landed:

- `AnnotationOptions` and `Annotation` now expose `AnnotationClip`, matching
  Matplotlib's explicit `annotation_clip=True/False` control for suppressing
  annotation text/arrow drawing when the annotated point lies outside the axes
  clip.
- The default `annotation_clip=None` policy now follows Matplotlib's data-only
  rule: outside data-coordinate annotations are clipped by default, while
  outside non-data annotations still draw unless explicitly clipped.
- Renderer-neutral annotation tests cover default clipping, explicitly clipped,
  and explicitly unclipped outside-point behavior.
- `AnnotationOptions` now accepts `BBox`, reusing `TextBBoxOptions` so
  annotation text can draw a styled background patch and route arrow start
  points from the expanded bbox instead of the raw text ink bounds.
- Renderer-neutral annotation tests cover bbox paint emission without backend
  pixels.
- `AnnotationOptions.Angle` now routes annotation text through the same rotated
  text renderer and rotated bbox path as `Text`, matching Matplotlib's
  annotation-as-text rotation behavior for static output.
- Renderer-neutral annotation tests cover rotated annotation text routing
  without backend pixels.
- `Annotation` now reuses multiline text layout for newline-separated
  annotation text, including bbox / arrow-start block geometry and
  artist-alpha routing through the shared text path.
- Renderer-neutral annotation tests cover multiline annotation text splitting
  without backend pixels.
- `Axes.AnnotationBbox` now provides a static Go-style equivalent for the
  common `AnnotationBbox(TextArea(...))` path: text-area content, separate
  annotated-point and box coordinate systems, box alignment, frame visibility,
  padding/styling, and optional arrow connection.
- Renderer-neutral annotation-box tests cover text placement, frame paint, and
  arrow endpoint routing without depending on backend pixels.
- `Axes.AnnotationBbox` also supports `Image` plus `ImageZoom`, covering the
  static `AnnotationBbox(OffsetImage(...))` path through the renderer-neutral
  `render.Image` contract. Box alignment converts Matplotlib's lower-left
  alignment semantics into the port's display coordinate convention.
- Renderer-neutral annotation-box tests cover zoomed image destination
  placement without backend pixels.
- `Axes.AddAnchoredSizeBar` now covers the common axes-grid
  `AnchoredSizeBar` static use case: data/axes/figure coordinate bar lengths,
  center-aligned label placement, optional frame, label-top mode, vertical bar
  thickness, and filled-bar behavior.
- Renderer-neutral anchored-layout tests cover data-scaled bar length, label
  placement, and frame paint.
- `Axes.AddAnchoredDrawingArea` now covers the common static
  `AnchoredDrawingArea` / `DrawingArea` path for fixed-size local path content,
  including anchored placement, optional frame, padding, and lower-left local
  coordinate mapping into display space. It also supports clipping children to
  the drawing-area bounds, matching upstream `DrawingArea(clip=True)`.
- Renderer-neutral anchored-layout tests cover local path coordinate mapping
  and frame paint, plus child clipping to the drawing-area bounds.
- `Axes.AddAnchoredPacker` now covers the common static `HPacker` / `VPacker`
  shape for anchored offset boxes: fixed drawing-area children, text-area
  children, and zoomed image children can be packed horizontally or vertically
  with explicit separation, padding, frame styling, and start/center/end
  cross-axis alignment.
- Renderer-neutral anchored-layout tests cover horizontal drawing/text packing,
  text placement, vertical child stacking, cross-axis alignment, and zoomed
  image child placement.
- Added the focused `text_annotation_matrix` parity fixture for Phase 12.4 text
  and annotation coverage. It exercises structured font properties, multiline
  text, rotated text, text bbox output, explicit annotation clipping,
  AnnotationBbox text/image content, anchored text, anchored drawing areas,
  packed offset boxes, and anchored size bars.

#### 12.4E Legend Handler and Layout Closure

- [x] Audit upstream `legend.py` and `legend_handler.py` for location,
      anchoring, column layout, title handling, frame styling, handle length,
      handle text padding, label spacing, marker scaling, scatter-point
      sampling, and handler-map behavior.
- [x] Add Go-style legend handler registration for custom samples and proxy-like
      entries without exposing Python's arbitrary object-dispatch surface.
- [x] Tighten built-in legend samples for lines, markers, patches,
      path/line/patch collections, error bars, stems, bars, filled bands, and
      scalar-mapped collections.
- [x] Implement multi-column and figure-level legend layout behavior needed by
      current catalog/showcase examples, including title placement and
      constrained-layout participation.
- [x] Verify `"best"` placement badness against upstream for representative
      line, scatter, image, and annotation cases; document any deliberate
      simplification with a migration note.
- [x] Add renderer-neutral legend-layout tests and catalog/parity fixtures for
      custom/proxy handlers, multi-column legends, scatter sample counts, and
      figure legends.

Current slice landed:

- `Legend` now exposes `NumColumns` and `ColumnSpacing` for Go-style
  multi-column static layout. Entries are split across columns using the same
  contiguous-column distribution as Matplotlib's `ncols` packing, and the
  shared layout calculation drives both drawing and `boxRect` placement.
- Renderer-neutral legend tests cover multi-column label origins and row
  alignment without depending on backend pixels.
- `Legend` also exposes `MarkerScale` and `ScatterPoints` for static marker
  legend samples, covering the high-value `markerscale` and `scatterpoints`
  controls from upstream legend handlers.
- Renderer-neutral legend tests cover scaled marker bounds and multiple
  left-to-right scatter sample positions.
- `Legend` now supports `Title` and `TitleFontSize`, drawing the title above
  entries and accounting for title width/height in the shared layout and
  `boxRect` placement path.
- Renderer-neutral legend tests cover title drawing, title placement above the
  first entry, and increased legend box height.
- Figure-level legends collect labeled artists across axes and participate in
  the figure-artist stacking path used for suptitle / supxlabel / supylabel
  composition.
- Renderer-neutral figure-layout tests cover figure legend collection and
  suptitle stacking.
- `Legend` now exposes `FrameOn`, mirroring the high-value upstream `frameon`
  control for suppressing only the legend frame while preserving samples, text,
  and the shared layout/placement calculation.
- Renderer-neutral legend tests cover frame suppression without dropping legend
  content.
- `Legend.AddEntry` now provides a typed Go proxy-entry surface through
  `LegendEntryOptions` and `LegendSampleLine` / `LegendSampleMarker` /
  `LegendSamplePatch`, reusing the same internal sample rendering path as
  collected artists.
- Renderer-neutral legend tests cover explicit proxy patch entries drawn
  without a backing artist.
- `Legend.SetHandler` now provides a typed per-artist handler override for
  collected artists, using `LegendEntryOptions` instead of Python's arbitrary
  object-dispatch handler maps. `Legend.ClearHandler` removes the override.
- Renderer-neutral legend tests cover a collected line artist rendered through a
  custom patch sample while preserving the artist label.
- Built-in `ErrorBar` legend entries now draw an errorbar-specific sample with
  a center line, x/y error stems as applicable, caps, and optional markers
  instead of degrading to a plain line-only sample.
- Renderer-neutral legend tests cover y-error stems and caps in the legend
  sample path output.
- Stem plots now collect as a single combined legend sample when the adjacent
  stem line collection and marker collection share a label, instead of drawing
  duplicate line-only and marker-only legend entries.
- Renderer-neutral legend tests cover combined stem legend collection.
- `LegendBest` placement now uses a Matplotlib-style static badness heuristic
  over display-space artist samples. It avoids line vertices, scatter /
  collection offsets, image extent corners, and annotation / AnnotationBbox
  anchors, then preserves Matplotlib's lower-location-code tie break. The Go
  heuristic intentionally stays point/anchor based for v1.0 instead of cloning
  Matplotlib's full bbox/path-intersection scoring and renderer-dependent text
  window extents.
- Renderer-neutral legend tests cover representative best-placement avoidance
  for line, scatter, image, and annotation cases.
- Added the focused `legend_layout_matrix` parity fixture for Phase 12.4 legend
  coverage. It exercises multi-column layout, title drawing, scatter sample
  counts, marker scaling, errorbar samples, proxy entries, frame suppression,
  and typed handler overrides separately from broad composition examples.

#### 12.4F Exit Criteria

- [x] `FoundationAPIGapAudit` rows for patch/hatch catalogs,
      text/font-surface gaps, annotation/offset boxes, and legend-handler
      behavior are closed or split into exact remaining rows.
- [x] Public-surface parity rows for `patches.py`, `hatch.py`, `text.py`,
      `font_manager.py`, `legend.py`, `legend_handler.py`, and `offsetbox.py`
      no longer contain broad "partial" notes without a precise remaining
      task.
- [x] 12.4 has catalog/parity coverage for box styles, hatch density,
      text/font variants, annotation clipping/boxes, offset boxes, and legend
      handler/layout behavior.
- [x] `go test ./core ./render ./internal/examplecatalog -count=1` and the
      relevant `go test ./test/ -run ...` catalog cases pass.

Current slice landed:

- Split the Phase 12.4 foundation audit into exact patch/hatch,
  text/font-layout, font-property, annotation-coordinate, offset-box, and
  legend-handler rows so remaining work is no longer hidden in a broad
  text/legend/offsetbox bucket.
- Expanded the committed public-surface inventory to include `hatch.py`,
  `font_manager.py`, `textpath.py`, and `legend_handler.py`, with explicit
  parity classifications for hatch geometry, font-manager policy, text paths,
  and legend-handler behavior.
- Added focused tests that keep those 12.4 audit and inventory rows from
  regressing back into broad untracked gaps.

Implementation notes:

- Compare against upstream `patches.py`, `hatch.py`, `text.py`,
  `font_manager.py`, `textpath.py`, `legend.py`, `legend_handler.py`, and
  `offsetbox.py`.
- Add small catalog fixtures for each style family; avoid a single giant patch
  fixture that is hard to debug.
- Keep API shapes Go-idiomatic, but the rendered output should follow
  Matplotlib where behavior is visual.
- Fix visible output through `core/patch*.go`, `core/text*.go`,
  `core/annotation*.go`, `core/legend.go`, renderer text/path capabilities, and
  shared geometry/transform helpers, not by tweaking examples.

#### 12.4G Display Coordinate and Backend Boundary Parity

**Goal:** remove y-orientation and renderer-boundary mismatches that make
Matplotlib display-space geometry diverge from the Python reference. This is a
foundational parity issue, not a Go-idiomatic exception: signed geometry such as
connection-style curvature, arrow normals, shrink/clipping, text rotation, and
backend path rasterization must match upstream Matplotlib semantics as closely
as possible.

**Reference sources:** `third_party/matplotlib/lib/matplotlib/transforms.py`,
`patches.py`, `text.py`, `backend_bases.py`, `_backend_agg.cpp`, local
`core/artist.go`, `core/arrow_patch.go`, `core/text.go`,
`backends/agg/`, and, when the mismatch is in the AGG port rather than the
matplotlib-go core boundary, `../agg_go` compared against `../agg_2.4`.

- [ ] Define and document the coordinate contract between core display-space
      geometry and renderer backends. The contract must state where Matplotlib's
      display-coordinate model is preserved, where raster backends convert to
      device pixels, and which layer owns any y-axis inversion.
- [ ] Add renderer-neutral regression tests for signed display-space geometry:
      `ConnectionStyle("arc3", rad=...)` control-point direction,
      `FancyArrowPatch` / `ConnectionPatch` shrink and patch clipping on curved
      paths, arrow-head normals, rotated text bbox orientation, and annotation
      arrow start points after bbox clipping.
- [ ] Audit existing y-sign compensations in `core/patch*.go`, `core/text*.go`,
      `core/annotation*.go`, transforms, and parity fixtures. Keep only
      source-backed conversions at explicit coordinate-system boundaries; remove
      example-level or fixture-level sign workarounds.
- [ ] Reconcile `text_annotation_matrix`'s `bbox arrow` against upstream
      `Annotation.update_positions` and `FancyArrowPatch._get_path_in_displaycoord`:
      the arrow must start from default `relpos=(0.5,0.5)`, clip to the text
      bbox patch, shrink in points, and follow the same `arc3` curvature as
      Matplotlib.
- [ ] Validate the same coordinate-boundary fix against broader fixtures that
      exercise signed display geometry: `patch_style_matrix`,
      `annotation_composition`, `transform_coordinates`, and any focused
      renderer-neutral tests added for this phase.
- [ ] If a residual is proven to originate in the Go AGG port rather than core
      coordinate conversion, inspect `../agg_go` against `../agg_2.4` and fix
      the port instead of adding compensating behavior in this repository.
- [ ] Track before/after metrics for the affected catalog cases. The immediate
      `text_annotation_matrix` target is `RMSE < 10` against the Matplotlib
      reference without changing the example source away from a 1:1 port.

Exit criteria:

- [ ] Signed display-space paths, annotations, rotated text bboxes, and arrow
      geometry are source-backed by upstream Matplotlib formulas under the
      documented coordinate contract.
- [ ] No `text_annotation_matrix`-specific hacks, sign flips, or extra settings
      exist in the example or core implementation.
- [ ] `TestMatplotlibRef/text_annotation_matrix` reports `RMSE < 10`, and
      related connection/annotation fixtures do not regress.
- [ ] Any remaining mismatch is classified with evidence as core behavior,
      renderer boundary behavior, AGG-port behavior, or documented upstream
      limitation.

#### 12.4G Execution Plan — Global y-up display-space conversion ("valley of tears")

**Decision:** be faithful to upstream matplotlib (`third_party`, Agg-based):
core display space is **y-up** (origin bottom-left), and **each backend owns the
device y-inversion at rasterization** (matplotlib `flipy()==True`: positions
flip, glyph/image bitmaps stay upright). Signed-geometry formulas (`arc3`
curvature, arrow normals, rotation, offsets) stay verbatim and become correct.

**Guiding invariant — net-neutral:** the flip merely MOVES from the transform
layer to the backend, so unsigned content rasterizes byte-identically; only
signed display geometry changes. Oracle: `go test ./test/ -run TestGolden`
(fresh render vs committed y-down golden) — a correct conversion makes a fixture
pass again; only signed-geometry/annotation fixtures legitimately change.

Status: [x] done · [~] in progress · [ ] todo.

- **G1 Contract & core pivot** — [x]
  - [x] `docs/adr/0003-display-coordinate-contract.md`.
  - [x] `transform/graph.go` `NewDisplayRectTransform` → y-up `(Min.Y,Max.Y)`.
  - [x] `core/artist.go` `layout()` → drop `1-frac` flip.
- **G2 AGG backend owns device flip** — [x]
  - [x] `backends/agg/agg_flip.go` helpers; convert y-up→device once at each
    public entry (Path/markers/gouraud/image/clip/text); markers snap in device;
    text/image bitmaps upright at `H-y`. `go build ./backends/agg/...` passes.
- **G3 Core positioning/text helpers → y-up** — [~]
  - [~] `textBaselineOffset`, multiline stacking, title/`titleTopExtent`,
    spines/extents, anchored locator, tick label pad + tick-mark outsets,
    `textInkRect`, rotation sense, `annotationVAlign`. Drive TestGolden failures
    (91 → only signed-geometry/annotation fixtures).
- **G4 AGG parity validation** — [ ]
  - [ ] `text_annotation_matrix` arc curves like mpl; `TestMatplotlibRef` RMSE
    `< 10`.
  - [ ] Net-neutral check: regenerate ONLY signed-geometry/annotation goldens;
    everything else byte-identical (`git diff --stat testdata/golden`).
- **G5 Examples → 1:1 ports** — [ ]
  - [ ] `text_annotation_matrix` `OffsetY +→-` (match `xytext` y-up); sweep
    others (most already use Python's sign).
- **G6 Vector/other backends own inversion** — [ ]
  - [ ] PDF/PS/PGF: remove redundant flip (now identity); audit text/image
    submatrices; regenerate structural goldens.
  - [ ] SVG: per-coordinate `H-y` at emission.
  - [ ] gobasic: per-draw flip (backs skia); webagg/desktop inherit (verify).
- **G7 Full-suite regen & revalidate** — [ ]
  - [ ] `go test ./...`; regenerate goldens; `just fmt && just lint && just test`.
- **G8 Renderer-neutral regression tests** — [ ]
  - [ ] arc3 control-point direction, FancyArrowPatch/ConnectionPatch
    shrink+clip on curves, arrow-head normals, rotated-text bbox orientation,
    annotation arrow start after bbox clip.

### 12.5 Images, Pyplot, Backends, Widgets, and Animation

**Goal:** finish the remaining high-value image, stateful wrapper, backend
lifecycle, widget, and animation parity decisions so v1.0 users get either a
working Go equivalent or a documented, explicit omission. Keep the
object-oriented API primary; use pyplot, widgets, and animation as migration
convenience layers over the core model.

#### 12.5A Landed Baseline

- [x] `Image2D`, `ImShow`, `MatShow`, and `Spy` cover matrix images with
      scalar mapping, alpha, origin, extent, transformed images, and colorbar
      integration.
- [x] AGG routes Matplotlib interpolation names through image filters,
      including `auto` / `antialiased` scale-dependent behavior and the current
      high-value interpolation fixtures.
- [x] `pyplot` provides current figure/current axes state, common plot wrappers,
      labels, legends, colorbars, save/show helpers, rc/rc-context support, and
      image/matrix helpers.
- [x] The canvas layer has dispatcher, picker, hover, scheduler, draw-idle,
      toolbar, manager, save, and backend capability contracts used by current
      headless, desktop, and web backends.
- [x] Static widget artists and interaction routing exist for buttons, sliders,
      range sliders, check buttons, radio buttons, text boxes, span selectors,
      and rectangle selectors.
- [x] The `animation` package provides deterministic `FuncAnimation` /
      `ArtistAnimation` style stepping, event-loop scheduling, animated artist
      tracking, and blit-region hooks.

#### 12.5B Image Class and Resampling Closure

- [x] Audit upstream `image.py` image artist classes and helper functions:
      `AxesImage`, `FigureImage`, `BboxImage`, `NonUniformImage`,
      `PcolorImage`, `imread`, `imsave`, `imresize`-style omissions,
      `figimage`, and image origin/extent/aspect defaults.
- [x] Decide which non-`AxesImage` classes belong in v1.0:
      implement `FigureImage` / `figimage` and `BboxImage` if they improve
      figure-level composition or annotation-box parity; implement
      `NonUniformImage` / `PcolorImage` as existing `PColor` / `PColorFast` /
      `PColorMesh` equivalents unless a visible fixture shows a meaningful
      difference from `Image2D` / `PColorMesh`.
- [x] Convert every unsupported image class/helper into either a typed Go
      equivalent, a clear runtime error, or a Phase 11 intentional omission
      with migration guidance.
- [x] Lock interpolation coverage for `nearest`, `none`, `bilinear`,
      `bicubic`, `hanning`, `hamming`, `lanczos`, `spline16`, `spline36`,
      `kaiser`, `quadric`, `catrom`, `gaussian`, `bessel`, `mitchell`, `sinc`,
      `blackman`, `hermite`, `antialiased`, and `auto` with parser tests,
      AGG-render tests, and at least one visible interpolation gallery case.
- [x] Verify transformed-image sampling, image alpha premultiplication, clipping,
      resample threshold decisions, and backend fallbacks against upstream
      `image.py` and AGG behavior.
- [x] Add catalog/parity fixtures for figure-level images or bbox images if
      implemented, plus an image-interpolation matrix fixture that is small
      enough to debug visually.

#### 12.5C Pyplot and Stateful Wrapper Surface

- [x] Audit high-traffic upstream `pyplot.py` and `_pylab_helpers.py` functions
      against the Go `pyplot` package; rank missing wrappers by migration value
      rather than trying to clone every overload.
- [x] Add wrappers for implemented core features where the current absence is
      a migration blocker: common axes creation, subplot/subplots variants,
      plotting/image/stat helpers, axis/tick scale helpers, annotations,
      legends, colorbars, rc/style helpers, and save/show helpers.
- [x] Preserve object-oriented APIs as the source of truth; pyplot wrappers
      should delegate to `core` without maintaining duplicate rendering or
      layout behavior.
- [x] Define clear error behavior for unsupported pyplot overloads, implicit
      figure-manager behavior, interactive mode toggles, and global state reset
      functions.
- [x] Add unit tests that pyplot wrappers delegate to the same core state as
      the object-oriented calls, plus migration-style examples for common
      pyplot workflows.
- [x] Update public-surface parity rows for `pyplot.py` and `_pylab_helpers.py`
      so each broad wrapper gap is either implemented, scoped to a smaller
      wrapper family, or intentionally omitted.

#### 12.5D Backend Canvas, Manager, and Tool Lifecycle

- [x] Audit upstream `backend_bases.py`, `backend_tools.py`, and
      `_pylab_helpers.py` against `canvas`, backend registry metadata, and the
      interactive backend implementations.
- [x] Complete canvas/manager lifecycle semantics for figure creation,
      current-manager tracking, draw vs draw-idle, resize, close/destroy,
      toolbar attachment, save dispatch, and backend capability reporting.
- [x] Tighten event contracts for mouse, key, scroll, pick, figure enter/leave,
      axes enter/leave, timer events, and callback connection/disconnection.
- [x] Complete toolbar/tool behavior needed by interactive backends: home,
      back, forward, pan, zoom, configure, save, cursor/status messages, mode
      state, and tool enablement.
- [x] Add backend-neutral lifecycle tests using fake canvases/managers and
      smoke tests for WebAgg/Gio where the environment allows them.
- [x] Document backend-specific omissions, especially GUI toolkit behaviors
      that cannot be represented in the current headless test environment.

#### 12.5E Widgets and Interaction Scope

- [x] Audit upstream `widgets.py` classes and decide v1.0 status for each:
      button, slider, range slider, check buttons, radio buttons, text box,
      span selector, rectangle selector, lasso selector, polygon selector,
      cursor/multi-cursor, annotated cursor, menu/tool widgets, and any
      deprecated or GUI-specific widgets.
- [x] Tighten existing widget behavior for callback ordering, active/disabled
      state, hover/press/release transitions, keyboard activation, value
      clamping, snapping, dragging, redraw policy, and axes ownership.
- [x] Implement only selector/cursor widgets that can be expressed through the
      current canvas event model; document GUI-only or callback-heavy widgets as
      intentional omissions until an interactive fixture requires them.
- [x] Add catalog/browser-demo coverage for the supported widget set and
      renderer-neutral tests for interaction state transitions.
- [x] Ensure widgets compose with artist picking, overlays, draw-idle, and
      animation without stealing unrelated user events.

#### 12.5F Animation Scope and Writers

- [x] Audit upstream `animation.py` for `Animation`, `TimedAnimation`,
      `FuncAnimation`, `ArtistAnimation`, frame sequence behavior, repeat /
      repeat-delay, blitting, save_count/cache behavior, HTML representation,
      and movie writer APIs.
- [x] Tighten `FuncAnimation` and `ArtistAnimation` behavior against upstream
      for initialization order, frame iteration, repeat semantics, event-source
      lifecycle, animated artist visibility, and blit background restoration.
- [x] Decide v1.0 writer scope: implement a small explicit writer surface for
      GIF/MP4 only if dependencies and backend output are deterministic, or
      document animation saving as intentionally omitted while interactive
      playback remains supported.
- [x] Add examples for at least one timer-driven line update, one artist-list
      animation, and one blit-capable animation path; browser-demo breadth is
      deferred to Phase 13 with the rest of the user-facing gallery work.
- [x] Add unit tests for frame sequencing, stop/start lifecycle, repeat-delay,
      blit fallback, and error handling for unsupported writer paths.

#### 12.5G Exit Criteria

- [x] `FoundationAPIGapAudit` rows for image class breadth, pyplot wrapper
      surface, backend lifecycle, widgets, and animation are closed or split
      into precise remaining rows.
- [x] Public-surface parity rows for `image.py`, `pyplot.py`,
      `_pylab_helpers.py`, `backend_bases.py`, `backend_tools.py`,
      `widgets.py`, and `animation.py` no longer contain broad "partial" notes
      without a precise remaining task.
- [x] Every unsupported image interpolation, pyplot wrapper family, widget, or
      animation writer path has a clear error or documented omission; none
      silently fall back to an incorrect behavior.
- [x] 12.5 has catalog/parity, example, browser-demo, or backend-neutral test
      coverage for image class/resampling decisions, pyplot migration
      workflows, backend lifecycle semantics, supported widgets, and animation
      playback.
- [x] `go test ./core ./pyplot ./canvas ./animation ./backends/agg`
      `./internal/examplecatalog -count=1` and relevant catalog/browser tests
      pass.

Current slice landed:

- Added explicit Phase 12.5 foundation audit rows for widget/selector
  interaction scope and animation playback/writer scope, complementing the
  existing image, pyplot, and backend lifecycle rows.
- Expanded the committed public-surface inventory to include
  `_pylab_helpers.py` so pyplot state-management parity is tracked alongside
  `pyplot.py`.
- Tightened 12.5 public-surface parity notes for image, pyplot,
  `_pylab_helpers`, backend bases/tools, widgets, and animation so each
  partial row names exact remaining families instead of broad catch-all gaps.
- Added pyplot `Text` and `Annotate` wrappers that delegate directly to the
  current axes, with a focused stateful-wrapper test covering returned core
  artists and current-axes ownership.
- Added pyplot `AxHLine`, `AxVLine`, `AxLine`, `AxLineSlope`, `AxHSpan`, and
  `AxVSpan` wrappers for the existing core reference-line/span helpers, with a
  focused delegation test covering returned core artists and current-axes
  ownership.
- Added `imshow_interpolation_matrix` as the focused Phase 12.5 image fixture:
  the AGG parser and render tests enumerate every Matplotlib interpolation
  registry name, and committed Go / Matplotlib PNGs make the full resampling
  matrix visually inspectable. The local Matplotlib reference script falls back
  from `auto` to `antialiased` only when the installed Matplotlib predates the
  vendored upstream `auto` registry name; Go behavior remains tested against
  the vendored upstream list.
- Added `MatShowOptions.Interpolation` so matrix-style `matshow` helpers can
  select the same renderer interpolation filters as `imshow` / `Image2D`
  instead of silently being limited to the renderer default.
- Added explicit public-surface classifications for image.py's remaining
  image classes and IO helpers: BboxImage maps to existing annotation/anchored
  image composition, NonUniformImage / PcolorImage map to PColor/PColorMesh,
  FigureImage and thumbnail are documented omissions with migration guidance,
  and imread / imsave now have explicit partial Go helpers. The broad image.py
  row now only carries transformed-resampling edge behavior as remaining
  partial scope.
- Added typed `PColorFast` aliases on `Axes` and `pyplot` so Matplotlib's
  axes-level fast pseudocolor entry point maps to the existing QuadMesh-backed
  `PColorMesh` path instead of remaining an unresolved image fast-path.
- Added explicit public-surface rows and pyplot delegation coverage for
  `pyplot.pcolor` and `pyplot.pcolormesh`, tying the stateful pseudocolor
  wrappers to the existing mesh parity fixtures.
- Added pyplot `ImShow` as a stateful wrapper over `Axes.ImShow`, including
  interpolation option delegation, and gave `pyplot.py:function:imshow` its own
  public-surface row instead of relying on the broad pyplot module note.
- Added pyplot `Close` / `CloseAll` registry lifecycle helpers so stateful
  figures can be removed and cached managers closed deterministically, with
  public-surface notes for the implemented `pyplot.close` subset.
- Added pyplot `Draw` as a current-figure manager/canvas redraw wrapper,
  preferring draw-idle capable canvases and documenting the implemented
  `pyplot.draw` subset.
- Added pyplot `Ion` / `Ioff` / `IsInteractive` as package-level interactive
  mode state helpers with restore callbacks, and documented the implemented
  `ion` / `ioff` / `isinteractive` subset.
- Added pyplot `CLF` / `CLA` reset helpers for clearing the current figure or
  current axes without destroying the figure/axes registry entry, with
  public-surface notes for the implemented `clf` / `cla` subset.
- Added pyplot `Grid` / `TickParams` wrappers that create current-axes grid
  artists as needed and delegate typed tick/grid styling through
  `core.TickParams`, with explicit unsupported-axis errors.
- Added pyplot `Suptitle` / `SupXLabel` / `SupYLabel` wrappers over the
  current figure-level label layout path.
- Added pyplot `Box` as the stateful current-axes frame visibility toggle.
- Added pyplot `Axes` for stateful rectangle axes creation on the current
  figure, marking the new axes current.
- Added pyplot `Axis` for explicit `on` / `off` visibility and `equal` /
  `auto` aspect modes, returning errors for unsupported mode strings.
- Added pyplot `MinorTicksOn` / `MinorTicksOff` and `LocatorParams` wrappers
  over the current axes' typed minor-locator and tick-density controls.
- Added pyplot `XTicks` / `YTicks` wrappers that install fixed locators and
  optional fixed labels on the current axes with label-count validation.
- Split already-implemented pyplot basics (`gcf`, `gca`, `title`, `xlabel`,
  `ylabel`, `xlim`, `ylim`, `xscale`, `yscale`) into explicit public-surface
  rows with scoped remaining overload behavior.
- Added pyplot `SCA` / `DelAxes` wrappers for current-axes selection and
  registered axes removal, including errors for unregistered axes.
- Split implemented pyplot figure/subplot, show/pause, legend/colorbar,
  savefig, and rc helper functions into explicit public-surface rows with
  scoped remaining Matplotlib overload behavior.
- Added pyplot `Subplot2Grid` over the current figure's spanning GridSpec
  helper, marking the new axes current.
- Added pyplot `SubplotMosaic` over the current figure's named GridSpec mosaic
  helper, registering returned axes and preserving the first visible axes as
  current.
- Added pyplot `SubplotsAdjust` over the current figure's managed GridSpec
  subplot margin/spacing adjustment path.
- Added pyplot `FigLegend` over the current figure-level legend path.
- Added pyplot `TickLabelFormat` over current-axes `ScalarFormatter` style,
  scientific-limit, and mathtext controls, with explicit unsupported-axis,
  unsupported-style, and non-scalar-formatter errors.
- Documented `pyplot.getp`, `pyplot.setp`, and `pyplot.subplot_tool` as
  intentional omissions; Go uses typed artist APIs and layout controls instead
  of dynamic property strings or GUI-only adjustment dialogs.
- Added pyplot wrappers for existing core plot/stat/layout helpers: `Step`,
  `Stairs`, `BrokenBarH`, `BarLabel`, `BoxPlot`, `StackPlot`, `ECDF`,
  `AutoScale`, `FigText`, `TightLayout`, `TwinX`, and `TwinY`, with focused
  stateful delegation tests and public-surface rows.
- Added pyplot `FillBetweenX` and `Arrow` wrappers over existing core
  horizontal-fill and FancyArrow artists, with focused current-axes delegation
  coverage and public-surface rows.
- Added pyplot `HLines` / `VLines` wrappers that build current-axes
  data-coordinate `LineCollection` segment artists with focused delegation
  coverage and public-surface rows.
- Added pyplot manager/event wrappers for `GetCurrentFigManager`, `Connect`,
  `Disconnect`, and `DrawIfInteractive`, delegating through the existing
  figure-manager/canvas contracts with focused fake-manager coverage.
- Added a typed `SwitchBackend` subset that validates a named backend, clears
  cached managers, and recreates future pyplot managers through the selected
  backend, leaving Matplotlib's GUI side effects as documented partial scope.
- Added pyplot `GetCMap` over the existing color registry, with focused
  colormap lookup/fallback coverage and a public-surface row for
  `pyplot.get_cmap`.
- Added a tested `examples/pyplot_workflow` migration example that builds a
  stateful pyplot figure through `Subplots`, `SCA`, plot/scatter/bar/image
  wrappers, annotation, legends, colorbar, and figure-level labels.
- Added explicit public-surface classifications for the remaining pyplot
  dynamic/global shortcuts: numeric figure lookup helpers, current-mappable
  `clim` / `sci` / `set_cmap`, blocking GUI input waits, `clabel`, polar grid
  shortcuts, backend switching, and manager factory behavior now map to typed
  Go APIs, partials, or intentional omissions.
- Added a focused Phase 12.5 inventory test so those remaining pyplot
  dynamic/global shortcut classifications stay explicit.
- Added pyplot `ImRead` / `ImSave` wrappers over the core image IO helpers,
  with a focused PNG round-trip delegation test and explicit public-surface
  rows for `pyplot.imread` / `pyplot.imsave`.
- Added core `ImRead` / `ImSave` image IO helpers for Go-supported file
  decoding and PNG `render.RGBAImage` output, including straight-alpha PNG
  round-trip tests and explicit unsupported-input errors.
- Added explicit pyplot image-helper rows for `matshow`, `spy`, and the
  intentionally omitted `figimage`, so the migration status of the high-traffic
  image wrapper family is concrete.
- Added explicit public-surface classifications for every upstream
  `widgets.py` class. Supported Go widgets/selectors/cursors are mapped to
  their typed core/canvas equivalents, base/helper classes are documented as
  idiomatic equivalents, and GUI-only/internal helpers such as `LockDraw`,
  `SubplotTool`, `ToolHandles`, and `ToolLineHandles` are documented omissions
  with migration guidance.
- Added explicit public-surface classifications for every upstream
  `animation.py` row. Playback classes map to the current Animation /
  FuncAnimation / ArtistAnimation scheduler model, while movie writers,
  HTMLWriter, external encoder families, writer registries, and writer-only
  helpers are documented omissions until a deterministic writer surface is
  chosen.
- Tightened broad 12.5 widget and animation audit notes so cursor /
  multi-cursor support, deterministic widget callback ordering, disabled-state
  handling, and repeat-delay playback are described as existing tested scope
  rather than remaining work.
- Promoted `widgets_gallery` into the catalog/parity matrix as Phase 12.5
  widget coverage, tying the existing widget/selector/cursor showcase to
  golden, Matplotlib-reference, public-surface, feature-coverage, demo-breadth,
  and interactive-coverage inventories.
- Added an `animation_gallery` showcase with deterministic FuncAnimation and
  ArtistAnimation constructors, static preview rendering, focused example
  tests, and catalog/parity coverage for the Phase 12.5 animation playback
  surface.
- Added `Animation.Save` as an explicit unsupported-writer path returning
  `ErrWriterUnsupported`, with unit coverage. This keeps writer export out of
  v1.0 while avoiding silent no-op behavior for users who look for a save path.
- Tightened `Animation.Start` lifecycle error handling so a timer start failure
  returns the backend/event-loop error without leaving the animation marked
  running or holding a stale timer.
- Tightened `Animation.Stop` during repeat-delay windows so a later `Start`
  resumes with an immediate frame tick instead of inheriting a stale skip.
- Tightened repeat-delay timer rescheduling so a failure to restart the regular
  interval timer propagates and clears the running/stale-timer state.
- Tightened clipped AGG image compositing so transformed and scaled images keep
  premultiplied-alpha semantics when rendered through path masks, with focused
  coverage for transformed-image clipping plus image alpha and existing
  straight-alpha clipped path/pattern regressions.
- Fixed `ButtonOptions.Disabled` propagation so constructor-specified disabled
  state participates in the same interaction/picking behavior as runtime
  `Button.Enabled=false`.
- Tightened disabled slider interaction handling so focused `Slider` and
  `RangeSlider` widgets ignore keyboard nudges without firing callbacks or
  requesting redraws, matching the mouse-drag disabled path.
- Added disabled-state support for `CheckButtons` and `RadioButtons`, including
  constructor options, disabled visual dimming, and mouse/keyboard suppression
  without callbacks or redraw requests.
- Added widget composition guards for draw-idle callback wiring and animation
  coexistence: widget-layer picking still wins after an animation marks a
  high-z data artist animated, and non-widget events continue through user
  handlers.
- Tightened slider value snapping so `Slider` and `RangeSlider` initial values
  honor `ValueStep`, sort range endpoints after snapping, and keep rounded
  values inside the configured bounds.
- Tightened widget callback dispatch so callbacks fire in registration order
  after removals instead of depending on Go map iteration order.
- Added explicit public-surface classifications for the high-value backend
  lifecycle/tool rows: `_pylab_helpers.Gcf`, canvas/manager/event/timer
  classes, backend registry lookup/registration, and home/back/forward/pan/
  zoom/save tool classes and registry entries. Remaining backend scope is now
  centered on lower-value GUI presentation, cursor/status, configure/help/copy,
  and exact toolkit lifecycle edge behavior.
- Promoted the backend lifecycle checklist to covered status based on existing
  runtime/headless manager tests, pyplot fake-manager lifecycle tests, toolbar
  controller tests, WebAgg/Gio event-loop smoke coverage, and public-surface
  omission rows for GUI-only backend tools.
Implementation notes:

- Compare against upstream `image.py`, `pyplot.py`, `_pylab_helpers.py`,
  `backend_bases.py`, `backend_tools.py`, `widgets.py`, and `animation.py`.
- Any unsupported interpolation or widget path must produce a clear error or
  documented omission, never silently fall back to a wrong default.
- Keep pyplot, widget, and animation APIs thin over `core`, `canvas`, and
  backend contracts. If a wrapper needs bespoke behavior, first check whether
  the object-oriented API is missing the shared capability.
- Interactive and browser-facing examples belong in Phase 13 as user-facing
  breadth once the Phase 12.5 API and behavior decisions are settled.

**Exit criteria:**

- [ ] Every `GapDecisionImplement` row in `FoundationAPIGapAudit` is either
      implemented with catalog coverage or deliberately reclassified with a
      documented rationale.
- [ ] Every `partial` core feature row in `FeatureCoverageMatrix` has moved to
      `implemented`, `intentional-omission`, or a smaller remaining partial row
      with a precise scope.
- [ ] `go test ./test/...` parity failures caused by newly changed behavior are
      resolved by core fixes or fixture updates that match Matplotlib output.

---

# Phase 13: User-Facing Example Breadth

**Goal:** ensure every major implemented public feature family has a
user-facing Go example that demonstrates meaningful Matplotlib-equivalent
variants, not just a parity fixture.

### 18.1 Core Plot Family Galleries

- [ ] Add or expand examples for line/marker grids, advanced scatter, bar
      variants, fill variants, histogram variants, and multi-series legend
      behavior.
- [ ] Each example should live under `examples/<id>/` and have a matching
      `test/parity/<id>/plot.go`, `test/parity/<id>/plot.py`, and
      `test/matplotlib_ref/plots/<id>.py` entry when it represents parity
      behavior.
- [ ] Update `internal/examplecatalog.Case` rows so examples are discoverable
      through the catalog and golden/reference tests.

Implementation notes:

- Keep Go examples close to upstream Matplotlib examples. If output diverges,
  fix the core library first.
- Use `DemoBreadthGaps` as the checklist; do not close a gap until the demo
  includes the target features listed there.

### 18.2 Color, Image, Text, and Annotation Galleries

- [ ] Add named-color swatches and colormap family galleries.
- [ ] Add image interpolation/alpha/matshow/spy galleries.
- [ ] Add colorbar norm/extension galleries.
- [ ] Add MathText, text layout, annotation, legend, and offset-box galleries.

Implementation notes:

- Prefer several focused examples over one overloaded gallery when visual
  differences need inspection.
- Include captions/descriptions in catalog metadata explaining what feature
  breadth the example validates.

### 18.3 Toolkit, Projection, 3D, and Backend Output Galleries

- [ ] Add a broad mplot3d gallery covering 3D line, scatter, surface,
      wireframe, trisurf, bar3d, voxels, quiver3d, stem3d, and fill-between3d.
- [ ] Expand projection/toolkit galleries for geographic projections, radar,
      Skew-T, axisartist, and axes_grid1.
- [ ] Add mixed raster/vector output examples for SVG/PDF behavior with dense
      rasterized artists and vector text/axes.
- [ ] Add or expand triangulation galleries covering triplot, tripcolor,
      tricontour, tricontourf, and masked meshes.

Implementation notes:

- For 3D and projection examples, include both Python and Go sources even when
  the Go implementation is intentionally approximate.
- Backend-output examples should save and compare SVG/PDF artifacts, not just
  PNG screenshots.

**Exit criteria:**

- [ ] Every high-priority `DemoBreadthGap` is closed by a user-facing example
      or split into a precise implementation gap in Phase 9C.
- [ ] Every medium-priority `DemoBreadthGap` has either a user-facing example
      or a scheduled follow-up rationale.
- [ ] `docs/matplotlib-parity-status.md` reports no `fixture-only` example
      status for implemented public feature families unless it has an
      intentional reason.

---

# Phase 19: Browser Gallery Alignment

**Goal:** make the browser gallery a catalog-backed inspection surface for the
same feature families covered by parity fixtures and CLI examples.

### 19.1 Wire Planned Web Reference Modules

- [ ] Wire `test/matplotlib_ref/webdemos/annotations.py`, `bars.py`,
      `errorbars.py`, `fills.py`, `heatmap.py`, `histogram.py`, `lines.py`,
      `patches.py`, `scatter.py`, and `subplots.py` into active browser demos
      or fold them into existing catalog-backed browser families.
- [ ] Keep `radialforce.py` reference-only until it is promoted to a catalog
      case.
- [ ] Add tests that every active web reference module maps to a catalog case
      and every catalog-backed planned row either has an active browser demo or
      remains explicitly planned.

### 19.2 Promote CLI-Only Showcases

- [ ] Promote CLI-only showcases listed in `BrowserDemoCoverageRows` into
      browser demos: basic lines, dashes, scatter, bars, fills, errorbars,
      multi-series, histograms, boxplots, heatmaps, figure labels, colorbars,
      annotations, projections, mplot3d, triangulation, axisartist, and
      axes_grid1.
- [ ] Browser demos must use the same catalog factories as parity tests or a
      documented wrapper around them.
- [ ] Add browser-demo smoke tests that render each promoted demo and verify it
      has a non-empty image/artifact.

### 19.3 Browser Parity Status Reporting

- [ ] Update `docs/matplotlib-parity-status.md` with active/planned/reference
      browser status for each feature family.
- [ ] Fail CI if a `Showcase: true` catalog row has no browser accounting row.
- [ ] Fail CI if a browser demo references a feature family that is not present
      in the catalog.

**Exit criteria:**

- [ ] Every `BrowserDemoPlanned` row is active, intentionally reference-only,
      or tied to a later documented feature gap.
- [ ] The browser gallery can be used to visually inspect the major parity
      families without manually running parity tests.
- [ ] Browser demo coverage is generated from or checked against the catalog.

---

# Phase 20: Documentation, Examples Polish, and v1.0 Release

**Goal:** make the project consumable by users who have not been following
the development thread, and tag a stable v1.0.

### 20.1 API Documentation

- [ ] Package-level GoDoc passes for every public package, with a worked
      example per package.
- [ ] Hosted documentation site (pkg.go.dev plus a curated landing page
      under the existing GitHub Pages deployment).
- [x] Migration guide from upstream Matplotlib: side-by-side Python / Go
      snippets for every plot family covered by the catalog.
- [x] Backend selection guide: when to use AGG / GoBasic / SVG / PDF /
      Skia, with capability matrix excerpts (`docs/backend-selection.md`).

### 20.2 Examples Gallery Polish

- [x] Review every `Showcase: true` catalog row for caption, description,
      and runnable snippet quality.
- [x] Add an "anti-gallery" of intentional Matplotlib-divergence cases with
      the reasons documented (where the Go port chose different defaults).
- [x] Promote the WASM browser gallery to a first-class entry point on the
      project README.

### 20.3 Performance Pass

- [ ] Profiling sweep across the catalog: identify hotspots that exceed the
      100k-point smoothness goal and the sub-second typical-plot goal.
- [ ] Reusable benchmark suite under `benchmarks/` with regression tracking
      in CI.
- [ ] Documented memory-usage targets and a tuning guide for long-running
      applications.

### 20.4 Release Readiness

- [ ] Semantic version policy decision and `CHANGELOG.md` baseline.
- [ ] Final golden / reference regeneration pass with explicit per-case
      tolerances frozen for v1.0.
- [ ] Public API stability audit: identify and either rename or hide any
      symbol that is not intended to be part of the v1.0 surface.
- [ ] CI gate: `just fmt && just lint && just test` plus catalog-driven
      parity checks must all pass on the release branch.
- [ ] Tag v1.0.

**Exit criteria:**

- [ ] A new user can install the module, follow the documentation, and
      reproduce every showcase plot.
- [ ] The public API surface is documented, audited, and frozen for v1.0.
- [ ] Performance and parity baselines are tracked in CI.

---

# Development Guidelines

## Backend Strategy

- **Primary raster backend:** AGG (`backends/agg/`) — anti-aliased,
  sub-pixel accurate, reference for parity fixtures.
- **AGG port ownership:** if a parity failure is caused by a fundamental
  rasterization, text, path, transform, or blending issue in the Go AGG port,
  fix `../agg_go` rather than adding compensating behavior in this repository.
- **Pure-Go fallback:** GoBasic (`backends/gobasic/`) — dependency-light
  correctness fallback.
- **Primary vector backend:** SVG (`backends/svg/`) — deterministic,
  browser-readable, structurally tested.
- **Publication vector backends:** PDF / PS / PGF (Phase 1).
- **Accelerated raster backend:** Skia (`backends/skia/`) — opt-in CPU and
  future GPU paths.

## Testing Strategy

- Catalog-driven parity tests (`internal/examplecatalog.Case` + `test/`).
- Golden image tests for raster backends, structural diff for vector
  backends.
- Property-based tests for data ranges and transforms.
- Visual regression against Matplotlib references with documented
  per-case tolerances.
- `go test ./...` runs the full suite; `go test ./test/ -run <id>` runs
  one parity case.

## API Design Principles

- Follow Matplotlib conventions where sensible; document and explain
  divergences.
- Use functional options for configuration; keep zero-value defaults
  useful.
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
- Examples demonstrate real-world usage rather than minimal API smoke
  tests.

---

This roadmap reflects the work remaining to bring matplotlib-go to a
stable, documented v1.0 release. Phases 1-3 close functional gaps in
output formats, effects, and math typography; Phases 4-6 add the
interactive runtime that the headless event infrastructure has been
waiting for; Phase 7 hardens the backend matrix; Phase 8 finishes the
release.
