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

# Phase 1A: PDF Publication Backend

**Goal:** make PDF a deterministic, publication-quality vector backend with
text, image, hatch, alpha, metadata, and resource-reuse behavior close to
Matplotlib.

**Reference sources:** `third_party/matplotlib/lib/matplotlib/backends/backend_pdf.py`,
current `backends/svg/` for serialization patterns.

### 1A.1 PDF Backend Foundation

**Goal:** make PDF a deterministic first-class backend with enough renderer
coverage to save common plots without raster fallbacks.

- [x] Scaffold `backends/pdf/` with `doc.go`, `init.go`, `pdf.go`,
      `pdf_test.go`, and `registry_test.go`.
- [x] Register `backends.PDF`, `backends.PDFExport`, `render.PDFExporter`,
      `render.PDFOptionExporter`, `render.PDFOptionSetter`, and `render.DPIAware`.
- [x] Add deterministic Catalog / Pages / Page / Contents / Info objects,
      xref/trailer writing, compact float formatting, literal-string escaping,
      and PDF name escaping.
- [x] Add the page content stream encoder: graphics state stack, top-left
      display-coordinate transform, path construction, fills, strokes, clips,
      and RGB color spaces.
- [x] Compress content streams with `/FlateDecode`.
- [x] Honor `SOURCE_DATE_EPOCH` for deterministic `/CreationDate`.
- [x] Route `.pdf` through `backends.SavePDF`, registry `SaveFormats`, and
      `backends/all` side imports.
- [x] Add `core.SaveFig(fig, r, "out.pdf")` dispatch through
      `render.PDFExporter`.

### 1A.2 PDF Text, Images, and Reuse

**Goal:** bring PDF output close to Matplotlib semantics for text, raster
images, hatches, alpha, and repeated resources.

- [x] Implement text-as-path output through `render.TextPather` for
      `TextPath`, font-keyed text, rotated text, vertical text, and simple
      `GlyphRun` fallback.
- [x] Implement embedded Type 0 / CIDFontType2 text resources for
      `render.WithPDFFontPolicy(render.PDFFontPolicyEmbed)`.
- [x] Add deterministic embedded TrueType font programs, `/CIDToGIDMap`, `/W`,
      and `/ToUnicode` maps.
- [x] Subset the PDF resource CID map to used glyphs while embedding the
      resolved font program bytes.
- [x] Emit raster image XObjects with `/FlateDecode` RGB streams and grayscale
      `/SMask` images for RGBA alpha.
- [x] Add JPEG `/DCTDecode` passthrough through the optional
      `render.JPEGImage` interface.
- [x] Add PNG predictor `/DecodeParms` for Flate image streams.
- [x] Reuse repeated raster image XObjects when encoded RGB/alpha payloads
      match.
- [x] Add transformed image support through `render.ImageTransformer` and PDF
      `cm` matrices.
- [x] Add native hatch fills through reusable PDF tiling pattern resources.
- [x] Add native marker and path-collection batches through reusable Form
      XObjects.
- [x] Add stroke/fill alpha through reusable PDF `/ExtGState` resources with
      separate `CA` / `ca` values.

### 1A.3 PDF Test and Fixture Hardening

**Goal:** keep PDF output stable enough that regressions are caught without
overfitting tests to object numbers or whitespace.

- [x] Add `internal/pdfcompare` for structural PDF comparison.
- [x] Parse indirect objects, ignore xref offset noise, normalize object token
      whitespace, and decode `/FlateDecode` streams before comparison.
- [x] Cover hatch pattern resources through decoded structural comparison.
- [x] Add golden fixtures for line, bar, scatter, hist, contour, imshow,
      polar, hatch_bars, text_layout, clipped, and image_transformed cases.
- [x] Add registry tests for PDF extension selection and export.

**PDF exit criteria:**

- [x] `core.SaveFig(fig, r, "out.pdf")` draws and exports through
      `render.PDFExporter`.
- [x] `.pdf` is registered in `SaveFormats`, selected by
      `backends.SelectBackendForExtension`, and covered by
      `cmd/example -format pdf` smoke output.
- [x] PDF docs define font, image, hatch, metadata, and deterministic-output
      semantics.

---

# Phase 1B: PostScript / EPS and PGF Backends

**Goal:** add the remaining publication vector formats after PDF, with small
backend-specific hardening tracks for the places PS/EPS and PGF differ from PDF.

**Reference sources:** `third_party/matplotlib/lib/matplotlib/backends/backend_ps.py`,
`backend_pgf.py`, current PDF / SVG backends for shared serialization patterns.

### 1B.1 PostScript / EPS Backend Foundation

**Goal:** support journal-style Level-2 PostScript/EPS output with clear,
documented limitations where PS cannot match PDF semantics.

- [x] Scaffold `backends/ps/` with `doc.go`, `init.go`, `ps.go`, and
      `registry_test.go`.
- [x] Register `backends.PS`, `backends.PSExport`, `render.PSExporter`, and
      `render.DPIAware`.
- [x] Include PS in `VectorOutput` through runtime-checked exporter
      interfaces.
- [x] Add deterministic Adobe headers, bounding boxes, document stream layout,
      and top-left display-coordinate transform.
- [x] Implement graphics state stack, path construction, rectangular/path
      clipping, stroke state, RGB stroke/fill color, background fills, and
      simple Helvetica text output.
- [x] Emit deterministic inline Level-2 `colorimage` payloads for RGBA raster
      images, pre-composited over white for the initial alpha slice.
- [x] Implement transformed images with arbitrary PostScript `concat`
      matrices.
- [x] Implement native hatch fills as clipped deterministic hatch stroke
      lines.
- [x] Implement native marker and path-collection batches as reusable
      PostScript procedures.
- [x] Register `.ps` and `.eps` in `SaveFormats`, side-import from
      `backends/all`, and route through `core.SaveFig` / `backends.SavePS`.
- [x] Add `cmd/example -format ps` smoke coverage.

### 1B.2 PostScript Parity Hardening

**Goal:** close the PS-specific gaps that remain after the first useful backend
slice.

- [x] Decide the PostScript font policy: embedded fonts, Type 3 glyph paths,
      or documented direct-text limitations.
- [x] Match PDF's embedded-font behavior where feasible.
- [x] Define alpha semantics for fills, strokes, images, and hatches in a way
      that is honest about Level-2 PS limitations.
- [x] Add JPEG passthrough or document why PS output always encodes raster
      image data.
- [x] Reuse repeated image payloads where the PostScript representation allows
      it.
- [x] Add PS structural or smoke fixtures for text, hatches, transformed
      images, collections, and alpha-limit behavior.

**PostScript exit criteria:**

- [x] `backends.SelectBackendForExtension("", ".ps" | ".eps", nil)` selects
      the PS backend and registry `SaveViaExtension` writes through
      `SaveFormats`.
- [x] `cmd/example -format ps` writes a non-empty file in smoke tests.
- [x] PS font, alpha, JPEG/image-reuse, and fixture policy is implemented or
      explicitly documented as a limitation.

### 1B.3 PGF Backend Foundation

**Goal:** provide generator-only PGF output for direct inclusion in LaTeX
documents without requiring a TeX engine during ordinary saves.

- [x] Scaffold `backends/pgf/` with `doc.go`, `init.go`, `pgf.go`, and
      `registry_test.go`.
- [x] Register `backends.PGF`, `backends.PGFExport`, `render.PGFExporter`,
      `render.DPIAware`, `render.FontTextDrawer`, and
      `render.FontRotatedTextDrawer`.
- [x] Emit deterministic `pgfpicture` output with paths, rectangular/path
      clips, graphics scopes, and the same top-left display-coordinate
      transform used by the other vector renderers.
- [x] Emit simple direct LaTeX text and rotated text through `\pgftext`.
- [x] Add explicit `\pgfsetfillopacity` and `\pgfsetstrokeopacity` commands
      for filled, stroked, hatched, and text output.
- [x] Implement native PGF hatch fills as clipped deterministic hatch stroke
      lines.
- [x] Register `.pgf` in `SaveFormats`, side-import from `backends/all`, and
      route through `core.SaveFig` / `backends.SavePGF`.
- [x] Add `cmd/example -format pgf` smoke coverage that writes a non-empty PGF
      file.

### 1B.4 PGF Verification and Parity Hardening

**Goal:** decide how far PGF should go as a pure generator backend, then fill
the chosen gaps in small pieces.

- [x] Decide whether CI/dev verification invokes `lualatex`, uses optional
      local TeX verification only, or keeps PGF as pure generator smoke output.
- [ ] Add PGF-specific option and metadata routing after the shared save-option
      surface lands.
- [x] Implement PGF raster image output.
- [x] Implement PGF transformed raster images or route them through the shared
      mixed-mode fallback.
