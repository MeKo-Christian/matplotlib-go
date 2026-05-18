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

**What's left** is the focused work in the phases below: PDF / vector output
beyond SVG, interactive backends and animation, widget interaction, MathText
and TeX completion, renderer effects (patterns / gradients / path effects),
final backend deepening, and the documentation polish for v1.0.

---

# Phase 1: PDF and Additional Vector Output

**Goal:** add the publication-quality vector formats that are still missing
after SVG landed, and finish the save-dispatch story so all formats route
through the same backend / canvas pipeline.

**Reference sources:** `third_party/matplotlib/lib/matplotlib/backends/backend_pdf.py`,
`backend_ps.py`, `backend_pgf.py`, current `backends/svg/` for serialization
patterns.

### 1.1 PDF Backend

- [x] Scaffold `backends/pdf/` following the SVG layout: deterministic
  serialization, dedicated `init.go` registration, `doc.go` capability notes.
- [x] Page model and content stream encoder: graphics state stack, path
  construction (`m`/`l`/`c`/`re`/`h`/`f`/`S`), clip operators, color spaces.
- [x] Embedded font subsetting (Type 0 / CIDFontType2) using the shared font
  resource strategy that already drives AGG and SVG.
- [x] Text-as-path opt-in via `render.TextPather` so PDF can render typography
  without an embedded font program. Real text-as-text via embedded fonts
  becomes the default once subsetting lands.
- [x] Image XObjects with `/FlateDecode` and `/DCTDecode`, including direct
  grayscale `/SMask` images for RGBA alpha.
- [x] Native hatch via tiling pattern color spaces.
- [x] Native marker / path collection batches via form XObjects.
- [x] Metadata, `SOURCE_DATE_EPOCH`, and reproducible-output options on par
  with the SVG renderer (`render.PDFOptions`, `WithPDFFontPolicy`,
  `WithPDFMetadata`, `WithPDFCreationDate`).
- [x] Structural test harness (analogous to `internal/svgcompare/`) for
  whitespace-insensitive PDF object comparison.
- [x] Golden fixtures for line, bar, scatter, hist, contour, imshow, polar,
  hatch_bars, text_layout, clipped, and image_transformed cases.

Current slice landed:

- `backends/pdf` package with `doc.go`, `init.go`, `pdf.go`, `pdf_test.go`,
  `registry_test.go`. Renderer implements `render.Renderer`,
  `render.PDFExporter`, `render.PDFOptionExporter`, `render.PDFOptionSetter`,
  and `render.DPIAware`.
- `backends.PDF` backend constant and `backends.PDFExport` capability with a
  runtime-checked `PDFExporter` interface mapping. `VectorOutput` capability
  now accepts either `SVGExporter` or `PDFExporter`.
- PDF document writer with deterministic xref/trailer, single-page Catalog /
  Pages / Page / Contents / Info object layout, FlateDecode-compressed content
  stream, compact `shortFloat` formatter, quadratic-to-cubic curve promotion,
  literal-string / name escaping, and `SOURCE_DATE_EPOCH`-aware
  `/CreationDate`.
- Save dispatch routes `.pdf` through `backends.SavePDF` and the registry
  `saveViaExtension` fallback. `backends/all/all.go` side-imports the new
  package.
- Text-as-path output now routes `TextPath`, font-keyed text, rotated text,
  vertical text, and simple `GlyphRun` fallback through the shared font outline
  pipeline. PDF advertises the matching text capabilities only via runtime
  checked renderer interfaces.
- `render.WithPDFFontPolicy(render.PDFFontPolicyEmbed)` now emits real PDF
  text objects backed by deterministic Type 0 / CIDFontType2 resources,
  embedded TrueType font programs, `/CIDToGIDMap`, `/W`, and `/ToUnicode`
  maps. The practical first slice subsets the resource CID map to used glyphs
  while embedding the resolved font program bytes intact.
- Raster `Image` draws now emit `/Image` XObjects with deterministic
  `/FlateDecode` streams and grayscale soft masks for RGBA alpha. JPEG
  `/DCTDecode` passthrough is supported through the optional
  `render.JPEGImage` interface. Flate image streams include PNG predictor
  `/DecodeParms`, matching Matplotlib's compressed image path.
- New `internal/pdfcompare` structural comparison helper parses indirect
  objects, ignores xref offset noise, normalizes object token whitespace, and
  decodes `/FlateDecode` streams before comparison. PDF backend tests exercise
  it against generated output.
