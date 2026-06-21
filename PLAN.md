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

_Former Phases 1–6, 8A, 9–19 are complete; see git history for the detailed
per-phase implementation logs._

---

# Phase 1: Backend Deepening (Skia Native + GPU)

**Goal:** finish the backend-specific Skia work. The historical blocker (no Skia
C-ABI binding) is **resolved**: a real Skia library is built locally and a narrow
C-ABI wrapper links it under `-tags "skia skiacgo"`. The remaining work is wiring
GPU mode — no longer external-access-blocked, just unbuilt.

## Done (summary)

C-ABI Skia wrapper + cgo bridge (`backends/skia/skia_cwrap.{h,cpp}`,
`native_cgo.go`; tag `skiacgo`) links Skia milestone 151. Native path collections,
quad meshes, transformed images, tiled-shader hatching, gradient fills, marker
batches, and Gouraud triangles are implemented via the wrapper. AGG non-text
parity diagnostics (`TestNonTextResidualDiagnostics`) are in place.

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

## Remaining work

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
      output and truthful per-mode capability reporting.

---

# Phase 2: Visual Parity Closure via Code Parity ✅

**Closed 2026-06-14** (full `TestReferenceCompare` sweep). All catalog rows are
below `RMSE 5` against the Matplotlib 3.10.9 reference set. High-residual
families closed via waves W1–W5: mplot3d, projections, axisartist, text
wrapping/rotated layout, MathText and annotation tails, legend/offsetbox, fills
and collections, contour/mesh, arrays, widgets, and mixed raster/vector. A
second sweep on 2026-06-20 closed the remaining above-5 rows (overlapping-axes
draw order, fill snap, scatter markers, polar/geo label bounds, `MaxNLocator`
edge labels, boxplot markers/alpha, matshow tick positions, and path-effect
offset/alpha). Catalog tolerances are ratcheted to actual metrics plus small
headroom. The one open follow-up (`stat_variants`, RMSE 3.24, cap `MaxRMSE 3.4`)
is below threshold with a documented rasterization-boundary exception.

---

# Phase 3: Parity Status Reporting ✅

**Closed.** `docs/matplotlib-parity-status.md` is generated from machine
inventories and is kept current by CI: `TestAllUpstreamPublicRowsAreClassified`
guards unclassified rows; `TestPartialAndOmissionRowsHaveNotes` guards rows
without rationale. Per-family status tables cover every upstream feature family
with status, local API, fixture, and browser demo columns.

---

# Phase 4: Large File Decomposition ✅

**Closed.** All tracked Go files above 1k lines are split into focused units
(or documented with an explicit keep-large decision in
`docs/large-file-decomposition.md`). Decomposed in batches L1–L8: test files
by behavior family, algorithm-heavy core files (`contour`, `axis`, `text`,
`axes3d_contour_surface`, `legend`, `colorbar`), facade entrypoints (`pyplot`,
`cmd/parityviewer`), and backend implementations (`agg`, `gobasic`, `ps`,
`pgf`). Full `just fmt && just lint && just test` and optional parity checks
pass after decomposition.

---

# Phase 5: Documentation, Performance, and v1.0 Release

**Goal:** make the project consumable by users who have not followed the
development thread, establish performance baselines, and tag a stable v1.0.

### 5.1 API Documentation ✅

Package-level GoDoc with worked examples for every stable public package;
guarded by `TestStablePublicPackagesHaveGoDocAndExamples`. Hosted docs at
pkg.go.dev; WASM gallery landing page links to README, examples, backend
selection, migration notes, and parity status.

### 5.2 Performance Pass ✅

Benchmarking harness (`benchmarks/`, `just bench-render`, `just profile-render`),
CI benchmark report (`.github/workflows/benchmark-report.yml`), and profiles
documented in `docs/performance-profiling.md`. Optimized: AGG FreeType text
measurement caching, font-face rune-support memoization, shaping buffer reuse,
100k scatter allocation pressure (combined affine path, marker prototype fast
path, devPath reuse), surface/image-copy reuse, and scalar-map setup caching.
User-facing memory targets and tuning guide documented.

