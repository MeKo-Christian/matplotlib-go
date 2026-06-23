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

**Completed-batch ledger** (guards the split via the `*SplitIsTracked` tests):

- [x] **L1 — Add a repeatable large-file audit.** `just large-file-audit` plus
      `docs/large-file-decomposition.md` capture the baseline inventory and rules.
- [x] `core/contour.go`: extract public API and `ContourSet` construction into `core/contour_api.go`.
- [x] `core/contour.go`: extract coordinate normalization, triangulation, and level selection into `core/contour_levels.go`.
- [x] `core/contour.go`: extract line segment generation, stitching, structured boundary handling, and polyline rotation into `core/contour_lines.go`.
- [x] `core/contour.go`: extract filled band polygon clipping, saddle handling, and band colors into `core/contour_filled.go`.
- [x] `core/contour.go`: extract clabel placement, inline erasing, label angle, and label width helpers into `core/contour_labels.go`.
- [x] `core/axis.go`: extract axis side/type definitions and constructors into `core/axis_types.go`.
- [x] `core/axis.go`: extract spine/frame drawing, snapping, and spine position helpers into `core/axis_spine.go`.
- [x] `core/axis.go`: extract major/minor tick drawing, tick target counts, and tick styling into `core/axis_ticks.go`.
- [x] `core/axis.go`: extract tick-label drawing, offset text, label bounds, and alignment helpers into `core/axis_ticklabels.go`.
- [x] `core/axis.go`: extract polar spine/tick/tick-label behavior into `core/axis_polar.go`.
- [x] **L7 — Decide generated-data strategy.** Keep-large curated catalog and
      generated tables are documented with drift guards in `docs/large-file-decomposition.md`.

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

**Completed P2 work** (guarded by the `TestPerformanceP2*` doc tests):

- [x] **P2 — Surface and image-copy reuse for repeated renders.**
  - [x] Add a benchmark that redraws the same figure into a reused renderer.
  - [x] Document or expose a supported renderer-reuse path (`agg.Renderer.Clear` + `agg.Renderer.ImageView`).
  - [x] Avoid `GetImage` copies in benchmark/save paths via the non-owning image view.
- [x] **P2 — Cache scalar mapping setup.**
  - [x] Cache resolved colormap and norm state on scalar-mapped artists (`ScalarMapInfo.Resolved`).
  - [x] Add focused benchmarks for scalar-mapped image, scatter, and mesh rows.
- [x] **Memory targets and tuning guide.**
  - [x] Define v1.0 memory targets for typical catalog plots, 100k scatter, and repeated-redraw and scalar-mapping scenarios.
  - [x] Document practical tuning advice: renderer reuse, avoiding unnecessary `GetImage` copies, marker batching, and backend selection.

### 5.3 Release Readiness

- [ ] Semantic-version policy decision and `CHANGELOG.md` baseline.
- [ ] Final golden / reference regeneration pass with per-case tolerances frozen
      for v1.0.
- [x] Public API stability audit: stable exported Go API frozen in
      `test/testdata/public_api/stable_public_api.json`; CI audit test guards
      against accidental surface drift; geometry primitives promoted from
      `internal/geom` to the public `geom` package.
- [ ] CI gate: `just fmt && just lint && just test` plus catalog-driven parity
      checks all pass on the release branch. The GitHub Actions workflows now
      build (cgo system deps + vendored FreeType) and the `build`/`vet`/`fmt`/
      `lint` jobs are green; the remaining `go test ./...` failures (pre-existing,
      never CI-validated) are catalogued in
      [`docs/ci-known-test-failures.md`](docs/ci-known-test-failures.md).
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

**Relationship to v1.0:** Phase 5 ships the release _mechanics_. Phases 6–10
close the _parity_ the project advertises and should land before a v1.0 that
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

