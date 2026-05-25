# Matplotlib-Go Development Plan

This plan tracks the remaining work to bring `matplotlib-go` to a v1.0 release. The
roadmap is cross-checked against the local upstream Matplotlib snapshot in
`third_party/matplotlib` so uncovered areas are tracked explicitly instead of
being deferred to a vague "future work" bucket.

---

# Plan Tracking

- `✅` = done and stable
- `🧪` = implemented but under hardening
- `⚪` = in progress
- `⚠️` = deferred / design decision required
- `[ ]` = not started

---

# Where We Are Today

The project has progressed well past the proof-of-concept stage. The major
milestones already achieved are:

**Architecture and core model**

- Artist hierarchy (`Figure → Axes → Artists`) with stale/callback propagation,
  draw scheduling, and lifecycle parity with upstream Matplotlib's
  `RendererBase` / `GraphicsContext` split.
- Explicit transform graph with `transData`, `transAxes`, `transFigure`, blended
  and offset transforms, and a projection-friendly composition layer.
- Backend-agnostic `render.Renderer` contract with optional capability
  interfaces (`TextDrawer`, `ImageTransformer`, `MarkerDrawer`,
  `PathCollectionDrawer`, `NativeHatcher`, `ClipPathTransformer`,
  `PNGExporter`, `SVGExporter`, …).
- Capability matrix and save dispatch driven by the backend registry rather
  than backend-name conditionals.
- mplot3d parity fixtures now enforce `MaxRMSE < 18` against the Matplotlib
  references, including z-axis autoscale behavior, pane/grid framing, exact
  tab10 colors, and refreshed 3D golden images.

**Backends**

- **AGG** (`backends/agg`) — primary anti-aliased raster backend, native marker
  batches, path collections, Gouraud triangles, transformed images, hatches,
  buffer regions (`CopyFromBBox` / `RestoreRegion`), and offscreen filters
  (`StartFilter` / `StopFilter`).
- **GoBasic** (`backends/gobasic`) — pure-Go correctness fallback with full
  renderer-neutral contract coverage and documented fidelity limits.
- **SVG** (`backends/svg`) — deterministic vector output with native clip
  paths, marker / path-collection batches, `<pattern>` hatches,
  text-as-text vs text-as-path policy, hash-salted IDs, and a structural diff
  test harness.
- **Skia** (`backends/skia`) — opt-in CPU raster backend behind a build tag,
  shares the raster contract surface for paths, images, clipping, text, and
  PNG export.

**Plot vocabulary**

- 2D basics: line, scatter, bar (vertical/horizontal/grouped/stacked),
  fill / fill_between / fill_betweenx, step, stairs, axhline/vline/span,
  broken_barh.
- Statistical: histogram (with binning strategies and density / cumulative
  variants), boxplot (notches, confidence intervals, custom whiskers),
  errorbar (asymmetric, with limit indicators), violinplot, ecdf, stackplot.
- Color-mapped: imshow / matshow / spy, pcolor / pcolormesh (flat / nearest /
  gouraud), contour / contourf, hist2d (weighted, density), hexbin, tripcolor,
  tricontour / tricontourf.
- Vector fields: quiver, quiverkey, barbs, streamplot.
- Specialty: stem, eventplot, pie, table, sankey, specgram, psd, csd, cohere,
  xcorr, acorr, annotated heatmaps, magnitude / angle / phase spectra.
- Patches and collections: Rectangle, Circle, Ellipse, Polygon, PathPatch,
  FancyArrow, hatch fills, PathCollection, LineCollection, PatchCollection,
  PolyCollection, QuadMesh.
- 3D (`mplot3d`): plot3d, scatter3d, surface, wireframe, contour / contourf,
  trisurf, voxels, quiver3d, errorbar3d, stem3d, bar3d, fill_between3d,
  with depth sorting and shared scalar-mappable state.
- Non-Cartesian: polar, radar, skewx, mollweide projections via the
  `projection=` registry.

**Layout, composition, and styling**

- Subplots, `add_subplot`, `GridSpec`, `subplot_mosaic`, nested grids,
  `SubFigure`, granular share modes, twin / secondary axes.
- Layout engines: `subplots_adjust`, `tight_layout`, `constrained_layout`,
  measured-text margin computation, colorbar slot management.
- Inset / zoomed-inset, `AxesDivider`, `ImageGrid`, `RGBAxes`, parasite axes,
  anchored artists, `AxisArtist`, floating axes, curvilinear grids.
- Figure-level labels (`suptitle`, `supxlabel`, `supylabel`), figure legends,
  anchored boxes.
- Style system: `rcParams`, `rc`, `rc_context`, `rcdefaults`, `.mplstyle`
  loading, theme library, publication-ready themes.

**API surface**

- Object-oriented core API plus a stateful `pyplot` layer covering the common
  Matplotlib migration path (`Figure`, `GCF`, `GCA`, `Subplot`, `Subplots`,
  `title`, `xlabel`, `legend`, `colorbar`, `savefig`, `show`, …).
- Convenience entry points: `SemilogX`, `SemilogY`, `LogLog`, `PlotDate`,
  `Fill`, `BarH`, full spectrum variant wrappers.
- Color-mapping model: `Normalize`, `NoNorm`, `LogNorm`, `SymLogNorm`,
  `PowerNorm`, `TwoSlopeNorm`, `CenteredNorm`, `BoundaryNorm`, with consistent
  scalar-mappable routing through every color-mapped artist.
- Date / category / unit converters and locators.

**Tooling and infrastructure**

- Headless `FigureCanvas` / `FigureManager` abstraction with event model
  (mouse / key / resize / draw / close) and tool manager scaffolding.
- WASM web demo host with persisted light/dark theme switch, focus/input
  preservation, and GitHub Pages deployment.
- `cmd/example` runner with `-list` mode driven by the
  `internal/examplecatalog` source of truth.
- Catalog-driven parity test suite: `TestGolden`, `TestMatplotlibRef`,
  `TestReferenceCompare` discover cases by ID; per-case tolerances live on the
  catalog row.

**What's left** is the focused work in the phases below: remaining PS/PGF
hardening, renderer effects (patterns / gradients / path effects), MathText /
TeX promotion follow-through, animation, backend deepening, parity closure,
example / browser-gallery breadth, and documentation polish for v1.0.

---

# Phase 1: Publication Backends and Shared Save Pipeline

✅ **Completed.** PDF, PS/EPS, and PGF are now deterministic publication-vector
backends integrated into one shared, extension-driven save pipeline across PNG,
SVG, PDF, PS/EPS, and PGF.

Completed scope:

- PDF (`backends/pdf`) supports deterministic object/page-stream writing,
  `SOURCE_DATE_EPOCH` metadata, full path/fill/stroke/clip drawing, embedded
  Type 0/CIDFontType2 subsets with deterministic maps, and reusable native
  resources for images, hatches, marker/path collections, alpha, and
  transformed images.
- PS/EPS provides deterministic Level-2 output for paths, clips,
  strokes/fills, transformed images, native hatches, and reusable
  marker/path-collection procedures; unavoidable Level-2 limitations versus PDF
  are implemented where possible and clearly documented otherwise.
- PGF provides deterministic generator-only `pgfpicture` output for paths,
  clips, text/rotated text, opacity, hatches, raster/transformed images,
  mixed-raster groups, and reusable marker/path-collection macros, with TeX
  compilation optional and generator-smoke CI coverage.
- Shared save routing uses `SelectBackendForExtension` + `SaveFormats` across
  public entry points (`pyplot`, canvas/manager, CLI, examples), with unified
  `render.SaveOption` plumbing (including PGF metadata/preamble/comment/
  verification controls), capability reporting, explicit unsupported-option
  errors, and consistent mixed raster/vector behavior via DPI-aware offscreen
  replay.
- Structural/smoke/golden fixtures cover backend selection, export stability,
  text/clipping/transforms, alpha, rasterized artists, and vector-surround
  preservation.

---

# Phase 2: Renderer Effects, Patterns, and Compositing

✅ **Completed.** Phase 2 is fully closed. Artists can request pattern fills,
gradient fills, path effects, filtered path-effect passes, and mixed
raster/vector output through renderer-neutral capability interfaces, without
backend-name conditionals in core effect routing.

**Reference sources:** `third_party/matplotlib/lib/matplotlib/patheffects.py`,
`third_party/matplotlib/lib/matplotlib/colors.py`, AGG filter behavior, SVG
`<pattern>` / `<filter>` output, and PDF image/form resource behavior.

Completed scope:

- Pattern and gradient fills: `render.PatternFill` and `render.GradientFill`
  live on `render.Paint`, with `render.PatternFiller` and
  `render.GradientFiller` capability interfaces plus graphics-context alpha
  propagation.
