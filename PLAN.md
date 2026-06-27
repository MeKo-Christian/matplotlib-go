# Matplotlib-Go Development Plan

This plan tracks the **remaining** work to bring `matplotlib-go` to a stable
v1.0 release. The foundation is built; what follows is the focused path to the
finish. The roadmap is cross-checked against the local upstream Matplotlib
snapshot in `third_party/matplotlib` so uncovered areas stay explicit instead of
sliding into a vague "future work" bucket.

Phases are ordered **closed first, open last**: Phases 1–9 are complete; the
remaining open work (Skia GPU, the v1.0 release stretch, and two deferred
infrastructure phases) is collected at the end under **Remaining Work**.

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

# Phase 1: Visual Parity Closure via Code Parity ✅

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

# Phase 2: Parity Status Reporting ✅

**Closed.** `docs/matplotlib-parity-status.md` is generated from machine
inventories and is kept current by CI: `TestAllUpstreamPublicRowsAreClassified`
guards unclassified rows; `TestPartialAndOmissionRowsHaveNotes` guards rows
without rationale. Per-family status tables cover every upstream feature family
with status, local API, fixture, and browser demo columns.

---

# Phase 3: Large File Decomposition ✅

**Closed.** All tracked Go files above 1k lines are split into focused units
(or documented with an explicit keep-large decision in
`docs/large-file-decomposition.md`). Decomposed in batches L1–L8: test files
by behavior family, algorithm-heavy core files (`contour`, `axis`, `text`,
`axes3d_contour_surface`, `legend`, `colorbar`), facade entrypoints (`pyplot`,
`cmd/parityviewer`), and backend implementations (`agg`, `gobasic`, `ps`,
`pgf`). The split is locked in by `just large-file-audit` plus the per-batch
`*SplitIsTracked` guard tests, and curated/generated-data files are a documented
keep-large exception with drift guards. Full `just fmt && just lint && just test`
and optional parity checks pass after decomposition.

---

# Phases 4–9: Parity-Breadth Closure

These phases are derived from the independent fidelity review in
[`REVIEW.md`](REVIEW.md) (2026-06-21). The review's verdict: the numerical and
rendering **core is faithful** (transforms, locators, contours, 3D projection,
norms, mathtext layout, FreeType text — verified line-for-line), but **breadth
is incomplete** — the long tail of Matplotlib kwargs/modes is missing across
nearly every subsystem, and several degradations are **silent**. The generated
`docs/matplotlib-parity-status.md` self-classifies most families as
"partial / thin"; these phases close that gap.

**Relationship to v1.0:** Phase 11 ships the release _mechanics_. Phases 4–9
closed the _parity_ the project advertises — silent-failure hardening (4),
formatter/layout/date fidelity (5), mathtext/text completeness (6), plot/colormap/
norm breadth (7), backend/renderer/styling completion (8), and the deferred
infrastructure depth (9) — so the Matplotlib-parity claim is now backed by code,
not scoped away. The genuinely-deferred residuals remain tracked in Phases 12 & 13.

## Phase 4: Silent-Failure Hardening & API Correctness ✅

**Closed.** Eliminated the worst UX failure mode — silently-wrong output with no
diagnostic — built on a shared swappable warning sink in `internal/diag`
(`diag.Warnf` / `diag.SetHandler`) that low- and high-level packages use without
coupling or enlarging the public API. All six paths now surface lost intent:
MathText warns on unknown commands instead of echoing them (shipped in
`github.com/cwbudde/mathtext v0.2.0`); an explicit `alpha=0` is baked into the
resolved colors at construction for the Bar/Fill/Hist/ErrorBar/BoxPlot families
(Step and the 3D bar/voxel paths already honored it); an unknown colormap warns
before the viridis fallback; Gouraud QuadMesh warns once when downgrading to flat
cells; invalid artist input (length mismatches, bad `errorevery`) logs a reason
instead of returning a bare `nil`; and the 3D naming traps `PlotSurface`/`Voxel`
warn once and document the real APIs (`Axes3D.Surface`, `Axes3D.Voxels`).
_Remaining (each needs a frozen-API break): `Violin`/`Grid` expose a plain
`float64` alpha where 0 is indistinguishable from "unset" (needs `*float64`), and
a mathtext strict (error) mode toggle is still open._

## Phase 5: Formatter, Layout & Date Fidelity ✅

**Closed.** Closed the highest-impact divergences on ordinary linear plots and
layouts. `ScalarFormatter` now ports Matplotlib's
`set_locs`/`_compute_offset`/`_set_order_of_magnitude`/`get_offset`, rendering
the shared additive offset and ×10ⁿ multiplier as offset text on both axes at
uniform tick precision; `constrained_layout` runs a real `LayoutGrid` constraint
solver (ported from `_layoutgrid.py`/`_constrained_layout.py` with per-margin
variables and suptitle/colorbar/legend reservations) that genuinely diverges from
`tight_layout`'s greedy heuristic; dates adopt Matplotlib days-since-epoch via
`Date2Num`/`Num2Date`/`SetEpoch`/`GetEpoch` with microsecond rounding; and
per-axes `Margins`/`SetXMargin`/`SetYMargin` + `round_numbers` autolimit mode,
`SetAdjustable`/`SetAnchor` aspect control, `NbinsAuto` axis-length-aware bin
counts, and non-mutating `label_outer()` shared-axes suppression all land.
Large/offset-magnitude and date figures match the references at RMSE ≤ ~2.

