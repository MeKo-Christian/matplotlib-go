# Matplotlib-Go Development Plan

This plan tracks the **remaining** work to bring `matplotlib-go` to a stable
v1.0 release. The foundation is built; what follows is the focused path to the
finish. The roadmap is cross-checked against the local upstream Matplotlib
snapshot in `third_party/matplotlib` so uncovered areas stay explicit instead of
sliding into a vague "future work" bucket.

Phases are ordered **closed first, open last**: Phases 1–9 are complete; the
remaining open work — Skia GPU (10), the v1.0 release stretch (11), two deferred
infrastructure phases (12–13), the second-review closure phases (14–18, from
the 2026-06-30 [`REVIEW.md`](REVIEW.md)), and the third-audit closure phases
(19–21, from the 2026-07-01 review appended to the same file) — is collected at
the end under **Remaining Work**, **Second Fidelity-Review Closure**, and
**Third-Audit Closure & Pre-v1.0 Break**. Phase 11 (v1.0 release) executes
**last**: Phases 19–21 include a breaking API rework and tolerance re-freeze
that the release must postdate.

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

## Parity Follow-up: Sketch figure-patch fill-coverage (done) ✅

**Status:** done (2026-06-28). `sketch_xkcd` reference-compare **RMSE 3.5 → 0.54**
(MaxDiff 255 → 65, MeanAbs 0.03 → 0.01, PSNR 64.4 dB); `TestGolden` is now a perfect
in-memory match (PSNR +Inf). No regressions: the change is gated to sketch-active
renders, so every other golden is byte-identical. Catalog tolerance tightened to
`MinPSNR 62 / MaxMeanAbs 0.05 / MaxRMSE 1.0` (was 60 / 0.1 / 4.0).

**Root cause (was):** the residual was **42 fully-transparent (255,255,255,α=0)
border pixels** in the Matplotlib reference. Matplotlib applies the global
`path.sketch` to the **figure background patch** (`figure.patch`) drawn on a
**transparent** RGBA buffer; its wiggled fill leaves those pixels uncovered. Go
painted the background as the raster backend's **opaque clear** (`agg.New(w,h,white)`),
which a sketch can never perforate.

**The missing insight (why the earlier attempt failed):** Matplotlib draws
`figure.patch` with **`antialiased=False`** (confirmed by instrumenting
`RendererAgg.draw_path`: `gc.get_antialiased()==0`). Binary/scanline-bin coverage is
what makes the notches **hard 2px α=0 holes with zero partial-alpha pixels** — the
reference has no gradient at all. The first attempt filled the patch **antialiased**,
producing a soft ~45px coverage gradient (~1000 semi-transparent px) instead of the
sharp notches, which is why it regressed. There was **no `../agg_go` fill-coverage
bug**; the fix is simply to fill the patch with AA off. Go's `sketch.Apply` on the
patch was already byte-exact (all 42 notch pixels land at the exact reference
coordinates — see `backends/agg/figure_patch_sketch_test.go`).

**What shipped:**

1. `render.TransparentClearer` capability + `agg.Renderer.ClearTransparent()` (clears
   the buffer to straight white-transparent without disturbing the draw/clip stack).
