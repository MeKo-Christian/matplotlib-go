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
connectivity computed by the standalone pure-Go `github.com/cwbudde/qhull-go`
engine — identical to Qhull's `qhull d Qt Qbb Qc Qz` backend for general position
and, via the faithful ridge build, for the cocircular diagonal too (Phase 12,
shipped and extracted). Live bbox-linked transforms
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
**Phase 13** is deferred-by-decision infrastructure depth, kept here so the open
design work does not vanish into "future work." (**Phase 12** is complete —
shipped and extracted to `github.com/cwbudde/qhull-go`; its compact record stays
below for provenance.)

## Parity Follow-up: Sketch figure-patch fill-coverage (open) 🔁

**Status:** open. The `sketch_xkcd` case was taken from **RMSE 7.7 → 3.5** (2026-06-28)
by enabling path simplification by default (`style.PathSimplify: true`, matching
Matplotlib's `path.simplify=True`) — Matplotlib simplifies a line _before_ the
sketch filter, so the segmentator+RNG must see the same simplified polyline or the
wiggle desyncs. The sine wiggle is now **vertex-exact** vs Matplotlib and the result
is visibly indistinguishable. This follow-up is the **last ~0.5 RMSE** (3.5 → ~1.8),
which is **invisible on a white display** and was deferred by decision.

**Root cause (fully diagnosed):** the residual is **42 fully-transparent
(255,255,255,α=0) pixels on the canvas border** of the Matplotlib reference.
Matplotlib applies the global `path.sketch` to the **figure background patch**
(`figure.patch`, a full-canvas white Rectangle, drawn on a transparent RGBA buffer);
its wiggled fill edges leave those 42 border pixels uncovered. Go instead paints the
figure background as the raster backend's **opaque clear** (`agg.New(w,h,white)`),
which a sketch can never perforate, so Go's border is fully opaque. (Confirmed
read-only: a Matplotlib render with `path.sketch` has 42 sub-opaque border pixels;
without it, 0.)

**What was tried and reverted** (the naive fix is not enough):

- Added a `render.TransparentClearer` capability + `agg` `ClearTransparent`, and had
  `DrawFigure` clear the buffer transparent and repaint the figure facecolor as a
  sketchable patch (`pixelRectPath(vp)`), gated to sketch-active renders.
- Verified Go's `sketch.Apply` on the figure-patch rectangle is **byte-identical** to
  Matplotlib's Sketch (1999 verts, exact coords — replicated the MSVC-LCG +
  vpgen_segmentator + sine in Python and it matched Go to 4 decimals).
- **Despite the identical sketched path, filling it diverges:** Go's AGG fill on the
  transparent canvas leaves **~1000 semi-transparent border pixels** vs Matplotlib's
  **42** (RGB _regressed_ 1.8 → 3.5, RGBA 6.5). So the blocker is **AGG fill coverage
  on a transparent canvas**, not the sketch.
- Also surfaced: `Renderer.GetImage()` returns **straight alpha in an `image.RGBA`**
  (whose Go contract is premultiplied — see the `SavePNG`/NRGBA workaround comment in
  `backends/agg/agg_text.go`). So `test/imagecmp.ComparePNG` (and any `image.RGBA`
  consumer) mis-handles semi-transparent pixels: Matplotlib's reference loads as
  NRGBA and gets premultiplied by `At()`, Go's mislabeled RGBA does not → transparent
  pixels are compared inconsistently. This must be fixed for the metric to even be
  trustworthy on transparent output.

**To actually close it (scoped):**

1. Fix the straight-vs-premultiplied alpha contract: make `GetImage()` return a
   correctly-labeled image (`image.NRGBA`, straight) — or have the parity harness
   compare via `ImageView()`/NRGBA — so semi-transparent pixels round-trip and compare
   correctly. Re-baseline; this alone may move the number.
2. Understand why Matplotlib's filled figure patch border is near-opaque (only 42
   sub-opaque px) while Go's identical path fills with a ~1px semi-transparent band:
   compare Go AGG scanline fill coverage vs Matplotlib's vendored AGG
   (`agg_rasterizer_scanline_aa` / `agg_scanline_p`) on the wiggly closed rectangle.
   Likely an `../agg_go` fill-coverage parity issue (per AGENTS.md, fix `../agg_go`).
3. Only then wire the transparent-canvas figure patch in `DrawFigure` (the reverted
   `drawSketchedFigurePatch` approach is the right shape — gate to sketch-active so
   non-sketch output is byte-identical).

**Exit:** `sketch_xkcd` reference-compare RMSE < 3 (target ~1.8) with **no regression**
to the other 10 dense-line goldens that now simplify, then tighten the catalog
tolerance (`internal/examplecatalog/catalog.go`) accordingly.

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

## Phase 12: Cocircular Qhull-Faithful Delaunay ✅ (shipped & extracted)