## Phase 6: MathText & Text Completeness ✅

**Closed.** Raised mathtext from ~13% to full symbol coverage and fixed text
fallback, all shipped in `github.com/cwbudde/mathtext v0.4.2` (no `replace`). The
complete 632-entry `tex2uni` table is generated and consulted as a final fallback
in both the layout parser and the plain-text normalizer (with operator/relation/
arrow spacing classes); the six math alphabets
(`\mathbb \mathcal \mathfrak \mathscr \boldsymbol \bm`) map to the Unicode
Mathematical-Alphanumeric block (Letterlike holes respected); a centered
separate-glyph accent model faithfully ports `Parser.accent` (full `_accent_map`,
char forms, wide accents, `\overline`-as-rule, `\overset`/`\underset`/`\substack`;
`\overbrace`/`\underbrace`/`\stackrel`/`\not` ship best-effort non-parity);
per-glyph multi-font fallback resolves Mathematical-Alphanumeric/symbol glyphs
DejaVu lacks to STIXGeneral instead of tofu; and `Text(bbox=boxstyle=…)` bridges
to the existing `FancyBboxPatch` styles via a Matplotlib boxstyle spec string.
Coverage is asserted ≥95% (100% today) by a symbol-table parity test, and parity
cases `mathtext_accents` and `text_bbox_styles` are green.

## Phase 7: Plot, Colormap & Norm Configuration Breadth ✅

**Closed.** Closed the per-artist configuration tail so common scientific plots
match Matplotlib defaults, not just the happy path. Boxplot gains the
`patch_artist=False` unfilled default (no color-cycle consumption), orientation,
the `showbox`/`showcaps`/`showmeans`/`meanline`/`sym` flags, `bootstrap` CI, and
percentile-clamped scalar `whis`; StackPlot gains `wiggle`/`weighted_wiggle`/`sym`
baselines (faithful `stackplot.py` port) with one-property-cycle-entry-per-layer
color cycling; Contour gains `negative_linestyles`, `extend`, `linestyles`,
contourf `hatches`, and `clabel` `fmt`/`rightside_up`, plus the contourpy
closed-loop start-vertex convention that fixes dashed-contour dash phase (RMSE
9→0.07); Colorbar gains norm-aware locators (SymLog/Power/TwoSlope/Centered/NoNorm),
`extendfrac`, and minor ticks; `imshow` gains native RGB/RGBA and `image.Image`
input (`ImShowRGB`/`ImShowImage`) bypassing colormap+norm; and `FuncNorm` plus the
`petroff10` color sequence and a `color_sequences`-style registry land. Misc
kwargs (`Stem` orientation, errorbar `capthick`, scatter `plotnonfinite`,
`LineCollection` linestyle strings) ship as faithful 3.10.9 ports. New parity
cases — `boxplot_default`, `stackplot_streamgraph`, `contour_styles` (2.50),
`colorbar_symlog_ticks` (0.65), `colorbar_extendfrac` (0.87), `imshow_rgb` (0.23),
`stem_horizontal`, `errorbar_capthick`, `scatter_plotnonfinite` (0.04),
`linecollection_linestyle` (0.01) — match Matplotlib styling, and two AGG backend
fixes gate the single-path half-pixel offset on a visible stroke (streamgraph
6.15→0.07) and always anti-alias contourf hatch strokes (hatched-contourf
26→0.55). _Deferred: `MultiNorm` (a 3.11 feature needing multivariate colormaps),
the full contourpy `locate_label` port, and long-tail kwargs (`hatch`-list
cycling, `sticky_edges`)._

## Phase 8: Backend, Renderer & Styling Completion ✅

