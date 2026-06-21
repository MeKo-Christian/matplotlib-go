# matplotlib-go — Translation Fidelity Review

**Date:** 2026-06-21
**Reference:** matplotlib 3.10.9 (`third_party/matplotlib`), FreeType 2.6.1, DejaVu Sans 2.35
**Scope:** Independent review of how faithfully the Go port translates upstream matplotlib, conducted package-by-package via parallel subagent analysis of `core/`, `transform/`, `geom/`, `color/`, `render/`, `backends/`, `style/`, `pyplot/`, `animation/`, and the external `mathtext` module. Emphasis on **deficits and gaps**, per the review request, with strengths noted for balance.

---

## 1. Executive Summary

matplotlib-go is a **remarkably broad and, in its core numerical algorithms, high-fidelity** port. The hard, easy-to-get-wrong parts — affine/non-linear transforms, the `MaxNLocator` tick algorithm, marching-squares contours, 3D projection + depth sorting, the norm/colormap pipeline, mathtext layout, and FreeType-exact text rasterization — are faithful, often line-for-line ports verified against the 3.10.9 source. The project's parity infrastructure (catalog-driven golden + reference tests at RMSE < 5) is genuinely strong and the visual output is close to matplotlib.

The weaknesses are **not in correctness of what is implemented, but in breadth of configuration and a handful of structural gaps**. The project's own generated `docs/matplotlib-parity-status.md` classifies most of its 16 feature families as *"partial / thin"* — and that self-assessment is accurate. The recurring pattern across every subsystem is: **the primary algorithm is faithful, but the long tail of matplotlib kwargs, modes, and edge cases is missing**, sometimes silently.

### Fidelity scorecard

| Subsystem | Core fidelity | Breadth | Most significant gap |
|---|---|---|---|
| Transforms & geometry | ★★★★★ | ★★★☆☆ | No Bézier toolkit, no triangulation, closed transform type set, affine/non-affine cache split unused |
| Ticks / locators / scales | ★★★★★ | ★★★★☆ | `ScalarFormatter` has **no offset/×10ⁿ text**; date numbers are Unix-seconds not mpl ordinals |
| 1D/2D plotting artists | ★★★★☆ | ★★★★☆ | `boxplot` default fills the box (should be unfilled); `alpha=0` treated as "unset" |
| Figure / Axes / layout | ★★★★☆ | ★★★☆☆ | **No real constrained/tight-layout solver** (one approximation for both); no `margins()`, no `datalim` aspect |
| Color / cmap / norm | ★★★★★ | ★★★★☆ | Colorbar tick locators wrong for SymLog/Power/TwoSlope norms; no native RGBA `imshow` |
| Contour | ★★★★☆ | ★★★☆☆ | No `negative_linestyles` (default dashing), no `extend`, no `hatches` |
| Text layout | ★★★★★ | ★★★★☆ | Per-glyph font fallback missing (tofu instead of family walk) |
| MathText | ★★★★☆ | ★★☆☆☆ | **~80 of 632 symbols**; unknown commands silently echo as literal text |
| mplot3d | ★★★★☆ | ★★★★☆ | `PlotSurface` is a mislabeled stub; no axis inversion, no `add_collection3d` |
| Renderer contract + AGG | ★★★★★ | ★★★★☆ | Sketch/xkcd is a no-op; Gouraud shading not antialiased; gradients clamped to 2–3 stops |
| Vector backends (PDF/SVG) | ★★★★☆ | ★★★★☆ | Vector `MeasureText` are crude stubs; PS/PGF lack gradients/patterns |
| Skia backend | ★★★☆☆ | — | GPU mode is **0% wired** (scaffolding only); CPU-only via agg fallback |
| Public API / pyplot | ★★★★★ | ★★★★☆ | Excellent matplotlib-shaped API; more thread-safe than mpl |
| Styling / rcParams | ★★★☆☆ | ★★☆☆☆ | **~13% rcParam coverage**; only 4 built-in themes; color-only cycler |

---

## 2. Cross-Cutting Themes