### 5.3 Release Readiness

- [ ] Semantic-version policy decision and `CHANGELOG.md` baseline.
- [ ] Final golden / reference regeneration pass with per-case tolerances frozen
      for v1.0.
- [x] Public API stability audit: stable exported Go API frozen in
      `test/testdata/public_api/stable_public_api.json`; CI audit test guards
      against accidental surface drift; geometry primitives promoted from
      `internal/geom` to the public `geom` package.
- [ ] CI gate: `just fmt && just lint && just test` plus catalog-driven parity
      checks all pass on the release branch.
- [ ] Tag v1.0.

**Exit criteria:**

- [ ] A new user can install the module, follow the docs, and reproduce every
      showcase plot.
- [ ] The public API surface is documented, audited, and frozen for v1.0.
- [ ] Performance and parity baselines are tracked in CI.

---

# Phases 6–11: Parity-Breadth Closure

These phases are derived from the independent fidelity review in
[`REVIEW.md`](REVIEW.md) (2026-06-21). The review's verdict: the numerical and
rendering **core is faithful** (transforms, locators, contours, 3D projection,
norms, mathtext layout, FreeType text — verified line-for-line), but **breadth
is incomplete** — the long tail of Matplotlib kwargs/modes is missing across
nearly every subsystem, and several degradations are **silent**. The generated
`docs/matplotlib-parity-status.md` self-classifies most families as
"partial / thin"; these phases close that gap.

**Relationship to v1.0:** Phase 5 ships the release *mechanics*. Phases 6–10
close the *parity* the project advertises and should land before a v1.0 that
claims Matplotlib parity, **or** the parity claim must be explicitly scoped in
docs ("the common path matches" vs "the full API is present"). Phase 11 is
deferred infrastructure depth. Ordered by impact (Phase 6 first).

## Phase 6: Silent-Failure Hardening & API Correctness

**Goal:** eliminate the worst UX failure mode — silently-wrong output with no
diagnostic. These are cheap, high-value, and gate trust in everything else.

Foundation: a shared, swappable warning sink lives in `internal/diag`
(`diag.Warnf` / `diag.SetHandler`) so low-level (`color`) and high-level (`core`)
packages can surface non-fatal problems without coupling or enlarging the
public API.

- [x] **MathText** warns on unknown commands instead of silently echoing them.
      The engine reports each unrecognized command through
      `mathtext.SetUnknownCommandHandler` (`../mathtext/unknown_command.go`,
      called from `parser.go` default case); `core` wires it to `diag.Warnf`,
      deduplicated per command name (`core/mathtext_warn.go`). Shipped in
      `github.com/cwbudde/mathtext v0.2.0` (no `replace` directive). _Follow-up:
      a strict (error) mode toggle._
- [x] **`alpha=0` vs unset (Bar + Fill family):** an explicit alpha (including
      0 for fully transparent) is baked into the resolved colors at construction
      via `bakeExplicitAlpha`, so it is honored while a nil alpha preserves the
      color's own channel (`core/plot.go` — `Bar`, `Fill`, `FillBetween`,
      `FillToBaselinePlot`, `FillBetweenX`, plus `Hist`, `ErrorBar`, `BoxPlot`,
      and `BoxPlots`). `Step` (delegates to `Plot`) and the 3D bar/voxel paths
      (`Axes3D.Bar`, `Bar3D`, `Voxels`) already multiplied the explicit alpha
      into their colors at construction, so they honor 0 too — locked in by
      regression tests in `core/alpha_zero_artists_test.go`. _Remaining:
      `Violin` and `Grid` expose a plain `float64` alpha option where 0 is
      structurally indistinguishable from "unset"; honoring an explicit 0 there
      needs a deliberate API change to `*float64` (frozen public API)._
- [x] **Unknown colormap** warns (via `diag.Warnf`) before falling back to
      viridis (`color/colormap.go` `GetColormap`). _Follow-up: case-sensitive
      lookup is deferred — the registry stores lowercase keys, so strict casing
      risks regressing existing callers; revisit with a rename table._