**Closed.** Finished the renderer/backend semantics and grew the styling system
from ~13% rcParam coverage. The crude vector `MeasureText` stubs in PDF/PS/PGF/SVG
are replaced by the shared pure-Go font shaper (`render/text_metrics.go`:
`MeasureTextMetrics`/`MeasureTextInkBounds`/`MeasureFontHeightMetrics`), so vector
backends anchor rotated/vertical text and report descents like AGG (PNG goldens
byte-identical). Gouraud triangles rasterize through agg_go's antialiased
`GouraudTriangle` (RMSE 3.95→0.26) and gradients honor an arbitrary number of
stops (new `FillLinearGradientStops`, needing agg_go v0.3.2). Matplotlib's
sketch/xkcd filter (LCG + segmentator + sine, ported from `path_converters.h`)
applies in y-up display space across every backend, driven by the `path.sketch`
rcParam / `style.WithXkcd()` (parity ~47 dB). PS/PGF gain gradient + pattern
fills, and PGF gains clip-path + vertical-text interfaces (fully vector via
`\pgfdeclare…shading`). `url`/`gid` metadata flows through `GraphicsContext` into
clickable SVG `<a>`/`<g>` wrappers and PDF `/Link` annotations; `RestoreRegion`'s
y-flip is finished for blit/anim. rcParams coverage now spans
`savefig.*`/`pdf.*`/`ps.*`/`svg.*`/`animation.*`/`boxplot.*`/`mathtext.*`/`hatch.*`/
`image.*`/`date.*` (all parsed + round-tripped; the high-impact groups functionally
wired and `savefig.*` applied at save time; the rest store-only by design). A new
top-level `cycler` package (faithful `Cycler` port with `+`/`*`) plus the embedded
3.10.9 stylelib add linestyle/marker/linewidth cyclers and the standard `.mplstyle`
sheets, registered **register-if-absent** so hand-tuned built-ins and their goldens
are preserved (`classic` skipped: unsupported continuation-comment syntax).

## Phase 9: Deferred Infrastructure Depth ✅

**Closed.** Landed the lower-priority structural depth the review surfaced; the
genuinely-deferred residuals were spun out into Phases 12 & 13 (cocircular Qhull
parity and renderer-wired cached transformed paths). A Bézier toolkit
(`geom/bezier.go`: `SplitDeCasteljau`, arc-length, `GetParallels`,
`MakeWedgedBezier2`, `InsideCircle`, …) and path-generator helpers
(`geom/path_generators.go`: `UnitCircle`/`Arc`/`Wedge`/`EllipseBezier`, regular
polygons/stars) replace the ad-hoc reimplementations across `core`/`render`,
verified against matplotlib 3.10.9. The `tri/` package now owns the
`Triangulation` type, point location (`TrapezoidMapTriFinder`), interpolation
(linear + reduced-HCT cubic), refinement, and analysis, with Delaunay
connectivity computed by the pure-Go exact-predicate `tri/qhull` engine
(`math/big.Rat`) — identical to Qhull for general-position points (cocircular
diagonal choice is the Phase 12 residual). Live bbox-linked transforms
(`BboxTransformTo` wired to the `TransformNode` invalidation graph) give each Axes
a persistent `axesBbox`/`transAxes`/`transData` graph that invalidates-on-change
instead of rebuilding every draw (RMSE 0, affine leg cached); the `AffineProvider`
capability opens the transform type set to third parties; `transform.splitAffine`/
`TransformedPath` cache the affine/non-affine split (renderer wiring deferred to
Phase 13); the path simplifier matches Matplotlib's single-pass algorithm; and a
teardown/introspection API (`Clear`/`Cla`/`Remove`/`DelAxes`/`Clf`, `Getp`/`Setp`/
`Findobj`/`FindobjType`) mirrors Matplotlib's lifecycle surface.

---

# Remaining Work (Open Phases)

The phases below still carry open to-do items and are ordered after the closed
phases. **Phase 10** finishes Skia GPU mode; **Phase 11** is the v1.0 release
stretch (docs and performance are done, only release mechanics remain); and
**Phases 12 & 13** are deferred-by-decision infrastructure depth, kept here so
the open design work does not vanish into "future work."

## Phase 10: Backend Deepening (Skia Native + GPU)

**Goal:** finish the backend-specific Skia work. The historical blocker (no Skia
C-ABI binding) is **resolved**: a real Skia library is built locally and a narrow
C-ABI wrapper links it under `-tags "skia skiacgo"`. The remaining work is wiring
GPU mode — no longer external-access-blocked, just unbuilt.

### Done (summary)

C-ABI Skia wrapper + cgo bridge (`backends/skia/skia_cwrap.{h,cpp}`,
`native_cgo.go`; tag `skiacgo`) links Skia milestone 151. Native path collections,
quad meshes, transformed images, tiled-shader hatching, gradient fills, marker
batches, and Gouraud triangles are implemented via the wrapper. AGG non-text
parity diagnostics (`TestNonTextResidualDiagnostics`) are in place.

### Build / test (carry-on entry points)

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

### Remaining work

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

## Phase 11: Documentation, Performance, and v1.0 Release

**Goal:** make the project consumable by users who have not followed the
development thread, establish performance baselines, and tag a stable v1.0.

### 11.1 API Documentation ✅

Package-level GoDoc with worked examples for every stable public package;
guarded by `TestStablePublicPackagesHaveGoDocAndExamples`. Hosted docs at
pkg.go.dev; WASM gallery landing page links to README, examples, backend
selection, migration notes, and parity status.

### 11.2 Performance Pass ✅

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

### 11.3 Release Readiness

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

## Phase 12: Cocircular Qhull-Faithful Delaunay (deferred) 🔬

