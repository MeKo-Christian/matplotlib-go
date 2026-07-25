# Matplotlib-Go Development Plan

This plan tracks the **remaining** work to bring `matplotlib-go` to a stable
v1.0 release. The foundation is built; what follows is the focused path to the
finish. The roadmap is cross-checked against the local upstream Matplotlib
snapshot in `third_party/matplotlib` so uncovered areas stay explicit instead of
sliding into a vague "future work" bucket.

Phases are ordered **closed first, open last**: Phases 1–17 are complete.
Phase 18 finishes Skia GPU mode, Phase 20 performs the pre-v1.0 API/package
rework, Phase 21 closes visual QA, and Phase 19 executes **last** to freeze and
tag the v1.0 release.

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

_Phases 1–17 are complete; see git history for detailed implementation logs._

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

# Phases 10–17: Completed Fidelity & Infrastructure Follow-up ✅

All are complete; detailed implementation logs remain available in git history.

- **Phase 10 — Delaunay:** extracted the Qhull-faithful engine to `qhull-go`.
- **Phase 11 — Render paths:** cached non-affine projections and added animated redraws.
- **Phase 12 — Failure honesty:** surfaced silent degradations and fixed text picking.
- **Phase 13 — Backends:** verified capability claims and explicit degradation paths.
- **Phase 14 — rcParams:** audited dead keys and honored targeted style defaults.
- **Phase 15 — Correctness:** closed artist, transform, layout, and algorithm gaps.
- **Phase 16 — Harness:** gated live renders, removed weak tolerances, and added fixtures.
- **Phase 17 — Defaults:** aligned plotting defaults with Matplotlib 3.10.9.

# Remaining Work (Open Phases)

The phases below still carry open to-do items and are ordered after the closed
phases. **Phase 18** finishes Skia GPU mode; **Phase 19** is the final v1.0
release stretch; Phases 20–21 complete the pre-release API and visual-QA work.

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

## Phase 18: Backend Deepening (Skia Native + GPU)

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

## Phase 19: Documentation, Performance, and v1.0 Release (LAST — after Phases 18, 20–21)

**Goal:** make the project consumable by users who have not followed the
development thread, establish performance baselines, and tag a stable v1.0.

> **Ordering (2026-07-01):** this phase executes **last**. Phase 20 is a
> deliberate pre-v1.0 breaking rework of the public API, so the "Public API
> stability audit ✅" checkbox below froze the _pre-break_ surface — Phase 20.4
> re-freezes it. The final golden/tolerance freeze must postdate Phases 17 and
> 21, and the changelog must include Phase 20's breaking section.

### 19.1 API Documentation ✅

Package-level GoDoc with worked examples for every stable public package;
guarded by `TestStablePublicPackagesHaveGoDocAndExamples`. Hosted docs at
pkg.go.dev; WASM gallery landing page links to README, examples, backend
selection, migration notes, and parity status.

### 19.2 Performance Pass ✅

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

### 19.3 Release Readiness

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

## Phase 20: Go-Idiomatic API Rework & `core/` Split (BREAKING, pre-v1.0)

**Goal:** one coordinated breaking pass — the only one before v1.0 — that makes
the API Go-idiomatic, splits the `core/` god-package (60,369 lines, 173 files,
1,529 exported symbols = 51% of the 3,019-symbol public surface), and
re-freezes the API afterwards. **Rendering behavior is untouched: goldens stay
byte-identical throughout — that invariant is the phase's regression gate**
(which is why Phase 17's golden churn must land first).

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

### 20.0 Ship-first, non-breaking (can land immediately, even before Phase 17)

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
      for Phase 19.

**Size:** XL (largest remaining phase; 20.2 and each 20.3 bullet are
independently executable sessions once 20.1's doc exists).
**Depends on:** Phase 17 (goldens must be stable so byte-identical is the gate).

**Exit criteria:**

- [ ] `core/` no longer contains plot3d/ticker/widgets; every plot method
      returns `(T, error)`; no raw-string enum fields in options; `Figure.Save`
      exists and examples use it; `stable_public_api.json` re-frozen and the CI
      audit green; goldens byte-identical to the Phase 17 baseline.

## Phase 21: Claude-Driven Visual QA Sweep (loose-tolerance closure)

**Goal:** visual inspection of every case whose gate is loose enough to hide a
real divergence — including the "RMSE passes but the output doesn't look right"
class — ending in a per-case disposition and a final tolerance ratchet that
Phase 19 freezes. Runs after 14/15/17 (which change images) and after 20 (whose
golden-stability gate must not be disturbed).

- [ ] **Re-arm the disabled gates:** `widgets_gallery` and `animation_gallery`
      (`MinPSNR 10 / MaxMeanAbs 95`) get real, binding thresholds after visual
      review (the mechanical removal of the theater thresholds is Phase 16's).
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
**Depends on:** Phases 14, 15, 17 (image-affecting) and 20 (freeze stability).
**Feeds:** Phase 19's "final golden/reference regeneration pass with per-case
tolerances frozen for v1.0".

**Exit criterion:**

- [ ] No catalog case has an effectively-disabled gate; every case with
      MaxRMSE ≥ 4 has a written disposition; the tolerance set handed to Phase 19
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
documented v1.0 release. **Phases 1–17 are closed**; their compact records and
detailed git history cover the parity foundation, extracted Delaunay engine,
render-path caching, fidelity reviews, harness hardening, and default alignment.
**Phase 18** finalizes Skia GPU acceleration, **Phase 20** performs the coordinated
pre-v1.0 API/package rework, and **Phase 21** completes visual QA. **Phase 19**
executes **last**: documentation and performance work is done, but the changelog,
CI gate, final freeze, and v1.0 tag must postdate Phases 18, 20, and 21.
