# Milestone 1 — Core skeleton (Figure/Axes/Artist + Transforms)

**Scope**

- Packages: `core` (figure, axes, artist), `transform`, `style`, `color`.
- Public interfaces: `Artist`, `Renderer`, `Figure`, `Axes`.
- Basic scales: linear/log; axis ticks (locators/formatters v0).

**Tasks**

- Define `Renderer` verbs: `Begin/End`, `Path`, `Stroke`, `Fill`, `Image`, `GlyphRun`, `Clip`, `MeasureText`.
- Data→axes→figure→pixel transforms; z-order & clipping.
- Minimal rc-style config (functional options).

**Exit criteria**

- Compile-time interfaces stable.
- Unit tests: transform invertibility; tick locator monotonicity.
- A headless “null renderer” can traverse and no-op draw a simple figure.

**Detailed plan**

## Phase A — Repo & package scaffold

Clean layout, CI, linting, and a place for examples/tests.

Layout: `core`, `transform`, `render`, `style`, `color`, `internal/geom`, `examples`, `test`.

Done

- [x] Module initialized (current path: `matplotlib-go`).
- [x] Tooling: Justfile/Makefile (`fmt`, `lint`, `lint-fix`, `build`, `test`, `cli`).
- [x] CI: `golangci-lint`, treefmt, `go vet`; jobs `test-lint.yml`, `test-fmt.yml`, `test-unit.yml`, orchestrated by `test.yml`.
- [x] Matrix: Go 1.22–1.24; Linux/macOS.
- [x] Repo meta: CODEOWNERS, CONTRIBUTING.md, LICENSE (content TBD).

---

## Phase B — Geometry & types (primitives)

Immutable-ish value types shared by all packages.

**Checklist**

- [x] Rect/Point utils with tests
- [x] Affine math + invert tests
- [x] Path structure with size/cap sanity checks

---

## Phase C — Renderer interface (+ null renderer)

Stable renderer verbs and a no-op backend.

**Checklist**

- [x] Define `Renderer` + `NullRenderer` with compile-time assert
- [x] No-op state stack & clipping stack
- [x] Unit test: calling verbs doesn’t panic, maintains Save/Restore balance

---

# Phase D — Transform subsystem

Data→axes→pixel transform chain with invertibility guarantees.

**Checklist**

- [x] Linear/Log implementations with domain checks
- [x] Compose affine + scale into a 2D transform: (x,y data) → (u,v axes) → pixels
- [x] Property tests: `p == inv(apply(p))` within ε for random points in domain
- [x] Edge cases: degenerate domain (min≈max), log with base≈1, values near 0

---

## Phase E — Axis ticks v0 (locators & formatters)

Done

- [x] Implemented `LinearLocator` with {1,2,5}×10^k steps covering [min,max]
- [x] Implemented `LogLocator` (majors at Base^k; optional minors at {2,5})
- [x] Formatters: `ScalarFormatter` (trim zeros; scientific for extremes), `LogFormatter` (10^k as 1eN/2eN/5eN)
- [x] Tests: monotonicity/coverage across ranges; log majors (bases 2,10) and minors containment; formatter behavior

Refs: `core/tick.go`, `core/tick_test.go`

---

## Phase F — Style (rc-like) & options

Done

- [x] RC defaults defined with colors, DPI, font, ticks
- [x] Options (`WithDPI/WithFont/...`) and `style.Apply` helper
- [x] Figure/Axes use options with inheritance (axes override figure)
- [x] Tests for defaults, last-wins order, and precedence

Refs: `style/style.go`, `style/style_test.go`, `core/artist.go`

---

## Phase G — Core: Artist, Figure, Axes (structure only)

**Goals:** Matplotlib-like Artist tree, traversal, and z-order.