**Goal:** make `tri/qhull`'s Delaunay byte-for-byte identical to matplotlib's
Qhull backend (`qhull d Qt Qbb Qc Qz`, qhull 8.0.2) even for **cocircular**
inputs (≥4 points on a common circle), where the Delaunay triangulation is
non-unique. General-position connectivity already matches exactly via the
exact-predicate engine; this phase only closes the cocircular diagonal choice.

**Status:** in progress — **full faithful port chosen** (close all 34). Stages 1,
2, 3a and 3b are done: the fan model is proven (34/34 + 27/27 from Qhull's true
creation order), the `qh_buildhull` incremental hull is ported, the faithful
`qh_partitionall` greedy first-facet partition + one-time `qh_furthestnext` seeding
reach 28/34, and the non-simplicial coplanarhorizon merge (`qh_premerge`→
`qh_mergecycle`) + first-clearly-outside `qh_findbestnew` + vertex-id cone ordering
bring the fully _computed_ engine to **31/34 cocircular + 27/27 general** (by
triangle set).

**Root cause of the last 3 (established by per-step trace, 2026-06-26).** The
divergence is _not_ confined to the final insertions, and it is _not_ the merge
tolerance. Measured against the captured ground truth, the computed **exact build
order** matches only **5/27 general + 31/34 cocircular**
(`TestComputedOrderMatchesGroundTruth`) — general still passes by _set_ because its
Delaunay is unique, which masks the order divergence. The per-step Qhull oracle
(`third_party/qhull-8.0.2/stepdump.c` + the `QHSTEP` patch in `qh_buildhull`) shows
the build orders diverge as early as **step 0**: the current facet model represents
a facet as a _vertex set_ and re-derives its boundary edges and neighbours
geometrically (angle-sort + opposite-edge matching), so it cannot reproduce the
order in which Qhull **creates and appends** new facets — which is driven by
Qhull's explicit **ridge graph** (`FOREACHridge_(visible->ridges)` in
ridge-storage order) and `qh_matchnewfacets` hashing, not by geometry. Closing the
last 3 requires replacing the vertex-set model with a faithful ridge-graph build so
the computed creation order matches Qhull bit-for-bit (Stage 3c, decomposed below).

**Update (faithful ridge engine landed, `tri/qhull/ridgebuild.go`).** The vertex-set
model is superseded by a faithful incremental hull that keeps Qhull's own layout
(inverse-id vertex set + parallel neighbour array, maintained incrementally) and
ports `qh_createsimplex`/`qh_initialhull` (toporient + global flip),
`qh_sethyperplane_det`/`qh_normalize2` orientation, `qh_makenew_simplicial`,
`qh_findbest` (directed walk, `noupper=False`), `qh_findbesthorizon`, and the
`qh_partitionpoint` facet_next bookkeeping. The computed build order now matches the
captured ground truth **27/27 general** (was 5/27 — exact order now, not just the
triangle set) **+ 24/34 cocircular** exact-order (`TestComputedOrderRidge`, general
hard-gated). The remaining 10 cocircular cases are the merge cases (Stage 3c.6,
partial/opt-in below). The shipped `Delaunay`/`DelaunayMatched` still use the
vertex-set engine until the merge is complete.

### What's already done (foundation, in-tree)

- **Differential harness + corpus:** `tri/qhull/corpus_test.go` +
  `testdata/corpus.json` (61 cases, 27 general + 34 cocircular, generated from
  live `matplotlib.tri` by `testdata/gen_corpus.py`). Compares the Go engine's
  triangle **set** + neighbor **graph** to Qhull (canonical, order-independent).
  General position passes 27/27; cocircular sits at 6/34 behind a ratchet const.
- **Validated faithful port of the deterministic front of the pipeline**
  (`tri/qhull/build.go` + `build_test.go`), all bit-matching Qhull: mean-subtract
  → paraboloid lift → **Qz infinity point** (xy centroid, last coord
  `1.1·maxboloid`) → `qh_maxmin` extremes/tolerances → **Qbb `qh_scalelast`**
  (last coord → `[0, MAXabs_coord]`); `qh_distround` (`DISTround`); `det2`/`det3`/
  `detsimplex`; and `qh_maxsimplex` (the `hull_dim=3` path). Example: the unit
  square projects to exact `{(-0.5,-0.5,0)…(0,0,0.5)}`, `DISTround=6.937e-16`,
  initial simplex append order `[p0,p1,p2,∞]` — all identical to Qhull.
- **Reference + introspection tooling** (gitignored under
  `third_party/qhull-8.0.2/`): the qhull 8.0.2 source (sha-matched download) plus
  `introspect.c` (dumps vertex creation order `vid:pid` and facet vertex order
  pre/post `qh_triangulate`) and `dump_state.c` (dumps projected coords, globals,
  flags). These give ground truth to validate every stage. `just qhull-corpus`
  builds `introspect` and regenerates the fixtures.
