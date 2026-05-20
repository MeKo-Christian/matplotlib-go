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
      three-stop radial uses the multi-stop variant. `SupportsGradientFill`
      advertises native support; `SupportsPatternFill` advertises `false` until
      tile rasterization lands.
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
  defs; PDF transparency groups / soft masks still need native vector
  implementations.
- Golden and Matplotlib-reference fixtures for text shadows, line halos,
  scatter marker shadows, and polygon effect stacks.

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