2. `core.drawSketchedFigurePatch` (in `figure_draw.go`): when the global sketch is
   active, clear transparent and fill `pixelRectPath(vp)` with the figure facecolor at
   **`AntialiasOff`** (the renderer's default sketch wiggles it). Gated to
   sketch-active, so non-sketch figures keep the opaque-clear path byte-for-byte.
3. Alpha contract: added `agg.Renderer.GetImageNRGBA()` — the straight-alpha buffer
   correctly labeled as `image.NRGBA` (vs `GetImage()`'s `image.RGBA`, whose Go
   contract is premultiplied). `sketch_xkcd`'s `Render()` returns it so the
   transparent notch pixels round-trip through `png.Encode`/`ComparePNG`'s
   premultiply consistently with the reference. Other examples are fully opaque
   (RGBA==NRGBA), so `GetImage()` was left unchanged (244 callers, incl. the Skia
   native bridge, rely on its `*image.RGBA` layout).

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

## Phase 11: Documentation, Performance, and v1.0 Release (LAST — after Phases 16–21)

**Goal:** make the project consumable by users who have not followed the
development thread, establish performance baselines, and tag a stable v1.0.

> **Ordering (2026-07-01):** this phase executes **last**. Phase 20 is a
> deliberate pre-v1.0 breaking rework of the public API, so the "Public API
> stability audit ✅" checkbox below froze the _pre-break_ surface — Phase 20.4
> re-freezes it. The final golden/tolerance freeze must postdate Phases 19 and
> 21, and the changelog must include Phase 20's breaking section.

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

# Phases 14–18: Second Fidelity-Review Closure (2026-06-30)

These phases are derived from the second independent fidelity review in
[`REVIEW.md`](REVIEW.md) (2026-06-30), an adversarial subagent audit run after the
Phase 4–9 breadth work. Its verdict: matplotlib-go is **a faithful port with a
scaffolding fringe, not a facade** — the numerical cores, the mathtext engine, the
Qhull triangulation, and the parity harness (which compares against _real_
matplotlib output and gates CI) are genuinely faithful. The remaining problems are
**concentrated and locatable**: a handful of silent-degradation footguns that
survived Phase 4, the secondary backends and the capability layer that advertises
more than it delivers, ~52 parsed-but-ignored rcParams, a few concrete algorithmic
bugs, and two holes in the parity harness. Phases 14–18 close those.

> **Note on overlap with Phase 4/8:** Phase 4 claimed the silent-failure class was
> closed, but the 2026-06-30 audit found several paths still silent (unknown
> colormap still falls back to viridis with only a debug warn; `Setp` still
> swallows unknown keys). Phase 8's vector-`MeasureText` fix is **confirmed real**.
> These phases pick up the genuine residuals, not the parts already done.

## Phase 14: Silent-Failure Hardening, Round 2 🧪

**Goal:** turn the remaining silent-wrong-output paths into loud errors or honest
warnings. This is the highest-value, lowest-cost bucket — each item is a small diff
with direct impact on anyone porting a real matplotlib script.

**Status (2026-07-01):** the four silent-degradation footguns plus the
`Text.Contains` metrics fix are **shipped**; the two genuinely-larger items (the
case-folding rework and the blit redraw) are descoped to Phase 17/15 with the
rationale recorded below. Parity suite green (zero golden regressions); new unit
tests cover the colormap strict variant and the `Setp` warning.

- [x] `color/colormap.go` — added `GetColormapStrict(name) (Colormap, error)` +
      exported `ErrUnknownColormap`; `GetColormap` now delegates to it and keeps the
      warn-then-viridis lenient fallback. Tests: `TestGetColormapStrict*`. _Deferred to
      Phase 17:_ stopping case-folding (`"Blues"=="blues"`) needs the builtin maps
      re-registered under canonical mixed-case names — a larger, parity-risky change
      (documented in the `GetColormapStrict` doc comment).
- [x] `core/introspection.go` — `Setp` now emits a one-shot `diag.Warnf` for
      unrecognized keys (and wrong-typed values) via the existing `SetProperty` bool,
      instead of silently dropping them. Test: `TestSetpWarnsOnUnknownProperty`.
- [x] `backends/pdf/pdf.go` (+`pdf_write.go`) & `backends/ps/procedures.go`
      (+`gradients.go`) — gradient/pattern-filled marker & path-collection items now
      emit a one-shot `diag.Warnf` (`warnGradientCollectionDrop`) when skipped, instead
      of vanishing undrawn. (A real gradient-XObject/procedure fill stays future work.)
- [x] `core/picker_contains.go` — `Text.Contains` now measures the glyph bbox via
      the shared sfnt shaper (`render.MeasureTextMetrics`, per-line width + real
      ascent/descent) using the resolved font key, replacing the `FontSize × rune-count`
      heuristic. Removed the now-orphaned `textRuneCount`.
- [x] `core/image.go` — the rotation-ignoring fallback now emits a one-shot
      `diag.Warnf` (`rotatedImageFallbackWarnOnce`) naming the renderer type; benign on
      AGG (which implements both transform paths), no longer silent-wrong on thin
      backends.
- [ ] `animation/animation.go:578` — the Blit "fast path" calls full `cnv.Draw()`
      anyway (zero blit benefit). **Deferred:** a real overlay redraw needs an
      only-animated canvas `Draw` entry point (`AnimatedFilterOnlyAnimated` plumbed
      through `canvas.FigureCanvas`), which is backend work better tracked with Phase 15;
      dropping the public `cfg.Blit` flag is an API break. Output is already correct —
      this is a perf non-win, not a silent-wrong-output footgun, so it is the lowest
      priority in this bucket. _If the `cfg.Blit` API question is ever acted on, it
      folds into Phase 20's breaking window rather than a standalone break._

## Phase 15: Backend Honesty & Capability Verification 🧪

**Goal:** stop advertising capabilities the backend doesn't actually provide, and
make the verification layer catch the gap. The review's strongest "facade" findings
live here, all off the AGG happy-path.

**Status (2026-07-01):** all nine items shipped. Verification is now behavioral
(unprobeable, non-intrinsic claims fail instead of rubber-stamping); Skia
registration/GPU honesty is test-locked; the pure-Go renderer grew real
round/bevel joins, full paint carry-forward + unsupported-fill warnings; the
glyph-ID→rune cmap resolver replaced the `rune(id)` cast across all five
backends; webagg now emits rubberband + history-button events; the stale gio doc
was removed; and SVG drop-shadow + the documented PGF font limitation closed the
vector-backend gaps. Zero golden regressions.

- [x] `backends/registry.go` — `VerifyRendererCapabilities` is now behavioral: a
      claimed capability must pass its runtime check or be one of three intrinsic,
      always-present capabilities (`AntiAliasing`/`SubPixel`/`PathClip`); any
      unprobeable, non-intrinsic claim is a verification failure rather than a silent
      pass. Guarded by `TestVerifyRendererCapabilitiesRejectsUnprobeableClaim`.
- [x] `backends/skia/` — registration honesty: item 1 now enforces that every
      registered skia capability is actually implemented, and `skia_render_test.go`
      locks that the CPU tier reports the bridged batch caps as `CapabilityBridged`
      (not native) and never advertises `GPUAccel`. `init.go` documents the bridge.
- [x] Skia "GPU" honesty — `GPU()` is gated on a new `BridgeInfo.Accelerated`
      flag that is false for every current bridge (CPU readback + cgo raster
      surface), so it returns false until a real `SkSurface::MakeRenderTarget` path
      lands; GPU _mode_ selection moved to `GPUModeRequested()`/`BridgeInfo().Mode`.
- [x] `backends/gobasic/stroke.go` — `JoinRound` now emits an adaptive arc fan
      and `JoinBevel` a corner triangle via `calculateJoinFiller`, appended as filled
      subpaths. Guarded by `TestJoinFillerGeometry`.
- [x] `backends/gobasic` `pathDevice` — carries the full `Paint` via struct copy
      (no more lossy reconstruction) and emits a one-shot `diag.Warnf` when a paint
      arrives with an unsupported gradient/pattern/composite fill.
- [x] Systemic `GlyphRun` bug — added `render.GlyphIDToRune` (cached cmap
      reverse-lookup); gobasic/agg/svg/pdf/ps/pgf now resolve glyph indices through
      it (one-shot warn on failure) instead of casting `rune(glyph.ID)`. Guarded by
      `TestGlyphIDToRuneRoundTrip`.
- [x] `backends/webagg` — the server now emits `rubberband` (via a new
      `Navigation.SetRubberbandHandler`) during/after zoom drags and `history_buttons`
      (on connect + after toolbar actions). Guarded by `TestRubberbandEmittedDuringZoomDrag`.
- [x] `backends/desktop/gio` — removed the stale `doc.go`; `gio.go` already
      carries the accurate package doc for the real, registered backend.
- [x] `backends/svg/path.go` — the `shadow` filter now emits a true
      `feDropShadow` (offset + blur + opacity) instead of collapsing to a symmetric
      blur. `backends/pgf/text.go` documents the intentional LaTeX-owned font
      selection as a known limitation. Guarded by `TestPathEffectShadowEmitsDropShadow`.

## Phase 16: rcParams Honesty & Coverage 🧪

**Goal:** end the "parse-then-ignore" pattern. Of matplotlib's ~309 rcParams, ~122
are parsed and only ~70 honored — ~52 are stored and never read, giving users false
confidence a setting took effect.

**Status (2026-07-01):** the audit + minimum-bar warning are **shipped**; honoring
the high-value subset and expanding parsing are deferred to a later pass. Zero
golden regressions; new unit tests cover the warn-on-non-default and dedup paths.

- [x] Audited the dead params: **51** parsed-but-never-consumed keys, captured as
      the data-driven source of truth in `style/unhonored.go` (`unhonoredRCParams`):
      `image.*` (6: origin/aspect/resample/interpolation*stage/lut/composite_image),
      `mathtext.*`(9: default/fallback/bf/bfit/cal/it/rm/sf/tt),`date._`(10),
 `pdf._`(4),`ps._`(5),`svg._`(4),`animation._`(11),`boxplot._`    (2: vertical/whiskers). Correction to the prior audit:`boxplot.notch`and
 `boxplot.patchartist` \_are* consumed (`core/plot.go`); their stale "Stored only"
      doc comments were fixed.
- [x] Minimum bar: `maybeWarnUnhonoredRCParam` emits a one-shot `diag.Warnf` (deduped
      per key, process-global) the first time an unhonored rcParam is _set to a
      non-default value_, injected at the `applyMPLStyleEntry` success path so it covers
      `.mplstyle` files, `UpdateParams`, `PushContext`, and `LoadRCFile`. Tests:
      `TestUnhonoredRCParam*` (warn-on-non-default, silent-on-default, dedup, via
      UpdateParams, honored-params-silent, audit guards).
- [ ] Honor the high-value subset where the drawing code already exists:
      `image.origin`/`image.aspect`/`image.resample`, `mathtext.default`/`.rm`/`.it`/…,
      `date.*` tick formatting.
- [ ] **Third-audit unparsed inventory (2026-07-01, REVIEW.md §3 of the third
      review).** Entire families never reach the `RC` struct and land silently in
      `report.Unsupported` (`style/mplstyle.go:977`):
  - ticks: `{x,y}tick.direction`, major/minor `size`/`width`/`pad`,
    `minor.visible`, side toggles (`top`/`bottom`/`left`/`right`), label sides,
    `alignment` — the Go defaults are hardcoded-correct but non-overridable;
  - legend layout (14 params: `loc`, `fancybox`, `shadow`, `numpoints`,
    `scatterpoints`, `markerscale`, `title_fontsize`, `borderpad`,
    `labelspacing`, `handlelength`, `handleheight`, `handletextpad`,
    `borderaxespad`, `columnspacing`);
  - axes: `axisbelow`, `titlepad`, `labelpad`, `titlelocation`, `spines.*`,
    `xmargin`/`ymargin`, `autolimit_mode`, `unicode_minus`,
    `axes.formatter.*` (limits/use_mathtext/useoffset/offset_threshold);
  - lines/markers: `lines.linestyle`, `marker`, `markersize`,
    `markeredgewidth`, the three `*_pattern` keys, `scale_dashes`,
    cap/join styles, `markers.fillstyle`, `lines.antialiased`;
  - figure: `edgecolor`, `frameon`, `subplot.*`, `autolayout`,
    `constrained_layout.*`, `titlesize`/`titleweight`;
  - patches: no `PatchRC` exists at all (`patch.linewidth`/`facecolor`/
    `edgecolor`/`force_edgecolor`/`antialiased`);
  - font: `font.weight`/`style`/`variant`/`stretch` + family-list values;
  - artist-default rc keys: `errorbar.capsize`, `hist.bins` (default fix in
    Phase 19), `scatter.marker`, `scatter.edgecolors`; `contour.*`.
    Prioritize by parity-case impact, not raw count.
- [ ] **Minimum bar for unparsed-but-known keys:** a key present in matplotlib
      3.10.9's rcsetup but not parsed here should trigger the same one-shot
      `maybeWarnUnhonoredRCParam`-style warning instead of landing silently in
      `report.Unsupported`. Ship a known-upstream-key table so silence means
      "genuinely unknown key", never "known key we ignore".
- [ ] **Explicit non-goals (document, don't parse):** `path.snap`,
      `text.hinting`, `polaraxes.grid`, `axes3d.*` — record in the
      unhonored/unparsed report with rationale.

## Phase 17: Artist Breadth & Algorithmic Correctness ⚪

**Goal:** fix the concrete divergences the audit pinned down and fill the narrow
artist gaps that have no workaround.

- [ ] `core/tick_locators.go:683` — `LogLocator` default stride uses
      `ceil(numDecades/numTicks)` vs matplotlib's `numdec//numticks + 1` (off-by-a-
      decade on dense log axes). Port the integer formula; add the "≤1 minor tick →
      AutoLocator" fallback.
- [ ] `core/contour_lines.go:215` — saddle disambiguation keys off `above[0]` with
      fixed corner order and never computes the cell-center mean that mpl2014 uses,
      emitting all four crossings as one polyline. Port the cell-mean saddle split (the
      single most likely source of contour parity drift). Add an ambiguous-saddle
      parity case.
- [ ] 2D `bar(yerr=/xerr=)` — `BarOptions` (`core/plot.go:508`) has no error field
      (only 3D `ErrorBar3D` does). Add error bars to the 2D bar path + an example.
- [ ] `hist(log=)` — add a `Log` field to `HistOptions` + an example exercising it.
- [ ] mathtext `cm`/`stix` fontsets — `core/mathtext.go:208` only remaps the font
      _family_ over the single DejaVu Unicode table; port matplotlib's
      `BakomaFonts`/`StixFonts` per-fontset glyph maps so non-DejaVu fontsets are
      parity-exact (currently only the DejaVu default is). Larger effort; scope first.
- [ ] `core/norm.go` `TwoSlopeNorm` — out-of-range should map to ±inf
      (`np.interp` left/right), not finite extrapolation; align log/logit clip
      semantics (`scale_registry.go:636`) with mpl's `clip` default and `-1000` floor.
- [ ] Transform type-set breadth — add `ScaledTranslation`, `TransformWrapper`, and
      `Affine2D`-style `rotate/skew` builders; the separable→affine extraction is
      diagonal-only (`transform/transform.go:61`), so rotation/shear need manual matrix
      construction today. Triage against real demand before building.
- [ ] Add the ~3 missing accents vs matplotlib's 20-entry `_accent_map` (mathtext
      module).