These patterns recur across multiple packages and are the most actionable findings.

### 2.1 Silent failure modes (highest-priority class of bug)
Several subsystems degrade **silently** rather than erroring or warning, which is the worst failure mode for users porting real matplotlib scripts:
- **MathText:** unknown commands echo as literal text (`mathtext .../parser.go:228`). A typo like `\inteGral`, or any of the ~550 unsupported symbols, renders as a broken literal string with no diagnostic.
- **`alpha=0` means "unset":** `Bar`/`Hist`/`Fill` promote a user-requested fully-transparent `alpha=0` to `1.0` (`core/plot.go:486`, `bar.go:77`), inconsistent with the pointer-for-unset convention used elsewhere in the same structs.
- **Unknown colormap → viridis:** `GetColormap` returns viridis for genuinely unknown names instead of erroring, and names are case-folded (`Blues`==`blues`), unlike matplotlib (`color/colormap.go:391,428`).
- **Gouraud QuadMesh silently falls back** to flat cell colors when the renderer lacks the capability (`core/collection_quadmesh.go:214`).
- **Invalid input → nil artist:** length-mismatched data returns a `nil` artist with no error.

### 2.2 "Primary algorithm faithful, configuration tail missing"
The dominant shape of every gap. Examples: `Bxp` (low-level boxplot) is faithful but high-level `BoxPlot2D` hardcodes vertical + filled; `clabel` core works but lacks `fmt`/`rightside_up`; contours compute correctly but ignore `extend`/`linestyles`/`hatches`; colorbars place ticks correctly only for linear/log/boundary norms.

### 2.3 Text metrics inconsistency across backends
The AGG path uses the real FreeType-backed shared font manager and is pixel-exact. But the vector backends each ship a **crude `MeasureText` stub** — PDF `0.5·size·len` (`backends/pdf/pdf_text.go:46`), PS/PGF `0.6·size·runes`, SVG a 7×13 bitmap font (`backends/svg/text.go:53`). Any layout relying on a vector backend's own metrics (rotated/vertical anchoring) misplaces text. Vector text is only reliable through the agg-backed shared font pipeline.

### 2.4 Declared-but-unused capability
Two notable cases where infrastructure exists but isn't exploited:
- The transform invalidation model declares an affine/non-affine split (`InvalidAffine`/`InvalidNonAffine`) but `TransformedPath` keeps one full-path cache and never separates the cheap display-affine update from the expensive projection — defeating the main performance reason matplotlib has the split (`transform/transformed_path.go:13`).
- `SketchParams` threads through the whole `Paint`/`GraphicsContext`/renderer contract but **no backend consumes it** — xkcd mode is a no-op (`render/render.go:277`).

---

## 3. Per-Subsystem Findings

### 3.1 Transforms & Geometry — `transform/`, `geom/`
**Excels:** Affine algebra is correct (`geom.go:563`); the SymLog `linscale/(1−base⁻¹)` adjustment is byte-identical to matplotlib (`scale_registry.go:341`); the full non-linear scale family (log/symlog/asinh/logit/func/funclog) is faithful; invalidation node model is a sound simplified port.
**Deficits:**
- No general Bézier module (`bezier.py` equivalent) — no `split_bezier`, arc-length, offset curves used by fancy arrows/annotation connectors.
- No triangulation (`tri/` — Delaunay/TriFinder/TriInterpolator) anywhere; tricontour/tripcolor build their own.
- No path-generator helpers (`unit_circle`, `arc`, `wedge`, 4-cubic circle) — pushed up into `core/`.
- Closed transform type set: `AsAffine`/`Frozen` switch on concrete types; a third-party transform can't participate in flattening (no `get_affine()` capability interface).
- No live bbox-linked transform (`BboxTransformTo`); resizing rebuilds transforms instead of invalidating a node.
- Path simplifier is **Douglas–Peucker** (`agg_path_simplify.go:5`), not matplotlib's single-pass running-segment algorithm — won't pixel-match on dense lines, and bails entirely on any curve/close.
- `internal/geom/` is empty.