- Native hatch fills now emit reusable PDF tiling pattern resources
  (`/PatternType 1`) with deterministic pattern names and decoded structural
  coverage through `internal/pdfcompare`.
- Marker and path collection batches now emit reusable Form XObjects,
  registered in the page `/XObject` resource dictionary and invoked with
  per-item paint state, matching Matplotlib's PDF batching strategy.
- Transformed images now implement `render.ImageTransformer` by reusing image
  XObjects with arbitrary PDF `cm` matrices, covering rotated image placement.
- Stroke/fill alpha now follows Matplotlib's PDF ExtGState model: paint draws
  register reusable `/ExtGState` resources with separate `CA` / `ca` values and
  select them through the `gs` operator.
- Repeated raster image draws reuse a single image XObject when the encoded
  RGB/alpha payload matches, mirroring Matplotlib's image-object cache.

### 1.2 PostScript and PGF Export

- [ ] PostScript (`backends/ps/`) for journal submissions that still require
  EPS. Scope limited to level-2 PS with the same font / image / hatch
  semantics as PDF.
- [ ] PGF / LaTeX export (`backends/pgf/`) for direct inclusion in LaTeX
  documents. Decision required on whether to ship as a generator-only backend
  or invoke `lualatex` for verification.
- [ ] Optionally vector / mixed-mode output: rasterized fallback for
  effects PDF/PS cannot represent (driven by renderer capability checks, not
  backend-name conditionals).

### 1.3 Save Dispatch Cleanup

- [ ] Remove the last hard-coded format paths in callers; every save route
  through `backends.SelectBackendForExtension` and the `SaveFormats` map.
- [ ] Expand `BackendComparisonReport` to enumerate PDF / PS / PGF capability
  status alongside AGG / GoBasic / SVG / Skia.
- [ ] Add `cmd/example -format pdf|ps|pgf` smoke coverage matching the
  existing PNG / SVG runners.

**Exit criteria:**

- [ ] `core.SaveFig(fig, "out.pdf")` works end-to-end with deterministic
  output and Matplotlib-comparable visual fidelity.
- [ ] PDF, PS, and PGF backends are reachable through the registry without
  any caller knowing about them.
- [ ] Vector backends share a documented font / hatch / image semantics
  surface so future formats are additions instead of rewrites.

---

# Phase 2: Renderer Effects, Patterns, and Compositing

**Goal:** finish the renderer-depth cleanup deferred from earlier phases so
artists can request pattern / gradient fills and post-render path effects
without backend-name conditionals.

**Reference sources:** `third_party/matplotlib/lib/matplotlib/patheffects.py`,
`third_party/matplotlib/lib/matplotlib/colors.py` (gradient stops),
`backends/agg/` filter support, `backends/svg/` `<pattern>` / `<filter>` defs.

### 2.1 Pattern and Gradient Fills

- [x] Renderer-neutral pattern fill API in `render/`: tile geometry, tile
  transform, and tile color description that AGG, SVG, PDF, and Skia can each
  implement natively. (`PatternFill` on `Paint` plus `PatternFiller` capability
  interface; `GraphicsContext.WithFillPattern` propagates through alpha/forced
  alpha just like solid fills.)
- [x] Linear and radial gradient fill description (stops, transform, spread
  method) routed through the same capability interface. (`GradientFill` on
  `Paint` plus `GradientFiller` capability interface; `GraphicsContext.WithFillGradient`
  applies the same forced-alpha bookkeeping as patterns and solid fills.)
- [x] AGG implementation using existing gradient span generators.
  Two-stop linear and radial gradients route through Agg2D's gradient API; a
  three-stop radial uses the multi-stop variant. `SupportsGradientFill`
  advertises native support; `SupportsPatternFill` advertises `false` until
  tile rasterization lands.
- [x] SVG implementation via `<linearGradient>` / `<radialGradient>` /
  `<pattern>` defs. Defs are deduplicated by content hash, honor the renderer's
  hash-salted ID strategy, and emit in registration order so document output
  remains deterministic. Hatch still wins precedence when both are set.
- [ ] PDF implementation via shading dictionaries (Type 2 / 3).
- [ ] Skia implementation via `SkShader` types.
- [ ] Golden fixtures: gradient fill bar, radial gradient pie wedge, pattern
  fill polygon, gradient streamline plot.

Current slice landed:

- New `backends/svg/gradients.go` registers `<linearGradient>`,
  `<radialGradient>`, and `<pattern>` defs from `Paint.FillGradient` and
  `Paint.FillPattern`. Unit tests cover linear/radial emission, stop-opacity,
  pattern emission, def deduplication, and hatch-over-gradient precedence.
- New `backends/agg/gradients.go` wires AGG's `FillLinearGradient` /
  `FillRadialGradient` / `FillRadialGradientMultiStop` into `Path()`. Unit
  tests verify left-to-right linear color falloff, center-to-edge radial
  falloff, and that subsequent solid fills are not painted through the active
  gradient span generator.
- `render.GradientFiller` / `render.PatternFiller` capability interfaces are
  now implemented on AGG and SVG; the backend capability comparison report
  reflects native vs unsupported truthfully.

### 2.2 Path Effects Pipeline

- [ ] Path effects model (`PathEffect` interface) covering Matplotlib's
  `Normal`, `Stroke`, `withStroke`, `SimplePatchShadow`, `SimpleLineShadow`,
  `PathPatchEffect`, `TickedStroke`.
- [ ] Backend hook for offscreen capture / replay: AGG uses
  `StartFilter` / `StopFilter`; SVG uses `<filter>` defs; PDF uses
  transparency groups + soft masks.
- [ ] Apply-time pipeline that walks the effects list and composes results
  back into the parent renderer.
- [ ] Golden fixtures: text with drop shadow, line with halo, scatter markers
  with shadow, polygon outline + fill effect stack.

### 2.3 Mixed Raster / Vector Output

- [ ] Artist-level "rasterize" flag honored by every vector backend, gated by
  renderer capability checks.
- [ ] DPI-aware rasterization at save time so dense scatter / image /
  contour plots embed as raster tiles inside PDF / PS / SVG without losing
  surrounding vector text and axes.
- [ ] Golden fixtures verifying the rasterized region honors clip, transform,
  and alpha state.

**Exit criteria:**

- [ ] Pattern fills, gradients, and path effects work uniformly across AGG,
  SVG, PDF, and Skia without backend-name conditionals.
- [ ] `Artist.SetRasterized(true)` produces correct mixed-mode output on
  every vector backend.
- [ ] All effects have committed golden and Matplotlib-reference fixtures.

---

# Phase 3: Mathematical Text and TeX

**Goal:** make MathText and `usetex` first-class across raster and vector
backends, and stabilize `internal/mathtext` for promotion.

**Reference sources:** `third_party/matplotlib/lib/matplotlib/_mathtext.py`,
`mathtext.py`, `texmanager.py`; current `internal/mathtext/`.

### 3.1 MathText Pipeline Completion

- [ ] Finish the shared shaping layer (carried over from prior backend work)
  so AGG text draw, text measurement, text bounds, and text-path output all
  consume the same shaped glyph runs.
- [ ] Complete the MathText grammar coverage gaps versus upstream: stacked
  fractions, accents, big operators, integral limits, matrix environments.
- [ ] Cache stabilization: deterministic cache keys, eviction policy, and
  cross-process safe storage so `internal/mathtext` can ship as its own
  module.
- [ ] MathText draw path through every backend: AGG (raster glyph composite),
  SVG (paths or text-as-text where the font is available), PDF, Skia.
- [ ] Golden fixtures: mathtext_basic, mathtext_fractions, mathtext_integrals,
  mathtext_matrices, mathtext_inline_labels.

### 3.2 `usetex` Support

- [ ] `usetex` import path that shells out to a system `latex` / `dvipng` /
  `dvisvgm` pipeline, behind a build tag / rc switch so the default build
  has no external dependency.
- [ ] DVI parser sufficient to read the geometry of the rasterized result
  back into the renderer's text bounds API.
- [ ] Shared clipping, alpha, and DPI semantics between MathText and `usetex`
  paths so the artist-side API does not branch.
- [ ] Golden fixtures gated by the presence of a TeX installation; skip with
  a clear diagnostic when missing.

### 3.3 MathText Module Promotion

- [ ] Stabilize the public API surface of `internal/mathtext` against the
  needs of the AGG, SVG, PDF, and Skia text drawers.
- [ ] Promote `internal/mathtext` to a top-level module / repo with its own
  versioning, once the grammar coverage and cache contracts are firm.

**Exit criteria:**

- [ ] MathText renders identically across all backends within documented
  tolerances.
- [ ] `usetex` is opt-in, dependency-free by default, and tested when the
  external TeX toolchain is present.