**Third-audit additions (2026-07-01, REVIEW.md third review §2), all
file:line-verified on both sides:**

- [ ] `core/vector_field_quiver.go:346` — quiver default scale uses
      `mean/(0.18·min(W,H)/√N)`; port matplotlib's `1.8·amean·sn/span` with
      `sn = clip(√N, 8, 25)` (`quiver.py:681`). Changes every default-scaled
      quiver figure — regen goldens.
- [ ] `core/axes_autoscale.go:92` — margins are applied in data space, not
      transform space (`axes/_base.py:3064`), so log/symlog margins are wrong.
      Also drop non-positive limits before log autoscale (`_base.py:3017`) and
      replace the ad-hoc zero-span expansion (`span=1` + linear margin) with
      `nonsingular(expander=0.05·|v|)` semantics.
- [ ] `core/axes_autoscale.go:51` — artists whose bounds are exactly
      `{0,0,0,0}` are skipped, so a single point at the origin is ignored by
      autoscale. Use an explicit has-data flag, not a zero-bbox sentinel.
- [ ] `core/axis_types.go:57` — spine `set_position(('outward', pts))` and
      `(('axes', frac))` are missing; only boundary + data modes exist. Port both
      (the standard detached/centered-spine idioms).
- [ ] `core/date_tick.go:648` — `AutoDateLocator` DAILY interval table is
      `{1,2,4,7,14}` vs matplotlib's `{1,2,3,7,14,21}`. Fix the table. (The full
      rrule alignment simplification stays a documented divergence.)
