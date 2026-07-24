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

**Goal:** turn the remaining silent-wrong-output paths into loud errors or honest
warnings. This is the highest-value, lowest-cost bucket — each item is a small diff
with direct impact on anyone porting a real matplotlib script.

**Status (2026-07-24):** all six scoped items are shipped. The colormap
case-folding rework remains explicitly deferred to Phase 17; the animation blit
path now performs real overlay-only redraws through WebAgg. Parity remains green,
with focused tests covering every shipped behavior.

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
- [x] `animation/animation.go` / `canvas/scheduler.go` / `backends/webagg` —
      `Blit=true` now captures one non-animated background, restores it, draws only
      animated artists through the optional `AnimatedDrawCanvas`, and blits without
      a full redraw. The first frame includes its overlay; unsupported canvases take
      a correct full-redraw fallback instead of dropping animation-managed artists.
      Focused animation, WebAgg, and sketched-background tests lock the behavior.
      _Shipped 2026-07-24._

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

**Status (2026-07-02):** audit, both minimum-bar warnings (unhonored-parsed AND
known-but-unparsed incl. non-goal rationale), the high-value honoring subset
(image origin/aspect, all date._), and three parsed-inventory families (axes
behavior, lines artist defaults, scatter/errorbar artist defaults) are
**shipped** — all with zero golden churn (only non-default rc values change
behavior). Open: the remaining unparsed families below (ticks, legend layout,
figure, patches, font, contour._, axes.formatter.\*, spines) and the
mathtext honoring blocked on a cwbudde/mathtext library hook.

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
- [x] Honor the high-value subset where the drawing code already exists
      (2026-07-02, zero golden churn):
  - `image.origin`/`image.aspect` are consumed by the imshow front-ends
    (`core/matrix_helpers.go`); MatShow keeps hardcoded upper/equal like
    matplotlib matshow. `image.resample` stays unhonored **by design**: the Go
    raster path has no adaptive downscale resampling engine to toggle (it
    structurally matches mpl's `resample=False` branch only) — noted in
    `style/unhonored.go`.
  - all ten `date.*` keys: `date.epoch` resolves lazily on first conversion
    (mirrors `get_epoch()`), `date.interval_multiples` drives
    `DateLocator` (incl. the `{1,2,3,7,14,21}` daily table for False),
    `date.autoformatter.*` override AutoDateFormatter buckets via a new
    strftime interpreter (`core/strftime.go`), `date.converter: concise`
    switches the default axis formatter.
  - `mathtext.default`/`.rm`/`.it`/… remain **blocked upstream**: the
    `cwbudde/mathtext` engine (v0.4.4) exposes no hook for the implicit-italic
    default or per-class family names (`parser.go` hardcodes them;
    `Options`/`FontResolver` carry no fontset table). Honoring these needs a
    mathtext library release first; `mathtext.cal`/`.bfit`/`fallback` further
    depend on the Phase 17 per-fontset glyph maps. Keys stay in the
    unhonored registry so the one-shot warning keeps firing.
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
  - axes: ~~`axisbelow`, `xmargin`/`ymargin`, `autolimit_mode`,
    `unicode_minus`~~ — **shipped 2026-07-02** (`RC.Axes` group, parsed +
    consumed: axes seed `SetAxisBelow`/margins/`autolimitMode` from rc at
    creation in `applyRCBehaviorDefaults`, `scalarFixMinus` consults
    `unicode_minus`; zero golden churn — only non-default rc values apply).
    Still open: `titlepad`, `labelpad`, `titlelocation`, `spines.*`,
    `axes.formatter.*` (limits/use_mathtext/useoffset/offset_threshold);
  - lines/markers: ~~`lines.linestyle`, `marker`, `markersize`,
    `markeredgewidth`~~ — **shipped 2026-07-02** (`RC.Lines` group; Axes.Plot
    seeds dash pattern / marker / sizes from non-default rc values, explicit
    options and prop-cycle entries still win). Still open: the three
    `*_pattern` keys, `scale_dashes`, cap/join styles, `markers.fillstyle`,
    `lines.antialiased`;
  - figure: `edgecolor`, `frameon`, `subplot.*`, `autolayout`,
    `constrained_layout.*`, `titlesize`/`titleweight`;
  - patches: no `PatchRC` exists at all (`patch.linewidth`/`facecolor`/
    `edgecolor`/`force_edgecolor`/`antialiased`);
  - font: `font.weight`/`style`/`variant`/`stretch` + family-list values;
  - artist-default rc keys: ~~`errorbar.capsize`, `scatter.marker`,
    `scatter.edgecolors`~~ — **shipped 2026-07-02** (`RC.Scatter`/`RC.Errorbar`
    groups; Axes.Scatter seeds marker/edge color — plus default size from
    `lines.markersize`² — and Axes.ErrorBar seeds cap size; explicit options
    win). Still open: `contour.*`. (`hist.bins` shipped with Phase 19:
    parsed, exported via `Params`, and consumed by `Axes.Hist`.)
    Prioritize by parity-case impact, not raw count.
- [x] **Minimum bar for unparsed-but-known keys:** shipped in
      `style/unparsed.go` — `knownUpstreamRCParams` carries all 321 user-facing
      matplotlib 3.10.9 rcParam keys (generated from `sorted(matplotlib.rcParams)`,
      internal `_internal.classic_mode` dropped); the `applyMPLStyleEntry`
      fallthrough now emits a one-shot `maybeWarnUnparsedRCParam` warning for
      known-but-unparsed keys before recording them in `report.Unsupported`, so
      silence means "genuinely unknown key". Guard tests pin table size and
      consistency with `supportedMPLStyleKeys`/`unhonoredRCParams`
      (`TestUnparsed*`, `TestKnownUpstreamRCParamsConsistency`).
- [x] **Explicit non-goals (document, don't parse):** `path.snap`,
      `text.hinting`/`text.hinting_factor`, `polaraxes.grid`, `axes3d.*` are
      registered in `nonGoalRCParams` (`style/unparsed.go`) with per-key
      rationale (parity-pinned snapping/hinting, unmodeled polar-grid toggle,
      fixed 3D pane/navigation behavior); their one-shot warning carries the
      rationale instead of the generic "not parsed" text.

## Phase 17: Artist Breadth & Algorithmic Correctness ⚪

**Goal:** fix the concrete divergences the audit pinned down and fill the narrow
artist gaps that have no workaround.

- [x] `core/tick_locators.go` — `LogLocator` default stride now uses matplotlib's
      `numDecades/numTicks + 1` (integer floor-div) instead of `ceil(numDecades/numTicks)`;
      the two only diverge when `numDecades` is an exact multiple of `numTicks` (Go and
      mpl share the same floating-point `numdec`, so the common few-decade axes were
      already stride-correct — **zero golden churn**). Added the "≤1 in-view minor tick →
      `AutoLocator`" fallback (mirrors `ticker.py` LogLocator.tick*values). Tests:
      `TestLogLocatorStrideAndInvalidDomains` gained `base10-/base2-exact-multiple` cases
      (verified against matplotlib 3.10.9) and `TestLogLocatorMinorFallsBackToAutoLocator`.
      \_Shipped 2026-07-06.*
- [x] `core/contour_lines.go` — saddle disambiguation now computes the cell-centre
      mean (bilinear centre value) and splits the ambiguous cell into **two** segments,
      isolating the diagonal corner-pair whose sign differs from the centre, instead of
      keying off `above[0]` and emitting all four crossings as one polyline. The centre
      uses a strict `mean > level` compare so an exact `mean == level` tie resolves as
      "below" — verified against matplotlib 3.10.9 `contour(...).allsegs` for the
      symmetric tie and both asymmetric pairings; the split direction matches the
      already-correct filled-contour path (`contourGridBandPolygons`). Tests:
      `TestStructuredContourLineClipsSingleSaddleQuadLikeMatplotlib` and
      `TestAxesContourUsesStructuredGridLinesLikeMatplotlib` corrected from the old
      single-polyline expectation to two segments, plus new
      `TestStructuredContourSaddleSplitUsesCellMean` (both pairings). Zero golden churn —
      the existing contour goldens (`contour_styles` et al.) don't cross an exact
      ambiguous saddle, so the exact-value unit tests (not a redundant pixel golden)
      provide the regression lock. _Shipped 2026-07-06._
- [x] 2D `bar(yerr=/xerr=)` — `BarOptions` (`core/plot.go`) gained `XErr`/`YErr`
      symmetric errors plus `ECol`/`CapSize`/`CapThick` and an `ErrorKw
*ErrorBarOptions` passthrough (asymmetric errors, errorevery). `Bar()` now
      builds a matplotlib-faithful error bar anchored at the bar top (vertical) /
      end (horizontal) — `ex=center, ey=baseline+height` — via the reused
      `Axes.ErrorBar` constructor (fmt="none", ecolor default black, capsize from
      the `errorbar.capsize` rc), added before autoscale so error extents widen the
      limits, and surfaced through the pre-existing `BarContainer.Errorbar` slot.
      Tests: `TestBarErrorBarsPlacementLikeMatplotlib` (anchor/color/container/
      autoscale exact-value lock) + `TestBarWithoutErrorDataHasNoErrorBar`. New
      parity case `bar_yerr` (golden byte-identical; matplotlib-ref RMSE 0.03 /
      PSNR 79.7 dB). Zero churn on existing bar goldens (error fields default off).
      _Shipped 2026-07-06._
- [x] `hist(log=)` — `HistOptions` (`core/plot.go`) gained a `Log bool` field;
      `Hist()` now calls `SetYScale("log", WithScaleNonPositive(NonPositiveClip))`
      when set, matching matplotlib's `hist(log=True)` →
      `set_yscale('log', nonpositive='clip')` (histograms here are vertical-only, so
      it always targets the y axis; the clip keeps the zero baseline finite instead
      of masking to NaN).
      Tests: `TestHistLogSetsYScaleToLog` + `TestHistWithoutLogKeepsLinearYScale`.
      New Showcase example `examples/hist_log` + parity case `hist_log` (golden
      byte-identical; matplotlib-ref RMSE 0.36 / PSNR 57.5 dB). Zero churn on the
      existing hist goldens (`Log` defaults off). _Shipped 2026-07-06._
- [ ] mathtext `cm`/`stix` fontsets — `core/mathtext.go:208` only remaps the font
      _family_ over the single DejaVu Unicode table; port matplotlib's
      `BakomaFonts`/`StixFonts` per-fontset glyph maps so non-DejaVu fontsets are
      parity-exact (currently only the DejaVu default is). Larger effort; scope first.
- [x] `core/norm.go` `TwoSlopeNorm` — out-of-range now maps to ±inf. `Map`/`Inverse`
      extrapolate `< VMin`/`> VMax` (and `< 0`/`> 1` for the inverse) to ∓inf, matching
      `np.interp(..., left=-inf, right=inf)` in `colors.py` TwoSlopeNorm. Coupled fix in
      `color/colormap.go` `AtValue`: `-inf`→under, `+inf`→over, only `NaN`→bad (was: all
      `Inf`→bad), so the default under/over color falls back to `At(0)`/`At(1)` exactly as
      matplotlib's `Colormap.__call__`. `twoslope_norm_image` golden stays byte-identical
      (below-vmin cells already routed to `At(0)`). Log/logit clip semantics aligned too:
      `Log.Fwd`/`Logit.transform` now pin the non-positive/out-of-range **log-space output**
      to the ∓1000 sentinel (matplotlib `scale.py`) before normalizing, instead of
      substituting a near-boundary input value; and matplotlib's LogScale default
      `nonpositive='clip'` is honored via a new unspecified-mode resolver (`resolveLogNonPositive`)
      plus `NewLog` defaulting to clip (logit default stays `mask`). `logClipFloor` is kept
      for domain repair only. Tests: out-of-range ±inf in `norm_test.go`, three-way Inf
      routing in `colormap_test.go`, sentinel/off-axis + default-clip in
      `scale_registry_test.go`. Only golden churn: `hist_log` regenerated — its
      matplotlib-ref parity **improved** from RMSE 0.36 to **0.16** (PSNR 57.5→68.0 dB).
      _Shipped 2026-07-06._
- [ ] Transform type-set breadth — add `ScaledTranslation`, `TransformWrapper`, and
      `Affine2D`-style `rotate/skew` builders; the separable→affine extraction is
      diagonal-only (`transform/transform.go:61`), so rotation/shear need manual matrix
      construction today. Triage against real demand before building.
- [ ] Add the ~3 missing accents vs matplotlib's 20-entry `_accent_map` (mathtext
      module).

**Third-audit additions (2026-07-01, REVIEW.md third review §2), all
file:line-verified on both sides:**

- [x] `core/vector_field_quiver.go` — quiver default (unset-scale) autoscale now
      ports matplotlib's `scale = 1.8·amean·sn/span` (`quiver.py:673-681`), replacing
      the home-grown `0.18·min(W,H)/√N` heuristic. **Note (stale audit):** the audit
      said `sn = clip(√N, 8, 25)`, but the vendored/installed matplotlib **3.10.9**
      (`quiver.py:674`) uses `sn = max(10, √N)` — that clip form is the separate
      _width_ default (`quiver.py:562`, already mirrored). We port the parity-correct
      `max(10, √N)` (verified against system matplotlib 3.10.9: `span==1` for the
      default `units="width"`, `scale==1.8·amean·sn`). In the Go pixel-space form
      (`mean = amean·Clip.W()`) this is `target = Clip.W()/(1.8·sn)`, `scale = mean/target`.
      `sn` uses the total anchor count (matplotlib `self.N`); `mean` averages only
      finite/positive arrows. Tests: `TestQuiverDefaultScaleMatchesMatplotlib` (exact
      value, sn-floor + √N sub-cases). New parity case `quiver_autoscale` (golden
      byte-identical; matplotlib-ref RMSE 1.61 / PSNR 62.6 dB). Zero churn on existing
      quiver goldens (every current 2D quiver sets an explicit scale). _Follow-up:_
      full per-`units`/`scale_units` generality (matplotlib's `span` for
      height/dots/inches, `a = lengths` for the `angles='xy'`+`scale_units='xy'` combo)
      is unmodeled — the default width-units path is exact and is what parity uses.
      _Shipped 2026-07-07._
- [x] `core/axes_autoscale.go` — autoscale margins now expand in scale-transform
      space and inverse-map back to data coordinates, matching
      `axes/_base.py:3064`: log padding is multiplicative and symlog padding
      follows the signed-log transform. The pre-margin domain now ports locator
      nonsingular behavior (generic `transforms.nonsingular(expander=.05)`;
      `LogLocator` positive filtering, minimum-positive replacement, adjacent
      decades for a lone point, and `[1, base]` for all-nonpositive data).
      The same upstream probe corrected the registry's symlog default
      `linthresh` from 1 to matplotlib 3.10.9's 2 (explicit overrides remain
      unchanged).
      Exact-value tests were probed against matplotlib 3.10.9 for linear origin
      and nonzero points, four log-domain cases, and symlog; all 173 catalog
      goldens/references remain green with zero fixture churn. _Shipped
      2026-07-23._