- [x] **Gouraud QuadMesh** warns once (per artist) when downgrading to flat
      cell colors because the renderer lacks `GouraudTriangleDrawer`
      (`core/collection_quadmesh.go`).
- [x] **Invalid artist input** (length mismatch, etc.) logs a reason via
      `diag.Warnf` instead of returning a bare `nil` artist. Wired into the
      primary plotting API: `Scatter` (x/y size), `Hist` (weights length),
      `FillBetween`/`FillBetweenX` (`where` length), `ErrorBar` (error/limit
      array lengths and invalid `errorevery`). The signatures stay unchanged
      (still return `nil`), so this is non-breaking; the same one-liner extends
      to any other silent-drop path as needed.
- [x] **3D naming traps:** `PlotSurface` (a 1D line strip, `core/axes3d.go`)
      and `Voxel` (edges-only, `axes3d_bar_voxel.go`) now emit a one-shot
      per-axes `diag.Warnf` and carry honest doc comments pointing at the real
      APIs — [`Axes3D.Surface(x, y, z)`](core/axes3d_surface.go) for a true
      surface and [`Axes3D.Voxels(grid)`](core/axes3d_bar_voxel.go) for filled
      cubes. Their signatures/return types (`*Line2D`, `*LineCollection`) are
      incompatible with the real functions (`*PolyCollection`, a collection
      map), so a clean rename needs a deliberate API break; until then the trap
      is no longer silent. _Follow-up: deprecate/rename in a future major
      version._

**Exit criterion:** no core artist silently discards user intent; every
unsupported input path produces a diagnostic. _Status: all 6 items done —
mathtext; alpha for Bar/Fill/Hist/ErrorBar/BoxPlot (Step/3D already correct);
colormap; Gouraud; invalid-input diagnostics on the primary plotting API; and
the 3D naming traps now warn + document the real APIs. Remaining tails:
`Violin`/`Grid` alpha need a `*float64` API change, and a mathtext strict-mode
toggle is still open._

## Phase 7: Formatter, Layout & Date Fidelity ✅

**Goal:** close the highest-impact divergences on ordinary linear plots and
layouts — defaults a typical Matplotlib script relies on.