- [ ] `core/pie.go:245` — pie axes framing is radius-scaled
      (`padding := Radius*1.25`); matplotlib uses a fixed `±1.25 + center` data
      window regardless of radius. Port the fixed framing.
- [ ] `core/layoutgrid.go:239` / `:448` — nested-mosaic constrained_layout is
      not modeled and outside legend/colorbar space is approximated. **Triage:**
      scope the shared-margin modeling cost; if not justified by a parity case,
      close as a documented limitation in `docs/matplotlib-parity-status.md`.
- [ ] `core/axes3d_contour.go:10` — stale "placeholder wireframe contour"
      comment over a real projected-contour implementation; fix the doc.

## Phase 18: Parity-Harness Rigor ⚪

**Goal:** close the two real holes in an otherwise-honest harness so a live-render
regression cannot pass green. **Ordering:** the two harness-hole items (live-render
compare, optional-visual CI) execute **before** Phases 19–21 — those phases
regenerate goldens and rely on the harness to catch regressions honestly.

- [x] `test/helpers_test.go:406` — `TestReferenceCompare` asserts on two committed
      files (golden vs ref); it renders `got` but never compares it. Compare the **live
      render** against the matplotlib reference (or assert `live == golden` in the same
      test) so the chain doesn't depend on `TestGolden` running.
      _Shipped 2026-07-01:_ `runReferenceCompareTest` now asserts `live == golden`
      (≤1 LSB, skipped under `-update-golden`) before the golden-vs-ref compare and
      writes a `_rendered_vs_golden_diff.png` artifact on failure. Verified
      red→green with a doctored golden (old harness passed, new harness fails).