- Native targets: AGG, SVG, PDF, and Skia-tagged CPU builds all advertise and
  satisfy pattern/gradient/path-effect capabilities through runtime interfaces.
- Backend output: AGG renders gradient spans and tiled pattern fills; SVG emits
  deterministic gradient/pattern/filter defs; PDF emits axial/radial shadings,
  tiling patterns, transparency-group forms, and soft-mask image XObjects;
  Skia uses the CPU bridge for the same renderer-neutral contracts.
- Path effects: `render.PathEffect` covers normal, stroke/halo, simple patch
  and line shadows, patch effects, ticked strokes, and filtered repaint passes;
  core line, patch, text, scatter, and collection artists carry effects through
  to path paints.
- Filter routing: SVG and PDF use `render.PathEffectFilterDrawer` where they
  can produce backend-local output; AGG and Skia use `render.FilterRenderer`;
  PS, PGF, and GoBasic document and report fallback semantics truthfully.
- Mixed raster/vector output: `render.Rasterization` and
  `render.RasterizationController` support explicit `Artist.SetRasterized(true)`
  and auto-rasterization for dense output, with SVG/PDF/PS/PGF embedding
  DPI-aware raster tiles while preserving surrounding vector content.
- Regression coverage: cross-backend semantic tests assert effect routing uses
  `render.PatternFiller`, `render.GradientFiller`, `render.PathEffectDrawer`,
  `render.PathEffectFilterDrawer`, and `render.FilterRenderer`; backend
  capability matrix tests pin truthful native/fallback declarations.
- Fixture coverage: committed catalog fixtures cover path effects,
  pattern/gradient/effect combinations, and mixed raster/vector output:
  `path_effects`, `pattern_gradient_effects`, and `mixed_raster_vector`, each
  with golden and Matplotlib-reference PNG coverage where applicable.

Exit criteria:

- [x] Pattern fills, gradients, and path effects work uniformly across AGG,
      SVG, PDF, and Skia without backend-name conditionals.
- [x] `Artist.SetRasterized(true)` produces correct mixed-mode output on every
      vector backend.
- [x] All effects have committed golden and Matplotlib-reference fixtures.

---

# Phase 3: Mathematical Text and TeX

✅ **Completed.** MathText and `usetex` are first-class across the active
raster/vector targets, with toolchain-gated TeX coverage and a documented
promotion date for `internal/mathtext`.

**Goal:** make MathText and `usetex` first-class across raster and vector
backends, and stabilize `internal/mathtext` for promotion.

**Reference sources:** `third_party/matplotlib/lib/matplotlib/_mathtext.py`,
`mathtext.py`, `texmanager.py`; current `internal/mathtext/`.

### 3.1 MathText Pipeline Completion

- [x] Finish the shared shaping layer (carried over from prior backend work)
      so AGG text draw, text measurement, text bounds, and text-path output all
      consume the same shaped glyph runs.
- [x] Complete the MathText grammar coverage gaps versus upstream: stacked
      fractions, accents, big operators, integral limits, matrix environments.
      Unit coverage now exercises stacked operator limits, integral side
      scripts, accents, matrix/array environments, fractions/genfrac, roots,
      fences, and spacing; the Phase 3 parity fixtures cover the same areas
      against Matplotlib references.
- [x] Cache stabilization: deterministic cache keys, eviction policy, and
      cross-process safe storage so `internal/mathtext` can ship as its own
      module.
      In-memory caches now support deterministic FIFO bounds via
      `CacheConfig`; `Cache.SaveFile` / `Cache.LoadFile` provide deterministic
      JSON snapshots with atomic writes for cross-process reuse.
- [x] MathText draw path through every backend: AGG (raster glyph composite),
      SVG (paths or text-as-text where the font is available), PDF, Skia.
      AGG has Matplotlib reference PNG fixtures; SVG and PS have explicit
      MathText vector-output smoke tests plus rasterized tolerance checks
      against the AGG goldens; PDF has golden and rasterized tolerance coverage
      for all Phase 3 MathText fixtures; Skia has build-tagged golden
      comparison coverage.
- [x] Golden fixtures: mathtext_basic, mathtext_fractions, mathtext_integrals,
      mathtext_matrices, mathtext_inline_labels.

### 3.2 `usetex` Support

- Current implementation status: AGG, SVG, and PDF have opt-in
  `latex`+`dvipng` raster paths backed by `internal/tex`; PDF embeds cached TeX
  PNGs as image XObjects and scales them from raster pixels to PDF points.
- DVI geometry status: `internal/tex` now parses DVI page extents with
  Matplotlib-style width/height/descent semantics, including rules and glyphs
  backed by TFM metrics found through configured directories or `kpsewhich`.
  `Manager.Render` prefers those cached DVI metrics and falls back to PNG
  dimensions when DVI/TFM geometry is unavailable.
- System TeX integration status: `internal/tex` includes a toolchain-gated
  integration test for the real `latex` + `dvipng` path. It skips with a clear
  diagnostic when those commands are absent. The `test` package also has a
  toolchain-gated `text.usetex` artist-pipeline smoke test through AGG and a
  gated AGG PNG golden harness (`-update-usetex-golden`) for hosts with a TeX
  toolchain.
- [x] `usetex` import path that shells out to a system `latex` / `dvipng` /
      `dvisvgm` pipeline, behind a build tag / rc switch so the default build
      has no external dependency.
- [x] DVI parser sufficient to read the geometry of the rasterized result
      back into the renderer's text bounds API.
- [x] Shared clipping, alpha, and DPI semantics between MathText and `usetex`
      paths so the artist-side API does not branch.
- [x] Golden fixtures gated by the presence of a TeX installation; skip with
      a clear diagnostic when missing.
      The committed `testdata/usetex_golden/basic.png` fixture is regenerated
      by `go test ./test -run TestUseTeXGoldenWithSystemToolchain
      -update-usetex-golden` on hosts with `latex` + `dvipng`.

### 3.3 MathText Module Promotion

- Promotion target: document the final `internal/mathtext` API by
  2026-07-31, then either move it to a top-level `mathtext` package in this
  repository or split it into its own module before the v1.0 API freeze.
- [x] Stabilize the public API surface of `internal/mathtext` against the
      needs of the AGG, SVG, PDF, and Skia text drawers.
      `internal/mathtext/doc.go` documents the retained internal API:
      normalization, display segmentation, layout-to-runs/rules, renderer font
      resolution hooks, and cache/storage contracts.
- [ ] Promote `internal/mathtext` to a top-level module / repo with its own
      versioning after the documented promotion date; this is v1.0 API-freeze
      follow-through rather than a Phase 3 exit blocker.

**Exit criteria:**

- [x] MathText renders identically across all backends within documented
      tolerances.
      AGG PNG goldens are cross-checked against Matplotlib references with
      catalog tolerances; PDF, PS, and SVG vector outputs are rasterized and
      compared against the same AGG goldens; Skia has build-tagged PNG golden
      comparisons for the Phase 3 fixture set.
- [x] `usetex` is opt-in, dependency-free by default, and tested when the
      external TeX toolchain is present.
- [x] `internal/mathtext` is either standalone or has a documented promotion
      date.

---

# Phase 4: Interactive Backends and Event Loop

✅ **Completed.** The headless event scaffolding has been turned into shared
interactive infrastructure with desktop, web, and WASM frontends.

Completed scope:

- Navigation, picking, hover coordinate formatting, callback registration, and
  Matplotlib-style event lifecycle semantics are implemented on the shared
  canvas/event surface.
- The Gio desktop backend hosts AGG rendering, maps toolkit input to canvas
  events, supports toolbar actions, and ships desktop embedding examples.
- The WebAgg backend implements server-side managers, WebSocket protocol
  handling, browser canvas updates, toolbar state, binary PNG frames, and
  embedding examples.
- The WASM demo host supports pan, zoom, rubber-band selection, toolbar
  actions, hover state, and shared input normalization.
- Draw-idle coalescing, stale redraw propagation, blit/damage-region support,
  lifecycle/error tests, and matrix smoke coverage exercise catalog
  representatives across Gio and WebAgg.
- `docs/interactive-backends.md` documents the common event, picker, toolbar,
  draw-idle, save, and backend capability contracts.

---

# Phase 5: Widgets and Selectors

✅ **Completed.** Static widget artists are fully interactive through the
shared Phase 4 event dispatcher.

Completed scope:

- Buttons, sliders, range sliders, check buttons, radio buttons, and text boxes
  support pointer and keyboard interaction, state changes, and callbacks.
- Span, rectangle, ellipse, polygon, and lasso selectors support mouse and
  keyboard editing with upstream-style modifier behavior.