- **Stage 1 — the fan model, proven (`tri/qhull/fanfromorder.go` +
  `fanfromorder_test.go`).** `delaunayFromOrder` takes the exact-predicate
  triangulation, groups adjacent triangles whose four points are cocircular within
  roundoff into convex cells (scale-invariant normalized in-circle residual <
  `cocircularRelTol`, robust to the thin base triangles a single lifted-distance
  test mis-scales), and re-fans each cell from its highest-creation-rank vertex
  (`apex = {boundary edge, apex} ∀ edges not at apex`). Fed Qhull's **captured**
  creation order (`testdata/creation_order.json`, from `introspect`),
  `TestFanFromGroundTruthOrder` reproduces Qhull **34/34 cocircular + 27/27
  general** bit-for-bit. This isolates all remaining work to _computing_ the
  creation order.

### Key insights (why it's hard)

- **No geometric rule.** Qhull's cocircular diagonal is governed by its
  incremental construction order, not geometry: the square's diagonal flips with
  input reordering; a 4×4 grid uses different diagonals in adjacent cells.
- **The fan rule** (now confirmed from source + proven end-to-end). Each Delaunay
  cell is fanned from its **last-created vertex**: `qh_triangulate_facet` apex =
  `SETfirst_(facet->vertices)` (poly2_r.c:3692), and Qhull keeps each facet's
  vertex set inverse-sorted by `vertex->id` (poly2_r.c:25, asserted :777), with
  `vertex->id = qh->vertex_id++` monotonic at creation. So the diagonal is fixed
  entirely by the **vertex creation order** — and Stage 1 proves that order is the
  _only_ missing input.
- **Creation order needs the build.** Order = `qh_maxsimplex` simplex (3 input
  pts + ∞, already ported) then `qh_buildhull`'s furthest-insertion. `qh_nextfurthest`
  walks the facet list and takes each facet's furthest outside point — not
  global-furthest — so it needs the real facet/hyperplane/horizon/cone structures.
- **Cocircular points are ordinary insertions, not `coplanarset` promotions**
  (corrected). For the lifted Delaunay, a near-cocircular point is _outside_ the
  tilted lower facets even while coplanar with the upper facet, so it enters via
  the outsideset→`qh_addpoint` path and gets a normal `qh_newvertex` id; Qhull's
  `qh_mergecycle` then folds its coplanar cone into the cell (changing facet
  topology, which Stage 1 reconstructs independently). General position and most
  cocircular cases therefore need **no** premerge for the order.
- **The build order has three drivers, not just premerge (corrected by the Stage 3
  trace).** Reproducing Qhull's `qhull d Qt Qbb Qc Qz` build order needs, in order
  of discovery: (1) `qh_partitionall`'s **greedy first-facet** assignment — each
  point joins the _first_ facet it is clearly outside of (`dist >= 2*MINoutside`),
  not its globally-furthest facet, with the running-furthest deferred to the end of
  the outside set; (2) the **one-time `qh_furthestnext`** in `qh_initbuild` that
  moves the globally-furthest facet to the list head before the build (PICKfurthest
  is _off_, so this is a single seeding, after which `qh_buildhull` walks plain
  facet-list order); and (3) the **coplanarhorizon merge** — when the apex is
  coplanar with a horizon facet (`qh_findhorizon` marks it, `dist` in
  `[-MAXcoplanar, MINvisible]`), `qh_premerge`→`qh_mergecycle` folds the new cone
  facet _into_ that horizon facet **before** `qh_partitionvisible`, so the absorbed
  outside points keep the horizon facet's early list position instead of moving to
  a tail facet. (1)+(2) are pure ordering and close 25→**28/34**; (3) is the
  genuine premerge work (it makes facets non-simplicial) and closes the last 6.

### Remaining work (if revived)

- [x] **Stage 1 — fan model proven** with captured creation order: 34/34
      cocircular + 27/27 general (`TestFanFromGroundTruthOrder`). All remaining
      work is now _computing_ the creation order.
- [x] **Stage 2 — `qh_buildhull` ported** (`tri/qhull/buildhull.go`): the
      incremental lifted lower-hull — `qh_setfacetplane`/`sethyperplane_det` +
      `qh_normalize2`, `qh_distplane`, `qh_createsimplex`/`qh_initialhull`,
      `qh_partitionall`, the `qh_nextfurthest` facet-list walk, and `qh_addpoint`
      (`qh_findhorizon` BFS, `qh_makenewfacets`/`qh_makenew_simplicial`,
      `qh_matchnewfacets`, `qh_partitionvisible`), with `qh_newvertex` ids. The
      self-contained engine `delaunayComputed` (no Qhull, no fixture) reaches
      **27/27 general + 25/34 cocircular** (`TestDelaunayComputed`); all 61 build
      with no Gaussian-fallback degeneracy.
- [x] **Stage 3a — faithful partition + furthest seeding** (`tri/qhull/buildhull.go`):
      `qh_partitionall`'s greedy first-facet block (with `2*MINoutside` and the
      deferred-furthest ordering) + block-2 leftover `qh_partitionpoint`, and the
      one-time `qh_furthestnext` seeding. Lifts the computed engine to **28/34
      cocircular + 27/27 general** (`TestDelaunayComputed`, ratchet 28); shipped
      `Delaunay` still uses the exact engine.