- [x] 49 optional-visual cases skip `TestGolden` (live-vs-golden) in default CI
      (`test/helpers_test.go:69–119`), so a live regression on the 3D/geo/gallery cases
      stays green. Either run their live-render check in CI or document the gap loudly.
      _Shipped 2026-07-01:_ the `RUN_OPTIONAL_VISUAL_TESTS` gate is **removed**
      entirely (all 173 golden cases pass ungated in ~20 s; each render costs
      ~0.05 s — the gate predated the vendored-FreeType default). The two strict
      text cases now always run their strict check (`strictMplRefIDs`); the
      `test-optional-visual` Just target and env var are gone; AGENTS.md updated.
- [x] Tolerance cleanup — the `MinPSNR 10 / MaxMeanAbs 95` overrides on big gallery
      cases never bind (the always-present `MaxRMSE` binds first); remove the redundant
      "theater" thresholds so the catalog reflects the gate that actually applies.
      _Third-audit expansion (2026-07-01):_ `widgets_gallery` and `animation_gallery`
      run with `MinPSNR 10 / MaxMeanAbs 95` — effectively **disabled** gates, not just
      redundant ones. The mechanical removal of the theater thresholds happens here;
      the actual re-tightening of these two plus the ~23 loose cases is owned by
      Phase 21 (visual QA sweep), which does the final ratchet.
      _Shipped 2026-07-01:_ cross-referenced every per-case `MinPSNR`/`MaxMeanAbs`
      against actual golden-vs-ref metrics: **103 override components were theater
      and every one passes the harness defaults (44 dB / 2.50)**, so all were
      removed (rule: drop `MinPSNR ≤ 44` and `MaxMeanAbs ≥ 2.5`); the 50 rows with
      genuinely stricter components keep them (e.g. `units_categories`
      `MaxMeanAbs 0.05` stays while its loose `MinPSNR 43` went). Net effect:
      defaults now bind on 103 rows — a tightening, zero failures. Phase 21's
      PSNR/MeanAbs loose-case queue is therefore **empty**; Phase 21 still owns
      the `MaxRMSE` ratchet (incl. `widgets_gallery` 5.0 / `animation_gallery` 1.5).