### 3.2 Ticks / Locators / Scales / Dates / Units — `core/tick_*.go`, `date_tick.go`, `units.go`
**Excels:** `MaxNLocator` is a genuine port of `_raw_ticks` incl. `_staircase`/`_Edge_integer` (`tick_locators.go:265`) — the hardest locator, done well. LogitLocator, SymmetricalLogLocator, AsinhLocator all faithful. Rich formatter set (EngFormatter with full SI table, PercentFormatter, LogFormatterMathText sparse-mantissa logic). Scale registry covers the full func-scale family. Units framework genuinely extensible (date/category/custom converters).
**Deficits:**
- **`ScalarFormatter` emits no offset/scientific-multiplier text** — the `+1e6` additive offset and `×10ⁿ` axis multiplier that matplotlib draws *by default* are absent (`tick_formatters.go:73`); the struct comment admits it. Biggest fidelity gap for ordinary large-magnitude linear plots. Offset text is also X-axis-only.
- **Date numbers are Unix seconds**, not matplotlib's days-since-1970 (`date_tick.go:892`). No `date2num`/`num2date`/`set_epoch` public API. Off by 86400× from any hand-computed mpl ordinal.
- `AutoLocator` hardcodes `nbins=9` instead of axis-length-aware `'auto'`; `LogLocator` same root cause; `AutoMinorLocator` ignores the leading-digit 4-vs-5 rule.
- Date locators don't implement `nonsingular` (degenerate range → single tick instead of ~4-year expansion).
- `LogLocator` with `stride>1` + subs returns nil (can blank an axis).

### 3.3 1D/2D Plotting Artists — `core/plot.go`, `scatter.go`, `boxplot.go`, …
**Excels:** Near-complete method coverage. Marker system is excellent — all marker strings, tuple/mathtext/path markers, half-fill clipping, per-marker snap (`scatter.go:280`). Scatter scalar mapping complete. `Bxp` faithfully mirrors `axes.bxp` (notch geometry, usermedians, conf_intervals). `fill_between` where+interpolate+step modes correct (`fill.go:359`). Histogram binning complete (Sturges/Scott/FD/auto, density/cumulative/weights).
**Deficits:**
- **High-level `BoxPlot2D` always fills the box** with the cycle color (`boxplot.go:558`); matplotlib default is `patch_artist=False` → unfilled box, black outline. No `patch_artist` toggle. Also vertical-only (no orientation), no showbox/showcaps/showmeans/sym on the high-level path.
- `bootstrap` accepted but ignored; whisker-percentile snaps to nearest datum instead of using the percentile value.
- `StackPlot` only supports `baseline='zero'` (no wiggle/weighted_wiggle/sym).
- `Stem` has no `orientation`; errorbar `capthick` not independent of `elinewidth`.
- Scatter keeps non-finite points in bounds (no `plotnonfinite`), can corrupt autoscale.
- One mega `PlotOptions` struct (`plot.go:15`) mixes ~40 unrelated 1D/contour/3D fields.

### 3.4 Figure / Axes / Layout — `core/layout_engine.go`, `gridspec.go`, `legend*.go`
**Excels:** GridSpec is the most matplotlib-accurate part of the layout stack — spans, ratios, nested grids, mosaic, subplot2grid, exact default margins (`gridspec.go`). Spine data-coordinate positioning with correct y-flip-then-snap ordering. Twin/secondary axes modeled correctly with functional forward/inverse scales. Legend "best" placement genuinely emulates the badness scan. Draw-order z-banding matches matplotlib.
**Deficits:**
- **No true constrained/tight-layout solver.** `TightLayout()` and `ConstrainedLayout()` run the *same* 2-iteration measured-margin approximation differing only by pad constants (`layout_engine.go:182`). matplotlib's iterative `LayoutGrid` constraint solver with per-row/col margin propagation and suptitle/colorbar/legend reservations is absent. **Single largest layout-fidelity gap.**
- Autoscaling uses a flat 5% margin with no locator rounding (`autolimit_mode='round_numbers'`) and a crude `span==0→1` degenerate fallback (mpl expands by ±0.5·|val|).
- Aspect only supports `adjustable='box'` — no `'datalim'`, no `set_adjustable`, no `anchor`. `axis('equal')` diverges where datalim is expected.
- No per-axes `Margins`/`SetXMargin`/`SetYMargin`.
- No `label_outer()` / automatic inner shared-tick-label suppression (redundant labels on shared subplots).
- Legend lacks `bbox_to_anchor`, `mode='expand'`, explicit handle/label ordering.
- No `cla()`/`clear()`/`remove()`/`delaxes`/`clf` teardown path found.