- [x] Implement reusable PGF marker batches.
- [x] Implement reusable PGF path-collection batches.
- [ ] Tighten PGF TeX/font metrics parity against upstream Matplotlib.
- [x] Add PGF fixtures for text, alpha scopes, hatches, image output, and
      collection batching.

**PGF exit criteria:**

- [x] `backends.SelectBackendForExtension("", ".pgf", nil)` selects a
      registered PGF backend.
- [x] `cmd/example -format pgf` writes a non-empty PGF file.
- [x] PGF verification policy is documented and reflected in CI or optional
      developer checks.
- [ ] PGF image, batching, option, and text-metrics limitations are either
      implemented or explicitly documented.

---

# Phase 1C: Shared Vector Save Pipeline

**Goal:** keep SVG, PDF, PS/EPS, and PGF on one extension-driven save path,
then add shared option routing and mixed raster/vector fallback without
backend-name conditionals.

**Reference sources:** `backends/`, `canvas/`, `core/`, `pyplot/`, `cmd/example/`,
and Matplotlib's mixed-mode/vector save behavior in `third_party/matplotlib`.

### 1C.1 Shared Save Dispatch and Capability Reporting

**Goal:** keep all public save routes on the same extension-driven registry
path, with capability reporting that makes missing formats obvious.

- [x] Remove hard-coded format fallbacks from `backends.SaveViaExtension`; it
      now requires an explicit `SaveFormats` handler.
- [x] Register `.pdf`, `.ps`, `.eps`, and `.pgf` in `SaveFormats`.
- [x] Select PDF / PS / PGF through `backends.SelectBackendForExtension`.
- [x] Side-import PDF / PS / PGF from `backends/all`.
- [x] Route `pyplot.SaveFig`, headless canvas / manager save, `cmd/example`,
      and CLI save helpers through `SelectBackendForExtension` and
      `SaveFormats`.
- [x] Keep `core.SaveFig` as the renderer-interface helper to avoid the
      existing `backends -> canvas -> core` cycle.
- [x] Define and test fallback behavior when `MATPLOTLIB_BACKEND` is pinned to
      a backend that cannot write the requested extension.
- [x] Add `cmd/example -format png|svg|pdf|ps|eps|pgf` support.
- [x] Add smoke coverage for PNG / SVG / PDF / PS / PGF.
- [x] Add `PGFExport` to the backend capability matrix.
- [x] Expand `BackendComparisonReport` with `PDFExport`, `PSExport`,
      `PGFExport`, and a `SaveFormats` column.
- [x] Document per-format status for fonts, hatches, alpha, raster images,
      transformed images, marker/path-collection batching, metadata, and
      deterministic output.

### 1C.2 Shared Save Options

**Goal:** let format-specific options flow through public save APIs without
backend-name conditionals.

- [x] Inventory existing option types: `render.SVGOptions`,
      `render.PDFOptions`, PS metadata needs, and future PGF options.
- [x] Choose one shared API shape: backend-neutral option bag,
      typed per-format options, or both.
- [x] Route PDF metadata, creation date, and font policy through `pyplot`,
      canvas manager saves, and `cmd/example`.
- [x] Preserve existing SVG option routing while moving it onto the shared
      option surface.
- [x] Add PS option placeholders only where the backend supports meaningful
      behavior.
- [x] Add PGF option placeholders for TeX preamble, metadata/comment policy,
      and future verification mode.
- [x] Add tests proving unsupported options produce clear errors instead of
      being silently ignored.

**Save-pipeline exit criteria:**

- [x] `pyplot.SaveFig`, canvas / manager save, `cmd/example`, and CLI helpers
      choose backends through `SelectBackendForExtension` and write through
      `SaveFormats`.
- [x] Backend comparison / capability output makes both export capabilities
      and registered save extensions obvious for PNG, SVG, PDF, PS/EPS, and
      PGF.
- [x] Format-specific save options can be passed through the shared save
      pipeline for SVG, PDF, PS, and PGF without backend-name conditionals.

### 1C.3 Mixed Raster / Vector Fallback

**Goal:** give vector backends a shared way to embed rasterized regions when an
artist or effect cannot be represented natively.

- [x] Define renderer capability checks for mixed raster/vector embedding.
- [x] Add an artist-level rasterize flag that vector backends honor.
- [x] Add DPI-aware rasterization at save time for dense scatter, image,
      contour, and unsupported-effect regions.
- [x] Embed raster tiles inside PDF without losing surrounding vector text and
      axes.
- [x] Embed raster tiles inside PS/EPS with documented alpha behavior.
- [x] Embed raster tiles inside SVG using existing image support.
- [x] Embed raster tiles inside PGF or document the fallback through external
      image files.
- [x] Add fixtures verifying clip, transform, alpha, and surrounding vector
      content.

**Mixed-output exit criteria:**

- [x] Rasterized fallback is driven by renderer capability checks, not
      backend-name conditionals.
- [x] `Artist.SetRasterized(true)` produces correct mixed-mode output on PDF,
      PS/EPS, SVG, and PGF where supported.
- [x] Unsupported vector effects have a deterministic raster fallback or a
      documented error.

Vector backend semantics matrix:

| Capability / semantic | SVG | PDF | PS / EPS | PGF |
| --- | --- | --- | --- | --- |
| Registry extension(s) | `.svg` | `.pdf` | `.ps`, `.eps` | `.pgf` |
| Deterministic output | native IDs / metadata policy | deterministic xref, metadata, `SOURCE_DATE_EPOCH` | deterministic document stream | deterministic `pgfpicture` stream |
| Text policy | text-as-text plus path policy | embedded Type 0 / CIDFontType2 or text-as-path | basic direct text, no embedded fonts yet | LaTeX text via `\pgftext`, approximate layout metrics |
| Font subsetting / embedding | n/a for text-as-text, paths available | implemented | missing | delegated to LaTeX, no subsetting |
| Hatch fills | native SVG patterns | native PDF tiling patterns | native clipped hatch strokes | native clipped hatch strokes |
| Stroke / fill alpha | native SVG opacity | PDF ExtGState | limited by PostScript semantics | PGF fill/stroke opacity commands |
| Raster images | embedded image data | Image XObjects with alpha masks | inline colorimage, alpha precomposited | self-contained PGF pixel rectangles |
| Transformed images | implemented | implemented | implemented | implemented through PGF transform scopes |
| Marker / path collections | reusable native batches | Form XObjects | reusable procedures | reusable PGF macros |
| Metadata options | `render.SVGOptions` | `render.PDFOptions` but not fully routed through shared save APIs | missing shared option surface | missing shared option surface |

Remaining shared-semantics work is concentrated in PostScript parity hardening
(Phase 1B.2), PGF hardening (Phase 1B.4), shared option routing (Phase 1C.2),
and mixed raster/vector fallback (Phase 1C.3).

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
      three-stop radial uses the multi-stop variant. `SupportsGradientFill` and
      `SupportsPatternFill` advertise native support.
- [x] SVG implementation via `<linearGradient>` / `<radialGradient>` /
      `<pattern>` defs. Defs are deduplicated by content hash, honor the renderer's
      hash-salted ID strategy, and emit in registration order so document output
      remains deterministic. Hatch still wins precedence when both are set.
- [x] PDF implementation via shading dictionaries (Type 2 / 3).
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
- AGG now also advertises `render.PatternFiller` / `backends.PatternFill` and
  replays tiled `Paint.FillPattern` cells through AGG path drawing clipped to
  the destination path. Unit tests cover tile repetition, tile transforms, and
  hatch-over-pattern precedence.
- `render.GradientFiller` / `render.PatternFiller` capability interfaces are
  now implemented on AGG and SVG; the backend capability comparison report
  reflects native vs unsupported truthfully.
- PDF now advertises `render.GradientFiller` / `backends.GradientFill` and
  emits native axial/radial shading resources for linear and radial gradient
  fills, including clipping the gradient to the path and preserving an
  overlaid stroke pass.
- PDF now also advertises `render.PatternFiller` / `backends.PatternFill` and
  maps renderer-neutral `Paint.FillPattern` values to colored tiling pattern
  resources, while keeping hatch-over-pattern and gradient-over-pattern
  precedence aligned with SVG.

### 2.2 Path Effects Pipeline

- [x] Path effects model (`PathEffect` value type) covering Matplotlib's
      `Normal`, `Stroke`, `withStroke`, `SimplePatchShadow`, `SimpleLineShadow`,
      `PathPatchEffect`, `TickedStroke`.
- [ ] Backend hook for offscreen capture / replay: AGG uses
      `StartFilter` / `StopFilter`; SVG uses `<filter>` defs; PDF uses
      transparency groups + soft masks.
- [x] Apply-time pipeline that walks the effects list and composes results
      back into the parent renderer.
- [ ] Golden fixtures: text with drop shadow, line with halo, scatter markers
      with shadow, polygon outline + fill effect stack.

Current slice landed:

- `render.PathEffect` now covers normal replay, alternate stroke / with-stroke
  stacks, simple line and patch shadows, path-patch repaint passes, and ticked
  strokes. Convenience constructors mirror the Matplotlib path-effects names.
- `render.DrawPathWithEffects` provides a renderer-neutral replay pipeline:
  each pass clears nested effects, applies offsets / alternate paint, generates
  tick segments when requested, and replays into the parent renderer.
- `PathEffectFilter` now routes through the shared `render.FilterRenderer`
  hook when a backend supports offscreen capture. AGG uses its existing
  `StartFilter` / `StopFilter` surface stack and supports deterministic
  identity / blur filter passes before compositing back to the parent surface.
- SVG now implements a native path-effect filter hook for blurred filter
  passes, registers deterministic `<filter>` / `<feGaussianBlur>` defs, and
  wraps only the affected replay pass while leaving the normal pass as vector
  path output.
- PDF now implements the native path-effect filter hook for identity / no-op
  filter passes by capturing the replay into deterministic transparency-group
  Form XObjects and invoking those groups from the page content stream.
- AGG, GoBasic, SVG, PDF, PS, PGF, and Skia now implement
  `render.PathEffectDrawer`; backend capability declarations advertise
  `PathEffects` through the runtime interface check.
- Core line, patch, text, scatter, and collection artists now carry
  `PathEffects` through to path paints. Collection batch optimizations fall
  back to per-path drawing when effects are present so pass ordering remains
  correct.

Remaining path-effects work:

- Native filter/offscreen variants for blurred shadows and vector soft masks.
  AGG has the first `FilterRenderer` replay path; SVG has native blur filter
  defs; PDF has transparency-group replay for identity filters, but blurred
  soft masks still need native vector implementations.
- Golden and Matplotlib-reference fixtures for text shadows, line halos,
  scatter marker shadows, and polygon effect stacks.

### 2.3 Mixed Raster / Vector Output

**Goal:** match Matplotlib's mixed-mode output model: selected artists can be
drawn into a DPI-aware raster tile while surrounding labels, axes, and simple
geometry remain native vector output.

#### 2.3.1 Renderer contract and capability gating

- [x] Add renderer-neutral rasterization policy plumbing:
      `render.Rasterization`, `RasterizationMode`, `Paint.Rasterization`, and
      `GraphicsContext.WithRasterization`.
- [x] Add `render.RasterizationController` with `StartRasterized` /
      `StopRasterized` so artists can request mixed output without backend-name
      conditionals.
- [x] Advertise mixed-mode support through the backend capability matrix as
      `backends.MixedRasterVector`, gated by the runtime controller interface.

#### 2.3.2 Artist opt-in and auto-rasterization

- [x] Add reusable `core.ArtistRasterization` with `SetRasterized`,
      `SetRasterization`, and `Rasterization` methods.
- [x] Honor explicit artist-level rasterization only when the active renderer
      supports `render.RasterizationController`; otherwise draw the artist
      normally.
- [x] Expose rasterization controls on common artist families: lines, scatter,
      images, contours, collections, bars, fills, patches / rectangles, text,
      and annotations.
- [x] Auto-rasterize dense scatter, collection, and contour output at figure DPI,
      with `RasterizeNever` preserving fully vector output for opt-out cases.
- [x] Preserve vector-native path-effect filter output when a backend can draw
      the effect natively, falling back to mixed raster output only when needed.

#### 2.3.3 DPI-aware offscreen replay

- [x] Add shared `backends/internal/mixedraster` session setup for transparent
      offscreen raster groups.
- [x] Allocate mixed-raster surfaces at the requested rasterization DPI while
      preserving the original vector placement rectangle.
- [x] Scale replayed paths, clips, images, and text into the high-DPI tile before
      embedding it back into the vector output.
- [x] Scale and replay affine-transformed images into the high-DPI tile through
      the shared GoBasic raster surface.

#### 2.3.4 Vector backend embedding

- [x] SVG embeds mixed-raster groups as `<image>` elements while keeping
      unaffected text and axes as vector content.
- [x] PDF embeds mixed-raster groups as image XObjects with alpha mask support
      where needed.
- [x] PostScript / EPS embeds mixed-raster groups through deterministic
      `colorimage` output.
- [x] PGF embeds mixed-raster groups as self-contained pixel rectangles.

#### 2.3.5 Clip, transform, and alpha correctness

- [x] Replay active rectangular clips into mixed-raster surfaces before embedding
      the resulting tile.
- [x] Replay active path clips, including transformed SVG clip paths, into the
      offscreen surface.
- [x] Preserve transparent pixels outside active path clips; PDF coverage asserts
      pixels outside the clip remain transparent in the embedded image.

#### 2.3.6 Fixture and regression coverage

- [x] Add core tests for explicit rasterization bracketing, DPI propagation,
      common artist API coverage, auto-rasterization thresholds, and
      path-effect fallback behavior.
- [x] Add shared mixed-raster helper tests for high-DPI surface sizing and path
      scaling.
- [x] Add backend tests showing rasterized artists embed pixels/images while
      surrounding vector content remains vector in SVG, PDF, PS, and PGF.
- [x] Add catalog golden / Matplotlib-reference fixtures that exercise the full
      save pipeline for clip, transform, and alpha state across mixed
      raster/vector output. (`mixed_raster_vector` covers a DPI-rasterized
      translucent scatter cloud inside a polar path clip, with vector axes,
      labels, legend, and line output preserved in the SVG / PDF structural
      goldens.)

No remaining Phase 2.3 work is currently known.

### 2.4 Native Pattern / Gradient Backend Parity

**Goal:** make the renderer-neutral pattern and gradient contracts genuinely
uniform across the primary native targets: AGG, SVG, PDF, and Skia.

- [x] Add AGG native or renderer-neutral tiled `PatternFill` support so
      `backends/agg` advertises `SupportsPatternFill() == true` for the same
      `render.PatternFill` values accepted by SVG and PDF.
- [x] Pin AGG Phase 2 capabilities in registry tests: `PatternFill`,
      `GradientFill`, `PathEffects`, and `OffscreenFilter` all report native
      support.
- [x] Start Skia parity-viewer workflow plumbing: the `skia` build tag exposes
      Skia as a selectable web-demo export backend, `cmd/webdemoexport` accepts
      `--backend`, and `web-parity-update-skia` / `web-parity-viewer-skia`
      compare Skia-tagged PNG artifacts against the Matplotlib web-demo
      baselines.
- [x] Add Skia `PatternFiller` / `GradientFiller` implementations behind the
      Skia CPU surface bridge, including linear gradients, radial gradients,
      transformed fills, stop opacity, and tiled pattern fills.
- [ ] Swap the CPU bridge shader implementation to external Skia `SkShader`
      primitives once the C ABI wrapper lands.
- [x] Add Skia unit tests matching the existing AGG/SVG/PDF gradient and
      pattern coverage: linear falloff, radial falloff, transformed fill
      geometry, pattern tile repetition, hatch-over-pattern precedence, and
      solid-fill reset after gradient/pattern draws.
- [x] Add backend capability matrix tests that fail if AGG, SVG, PDF, or Skia
      regress from native pattern/gradient/path-effect support once the work
      above lands.

### 2.5 Native Path-Effect Filter Parity

**Goal:** make filtered path effects route through capability interfaces and
produce equivalent native output where the backend can support it.

- [x] Keep AGG as the raster reference for filtered path effects through
      `render.FilterRenderer`, with coverage for offscreen capture/replay and
      blurred path-effect compositing.
- [ ] Complete native filtered path-effect behavior across vector backends:
      AGG offscreen blur remains the raster reference, SVG keeps `<filter>`
      output, PDF adds blurred transparency-group / soft-mask output, and Skia
      renders the same filter passes without core knowing the backend name.
- [ ] Define the intentional fallback semantics for PS, PGF, and GoBasic in
      backend docs and capability declarations so the "uniform" contract is
      explicit: AGG/SVG/PDF/Skia are native targets; fallback backends either
      replay renderer-neutral effects or report unsupported truthfully.

### 2.6 Phase 2 Routing and Regression Audit

**Goal:** lock in the no-backend-name-conditionals requirement and keep it from
regressing after native backend work lands.

- [ ] Add a cross-backend semantic test that draws the same pattern, gradient,
      and path-effect scene through AGG, SVG, PDF, and Skia and asserts only
      capability interfaces (`render.PatternFiller`, `render.GradientFiller`,
      `render.PathEffectDrawer`, `render.PathEffectFilterDrawer` /
      `render.FilterRenderer`) are used for routing.
- [ ] Audit Phase 2 routing code for backend-name conditionals in `core/`,
      `render/`, and shared backend helpers; replace any remaining conditionals
      with capability interfaces or document why they are save-format dispatch
      rather than effect rendering logic.

**Exit criteria:**