The **core completeness wave** (symbols, alphabets, per-glyph fallback, coverage
test) is **done** in `github.com/cwbudde/mathtext v0.3.0` (no `replace`); the
accent overhaul and the `Text(bbox=…)` bridge remain as a fast-follow.

- [x] **Expand the `tex2uni` symbol table** to the full 632 entries (arrows,
      relations, binary ops, `\cdots`/`\vdots`/`\ddots`, `\var*` Greek). The
      complete matplotlib table is generated into `mathtext/tex_tables.go` and
      consulted as a final fallback in both the layout parser
      (`parser.go:parseCommandNode`) and the plain-text normalizer
      (`normalize.go:parseCommand`), so existing hand-tuned maps still win; the
      binary-operator/relation/arrow classes (`tex_spacing.go`) drive operator
      spacing.
- [x] **Math alphabets** `\mathbb \mathcal \mathfrak \mathscr \boldsymbol \bm`
      mapped to the Unicode Mathematical Alphanumeric block with the
      Letterlike-Symbols reserved holes (`mathtext/alphabets.go`).
- [ ] **Accent model:** centered separate-glyph accents over the nucleus;
      add `\widehat \widetilde \overbrace \underbrace \overline`-as-rule
      `\stackrel \substack \overset \underset \not`. _(deferred fast-follow)_
- [x] **Per-glyph multi-font fallback** — `render/text_fallback.go`
      `ResolveTextRuns` walks the requested family list, then generics, then
      `STIXGeneral` per missing glyph, so Mathematical-Alphanumeric/symbol
      glyphs DejaVu lacks resolve to STIX instead of tofu.
- [ ] **Bridge `Text(bbox=boxstyle=…)`** to the existing `FancyBboxPatch`
      styles (sawtooth/arrow/circle) (`core/text_bbox.go`, `patch_fancybbox.go`).
      _(deferred fast-follow)_
- [x] Track mathtext coverage with a symbol-table parity test
      (`mathtext/coverage_test.go` against the vendored
      `testdata/tex2uni_symbols.json`; currently 100%).

**Exit criterion:** common Matplotlib labels render without literal-echo;
symbol coverage is measured and reported. _Status: met for the core wave —
632/632 symbols + the six math alphabets resolve to glyphs; coverage is asserted
≥95% (100% today). Accents and the bbox→FancyBboxPatch bridge are the remaining
fast-follow items._

## Phase 9: Plot, Colormap & Norm Configuration Breadth

**Goal:** close the per-artist configuration tail so common scientific plots
match Matplotlib defaults, not just the happy path.