### 3.5 Color / Colormaps / Norm / Colorbar / Contour — `color/`, `core/norm.go`, `contour*.go`
**Excels:** Color parsing near-complete (named/CSS4/xkcd/hex/tuple/Cn/grayscale). Colormap coverage ~80 maps baked as 256-sample LUTs — viridis family quantizes identically to matplotlib. Norm family complete and verified (`BoundaryNorm._n_regions` ported exactly). ScalarMappable pipeline clean. Collections broad (PathCollection/QuadMesh/LineCollection/PolyCollection/PatchCollection). Contour marching-squares with saddle splitting validated against fixtures.
**Deficits:**
- **Colorbar tick locators only specialized for linear/Log/Boundary/Asinh** — SymLogNorm, PowerNorm, TwoSlopeNorm, CenteredNorm fall through to a generic FuncScale with no norm-aware locator (`colorbar_scale.go:118`). Tick placement diverges.
- Contour: **no `negative_linestyles`** (matplotlib dashes negative contours by default — monochrome contour plots visibly diverge); no `extend`, no `linestyles`, no contourf `hatches`.
- No native RGBA `imshow` — `(M,N,3/4)` arrays unsupported, everything routes through scalar mapping (`image.go:97`). No image `aspect`.
- `FuncNorm` and `MultiNorm` absent; `petroff10` colormap and bivariate/multivariate maps missing.
- `extendfrac` hardcoded 5%; colorbar minor ticks not exposed.
- LineCollection `DashPatterns` field defined but never populated; no linestyle-string conversion.

### 3.6 Text & MathText — `core/text_layout.go`, `core/mathtext.go`, external `mathtext@v0.1.1`
**Excels:** Layout is a faithful line-by-line port of `Text._get_layout` (per-line height/descent clamped to "lp" ref, linespacing, corner-rotation bbox, anchor/default rotation modes, xtick/ytick auto ha/va tables). MathText pixel parity is genuinely close — ink-bbox alignment matching `to_raster`, faithful rule rasterization, ported decimal-kern and integral-drop-sub hacks. Sub/superscript, `\frac`/`\binom`/`\genfrac`, `\sqrt[n]`, `\left…\right` fences, matrix environments all present. **usetex is real and multi-backend** (external latex+dvipng, DVI/TFM parsing) — many ports skip it entirely.
**Deficits:**
- **MathText symbol table is ~80 of matplotlib's 632** (`normalize_tables.go:50`). Missing whole classes: most arrows, most relations, many binary ops, `\varsigma`/`\varpi`/`\digamma`, `\cdots`/`\vdots`/`\ddots`, and the entire `\mathbb`/`\mathcal`/`\mathfrak`/`\mathscr` letter families.
- **Unknown commands silently echo as literal text** (see §2.1).
- Accents use combining marks appended per-character, not a centered separate accent glyph; `\widehat`/`\widetilde` absent; `\overbrace`/`\underbrace`/`\overline`-as-rule/`\stackrel`/`\substack` missing.
- `\mathbf` etc. only swap font face, don't map to Unicode Mathematical Alphanumeric block (not glyph-identical for bold-italic).
- **Font fallback is family→path, not per-glyph** (`render/font_manager.go:757`) — a glyph absent from the primary font renders as tofu instead of walking the family list.
- `Text(bbox=boxstyle=…)` only reaches square/round; sawtooth/arrow/circle FancyBboxPatch styles exist but aren't bridged from the Text artist.