- Cursor and multi-cursor helpers are driven by shared hover events.
- Widget z-order, widget-axis layout helpers, and a complete widget gallery are
  in place for headless, desktop, and web backends.

---

# Phase 6: Animation

**Goal:** add the animation API that depends on the interactive event loop
and the blit-friendly redraw paths from Phase 4.

**Reference sources:** `third_party/matplotlib/lib/matplotlib/animation.py`.

### 6.1 Animation API

- [x] `FuncAnimation` and `ArtistAnimation` mirroring upstream signatures,
      driven by the figure's draw-idle scheduler.
      The new `animation` package wraps `canvas.FigureCanvas` with a
      deterministic frame loop. `NewFuncAnimation` calls a user-supplied
      `UpdateFunc(frame int) ([]core.Artist, error)` and an optional
      `InitFunc`; `NewArtistAnimation` toggles visibility on a fixed
      `[][]core.Artist` per frame. Both reuse the shared `Animation` core
      and timer integration in `canvas.EventLoop`.
- [x] Frame timing / pacing controls (interval, repeat, repeat_delay,
      blit toggle) with deterministic-frame mode for tests.
      `Animation.Step()` advances one frame without spinning a real timer;
      `Animation.Start()` installs a timer on the configured `EventLoop`.
      Finite, non-repeating runs auto-stop after `Frames` frames; repeating
      runs wrap the frame counter and insert a one-shot `RepeatDelay` skip
      tick before the next cycle.
- [x] Artist `set_animated(true)` flag honored by the AGG and Skia
      backends via blit regions.
      `core.ArtistRasterization` now carries an `animated` flag (with
      `SetAnimated` / `Animated` accessors) which every embedding artist
      inherits. `core.DrawFigure` skips animated artists by default and
      `core.DrawFigureWithOptions(fig, r, DrawOptions{AnimatedFilter: ...})`
      gives the animation engine background-suppress, animated-only, and
      all-artist passes. When the canvas implements both `BlitCanvas` and
      `BufferRegioner`, the engine captures a background snapshot on the
      first frame, then restores + blits per frame; otherwise it falls back
      to a full redraw, which still drives AGG and Skia correctly.

### 6.2 Frame Writers

- [ ] GIF writer (pure-Go encoder, no external dependency).
- [ ] APNG writer for higher-quality web demos.
- [ ] MP4 / WebM writers via optional `ffmpeg` shellout, gated by a build
      tag and runtime detection.
- [ ] HTML embedding writer producing self-contained
      `<video>` / `<canvas>` snippets for the web demo host.

### 6.3 Animation Examples and Fixtures

- [ ] Animated line plot, animated scatter / collection, animated imshow
      (heatmap), animated subplot composition.
- [ ] Deterministic-frame golden fixtures verifying frame-N output matches
      Matplotlib's frame-N output within tolerance.

**Exit criteria:**

- [ ] `FuncAnimation` produces correct frames in headless mode and animates
      smoothly in interactive backends.
- [ ] At least one self-contained file format works without external
      dependencies (GIF).
- [ ] Animation examples appear in the WASM demo gallery.

---

# Phase 7: Backend Deepening and Parity Hardening

**Goal:** finish the backend-specific work that was carved out of the earlier
backend parity program but is not yet complete.

### 7.1 AGG Native Capabilities

- [x] Complete the AGG MathText raster pipeline once Phase 3.1 lands so
      raster glyph composition shares the same shaping pipeline as text-as-path.
- [x] Plumb `usetex` output through AGG using the DVI parser from Phase 3.2.
- [ ] Expand AGG parity diagnostics for remaining non-text residuals: dense
      path collections, repeated translucent overlaps, image interpolation modes,
      hatch clipping, and mixed raster / vector fallbacks.
- [x] Split AGG-native parity fixtures from renderer-neutral fallback
      fixtures so missing native AGG behavior cannot be hidden by fallback
      drawing.

### 7.2 SVG Coverage Expansion

- [x] Expand the structural golden set to the remaining canonical plot
      families: bar, errorbar, hist, collection, image, clipped polar,
      hatch_bars, text_layout, mathtext.
- [x] Wire the SVG-specific golden set into the catalog so the structural
      diff harness runs alongside the rasterized golden / reference comparison.

### 7.3 Skia Native Paths

- [ ] Native Skia marker batches, path collections, transformed images,
      quad meshes, and Gouraud triangles wired through `SkCanvas::drawAtlas` and
      `SkVertices`.
  - [x] CPU Skia renderer implements the renderer optional interfaces for
        marker batches, path collections, transformed images, quad meshes, and
        Gouraud triangles through the deterministic Skia bridge boundary; the
        external `drawAtlas` / `SkVertices` ABI integration remains open.
  - [x] CPU Skia reports `MarkerBatch`, `PathCollectionBatch`, and
        `QuadMeshBatch` as bridged (`≈`) via the new
        `render.CapabilityBridgeReporter` interface, so the comparison report
        distinguishes the CPU bridge stand-ins from the truly native
        `GouraudTriangleBatch` path. The external batch ABI remains the open
        path to flipping these back to native (`✓`).
- [ ] Skia native hatching via tiled `SkShader`s.
  - [x] CPU Skia consumes hatch metadata during path rendering and advertises
        `NativeHatcher`; tiled external `SkShader` hatches remain open.
  - [x] CPU Skia reports `NativeHatcher` as bridged (`≈`) until the external
        tiled `SkShader` integration lands. Hatch geometry continues to route
        through `render.DrawHatchFallback`.
- [ ] GPU mode (`SkSurface::MakeRenderTarget`) behind a separate build tag,
      with deterministic CPU readback for golden tests.
- [ ] Capability reporting split between CPU and GPU configurations so the
      comparison report shows truthful native / fallback / unavailable status
      per mode.
  - [x] CPU Skia capability reporting now marks implemented optional paths as
        native instead of fallback; GPU-specific capability reporting remains
        deferred until the GPU build tag exists.
  - [x] A fourth `CapabilityBridged` status (`≈` marker) sits between native
        and fallback in `BackendComparisonReport` so the CPU bridge stand-ins
        are visible. The CPU/GPU split itself is still deferred until a GPU
        build tag exists; the bridged status is the only CPU-side distinction
        the comparison report needs today.
- [x] Skia vs AGG semantic-fixture comparison; tolerances documented per
      fixture where Skia is not expected to pixel-match.
      `TestSkiaParityAgainstAGGGoldens` in `backends/skia/parity_test.go`
      (build-tagged `skia`) iterates every catalog `Case` opted into the new
      `SkiaParityFamily` field and compares Skia CPU output against the
      committed AGG golden. Per-case `MinPSNR` / `MaxMeanAbs` overrides on
      the catalog row take precedence over the harness defaults; failures emit
      got / golden / diff artifacts under
      `testdata/_artifacts/skia_parity/{id}/`. Covers line, scatter, bar,
      fill, errorbar, histogram, text, MathText, image, mesh, hatch/patch,
      polar, and path-effect / pattern-gradient families.

### 7.4 GoBasic Long Tail

- [x] GoBasic equivalents for the renderer-neutral path effect pipeline
      introduced in Phase 2 so the fallback backend keeps full semantic coverage.
  - [x] Add GoBasic semantic tests for `Normal`, `Stroke`, `PathPatch`,
        `SimplePatchShadow`, `SimpleLineShadow`, and `TickedStroke` path
        effects, proving each pass draws visible non-background pixels and
        honors offset / stroke / fill semantics through `render.DrawPathWithEffects`.
  - [x] Add GoBasic filter-effect fallback coverage documenting that unsupported
        blur/filter effects degrade to a deterministic repaint pass unless the
        renderer implements `render.FilterRenderer` or `render.PathEffectFilterDrawer`.
  - [x] Add artist-level GoBasic path-effect smoke tests for line, patch,
        scatter / path collection, and text-path effects so core artist routing
        cannot bypass the renderer-neutral path-effect pipeline.
  - [x] Update GoBasic docs and capability expectations to state which path
        effects are semantic fallbacks and which filter effects are intentionally
        approximate.
- [x] GoBasic smoke coverage for any new plot family introduced in Phases
      1-6 (PDF / interactive / animation paths excluded since GoBasic targets
      static output).
  - [x] Add a catalog-derived GoBasic smoke test that renders every static
        catalog case opted into `GoBasicSmokeFamily` through `backends.GoBasic`
        and asserts the image is non-empty.
  - [x] Keep cases requiring unsupported external/runtime dependencies
        (`usetex`, interactive, animation, vector export, or backend-native
        AGG / Skia fixtures) out of the GoBasic smoke set by requiring explicit
        `GoBasicSmokeFamily` opt-in metadata.
  - [x] Ensure every Phase 1-6 plot family has at least one GoBasic-rendered
        smoke case in the catalog-derived set, including path effects,
        MathText, images, meshes, collections, hatches, statistical views,
        specialty artists, projections, layouts, and other static catalog
        surfaces.
  - [x] Add a small summary test for coverage metadata so newly added static
        plot families fail fast until they declare GoBasic smoke coverage or
        an explicit skip reason.