- [ ] Pattern fills, gradients, and path effects work uniformly across AGG,
      SVG, PDF, and Skia without backend-name conditionals.
- [x] `Artist.SetRasterized(true)` produces correct mixed-mode output on
      every vector backend.
- [ ] All effects have committed golden and Matplotlib-reference fixtures.

---

# Phase 3: Mathematical Text and TeX

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
- [ ] Golden fixtures gated by the presence of a TeX installation; skip with
      a clear diagnostic when missing.
      Harness exists, but committed TeX-generated PNG fixtures still need to
      be produced on a host with `latex` + `dvipng`.

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
      versioning, once the grammar coverage and cache contracts are firm.

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

**Goal:** turn the existing headless canvas / event scaffolding into a
working interactive runtime that supports pan, zoom, picking, and live
updates.

**Reference sources:** `third_party/matplotlib/lib/matplotlib/backend_bases.py`
(NavigationToolbar2, FigureCanvasBase event flow), upstream `backend_qtagg.py`
and `backend_tkagg.py` for desktop reference, `webagg_core.py` for web
reference; current `core/events.go`, `internal/webdemo/`.

### 4.1 Navigation and Hit Testing

- [x] Pan, zoom-to-rect, and box-zoom interactions wired through the event
      dispatcher and the existing draw-idle queue.
      `canvas.Navigation` attaches to a `Dispatcher` and mutates axes view
      limits through `PanAxesPixels` / `ZoomAxesToRect` /
      `ZoomAxesByFactor`; mouse-press / move / release events drive pan and
      rubber-band zoom, and scroll events trigger anchored zoom-in/out.
- [x] Picking / hit testing: `Artist.Contains(MouseEvent)` for every artist
      family, with shared bounding-box and path-contains helpers.
      `core.ArtistPicker` / `core.PickInfo` define the contract; Line2D,
      Rectangle, Circle, Ellipse, Polygon, PathPatch, FancyBboxPatch,
      FancyArrow, Image2D, and Text implement `Contains`. Shared
      `distancePointToSegment` / `containsPath` helpers back the patch
      family. `canvas.Pick` and `canvas.EmitPick` walk the figure tree and
      emit `PickEvent`s.
- [x] Coordinate inspection: hover-driven formatter callbacks, cursor
      rendering hook, and a default `format_coord` implementation.
      `Axes.FormatCoord` returns the data-space `x=…, y=…` string for a
      figure-pixel cursor; `Axes.SetFormatCoord(CoordFormatter)` installs
      custom formatters per axes.
- [x] Callback registration matching `mpl_connect` / `mpl_disconnect`
      semantics; covered by event-lifecycle tests.
      `canvas.Connect` / `canvas.Disconnect` wrap `FigureCanvas` with the
      Matplotlib-style API; existing `Dispatcher` already supports the
      multiplexing semantics. Lifecycle covered by `canvas/connect_test.go`.

### 4.2 Desktop Interactive Backend

- [x] Decision and ADR on the desktop toolkit: Fyne, Ebiten, Gio, or a thin
      SDL2 binding. Decision criteria: pure-Go preference, AGG framebuffer
      embedding, keyboard / mouse event fidelity, and CI availability.
      Decided **Gio** in `docs/adr/0001-desktop-interactive-backend.md`;
      the only option that satisfies all five criteria simultaneously
      (pure-Go end-to-end, headless test driver, direct `*image.RGBA`
      blit via `paint.NewImageOp`, multi-window embeddability, CI
      without system packages).
- [x] Backend implementation that hosts an AGG renderer, drives the event
      dispatcher, and supports the standard NavigationToolbar actions
      (home / pan / zoom / save).
      `backends/desktop` defines the toolkit-agnostic `Backend` contract
      and `Options` / `Constructor` / `Register` / `New` machinery.
      `backends/desktop/gio` is the Gio (`gioui.org` v0.10.0)
      implementation: one `app.Window` per figure, an event-pump that
      maps `pointer.Event` / `key.Event` / `app.FrameEvent` /
      `app.DestroyEvent` onto the canvas dispatcher, and an AGG bitmap
      blit via `paint.NewImageOp` each frame. `desktop.New` returns a
      ready-to-Run backend; the AGG renderer is supplied via
      `Options.Renderer`. Covered by `backends/desktop/gio/gio_test.go`
      (7 tests, cgo-free).
- [x] Toolbar abstraction generic enough for a future Qt or GTK binding.
      `canvas.NavigationToolbar` interface plus `canvas.ToolbarController`
      reference implementation in `canvas/toolbar.go`; pan/zoom toggles
      drive the existing `canvas.Navigation` helper, other actions go
      through registered `ToolbarHandler`s. Covered by
      `canvas/toolbar_test.go`.
- [x] Embedding example in `examples/embed/desktop/`.
      `examples/embed/desktop/main.go` opens the `basic_line` figure in
      a Gio window, wires `p` / `z` / `r` / `s` / `q` keyboard
      shortcuts to the toolbar actions, and streams hover coordinates
      into the toolbar status line via `Axes.FormatCoord`.
      `SaveHandler` writes `out.png` through `core.SavePNG`.

### 4.3 Web Interactive Backend (WebAgg-style)

- [x] Server-side WebAgg implementation that broadcasts AGG diff regions
      over WebSockets, mirroring upstream's protocol shape.
      `backends/webagg.Manager` wraps a `*core.Figure` with a per-session
      PNG-diff buffer, image-mode tracking (`full` / `diff`), a
      `canvas.Navigation` + `canvas.ToolbarController`, and a hub that
      fans frames out to every connected client. The wire protocol —
      message names, payload shapes, and the image-mode handshake — is
      a 1:1 port of `third_party/.../backend_webagg_core.py`. JSON text
      frames carry events both directions; binary frames carry PNGs
      server→client. Transport is `golang.org/x/net/websocket` (no new
      deps). Covered by `backends/webagg/webagg_test.go` and
      `server_test.go` (15 tests, including a real httptest +
      WebSocket round-trip).
- [x] Browser-side JS shim handling event encoding, diff application, and
      cursor rendering.
      `backends/webagg/static/mpl.js` opens the WebSocket, decodes
      `image_mode` / `resize` / `figure_label` / `cursor` /
      `rubberband` / `history_buttons` / `navigate_mode` /
      `message` events, blits binary PNG frames over a `<canvas>`
      (full or diff via the source-over compositor), and encodes
      mouse, scroll, key, dblclick, and toolbar events upstream-style.
      `static/index.html` ships a minimal toolbar; hosts can swap in
      their own page via `webagg.ServerOptions.Assets`.
- [x] WASM interactive mode for the existing browser demo host so the
      GitHub Pages gallery is clickable.
      `canvas/wasm.Manager` now attaches a `canvas.Navigation` to its
      `Dispatcher` (pan, scroll-zoom, rubber-band zoom with an
      on-canvas dashed overlay drawn after each `putImageData`) and a
      `canvas.ToolbarController` wired to a home-view snapshot. The
      WASM API exports `setNavigationMode` / `navigationMode` /
      `triggerToolbar`; `web/index.html` ships Home / Pan / Zoom
      toolbar buttons above the canvas and `web/main.js` mirrors the
      active mode via `aria-pressed`. PNG export still goes through
      the existing browser-side `canvas.toBlob` download path.
- [x] Embedding example in `examples/embed/web/`.
      `examples/embed/web/main.go` mounts a `webagg.Server` at `/`,
      serves the `basic_line` figure, wires `Save` to a server-side
      PNG writer, and forwards hover events to the toolbar status
      line via `Axes.FormatCoord`.

### 4.4 Real-Time Redraw

- [x] Blit / damage-region optimizations for animated artists, riding on the
      existing AGG `CopyFromBBox` / `RestoreRegion` surface.
      WebAgg managers now expose the shared `canvas.BlitCanvas`
      surface (`CopyFromBBox` / `RestoreRegion` / `Blit`) and can
      broadcast a damaged renderer buffer without a full figure draw.
- [x] `draw_idle` scheduling parity: coalesce redraw requests, drop stale
      frames, honor the figure's `stale` propagation.
      `webagg.Manager.DrawIdle` schedules work through a backend event
      loop, coalesces repeated requests before the idle callback runs,
      and wires stale artist callbacks into idle redraws.
- [x] Tests that verify event-driven mutations produce exactly one redraw
      per idle tick, not one per mutation.
      `backends/webagg` covers coalesced draw-idle bursts, stale artist
      redraw scheduling, and blit broadcasts that avoid a full
      `DrawFigure` traversal.

**Exit criteria:**

- [x] At least one desktop and one web interactive backend can drive pan /
      zoom / pick across every plot category committed in earlier phases.
      Pan and zoom are wired for Gio, WebAgg, and WASM, and Gio/WebAgg
      now emit `pick_event` on mouse press when an artist is hit.
      `internal/examplecatalog.InteractiveCoverageMatrix` lists every
      catalog topic with a figure-backed representative, and WebAgg/Gio
      matrix tests drive draw, pan drag, scroll zoom, and pick-click
      input for every row.