- [x] `core/axes_autoscale.go` — zero-valued bounds no longer conflate an
      origin-only data artist with “no data”: the autoscale collector has an
      explicit bounds-presence record plus per-axis minimum-positive values,
      implemented by `Line2D` and `Scatter2D` from their finite point data.
      `TestAutoScaleSingleOriginPointMatchesMatplotlib` locks matplotlib's
      `[-0.055, 0.055]` default window and the scatter variant covers the same
      zero-rectangle ambiguity. _Shipped 2026-07-23._
- [x] `core/axis_types.go` / `core/axis_spine.go` — spine positioning now
      supports Matplotlib's `('outward', points)` and `('axes', fraction)`
      modes through `SetSpinePositionOutward` / `SetSpinePositionAxes` and
      typed `AxisSpinePositionMode` constants. A shared perpendicular
      display-coordinate resolver keeps the spine, tick bases, tick labels,
      and axis-label anchor together; outward distances convert points through
      renderer DPI (negative values move inward), while axes fractions remain
      unclamped like Matplotlib. Exact four-side geometry tests match
      `spines.py:get_spine_transform`, including DPI scaling, and new catalog
      case `spine_positions` visually/reference-checks both modes at RMSE 0.61 /
      PSNR 74.1 dB. The intentional API addition is recorded in the frozen
      public audit. _Shipped 2026-07-23._
- [x] `core/date_tick.go` — DAILY interval table: **already correct** (the audit
      note was stale). `chooseDateTickInterval` (date_tick.go:711-726) already carries
      **both** matplotlib tables keyed on `date.interval_multiples`: `{1,2,3,7,14,21}`
      for `False` and `{1,2,4,7,14}` for `True` (the AutoDateLocator default), byte-matching
      `dates.py:1300/1312`. Covered by `date_rc_test.go:35/79`. No change needed.
      (The full rrule alignment simplification stays a documented divergence.)
- [x] `core/pie.go` — pie framing is now radius-independent: the default (`frame=false`)
      pie uses a fixed `±1.25 + center` data window regardless of radius (was
      `Radius*1.25`), and `frame=true` keeps the frame and lets autoscale fit the wedges,
      matching matplotlib's pie() `_request_autoscale_view()` / fixed-xlim branches.
      Test: `TestAxesPieFixedWindowIsRadiusIndependent` (radii 0.5/1/2, off-origin center).
      Zero golden churn (`specialty_artists` uses the default radius, so the window was
      already `≈±1.25`). _Shipped 2026-07-06._
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