```go
// core/artist.go
package core

import (
  "sort"
  "github.com/meko-christian/mathplotlib-go/internal/geom"
  "github.com/meko-christian/mathplotlib-go/render"
)

type Artist interface {
    Draw(r render.Renderer, ctx *DrawContext)
    Z() float64                      // for z-order sorting
    Bounds(ctx *DrawContext) geom.Rect // optional rough bbox
}

type DrawContext struct {
    // transforms
    DataToPixel Transform2D
    // styling
    RC      style.RC
    // clip (axes rect in pixels)
    Clip    geom.Rect
}

// Transform2D wires x/y scales + axes->pixel affine
type Transform2D struct {
    XScale transform.Scale
    YScale transform.Scale
    AxesToPixel transform.AffineT
}

type Figure struct {
    SizePx geom.Pt
    RC     style.RC
    Children []*Axes
}

type Axes struct {
    RectFraction geom.Rect // [0..1] in fig coords
    RC           *style.RC // nil => inherit
    XScale transform.Scale
    YScale transform.Scale
    Artists []Artist
    zsorted bool
}

func NewFigure(w, h int, opts ...style.Option) *Figure
func (f *Figure) AddAxes(r geom.Rect, opts ...style.Option) *Axes
func (a *Axes) Add(art Artist)
func (a *Axes) layout(f *Figure) (pixelRect geom.Rect) // for context
```

**Traversal (v0)**

- Figure begins renderer with viewport = full pixel rect.
- For each Axes: compute pixel rect, set clip, build `DrawContext`, sort artists by `Z()`, call `Draw`.
- Axes draws spines/grid/ticks later (in later milestones) — for now focus on structure.

**Checklist**

- [x] Artist interface + adapter (`ArtistFunc`)
- [x] Figure/Axes construction & artist registration
- [x] Z-order stable sort by Z then insertion order
- [x] Inheritance of RC: Figure → Axes (artists can ignore for now)

Refs: `core/artist.go`, `core/artist_test.go`

---

## Phase H — Clipping & z-order semantics (baseline)

**Goals:** predictable rules from day one.

**Rules**

- Each Axes sets a **single rectangular clip** (its pixel rect) before drawing children.
- Artists **may** push additional clip paths/rects (not needed in M1).
- Z-order: float64 where larger Z draws later. Default Z=0. Axes background uses Z=-∞ (implicit).

**Checklist**

- [ ] `Axes` calls `r.Save(); r.ClipRect(axRect); ...; r.Restore()`
- [ ] Unit test: Artists with Z=-1,0,1 are visited in that order
- [ ] NullRenderer records call sequence in a debug buffer (build tag `debug`) for tests

---

## Phase I — Example no-op traversal (compiles, runs)

**Goals:** one tiny artist that does nothing but prove traversal.

```go
// examples/noop/main.go
fig := core.NewFigure(800, 600)
ax := fig.AddAxes(geom.Rect{Min: geom.Pt{.1,.1}, Max: geom.Pt{.9,.9}})
ax.Add(core.ArtistFunc(func(r render.Renderer, ctx *core.DrawContext) {
    // nothing yet
}))
r := &render.NullRenderer{}
_ = r.Begin(geom.Rect{Min: geom.Pt{0,0}, Max: geom.Pt{800,600}})
defer r.End()
core.DrawFigure(fig, r) // helper that performs traversal
```

_(Implement `ArtistFunc` as a small adapter type for tests/examples.)_

**Checklist**

- [x] Buildable example under `examples/noop`
- [x] `core.DrawFigure` exists and calls into traversal

Refs: `examples/noop/main.go`, `core/artist.go`

---

## Phase J — Tests (what “done” means for M1)

**Transform invertibility**

- [ ] Property test (testing/quick): random affine matrices with invertible determinant × random points → `Invert(Apply(p)) ≈ p` (ε=1e-9)
- [ ] Linear scale: random domains (including negative), 1e6 magnitude spread → round-trip within ε
- [ ] Log scale: random domains with min>0, various bases (2, e≈2.718281828, 10) → round-trip within ε; reject non-positive input

**Tick monotonicity**

- [ ] LinearLocator: for ranges {\[-1,1], \[0,1e-9], \[1, 1e6], \[-1e6,-1], \[2,2]} and targetCounts {3,5,7} → strictly increasing; first≤min; last≥max
- [ ] LogLocator: base∈{2,10}; domains like \[1, 1e6] → majors monotone; optional minors in-between
- [ ] Formatter: no trailing zeros; `Format(1.0)` → `"1"`; large/small use scientific if |x|≥1e6 or |x|≤1e-4