- [x] Event lifecycle and redraw scheduling match upstream Matplotlib for
      the documented event set.
      Draw-idle coalescing and WebAgg stale redraw scheduling are in
      place. Figure and axes enter / leave, mouse press-before-pick,
      release-without-pick, WebAgg double-click metadata, scroll steps,
      key normalization, and modifier payloads are covered for the
      touched WebAgg/Gio paths. Headless, Gio, WebAgg, and WASM now use
      the shared normalization and lifecycle surface where applicable;
      focused tests cover draw, draw_idle, resize, close, stale redraw,
      and renderer-error propagation for the concrete backends.
- [x] Interactive backends share the same artist / event / renderer surface
      as the headless backends.
      The common `FigureCanvas`, `DrawIdleCanvas`, `BlitCanvas`,
      `Dispatcher`, `Navigation`, picker, and toolbar pieces are covered
      by runtime assertions, WebAgg/Gio matrix tests, WASM compile
      verification, embedder docs, and the switchable embed example.

### 4.5 Interactive Coverage Matrix

- [x] Add a catalog-driven interactive smoke harness that drives at least
      one desktop backend (Gio) and one web backend (WebAgg) across every
      plot category committed in Phases 1–3.
      The harness should live beside the existing backend tests, reuse
      `internal/examplecatalog.Case`, synthesize pan drag, scroll zoom,
      box zoom, and pick clicks, and assert semantic outcomes rather than
      exact pixels: draw events fire, axis limits change when expected,
      pick events fire for pickable cases, and the renderer does not
      panic.
      `backends/webagg.TestInteractiveCoverageMatrixWebAgg` and
      `backends/desktop/gio.TestInteractiveCoverageMatrixGio` drive each
      `InteractiveCoverageMatrix` representative through draw, pan drag,
      scroll zoom, and pick-click input.
- [x] Extend `internal/examplecatalog.Case` with optional interactive
      metadata: `PickPointData`, `PickPointPixel`, `Pickable` /
      `NoPickReason`, and per-case pan/zoom skip flags for non-Cartesian
      or intentionally static figures.
      Existing catalog rows should default to the conservative path:
      pan/zoom smoke runs where axes support pixel inversion; pick checks
      run only when metadata identifies a pickable artist.
      The fields are present on `Case`; the current matrix defaults to
      conservative smoke checks and reserves exact pick assertions for
      rows that later declare pick points.
- [x] Add WebAgg protocol-level integration coverage for pan, zoom, and
      pick using real `httptest` WebSocket clients, not only direct
      `HandleClientMessage` calls.
      `backends/webagg.TestWebSocketDrivesPanScrollAndPick` sends
      toolbar, mouse, scroll, and pick-click JSON over a real
      `httptest` WebSocket connection.
- [x] Add a cgo-free Gio synthetic-input suite for pan, zoom, and pick
      over representative catalog categories, plus a documented manual
      command for visual Gio smoke checks.
      Gio's matrix test uses synthetic `pointer.Event` input over every
      catalog topic representative. `docs/interactive-backends.md`
      documents switchable Gio/WebAgg/headless smoke commands.

**Exit criteria:**

- [x] The Phase 4 first exit criterion is backed by automated coverage
      listing every catalog category and whether it passed, skipped with
      reason, or is explicitly unsupported.
      `backends.TestInteractiveMatrixCoversEveryCatalogTopic` and
      `TestInteractiveMatrixHasFigureFactoryForEveryCatalogTopic` assert
      coverage for every catalog topic.

### 4.6 Event Lifecycle Parity Closure

- [x] Add explicit figure enter / leave event types to `canvas.EventType`
      and map WebAgg `figure_enter` / `figure_leave` plus Gio pointer
      enter/leave into those events instead of collapsing them into
      mouse move.
      `canvas.EventFigureEnter` / `EventFigureLeave` are covered by
      `canvas/architecture_test.go`; WebAgg and Gio dispatch tests
      assert backend mapping and payload positions.
- [x] Add axes enter / leave semantics on top of figure mouse motion.
      `canvas.AxesHoverTracker` keeps per-canvas hover state and
      synthesizes `canvas.EventAxesEnter` / `EventAxesLeave` when the
      resolved axes under the cursor changes. WebAgg and Gio dispatch
      tests cover same-axes motion suppression, axes switching, and
      leaving the figure.
- [x] Add event-ordering tests against the documented Matplotlib flow:
      mouse press emits the button event then pick event; release emits
      release only; double-click preserves click count / double-click
      metadata; scroll uses upstream step semantics; key events preserve
      normalized key plus modifier bitfields.
      WebAgg and Gio assert mouse-press before pick-event ordering and
      release / scroll / key payload behavior. WebAgg `dblclick` now
      dispatches `canvas.EventMousePress` with `DoubleClick=true`.
      Gio's pointer event type does not expose double-click metadata.
- [x] Centralize backend event normalization helpers for mouse buttons,
      modifiers, scroll deltas, key strings, and double-click metadata so
      Gio, WebAgg, WASM, and headless simulations report the same
      canvas-level event payloads.
      Shared helpers now live in `canvas/input.go` for modifier sets,
      modifier names, browser mouse-button indices, and key
      normalization. WebAgg and Gio use the shared modifier/key helpers
      on the exercised paths; WASM now uses the same helpers and emits
      pick / enter / leave through the shared paths.
- [x] Add lifecycle tests for draw, draw_idle, resize, close, stale artist
      redraw, and error propagation across headless, Gio, WebAgg, and
      WASM where applicable.
      Headless, Gio, and WebAgg have focused lifecycle and error tests;
      WASM compiles with the shared lifecycle path under `GOOS=js
      GOARCH=wasm`.

**Exit criteria:**

- [x] The Phase 4 second exit criterion has a parity table mapping each
      documented Matplotlib event to the corresponding canvas event,
      backend mappings, and tests.
      `docs/interactive-backends.md` contains the event parity table and
      links the common payload contract to backend behavior.

### 4.7 Shared Interactive Surface Hardening

- [x] Add compile-time and runtime assertions that every interactive
      backend exposes the expected common surface:
      `FigureCanvas`, `DrawIdleCanvas`, `Dispatcher` event flow,
      navigation, toolbar, and save hooks; `BlitCanvas` where the
      renderer supports buffer regions.
      Headless compile-time interface assertions now exist alongside
      the existing WebAgg and Gio compile-time checks; runtime assertions
      cover headless, WebAgg, and Gio. WASM is covered by compile
      verification.
- [x] Decide whether desktop Gio should expose `canvas.BlitCanvas` by
      reusing its retained renderer/image buffer, or document why full
      frame blits remain the desktop path until animation work in Phase 6.
      Gio remains a full-frame redraw backend for now; the rationale is
      documented in `docs/interactive-backends.md`.
- [x] Align WebAgg, Gio, and WASM hover status, cursor, rubber-band,
      toolbar history, and navigation-mode announcements through shared
      canvas-level APIs instead of backend-local conventions.
      `ToolbarController` now exposes shared mode/message announcement
      callbacks; WebAgg broadcasts `navigate_mode` and `message` through
      those callbacks, Gio keeps the same toolbar state locally, and
      WASM shares the normalized hover/pick lifecycle.
- [x] Add docs for embedders describing the common interactive contracts:
      event registration, pick handling, draw idle, blit regions,
      toolbar actions, save handlers, and backend capability detection.
      `docs/interactive-backends.md` documents the common canvas
      surface, optional `DrawIdleCanvas` / `BlitCanvas` capabilities,
      event payload mapping, picker behavior, backend notes, and Gio's
      full-frame redraw decision.

**Exit criteria:**

- [x] The Phase 4 third exit criterion is backed by shared-interface
      assertions, embedder docs, and at least one example that can switch
      between headless, Gio, and WebAgg without changing artist/event
      code.
      Runtime assertions live in `backends.TestInteractiveBackendsExposeCommonSurface`,
      docs live in `docs/interactive-backends.md`, and
      `examples/embed/switchable` uses one figure/event setup across the
      three backend choices.

---

# Phase 5: Widgets and Selectors

**Goal:** turn the static widget artist surface introduced in Phase 7 into
fully interactive widgets that participate in the event dispatch from
Phase 4.

**Reference sources:** `third_party/matplotlib/lib/matplotlib/widgets.py`.

### 5.1 Interactive Widget Behaviors

- [x] `Button` click activation with hover, press, and disabled states.
- [x] `Slider` and `RangeSlider` with click-to-set, drag, keyboard nudging,
      and value formatting.
- [x] `CheckButtons` and `RadioButtons` with keyboard navigation and
      value-change callbacks.
- [x] `TextBox` with focus, caret, selection, copy / paste, and submit /
      cancel callbacks.