**Third-audit additions (2026-07-01, REVIEW.md third review §4):**

- [ ] **Zero-fixture public APIs** — these ship with no parity case or example,
      so a regression is invisible to the harness. Add catalog fixtures (matplotlib
      reference + golden) for: `PSD`, `Specgram`, `Cohere`, `CSD` (HIGH — real
      Welch-segmentation/detrend/window numerics behind them);
      `LogLog`/`SemilogX`/`SemilogY`, `SecondaryYAxis`, `TwinY` (MED);
      `AxLineSlope`, single-series `BoxPlot` (LOW). HIGH group first.

---

# Phases 19–21: Third-Audit Closure & Pre-v1.0 Break (2026-07-01)

Derived from the third audit (2026-07-01, appended to [`REVIEW.md`](REVIEW.md)):
three parallel subagent audits over defaults, algorithms, and API/organization,
excluding everything Phases 14–18 already track. Maintainer decisions: (1) the
**breaking** Go-idiomatic API rework happens **before v1.0** and the API JSON is
re-frozen afterwards; (2) the `core/` god-package is **split** before the
freeze; (3) a **Claude-driven visual QA sweep** closes the loose-tolerance
cases. Fold-ins from the same audit extend Phases 16, 17, and 18.

**Ordering:** Phase 18's harness holes first (so regens are honestly gated) →
the image-affecting fidelity phases 16/17/19 (each regenerates its own goldens;
19 last of the three — it owns the broadest regen sweep) → Phase 20 (whose
regression gate is _goldens byte-identical throughout_, hence after all image
churn) → Phase 21 (QA fixes change pixels; also inspects the renderer in its
final pre-v1.0 state) → Phase 11 (release, LAST). Phase 10 (Skia) is parallel —
it never touches AGG goldens or the core API — and must simply land or be
descoped before the tag.

## Phase 19: Default-Value Fidelity & Golden Regeneration

**Goal:** an unstyled plot must use matplotlib 3.10.9's defaults. Today three
headline defaults diverge and one rc default is dead code. Every fix here moves
Go output _toward_ the committed matplotlib references, so this phase
regenerates goldens (never references) and should tighten tolerances.

- [ ] `lines.linewidth` — two coupled bugs: `style/style.go:333` defaults
      `RC.LineWidth` to **1.25** (mpl: 1.5) and `core/plot.go:87` hardcodes
      `lineWidth := 1.5` without reading `RC.LineWidth` (only the scatter
      edge-width fallback reads it), so `lines.linewidth` in an `.mplstyle` is a
      no-op for lines. Fix the default to 1.5 AND route `plot()` through the rc
      value.
- [ ] hist default bins — `core/histogram.go:285` / `core/plot.go:929` default
      to auto-selection (Sturges <1000 else Scott); matplotlib defaults to fixed
      **10** (rc `hist.bins`, `_axes.py:7033`). Also fix `'auto'` semantics to
      numpy's `min(fd, sturges)` bin width, FD IQR to interpolated (linear)
      percentiles instead of nearest-rank, and Scott to ddof=0. Wire the
      `hist.bins` rc key (Phase 16 cross-ref).
- [ ] scatter default size — `core/scatter.go:802`: unset size renders
      **invisible** (zero area); matplotlib uses `s = 36` pt². Default to 36.
- [ ] minor ticks — size 2.1 → **2.0** (`core/axis_ticks.go:73`); distinguish
      minor pad **3.4** from major 3.5 (`core/axis_types.go:23`).
