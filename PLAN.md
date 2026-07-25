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

# Phases 1–9: Completed Parity Foundation ✅

All are complete; detailed implementation logs remain available in git history.

- **Phase 1 — Visual parity:** all catalog cases closed below RMSE 5.
- **Phase 2 — Reporting:** generated parity matrix plus CI coverage guards.
- **Phase 3 — Decomposition:** oversized Go files split and audit-locked.
- **Phase 4 — Correctness:** silent failures now handle or diagnose lost intent.
- **Phase 5 — Fidelity:** formatters, layout, dates, margins, and axes completed.
- **Phase 6 — Text:** MathText coverage, fallback, accents, and boxes completed.
- **Phase 7 — Plot breadth:** key artist, contour, colorbar, and image gaps closed.
- **Phase 8 — Backends/style:** metrics, effects, vectors, rcParams, cyclers done.
- **Phase 9 — Infrastructure:** geometry, triangulation, transforms, lifecycle done.

# Remaining Work (Open Phases)

The phases below still carry open to-do items and are ordered after the closed
phases. **Phase 10** finishes Skia GPU mode; **Phase 11** is the v1.0 release
stretch (docs and performance are done, only release mechanics remain).
**Phases 12 and 13 are complete**; their compact records stay below for
provenance.

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

**Completed 2026-06-27.** The Qhull-faithful Delaunay engine now reproduces general and cocircular construction order and lives in `github.com/cwbudde/qhull-go` v0.1.0; `tri` consumes it, and further work is tracked in that repository.

---

## Phase 13: Cached Transformed Paths in the Render Path ✅

**Goal:** make artists (`Line2D`, patches, collections) draw through a
persistent `transform.TransformedPath` so a redraw that changes only the trailing
affine (axes resize/pan/zoom) reuses the cached non-affine projection instead of
re-running it per vertex — and an unchanged redraw skips the per-vertex transform
entirely. This realizes, in the renderer, the affine/non-affine cache split that
Phase 9 already built into `TransformedPath`.

**Status:** **complete** (2026-07-24). The Line2D pilot, leg-change detection,
patch/collection rollout, and the interactive animation redraw loop are shipped.
WebAgg's animation path now captures a static background once and repeatedly
drives `DrawFigureWithOptions(..., AnimatedFilterOnlyAnimated)` into the existing
renderer before blitting. Unsupported canvases use a correct repeated full-draw
fallback. This gives end users the persistent redraw loop the cache mechanism
previously exercised only in tests. What landed across the four passes:

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
4. **Redraw loop:** animation blitting now restores the captured background,
   redraws only animated artists through `canvas.AnimatedDrawCanvas`, and presents
   the damaged region. WebAgg implements the overlay draw against its persistent
   renderer; non-blitting canvases repeat full draws without losing
   animation-managed artists.

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
  `transData` with `dataNode`) drives cache invalidation across redraws.

### Key insights (why it's more than a drop-in)

- **`refreshDataTransform` fired the wrong stage** (✅ fixed). It used to
  invalidate `dataNode` with `InvalidAffine` for every change, including the
  non-affine-leg branch (log/symlog, polar/geo/3D), so a split-aware consumer
  would reuse a **stale projection** on a log-domain/limits change. Now fires
  `InvalidNonAffine` for non-affine legs and `InvalidAffine` only for the
  affine-leg/bbox case (`TestRefreshDataTransformInvalidationStage`).
- **Artists hold no axes/node reference** (✅ addressed). Artists get
  their transform only from the per-draw `DrawContext`; the invalidation nodes are
  now plumbed through `ctx.dataTransformDeps()` (`axesBbox.Node()` + `dataNode`),
  which `Line2D` and the shared patch/collection display-path cache wire their
  `TransformedPath` instances to.
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

### Completed work

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
- [x] Add the repeated redraw loop: `canvas.AnimatedDrawCanvas` plus WebAgg's
      persistent-renderer implementation restore one captured background, draw
      only animated artists, and blit each frame. Unsupported canvases fall back
      to full draws that temporarily include animation-managed artists.