### 5.2 Selectors

- [x] `SpanSelector`, `RectangleSelector`, `EllipseSelector`,
      `PolygonSelector`, and `LassoSelector` with mouse and keyboard editing.
- [x] Modifier-key behaviors (shift / ctrl / alt) matching upstream
      defaults.
- [x] `Cursor` and `MultiCursor` helpers driven by hover events.

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
- [ ] Skia native hatching via tiled `SkShader`s.
  - [x] CPU Skia consumes hatch metadata during path rendering and advertises
        `NativeHatcher`; tiled external `SkShader` hatches remain open.
- [ ] GPU mode (`SkSurface::MakeRenderTarget`) behind a separate build tag,
      with deterministic CPU readback for golden tests.
- [ ] Capability reporting split between CPU and GPU configurations so the
      comparison report shows truthful native / fallback / unavailable status
      per mode.
  - [x] CPU Skia capability reporting now marks implemented optional paths as
        native instead of fallback; GPU-specific capability reporting remains
        deferred until the GPU build tag exists.
- [ ] Skia vs AGG semantic-fixture comparison; tolerances documented per
      fixture where Skia is not expected to pixel-match.

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

**Goal:** turn individual RMSE fixes into general Matplotlib parity fixes. A
change that improves one catalog case is not complete until it is checked
against a varied set of related fixtures and shown not to rely on example-local
behavior.

This phase exists because cases such as `figure_labels_composition` can expose
real core issues in constrained layout, rotated labels, figure-level labels,
text metrics, and AGG rasterization. The correct outcome is a reusable library
fix that also improves or preserves the neighboring examples, not a narrow
patch that only makes one PNG pass.

**Policy:**

- [x] No example-source workaround is acceptable unless the Go example was
      demonstrably not equivalent to the Matplotlib reference source.
- [x] No fixture-specific logic, magic offsets, or catalog-ID conditionals are
      acceptable in core, layout, renderer, or backend code.
- [x] Any empirical correction must either be replaced with a source-backed
      model from `third_party/matplotlib`, converted into a renderer/geometry
      invariant with tests, or explicitly tracked as unresolved before merge.
- [x] Validate every parity fix against a fixture cluster, not just the first
      failing case. For layout/text work this includes at least
      `figure_labels_composition`, `text_labels_strict`,
      `axes_top_right_inverted`, `axes_control_surface`,
      `transform_coordinates`, `annotation_composition`, and
      `colorbar_composition` when relevant.
- [x] For image, mesh, colorbar, and rasterizer work, also check the related
      image/mesh fixtures such as `image_heatmap`, `imshow_clipped`,
      `imshow_transformed`, `spy_image`, `pcolor_flat`,
      `pcolormesh_gouraud`, `boundarynorm_pcolormesh`, `lognorm_imshow`,
      `twoslope_norm_image`, and `colorbar_extensions`.
- [x] For 3D or projection fixes, check the corresponding `mplot3d_*`,
      polar, geographic, radar, and skew fixtures so a local improvement does
      not regress another coordinate system.
- [x] If the residual points to a fundamental issue in the AGG Go port, it is
      valid to modify `../agg_go` instead of compensating inside
      `matplotlib-go`. Such changes must be treated as renderer fixes and
      covered by AGG-level tests or cross-checked through parity fixtures.

**Validation clusters:**

- [x] `internal/examplecatalog.ValidationClusters` defines named Phase 8A
      clusters for `layout-text`, `image-mesh-colorbar`, and `projection-3d`.
- [x] Catalog tests enforce the required cluster members from this phase so
      future parity fixes can cite a stable validation group instead of a
      single fixture.
- [x] `TestNoCatalogIDsInLibraryImplementation` scans non-test
      `core/`, `render/`, and `backends/` implementation files for quoted
      catalog IDs so library code cannot silently grow fixture-specific
      branches.
- [x] `internal/examplecatalog.ParityFixValidationTargets` maps every Phase 8
      case to the named cluster(s) that must be used before accepting a fix.
      `TestParityFixValidationTargetsNameClusters` enforces complete coverage,
      valid cluster IDs, and cluster membership.

**Resolved empirical corrections / renderer nudges:**

- [x] Owner: layout/text parity. `core/layout_engine.go` now derives
      constrained-layout padding and default inter-cell spacing from
      Matplotlib's rc defaults (`figure.constrained_layout.{h_pad,w_pad}` =
      3 pt; `{hspace,wspace}` = `0.02`) and its padding rule in
      `third_party/matplotlib/lib/matplotlib/_constrained_layout.py`. The
      previous `AxisLineWidth`-scaled row/x-label nudges were removed.
- [x] Owner: figure/text layout parity. `core/figure_layout.go` now derives
      figure-label autopositions from Matplotlib's `Figure.suptitle`,
      `supxlabel`, and `supylabel` defaults (`y=0.98`, `y=0.01`, `x=0.02`),
      constrained-layout figure-label pads from Matplotlib's constrained
      layout pad, and figure-artist stacking separation from
      `legend.borderaxespad = 0.5` font-size units. The previous fixed 4 pt
      gaps and rotated `SupYLabel` `AxisLineWidth` y-anchor nudge were removed.

**Exit criteria:**

- [x] Every accepted Phase 8 fix names the fixture cluster used for validation.
- [x] Any remaining empirical constants or renderer-specific nudges are listed
      with an owner, rationale, and removal path.
- [x] The catalog contains enough varied fixtures that a layout, text, image,
      mesh, projection, or 3D fix cannot silently overfit one example.

---

# Phase 9: Matplotlib API Coverage Audit

**Goal:** close the *existence* gap, not just the *quality* gap. Phase 8's RMSE
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

# Phase 9A: Feature and Demo Coverage Audit

**Goal:** make missing Matplotlib parity coverage visible before v1.0 by
auditing each fundamental Matplotlib feature area for three things: direct Go
equivalent, catalog / demo coverage, and whether the demo actually exercises
the feature breadth.

**Reference sources:** `third_party/matplotlib/lib/matplotlib/`,
`third_party/matplotlib/galleries/examples/`, `internal/examplecatalog/`,
`examples/`, `test/parity/`, and `test/matplotlib_ref/`.

### 9A.1 Coverage Matrix

- [x] Add a machine-readable coverage inventory that maps upstream Matplotlib
      modules / gallery families to local Go implementation files, catalog IDs,
      user-facing examples, browser demos, and intentional omissions.
- [x] For each row, record status for direct Go translation / equivalent,
      parity fixture, user-facing showcase, browser demo, and breadth quality.
- [x] Fail CI when a catalog topic or upstream gallery family is marked
      implemented but has neither a fixture nor an intentional omission note.

Current slice landed:

- `internal/examplecatalog.FeatureCoverageMatrix` records the initial Phase 9A
  inventory across foundational Matplotlib areas: artists, axes, figure/layout,
  axis/ticker/scale, transforms, lines, collections, patches, text/annotation/
  legend, images, colorbars, colors/colormaps/norms, pyplot, renderer/backends,
  widgets/events/animation, and toolkit/projection families.
- Catalog tests now enforce stable coverage row IDs, non-empty upstream
  references and coverage statuses, valid catalog/showcase/web-demo references,
  existing Go implementation file paths, and the implemented-row rule that each
  implemented area must have parity fixtures or an explicit omission note.

### 9A.2 Foundation API Gaps

- [x] Audit `artist.py`, `axis.py`, `ticker.py`, `scale.py`, `transforms.py`,
      `lines.py`, `collections.py`, `patches.py`, `text.py`, `image.py`,
      `colorbar.py`, `cm.py`, `colors.py`, `pyplot.py`, and
      `backend_bases.py` against the Go equivalents.
- [x] Track missing or thin fundamentals: artist property / stale / callback
      APIs, richer locator / formatter catalog, explicit tick-artist behavior,
      transform / BBox breadth, collection variants, patch and box / arrow
      style registries, advanced text / font behavior, image classes, colorbar
      orientation / tick behavior, advanced norms / LightSource, and high-value
      pyplot wrappers.
- [x] For each gap, decide: implement, expose through an idiomatic Go
      equivalent, or document as an intentional divergence.

Current slice landed:

- `internal/examplecatalog.FoundationAPIGapAudit` records stable Phase 9A.2
  gap decisions across artist properties and per-artist clipping, ticker /
  formatter / scale gaps, tick artist behavior, transform / BBox breadth,
  Line2D marker and data semantics, collection variants and scalar-mappable
  updates, patch style registries, text / font / annotation layout, image
  classes and interpolation policy, colorbar placement / ticks, advanced
  colors / norms / LightSource, pyplot wrappers, and backend canvas / manager
  lifecycle.
- Catalog tests now enforce coverage for every Phase 9A.2 upstream module,
  stable required gap IDs, valid linked Phase 9A coverage rows, non-empty
  current-equivalent / gap / decision / rationale text, and existing Go file
  references for each gap row.