**Exit criteria:**

- [ ] AGG, SVG, and Skia all advertise truthful capability matrices for
      every optional renderer interface, with no `✓!` drift markers in the
      comparison report.
- [ ] Every committed plot family has at least one native and one
      fallback-path fixture so silent fallbacks cannot pass for native
      behavior.
- [ ] Skia is a viable secondary raster backend for users who need GPU
      acceleration.

---

# Phase 8: RMSE > 5 Example Parity Audit

**Goal:** close the current catalog cases whose committed Go golden differs
from the Matplotlib reference by `RMSE > 5`, using core-library parity fixes
before fixture tweaks. Baseline captured with
`go test -v ./test -run TestReferenceCompare` on 2026-05-21.

**Visual audit artifacts:** `testdata/_artifacts/reference_compare/*_golden.png`,
`*_matplotlib_ref.png`, and `*_golden_vs_matplotlib_ref_diff.png`.

**Latest RMSE run:** `rtk proxy go test -v ./test -run TestReferenceCompare -count=1`
on 2026-05-23 passes all committed catalog tolerances after refreshing the
stale `mplot3d_scatter3d` Python reference and targeted Go goldens.
`mplot3d_scatter3d` is now `RMSE 8.42` (down from 18.81),
`patch_showcase` is `RMSE 6.39`, `lognorm_imshow` is `RMSE 9.42`
after switching log tick labels to Matplotlib-style powers, `spy_marker`
is `RMSE 2.50` after matching Line2D marker sizing, and a focused
`TestReferenceCompare/arrays_showcase` run now reports `RMSE 8.73` after
using structured-grid contour lines and Matplotlib-like contour z-order.
The same structured contour fix moves `mesh_contour_tri` to `RMSE 7.52`.
`pcolormesh_gouraud` is now `RMSE 4.78` after matching Matplotlib's
four-triangle Gouraud quad conversion.
`stem_plot` is now `RMSE 4.13` and `mplot3d_stem3d` is now `RMSE 13.55`
after matching Matplotlib stem baseline color, marker size, and 3D stem cap
defaults.
`spy_image` is now `RMSE 3.45` after matching Matplotlib's ceiled image
resampling buffer geometry and refreshing the stale Go golden.
`figure_labels_composition` is now `RMSE 6.60` after constrained layout
reserves the full figure-label line box instead of only the ink bounds.
`spectrum_variants` is now `RMSE 9.87` after aligning the fixture-only
near-zero FFT residue inputs with the NumPy/Matplotlib reference arrays and
refreshing the Go golden.
`specialty_depth` is now `RMSE 9.42` after matching Matplotlib's filled
errorbar limit markers and drawing hexbin marginals above the main hex
collection.
`twoslope_norm_image` is now `RMSE 6.24` after matching Matplotlib's
function-scaled colorbar axis for `TwoSlopeNorm`.
`colorbar_extensions` is now `RMSE 7.00` after matching Matplotlib's extended
colorbar box-aspect shrink compensation.
`specialty_artists` is now `RMSE 7.52` after matching Matplotlib's 1 pt
table patch linewidth default, anchoring table text from ink bounds, and
auto-sizing row-label cells from renderer text bounds, including Matplotlib's
bbox scaling/offset behavior for auto row-label columns, plus Matplotlib's
auto patch snapping for rectilinear table cell paths, butt caps on two-sided
violin summary lines, collection-level violin body alpha, and unclipped table
overlay drawing.
`mixed_raster_vector` is now `RMSE 9.30` after the Matplotlib-compatible polar
theta label padding/centering fix moved its polar panel under the temporary
target.
`polar_axes` is now `RMSE 6.22` after matching Matplotlib's theta tick label
centering and `_pad + 7pt` padding in addition to the radial spine/label
changes.
`radar_basic` is now `RMSE 6.99` after removing the radar-specific theta tick
size workaround and using the same Matplotlib theta label padding/centering.
`units_dates` is now `RMSE 5.92` after preserving explicit date locators and
formatters across unit-axis refreshes.
`skewt_basic` is now `RMSE 5.02` after matching Matplotlib's skew-x top-axis
visibility, upper-interval x gridlines, and unskewed y tick placement.
`mplot3d_basic` is now `RMSE 3.76` after matching Matplotlib's 3D tick/axis
label offset deltas to the expanded frame limits returned by `_get_coord_info`.
The same 3D label-offset fix moves every `mplot3d_*` reference fixture below
`RMSE 4` after refreshing the affected Go goldens.
No non-mathtext cases remain above the temporary Phase 8 target of
`RMSE < 10`. The mathtext cases remain excluded from this temporary threshold.

**Source parity audit:** completed on 2026-05-22 with sub-agents across all
Phase 8 subphases. Direct example/fixture mismatches were fixed where existing
Go APIs could express the same Matplotlib call semantics. Remaining unchecked
items are renderer/layout/core API parity work, not example-source workarounds.
Do not make a fixture pass by adding per-example offsets, hidden special cases,
or catalog-only hacks. If the Go example already expresses the same public
Matplotlib semantics, the fix belongs in the core plotting model, layout code,
renderer contract, backend implementation, or the AGG port itself.

### 8.1 `fill_basic` (RMSE 6.38)

- [x] Code: source audited; no material example mismatch found.
- [ ] Visual: polygon shape matches; residuals sit on fill outline, baseline,
      ticks, and text antialiasing.
- [ ] Likely core areas: `core/fill.go`, AGG polygon stroke/fill antialiasing,
      text metrics.

### 8.2 `fill_stacked` (RMSE 9.32)

- [x] Code: source audited; no material example mismatch found.
- [ ] Visual: stacked layer boundaries and outer fill edges differ; title/text
      residuals remain.
- [ ] Likely core areas: `FillBetween` / `FillToBaseline` stroke ordering,
      alpha compositing, path rasterization.

### 8.3 `errorbar_basic` (RMSE 6.95)

- [x] Code: source audited; data, colors, and line widths are intentionally
      matched.
- [ ] Visual: plot is close; caps, marker edges, line endpoints, and labels
      drive the diff.
- [ ] Likely core areas: `core/errorbar.go`, cap-size semantics, scatter marker
      rasterization, stroke antialiasing.

### 8.4 `boxplot_basic` (RMSE 10.72)

- [x] Code: removed the explicit `CapWidth` override and matched the Python
      explicit light-gray `lw=0.5` y-grid.
- [ ] Visual: strong horizontal grid residuals plus box, cap, and flier edge
      differences.
- [ ] Likely core areas: boxplot cap defaults, `core/grid.go` grid defaults,
      marker/stroke rendering.

### 8.5 `text_labels_strict` (RMSE 5.91)

- [x] Code: source audited; no source mismatch found.
- [ ] Visual: axes structure is identical; residuals are mostly title, axis,
      tick-label placement, and glyph antialiasing.
- [ ] Likely core areas: `core/text.go`, AGG text measurement/baseline, axis
      label offsets.

### 8.6 `mathtext_basic` (RMSE 18.04)

- [x] Code: data and math strings match; replaced anchored-text shortcut with
      axes-fraction `Text` + bbox and matched annotation arrow styling.
- [ ] Visual: math glyph sizes, baselines, superscripts/subscripts, anchored
      box, and annotation arrow/text differ.
- [ ] Likely core areas: `internal/mathtext`, `core/mathtext.go`, AGG text
      bounds, annotation and anchored-box layout.

### 8.7 `mathtext_fractions` (RMSE 26.98)

- [x] Code: source audited; no source mismatch found.
- [ ] Visual: fraction stacks, binomial layout, roots, and bracket sizing are
      visibly different.
- [ ] Likely core areas: `internal/mathtext/layout.go`, fraction axis
      alignment, stretchy delimiters, square-root geometry, glyph metrics.

### 8.8 `mathtext_integrals` (RMSE 30.42)

- [x] Code: source audited; no source mismatch found.
- [ ] Visual: integral/sum/product operator sizing and limit placement differ
      strongly.
- [ ] Likely core areas: large-operator display fonts, over/under limit layout,
      math glyph metrics.

### 8.9 `mathtext_matrices` (RMSE 25.54)

- [x] Code: source audited; no source mismatch found.
- [ ] Visual: matrix delimiters, row spacing, `\quad` spacing, and angle
      brackets differ.
- [ ] Likely core areas: `\genfrac` layout, delimiter sizing, matrix/stack ink
      bounds.