- [ ] `internal/mathtext` is either standalone or has a documented promotion
  date.

---

# Phase 4: Interactive Backends and Event Loop

**Goal:** turn the existing headless canvas / event scaffolding into a
working interactive runtime that supports pan, zoom, picking, and live
updates.

**Reference sources:** `third_party/matplotlib/lib/matplotlib/backend_bases.py`
(NavigationToolbar2, FigureCanvasBase event flow), upstream `backend_qtagg.py`
and `backend_tkagg.py` for desktop reference, `webagg_core.py` for web
reference; current `core/events.go`, `internal/webdemo/`.

### 4.1 Navigation and Hit Testing

- [ ] Pan, zoom-to-rect, and box-zoom interactions wired through the event
  dispatcher and the existing draw-idle queue.
- [ ] Picking / hit testing: `Artist.Contains(MouseEvent)` for every artist
  family, with shared bounding-box and path-contains helpers.
- [ ] Coordinate inspection: hover-driven formatter callbacks, cursor
  rendering hook, and a default `format_coord` implementation.
- [ ] Callback registration matching `mpl_connect` / `mpl_disconnect`
  semantics; covered by event-lifecycle tests.

### 4.2 Desktop Interactive Backend

- [ ] Decision and ADR on the desktop toolkit: Fyne, Ebiten, Gio, or a thin
  SDL2 binding. Decision criteria: pure-Go preference, AGG framebuffer
  embedding, keyboard / mouse event fidelity, and CI availability.
- [ ] Backend implementation that hosts an AGG renderer, drives the event
  dispatcher, and supports the standard NavigationToolbar actions
  (home / pan / zoom / save).
- [ ] Toolbar abstraction generic enough for a future Qt or GTK binding.
- [ ] Embedding example in `examples/embed/desktop/`.

### 4.3 Web Interactive Backend (WebAgg-style)

- [ ] Server-side WebAgg implementation that broadcasts AGG diff regions
  over WebSockets, mirroring upstream's protocol shape.
- [ ] Browser-side JS shim handling event encoding, diff application, and
  cursor rendering.
- [ ] WASM interactive mode for the existing browser demo host so the
  GitHub Pages gallery is clickable.
- [ ] Embedding example in `examples/embed/web/`.

### 4.4 Real-Time Redraw

- [ ] Blit / damage-region optimizations for animated artists, riding on the
  existing AGG `CopyFromBBox` / `RestoreRegion` surface.
- [ ] `draw_idle` scheduling parity: coalesce redraw requests, drop stale
  frames, honor the figure's `stale` propagation.
- [ ] Tests that verify event-driven mutations produce exactly one redraw
  per idle tick, not one per mutation.

**Exit criteria:**

- [ ] At least one desktop and one web interactive backend can drive pan /
  zoom / pick across every plot category committed in earlier phases.
- [ ] Event lifecycle and redraw scheduling match upstream Matplotlib for
  the documented event set.
- [ ] Interactive backends share the same artist / event / renderer surface
  as the headless backends.

---

# Phase 5: Widgets and Selectors

**Goal:** turn the static widget artist surface introduced in Phase 7 into
fully interactive widgets that participate in the event dispatch from
Phase 4.

**Reference sources:** `third_party/matplotlib/lib/matplotlib/widgets.py`.

### 5.1 Interactive Widget Behaviors

- [ ] `Button` click activation with hover, press, and disabled states.
- [ ] `Slider` and `RangeSlider` with click-to-set, drag, keyboard nudging,
  and value formatting.
- [ ] `CheckButtons` and `RadioButtons` with keyboard navigation and
  value-change callbacks.
- [ ] `TextBox` with focus, caret, selection, copy / paste, and submit /
  cancel callbacks.

### 5.2 Selectors

- [ ] `SpanSelector`, `RectangleSelector`, `EllipseSelector`,
  `PolygonSelector`, and `LassoSelector` with mouse and keyboard editing.
- [ ] Modifier-key behaviors (shift / ctrl / alt) matching upstream
  defaults.
- [ ] `Cursor` and `MultiCursor` helpers driven by hover events.

### 5.3 Widget Composition

- [ ] Widget z-order separate from artist z-order so widgets always sit on
  top of plot data.
- [ ] Layout helpers for widget axes that compose with `GridSpec` and
  `constrained_layout`.
- [ ] Widget gallery example covering every widget, mirroring the upstream
  `gallery/widgets/` family.