### 9A.3 Demo Breadth Gaps

- [x] Promote fixture-heavy basics into user-facing demos where the current
      showcase is too thin: marker grids, advanced scatter,
      grouped / horizontal / stacked bars, fill_between / stacked fill,
      histogram density / strategies, named colors, colormap families, image
      interpolation / alpha / matshow / spy, and colorbar norm / extension
      variants.
- [x] Add or expand feature-breadth galleries for MathText,
      ticks / scales / formatters, text alignment / rotation / wrapping,
      annotations / legends / offset boxes, mplot3d,
      geographic / radar / skew projections, axisartist, axes_grid1,
      unstructured triangulation, and mixed raster / vector output.
- [x] Keep examples close to the upstream Matplotlib examples; fix core
      behavior rather than adjusting examples to hide parity gaps.

Current slice landed:

- `internal/examplecatalog.DemoBreadthGaps` records stable Phase 9A.3 demo
  breadth rows for marker/line grids, advanced scatter, bar and fill variants,
  histogram strategies, named colors, colormap families, image variants,
  colorbar norm/extension variants, MathText, ticks/scales/formatters,
  text layout, annotation/legend/offset boxes, mplot3d, projection/toolkit
  breadth, unstructured triangulation, and mixed raster/vector output.
- Catalog tests now enforce required demo-breadth IDs, valid catalog/showcase/
  web-demo references, non-empty current-coverage / need / recommended-demo
  text, supported priorities, actionable high-priority target-feature lists,
  and linkage from high-priority demo gaps back to thin or fixture-only Phase
  9A coverage rows.

### 9A.4 Browser Demo Coverage

- [x] Reconcile `test/matplotlib_ref/webdemos/` with
      `internal/examplecatalog.WebDemos()`: wire unused web reference modules
      into the browser gallery or mark them as reference-only.
- [x] Add browser demos for existing showcase families that are currently
      CLI-only but important for inspection: annotations, bars, errorbars,
      fills, heatmaps, histograms, lines, patches, scatter, subplots, mplot3d,
      and projection / toolkit demos.
- [x] Ensure browser demos use the same catalog source as parity tests so web
      coverage cannot drift from reference coverage.

Current slice landed:

- `internal/examplecatalog.BrowserDemoCoverageRows` records stable Phase 9A.4
  reconciliation rows for every inactive `test/matplotlib_ref/webdemos/*.py`
  module, marking catalog-backed modules as planned browser work and keeping
  `radialforce` reference-only until it is promoted to a catalog case.
- The same inventory records every current `Showcase: true` catalog row without
  a `WebDemoID`, so CLI-only examples have explicit browser-demo follow-up
  actions tied back to the catalog source of truth.
- Catalog tests now enforce stable browser coverage rows, valid catalog and
  active-web-demo references, complete reconciliation of Python web reference
  modules, and complete accounting for CLI-only showcases.

### 9A.5 Reference Consistency

- [x] Ensure every catalog case has matching Go and Python parity sources under
      `test/parity/<id>/`.
- [x] Add any missing Matplotlib reference plot modules for catalog IDs that
      currently rely on generated images without a visible source counterpart.
- [x] Add a catalog test that reports cases with fixture-only coverage but no
      user-facing showcase when the topic is considered public API.

Current slice landed:

- Catalog tests now enforce canonical `test/parity/<id>/plot.go` and
  `test/parity/<id>/plot.py` source pairs for every catalog case.
- Catalog tests now enforce a visible
  `test/matplotlib_ref/plots/<id>.py` module for every catalog case; the
  missing bilinear and bicubic imshow reference modules were added.
- `internal/examplecatalog.ReferenceConsistencyClassifications` records
  fixture-only public API cases that are intentionally backend-stress,
  mesh-shading, signal-demo, or compound-statistics fixtures instead of
  standalone user-facing showcases.
- `docs/phase-9a-coverage-audit.md` explains how to read the Phase 9A
  inventories, summarizes the current findings, and clarifies that the exit
  criteria are audit criteria unless later phases explicitly promote the
  discovered gaps into implementation work.
- `internal/examplecatalog.CoverageAuditExitCriteria` makes that exit-criteria
  interpretation testable, including which criteria are audit-satisfied and
  which still expose follow-up implementation or browser-demo work.

**Audit exit criteria:**

- [x] Every fundamental Matplotlib feature area is classified as implemented,
      partially implemented, intentionally omitted, or pending.
- [x] Every implemented public feature has at least one parity fixture or a
      documented reason why visual parity testing is not applicable.
- [x] Every major user-facing feature family has an audit row explaining
      whether showcase coverage is broad, thin, fixture-only, pending, or
      intentionally omitted.
- [x] Browser demo coverage is reconciled against the catalog so planned,
      active, and reference-only browser work cannot drift silently.

**Important:** Phase 9A answers "what is missing?" at audit granularity. It
does not claim that every missing Matplotlib API or example has been
implemented. The implementation work discovered by Phase 9A continues in
Phases 9B-9E below.

---

# Phase 9B: Exhaustive Public Surface Parity Map

**Goal:** turn the coarse Phase 9A audit into the detailed answer originally
needed: for each relevant upstream Matplotlib public API, enumerable registry,
and gallery family, state whether the Go port has a direct equivalent, an
idiomatic equivalent, an intentional omission, or no implementation yet.

**Reference sources:** `third_party/matplotlib/lib/matplotlib/`,
`third_party/matplotlib/galleries/examples/`, `internal/examplecatalog/`,
`core/`, `transform/`, `render/`, `color/`, `style/`, `pyplot/`, `canvas/`,
`backends/`, `examples/`, and `test/parity/`.

### 9B.1 Public API Inventory Generator

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

### 9B.2 Go Equivalent Mapping

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

### 9B.3 Human Parity Status Report

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

# Phase 9C: Foundational API Parity Closure

**Goal:** implement or explicitly omit the missing fundamental APIs surfaced by
`FoundationAPIGapAudit`, prioritizing behavior that affects visual parity,
Matplotlib migration examples, or broad public use.

### 9C.1 Artist, Line2D, and Marker Semantics

**Goal:** close the foundational `artist.py`, `lines.py`, and `markers.py`
gaps that affect static rendering, legends, migration examples, and parity
fixtures. Keep the Go API explicit and typed, but match Matplotlib behavior
where the output is visible.

#### 9C.1A Landed Baseline

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

#### 9C.1B Shared Artist Metadata Remainder

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

#### 9C.1C Line2D Data and Stroke Semantics

- [ ] Add typed data getters/setters for `Line2D`: clone-returning `Data`,
      `SetData`, `SetXData`, and `SetYData`, with stale invalidation.
- [ ] Define and implement NaN / Inf segment splitting to match Matplotlib's
      line-break behavior instead of drawing through invalid points.
- [ ] Add `gapcolor` support for dashed lines where Matplotlib paints dash gaps
      with an alternate color.
- [ ] Expand `MarkEvery` beyond every-N integers:
      start/step tuple, explicit index list, and slice-like range where that
      maps cleanly to Go.
- [ ] Add catalog/parity cases for data mutation, invalid-point line breaks,
      dashed `gapcolor`, and at least two nontrivial `markevery` forms.

#### 9C.1D Line2D Marker Completion

- [ ] Verify current `Line2D` marker size conversion, marker face/edge fallback,
      marker edge width, marker alpha, and legend samples against upstream
      `lines.py` and `markers.py`.
- [ ] Support marker face color `"none"` semantics through an explicit Go option
      or sentinel that does not rely on ambiguous zero-alpha fallback.
- [ ] Support marker edge color `"auto"` / face-color fallback semantics where
      Matplotlib does so for common markers.
- [ ] Ensure line-only markers (`+`, `x`, ticks, carets, etc.) use stroke-only
      marker rendering in both plot output and legends.
- [ ] Add catalog/parity cases for Line2D markers in plots and legends,
      including filled, unfilled, line-only, custom path, tuple marker, and
      mathtext marker variants.

#### 9C.1E Half-Filled Marker Paths

- [ ] Replace the current single filled marker path behavior for
      `MarkerFillLeft`, `MarkerFillRight`, `MarkerFillTop`, and
      `MarkerFillBottom` with split marker drawing: primary half uses
      markerfacecolor, alternate half uses markerfacecoloralt / transparent
      fallback, and edge drawing remains whole-marker.
- [ ] Implement split paths for circle, square, diamond, triangle, and polygon
      markers first; explicitly document or implement behavior for line-only,
      tuple, custom path, and mathtext markers.
- [ ] Add renderer-neutral tests that inspect the two fill paths and one edge
      path for a half-filled marker.
- [ ] Add catalog/parity fixture coverage for all four half-fill directions.

#### 9C.1F Exit Criteria

- [ ] `FoundationAPIGapAudit` row `artist-clipping-transform` is either closed
      or split into smaller remaining rows with exact implementation scope.