- [ ] tick-label pad DPI fallback 96 → **100** (`core/axis_ticklabels.go:203`).
- [ ] `PlotOptions` — add a typed linestyle field (`"--"`, `":"`, …); today
      only `Dashes []float64` exists, so the most common mpl idiom has no direct
      spelling. (Use the typed-constant enum style from day one so Phase 20
      doesn't have to re-break it.)
- [ ] **Golden regen pass:** rerun `-update-golden` for affected cases;
      matplotlib references are untouched (they already embody these defaults).
      Assert per-case reference-compare RMSE is non-increasing; ratchet tolerances
      where cases improve.

**Size:** M. **Depends on:** Phase 18's harness-hole items.

**Exit criterion:**

- [ ] Unit tests assert the five default values against matplotlib 3.10.9
      literals; `lines.linewidth` from an `.mplstyle` visibly changes `plot()`
      output; goldens regenerated with no reference-compare regression.

## Phase 20: Go-Idiomatic API Rework & `core/` Split (BREAKING, pre-v1.0)

**Goal:** one coordinated breaking pass — the only one before v1.0 — that makes
the API Go-idiomatic, splits the `core/` god-package (60,369 lines, 173 files,
1,529 exported symbols = 51% of the 3,019-symbol public surface), and
re-freezes the API afterwards. **Rendering behavior is untouched: goldens stay
byte-identical throughout — that invariant is the phase's regression gate**
(which is why Phase 19's golden churn must land first).

Regen workflow (after every stage, final freeze at the end):
`UPDATE_PUBLIC_API_AUDIT=1` regenerates
`test/testdata/public_api/stable_public_api.json`; the
`internal/examplecatalog/public_surface_parity.go` local-API strings are
remapped and `docs/matplotlib-parity-status.md` regenerated
(`go run ./cmd/paritystatusdoc`; guards `TestAllUpstreamPublicRowsAreClassified`,
`TestPartialAndOmissionRowsHaveNotes`); doc-coupled string tests
(`stable_test_names_test.go`, `apidoc_coverage_test.go`,
`TestStablePublicPackagesHaveGoDocAndExamples`) updated in the same commit as
each move.

### 20.0 Ship-first, non-breaking (can land immediately, even before Phase 19)

- [x] **DATA RACE:** `color/colormap.go:388` `RegisterColormap` mutated the
      package-global `colormaps` map with no mutex while draws read it.
      **Shipped 2026-07-01:** `colormapMu sync.RWMutex` guards every map access
      (`RegisterColormap` write-locks; the `GetColormap` fallback,
      `GetColormapStrict`, and `ColormapNames` read-lock); the two `init()`
      writers (`petroff.go`, `listed_colormaps.go`) now route through
      `RegisterColormap` so the every-access-holds-the-lock invariant is
      greppable. Guarded by `TestColormapRegistryConcurrentAccess`, which fails
      under `-race` without the lock (verified red first).

### 20.1 Surface tiering (decision stage, no code)

- [ ] Classify all 3,019 frozen symbols keep / demote / delete in a design doc.
      Demotion candidates from the audit: the introspection cluster
      (`Setp`/`Getp`/`GetpAll`/`Findobj`/`FindobjType` — Python-isms duplicating
      typed setters), the `*Units` any-variants (superseded by the unified error
      convention in 20.3), and render-extension interfaces only backends consume
      (→ tiered or internal). Symbols marked delete are removed during 20.2,
      never moved.

### 20.2 Package split (mechanical stage)

- [ ] `core/axes3d*.go` + 3D projection files → new `plot3d` package
      (~98 exported symbols, ~7k lines).
- [ ] `core/tick_locators.go` (1,149) + `core/tick_formatters.go` (1,045) +
      `core/date_tick.go` (968) → new `ticker` package (~3,162 lines), mirroring
      matplotlib's own `ticker`/`dates` boundary where natural.
- [ ] `core/widget_*.go` + selector files → new `widgets` package (~2.5k
      lines; conceptually belongs beside `canvas/`).
- [ ] Gate: `go build ./... && just test` green, goldens **byte-identical**,
      full regen workflow run. Refresh the stale
      `docs/large-file-decomposition.md` snapshot (plot.go +256 lines,
      mplstyle.go +724 since recorded) and re-run `just large-file-audit`.

### 20.3 Idiomatic conventions (semantic stage, per-package after the split)

- [ ] **One error convention.** Today plot methods return artist-only +
      `diag.Warnf` on bad input (16 sites, e.g. `core/plot.go:263,755,821,896,1016`)
      while `*Units` variants return `(T, error)`. Adopt `(T, error)` everywhere;
      fold the `*Units` variants into the primary methods (per 20.1);
      `diag.Warnf` remains for degradations only, never for rejected input.
- [ ] **Options model.** 83 Options structs, 408 pointer-to-primitive optional
      fields, and `opts ...FooOptions` that silently uses only the first element.
      Decide one pattern (single-struct arg or functional options), make extra
      variadic elements impossible or an error, and replace string enums
      (`Orientation`, `LineStyle`, `Location`, `Where`, `Colormap`,
      `Interpolation *string` at `core/image.go:68`) with the existing
      typed-constant pattern (`SignalSpectrumScale` style).
- [ ] **io.Writer surface.** Add `Figure.Save(path)` / `Figure.WriteTo(w,
format)` / `Figure.Image()`; delete the hand-rolled 8-line agg dance from
      every example (the 3,019-symbol surface currently has zero `io.Writer`).
- [ ] **Naming.** `GetX()` getters → `X()`; resolve exported-mutable-field vs
      setter duplication (pick one per type); drop/demote Python-ism names per
      20.1.
- [ ] **Concurrency contract.** Document what is and is not safe (global rc
      state, registries, Figure); pyplot examples stop discarding errors under
      the new convention.
- [ ] **Dedup (falls out of the redesign):** shared alpha-baking helper (~62
      duplicated sites), option-unpack helper (replaced by the options model),
      single scalar-map resolution path (Scatter2D vs Collection).

### 20.4 Re-freeze

- [ ] Final `UPDATE_PUBLIC_API_AUDIT=1` regen + classification remap + parity
      status doc regen; migration notes for every break in
      `docs/matplotlib-migration-notes.md`; CHANGELOG "breaking" section drafted
      for Phase 11.

**Size:** XL (largest remaining phase; 20.2 and each 20.3 bullet are
independently executable sessions once 20.1's doc exists).
**Depends on:** Phase 19 (goldens must be stable so byte-identical is the gate).

**Exit criteria:**

- [ ] `core/` no longer contains plot3d/ticker/widgets; every plot method
      returns `(T, error)`; no raw-string enum fields in options; `Figure.Save`
      exists and examples use it; `stable_public_api.json` re-frozen and the CI
      audit green; goldens byte-identical to the Phase 19 baseline.

## Phase 21: Claude-Driven Visual QA Sweep (loose-tolerance closure)

**Goal:** visual inspection of every case whose gate is loose enough to hide a
real divergence — including the "RMSE passes but the output doesn't look right"
class — ending in a per-case disposition and a final tolerance ratchet that
Phase 11 freezes. Runs after 16/17/19 (which change images) and after 20 (whose
golden-stability gate must not be disturbed).

- [ ] **Re-arm the disabled gates:** `widgets_gallery` and `animation_gallery`
      (`MinPSNR 10 / MaxMeanAbs 95`) get real, binding thresholds after visual
      review (the mechanical removal of the theater thresholds is Phase 18's).
- [ ] **MaxRMSE ≥ 4 queue (~23 cases):** side-by-side vs the matplotlib
      reference, classify the residual (real divergence → fix; acceptable
      rasterization/text difference → documented exception; "the Python original
      looks bad too" → upstream-difference note), then ratchet: `skewt_basic`,
      `annotation_legend_offsetbox_gallery`, `text_bbox_styles`,
      `text_layout_gallery`, `mathtext_accents`, `mathtext_fractions`,
      `mathtext_inline_labels`, `projection_toolkit_gallery`, `specialty_depth`,
      `formatter_log_mathtext_labels`, `named_colors`, `stem_plot`,
      `stem_horizontal`, `lognorm_imshow`, `colorbar_variants_gallery`,
      `errorbar_capthick`, `patch_style_matrix`, `text_annotation_matrix`,
      `scatter_gallery`, `animation_subplots_frame`, `animation_line_frame`,
      `animation_scatter_frame`, `mplot3d_stem3d`, `scale_logit_ticks`.
- [ ] **Low-PSNR, low-RMSE dense galleries** (where RMSE alone is a weak
      gate): `mathtext_gallery` (PSNR 16 / MeanAbs 35), `image_variants_gallery`,
      `triangulation_gallery`, `pcolormesh_gouraud` — inspect and add a binding
      PSNR floor where warranted.
- [ ] **Deliverable:** a per-case disposition table (fixed / exception with
      rationale / upstream-difference) committed alongside the catalog; every
      reviewed case's tolerance ratcheted to actual + small headroom.

**Size:** L (breadth-heavy, low per-case depth; parallelizable by gallery).
**Depends on:** Phases 16, 17, 19 (image-affecting) and 20 (freeze stability).
**Feeds:** Phase 11's "final golden/reference regeneration pass with per-case
tolerances frozen for v1.0".

**Exit criterion:**

- [ ] No catalog case has an effectively-disabled gate; every case with
      MaxRMSE ≥ 4 has a written disposition; the tolerance set handed to Phase 11
      is the ratcheted one.

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
(CPU native primitives are done); **Phases 12 & 13** track genuinely-deferred
infrastructure depth (cocircular Qhull parity, renderer-wired cached transformed
paths); **Phases 14–18** close the second fidelity review (2026-06-30); and
**Phases 19–21** close the third audit (2026-07-01) — default-value fidelity,
the one coordinated pre-v1.0 breaking pass (Go-idiomatic API rework + `core/`
split + re-freeze), and the visual QA sweep. **Phase 11** is the final v1.0
stretch and executes **last**: documentation and performance work is done, but
the release mechanics (changelog, CI gate, final freeze, v1.0 tag) must postdate
the Phase 20 re-freeze and the Phase 21 tolerance ratchet.