- [x] **Stage 3b — coplanarhorizon merge** (`tri/qhull/buildhull.go`): the
      non-simplicial `qh_premerge`→`qh_mergecycle` merge (a coplanar cone facet is
      folded into its horizon facet, which keeps its plane, moves to the tail and
      becomes a new facet), plus the first-clearly-outside `qh_findbestnew` and
      Qhull's vertex-id cone/horizon iteration order. Facets are modelled as vertex
      _sets_ — boundary edges derived by in-plane angle-sort, neighbours rebuilt by
      matching opposite directed edges — so the merge is a vertex-set union. Lifts
      the computed engine to **31/34 cocircular + 27/27 general**
      (`TestDelaunayComputed`, ratchet 31), no Gaussian-fallback degeneracy.
- [ ] **Stage 3c — faithful ridge-graph build order → 34/34.** Replace the
      vertex-set facet model with Qhull's explicit ridge graph so the computed vertex
      creation order matches Qhull bit-for-bit. Decomposed (divide and conquer):
  - [x] **3c.0 — exact-order oracle test.** `TestComputedOrderMatchesGroundTruth`
        compares computed `buildHullOrder` against captured `creation_order.json`
        per case (the tight loop; the fan model proves order ⇒ diagonal). Baseline
        **5/27 general + 31/34 cocircular**. Per-step oracle: `stepdump.c` + the
        `QHSTEP` patch dump the facet list + outside sets before every pick.
  - [x] **3c.1–3c.5 — faithful simplicial build → 27/27 general exact-order.**
        New `tri/qhull/ridgebuild.go` replaces the vertex-set model with Qhull's own
        layout: each facet keeps a vertex set inverse-sorted by creation id and a
        PARALLEL neighbour array (`nbr[i]` across the ridge opposite `verts[i]`),
        maintained incrementally (no geometric rebuild). Ported faithfully:
        `qh_createsimplex`/`qh_initialhull` (alternating `toporient` + global flip),
        `qh_setfacetplane`→`qh_sethyperplane_det`+`qh_normalize2` (sign from
        `toporient`, NOT an interior test — this is what orients the Qz infinity
        "ceiling" facets correctly), `qh_makenew_simplicial` (FOREACHneighbor,
        apex-first, horizon relink), the new-facet sibling matching, and — the
        decisive correction — `qh_partitionvisible` uses **`qh_findbest`** (directed
        greedy walk from the replacement; `qh_USEfindbestnew` is false for these
        low-merge inputs), with `noupper=False` (a clearly-outside facet of EITHER
        kind early-returns), the `qh_findbesthorizon` frontier fallback, and the
        `qh_partitionpoint` `facet_next`/`isnewoutside` bookkeeping. Result: computed
        build order matches the captured ground truth **27/27 general + 24/34
        cocircular** (`TestComputedOrderRidge`, general hard-gated). Per-step oracle:
        `stepdump.c` + the `QHSTEP` patch.
  - [~] **3c.6 — coplanar-horizon merge ported, ON by default → 30/34.** The merge
        — `linkSamecycle`/`premerge`/`mergeInto`, explicit `redge`/`rnbr`/`rtop` ridge
        lists via `makeRidges`, a unified `boundary()` so non-simplicial facets walk
        their ordered ridges, `qh_makenew_nonsimplicial` ridge-top orientation, the
        `qh_getreplacement` chain through merged cones — is **strictly better than the
        simplicial-only build (30/34 vs 24/34 cocircular, fixing 6 and regressing 0;
        general stays 27/27)**, so it is on by default (`QHULL_MERGE=0` to disable).
        A cone facet across a coplanar horizon ridge is folded into the horizon facet,
        which absorbs the apex at position 0, drops the shared/interior ridges and
        vertices, keeps its plane, and moves to the tail as a non-simplicial new facet.
        Four bugs fixed this pass: (1) `mergehz` is per-addPoint state — must clear
        each step (a stale flag caused runaway over-merging that collapsed the hull);
        (2) cone facets from a non-simplicial visible facet need the ridge-top
        orientation; (3) **the `horizonskip` parity is unreliable across a merged
        horizon** (its neighbour order ≠ Qhull's yet) — orient those cone facets
        geometrically (interior-below; safe because non-simplicial horizons occur only
        in cocircular cases, never in the 27/27 general build); (4) the post-merge
        new-facet neighbour graph can be locally disconnected, so the directed
        `qh_findbest` walk misses a facet a point is clearly outside — added a
        full-scan fallback. NB: the earlier "`qh_partitioncoplanar` coplanar-promotion"
        framing was a red herring — points drop because the wrong facets exist.
        Replacement insight (closed grids 3×4/4×3/4×4/5×2/5×3/5×4, rings 1.0_2.0_6/8,
        and reg12 → **31/34**): the seed of `qh_partitionvisible`'s directed walk is
        `visible->f.replace` = `qh_makenewfacets`' returned facet. When the visible
        facet is adjacent to a **merged** facet it carries a materialised ridge, so
        Qhull runs `qh_makenew_nonsimplicial` too and sets the replacement to **that
        ridge cone (newfacet2)**, not the last simplicial cone. Approximating "has a
        ridge to nb" as "nb is a coplanar (`mergehz`) non-simplicial horizon" reaches
        31/34. (Verified dead-ends along the way: `max_outside` stays at roundoff so
        `qh_findbestnew == qh_findbest`; global-furthest/first-cone replacement crater
        general to 4–5/27; reversing cone-creation order craters everything; the exact
        `ridgeTo` materialisation I tried over- or under-sets vs Qhull.)
        _Remaining (last 3: grid6x3, grid6x4, rings_0.5_1.0_5)._ **Exact rule now
        verified from source** (`poly2_r.c:2505-2515`): `visible->f.replace =
        newfacet2` (the `qh_makenew_nonsimplicial` **last** cone) iff
        `visible->ridges` is non-empty, else `newfacet` (the `qh_makenew_simplicial`
        last cone) — in both paths simply _the last cone built for that visible
        facet_, which `makeNewFacets` now records by overwriting `replace[vf]` per
        cone. The one hypothesised mechanism for the last 3 — `qh_makeridges` on a
        **simplicial neighbour touched twice in one merge cycle**
        (`qh_mergecycle_neighbors`, `visitid==visit_id` branch) — was implemented
        faithfully and **verified to NEVER fire on the 34-case corpus** (debug
        counter), so the last 3 do **not** hinge on neighbour ridge materialisation.
        They hinge purely on the **merged horizon's own ridge-append order**
        (`qh_mergecycle_ridges`): for a non-simplicial merged-horizon `vf`, `boundary()`
        does not yet reproduce that order, so its "last cone" ≠ Qhull's `newfacet2`.
        The `mergehz` override pins the right cone for all but these 3. _Next (needs the
        QHATTACH oracle):_ rebuild the patched qhull, read Qhull's actual `newfacet2`
        for the 3 cases, and reproduce `qh_mergecycle_ridges`' ridge order in
        `mergeInto`. _Gate:_ 31/34 → 34/34. (Verified dead-ends: `max_outside` stays at
        roundoff so `qh_findbestnew == qh_findbest`; global-furthest/first-cone
        replacement crater general to 4–5/27; reversing cone-creation order craters
        everything; the exact `ridgeTo` materialisation over/under-sets vs Qhull;
        pure last-cone without the `mergehz` override drops to 30/34.) Ridge now ties
        the shipped vertex-set engine at 31/34-by-set.
  - [ ] **3c.7 — close + lock.** Computed order = ground truth **61/61**
        (`QHULL_ORDER_STRICT=1`); switch `buildHullOrder`/`DelaunayMatched` to the
        ridge engine; `delaunayComputed` → 34/34 cocircular + 27/27 general; bump
        `computedCocircularRatchet` (and the fan ratchet) to 34; delete the
        vertex-set model + per-step trace scaffolding. _Gate:_ both ratchets at 34.
- [x] **Stage 4 — wired into `tri.delaunayTriangles`** via `qhull.DelaunayMatched`
      (`tri/qhull/fanfromorder.go`, `tri/delaunay.go`): general position takes a
      fast path returning the exact triangulation unchanged (the order computation
      is skipped — no cost or risk on large inputs); cocircular inputs get the
      computed fan; and if `buildHullOrder` bails the exact (still valid) Delaunay is
      returned. Three optional-visual 3D goldens shifted to the new diagonal and were
      verified against the matplotlib references (`TestReferenceCompare` PSNR
      50–57 dB): `mplot3d_gallery` and `mplot3d_trisurf3d` regenerated (deterministic);
      `mplot3d_tricontourf3d` left as-is — it is a pre-existing non-deterministic
      optional-visual case (a 3D depth-order tie, orthogonal to triangulation).

**Exit criterion:**

- [ ] `TestComputedOrderMatchesGroundTruth` reports 61/61 exact build-order match
      (27/27 general + 34/34 cocircular) under `QHULL_ORDER_STRICT=1`.
- [ ] `go test ./tri/qhull/` differential harness reports 34/34 cocircular (and
      27/27 general) bit-for-bit against Qhull, with no regression in the
      `just test` parity goldens.

---

## Phase 13: Cached Transformed Paths in the Render Path (deferred) 🔁

**Goal:** make artists (`Line2D`, patches, collections) draw through a
persistent `transform.TransformedPath` so a redraw that changes only the trailing
affine (axes resize/pan/zoom) reuses the cached non-affine projection instead of
re-running it per vertex — and an unchanged redraw skips the per-vertex transform
entirely. This realizes, in the renderer, the affine/non-affine cache split that
Phase 9 already built into `TransformedPath`.

**Status:** deferred by decision — the payoff requires **repeated redraws** of the
same artist (interactive pan/zoom or animation). The current pipeline draws each
figure once to a PNG, so a cross-draw cache is never hit; wiring it in now adds
hot-path complexity and parity risk for no runtime benefit. Revive this alongside
an interactive/animation redraw loop.

### What's already done (foundation, in-tree)

- **`transform.splitAffine`** (`transform/transform.go`) decomposes any `T` into
  its maximal trailing affine + non-affine remainder
  (`t.Apply(p) == trailing.Apply(nonAffine.Apply(p))`), reusing `Frozen`/
  `AsAffine`. Validated against every transform type incl. `Chain`/`OffsetT`
  nesting and third-party `AffineProvider`.
- **`transform.TransformedPath`** caches the non-affine vertex pass and re-applies
  the trailing affine separately (`TransformedPointsAndAffine()`/`Affine()`;
  `Transformed()` unchanged). `InvalidAffine` (e.g. `Bbox.Set`) refreshes only the
  affine; `InvalidNonAffine`/`InvalidAll` re-runs the projection. Covered by
  `TestSplitAffine` + `TestTransformedPathAffineNonAffineSplit`.
- The persistent per-axes transform graph from Phase 9 (`axesBbox`/`transAxes`/
  `transData` with `dataNode`) is the natural invalidation source to wire to.

### Key insights (why it's more than a drop-in)

- **`refreshDataTransform` fires the wrong stage.** `core/axes_transform.go:
refreshDataTransform` currently invalidates `dataNode` with **`InvalidAffine`
  for every change**, including the "non-affine leg changed every draw" branch
  (log/symlog, polar/geo/3D). A split-aware consumer wired to `dataNode` would
  therefore reuse a **stale projection** on a log-domain/limits change and render
  incorrectly. Fix first: fire `InvalidNonAffine` when the data leg is non-affine,
  keep `InvalidAffine` only for the affine-leg/bbox case.
- **Artists hold no axes/node reference.** `Line2D` (and the shared
  `core/patch_paths.go:buildArtistDisplayPath`) get their transform only from the
  per-draw `DrawContext`. To register a `TransformedPath` as a dependent of
  `axesBbox.Node()`/`dataNode`, those nodes must be plumbed through `DrawContext`.
- **Segmentation must stay on final points.** `Line2D.displayPath`
  (`core/line.go:447`) interleaves NaN-segmentation with the per-vertex transform.
  To stay byte-identical, apply the trailing affine to the cached non-affine
  points, then run the **existing** finiteness/segmentation loop on the final
  points (affine of finite is finite; affine of NaN is NaN, so the mask matches).
- **Log/polar resize win needs leg-change detection.** Even after the stage fix,
  non-affine legs are treated as "changed every draw" (the Phase 9 deferral), so
  log/polar resizes re-project every draw. The real resize win for those needs a
  cheap non-affine-leg equality check (e.g. a scale/version identity) — itself a
  follow-up. The affine (linear-axes) and unchanged-redraw wins land without it.

### Remaining work (if revived)

- [ ] Refine `refreshDataTransform` to fire `InvalidNonAffine` for non-affine data
      legs (and `InvalidAffine` otherwise). _Gate:_ existing
      `core/axes_transform_test.go` stays green; add a test asserting a log-domain
      change invalidates non-affine.
- [ ] Expose the axes invalidation nodes (`axesBbox`/`dataNode`) on `DrawContext`.
- [ ] Pilot on `Line2D`: persistent `transformedPath` field, `SetPath` on data
      change, dependency-wire to the nodes, draw via `TransformedPointsAndAffine` +
      the unchanged segmentation loop. _Gate:_ golden/reference RMSE 0.
- [ ] Extend the same pattern through `buildArtistDisplayPath` to patches and
      collections. Re-run the parity gate after each.
- [ ] (Optional, larger) Non-affine-leg change detection so log/polar resizes also
      reuse the projection.

**Out of scope:** passing the affine _separately_ to the agg backend so it
composes with the device y-flip and does zero per-vertex affine work
(matplotlib's deepest win) — needs a `render.Renderer` capability/signature
change.

**Exit criterion:**

- [ ] With an interactive/animation redraw loop in place, an affine-only redraw of
      a linear-axes artist re-runs no non-affine vertex pass (asserted via a
      counter probe), and the full `just test` golden+reference suite stays RMSE 0.

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
documented v1.0 release. **Phases 1–9** are **closed**: the visual/code parity
closure (1–3) and the parity-breadth closure derived from [`REVIEW.md`](REVIEW.md)
(4–9, covering silent-failure hardening, formatter/layout/date fidelity,
mathtext/text completeness, plot/colormap/norm breadth, backend/renderer/styling
completion, and deferred infrastructure depth). The remaining open work is
collected at the end: **Phase 10** finalizes GPU acceleration for the Skia backend
(CPU native primitives are done); **Phase 11** is the final v1.0 stretch —
documentation and performance work is done, only the release mechanics (changelog,
CI gate, v1.0 tag) remain; and **Phases 12 & 13** track genuinely-deferred
infrastructure depth (cocircular Qhull parity, renderer-wired cached transformed
paths).