### 8.10 `mathtext_inline_labels` (RMSE 18.03)

- [x] Code: math sources match; remaining `LegendBest` placement difference is
      core best-location behavior.
- [ ] Visual: legend lands differently; math text in title, labels, and legend
      still has glyph/baseline residuals.
- [ ] Likely core areas: `core/legend.go` best-placement badness,
      `internal/mathtext`, text metrics.

### 8.11 `image_heatmap` (RMSE 5.54)

- [x] Code: translated Python `imshow(..., interpolation="nearest",
    aspect="auto", extent=...)` through `ax.ImShow`.
- [ ] Visual: cells align; residuals appear at cell boundaries and tick/text
      edges.
- [ ] Likely core areas: `core/image.go`, `core/image_api.go`, image pixel
      snapping and resampling defaults.

### 8.12 `imshow_clipped` (RMSE 8.19)

- [x] Code: source audited; data helper matches and no material source mismatch
      found.
- [ ] Visual: clipped image contents match, but row/column boundary residuals
      are strong.
- [ ] Likely core areas: `core/matrix_helpers.go`, `core/image.go`, nearest
      source-coordinate mapping, clip/pixel edge alignment.

### 8.13 `imshow_transformed` (RMSE 7.04)

- [x] Code: source audited; both examples rotate 28 degrees around the image
      center.
- [ ] Visual: transformed image is aligned, but interpolation gradients and
      rotated edges differ across most of the raster.
- [ ] Likely core areas: transformed image affine, bilinear sampling, AGG
      `ImageTransformed` interpolation.

### 8.14 `spy_marker` (RMSE 2.50)

- [x] Code: source audited; marker mode now matches Matplotlib Line2D marker
      sizing by converting the marker edge width from points and rounding the
      marker footprint to output pixels.
- [x] Visual: focused `TestReferenceCompare/spy_marker` reports `RMSE 2.50`
      after refreshing the Go golden.
- [ ] Likely core areas: remaining residual is minor text/axis antialiasing.

### 8.15 `spy_image` (RMSE 19.42)

- [x] Code: source audited; no source mismatch found.
- [ ] Visual: sparsity image looks almost identical, but diff shows many
      one-pixel cell-edge shifts.
- [ ] Likely core areas: `MatShow` / `Spy` image extents, binary colormap
      rasterization, nearest interpolation and pixel boundary alignment.

### 8.16 `axes_top_right_inverted` (RMSE 5.06)

- [x] Code: source audited; mostly equivalent, with remaining differences in
      mirrored/inverted axis layout behavior.
- [ ] Visual: data marks align; residuals are small title/tick/text and y-label
      differences.
- [ ] Likely core areas: inverted-axis tick layout, mirrored axis tick/label
      positioning, text metrics and antialiasing.

### 8.17 `axes_control_surface` (RMSE 11.87)

- [x] Code: source audited; Go manually models Matplotlib `tick_top`,
      `tick_right`, `set_aspect`, `set_box_aspect`, `twinx`, and
      `secondary_xaxis`; remaining differences are core/layout behavior.
- [ ] Visual: strongest residuals are around the left axes box/ticks and right
      twin/secondary axes ticks/spines.
- [ ] Likely core areas: box aspect adjustment, tick-param propagation,
      twin/secondary-axis transforms and spine layout.

### 8.18 `transform_coordinates` (RMSE 10.99)

- [x] Code: removed the Go-only blended-text offset, used `fig.Text` for
      figure text, and matched annotation arrow/alignment plus grid styling.
- [ ] Visual: annotation/arrow and figure/axes/blended text placement differ
      clearly.
- [ ] Likely core areas: coordinate transforms, blended figure/axes coords,
      annotation offset-pixel handling, arrow rendering.

### 8.19 `figure_labels_composition` (RMSE 4.96)

- [x] Code: switched from manual subplot padding/spacing to
      `ConstrainedLayout()` + `Subplots(2, 2)` to mirror Python
      `constrained_layout=True`; legend locator remains the direct
      `bbox_to_anchor` translation.
- [x] Code: tightened constrained-layout line-box measurement, figure-level
      label auto-padding, rotated-label anchoring, text bbox stroke sizing, and
      inter-column/inter-row spacing to match Matplotlib's layout model.
- [x] Visual: focused `TestMatplotlibRef/figure_labels_composition` now reports
      PSNR 51.9 dB / MeanAbs 0.21; direct artifact RMSE is 4.96.
- [x] Likely core areas: constrained layout parity, figure-level labels, figure
      legend anchoring, text bbox sizing.

### 8.20 `colorbar_composition` (RMSE 18.60)

- [x] Code: translated the Python `imshow(..., aspect="auto", extent=...)`
      call through `ax.ImShow`; remaining differences are image/colorbar
      rendering and constrained-layout behavior.
- [ ] Visual: heatmap/colorbar are similar, but raster sampling and colorbar
      tick/label residuals cover much of the image.
- [ ] Likely core areas: image extent/aspect handling, colormap
      interpolation/normalization, colorbar layout and ticks.

### 8.21 `annotation_composition` (RMSE 8.90)

- [x] Code: matched explicit Python grid styling and annotation arrow width.
- [ ] Visual: annotation text/arrow placement differs, with smaller legend,
      grid, and text residuals.
- [ ] Likely core areas: annotation arrows, offset-pixel coords, default
      plot/legend styling, Unicode text metrics.

### 8.22 `patch_showcase` (RMSE 10.13)

- [x] Code: source audited; no direct example mismatch fixed. Remaining
      `FancyArrow(..., length_includes_head=True)` parity belongs in core patch
      semantics.
- [x] Code: added `FancyArrow.LengthIncludesHead` to model Matplotlib's
      `length_includes_head` option and set it explicitly in the translated
      patch showcase call.
- [x] Visual: refreshed `testdata/golden/patch_showcase.png`; focused
      `TestReferenceCompare/patch_showcase` now reports `RMSE 6.39`.
- [ ] Likely core areas: remaining residuals are hatches, ellipse/star/fancy
      box antialiasing, and alpha compositing.

### 8.23 `mesh_contour_tri` (RMSE 7.52)

- [x] Code: removed explicit half-step tick locator/formatter overrides so the
      example relies on core locator defaults like the Python source. Structured
      `Axes.Contour` now uses quad-grid contour lines and line z-order above
      filled contours.
- [x] Visual: focused `TestReferenceCompare/mesh_contour_tri` now reports
      `RMSE 7.52`; contourf/contour labels and triangulation coloring/lines
      remain the dominant residuals.
- [ ] Likely core areas: contour marching/fill bands, contour label placement,
      tripcolor flat shading, mesh edge strokes.

### 8.24 `plot_variants` (RMSE 7.11)

- [x] Code: replaced subplot-grid approximation with explicit `AddAxes`
      rectangles, replaced broken-bar `BarLabel` shortcuts with explicit
      `Text`, and matched stacked-bar label font/padding.
- [ ] Visual: fill-between-x polygon/edge, stairs/step edges, bar labels, grid,
      and text differ.
- [ ] Likely core areas: stairs and fill-between polygon construction, axline
      clipping, dash scaling, bar label padding.

### 8.25 `spectrum_variants` (RMSE 9.87)

- [x] Code: source audited; intent matches, but Python uses Matplotlib `mlab`
      spectrum helpers
      with explicit one-sided/two-sided handling and FFT normalization.
      `Axes.MagnitudeSpectrum` now matches Matplotlib's return contract by
      returning the linear magnitude spectrum while plotting dB-scaled values
      when `Scale: "dB"` is requested.
- [x] Visual: focused `TestReferenceCompare/spectrum_variants` reports
      `RMSE 9.87` after refreshing the Go golden; the catalog now enforces
      `MaxRMSE: 10.0` for this case.
- [x] Diagnostic: Go fixture samples differ from NumPy's generated samples at
      about 1e-16 because Go and NumPy/libm trig are not bit-identical; NumPy's
      FFT on Go-generated samples reproduces Go's near-zero-bin floor, while
      Go's FFT on NumPy-generated samples reproduces the Matplotlib magnitude
      floor. The fixture now uses the NumPy-generated signal, and the
      implementation-specific angle/phase residues are stored as fixture-only
      Matplotlib reference arrays.
- [x] Diagnostic: changing Line2D's default capstyle toward Matplotlib's
      `projecting` default barely moves `spectrum_variants` and pushes other
      non-mathtext fixtures over `RMSE 10`; leave this as a separate parity
      hardening task, not a Phase 8 spectrum fix.
- [ ] Likely core areas: general NumPy/pocketfft numerical parity for
      exact-zero angle/phase residues remains future hardening; the fixture is
      under the Phase 8 threshold without broad renderer or FFT changes.

### 8.26 `specialty_depth` (RMSE 9.42)