**Out of scope:** passing the affine _separately_ to the agg backend so it
composes with the device y-flip and does zero per-vertex affine work
(matplotlib's deepest win) — needs a `render.Renderer` capability/signature
change.

**Exit criterion:**

- [x] Mechanism and redraw loop complete. Reuse fires through a **full figure
      draw**: with leg-change detection the every-draw
      `refreshDataTransform` no longer re-projects an unchanged non-affine leg, so a
      resize that only moves the bbox reuses the cached projection for `Line2D`,
      patches, and collections (`TestLine2DDisplayPathReusesProjectionThroughRefresh`,
      `TestBuildArtistDisplayPathReusesProjection`, `TestCollectionPerElementCacheReuse`).
      The WebAgg animation loop now issues repeated overlay-only figure draws against
      a persistent renderer, with full-redraw fallback for other canvases. Golden and
      reference output remain unchanged.

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

## Phase 14: Silent-Failure Hardening, Round 2 ✅

**Completed 2026-07-24.** Silent degradations now fail or warn honestly, text
picking uses real metrics, and animation blitting performs true overlay redraws.
Focused tests cover every item; parity remained green.

- [x] Added strict colormap lookup (`GetColormapStrict`,
      `ErrUnknownColormap`) while retaining the warned viridis fallback. Canonical
      case-sensitive map names remain deferred to Phase 17.
- [x] Added one-shot diagnostics for invalid `Setp` values, unsupported PDF/PS
      gradient collections, and renderer fallbacks that ignore image rotation.
- [x] Replaced `Text.Contains`' rune-count estimate with shaped multiline glyph
      metrics.
- [x] Implemented real WebAgg blitting: capture the static background, restore
      it, draw only animated artists, then blit; unsupported canvases use a
      correct full-redraw fallback.

## Phase 15: Backend Honesty & Capability Verification ✅

**Completed 2026-07-01.** Backend claims are behaviorally verified, secondary
backends report their real implementation tier, and known degradations are
explicit. All nine audit findings closed with zero golden regressions.

- [x] Capability verification rejects unprobeable non-intrinsic claims; Skia
      reports bridged CPU features accurately and advertises GPU acceleration
      only when `BridgeInfo.Accelerated` is true.
- [x] GoBasic gained real round/bevel joins, lossless `Paint` carry-forward, and
      warnings for unsupported gradient/pattern/composite fills.
- [x] Added cached cmap-based `render.GlyphIDToRune`; AGG, GoBasic, SVG, PDF,
      PS, and PGF no longer cast glyph IDs directly to runes.
- [x] WebAgg emits rubberband and history-button events; stale Gio documentation
      was removed.
- [x] SVG shadows use `feDropShadow`; PGF documents its intentional
      LaTeX-controlled font-selection limitation.

## Phase 16: rcParams Honesty & Coverage ✅

**Completed 2026-07-25:** audited 51 dead keys; warned on unsupported non-default
values; classified intentional non-goals; and honored the targeted rcParams with
explicit-option precedence and zero golden churn. Audit registries and rationales
live in `style/{unhonored,unparsed}.go`.

- [x] Honored tick, legend, axes label/title, scalar formatter, and spine
      defaults.
- [x] Honored line and marker styles across lines, errorbars, collections, and
      contours.
- [x] Applied figure and layout defaults; documented GUI-only controls as
      headless non-goals.
- [x] Applied patch defaults across standalone patches and plot-generated
      patches.
- [x] Preserved font properties and ordered generic-family fallback lists
      through renderer selection.
- [x] Applied contour defaults across structured and triangular contour paths.
- [x] **MathText and exit gate:** all nine settings are honored; audit guards,
      focused tests, API audit, and pinned-FreeType parity suites pass.

## Phase 17: Artist Breadth & Algorithmic Correctness ⚪

**Goal:** fix the concrete divergences the audit pinned down and fill the narrow
artist gaps that have no workaround.

**Completed:** corrected log tick selection, contour saddle splitting,
`TwoSlopeNorm`/log clipping, quiver scaling, transformed-space autoscaling, and
zero-bound handling.

- [x] Added 2D bar errorbars and logarithmic histograms with parity cases.
- [x] Ported CM/STIX MathText fontsets and the complete accent set.
- [x] Added outward/axes-relative spine positioning.
- [x] Fixed radius-independent pie framing.
- [x] Verified the daily date interval tables already match Matplotlib 3.10.9.

### Remaining work

- [ ] Transform type-set breadth — add `ScaledTranslation`, `TransformWrapper`, and
      `Affine2D`-style `rotate/skew` builders; the separable→affine extraction is
      diagonal-only (`transform/transform.go:61`), so rotation/shear need manual matrix
      construction today. Triage against real demand before building.
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

## Phase 19: Default-Value Fidelity & Golden Regeneration ✅

**Completed 2026-07-02.** Matplotlib 3.10.9 defaults now govern line width, histogram bins, scatter size, minor ticks, and DPI fallback; typed line styles also landed. Focused tests pin the values, while all 173 goldens and reference comparisons remained unchanged and green.

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
(CPU native primitives are done); **Phases 12 and 13** record the completed
cocircular Qhull and renderer-wired transformed-path work; **Phases 14–18** close
the second fidelity review (2026-06-30); and
**Phases 19–21** close the third audit (2026-07-01) — default-value fidelity,
the one coordinated pre-v1.0 breaking pass (Go-idiomatic API rework + `core/`
split + re-freeze), and the visual QA sweep. **Phase 11** is the final v1.0
stretch and executes **last**: documentation and performance work is done, but
the release mechanics (changelog, CI gate, final freeze, v1.0 tag) must postdate
the Phase 20 re-freeze and the Phase 21 tolerance ratchet.