**Goal:** make the Delaunay triangulation byte-for-byte identical to matplotlib's
Qhull backend (`qhull d Qt Qbb Qc Qz`, qhull 8.0.2), including **cocircular**
inputs (≥4 points on a common circle) where the triangulation is non-unique and
Qhull's diagonal is fixed by construction order, not geometry.

**Outcome — done.** A faithful incremental **ridge-graph** hull (Qhull's own
layout: inverse-id vertex sets + parallel neighbour arrays + explicit ridge
lists, ported from `qh_createsimplex`/`qh_initialhull`, `qh_sethyperplane_det`,
`qh_findbest`/`qh_findbesthorizon`, `qh_makenewfacets`, and the
`qh_premerge`→`qh_mergecycle` coplanar-horizon merge) computes Qhull's exact
**vertex creation order**; each cocircular cell is then re-fanned from its
last-created vertex — proven to reproduce Qhull's diagonal. General position is
exact (27/27); cocircular matches Qhull's exact build order. Wired into
`tri.delaunayTriangles` (general position takes a fast exact path; cocircular gets
the computed fan; falls back to the valid exact Delaunay if the order computation
ever bails). Three 3D goldens shifted to the more-faithful diagonal and were
regenerated against the matplotlib references.

**Extracted to a standalone module (2026-06-27).** The package now lives at
**`github.com/cwbudde/qhull-go` (v0.1.0, tagged & pushed)**; matplotlib-go depends
on it via `go.mod` (local `replace => ../qhull-go` for co-development). The
in-tree `tri/qhull/` copy and its gitignored Qhull oracle were deleted —
connectivity is unchanged by the move, so no goldens changed.

**grid5x4 holdout — closed upstream.** The lone 60/61 order-parity holdout
(historical stage "3c.6f": intermediate ridge-order bookkeeping of a merged quad
across `addPoint`s) is **resolved in qhull-go → 34/34 cocircular, 61/61 order**
(faithful `qh_findbestnew` + a non-simplicial `findbestnew` flag).

**Open in matplotlib-go: none.** Further qhull-go work (publishing polish, doc
freeze, optional richer result type, dropping the local `replace`) is tracked in
that repo's own `PLAN.md`.

---

## Phase 13: Cached Transformed Paths in the Render Path (deferred) 🔁

**Goal:** make artists (`Line2D`, patches, collections) draw through a
persistent `transform.TransformedPath` so a redraw that changes only the trailing
affine (axes resize/pan/zoom) reuses the cached non-affine projection instead of
re-running it per vertex — and an unchanged redraw skips the per-vertex transform
entirely. This realizes, in the renderer, the affine/non-affine cache split that
Phase 9 already built into `TransformedPath`.

**Status:** **Line2D pilot + leg-change detection + patch/collection rollout
shipped** (2026-06-28). The only piece still open is the interactive/animation
**redraw loop** that exercises repeated redraws of the same figure (the
single-PNG pipeline never redraws), so the formal exit criterion stays open. What
landed across the three passes:

1. `refreshDataTransform` fires the stage matching the leg, the axes invalidation
   nodes are on `DrawContext`, and `Line2D` draws data-coordinate paths through a
   persistent `transform.TransformedPath` (the pilot).
2. **Leg-change detection:** `refreshDataTransform` now compares the non-affine
   leg structurally (`reflect.DeepEqual`) against the previous draw and fires
   `InvalidNonAffine` only on an actual change, so an unchanged leg reuses the
   projection through a full figure draw — a resize that only moves the axes bbox
   refreshes the trailing affine alone. End-to-end reuse is proven by
   `TestLine2DDisplayPathReusesProjectionThroughRefresh`.
3. **Patch/collection rollout:** the shared `buildArtistDisplayPath` routes
   data-coordinate draws through a centralized `displayPathCache`. Single-shape
   patches inherit one cache via an interface method promoted from the embedded
   `Patch`; the three collections keep one cache per element (`Collection`
   grows a `[]displayPathCache`). Sources are rebuilt fresh each draw (collections
   per element), so change detection is value-based (`pathsEqualValue`).

Parity is held at RMSE 0 by gating the cache to **genuine non-affine legs only**
(linear axes keep the direct path — see the ULP note below) and by applying the
trailing affine **per axis** so a vertex outside the data domain (NaN under a log
scale) keeps NaN local to one coordinate, exactly like the direct separable chain.

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

- **`refreshDataTransform` fired the wrong stage** (✅ fixed). It used to
  invalidate `dataNode` with `InvalidAffine` for every change, including the
  non-affine-leg branch (log/symlog, polar/geo/3D), so a split-aware consumer
  would reuse a **stale projection** on a log-domain/limits change. Now fires
  `InvalidNonAffine` for non-affine legs and `InvalidAffine` only for the
  affine-leg/bbox case (`TestRefreshDataTransformInvalidationStage`).
