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
- [ ] Decide on `figimage` and `pcolorfast`: implement or document as
      intentional omissions.
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
- [ ] Tighten text bounding-box, baseline, rotation-mode, multiline, wrapping,
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

- [ ] Audit upstream `image.py` image artist classes and helper functions:
      `AxesImage`, `FigureImage`, `BboxImage`, `NonUniformImage`,
      `PcolorImage`, `imread`, `imsave`, `imresize`-style omissions,
      `pcolorfast`, `figimage`, and image origin/extent/aspect defaults.
- [ ] Decide which non-`AxesImage` classes belong in v1.0:
      implement `FigureImage` / `figimage` and `BboxImage` if they improve
      figure-level composition or annotation-box parity; implement
      `NonUniformImage` / `PcolorImage` / `pcolorfast` only if a visible
      fixture shows a meaningful difference from `Image2D` / `PColorMesh`.
- [ ] Convert every unsupported image class/helper into either a typed Go
      equivalent, a clear runtime error, or a Phase 11 intentional omission
      with migration guidance.
- [x] Lock interpolation coverage for `nearest`, `none`, `bilinear`,
      `bicubic`, `hanning`, `hamming`, `lanczos`, `spline16`, `spline36`,
      `kaiser`, `quadric`, `catrom`, `gaussian`, `bessel`, `mitchell`, `sinc`,
      `blackman`, `hermite`, `antialiased`, and `auto` with parser tests,
      AGG-render tests, and at least one visible interpolation gallery case.
- [ ] Verify transformed-image sampling, image alpha premultiplication, clipping,
      resample threshold decisions, and backend fallbacks against upstream
      `image.py` and AGG behavior.
- [ ] Add catalog/parity fixtures for figure-level images or bbox images if
      implemented, plus an image-interpolation matrix fixture that is small
      enough to debug visually.

#### 12.5C Pyplot and Stateful Wrapper Surface

- [ ] Audit high-traffic upstream `pyplot.py` and `_pylab_helpers.py` functions
      against the Go `pyplot` package; rank missing wrappers by migration value
      rather than trying to clone every overload.
- [ ] Add wrappers for implemented core features where the current absence is
      a migration blocker: common axes creation, subplot/subplots variants,
      plotting/image/stat helpers, axis/tick scale helpers, annotations,
      legends, colorbars, rc/style helpers, and save/show helpers.
- [ ] Preserve object-oriented APIs as the source of truth; pyplot wrappers
      should delegate to `core` without maintaining duplicate rendering or
      layout behavior.
- [ ] Define clear error behavior for unsupported pyplot overloads, implicit
      figure-manager behavior, interactive mode toggles, and global state reset
      functions.
- [ ] Add unit tests that pyplot wrappers delegate to the same core state as
      the object-oriented calls, plus migration-style examples for common
      pyplot workflows.
- [ ] Update public-surface parity rows for `pyplot.py` and `_pylab_helpers.py`
      so each broad wrapper gap is either implemented, scoped to a smaller
      wrapper family, or intentionally omitted.

#### 12.5D Backend Canvas, Manager, and Tool Lifecycle

- [ ] Audit upstream `backend_bases.py`, `backend_tools.py`, and
      `_pylab_helpers.py` against `canvas`, backend registry metadata, and the
      interactive backend implementations.
- [ ] Complete canvas/manager lifecycle semantics for figure creation,
      current-manager tracking, draw vs draw-idle, resize, close/destroy,
      toolbar attachment, save dispatch, and backend capability reporting.
- [ ] Tighten event contracts for mouse, key, scroll, pick, figure enter/leave,
      axes enter/leave, timer events, and callback connection/disconnection.
- [ ] Complete toolbar/tool behavior needed by interactive backends: home,
      back, forward, pan, zoom, configure, save, cursor/status messages, mode
      state, and tool enablement.
- [ ] Add backend-neutral lifecycle tests using fake canvases/managers and
      smoke tests for WebAgg/Gio where the environment allows them.
- [ ] Document backend-specific omissions, especially GUI toolkit behaviors
      that cannot be represented in the current headless test environment.

#### 12.5E Widgets and Interaction Scope

- [x] Audit upstream `widgets.py` classes and decide v1.0 status for each:
      button, slider, range slider, check buttons, radio buttons, text box,
      span selector, rectangle selector, lasso selector, polygon selector,
      cursor/multi-cursor, annotated cursor, menu/tool widgets, and any
      deprecated or GUI-specific widgets.
- [ ] Tighten existing widget behavior for callback ordering, active/disabled
      state, hover/press/release transitions, keyboard activation, value
      clamping, snapping, dragging, redraw policy, and axes ownership.
- [x] Implement only selector/cursor widgets that can be expressed through the
      current canvas event model; document GUI-only or callback-heavy widgets as
      intentional omissions until an interactive fixture requires them.
- [ ] Add catalog/browser-demo coverage for the supported widget set and
      renderer-neutral tests for interaction state transitions.
- [ ] Ensure widgets compose with artist picking, overlays, draw-idle, and
      animation without stealing unrelated user events.

#### 12.5F Animation Scope and Writers

- [x] Audit upstream `animation.py` for `Animation`, `TimedAnimation`,
      `FuncAnimation`, `ArtistAnimation`, frame sequence behavior, repeat /
      repeat-delay, blitting, save_count/cache behavior, HTML representation,
      and movie writer APIs.