- [x] Code: removed the separate scatter workaround for `errorbar(fmt="o")`
      and matched boxplot flier size, violin edge color, and pie wedge
      edge/linewidth. Boxplot median defaults now match Matplotlib's
      `boxplot.medianprops.color = C1` and linewidth `1.0`. Errorbar limit
      markers now render as filled cap markers, and hexbin marginal bars draw
      above the main hex collection like Matplotlib.
- [x] Visual: focused `TestReferenceCompare/specialty_depth` reports
      `RMSE 9.42` after refreshing the Go golden.
- [ ] Likely core areas: errorbar limit caps, boxplot statistics/notches/fliers,
      violin KDE/side option, pie wedge styling, hexbin log scales/marginals.

### 8.27 `stem_plot` (RMSE 4.13)

- [x] Code: matched explicit grid styling and removed the Go-only legend label;
      `Axes.Stem` now matches Matplotlib's default `basefmt='C3-'` baseline and
      Line2D-style point marker diameter.
- [x] Visual: focused `TestReferenceCompare/stem_plot` now reports `RMSE 4.13`.
- [ ] Likely core areas: residual grid/tick/text antialiasing.

### 8.28 `specialty_artists` (RMSE 7.52)

- [x] Code: matched Python hexbin `mincnt=1`, pie white wedge
      edge/linewidth, removed extra labels, and simplified table styling to
      Python defaults where possible. Shared hexbin marginal draw-order and
      errorbar limit-marker fixes move the current render closer. Table cells
      now match Matplotlib's 1 pt patch linewidth, ink-bounds text anchoring,
      renderer-measured row-label auto width, and bbox scaling/offset behavior
      for auto row-label columns. Table cell paths now use Matplotlib-style
      auto patch snapping for rectilinear paths. Two-sided violin summary lines
      now use Matplotlib's default butt caps, violin body alpha is applied at
      the collection level, and tables draw as unclipped overlays like
      Matplotlib's `clip_on(False)` table artist.
- [x] Visual: focused `TestReferenceCompare/specialty_artists` reports
      `RMSE 7.52` after refreshing the Go golden.
- [ ] Likely core areas: event collection widths, hexbin bin geometry/mincnt and
      color normalization, pie wedge/text placement, violin KDE, table layout.

### 8.29 `units_overview` (RMSE 6.45)

- [x] Code: source audited; no substantive data/layout mismatch. Go uses units
      converters idiomatically for the same output.
- [ ] Visual: data align; diff is mostly text/tick labels and scatter/bar edge
      antialiasing.
- [ ] Likely core areas: unit formatters/locators, text metrics, marker/bar
      stroke rasterization.

### 8.30 `units_dates` (RMSE 5.92)

- [x] Code: added a `DayLocator` and changed the Go example to pass
      `time.Time` values through `FillBetweenUnits` / `PlotUnits`, matching the
      Python datetime + `mdates.DayLocator` structure without pre-conversion.
- [x] Code: unit-axis refresh now preserves explicit locator/formatter choices
      after date conversion, matching Matplotlib's default-axis-info guard.
- [ ] Visual: line/fill positions align; diff follows fill polygon edges, line
      strokes, and labels.
- [ ] Likely core areas: date units/date formatter parity, `FillBetween`
      edge/antialias behavior, text.

### 8.31 `units_categories` (RMSE 6.76)

- [x] Code: source audited; no major example mismatch. Go uses `BarUnits`
      including horizontal orientation as the idiomatic equivalent of Python
      `bar` / `barh`.
- [ ] Visual: bars align; diff is labels, grid/stroke edges, and bar outlines.
- [ ] Likely core areas: categorical units, horizontal `Bar2D` geometry/strokes,
      grid draw ordering, text.

### 8.32 `units_custom_converter` (RMSE 5.25)

- [x] Code: source audited; Go uses registered `TestDistanceKM` converter as
      the idiomatic equivalent of Python floats plus `FuncFormatter`.
- [ ] Visual: line/points align; marker fill/edge and tick/title text drive the
      diff.
- [ ] Likely core areas: custom unit `AxisInfo`, scatter marker area/edge
      linewidth, text metrics.

### 8.33 `vector_fields` (RMSE 7.99)

- [x] Code: changed barb length defaults to point units and quiver-key default
      label separation to the Matplotlib-equivalent 0.1 inch at 100 DPI, then
      removed the Go-only barb pre-scaling and explicit key label separation.
- [ ] Visual: fields align, but quiver/barb/stream glyph shapes and strokes
      differ, especially barbs and stream arrows.
- [ ] Likely core areas: `core/vector_field.go` quiver scaling/arrow polygons,
      barb decomposition, streamplot integration/arrows, quiver-key layout.

### 8.34 `polar_axes` (RMSE 6.22)

- [x] Code: source audited; `FillToBaseline` is the current idiomatic
      equivalent of Python `fill_between(..., 0)` for this case.
- [x] Code: matched Matplotlib polar defaults by hiding polar tick marks,
      removing polar minor tick locators, and drawing polar/radar grids between
      patch and line z-orders.
- [x] Code: polar radial spine is hidden, full-circle radial labels no longer
      receive tick padding, scalar radial tick labels use step precision, and
      theta tick labels are center-aligned with Matplotlib's `_pad + 7pt`
      padding.
- [x] Visual: focused `TestReferenceCompare/polar_axes` reports `RMSE 6.22`
      after refreshing the Go golden.
- [ ] Likely core areas: polar transforms, polar tick labels, polar grid paths,
      fill clipping/antialiasing.

### 8.35 `geo_mollweide_axes` (RMSE 7.98)

- [x] Code: source audited; Go helper relies on projection defaults while the
      Python explicitly sets ticks/formatters, but values match.
- [ ] Visual: line is close; diff concentrates on oval frame, gridline sampling,
      and labels.
- [ ] Likely core areas: `core/geo.go` Mollweide transform/frame/grid sampling,
      geo tick label transforms.

### 8.36 `geo_aitoff_axes` (RMSE 7.34)

- [x] Code: source audited; same projection-default versus explicit tick setup
      as Mollweide, with matching values.
- [ ] Visual: close overall; outer frame, meridians/parallels, and labels
      differ.
- [ ] Likely core areas: Aitoff projection transform, geo grid path sampling,
      label padding.

### 8.37 `geo_hammer_axes` (RMSE 6.98)

- [x] Code: source audited; same projection-default versus explicit tick setup
      as Mollweide, with matching values.
- [ ] Visual: sine trace aligns; frame/grid and text dominate the diff.
- [ ] Likely core areas: Hammer projection transform, geo frame/grid drawing,
      clipping.

### 8.38 `geo_lambert_axes` (RMSE 7.78)

- [x] Code: source audited; Go sets Lambert x locator and relies on projection
      formatter / y-label hiding while Python explicitly formats x only.
- [ ] Visual: line is close; circular frame/grid and longitude label placement
      differ.
- [ ] Likely core areas: `lambertDataTransform`, Lambert box aspect/frame, geo
      grid/text transforms.

### 8.39 `radar_basic` (RMSE 6.99)

- [x] Code: changed the fill call from baseline fill to polygon `Fill` to match
      Python `ax.fill(closed_angles, closed_values, ...)`; remaining
      projection/frame differences are core radar behavior.
- [x] Code: radar theta ticks keep Matplotlib's default tick size for label
      padding while tick marks remain hidden; theta labels are center-aligned
      with `_pad + 7pt` padding.
- [x] Visual: focused `TestReferenceCompare/radar_basic` reports `RMSE 6.99`
      after refreshing the Go golden.
- [ ] Likely core areas: radar projection configuration, polar grid
      interpolation, tick label padding, polygon fill/stroke rasterization.

### 8.40 `skewt_basic` (RMSE 5.02)

- [x] Code: Go source now directly mirrors the Python `set_yscale`,
      locator/formatter, and grid setup while relying on the core skewx
      projection for Matplotlib-compatible behavior.
- [x] Code: `AddSkewXAxes` keeps the top spine but hides top tick marks/labels,
      extends x-grid ticks through the skewed upper view interval, and positions
      y-axis ticks/labels on the left spine instead of through the skewed data
      transform.
- [x] Visual: focused `TestReferenceCompare/skewt_basic` reports `RMSE 5.02`
      after refreshing the Go golden.
- [ ] Likely core areas: residual legend/text antialiasing and line/grid
      rasterization.

### 8.41 `mplot3d_basic` (RMSE 3.76)

- [x] Code: added Python `cmap="viridis"` to the Go `Surface(... Alpha)`
      call.
- [x] Code: 3D tick labels and axis labels now compute Matplotlib-style
      label-offset centers/deltas from the expanded frame limits, not the
      projection/view limits.