**Core traversal**

- [ ] Visit order obeys z-order and insertion stability
- [ ] Clip stack balanced (Save/Restore count matches)
- [ ] NullRenderer `Begin`/`End` must be called exactly once each

**Compile-time interface stability**

- [ ] `var _ render.Renderer = (*render.NullRenderer)(nil)`
- [ ] `var _ core.Artist = (*testArtist)(nil)`

---

## Phase K — Polishing & docs

**Goals:** lock API names, document package boundaries.

**Checklist**

- [ ] Package docs (`doc.go`) for `core`, `render`, `transform`, `style`
- [ ] README section: “Milestone 1 status” + minimal example
- [ ] Changelog entry capturing stable interfaces
- [ ] Tag `v0.1.0-m1` (pre-release) once tests are green

---

### Reference snippets you can drop in

#### 1) ArtistFunc adapter (testing convenience)

```go
package core
type ArtistFunc func(r render.Renderer, ctx *DrawContext)
func (f ArtistFunc) Draw(r render.Renderer, ctx *DrawContext) { f(r, ctx) }
func (f ArtistFunc) Z() float64 { return 0 }
func (f ArtistFunc) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }
```

#### 2) Draw traversal

```go
package core
func DrawFigure(fig *Figure, r render.Renderer) {
    vp := geom.Rect{Min: geom.Pt{0,0}, Max: geom.Pt{fig.SizePx.X, fig.SizePx.Y}}
    _ = r.Begin(vp); defer r.End()
    for _, ax := range fig.Children {
        px := ax.layout(fig)
        r.Save(); r.ClipRect(px)
        // build DrawContext with composed transform
        ctx := &DrawContext{
            DataToPixel: Transform2D{
                XScale: ax.XScale, YScale: ax.YScale,
                AxesToPixel: transform.NewAffine(axesToPixel(px, ax)),
            },
            RC:   ax.effectiveRC(fig),
            Clip: px,
        }
        if !ax.zsorted { sort.SliceStable(ax.Artists, func(i,j int) bool {
            zi, zj := ax.Artists[i].Z(), ax.Artists[j].Z()
            if zi == zj { return i < j }
            return zi < zj
        }); ax.zsorted = true }
        for _, art := range ax.Artists { art.Draw(r, ctx) }
        r.Restore()
    }
}
```

---

# Milestone 2 — AGG backend (PNG) + first plot

## Phase A — Backend plumbing & “first line” artist

**Goals:** finalize the link between `core` and `render.Renderer`, add a tiny `Line2D` artist so we can actually draw something.

**Scope**

- `core/line.go`: minimal polyline artist (stroke only).
- `core/savepng.go`: convenience `SavePNG(fig, r, path)`.

**Key APIs**

```go
// core/line.go
type Line2D struct {
    XY   []geom.Pt            // data space
    W    float64              // stroke width (px for now)
    Col  render.Color
    z    float64
}
func (l *Line2D) Draw(r render.Renderer, ctx *DrawContext) {
    p := geom.Path{}
    for i, v := range l.XY {
        q := ctx.DataToPixel.Apply(v)
        if i == 0 { p.C = append(p.C, geom.MoveTo) } else { p.C = append(p.C, geom.LineTo) }
        p.V = append(p.V, q)
    }
    r.Path(p, render.Paint{LineWidth: l.W, Stroke: l.Col})
}
func (l *Line2D) Z() float64 { return l.z }
func (l *Line2D) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }
```

**Checklist**

- [ ] Add `Line2D` + unit test (doesn’t panic with empty/singleton data).
- [ ] `SavePNG` helper delegates to a renderer that owns a surface and encodes it.

---

## Phase B — “GoBasic” renderer (std/X image first)

**Goals:** get a PNG out using pure Go deps before touching AGG.

**Tech choices**

- Surface: `image.RGBA`.
- Rasterization: `golang.org/x/image/vector` (`vector.Rasterizer`) for fill/stroke.
- Dashes: manual path decomposition (v1: none; v2: simple dash).
- PNG: `image/png`.

**Package**

```
/backends/gobasic   // first, deterministic, no cgo
```

**Mapping `Renderer` → GoBasic**