- [ ] Tighten `FuncAnimation` and `ArtistAnimation` behavior against upstream
      for initialization order, frame iteration, repeat semantics, event-source
      lifecycle, animated artist visibility, and blit background restoration.
- [x] Decide v1.0 writer scope: implement a small explicit writer surface for
      GIF/MP4 only if dependencies and backend output are deterministic, or
      document animation saving as intentionally omitted while interactive
      playback remains supported.
- [ ] Add examples and browser demos for at least one timer-driven line update,
      one artist-list animation, and one blit-capable animation path.
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
- [ ] Every unsupported image interpolation, pyplot wrapper family, widget, or
      animation writer path has a clear error or documented omission; none
      silently fall back to an incorrect behavior.
- [ ] 12.5 has catalog/parity, example, or browser-demo coverage for image
      class/resampling decisions, pyplot migration workflows, backend
      lifecycle semantics, supported widgets, and animation playback.
- [ ] `go test ./core ./pyplot ./canvas ./animation ./backends/agg`
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
  row now only carries pcolorfast / transformed-resampling edge behavior as
  remaining partial scope.
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
- Added pyplot `FigLegend` over the current figure-level legend path.
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
- Added `Animation.Save` as an explicit unsupported-writer path returning
  `ErrWriterUnsupported`, with unit coverage. This keeps writer export out of
  v1.0 while avoiding silent no-op behavior for users who look for a save path.
- Tightened `Animation.Start` lifecycle error handling so a timer start failure
  returns the backend/event-loop error without leaving the animation marked
  running or holding a stale timer.
- Fixed `ButtonOptions.Disabled` propagation so constructor-specified disabled
  state participates in the same interaction/picking behavior as runtime
  `Button.Enabled=false`.
- Tightened disabled slider interaction handling so focused `Slider` and
  `RangeSlider` widgets ignore keyboard nudges without firing callbacks or
  requesting redraws, matching the mouse-drag disabled path.
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

### 13.1 Core Plot Family Galleries

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

### 13.2 Color, Image, Text, and Annotation Galleries

- [ ] Add named-color swatches and colormap family galleries.
- [ ] Add image interpolation/alpha/matshow/spy galleries.
- [ ] Add colorbar norm/extension galleries.
- [ ] Add MathText, text layout, annotation, legend, and offset-box galleries.

Implementation notes:

- Prefer several focused examples over one overloaded gallery when visual
  differences need inspection.
- Include captions/descriptions in catalog metadata explaining what feature
  breadth the example validates.

### 13.3 Toolkit, Projection, 3D, and Backend Output Galleries

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

# Phase 14: Browser Gallery Alignment

**Goal:** make the browser gallery a catalog-backed inspection surface for the
same feature families covered by parity fixtures and CLI examples.

### 14.1 Wire Planned Web Reference Modules

- [ ] Wire `test/matplotlib_ref/webdemos/annotations.py`, `bars.py`,
      `errorbars.py`, `fills.py`, `heatmap.py`, `histogram.py`, `lines.py`,
      `patches.py`, `scatter.py`, and `subplots.py` into active browser demos
      or fold them into existing catalog-backed browser families.
- [ ] Keep `radialforce.py` reference-only until it is promoted to a catalog
      case.
- [ ] Add tests that every active web reference module maps to a catalog case
      and every catalog-backed planned row either has an active browser demo or
      remains explicitly planned.

### 14.2 Promote CLI-Only Showcases

- [ ] Promote CLI-only showcases listed in `BrowserDemoCoverageRows` into
      browser demos: basic lines, dashes, scatter, bars, fills, errorbars,
      multi-series, histograms, boxplots, heatmaps, figure labels, colorbars,
      annotations, projections, mplot3d, triangulation, axisartist, and
      axes_grid1.
- [ ] Browser demos must use the same catalog factories as parity tests or a
      documented wrapper around them.
- [ ] Add browser-demo smoke tests that render each promoted demo and verify it
      has a non-empty image/artifact.

### 14.3 Browser Parity Status Reporting

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

# Phase 15: Documentation, Examples Polish, and v1.0 Release

**Goal:** make the project consumable by users who have not been following
the development thread, and tag a stable v1.0.

### 15.1 API Documentation

- [ ] Package-level GoDoc passes for every public package, with a worked
      example per package.
- [ ] Hosted documentation site (pkg.go.dev plus a curated landing page
      under the existing GitHub Pages deployment).
- [x] Migration guide from upstream Matplotlib: side-by-side Python / Go
      snippets for every plot family covered by the catalog.
- [x] Backend selection guide: when to use AGG / GoBasic / SVG / PDF /
      Skia, with capability matrix excerpts (`docs/backend-selection.md`).

### 15.2 Examples Gallery Polish

- [x] Review every `Showcase: true` catalog row for caption, description,
      and runnable snippet quality.
- [x] Add an "anti-gallery" of intentional Matplotlib-divergence cases with
      the reasons documented (where the Go port chose different defaults).
- [x] Promote the WASM browser gallery to a first-class entry point on the
      project README.

### 15.3 Performance Pass

- [ ] Profiling sweep across the catalog: identify hotspots that exceed the
      100k-point smoothness goal and the sub-second typical-plot goal.
- [ ] Reusable benchmark suite under `benchmarks/` with regression tracking
      in CI.
- [ ] Documented memory-usage targets and a tuning guide for long-running
      applications.

### 15.4 Release Readiness

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