- [x] **Boxplot:** `patch_artist=False` unfilled default (no color-cycle
      consumption), orientation, and `showbox`/`showcaps`/`showmeans`/`meanline`/
      `sym` on the high-level artist (`core/boxplot.go`, `core/plot.go`); honor
      `bootstrap` CI; configurable scalar `whis` + percentile-value whisker
      semantics (Q1/Q3-clamped per `cbook.boxplot_stats`). New `boxplot_default`
      parity case covers the unfilled default styling; orientation/means/bootstrap
      covered by `core/boxplot_test.go`. Caveat: `bootstrap` uses a seeded RNG
      (Matplotlib's is unseeded, so no byte-exact parity image); `sym` parsing is
      a minimal color+marker shorthand (structured options remain the primary
      surface).
- [x] **StackPlot** `wiggle`/`weighted_wiggle`/`sym` baselines
      (`core/stat_variants.go`): `StackPlotOptions.BaselineMode` selects the
      `StackBaselineZero`/`Sym`/`Wiggle`/`WeightedWiggle` layouts as a faithful
      port of Matplotlib's `stackplot.py` baselines. Color cycling now consumes
      exactly one property-cycle entry per layer (previously two), matching
      Matplotlib. New `stackplot_streamgraph` parity case covers the
      `weighted_wiggle` (streamgraph) layout with default colors; the baseline
      math is unit-tested in `core/stat_variants_test.go`. `hatch` list cycling
      and `sticky_edges` remain out of scope. Also fixed an AGG-backend bug
      (`backends/agg/agg_draw.go`): the single-path-collection half-pixel
      placement offset (`translatePath(0.5, -0.5)`) was applied to _unstroked_
      fills, shifting `fill_between`/`stackplot` polygons ~0.5px off Matplotlib;
      it is now gated on a visible stroke, dropping streamgraph parity from
      RMSE 6.15 → 0.07 with no change to stroked fills.
- [x] **Contour** `negative_linestyles` (default dashing), `extend`,
      `linestyles`, and contourf `hatches` (`core/contour_api.go`); `clabel`
      `fmt` (dict/callable/format-string) + `rightside_up`. `ContourOptions`
      gained `LineStyles`/`NegativeLineStyles` (port of `_process_linestyles`:
      monochrome contours dash negative levels by default), `Extend`
      (`contourExtendedLevels` sentinel bands → colormap under/over via
      `AtValue`), and `Hatches` (cycled per band; hatched contourf merges each
      band into one compound path so the hatch tiles continuously, matching
      Matplotlib's one-path-per-level model). `ClabelOptions` gained
      `FormatString`/`FormatDict` (`get_text` port) and `RightSideUp`. Backend
      fix: AGG hatch strokes are now always anti-aliased independent of the
      fill's AA (contourf bands are `antialiased=False` but their hatch lines
      are AA in Matplotlib) — dropped hatched-contourf reference RMSE 26→0.55.
      Also ported contourpy's closed-loop start-vertex convention
      (`rotateClosedLoopToContourpyStart`, `core/contour_lines.go`): Go's marching
      squares produced the same vertices/winding as mpl2014 but rotated to a
      different start, throwing dashed-contour dash phase ~anti-phase (RMSE 9);
      aligning the start (bottom-boundary-tangent loops at their leftmost boundary
      vertex; interior loops at the leftmost vertical-edge crossing in the
      bottommost row-band) drops dashed parity to RMSE 0.07 with no regression
      (mesh*contour_tri stays 2.32). New `contour_styles` parity case (RMSE 2.50,
      negative dashing + `%.2f` labels + extend + hatches); style resolution,
      extend, hatch cycling, label formatting, and the loop-start rule unit-tested
      (`core/contour_styles_test.go`, `core/contour_label_format_test.go`,
      `core/contour_loopstart_test.go`). \_Deferred: full contourpy `locate_label`
      port (Go's inline auto-label placement still differs on asymmetric loops;
      manual placement matches); colorbar extend triangles, log-scale extend,
      accent/`format_ticks` list semantics.*
- [x] **Colorbar** norm-aware locators for SymLog/Power/TwoSlope/Centered +
      `NoNorm` IndexLocator (`core/colorbar_scale.go`); `extendfrac`; minor
      ticks. Refactored `configureColorbarScale`/`configureHorizontalColorbarScale`
      into a shared `applyColorbarNormScale` (via `colorbarAxisOps`) and added
      per-norm cases: `SymLogNorm` → symlog scale + `SymmetricalLogLocator`
      (major+minor, shown by default like log); `PowerNorm`/`TwoSlopeNorm`/
      `CenteredNorm` → `FuncScale` with `AutoLocator` (mpl's function-scale
      default, nice data-space ticks, replacing `LinearLocator`); `NoNorm` →
      `IndexLocator{base: 1+⌊N/10⌋, offset: 0.5}`. `extendfrac` lands as
      `ColorbarOptions.ExtendFrac` (scalar or per-side `[min,max]`) +
      `ExtendFracAuto` (`'auto'`), threading per-side fractions through the
      body inset, slot shrink, extension patches, and outline (replacing the
      hard-coded 5%). Minor ticks: log/symlog/asinh show their scale minor
      ticks by default; linear/function/boundary are opt-in via
      `ColorbarOptions.MinorTicks` (`AutoMinorLocator`/boundary `FixedLocator`).
      New parity cases `colorbar_symlog_ticks` (RMSE 0.65) and
      `colorbar_extendfrac` (RMSE 0.87); TwoSlope gallery ticks now match mpl.
      _Deferred: gradient-`NoNorm`-without-values auto extendfrac (5% fallback);
      `NoNorm` value-count ambiguity makes it unit-test-only._
- [x] **Image:** native RGB/RGBA `imshow` for `(M,N,3/4)` arrays plus Go
      `image.Image` input, bypassing colormap+norm. New `Axes.ImShowRGB`
      (float `[0,1]` arrays) and `Axes.ImShowImage` (`image.Image`, e.g.
      `ImRead` output) with `ImShowRGBOptions` (`core/matrix_helpers.go`); the
      array path ports matplotlib's `_normalize_image_array` clip/dtype handling
      (`core/image_rgba.go`), `(M,N,1)` squeezes to the scalar colormap path.
      Pre-colored pixels ride a new `Image2D.rgba` field and a true-color
      rasterize branch (`core/image.go`, `core/image_api.go`); origin flip and
      per-pixel alpha (multiplied by the scalar `Alpha`) are preserved. Image
      `aspect` and scalar `norm` were already shipped via `ImShow`/`MatShow`
      (default `aspect="equal"`, full `norm` family). New `imshow_rgb` parity
      case (RMSE 0.23) plus `core/image_rgba_test.go` cover RGB/RGBA
      classification, clipping, origin, and alpha.
- [x] **Norms/cmaps:** `FuncNorm` and the `petroff10` color sequence are
      implemented; `MultiNorm` is deferred. `core.FuncNorm` (`core/norm.go`)
      ports Matplotlib's scale-backed forward/inverse normalizer: `Forward`/
      `Reverse` callbacks plus `VMin`/`VMax`/`Clip`, normalizing
      `Forward(value)` between `Forward(VMin)` and `Forward(VMax)` (clip clamps
      in data space, then transforms), with inverse, transform-domain-finite
      autoscale, and validation; unit-tested in `core/funcnorm_test.go`. No
      transpiler-backed golden exists because arbitrary Go callbacks do not
      transpile to a Matplotlib reference script. `color.Petroff10`
      (`color/petroff.go`) adds the ten-color `petroff10` sequence
      (byte-identical to `matplotlib._cm._petroff10_data`), a `color_sequences`-
      style registry (`ColorSequence`/`RegisterColorSequence`/
      `ColorSequenceNames`, seeded with `petroff10` and `tab10`), and registers
      `petroff10` as a `ListedColormap` so `GetColormap("petroff10")` resolves;
      kept out of `matplotlibListedColormapNames` because upstream petroff10 is
      a color sequence, not a registered colormap. The FuncNorm public-surface
      row flips from intentional-omission to idiomatic-equivalent. _Deferred:
      `MultiNorm` is a Matplotlib 3.11 feature (absent from the 3.10.9
      reference) that needs tuple-valued component arrays no Go artist accepts
      and depends on the deferred multivariate/bivariate colormap machinery; it
      stays an intentional omission until a multivariate-colormap consumer and a
      visible fixture exist._
- [x] **Misc artist kwargs:** `Stem` orientation, errorbar `capthick`, scatter
      `plotnonfinite`, `LineCollection` linestyle-string → dash conversion. All
      four ship as faithful ports of the Matplotlib 3.10.9 kwargs with parity
      fixtures. `StemOptions.Orientation` (`core/container.go`) swaps locs↔heads
      and the baseline axis for `"horizontal"` (invalid values warn + default
      vertical). `ErrorBarOptions.CapThick`/`ErrorBar.CapThick`
      (`core/plot.go`, `core/errorbar.go`) set the cap-line width in points
      (the cap markeredgewidth alias), replacing the hard-coded 1pt default.
      `ScatterOptions.PlotNonfinite` (`core/plot.go`) adds a `_combine_masks`
      port: by default any non-finite x/y/size/scalar/color masks the point;
      with the flag, non-finite color/scalar values are kept and ride the
      colormap "bad" color (already wired via `AtValue(NaN)` → bad, default
      transparent), and only non-finite positions/sizes are dropped.
      `LineCollection` gained `LineStyle`/`LineStyles` string fields
      (`core/collection_line.go`) resolved to dash patterns via the promoted
      shared helper `lineStyleToDashes` (`core/contour_styles.go`, formerly the
      contour-only `contourLineStyleDashes`); explicit numeric dashes keep
      precedence, and the support reaches `HLines`/`VLines` for free. New parity
      cases `stem_horizontal` (RMSE 4.10), `errorbar_capthick` (RMSE 3.91 —
      caps byte-match; residual is benign scatter-marker-center AA),
      `scatter_plotnonfinite` (RMSE 0.04), and `linecollection_linestyle`
      (RMSE 0.01); kwarg behavior unit-tested in `core/stem_orientation_test.go`,
      `core/errorbar_test.go`, `core/scatter_test.go`, and
      `core/collection_line_linestyle_test.go`.

**Exit criterion:** contour, colorbar, boxplot, and image cases match
Matplotlib default styling.

## Phase 10: Backend, Renderer & Styling Completion

**Goal:** finish the renderer/backend semantics and grow the styling system from
its current ~13% rcParam coverage.

- [ ] **Vector text metrics:** replace the crude `MeasureText` stubs in PDF/PS/
      PGF/SVG with the shared FreeType font manager (`backends/pdf/pdf_text.go:46`,
      `backends/svg/text.go:53`, etc.) so rotated/vertical anchoring matches AGG.
- [x] **Antialiased Gouraud** + **multi-stop gradients**. Gouraud triangles now
      rasterize through agg*go's antialiased `Agg2D.GouraudTriangle`
      (`span_gouraud_rgba`, dilation 0.5 to match Matplotlib's `_backend_agg.h`)
      when the batch's `Antialiased` flag is set; the binary point-sampled loop
      remains for `Antialiased=false` (`backends/agg/agg_gouraud.go`,
      `agg_draw.go`, `surface.go`). `gouraud_triangles` parity vs the Matplotlib
      reference drops **RMSE 3.95 → 0.26** (MaxDiff 227 → 2); `pcolormesh_gouraud`
      is unchanged (continuous mesh, no exposed AA edges). Gradients honor an
      arbitrary number of stops: radial via the existing
      `FillRadialGradientStops` and linear via a new
      `Agg2D.FillLinearGradientStops` (mirrors `FillLinearGradient` but builds the
      LUT from every stop via `buildNStopGradient`) added to agg_go
      (`internal/agg2d/gradient.go`, `agg.go`, float twins) — previously the
      linear path silently dropped interior stops and radial collapsed ≥3 stops
      to first/middle/last (`backends/agg/gradients.go`). Catalog tolerance for
      `gouraud_triangles` ratcheted to `MaxRMSE 0.6`. Unit coverage:
      `TestDrawGouraudTrianglesAntialiasesEdges`,
      `TestLinearGradientFourStopsHonorsInteriorColors`,
      `TestRadialGradientFourStopsHonorsInteriorColors`, and agg_go
      `TestFillLinearGradientStops`. \_Landing note: the linear-stops consumption
      requires publishing agg_go **v0.3.2** (the new method) + a `go.mod` bump;
      antialiased Gouraud and radial multi-stop build against the stock v0.3.1.*
- [x] **Sketch/xkcd** rendering pass. Matplotlib's sketch filter (LCG RNG +
      vpgen_segmentator + sinusoidal perturbation, faithfully ported from
      `third_party/matplotlib/src/path_converters.h`) now lives in
      `internal/sketch` and is applied in y-up display space at every backend's
      `Path()` entry (AGG, gobasic, SVG, PDF, PS, PGF; Skia via its CPU
      fallback). AGG skips snap/simplify when sketch is active so the dense
      wiggle survives. A global default flows from the new `path.sketch` rcParam
      / `style.WithXkcd()` via the `render.SketchAware` capability
      (`SetDefaultSketch`, set by `core.DrawFigure`); per-artist overrides are
      exposed on `Line2D.Sketch`/`Patch.Sketch` and win over the default. Parity
      case `sketch_xkcd` matches the Matplotlib reference at PSNR ~47 dB /
      MeanAbs ~1.4 (RNG-exact wiggle shape; residual is sub-pixel edge placement
      because the port applies sketch ahead of the device-space snap).
- [x] **PS/PGF** gradient + pattern fills; PGF clip-path and vertical-text
      interfaces. **PS** now implements `render.GradientFiller`/`PatternFiller`
      (`backends/ps/gradients.go`): gradients paint via Level-3 axial/radial
      shading dictionaries (`shfill`) whose dict/function format is byte-shared
      with the PDF backend (ShadingType 2/3 + FunctionType 2/3 stitching);
      patterns tile the cell path inside a path clip (bg rect + fg fill/stroke,
      mirroring AGG). **PGF** implements the same two fillers plus
      `render.ClipPathTransformer` and `render.FontVerticalTextDrawer`
      (`backends/pgf/gradients.go`, `lifecycle.go`, `text.go`): gradients are
      fully vector via `\pgfdeclare{horizontal,radial}shading` + clip +
      `\pgftext{\pgfuseshading}` (linear shadings rotated to the gradient axis,
      box fitted to the path bbox; matplotlib's own PGF backend rasterizes
      these), patterns tile within a `\pgfscope` clip, `ClipPathTransformed`
      applies the affine before `\pgfusepath{clip}`, and vertical text stacks
      glyphs centered to match PS/PDF. Both backends are y-up display space, so
      gradient/clip geometry is emitted without a device flip. Unit coverage:
      `backends/ps/gradients_test.go`, `backends/pgf/gradients_test.go`.
- [x] **`url`/`gid` metadata** in `GraphicsContext` for clickable vector output.
      Mirrors matplotlib's `Artist.set_url`/`set_gid`: a `render.URLMarker`
      capability (`SetURL`/`URL`/`SetGID`/`GID`) carries the active hyperlink
      target and element id; `core.drawArtist` snapshots/restores them around
      every artist draw (`core/rasterization.go`), and metadata lives on the
      embedded `ArtistRasterization` so all artists gain `SetURL`/`SetGID` for
      free. The **SVG** backend stamps each emitted node via a shared
      `newNode` helper and wraps it in `<a xlink:href>` + `<g id>`
      (`backends/svg/svg.go`, `export.go`); the **PDF** backend records
      `/Link` URI annotations (rect from `pathBounds`/measured text box) and
      emits an `/Annots` array on the page (`backends/pdf/pdf.go`,
      `pdf_document.go`), mirroring matplotlib's `_get_link_annotation`. PS/PGF
      ignore url (matplotlib's PS backend likewise). Unit coverage:
      `render/url_marker_test.go`, `core/url_metadata_test.go`,
      `backends/svg/svg_url_gid_test.go`, `backends/pdf/pdf_url_test.go`;
      verified end-to-end through `core.SaveSVG`/`SaveFig`.
- [x] finish `RestoreRegion` y-flip (`backends/agg/agg.go:371`) for blit/anim.
      `RestoreRegion` now takes its `bbox` (absolute crop sub-rect) and `offset`
      (translation delta) in y-up display space — flipped to the device buffer via
      `devRect`, consistent with `CopyFromBBox` and the backend's y-up public
      boundary. `region.Rect` stays device-space (matches matplotlib's
      `BufferRegion.get_extents`). Covered by `TestRestoreRegionWithBBoxAndOffset`.
- [x] **rcParams coverage** for `savefig.*`, `pdf.*`/`ps.*`/`svg.*`,
      `animation.*`, `boxplot.*`, `mathtext.*`, `hatch.*`, `image.*`, `date.*`.
      All eight groups are parsed, validated, stored on nested `style.RC`
      sub-structs (`ImageRC`/`HatchRC`/`BoxplotRC`/`MathtextRC`/`DateRC`/`PDFRC`/
      `PSRC`/`SVGRC`/`AnimationRC`/`SavefigRC` in `style/style.go`), and
      round-trip losslessly through `paramsFromRC` (guarded by
      `TestCurrentParamsRoundTripsWithoutUnsupported`). **Functionally wired:**
      `hatch.color`/`linewidth` (`core/contour_filled.go`), `image.cmap`/
      `interpolation` (`core/image_api.go`), `boxplot.*` show-flags/widths/colors
      (`core/boxplot.go`, `core/plot.go`), `mathtext.fontset` (`core/mathtext.go`
      resolver), `pdf`/`ps` fonttype/useafm/use14corefonts → backend font policy
      (`core/rcsaveopts.go`, seeded in `SavePDF`/`SavePS`). **`savefig.*` is
      applied at save time:** `render.FigureOptions` + `WithSave*` options
      (`render/extensions.go`) layered over `RC.Savefig` in `core/savefig_options.go`
      drive `facecolor`/`edgecolor`/`transparent` (figure + axes patches),
      `dpi` (vector + raster via new `agg.Renderer.Resize`/`render.Resizer`),
      `format` (extension-independent dispatch in `SaveFig`), and
      `bbox=tight`+`pad_inches` (raster pixel-crop in `core/savefig_tightbbox.go`;
      vector formats return a clear unsupported error). **Store-only** (validated
      + round-tripped, consumer pending, by design): `image.origin`/`aspect`/
      `resample`/`lut`, `mathtext.default`+per-style font patterns, all `date.*`
      (blocked on a strftime→Go-layout converter), `svg.fonttype`/`image_inline`
      (port's SVG default is native text), and all `animation.*`. Public-API
      audit + doc coverage regenerated.
- [ ] **Cyclers** for `linestyle`/`marker`/`linewidth`; bundle the common
      `.mplstyle` sheets (`seaborn-*`, `fivethirtyeight`, `bmh`,
      `Solarize_Light2`) (`style/theme.go:23`).

**Exit criterion:** vector text parity with AGG, antialiased Gouraud, and broad
rcParam + stylesheet coverage.

## Phase 11: Deferred Infrastructure Depth ⚠️

**Goal:** lower-priority structural depth surfaced by the review; not required
for the v1.0 parity claim but worth tracking so it does not vanish into "future
work."

- [x] **Bézier toolkit** (`bezier.py` equivalent): `split_bezier`, arc-length,
      offset/parallel curves, `inside_circle` — used by fancy arrows and
      annotation connectors. Ported to `geom/bezier.go`: `BezierSegment`
      (eval/`ArcLength`/`PolynomialCoefficients`/`AxisAlignedExtrema`),
      `SplitDeCasteljau`, `GetParallels`, `MakeWedgedBezier2`, `InsideCircle`,
      and the supporting helpers (`GetCosSin`, `GetNormalPoints`,
      `GetIntersection`, `CheckIfParallel`, `FindControlPoints`,
      `FindBezierTIntersectingWithClosedPath`,
      `SplitBezierIntersectingWithClosedPath`). Values verified against
      matplotlib 3.10.9 in `geom/bezier_test.go`; `core/arrow_patch.go` now
      delegates its wedge construction to `geom.MakeWedgedBezier2` instead of an
      ad-hoc reimplementation.
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