- [x] Visual: focused `TestReferenceCompare/mplot3d_basic` reports `RMSE 3.76`
      after refreshing the Go golden.
- [ ] Likely core areas: remaining residual is subpixel text/rasterization and
      minor 3D data-artist antialiasing.

### 8.42 `mplot3d_terrain` (RMSE 2.79)

- [x] Code: source audited; fixture bodies mostly match and Go
      axes-coordinates text corresponds to Python `text2D(..., transAxes)`.
- [x] Code: shared 3D tick/axis label offsets now use expanded frame limits.
- [x] Visual: focused `TestReferenceCompare/mplot3d_terrain` reports
      `RMSE 2.79` after refreshing the Go golden.
- [ ] Likely core areas: remaining residual is minor contour/surface
      rasterization and subpixel frame/text antialiasing.

### 8.43 `mplot3d_plot3d` (RMSE 1.19)

- [x] Code: source audited; no obvious source mismatch found.
- [x] Code: shared 3D tick/axis label offsets now use expanded frame limits.
- [x] Visual: focused `TestReferenceCompare/mplot3d_plot3d` reports
      `RMSE 1.19` after refreshing the Go golden.
- [ ] Likely core areas: residual line/text antialiasing.

### 8.44 `mplot3d_scatter3d` (RMSE 1.51)

- [x] Code: resolved the deterministic source-parity issue by using the shared
      Go-compatible PCG stream in the Python reference and generating the Go
      scatter values from the same stream instead of hardcoding point arrays.
- [x] Code: matched Matplotlib `Axes3D.scatter` defaults by applying
      `s=20` for 3D scatter calls and depth-shading/z-sorting edge colors along
      with face colors.
- [x] Code: shared 3D tick/axis label offsets now use expanded frame limits.
- [x] Visual: refreshed stale `testdata/matplotlib_ref/mplot3d_scatter3d.png`
      and `testdata/golden/mplot3d_scatter3d.png`; latest
      `TestReferenceCompare/mplot3d_scatter3d` reports `RMSE 1.51`.
- [ ] Likely core areas: remaining residual is marker/text antialiasing rather
      than source fixture drift.

### 8.45 `mplot3d_surface3d` (RMSE 2.89)

- [x] Code: source audited; fixture data/options match. Remaining default-alpha
      difference is core `Surface` behavior.
- [x] Code: shared 3D tick/axis label offsets now use expanded frame limits.
- [x] Visual: focused `TestReferenceCompare/mplot3d_surface3d` reports
      `RMSE 2.89` after refreshing the Go golden.
- [ ] Likely core areas: residual surface/text antialiasing and minor colormap
      raster differences.

### 8.46 `mplot3d_wire3d` (RMSE 1.29)

- [x] Code: source audited; no obvious mismatch. Go helper mirrors upstream
      `axes3d.get_test_data`.
- [x] Code: shared 3D tick/axis label offsets now use expanded frame limits.
- [x] Visual: focused `TestReferenceCompare/mplot3d_wire3d` reports
      `RMSE 1.29` after refreshing the Go golden.
- [ ] Likely core areas: residual line/text antialiasing.

### 8.47 `mplot3d_trisurf3d` (RMSE 1.93)

- [x] Code: added automatic Delaunay triangulation for `Trisurf` when triangles
      are omitted, then removed the manual fan/ring triangle construction from
      the example to match Python `plot_trisurf(x, y, z, ...)`.
- [x] Code: shared 3D tick/axis label offsets now use expanded frame limits.
- [x] Visual: focused `TestReferenceCompare/mplot3d_trisurf3d` reports
      `RMSE 1.93` after refreshing the Go golden.
- [ ] Likely core areas: residual facet/text antialiasing.

### 8.48 `mplot3d_bar3d` (RMSE 3.20)

- [x] Code: source audited; fixture values match.
- [x] Code: shared 3D tick/axis label offsets now use expanded frame limits.
- [x] Visual: focused `TestReferenceCompare/mplot3d_bar3d` reports
      `RMSE 3.20` after refreshing the Go golden.
- [ ] Likely core areas: residual face/text antialiasing.

### 8.49 `mplot3d_voxels` (RMSE 2.93)

- [x] Code: source audited; fixture values/options match.
- [x] Code: shared 3D tick/axis label offsets now use expanded frame limits.
- [x] Visual: focused `TestReferenceCompare/mplot3d_voxels` reports
      `RMSE 2.93` after refreshing the Go golden.
- [ ] Likely core areas: residual face/edge/text antialiasing.

### 8.50 `mplot3d_quiver3d` (RMSE 3.38)

- [x] Code: source audited; loop order appears to match NumPy `meshgrid`
      flattening.
- [x] Code: shared 3D tick/axis label offsets now use expanded frame limits.
- [x] Visual: focused `TestReferenceCompare/mplot3d_quiver3d` reports
      `RMSE 3.38` after refreshing the Go golden.
- [ ] Likely core areas: residual arrow/text antialiasing.

### 8.51 `mplot3d_stem3d` (RMSE 3.61)

- [x] Code: source audited; `Stem3D` now matches Matplotlib's default
      `basefmt='C3-'`, Line2D marker point diameter, and Line3DCollection butt
      capstyle for stem lines.
- [x] Code: shared 3D tick/axis label offsets now use expanded frame limits.
- [x] Visual: focused `TestReferenceCompare/mplot3d_stem3d` now reports
      `RMSE 3.61` after refreshing the Go golden.
- [ ] Likely core areas: residual stem marker/line/text antialiasing.

### 8.52 `mplot3d_fill_between3d` (RMSE 3.50)

- [x] Code: removed the explicit fill color so the Go call mirrors Python
      `fill_between(..., alpha=0.5)`; remaining difference is core 3D ruled
      surface shading/z-order.
- [x] Code: shared 3D tick/axis label offsets now use expanded frame limits.
- [x] Visual: focused `TestReferenceCompare/mplot3d_fill_between3d` reports
      `RMSE 3.50` after refreshing the Go golden.
- [ ] Likely core areas: residual fill/text antialiasing.

### 8.53 `unstructured_showcase` (RMSE 7.87)

- [x] Code: changed `TriColor` edge color alpha to opaque white to match Python
      `edgecolors="white"`; larger remaining mismatch is core `tricontour`
      behavior.
- [ ] Visual: triplot is close; tripcolor/tricontour and tricontourf have
      different topology, label placement, and filled-band shapes.
- [ ] Likely core areas: `core/contour.go` tri contour/fill generation and
      inline labels, `core/triangulation.go` tripcolor edge/color handling.

### 8.54 `arrays_showcase` (RMSE 8.73)

- [x] Code: source audited; heatmap, mesh, and spy data match. Spy marker
      panel picked up the Line2D marker sizing fix; contour lines now use
      structured-grid cell clipping instead of triangulating quads.
- [x] Visual: center contour paths now draw above the pcolormesh cells by
      defaulting line contours to Matplotlib's line z-order. Focused
      `TestReferenceCompare/arrays_showcase` now reports `RMSE 8.73`.
- [ ] Likely remaining areas: contour label placement/inline clipping, text and
      mesh antialiasing.

### 8.55 `axisartist_showcase` (RMSE 13.25)

- [x] Code: replaced floating cloned axes with explicit `AxHLine` / `AxVLine`,
      matched dash/linewidth values, light y-grid color, tick direction, and
      axes-fraction note text/bbox.
- [ ] Visual: plot geometry is close, but gridlines, dashed zero axes,
      tick/text antialiasing, and legend/text box pixels dominate.
- [ ] Likely core areas: `core/axis_artist.go`, `core/axis.go` dash/line-width
      parity for data-position spines, `core/grid.go` grid style configuration.

### 8.56 `axes_grid1_showcase` (RMSE 10.73)

- [x] Code: matched tile label font size and `round,pad=0.25` bbox semantics;
      remaining `axes_pad` inch-vs-fraction behavior belongs in
      `core/image_grid.go`.
- [ ] Visual: near-identical overall; diff is mostly text/bbox/tick/axis
      antialiasing and small label-box padding.
- [ ] Likely core areas: `core/image_grid.go` inch-vs-fraction divider spacing,
      anchored text bbox padding, renderer text/line antialiasing.

### 8.57 `pcolor_flat` (RMSE 8.74)

- [x] Code: source audited; remaining Go `PColor` as `PColorMesh` versus
      Matplotlib `PolyQuadMesh` is core mesh/collection behavior.
- [ ] Visual: data cells align, but every cell edge/stroke and text edge
      differs.
- [ ] Likely core areas: `core/mesh.go`, `core/collection.go` pcolor versus
      pcolormesh semantics, quad edge antialias/snap.

### 8.58 `pcolormesh_gouraud` (RMSE 4.78)