- [x] **`ScalarFormatter` offset / ×10ⁿ multiplier text** rendered by default on
      both axes. `ScalarFormatter` now ports Matplotlib's
      `set_locs`/`_compute_offset`/`_set_order_of_magnitude`/`_set_format`/
      `get_offset` (`core/tick_formatters.go`): it factors a shared additive
      offset and order-of-magnitude into axis offset text (`OffsetText`) and
      renders ticks at uniform precision. The X-only guard at
      `axis_ticklabels.go` is dropped and left/right (Y-axis) offset positioning
      added. The `formatter_scalar_scientific_labels` fixture + reference were
      retargeted from a `FuncFormatter` workaround to the real `ScalarFormatter`
      (Go output now byte-matches Matplotlib's offset rendering).
- [x] **Real `constrained_layout` solver** — a `LayoutGrid` constraint
      propagation solver (`core/layoutgrid.go`, ported from
      `_layoutgrid.py`/`_constrained_layout.py`) with per-margin variables,
      a twice-run solve, and suptitle/colorbar/legend reservations. Only
      `LayoutEngineConstrained` routes through it; `tight_layout` keeps its
      greedy heuristic, so the two engines now genuinely differ.
- [x] **Date convention:** adopted Matplotlib days-since-epoch and shipped
      `Date2Num`/`Num2Date`/`SetEpoch`/`GetEpoch` (`core/dates.go`); the internal
      converters in `core/date_tick.go` use days with microsecond rounding (like
      `num2date`), and `internal/parityutil` reuses `Date2Num` so fixtures agree.
- [x] **Per-axes margins:** `Margins`/`SetXMargin`/`SetYMargin` and
      `SetAutolimitMode("round_numbers")` snap-to-locator rounding
      (`core/axes_autoscale.go`).
- [x] **Aspect:** `SetAdjustable("box"|"datalim")`, `SetAnchor` (cardinal
      anchors), anchor-aware box positioning, and a `datalim` data-limit
      expansion applied from autoscale (`core/axes_limits.go`).
- [x] **Axis-length-aware bins:** `MaxNLocator`/`AutoLocator` opt-in
      `NbinsAuto` clamps the axis-length-aware target count to
      `[max(1,minTicks-1), 9]` (`tick_locators.go`); date-locator `nonsingular`
      (~4-year) expansion in `DateLocator` (`date_tick.go`).
- [x] **`label_outer()`** + inner shared-axes tick-label / offset / axis-label
      suppression via per-Axes flags consulted at draw time (`core/gridspec.go`,
      `figure_draw.go`, `axes_labels.go`) — non-mutating so shared `*Axis`
      artists are unaffected.

**Exit criterion:** ✅ large/offset-magnitude linear data and date figures match
the Matplotlib references at RMSE ≤ ~2; the two layout engines diverge as
intended. (Pre-existing, unrelated branch drift remains in some SVG goldens and
a few non-Phase-7 reference-compare rows.)

## Phase 8: MathText & Text Completeness

**Goal:** raise mathtext from ~13% symbol coverage to real-world usability and
fix text fallback.

- [ ] **Expand the `tex2uni` symbol table** toward the full 632 entries
      (arrows, relations, binary ops, `\cdots`/`\vdots`/`\ddots`, `\var*` Greek)
      (`mathtext .../normalize_tables.go:50`).
- [ ] **Math alphabets** `\mathbb \mathcal \mathfrak \mathscr \boldsymbol \bm`
      mapped to the Unicode Mathematical Alphanumeric block.
- [ ] **Accent model:** centered separate-glyph accents over the nucleus;
      add `\widehat \widetilde \overbrace \underbrace \overline`-as-rule
      `\stackrel \substack \overset \underset \not`.
- [ ] **Per-glyph multi-font fallback** — walk the family list per missing glyph
      instead of one font per family (`render/font_manager.go:757`); no tofu.
- [ ] **Bridge `Text(bbox=boxstyle=…)`** to the existing `FancyBboxPatch`
      styles (sawtooth/arrow/circle) (`core/text_bbox.go`, `patch_fancybbox.go`).
- [ ] Track mathtext coverage with a symbol-table parity test.

**Exit criterion:** common Matplotlib labels render without literal-echo;
symbol coverage is measured and reported.

## Phase 9: Plot, Colormap & Norm Configuration Breadth

**Goal:** close the per-artist configuration tail so common scientific plots
match Matplotlib defaults, not just the happy path.

- [ ] **Boxplot:** `patch_artist=False` unfilled default, orientation, and
      `showbox`/`showcaps`/`showmeans`/`meanline`/`sym` on the high-level artist
      (`core/boxplot.go:558`); honor `bootstrap` CI; percentile-*value* whisker
      semantics.
- [ ] **StackPlot** `wiggle`/`weighted_wiggle`/`sym` baselines
      (`core/stat_variants.go:48`).
- [ ] **Contour** `negative_linestyles` (default dashing), `extend`,
      `linestyles`, and contourf `hatches` (`core/contour_api.go`); `clabel`
      `fmt` (dict/callable/format-string) + `rightside_up`.
- [ ] **Colorbar** norm-aware locators for SymLog/Power/TwoSlope/Centered +
      `NoNorm` IndexLocator (`core/colorbar_scale.go:118`); `extendfrac`; minor
      ticks.
- [ ] **Image:** native RGBA `imshow` for `(M,N,3/4)` arrays, image `aspect`,
      and image normalization (`core/image.go:97`).
- [ ] **Norms/cmaps:** `FuncNorm`, `MultiNorm`, `petroff10` colormap.
- [ ] **Misc artist kwargs:** `Stem` orientation, errorbar `capthick`, scatter
      `plotnonfinite`, `LineCollection` linestyle-string → dash conversion.

**Exit criterion:** contour, colorbar, boxplot, and image cases match
Matplotlib default styling.

## Phase 10: Backend, Renderer & Styling Completion

**Goal:** finish the renderer/backend semantics and grow the styling system from
its current ~13% rcParam coverage.

- [ ] **Vector text metrics:** replace the crude `MeasureText` stubs in PDF/PS/
      PGF/SVG with the shared FreeType font manager (`backends/pdf/pdf_text.go:46`,
      `backends/svg/text.go:53`, etc.) so rotated/vertical anchoring matches AGG.
- [ ] **Antialiased Gouraud** via agg_go `span_gouraud_rgba`
      (`backends/agg/agg_gouraud.go:10`); **>3-stop gradients**
      (`gradients.go:34`).
- [ ] **Sketch/xkcd** rendering pass — currently a no-op despite full contract
      plumbing (`render/render.go:277`).
- [ ] **PS/PGF** gradient + pattern fills; PGF clip-path and vertical-text
      interfaces.
- [ ] **`url`/`gid` metadata** in `GraphicsContext` for clickable vector output;
      finish `RestoreRegion` y-flip (`backends/agg/agg.go:360`) for blit/anim.
- [ ] **rcParams coverage** for `savefig.*`, `pdf.*`/`ps.*`/`svg.*`,
      `animation.*`, `boxplot.*`, `mathtext.*`, `hatch.*`, `image.*`, `date.*`
      (`style/mplstyle.go:31`).
- [ ] **Cyclers** for `linestyle`/`marker`/`linewidth`; bundle the common
      `.mplstyle` sheets (`seaborn-*`, `fivethirtyeight`, `bmh`,
      `Solarize_Light2`) (`style/theme.go:23`).

**Exit criterion:** vector text parity with AGG, antialiased Gouraud, and broad
rcParam + stylesheet coverage.

## Phase 11: Deferred Infrastructure Depth ⚠️

**Goal:** lower-priority structural depth surfaced by the review; not required
for the v1.0 parity claim but worth tracking so it does not vanish into "future
work."

- [ ] **Bézier toolkit** (`bezier.py` equivalent): `split_bezier`, arc-length,
      offset/parallel curves, `inside_circle` — used by fancy arrows and
      annotation connectors.
- [ ] **Path-generator helpers** in `geom`: `unit_circle`, `arc`, `wedge`,
      4-cubic circle approximation (currently rebuilt ad hoc in `core/`).
- [ ] **Triangulation library** (`tri/`: Delaunay, `TriFinder`,
      `TriInterpolator`) instead of per-call implementations.
- [ ] **Live bbox-linked transforms** (`BboxTransformTo`) so axes resize
      invalidates rather than rebuilds the transform graph.
- [ ] **Open transform type set:** a `get_affine()` capability interface so a
      third-party `T` can participate in flattening (`transform/transform.go:32`).
- [ ] **Exploit the affine/non-affine cache split** in `TransformedPath`
      (declared but unused, `transform/transformed_path.go:13`).
- [ ] **Path simplifier:** replace Douglas–Peucker with Matplotlib's single-pass
      running-segment algorithm for pixel parity on dense lines
      (`backends/agg/agg_path_simplify.go:5`).
- [ ] Teardown API: `Axes.cla()`/`clear()`/`remove()`, `Figure.delaxes`/`clf`;
      `setp`/`getp`/`findobj` introspection.

**Exit criterion:** geometry primitives and the transform graph reach
Matplotlib's structural flexibility; no per-call reimplementations of shared
infrastructure.

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
documented v1.0 release. **Phase 1** finalizes GPU acceleration for the Skia
backend (CPU native primitives are done); **Phases 2–4** are closed. **Phase 5**
is in the final stretch — documentation and performance work is done; only the
release mechanics (changelog, CI gate, v1.0 tag) remain. **Phases 6–11** are the
parity-breadth closure derived from [`REVIEW.md`](REVIEW.md): Phase 6 hardens
silent-failure modes, Phases 7–10 close the missing Matplotlib configuration
breadth that gates the parity claim, and Phase 11 tracks deferred infrastructure
depth. They are ordered by impact and should land before — or explicitly scope —
a v1.0 that advertises Matplotlib parity.