**Exit criteria:**

- [ ] Every widget responds to mouse and keyboard events through the shared
  event dispatcher.
- [ ] Selectors emit semantic callbacks (with data-coordinate payloads) and
  are usable for ROI selection workflows.
- [ ] Widget examples render correctly in headless mode and remain
  interactive in desktop and web backends.

---

# Phase 6: Animation

**Goal:** add the animation API that depends on the interactive event loop
and the blit-friendly redraw paths from Phase 4.

**Reference sources:** `third_party/matplotlib/lib/matplotlib/animation.py`.

### 6.1 Animation API

- [ ] `FuncAnimation` and `ArtistAnimation` mirroring upstream signatures,
  driven by the figure's draw-idle scheduler.
- [ ] Frame timing / pacing controls (interval, repeat, repeat_delay,
  blit toggle) with deterministic-frame mode for tests.
- [ ] Artist `set_animated(true)` flag honored by the AGG and Skia
  backends via blit regions.

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

- [ ] Complete the AGG MathText raster pipeline once Phase 3.1 lands so
  raster glyph composition shares the same shaping pipeline as text-as-path.
- [ ] Plumb `usetex` output through AGG using the DVI parser from Phase 3.2.
- [ ] Expand AGG parity diagnostics for remaining non-text residuals: dense
  path collections, repeated translucent overlaps, image interpolation modes,
  hatch clipping, and mixed raster / vector fallbacks.
- [ ] Split AGG-native parity fixtures from renderer-neutral fallback
  fixtures so missing native AGG behavior cannot be hidden by fallback
  drawing.

### 7.2 SVG Coverage Expansion

- [ ] Expand the structural golden set to the remaining canonical plot
  families: bar, errorbar, hist, collection, image, clipped polar,
  hatch_bars, text_layout, mathtext.
- [ ] Wire the SVG-specific golden set into the catalog so the structural
  diff harness runs alongside the rasterized golden / reference comparison.

### 7.3 Skia Native Paths

- [ ] Native Skia marker batches, path collections, transformed images,
  quad meshes, and Gouraud triangles wired through `SkCanvas::drawAtlas` and
  `SkVertices`.
- [ ] Skia native hatching via tiled `SkShader`s.
- [ ] GPU mode (`SkSurface::MakeRenderTarget`) behind a separate build tag,
  with deterministic CPU readback for golden tests.
- [ ] Capability reporting split between CPU and GPU configurations so the
  comparison report shows truthful native / fallback / unavailable status
  per mode.
- [ ] Skia vs AGG semantic-fixture comparison; tolerances documented per
  fixture where Skia is not expected to pixel-match.

### 7.4 GoBasic Long Tail

- [ ] GoBasic equivalents for the renderer-neutral path effect pipeline
  introduced in Phase 2 so the fallback backend keeps full semantic coverage.
- [ ] GoBasic smoke coverage for any new plot family introduced in Phases
  1-6 (PDF / interactive / animation paths excluded since GoBasic targets
  static output).

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

# Phase 8: Documentation, Examples Polish, and v1.0 Release

**Goal:** make the project consumable by users who have not been following
the development thread, and tag a stable v1.0.

### 8.1 API Documentation

- [ ] Package-level GoDoc passes for every public package, with a worked
  example per package.
- [ ] Hosted documentation site (pkg.go.dev plus a curated landing page
  under the existing GitHub Pages deployment).
- [ ] Migration guide from upstream Matplotlib: side-by-side Python / Go
  snippets for every plot family covered by the catalog.
- [ ] Backend selection guide: when to use AGG / GoBasic / SVG / PDF /
  Skia, with capability matrix excerpts.

### 8.2 Examples Gallery Polish

- [ ] Review every `Showcase: true` catalog row for caption, description,
  and runnable snippet quality.
- [ ] Add an "anti-gallery" of intentional Matplotlib-divergence cases with
  the reasons documented (where the Go port chose different defaults).
- [ ] Promote the WASM browser gallery to a first-class entry point on the
  project README.

### 8.3 Performance Pass

- [ ] Profiling sweep across the catalog: identify hotspots that exceed the
  100k-point smoothness goal and the sub-second typical-plot goal.
- [ ] Reusable benchmark suite under `benchmarks/` with regression tracking
  in CI.
- [ ] Documented memory-usage targets and a tuning guide for long-running
  applications.

### 8.4 Release Readiness

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
