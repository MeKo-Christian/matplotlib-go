# matplotlib-go — Translation Fidelity Review

**Date:** 2026-06-30
**Reference:** matplotlib 3.10.9 (`third_party/matplotlib`), FreeType 2.6.1, DejaVu Sans 2.35
**Method:** Independent, skeptical audit conducted via parallel subagents across `core/`, `transform/`, `geom/`, `color/`, `render/`, every backend under `backends/`, `style/`, `pyplot/`, `animation/`, `tri/`, the external `mathtext` module, the example gallery, and the parity-test harness. Emphasis on **locating workarounds, stubs, and silent degradations**, with strengths noted for balance. Every claim below is anchored to `file:line`.

---

## 1. Executive Summary

matplotlib-go is a **genuinely faithful port where it counts, with a scaffolding fringe** — not a facade. The hard, easy-to-get-wrong numerical cores are real, often line-for-line ports verified against the 3.10.9 source: affine transforms, `MaxNLocator`, the norm/colormap pipeline, the mathtext box/glue engine, an exact-predicate Qhull triangulation, and FreeType-exact text rasterization. The parity harness compares against **real matplotlib output** (not the port's own), and the example gallery exercises real features on realistic data rather than dodging hard cases.

The weaknesses are concentrated and identifiable: **(a) the secondary backends** (Skia and its "GPU" mode are scaffolding; the default pure-Go `gobasic` renderer has real gaps), **(b) the capability-advertising layer** that rubber-stamps those gaps, **(c) rcParams breadth** — only ~40% of params are parsed and a third of those are dead storage, and **(d) a handful of silent-degradation hacks** that render plausible-but-wrong output instead of erroring as matplotlib does.

### Fidelity scorecard

| Subsystem                        | Core fidelity | Breadth | Most significant issue                                                                                   |
| -------------------------------- | ------------- | ------- | -------------------------------------------------------------------------------------------------------- |
| Parity test harness              | ★★★★★         | ★★★★☆   | Diffs committed files, not the live render; 49 cases skip live-regression in default CI                  |
| Transforms & geometry            | ★★★★☆         | ★★★☆☆   | Closed transform type set; no `ScaledTranslation`/`TransformWrapper`; diagonal-only separable extraction |
| Ticks / locators                 | ★★★★☆         | ★★★★☆   | `MaxNLocator`/`ScalarFormatter` faithful; **`LogLocator` has an off-by-a-decade stride bug**             |
| Color / norms / cmaps            | ★★★★★         | ★★★★☆   | All six norms faithful; **unknown colormap → viridis silently**                                          |
| MathText & text layout           | ★★★★★         | ★★★★☆   | 632/632 symbols, real box/glue engine; `cm`/`stix` fontsets only remap families over the DejaVu table    |
| Triangulation                    | ★★★★★         | ★★★★★   | Real exact-predicate Qhull port + byte-identical HCT interpolator                                        |
| 1D/2D plotting artists           | ★★★★☆         | ★★★★☆   | 2D `bar(yerr=)` and `hist(log=)` unimplemented                                                           |
| Contour                          | ★★★★☆         | ★★★★☆   | extend/hatches/negative_linestyles faithful; **saddle disambiguation weak**                              |
| Figure / Axes / layout           | ★★★★☆         | ★★★☆☆   | No shared-margin solver for nested/mosaic constrained_layout                                             |
| mplot3d                          | ★★★☆☆         | ★★★★☆   | No true depth-buffered pipeline; `PlotSurface` is a misnamed line-strip                                  |
| AGG backend                      | ★★★★★         | ★★★★★   | The reference backend — fully native, FreeType text, every capability real                               |
| Vector backends (svg/pdf/ps/pgf) | ★★★★☆         | ★★★★☆   | Real shared text shaper; gradient-filled collection items silently dropped                               |
| gobasic (default pure-Go)        | ★★★☆☆         | ★★★☆☆   | Round/bevel joins fake; no gradients; drops most `Paint` state                                           |
| Skia backend                     | ★★☆☆☆         | —       | Default build is a no-op stub; "GPU" is self-admitted CPU-readback scaffolding                           |
| rcParams / styling               | ★★★☆☆         | ★★☆☆☆   | ~122/309 parsed, ~52 of those parsed-then-ignored                                                        |
| Public API / pyplot              | ★★★★★         | ★★★★☆   | Broad, matplotlib-shaped, ~157 pyplot wrappers                                                           |

---

## 2. Where the port is genuinely faithful

These were verified by reading the Go and the matplotlib source side by side.

- **Parity harness is real, not circular.** `TestReferenceCompare` (`test/reference_compare_test.go:25`) diffs Go output against committed **real matplotlib 3.10.9 / FreeType 2.6.1** references over the full, unmasked image (`test/imagecmp/imagecmp.go:27`), failing the CI build on all 178 cases. The self-referential `TestGolden` snapshot is a separate regression layer, not the fidelity proof. Versions are pinned and asserted (`test/helpers_test.go:217`).
- **MathText** (external module `github.com/cwbudde/mathtext@v0.4.4`): **632/632 symbols** ported verbatim from `_mathtext_data.py` (`tex_tables.go:10`); a real hlist/vlist/Hrule/Glue layout engine reproducing `_genfrac` (`frac.go:43`), `subsuper` with the DejaVu `FontConstantsBase` constants (`script.go:5`), `AutoHeightChar` radicals (`sqrt.go`), big-operator limit stacking (`script.go:145`), and matrices (`matrix.go:118`).
- **Per-glyph font fallback** is real — `render/text_fallback.go:88` walks families → generic sans/serif/mono → `STIXGeneral` math fallback rune-by-rune, with tofu only as the last resort.
- **usetex** is a working `latex`+`dvipng` pipeline (`internal/tex/manager.go:133,170`), wired into agg/svg/pdf.
- **Triangulation** (`tri/`): a 21-line shim into the external `qhull-go` module — a genuine exact-predicate engine (`orient2d`/`inCircle` over `math/big.Rat`) that re-ports Qhull 8.0.2's incremental-hull vertex order. `TrapezoidMapTriFinder` and the HCT `CubicTriInterpolator` matrices are byte-identical to `_triinterpolate.py:600`.
- **Color norms** (`core/norm.go`): Normalize, LogNorm, SymLogNorm, PowerNorm, BoundaryNorm faithful forward+inverse; 87 colormaps stored as real 256-entry LUTs (`color/colormap.go:93`). `ToRGBA` faithfully mirrors `to_rgba` incl. `#hex`, tuples, `Cn` cycle refs, grayscale strings; named tables exact (css4=148, xkcd=949).
- **ScalarFormatter** (`core/tick_formatters.go:154`): full additive-offset + ×10ⁿ scientific notation — `_compute_offset`, `_set_order_of_magnitude`, banker's rounding, `get_offset`. **Not a stub.**
- **MaxNLocator** (`core/tick_locators.go`): near line-by-line — `scale_range`, `_Edge_integer`, `_raw_ticks`, `nonsingular`.
- **Bezier toolkit** (`geom/bezier.go`) and **path simplifier** (`backends/agg/agg_path_simplify.go`, a real `PathSimplifier` port gated ≥128 verts) are faithful.
- **Cycler** (`cycler/cycler.go`): full multi-key composition (`New`/`Concat`/`Multiply`/`ByKey`), not color-only.
- **Style sheets**: 28 real `.mplstyle` ports in `style/stylelib/` (vs mpl's 29), copied near-verbatim.
- **Animation writers**: real GIF/APNG/HTML/ffmpeg encoders (`animation/writer_*.go`); unsupported formats return `ErrWriterUnsupported` loudly.
- **Examples are a representative showcase**, not a curated cover-up: `imshow/main.go:33` uses interpolation+extent+aspect together; `contour_styles/example.go:96` uses `extend="both"` with cycled hatches; `stat_variants` uses step-filled and stacked histograms; `vector_fields` drives quiver+barbs+streamplot+quiverkey — on realistic (PCG-normal, dipole) data.

---

## 3. Located workarounds and hacks

Severity key: **CRITICAL** = advertises/pretends to work but doesn't · **SILENT** = degrades silently without telling the caller · **BUG** = simply diverges · **MINOR** = honest, documented limitation.

### 3.1 Secondary backends (the densest cluster)

- **CRITICAL — `backends/skia/`**: The default (untagged) Skia backend is a pure stub — `New` returns an error and `MeasureText` returns `{}` (`skia_stub.go:16,65`). Under `-tags skia` it registers the **full native capability set** while running every method on an embedded `*gobasic.Renderer`. Real native Skia exists only under `-tags "skia skiacgo"`, and even then it is CPU raster.
- **CRITICAL — Skia "GPU" is scaffolding.** `GPU()` returns `r.useGPU` (true under `-tags skiagpu`) but the surface is always the CPU readback bridge; `gpu_enabled.go` says so explicitly, and `SkSurface::MakeRenderTarget` is `StatusDeferred` (`backends/skia/strategy.go:142`). `FlushGPU` is a no-op TODO (`skia.go:393`); `GetSurface()` always returns `nil` (`skia.go:380`).
- **CRITICAL — `backends/registry.go:98`**: `capabilityRuntimeChecks` / `VerifyRendererCapabilities` test _interface presence_ (type assertion), not behavior — so the matrix reports "✓ native" for all the fakes above. The advertising layer cannot catch any of these.
- **CRITICAL — `backends/gobasic/stroke.go:316,340`**: `JoinRound` draws no arc (reuses the miter point); `JoinBevel` is an empty case. The default pure-Go renderer mis-strokes joins.
- **SILENT — `backends/gobasic`**: `pathDevice` copies ~7 of ~27 `Paint` fields, dropping `CompositeMode`, `Alpha`, `FillPattern`, `FillGradient`, `Antialias`, `Snap`. No gradients (capability not declared); images always nearest-neighbor.
- **SILENT — `backends/pdf/pdf_write.go:390` & `backends/ps/paths.go:186`**: gradient/pattern-filled markers and path-collection items are silently dropped (the item `continue`s undrawn).
- **SILENT — `backends/pdf/pdf_write.go:282,451`**: `subsetFontName` prepends an `ABCDEF+` subset tag but embeds the whole font; every embedded font gets hardcoded `/Flags 32 /ItalicAngle 0 /StemV 0` (StemV 0 is invalid).
- **SILENT — `backends/svg/path.go:127`**: `normalizePathEffectFilter` collapses `"shadow"` → `"blur"` (no offset/color). `backends/svg/tex.go:26`: `DrawTeX` rasterizes mathtext to an embedded PNG data-URI instead of vector glyphs.
- **SILENT — `backends/pgf/text.go:70`**: `DrawTextWithFont` discards the `fontKey`; every label renders in the LaTeX default font.
- **SILENT — `backends/webagg/protocol.go:26`**: `MsgRubberband` and `MsgHistoryButtons` are consumed by the JS client but never emitted by the Go server — zoom-box and history buttons never work.
- **SILENT (systemic) — every backend's `GlyphRun`** reinterprets a glyph _ID_ as a Unicode rune (`gobasic/text.go:24`, `agg/agg_text.go:104`, plus pdf/svg/pgf). Correct only when IDs equal codepoints.
- **MINOR (doc wrong) — `backends/desktop/gio/doc.go:4`** claims importing the package is a no-op and `New` returns `ErrNotImplemented`, but `gio.go` is a fully implemented Gio backend. The `backends/desktop` _core_ package, by contrast, genuinely returns `ErrNotImplemented` (`desktop.go:110`).

**Correction to a prior accusation:** the vector backends do **not** ship crude `0.5·size·len` text-metric stubs anymore. svg/pdf/ps/pgf delegate `MeasureText` to a shared pure-Go sfnt shaper (`render/text_metrics.go:26`, `svg/text.go:57`); the crude `0.6·size·len` heuristic survives only as a font-load-failure fallback (`render/text_metrics.go:35`).

### 3.2 Silent degradations in `core/`

- **SILENT — `color/colormap.go:390`**: `GetColormap` returns **viridis** (with only a `diag.Warnf`) for any unknown name, and folds names case-insensitively (`"Blues"=="blues"`), where matplotlib raises and is case-sensitive. A typo renders a plausible-but-wrong plot.
- **SILENT — `core/introspection.go:75`**: `Setp` silently drops unknown property keys; matplotlib raises `AttributeError`.
- **SILENT — `core/picker_contains.go:271`**: `Text.Contains` fakes the glyph bbox from `FontSize × rune-count`; interactive picks on rotated/proportional text are wrong.
- **SILENT — `core/image.go:48`**: a final fallback ignores rotation entirely on backends implementing neither `rasterizeTransformed` nor `ImageTransformer` (benign on AGG, which implements both).
- **SILENT — `canvas/wasm/scaled_renderer.go:19`**: the HiDPI wrapper (used when `pixelRatio != 1`) forwards only a few capability interfaces; on HiDPI displays the native marker/collection/quadmesh/gouraud fast paths and PNG export silently disappear into fallback rendering.
- **SILENT — `animation/animation.go:578`**: the Blit "fast path" restores the background then calls a full `cnv.Draw()` anyway — zero blit benefit (output is correct).

### 3.3 Concrete correctness bugs (divergent, not fake)

- **BUG — `core/tick_locators.go:683`**: `LogLocator` default stride uses `ceil(numDecades/numTicks)` instead of mpl's `numdec//numticks + 1` — these disagree (e.g. 18 decades / 9 ticks: mpl→3, Go→2), so dense log axes get a different decade stride.
- **BUG/weak — `core/contour_lines.go:215`**: saddle disambiguation keys off `above[0]` with fixed corner order, never computes the cell-center mean that mpl2014 uses, and emits all four crossings as one connected polyline. The least-faithful part of the contour port.
- **APPROXIMATE — TwoSlopeNorm** maps out-of-range to finite extrapolation instead of mpl's ±inf (`core/norm.go`); log/logit clip semantics and date nice-snapping (14-day→`[1,15]`) also diverge.

### 3.4 Honest, documented limitations (to be fair)

- `core/axes3d.go:391` `PlotSurface` draws a line strip and emits a one-shot `diag.Warnf` pointing at the real `Surface()`; same for `Voxel` vs `Voxels`. The whole 3D family is documented (`axes3d.go:31`) as a 2D artist model with pre-projected coordinates — there is no depth-buffered painter pipeline (`mplot3d_gallery` ≈ 22 RMSE).
- `render/render.go:495` `NullRenderer` and the ~50 capability interfaces in `render/extensions.go` are honest contracts, not pretend-fakes.
- Throughout `core/` and `pyplot/`, unknown-mode `switch default` cases return explicit `unsupported …` errors rather than swallowing input (e.g. `pyplot/axes_wrappers.go:71`).

---

## 4. Structural weakness: rcParams "parse-then-ignore"

The most user-hostile pattern in the codebase. Of matplotlib's ~309 dotted rcParams, **~122 are parsed** (`style/mplstyle.go` switch → `style/style.go` `RC` struct) and **only ~70 are actually honored** by drawing/layout/save code. The remaining **~52 are parsed, stored, and never read**:

- All 10 `date.*` (`style/mplstyle.go:811`), all 11 `animation.*` (`:905`, struct comment: _"stored only; not yet consumed"_), all `pdf.*`/`ps.*`/`svg.*` backend params (`:839–903`), 6 of 8 `image.*` (`:635–665`), 9 of 10 `mathtext.*` (`:677–701`).

Setting one of these gives **false confidence that it took effect**. To the project's credit the structs self-document the gap (`"Stored only"`), so it is known rather than hidden — but parse-then-ignore is arguably worse than not parsing, because the user has no signal. The larger remainder (309 − 122) is simply not parsed: `lines.dash_capstyle`, `markers.fillstyle`, `axes.spines.*`, `xtick.direction`, `figure.subplot.*`, `font.weight/stretch`, all polar params.

---

## 5. Coverage gaps in plotting artists

Verified absent (not papered over by examples):

- **2D `bar(yerr=/xerr=)`** — `BarOptions` (`core/plot.go:508`) has no error field; only 3D `ErrorBar3D` (`core/axes3d.go:479`) has asymmetric error.
- **`hist(log=)`** — no `Log` field in `HistOptions`.
- **mathtext `cm`/`stix` fontsets** — `core/mathtext.go:208` only remaps the font _family_ over the single DejaVu Unicode table; matplotlib's `BakomaFonts`/`StixFonts` per-fontset glyph maps are not ported, so only the DejaVu default (what every parity reference uses) is parity-exact.
- **~3 accents** missing vs matplotlib's 20-entry `_accent_map`.
- **Transform type system** — closed `switch` over built-in types; missing `ScaledTranslation`, `TransformWrapper`, `TransformedBbox`, and the `Affine2D` fluent `rotate/skew` builders; separable→affine extraction is diagonal-only.

Implemented but undemonstrated by any example: `boxplot(notch=)` (`core/boxplot.go:40`), errorbar asymmetric `XErrLower/Upper` (`core/errorbar.go:16`).

---

## 6. Parity-harness caveats

The gate is genuine but has two holes worth closing:

1. **`TestReferenceCompare` asserts on two committed files** (golden vs ref, `test/helpers_test.go:406`) — it renders `got` but never compares it. Live-render correctness is enforced only by `TestGolden`. The chain _live ≈ golden_ + _golden ≈ mpl_ ⇒ _live ≈ mpl_ is valid **except** for the 49 optional-visual cases (mostly 3D/geo/gallery) where `TestGolden` is skipped in default CI — there, a live-render regression stays green.
2. **Tolerances are honest but loose on big composite cases** (worst ≈ `MaxRMSE 5.0` ≈ 2% RMS). The eye-catching `MinPSNR 10 / MaxMeanAbs 95` overrides on gallery cases are mathematically redundant — the always-present `MaxRMSE` binds first — so no gate degenerates to a no-op, but the loose PSNR/MeanAbs numbers are misleading noise.

---

## 7. Prioritized recommendations

**Stop the silent footguns (cheap, high value):**

1. `color/colormap.go:390` — raise on unknown colormap name; make lookup case-sensitive.
2. `core/introspection.go:75` — make `Setp` error on unknown keys.
3. `core/tick_locators.go:683` — fix the `LogLocator` stride to `numdec//numticks + 1`.
4. Vector backends — error (or visibly warn) when dropping gradient-filled collection items rather than `continue`-ing silently.

**Stop advertising what isn't there:** 5. `backends/registry.go:98` — make capability verification behavioral, not interface-presence, OR have stub renderers (default Skia) decline the native capabilities. 6. Relabel the Skia "GPU" mode honestly (it is CPU readback) and gate `GPU()`/`FlushGPU` accordingly. 7. Fix the `backends/desktop/gio/doc.go` doc that states the opposite of reality.

**Close the structural gaps (larger):** 8. Either honor or stop parsing the ~52 dead rcParams; at minimum emit a one-shot warn when an ignored param is set. 9. Improve contour saddle handling to use the cell-center mean (`core/contour_lines.go`). 10. Implement 2D `bar(yerr=)` and `hist(log=)`.

---

## 8. One-paragraph verdict

This is a serious, high-fidelity port, not a facade. The numerically hard cores — transforms, locators, norms, mathtext, triangulation, FreeType text — are real, often line-for-line, and the parity claim is honestly measured against genuine matplotlib output and enforced in CI. The "facade" instinct is nonetheless correct in a small, well-defined ring: the Skia backend and its GPU mode are scaffolding, the capability registry rubber-stamps backend stubs by checking interfaces rather than behavior, the default pure-Go renderer fakes stroke joins, and a third of the parsed rcParams are silently ignored. The single most dangerous class is the silent degradation — unknown-colormap-to-viridis, `Setp` key-dropping, dropped gradient items — because these render plausible-but-wrong output for someone porting a real matplotlib script. Tighten those into loud errors and honestly label the backend scaffolding, and the gap between the project's parity claims and its reality closes substantially.

---

# Third Fidelity & Quality Review (2026-07-01)

**Date:** 2026-07-01
**Reference:** matplotlib 3.10.9 (`third_party/matplotlib`), FreeType 2.6.1, DejaVu Sans 2.35
**Method:** Three parallel subagent audits — (1) defaults vs `matplotlibrc`/`rcsetup.py` + test/showcase coverage + tolerance analysis, (2) confession-marker hunt + side-by-side algorithm diffs across 12 subsystems, (3) public-API idiomaticity + code-organization critique. Findings already tracked by the 2026-06-30 review (Phases 14–18) were excluded up front. Every claim is anchored to `file:line`; the headline claims (linewidth default, colormap race, hist auto-binning, `io.Writer` absence) were independently re-verified against source.

**Verdict:** The rendering core keeps its five-star fidelity — this round found no new "facade" ring. What it found instead is (a) a small set of **wrong or dead defaults** on the most-used entry points, (b) a handful of **new algorithmic divergences** in autoscale/quiver/hist/dates/pie, (c) large rcParams families that are **not even parsed**, and (d) a public API that is a fairly literal **Python transliteration rather than idiomatic Go** — warn-and-continue instead of errors, pointer-optionals, string enums, zero `io.Writer`, and a 60k-line `core/` god-package holding half the frozen surface. Resolution is tracked as PLAN.md fold-ins to Phases 16–18 plus new Phases 19–21.

## 1. Default-value mismatches (Phase 19)

- **HIGH — `lines.linewidth` is both wrong and dead.** `style/style.go:333` defaults `RC.LineWidth` to **1.25** (matplotlib: **1.5**, `matplotlibrc:116`), and `core/plot.go:87` hardcodes `lineWidth := 1.5` without ever reading `RC.LineWidth` — the only consumer is the scatter edge-width fallback (`core/scatter.go:415`). Net: `lines.linewidth` in an `.mplstyle` is a no-op for line plots.
- **MED — hist default bins.** `core/histogram.go:285` + `core/plot.go:929` default to auto-selection (Sturges for n<1000, else Scott); matplotlib defaults to a fixed **10** (`hist.bins`, `_axes.py:7033`). Every default histogram has a different bar count.
- **MED — scatter default size renders invisible.** `Scatter2D.Size` zero-value is 0 and `scatterAreaScale` returns 0 for `area<=0` (`core/scatter.go:802-810`); matplotlib defaults `s = lines.markersize² = 36` pt².
- **LOW — minor ticks.** Size `3.5×0.6 = 2.1` vs mpl `xtick.minor.size: 2` (`core/axis_ticks.go:73`); minor pad not distinguished from major (both 3.5 vs mpl 3.4/3.5, `core/axis_types.go:23`); tick-label pad DPI fallback uses 96 instead of 100 (`core/axis_ticklabels.go:203`).
- **MED gap — `plot()` has no linestyle field.** `PlotOptions` exposes only `Dashes []float64` (`core/plot.go:21`); `lineStyleToDashes` is reachable only via the prop cycle, so the most common mpl idiom (`linestyle="--"`) has no direct spelling.
- **Verified correct (regression documentation):** DPI 100, figsize 6.4×4.8, font.size 10/title 12, major tick geometry (3.5/0.8/0.6/pad 3.5/direction out), grid `#b0b0b0` lw 0.8, `axes.axisbelow="line"` z=1.5, legend framealpha 0.8/frameon/edgecolor 0.8, `errorbar.capsize 0`, `axes.unicode_minus` (U+2212 conversion at `core/tick_formatters.go:447`), and the full dash-pattern/cap/join set — dashed `[3.7,1.6]`, dashdot `[6.4,1.6,1,1.6]`, dotted `[1,1.65]`, width-scaled, solid cap projecting / dashed cap butt, join round (`core/contour_styles.go:26`, `core/line.go:339`).

## 2. New algorithmic divergences (Phase 17 fold-in)

- **BUG — quiver default scale.** `core/vector_field_quiver.go:346-353` uses `scale = mean/(0.18·min(W,H)/√N)`; matplotlib uses `1.8·amean·sn/span` with `sn = clip(√N, 8, 25)` (`quiver.py:681`). Default arrow lengths do not match.
- **BUG — autoscale margins in data space.** `core/axes_autoscale.go:92-93` pads `span·margin` in data coordinates; matplotlib pads in transform space and inverse-maps (`axes/_base.py:3064-3070`) — wrong padding on every log/symlog axis. Also: non-positive limits are not dropped before log autoscale (`_base.py:3017`), and zero-span expansion uses `span=1` + linear margin instead of `nonsingular(expander=0.05·|v|)`.
- **BUG — zero-bbox artists dropped from autoscale.** `core/axes_autoscale.go:51` skips artists whose bounds are exactly `{0,0,0,0}` — a single point at the origin is ignored. matplotlib has no such exclusion; the sentinel needs an explicit has-data flag.
- **MISSING — spine positions.** Only boundary + data modes exist (`core/axis_types.go:57-62`); matplotlib's `set_position(('outward', pts))` and `(('axes', frac))` (`spines.py`) — the standard detached/centered-spine idioms — are absent.
- **BUG — AutoDateLocator DAILY table.** `{1,2,4,7,14}` (`core/date_tick.go:648`) vs matplotlib `{1,2,3,7,14,21}`; the broader rrule alignment stays a simplified `align()` (documented divergence).
- **BUG — pie framing.** Axis window is radius-scaled (`padding := Radius*1.25`, `core/pie.go:245`); matplotlib fixes it at `±1.25 + center` regardless of radius.
- **Hist internals (with the Phase 19 default fix):** 'auto' should be numpy's `min(fd, sturges)` bin width, not an n<1000 Sturges/Scott switch (`core/histogram.go:285-289`); FD IQR uses nearest-rank instead of interpolated percentiles (`:375`); Scott uses ddof=1 instead of numpy's ddof=0 (`:352-367`).
- **APPROXIMATE — constrained_layout edges.** Nested-mosaic grids are not modeled (`core/layoutgrid.go:239`) and outside legend/colorbar space is reserved approximately (`:448`).
- **Doc rot:** `core/axes3d_contour.go:10` still claims a "placeholder wireframe contour"; the code beneath runs a real projected contour.
- **Verified faithful (no action):** legend `loc="best"` (full `_find_best_position` candidate loop + 4-term badness), violin KDE (Scott/Silverman exact), imshow (all 18 AGG filters + the `antialiased` auto-select), `ArrowStyle`/`ConnectionStyle` registries (complete), the ~40-marker table + fillstyles (geometric-clip mechanism, equivalent output), pie defaults, hexbin (gridsize/extent/mincnt/marginals/reduce_C).

## 3. rcParams not parsed at all (Phase 16 fold-in)

Beyond the 51 known parsed-but-unhonored keys (`style/unhonored.go`), whole families never reach the `RC` struct — they land silently in `report.Unsupported` (`style/mplstyle.go:977`), which nothing inspects by default:

- **HIGH:** the entire ticks family (`{x,y}tick.direction`, major/minor `size`/`width`/`pad`, `minor.visible`, side toggles, label sides, `alignment`); the legend layout family (14 params: `loc`, `fancybox`, `shadow`, `numpoints`, `scatterpoints`, `markerscale`, `title_fontsize`, `borderpad`, `labelspacing`, `handlelength`, `handleheight`, `handletextpad`, `borderaxespad`, `columnspacing`); the axes family (`axisbelow`, `titlepad`, `labelpad`, `titlelocation`, `spines.*`, `xmargin`/`ymargin`, `autolimit_mode`, `unicode_minus`, `axes.formatter.*`); lines/markers (`linestyle`, `marker`, `markersize`, `markeredgewidth`, the three `*_pattern` keys, `scale_dashes`, cap/join styles, `markers.fillstyle`, `lines.antialiased`); figure (`edgecolor`, `frameon`, `subplot.*`, `autolayout`, `constrained_layout.*`, `titlesize`/`titleweight`).
- **MED:** patches (no `PatchRC` exists at all); `font.weight`/`style`/`variant`/`stretch` + family lists; the artist-default keys `errorbar.capsize`, `hist.bins`, `scatter.marker`, `scatter.edgecolors`; `contour.*`.
- **LOW:** `path.snap`/`path.effects`, `pcolor.shading`, `text.hinting`/`antialiased`, `polaraxes.grid`, `axes3d.*`, `{x,y}axis.labellocation`, `pgf.*`.

## 4. Zero-fixture public APIs (Phase 18 fold-in)

Implemented on `*Axes` but exercised by no parity case and no example: **`PSD`, `Specgram`, `Cohere`, `CSD`** (HIGH — the whole spectral family with real Welch/detrend/window numerics behind it); `LogLog`/`SemilogX`/`SemilogY`, `SecondaryYAxis`, `TwinY` (MED); `AxLineSlope`, single-series `BoxPlot` (LOW).

## 5. Loose-tolerance / visual-QA candidates (Phase 21)

- **Effectively disabled gates:** `widgets_gallery` and `animation_gallery` run with `MinPSNR 10 / MaxMeanAbs 95` — 95/255 mean absolute difference passes.
- **MaxRMSE ≥ 4 queue (~23 cases):** `skewt_basic` 5.0, `annotation_legend_offsetbox_gallery` 5.0, `text_bbox_styles` 5.0, `text_layout_gallery` 5.0, `mathtext_accents` 5.0, `projection_toolkit_gallery` 4.9, `specialty_depth` 4.8, `formatter_log_mathtext_labels` 4.8, `named_colors` 4.8, `animation_subplots_frame` 4.5, `stem_horizontal` 4.5, `mathtext_fractions` 4.5, `lognorm_imshow` 4.4, `colorbar_variants_gallery` 4.4, `stem_plot` 4.3, `errorbar_capthick` 4.2, `patch_style_matrix` 4.2, `text_annotation_matrix` 4.2, `scatter_gallery` 4.0, `animation_line_frame`/`animation_scatter_frame` 4.0, `mplot3d_stem3d` 4.0, `scale_logit_ticks` 4.0, `mathtext_inline_labels` 4.0.
- **Dense galleries where RMSE alone is a weak gate:** `mathtext_gallery` (PSNR 16 / MeanAbs 35), `image_variants_gallery`, `triangulation_gallery`, `pcolormesh_gouraud`.

## 6. API idiomaticity (Phase 20)

- **Split error convention.** Primary plot methods return the artist only and swallow invalid input via `diag.Warnf` (16 sites; `core/plot.go:263,755,821,896,1016`) — no programmatic failure detection — while the `*Units` variants return `(T, error)`. Two conventions for the same operations.
- **Options ergonomics.** 83 `…Options` structs with **408 pointer-to-primitive optional fields** (callers declare locals just to take addresses); variadic `opts ...FooOptions` consumed first-only (a second struct is silently dropped, `core/plot.go:73`); string-typed enums (`Orientation`, `LineStyle`, `Location`, `Where`, `Colormap`, and `Interpolation *string` at `core/image.go:68`) next to the existing typed-constant pattern (`SignalSpectrumScale`).
- **Zero `io.Writer` in the 3,019-symbol surface.** Every save takes a file path; there is no `Figure.Save`/`WriteTo`/`Image()`, so every example hand-rolls the 8-line `agg.New` + `DrawFigure` + `GetImage` dance with `panic` on error.
- **Python-ism naming.** `Setp`/`Getp`/`GetpAll`/`Findobj` (`core/introspection.go`); `GetX()` getters beside idiomatic `X()` siblings; exported mutable fields (`fig.SizePx.X =`) alongside setter methods.
- **DATA RACE.** `color/colormap.go:388` `RegisterColormap` writes the package-global `colormaps` map with no mutex while draws read it concurrently. Figure/Axes concurrency is otherwise undocumented (one concurrency comment in all of `core/`).
- pyplot's global current-figure registry is mutex-guarded (fine), but its errors are discarded in examples (`_ = pyplot.SCA(...)`).

## 7. Code organization (Phase 20)

- **`core/` god-package:** 60,369 lines / 173 files / **1,529 exported symbols = 51% of the frozen public API**, bundling figure/axes/artists, the full 3D subsystem (98 symbols), locators/formatters/date ticks (~3,162 lines), units, the mathtext bridge, interactive widgets/selectors (~2.5k lines), contour, colorbar, legend, and layout. Split candidates: `plot3d`, `ticker`, `widgets` — all break import paths, so they must precede the v1.0 freeze.
- **Duplication:** alpha-baking inlined ~62×, the `opts[0]` unpack dozens×, scalar-map resolution duplicated between `Scatter2D` and `Collection`.
- **Stale audit doc:** `docs/large-file-decomposition.md` records `plot.go` 256 lines and `mplstyle.go` 724 lines smaller than reality; 24 non-test files exceed 800 lines, several undocumented (`tick_locators.go` 1149, `render/extensions.go` 1074, `tick_formatters.go` 1045, `line.go` 1028, `date_tick.go` 968, `signal_helpers.go` 954, `boxplot.go` 907, …).
- **Frozen surface too large:** 3,019 symbols is a heavy v1.0 stability promise; tiering candidates are the introspection cluster, the `*Units` any-variants, and backend-only render extensions.

## 8. Deliberately not acted on

- The verified-faithful subsystems above (kept as regression documentation).
- A full rrule-based `AutoDateLocator` port (low parity yield; the DAILY-table fix suffices).
- Nested-mosaic constrained_layout (Phase 17 triage item; expected outcome "documented limitation").
- 309-param rcParams parse completeness (Phase 16 parses by parity impact; `path.snap`, `text.hinting`, `polaraxes.grid`, `axes3d.*` become documented non-goals).
- The mplot3d depth-buffered pipeline (pre-existing documented limitation; unchanged).
- pyplot examples discarding errors (dissolves into Phase 20's error-convention change).