### 3.7 mplot3d — `core/axes3d*.go`
**Excels:** Projection math is a near-verbatim port of `proj3d.py` (numerically-stable rotation form, perspective/ortho selection). Two-tier depth model faithfully mirrors matplotlib (per-face zsort + collection min-depth zorder). Lighting/shading algebraically identical to `_shade_colors`. Broad plot-type coverage incl. voxels with internal-face culling, fill_between3d with coplanarity test, quiver 3D arrowheads. Reprojector pattern means post-plot `SetView`/`SetZLim` correctly re-renders — an ergonomic win over matplotlib.
**Deficits:**
- **`PlotSurface` is a mislabeled stub** aliased to `Plot3D` (a line strip) — a user calling the matplotlib-named method gets silently-wrong output (`axes3d.go:382`). The real surface is `Surface`/`PlotSurfaceGrid`.
- `Voxel` (singular) ignores shading, renders edges only.
- **No axis inversion** (`invert_zaxis` not honored in projection).
- No `add_collection3d`, no 3D `clabel`, no oriented `text(zdir=…)`/`text2D`.
- `axlim_clip` for lines splits runs instead of masking (edge-crossing segments differ).

### 3.8 Renderer Contract & Raster Backends — `render/`, `backends/agg/`
**Excels:** Clean minimal 8-verb `Renderer` core + ~50 optional capability interfaces with a faithful default-fallback model (`core` checks `r.(MarkerDrawer)`, falls back to per-path loop — exactly matplotlib's `RendererBase.draw_markers`). AGG asserts 26 optional interfaces — unusually complete. AA-off binary-coverage and SNAP_AUTO rectilinear test correctly ported. Path effects are a faithful renderer-neutral replay. Image interpolation comprehensive (~17 filters + auto-resample rule). The queryable capability matrix is a genuine strength matplotlib has no equivalent of.
**Deficits:**
- **Sketch/xkcd is a no-op** everywhere (contract-only, see §2.4).
- **Gouraud shading is a hand-rolled CPU barycentric rasterizer, not antialiased** (`agg_gouraud.go:10`) — `pcolormesh(shading='gouraud')`/`tripcolor` show aliased seams vs reference.
- **Gradients clamped to 2–3 stops** (`gradients.go:34`) — multi-stop colormap gradients misrendered (limited impact; mostly decorative).
- `RestoreRegion` y-flip admittedly unfinished (`agg.go:360`) — latent bug for future blit/animation.
- No `points_to_pixels` hook on the contract — DPI scaling scattered.
- No `url`/`gid` metadata in GraphicsContext — vector backends can't emit hyperlinks/clickable elements.
- Hatch is a bespoke geometric reimplementation with empirical spacing constants, not matplotlib's unit-tile `get_hatch_path`.
- `Paint` vs `GraphicsContext` duplicate most fields; `EffectivePaint` alpha-forcing logic is intricate and a bug magnet.

### 3.9 Vector Backends, Public API & Styling — `backends/{svg,pdf,ps,pgf,skia}/`, `pyplot/`, `animation/`, `style/`
**Excels:** PDF backend is production-grade — real TrueType/CID font embedding, clip path, hatches, axial+radial shadings, ExtGState alpha, image XObjects, path-effect SMask blur. SVG rich — gradients with focal points, `<pattern>` hatches, clip-path, filters, `@font-face` base64 embedding, defs dedup. All four vector backends support mixed rasterization. Public `pyplot` API is excellent and **more thread-safe than matplotlib** (mutex-guarded, per-figure-isolated registry); ~150 top-level functions. Animation complete (FuncAnimation + ArtistAnimation, blit, GIF/APNG/HTML zero-dep writers, optional ffmpeg MP4/WebM). Defaults match matplotlib (640×480 @ 100dpi, Tab10 cycle).
**Deficits:**
- **Vector `MeasureText` are crude stubs** (see §2.3).
- PS/PGF lack gradient and pattern fills (degrade to flat). PGF is the thinnest backend (rect-clip only, no vertical-text/TeX-drawer interfaces). PS Base14 path replaces non-ASCII with `?` and hardcodes Helvetica.
- **Skia GPU is 0% wired** — real C-ABI wrapper and ~95% of the contract exist, but all rendering goes through CPU/agg fallback; `SkSurface::MakeRenderTarget` is `StatusDeferred`.
- **rcParams ~13% coverage** (43 of ~320). Entirely missing: `savefig.*`, `pdf.*`/`ps.*`/`svg.*` export controls, `animation.*`, `boxplot.*` (37 params), `mathtext.*`, `hatch.*`, `image.*`, `keymap.*`, `date.*`.
- Cycler supports only the `color` property (no linestyle/marker/linewidth cyclers).
- Only 4 built-in themes; no bundled `.mplstyle` sheets (`seaborn-*`, `fivethirtyeight`, `bmh`, `Solarize_Light2`).
- Pyplot gaps: `setp`/`getp`/`findobj`, `ginput`, `xkcd()`, `figimage`, per-colormap shortcuts, `clim`/`set_cmap`.

---

## 4. Ease of Use

The library is **deliberately and successfully matplotlib-shaped**, which is its biggest ergonomic win for the target user (a matplotlib veteran moving to Go):

```go
pyplot.FigureSized(640, 480)
pyplot.Plot([]float64{0,1,2,3,4}, []float64{0,1,4,9,16})
pyplot.Title("y = x^2"); pyplot.XLabel("x"); pyplot.YLabel("y")
pyplot.Savefig("parabola.png")
```

**Strengths:** Both paradigms available (stateful pyplot + first-class OO core API). Defaults match matplotlib so new users get matplotlib-looking output with zero config. Thread-safety genuinely better than matplotlib (usable in Go server contexts where `plt` is not). Save-by-extension removes backend-selection friction. `rc_context` via a returned `restore()` closure is the idiomatic Go translation of Python's `with`. Functional options are discoverable and idiomatic. Container parity (BarContainer, ErrorbarContainer, …) eases porting.

**Friction points:**
- Go's lack of kwargs/format-strings means `plt.plot(x, y, 'r--', lw=2)` becomes verbose `...Option` variadics + explicit slices; the `'r--'` shorthand isn't idiomatically reproduced.
- The `alpha=0`-means-unset trap (§2.1) is inconsistent with the pointer convention used in the same structs.
- Setters split across `axes.go`/`axes_limits.go`/`axes_scale.go`/`axes_secondary.go` with no single `Axes` doc surface.
- `TightLayout()` vs `ConstrainedLayout()` behave nearly identically — will surprise users expecting matplotlib's distinct results.
- Styling discoverability weak (4 themes, 13% rcParams) — though the parser *reports* unsupported keys rather than failing silently (good).
- Naming traps in 3D: `PlotSurface`≠surface, `Voxel`≠`Voxels`.

**Overall code quality:** idiomatic Go throughout — value-type transforms/locators/norms with useful zero-values, functional options, no hidden global state (except the deliberately-guarded pyplot registry), error returns on invalid setters. Large files have been decomposed into focused units (per PLAN.md Phase 4). The AGG backend is well-factored into ~15 focused files. The capability-matrix architecture is a standout design.

---

## 5. What Is Missing (beyond incomplete Skia GPU)

Consolidated list of features absent or stubbed, roughly by impact:

**High impact (likely hit by real matplotlib scripts):**
1. `ScalarFormatter` offset / ×10ⁿ axis-multiplier text (default matplotlib behavior).
2. True `constrained_layout` / `tight_layout` constraint solver.
3. MathText symbol long tail (~550 of 632 symbols) + `\mathbb`/`\mathcal`/`\mathfrak`/`\mathscr` + stretchy accents/braces.
4. `date2num`/`num2date`/`set_epoch` + matplotlib date-ordinal convention.
5. Contour `negative_linestyles` (default), `extend`, `linestyles`, `hatches`.
6. Native RGBA `imshow` for `(M,N,3/4)` arrays; image `aspect`.
7. Colorbar tick locators for SymLog/Power/TwoSlope/Centered norms.
8. Per-glyph multi-font fallback (currently tofu).
9. `boxplot` unfilled default (`patch_artist=False`) + orientation on the high-level artist.
10. Per-axes `margins()` / `autolimit_mode='round_numbers'`.

**Medium impact:**
11. `FuncNorm`, `MultiNorm`; `petroff10` and bivariate/multivariate colormaps.
12. Aspect `adjustable='datalim'`, `set_adjustable`, `anchor`.
13. `label_outer()` / shared-axes inner-label suppression.
14. Legend `bbox_to_anchor` / `mode='expand'`.
15. StackPlot `wiggle`/`weighted_wiggle`/`sym` baselines.
16. Antialiased Gouraud shading; >3-stop gradients; sketch/xkcd rendering.
17. mplot3d: `PlotSurface` real implementation, axis inversion, `add_collection3d`, 3D `clabel`, oriented text.
18. Higher rcParams coverage + bundled style sheets + linestyle/marker cyclers.
19. Vector-backend real text metrics (PDF/PS/PGF/SVG `MeasureText`).
20. PS/PGF gradient + pattern fills.

**Lower impact / infrastructure:**
21. Bézier toolkit (`bezier.py`), triangulation library (`tri/`), path-generator helpers, live bbox-linked transforms, open/extensible transform type set.
22. `url`/`gid` metadata for clickable vector output.
23. `cla()`/`clear()`/`remove()`/`delaxes` teardown; `setp`/`getp`/`findobj`.
24. ImageMagick animation writer; `RestoreRegion` y-flip completion (blit correctness).

---

## 6. Recommendations (priority order)

1. **Fix the silent-failure class first** (§2.1) — make unknown mathtext commands and unknown colormaps error or warn; distinguish `alpha=0` from unset (use `*float64`). These are cheap and prevent the worst "broken output, no diagnostic" experience.
2. **Implement `ScalarFormatter` offset/×10ⁿ text** — single highest-impact fidelity gap for ordinary linear plots, and currently users are told to hand-roll a `FuncFormatter`.
3. **Expand the mathtext symbol table** toward the full `tex2uni` set and add the math-alphabet families — the current ~13% coverage is the narrowest subsystem relative to matplotlib.
4. **Either implement a real constrained-layout solver or document that the two layout engines are approximations** that differ only by padding — the current naming implies parity that isn't there.
5. **Add contour `negative_linestyles` default** + colorbar locators for the nonlinear norms — both cause visible divergence in common scientific plots.
6. **Decide the date-number convention** — adopt matplotlib's days-since-epoch (and ship `date2num`/`num2date`), or prominently document the Unix-seconds divergence.
7. **Replace vector-backend `MeasureText` stubs** with the shared FreeType font manager so vector text anchoring matches the raster path.
8. Address the 3D naming traps (`PlotSurface`, `Voxel`) — rename or implement.

---

## 7. Bottom Line

This is a **serious, well-engineered port** whose numerical and rendering core is faithful to matplotlib 3.10.9 at a level most ports never reach — verified line-for-line in the transform, locator, contour, 3D-projection, norm, and mathtext-layout code, and pixel-validated by a strong catalog-driven parity harness. The architecture (capability-interface renderer contract, queryable capability matrix, idiomatic value-type API, better-than-matplotlib thread safety) is a genuine strength.

The honest gap is **breadth, not correctness**: the long tail of matplotlib kwargs/modes is incomplete across nearly every subsystem, a handful of features are stubbed or aliased in misleading ways, and several degradations are silent. The project's own "partial / thin" self-classification is accurate. For a v1.0 that advertises matplotlib parity, the priorities are closing the silent-failure modes, the `ScalarFormatter` offset text, the mathtext symbol tail, and being precise in docs about where "parity" means "the common path matches" versus "the full matplotlib API is present."