- [ ] `FoundationAPIGapAudit` row `line2d-marker-data-semantics` is either
      closed or split into smaller remaining rows with exact implementation
      scope.
- [ ] Public-surface parity rows for `Artist`, `Line2D`, `MarkerStyle`, marker
      registries, and fillstyle registries are updated from `partial` to
      precise final statuses or linked to the remaining 9C.1 subtask.
- [ ] 9C.1 has at least one catalog/parity case each for visibility, alpha,
      clipping, Line2D markers in legends, invalid line data, `markevery`,
      `gapcolor`, and half-filled markers.
- [ ] `go test ./core -count=1` and the relevant catalog cases under
      `go test ./test/ -run ...` pass after the behavior changes.

Implementation notes:

- Compare against `third_party/matplotlib/lib/matplotlib/artist.py`,
  `lines.py`, and `markers.py` before changing behavior.
- Add focused catalog cases for marker fill styles, Line2D markers in legends,
  clipping, alpha, and visibility.
- Prefer adding shared Go option structs/mixins over copying Python's dynamic
  setter model wholesale.

### 9C.2 Axis, Ticker, Formatter, Scale, and Transform Breadth

- [ ] Expand locator/formatter coverage for engineering, percent, index, null,
      scalar, log-mathtext, minor ticks, multi-level date formatting, and
      scale-specific formatting.
- [ ] Expose tick styling through Go axis/tick options without requiring a full
      Python-style `Tick` artist clone.
- [ ] Add parity-driven transform/BBox/path helpers for frozen transforms,
      transformed paths, invalidation, clipping, annotation coordinate modes,
      and layout calculations.

Implementation notes:

- Compare against upstream `axis.py`, `ticker.py`, `scale.py`,
  `transforms.py`, `path.py`, and `bezier.py`.
- Add one catalog case per formatter/locator family instead of one enormous
  fixture.
- Keep the transform graph lean; add only helpers needed by rendering,
  annotations, layout, and clipping parity.

### 9C.3 Collections, Scalar Mapping, Meshes, and Colorbars

- [ ] Complete collection scalar-mappable behavior: mutable arrays,
      face/edge-color updates, norm/colormap updates, offset transforms, and
      colorbar synchronization.
- [ ] Expand colorbar orientation, placement, anchor/location, custom tick,
      boundary, spacing, drawedges, extension, and multi-axes behavior.
- [ ] Add or omit remaining advanced norms and color machinery: `FuncNorm`,
      `AsinhNorm`, multivar/bivar colormaps, and `LightSource`.

Implementation notes:

- Compare against upstream `collections.py`, `cm.py`, `colors.py`,
  `colorbar.py`, and `colorizer.py`.
- Core behavior should be fixed in `core/collection.go`,
  `core/scalar_mappable.go`, `core/norm.go`, and `core/colorbar.go`, not by
  tweaking examples.
- Add Matplotlib reference cases for mutable scalar-mapping and horizontal /
  boundary colorbars.

### 9C.4 Patches, Text, Annotation, Legend, and Offset Boxes

- [ ] Audit `FancyBboxPatch` box-style coverage against
      `BoxStyle._style_list` and implement or document every missing style.
- [ ] Verify hatch pattern characters and repeat-density semantics against
      `hatch.py`.
- [ ] Expand text/font property support: family, style, weight, stretch,
      variant, math font, parse-math behavior, font features, and per-text
      font options.
- [ ] Implement missing annotation coordinate modes, annotation clipping,
      `AnnotationBbox`, offset-box families, legend handler maps, proxy-like
      legend entries, and legend layout behavior.

Implementation notes:

- Compare against upstream `patches.py`, `hatch.py`, `text.py`,
  `font_manager.py`, `legend.py`, `legend_handler.py`, and `offsetbox.py`.
- Add small catalog fixtures for each style family; avoid a single giant patch
  fixture that is hard to debug.
- Keep API shapes Go-idiomatic, but the rendered output should follow
  Matplotlib where behavior is visual.

### 9C.5 Images, Pyplot, Backends, Widgets, and Animation

- [ ] Add remaining image class decisions: `FigureImage`, `BboxImage`,
      `NonUniformImage`, `PcolorImage`, `pcolorfast`, and `figimage`.
- [ ] Complete interpolation policy for `lanczos`, `spline16`, `spline36`,
      `kaiser`, `quadric`, `catrom`, `gaussian`, `bessel`, `mitchell`, `sinc`,
      `blackman`, `hermite`, `antialiased`, and `auto`.
- [ ] Expand high-value `pyplot` wrappers where they materially improve
      migration, while keeping object-oriented Go APIs primary.
- [ ] Complete backend canvas/manager/tool lifecycle semantics needed for
      interactive backends.
- [ ] Decide which widgets and animation APIs are in scope for v1.0, then add
      fixtures/examples or intentional omissions.

Implementation notes:

- Compare against upstream `image.py`, `pyplot.py`, `_pylab_helpers.py`,
  `backend_bases.py`, `backend_tools.py`, `widgets.py`, and `animation.py`.
- Any unsupported interpolation or widget path must produce a clear error or
  documented omission, never silently fall back to a wrong default.

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

# Phase 9D: User-Facing Example Breadth

**Goal:** ensure every major implemented public feature family has a
user-facing Go example that demonstrates meaningful Matplotlib-equivalent
variants, not just a parity fixture.

### 9D.1 Core Plot Family Galleries

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

### 9D.2 Color, Image, Text, and Annotation Galleries

- [ ] Add named-color swatches and colormap family galleries.
- [ ] Add image interpolation/alpha/matshow/spy galleries.
- [ ] Add colorbar norm/extension galleries.
- [ ] Add MathText, text layout, annotation, legend, and offset-box galleries.

Implementation notes:

- Prefer several focused examples over one overloaded gallery when visual
  differences need inspection.
- Include captions/descriptions in catalog metadata explaining what feature
  breadth the example validates.

### 9D.3 Toolkit, Projection, 3D, and Backend Output Galleries

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

# Phase 9E: Browser Gallery Alignment

**Goal:** make the browser gallery a catalog-backed inspection surface for the
same feature families covered by parity fixtures and CLI examples.

### 9E.1 Wire Planned Web Reference Modules

- [ ] Wire `test/matplotlib_ref/webdemos/annotations.py`, `bars.py`,
      `errorbars.py`, `fills.py`, `heatmap.py`, `histogram.py`, `lines.py`,
      `patches.py`, `scatter.py`, and `subplots.py` into active browser demos
      or fold them into existing catalog-backed browser families.
- [ ] Keep `radialforce.py` reference-only until it is promoted to a catalog
      case.
- [ ] Add tests that every active web reference module maps to a catalog case
      and every catalog-backed planned row either has an active browser demo or
      remains explicitly planned.

### 9E.2 Promote CLI-Only Showcases

- [ ] Promote CLI-only showcases listed in `BrowserDemoCoverageRows` into
      browser demos: basic lines, dashes, scatter, bars, fills, errorbars,
      multi-series, histograms, boxplots, heatmaps, figure labels, colorbars,
      annotations, projections, mplot3d, triangulation, axisartist, and
      axes_grid1.
- [ ] Browser demos must use the same catalog factories as parity tests or a
      documented wrapper around them.
- [ ] Add browser-demo smoke tests that render each promoted demo and verify it
      has a non-empty image/artifact.

### 9E.3 Browser Parity Status Reporting

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

# Phase 10: Documentation, Examples Polish, and v1.0 Release

**Goal:** make the project consumable by users who have not been following
the development thread, and tag a stable v1.0.

### 10.1 API Documentation

- [ ] Package-level GoDoc passes for every public package, with a worked
      example per package.
- [ ] Hosted documentation site (pkg.go.dev plus a curated landing page
      under the existing GitHub Pages deployment).
- [x] Migration guide from upstream Matplotlib: side-by-side Python / Go
      snippets for every plot family covered by the catalog.
- [x] Backend selection guide: when to use AGG / GoBasic / SVG / PDF /
      Skia, with capability matrix excerpts (`docs/backend-selection.md`).

### 10.2 Examples Gallery Polish

- [x] Review every `Showcase: true` catalog row for caption, description,
      and runnable snippet quality.
- [x] Add an "anti-gallery" of intentional Matplotlib-divergence cases with
      the reasons documented (where the Go port chose different defaults).
- [x] Promote the WASM browser gallery to a first-class entry point on the
      project README.

### 10.3 Performance Pass

- [ ] Profiling sweep across the catalog: identify hotspots that exceed the
      100k-point smoothness goal and the sub-second typical-plot goal.
- [ ] Reusable benchmark suite under `benchmarks/` with regression tracking
      in CI.
- [ ] Documented memory-usage targets and a tuning guide for long-running
      applications.

### 10.4 Release Readiness

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