- **Artists hold no axes/node reference** (✅ addressed for `Line2D`). Artists get
  their transform only from the per-draw `DrawContext`; the invalidation nodes are
  now plumbed through `ctx.dataTransformDeps()` (`axesBbox.Node()` + `dataNode`),
  which `Line2D` wires its `TransformedPath` to. The shared
  `core/patch_paths.go:buildArtistDisplayPath` is not yet converted (deferred).
- **Segmentation must stay on final points** (✅ done). `Line2D.displayPath`
  interleaved NaN-segmentation with the per-vertex transform; the cached path
  applies the trailing affine to the cached non-affine points, then runs the
  existing finiteness/segmentation loop on the final points (`segmentFinitePath`).
  Byte-identical: a single finiteness check on the final point matches the direct
  pre+post checks, validated by `TestLine2DDisplayPathCacheParity`.
- **Fully-affine collapse must not be cached** (parity gate). For linear axes the
  trailing affine collapses scale+bbox into one matrix and `M.Apply(v)` diverges
  from the direct `bbox.Apply(scale.Apply(v))` by a ULP (and there is no
  projection to cache). `cachedDisplayPoints` therefore engages only when
  `transform.AsAffine` reports a non-affine remainder; linear axes keep the direct
  path. For a genuine non-affine leg the trailing affine is the single bbox matrix,
  which matches the direct chain exactly.
- **Log/polar resize win via leg-change detection** (✅ done). `refreshDataTransform`
  compares the rebuilt non-affine leg structurally against the previous draw
  (`reflect.DeepEqual`, since legs are value transforms rebuilt fresh each draw)
  and fires `InvalidNonAffine` only on a real change. `Log`/`Linear` encode the
  view limits, so a limit change alters the leg (re-project) while a pure resize
  leaves it untouched (reuse the projection, refresh only the bbox affine).
  Func-based scales compare unequal and conservatively re-project.
- **NaN must stay axis-local in the trailing affine** (parity gate, patch
  rollout). The cached path applies the trailing affine; doing so via a full 2×3
  matrix lets the zero cross-term poison a finite axis when the other is NaN
  (`0*NaN == NaN`), so a log-domain-outside vertex `{10, NaN}` (direct, per-axis)
  became `{NaN, NaN}`. `Line2D` never saw this (it culls non-finite points during
  segmentation) but patches keep every vertex. Fix: `applyTrailingAffinePath`
  applies a shear-free affine (the separable bbox) **per axis**, byte-identical to
  the direct separable chain for finite coords and NaN-correct otherwise.

### Remaining work (if revived)

- [x] Refine `refreshDataTransform` to fire `InvalidNonAffine` for non-affine data
      legs (and `InvalidAffine` otherwise). `core/axes_transform_test.go` stays
      green; `TestRefreshDataTransformInvalidationStage` asserts a log leg fires
      non-affine and an affine leg fires affine.
- [x] Expose the axes invalidation nodes (`axesBbox`/`dataNode`) on `DrawContext`
      (`ctx.dataTransformDeps()`).
- [x] Pilot on `Line2D`: persistent `transformedPath` field, `SetPath` on data
      change, dependency-wire to the nodes, draw via `TransformedPointsAndAffine` +
      the unchanged segmentation loop. Golden/reference RMSE 0 (cache gated to
      non-affine legs). Reuse proven by `TestLine2DDisplayPathReusesNonAffineProjection`.
- [x] Extend the same pattern through `buildArtistDisplayPath` to patches and
      collections (centralized `displayPathCache`; one slot per `Patch`, one per
      collection element). Parity gate re-run: `TestBuildArtistDisplayPathCacheParity`
      (incl. log-domain NaN vertices), `TestCollectionPerElementCacheReuse`.
- [x] Non-affine-leg change detection so log/polar resizes also reuse the
      projection and the reuse fires through a full figure draw, not just a direct
      `displayPath` call (`reflect.DeepEqual` leg compare in `refreshDataTransform`;
      `TestRefreshDataTransformInvalidationStage`,
      `TestLine2DDisplayPathReusesProjectionThroughRefresh`).

**Out of scope:** passing the affine _separately_ to the agg backend so it
composes with the device y-flip and does zero per-vertex affine work
(matplotlib's deepest win) — needs a `render.Renderer` capability/signature
change.

**Exit criterion:**

- [~] Mechanism complete; only the redraw loop is missing. Reuse now fires through
  a **full figure draw**: with leg-change detection the every-draw
  `refreshDataTransform` no longer re-projects an unchanged non-affine leg, so a
  resize that only moves the bbox reuses the cached projection for `Line2D`,
  patches, and collections (`TestLine2DDisplayPathReusesProjectionThroughRefresh`,
  `TestBuildArtistDisplayPathReusesProjection`, `TestCollectionPerElementCacheReuse`).
  Golden+reference suite stays RMSE 0 (no new failures vs. the pre-existing branch
  baseline). Still open: an interactive/animation redraw loop (or public re-render
  API) that actually issues repeated figure draws — the single-PNG pipeline draws
  once, so the win is currently exercised only by tests, not end users.

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