- [x] Code: source audited; fixture inputs match. `QuadMesh.drawGouraudMesh`
      now matches Matplotlib's `_convert_mesh_to_triangles` by splitting each
      quad into four center-fan Gouraud triangles with averaged center RGBA.
- [x] Visual: focused `TestReferenceCompare/pcolormesh_gouraud` now reports
      `RMSE 4.78`; remaining residual is mostly colorbar/text and minor
      rasterization drift.
- [ ] Likely core areas: AGG Gouraud rasterization and colorbar/text
      antialiasing.

### 8.59 `hist2d_weighted_density` (RMSE 7.48)

- [x] Code: source audited; no obvious fixture/binning mismatch and weighted
      density setup matches Python.
- [ ] Visual: mesh bins align and values look correct; differences are mainly
      colorbar, tick/text, and edge rendering.
- [ ] Likely core areas: colorbar axis rendering, text metrics, quad mesh
      antialiasing.

### 8.60 `boundarynorm_pcolormesh` (RMSE 5.86)

- [x] Code: source audited; data, boundaries, `BoundaryNorm`, and pcolormesh
      setup match.
- [ ] Visual: discrete bands align; residual is mostly colorbar ticks/label,
      border, and text.
- [ ] Likely core areas: `core/colorbar.go`, `core/norm.go` BoundaryNorm
      colorbar rendering, axis/text placement.

### 8.61 `lognorm_imshow` (RMSE 9.42)

- [x] Code: source audited; fixture values and `LogNorm(1,1000)` match.
      `LogFormatter` now emits Matplotlib-style base-10 power labels instead
      of `1eN` labels for exact decades.
- [x] Visual: focused `TestReferenceCompare/lognorm_imshow` reports
      `RMSE 9.42` after refreshing the Go golden.
- [ ] Likely core areas: remaining residual is minor image/colorbar
      rasterization and text antialiasing drift.

### 8.62 `twoslope_norm_image` (RMSE 6.24)

- [x] Code: source audited; data, custom diverging colormap stops, and
      `TwoSlopeNorm(-3,0,6)` match. `Figure.AddColorbar` now installs a
      Matplotlib-like function scale for non-linear continuous norms so the
      colorbar axis transforms through `norm`/`inverse`.
- [x] Visual: focused `TestReferenceCompare/twoslope_norm_image` reports
      `RMSE 6.24` after refreshing the Go golden.
- [ ] Likely core areas: remaining residual is minor image/colorbar
      rasterization and text antialiasing drift.

### 8.63 `colorbar_extensions` (RMSE 7.00)

- [x] Code: source audited; fixture requests `Extend: "both"` like Python
      `extend="both"`. `core/colorbar.go` now draws extension patches outside
      the clipped artist pass and shrinks the inner colorbar axes for extension
      space, matching Matplotlib's colorbar layout model.
- [x] Code: extended colorbars now compensate the box aspect by the extension
      shrink factor, matching Matplotlib's `_ColorbarAxesLocator` behavior.
- [x] Visual: focused `TestReferenceCompare/colorbar_extensions` reports
      `RMSE 7.00` after refreshing the Go golden.
- [ ] Likely core areas: residual colorbar outline/text antialiasing.

**Exit criteria:**

- [ ] Every listed subphase is either fixed in core behavior or explicitly
      justified as an intentional divergence. Source parity has been audited
      and direct example mismatches have been fixed where current APIs allow.
- [ ] `TestReferenceCompare` records no catalog case above `RMSE 5.00` unless
      that case has a documented, frozen tolerance exception.
- [ ] All fixes are validated against both source parity and visual artifacts,
      not just metric deltas.

---

# Phase 8A: Cross-Fixture Parity Hardening

✅ **Completed.** Individual RMSE fixes are now governed by reusable
cross-fixture validation rules so parity work cannot overfit one catalog case.

Completed scope:

- Parity-fix policy forbids example-source workarounds, fixture-specific core
  branches, catalog-ID conditionals, and unexplained empirical constants.
- `internal/examplecatalog.ValidationClusters` defines stable validation groups
  for layout/text, image/mesh/colorbar, and projection/3D work.
- Catalog tests enforce cluster membership, required validation targets, and
  the absence of quoted catalog IDs in non-test implementation files.
- Accepted layout/text empirical corrections were replaced with source-backed
  Matplotlib models for constrained-layout padding, figure-label
  autopositioning, and figure-artist spacing.
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
axis and display-coordinate residuals.

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

- [x] Mutable collection/scalar-mappable state updates propagate
      deterministically.
- [x] Mesh shading/shape rules and colorbar orientation/boundary/tick behavior
      are source-backed and fixture-covered.
- [x] Advanced norm/color-helper scope is implemented or explicitly classified.
- [x] Audit and public-surface rows for `collections.py`, `cm.py`, `colors.py`,
      `colorbar.py`, and `colorizer.py` are closed to precise statuses.

---

# Phase 15: Patches, Text, Annotation, Legend, and Offset Boxes (Core Closure)

**Goal:** complete the non-coordinate-boundary portion of former 12.4 (12.4A-F)
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
geometry follows upstream Matplotlib semantics.

**Reference sources:** `third_party/matplotlib/lib/matplotlib/transforms.py`,
`patches.py`, `text.py`, `backend_bases.py`, `_backend_agg.cpp`, local
`core/artist.go`, `core/arrow_patch.go`, `core/text.go`, `backends/agg/`, and
when needed `../agg_go` vs `../agg_2.4`.

- [x] Define and document the coordinate contract between core display-space
      geometry and renderer backends. (ADR `docs/adr/0003-display-coordinate-contract.md`)
- [ ] Add renderer-neutral regressions for signed display-space geometry:
      `ConnectionStyle("arc3", rad=...)`, arrow shrink/clip on curves,
      arrow-head normals, rotated-text bbox orientation, annotation arrow start
      after bbox clipping. (G8, pending)
- [x] Audit and remove non-source-backed y-sign compensations in core,
      transforms, and fixtures. (core gradients/colorbar/image/layout/legend/
      widgets converted; `text_annotation_matrix` `OffsetY` hack removed → 1:1)
- [x] Reconcile `text_annotation_matrix` bbox-arrow behavior against upstream
      `Annotation.update_positions` and
      `FancyArrowPatch._get_path_in_displaycoord`. (MeanAbs 1.39, PSNR 45.9 dB)
- [x] Validate boundary fixes across `patch_style_matrix` (1.15),
      `annotation_composition` (<0.7), and `transform_coordinates` (<0.7).
- [ ] If residuals are AGG-port specific, fix `../agg_go` directly instead of
      compensating in this repository. (pending: rotated vertical-label glyph
      orientation; widgets selector shapes)
- [x] Track before/after metrics; immediate target:
      `text_annotation_matrix` `RMSE < 10` against Matplotlib reference with
      1:1 example sources. (achieved: PSNR 45.9 dB ≈ RMSE 1.3, was 14.94)

Execution track (kept from former 12.4G):

Status: [x] done · [~] in progress · [ ] todo.

- [x] G1 Contract & core pivot.
- [x] G2 AGG backend owns device flip.
- [x] G3 Core positioning/text helpers y-up conversion.
- [~] G4 AGG parity validation. (parity suite validated — only 4 fixtures >2
      MeanAbs, all classified; AGG backend unit/golden tests still need y-up
      reconciliation: transformed-image/TeX pixel expectations, path-effect and
      pattern-gradient backend goldens.)
- [~] G5 Example 1:1 port sweep. (`text_annotation_matrix`,
      `pattern_gradient_effects` ported; broader sweep pending.)
- [ ] G6 Vector/other backend inversion ownership. (pdf/ps/pgf/svg/gobasic
      backend tests currently pass under y-up — verify output correctness.)
- [ ] G7 Full-suite regen and revalidation.
- [ ] G8 Renderer-neutral signed-geometry regression set.

Residual parity offenders (TestMatplotlibRef MeanAbs, 2026-05-25):
`colorbar_horizontal_ticks` 7.28, `widgets_gallery` 6.41,
`imshow_interpolation_matrix` 4.31, `pattern_gradient_effects` 3.09. All other
fixtures <2.0.

Exit criteria:

- [x] Signed display-space paths/annotations/text bboxes/arrow geometry are
      source-backed under the documented coordinate contract.
- [x] No `text_annotation_matrix`-specific sign hacks exist in example or core.
- [x] `TestMatplotlibRef/text_annotation_matrix` reports `RMSE < 10` without
      regressions in related fixtures.
- [~] Remaining mismatch is classified with evidence as core, renderer
      boundary, AGG-port, or upstream limitation. (4 residual fixtures listed
      above; AGG backend golden/unit failures pending classification.)

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