```go
type Renderer struct {
    dst   *image.RGBA
    vp    geom.Rect
    stack []state
    rast  *vector.Rasterizer
}
func New(w, h int, bg render.Color) *Renderer
func (r *Renderer) Begin(vp geom.Rect) error
func (r *Renderer) End() error
func (r *Renderer) Save()
func (r *Renderer) Restore()
func (r *Renderer) ClipRect(rr geom.Rect)
func (r *Renderer) Path(p geom.Path, paint render.Paint)
func (r *Renderer) Image(img render.Image, dst geom.Rect)
func (r *Renderer) GlyphRun(run render.GlyphRun, color render.Color) // stub: no-op in M2
func (r *Renderer) MeasureText(text string, size float64, fontKey string) render.TextMetrics
```

**Stroke plan (v0)**

- Use `vector.Stroke` with `vector.Path` conversion; map `LineCap`, `LineJoin`, `MiterLimit`.
- If `Fill.A > 0`: rasterize fill (nonzero rule) then stroke.
- No dash in v0; set expectation in tests (dash added in Phase E).

**Checklist**

- [ ] Implement surface + `Begin/End` + state stack and clip rect.
- [ ] Convert `geom.Path` → `vector.Path`.
- [ ] Implement fill + stroke (no dash yet).
- [ ] `SavePNG(fig, gobasic.New(...), path)` renders a diagonal line.

---

## Phase C — PNG encode & example

**Goals:** wire an example that anyone can run; assert dimensions, background.

**Example**

```
/examples/lines/basic.go
```

```go
fig := core.NewFigure(640, 360)
ax  := fig.AddAxes(geom.Rect{Min: pt(.1,.15), Max: pt(.95,.9)})
ax.XScale = transform.NewLinear(0, 10)
ax.YScale = transform.NewLinear(0,  1)
ax.Add(&core.Line2D{
  XY: []geom.Pt{{0,0},{1,.2},{3,.9},{6,.4},{10,.8}},
  W: 2, Col: render.Color{0,0,0,1},
})
r := gobasic.New(640,360, render.Color{1,1,1,1})
core.DrawFigure(fig, r)
core.SavePNG(r, "out.png")
```

**Checklist**

- [ ] `examples/lines/basic.go` builds & produces `out.png`.
- [ ] Background is uniform white; 640×360; no clipping artifacts.

---

## Phase D — Golden image test harness (deterministic)

**Goals:** add snapshot testing that’s stable cross-platform.

**Design**

- Baseline PNGs live under `testdata/golden/...`.
- Compare with **pixel diff** (absolute RGBA tolerance), and compute **PSNR** (to detect non-obvious changes).
- Optional mask to ignore text (not used in M2).

**Helpers (`/test/imagecmp`)**

```go
type DiffResult struct { MaxDiff uint8; MeanAbs float64; PSNR float64 }
func ComparePNG(got, want image.Image, tol uint8) (DiffResult, error)
func LoadPNG(path string) (image.Image, error)
func HashPNG(img image.Image) string // SHA256 on raw RGBA for CI assertions
```

**Go test pattern**

```go
func TestBasicLine_Golden(t *testing.T) {
    img := renderBasicLine() // returns image.Image
    want, _ := imagecmp.LoadPNG("testdata/golden/basic_line.png")
    diff, err := imagecmp.ComparePNG(img, want, 1) // ≤1 LSB tolerance
    if err != nil { t.Fatal(err) }
    if diff.MaxDiff > 1 { t.Fatalf("maxdiff=%d psnr=%.2f", diff.MaxDiff, diff.PSNR) }
}
```

**Checklist**

- [ ] Implement `imagecmp` + unit tests on synthetic images.
- [ ] Add `basic_line.png` baseline.
- [ ] CI job uploads diff image on failure (store in `_artifacts/`).

---

## Phase E — Stroke/fill parity, joins/caps, and dashes

**Goals:** correctness for common stroke styles; minimal dash support.

**Tasks**

- Add `LineJoin` (miter/round/bevel) and `LineCap` (butt/round/square) mapping to `vector.StrokeOptions`.
- Implement **dash**: decompose each subpath into dash segments (user-space lengths).
- Add table-driven tests that render the same path with each style and compare to golden images (`joins_caps.png`, `dashes.png`).

**Checklist**

- [ ] Join & cap styles match expected shapes (visual baselines).
- [ ] Dash phase starts at subpath start; consistent across platforms.
- [ ] Parity test: fill + stroke order produces identical result independent of internal batching.

---

## Phase F — Determinism hardening

**Goals:** lock down sources of cross-platform drift.

**Actions**

- Force **premultiplied alpha** consistently; ensure `vector` inputs premultiplied.
- Disable any non-deterministic parallelism.
- Pin Go toolchain in CI (e.g., 1.22.x) and pin `golang.org/x` commit SHAs.
- Quantize float inputs before rasterization (e.g., snap to 1e-6) to avoid tiny stroke joins differences.

**Checklist**

- [ ] Re-run CI on Linux/macOS → identical hashes for golden images.
- [ ] Document determinism assumptions in `docs/determinism.md`.

---

## Phase G — AGG backend scaffolding (build-tagged)

**Goals:** introduce AGG backend while keeping GoBasic as the default “safe” backend.

**Structure**

```
/backends/agg
    agg.go          // implements render.Renderer
    image.go        // surface wrapping
// build tags if AGG uses cgo: //go:build agg
```

**Mapping `Renderer` → AGG**

- Convert `geom.Path` to AGG path sink.
- Map joins/caps/miter to AGG stroker.
- ClipRect/ClipPath: start with rect only.
- Image: nearest-neighbor blit (bilinear later).
- GlyphRun: stub (no text yet in M2).

**Checklist**

- [ ] `agg.New(width,height,bg)` returns a renderer with in-memory surface.
- [ ] Implement `Path` fill + stroke with parity to GoBasic.
- [ ] PNG export (via AGG or copy RGBA buffer → `image/png`).

---

## Phase H — Cross-backend parity & goldens

**Goals:** ensure AGG and GoBasic produce the **same** outputs within tolerance.

**Tests**

- Reuse all M2 goldens; run twice (GoBasic and AGG) guarded by build tags/env var:

  - Default CI: GoBasic (no cgo).
  - Separate CI job with `-tags=agg` to test AGG.

```bash
go test ./...                          # GoBasic
go test -tags=agg ./...                # AGG
```

**Checklist**

- [ ] Parity: max pixel diff ≤1, PSNR ≥ 50 dB on all goldens.
- [ ] If tiny differences, document per-backend tolerances and freeze goldens per backend (`golden/basic_line.agg.png` if needed).

---

## Phase I — CI wiring & artifacts

**Goals:** make failures actionable.

**CI Steps**

- Job 1: Linux, Go 1.22.x, GoBasic tests + upload diffs if any.
- Job 2: macOS, Go 1.22.x, GoBasic tests.
- Job 3 (optional): Linux with `-tags=agg` if AGG is available in CI image.
- Cache `go` build and `golangci-lint`.

**Artifacts**

- On golden failure, upload: `got.png`, `want.png`, `diff.png`, and a small HTML report with PSNR/MaxDiff.

**Checklist**

- [ ] GitHub Actions workflows committed.
- [ ] Badges in README for build status.

---

## Phase J — Example gallery seed (for humans)

**Goals:** provide 2–3 tiny examples using GoBasic and AGG.

- `examples/lines/styles.go` – caps/joins comparison.
- `examples/lines/dash.go` – simple dash pattern.
- `examples/axes/basic.go` – same as Milestone 1 demo but with visible axes box (if you already have spines; if not, plot only).

**Checklist**

- [ ] Examples compile; `go run` produces PNGs to `examples/out/`.
- [ ] Screenshots added to README.

---

## Tips & pitfalls (quick hits)

- **Coordinate precision:** keep all math in float64 until the rasterizer; avoid early rounding.
- **Clipping first:** always set `ClipRect` (axes pixel rect) before drawing—this stabilizes dash clipping at edges.
- **Dashes & transforms:** compute dash segmentation **in user space** (post-transform) to match visual expectations.
- **Gamma:** both GoBasic and AGG effectively assume sRGB-ish; don’t mix linear/sRGB without a plan. Keep everything in premultiplied sRGB for M2.

---

# Milestone 3 — Text MVP (ASCII) + layout basics

**Scope**

- Single-line Latin text; labels/titles; rotation; alignment.
- Font manager & caching.

**Tasks**

- Integrate freetype for metrics; pick default font set.
- Implement `GlyphRun`, `MeasureText`, renderer-side glyph cache.
- Axes layout: margins, title/label/ticks placement; grid lines.

**Exit criteria**

- Title/xlabel/ylabel render correctly; tick labels don’t overlap for common ranges.
- Font snapshot tests assert metrics consistency across platforms.

---

# Milestone 4 — Core artists v1 (Line/Scatter/Patch/Image)

**Scope**

- `Line2D`, `Scatter` (markers + size/color), `Patch` (Rect/Circle/Poly), `Image` (imshow), grid.
- Colormaps + normalization (linear/log), colorbar (basic).

**Tasks**

- Colormap registry (viridis, plasma, gray).
- Scalar mappers; image sampling & aspect handling.
- Legend (basic): entries for line/scatter/patch.

**Exit criteria**

- Gallery: line, scatter, bar/hist (from patches), imshow + colorbar.
- Golden images added for each gallery plot.

---

# Milestone 5 — SVG export (vector) via recorder

**Scope**

- Recording backend that translates draw ops → SVG.
- Savefig to `.svg`.

**Tasks**

- Implement vector `Path`, text as `<text>` with transforms; clipping.
- Baseline tests comparing structural SVG (XML diffs normalized).

**Exit criteria**

- PNG and SVG snapshots visually match within tolerance (render SVG → raster for compare).

---

# Milestone 6 — Text shaping (HarfBuzz) + font fallbacks

**Scope**

- Complex scripts, ligatures, RTL; font fallback stacks.

**Tasks**

- Bind HarfBuzz; layout → `GlyphRun`s with bidi and shaping.
- Font discovery (system + user) and fallback selection.
- Cache shaped runs (keyed by font+size+text).

**Exit criteria**

- Samples for Arabic/Devanagari/Thai/emoji render correctly (visual baselines).
- API unchanged for simple ASCII users.

---

# Milestone 7 — Layout polish + themes

**Scope**

- Tight layout / constrained layout; style presets.

**Tasks**

- Compute artist bounding boxes; auto-adjust axes to avoid clipping.
- rc-like config loader (file/env); a “nice defaults” theme.

**Exit criteria**

- Gallery renders without label clipping under default DPI/sizes.
- Theme snapshot tests (style toggles don’t change geometry).

---

# Milestone 8 — Skia backend (CPU/GPU) + PDF

**Scope**

- `backends/skia` with CPU first; optional GPU; PDF export.

**Tasks**

- Map renderer verbs to Skia; ensure parity with AGG.
- PDF surface output; font bridge consistent with shaping layer.
- Cross-backend parity tests (Skia vs AGG).

**Exit criteria**

- All gallery plots match AGG within tolerance.
- `Savefig("plot.pdf")` works for core artists.

---

# Milestone 9 — Interactivity basics (desktop)

**Scope**

- Event model; pan/zoom; redraw loop.

**Tasks**

- GLFW canvas backend using Skia surface.
- Events: mouse move/down/up, scroll, key, resize; picking (hit test).
- Simple toolbar: zoom, pan, reset.

**Exit criteria**

- Example app: interactively pan/zoom a scatter of 100k points smoothly.

---

# Milestone 10 — Performance & DX

**Scope**

- Benchmarks, decimation, docs & examples gallery.

**Tasks**

- Micro-benchmarks: stroking, text layout, image upload.
- Optional point decimation/tiling; caching transformed paths.
- Docs site: API reference, “Matplotlib to Go” guide; CI builds gallery.

**Exit criteria**

- Benchmarks tracked in CI; regressions alert.
- Public examples folder (20–30 plots) doubles as test suite.

---

## Cross-cutting “always-on” tracks

- **Testing:** golden images, SSIM/PSNR checks, property tests on transforms.
- **Determinism:** pin font versions for CI; document backend settings.
- **API hygiene:** functional options, no global state; stability review before M8.
