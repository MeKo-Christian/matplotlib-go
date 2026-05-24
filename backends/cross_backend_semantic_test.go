package backends_test

import (
	"image"
	"testing"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/agg"  // register AGG backend
	_ "github.com/cwbudde/matplotlib-go/backends/pdf"  // register PDF backend
	_ "github.com/cwbudde/matplotlib-go/backends/skia" // register Skia backend
	_ "github.com/cwbudde/matplotlib-go/backends/svg"  // register SVG backend
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// TestCrossBackendSemanticParity drives the same Figure through both the AGG
// and SVG backends and verifies they receive equivalent semantic operations.
// We compare counts of high-level draw verbs (clip pushes, marker batches,
// path-collection batches, glyph runs, text draws, transformed clips/images)
// rather than pixel or byte parity, because the two backends produce
// different output formats by design.
//
// This is the cross-backend half of 14.3.6: it guards against regressions where
// core artist code accidentally routes one backend through a different code
// path (for example, falling back to per-marker Path calls for SVG while AGG
// keeps using DrawMarkers).
func TestCrossBackendSemanticParity(t *testing.T) {
	makeFigure := func() *core.Figure {
		fig := core.NewFigure(200, 150)
		ax := fig.AddAxes(geom.Rect{
			Min: geom.Pt{X: 0.1, Y: 0.1},
			Max: geom.Pt{X: 0.9, Y: 0.9},
		})
		ax.SetTitle("semantic parity")
		ax.Plot([]float64{0, 1, 2, 3}, []float64{0, 1, 0.5, 2})
		ax.Scatter(
			[]float64{0.5, 1.5, 2.5},
			[]float64{0.25, 1.25, 0.75},
			core.ScatterOptions{Label: "samples"},
		)
		return fig
	}

	aggCounts := drawAndCount(t, backends.AGG, makeFigure())
	svgCounts := drawAndCount(t, backends.SVG, makeFigure())

	compareCounts(t, &aggCounts, &svgCounts)
}

func TestCrossBackendPatternGradientSemanticParity(t *testing.T) {
	makeFigure := func() *core.Figure {
		fig := core.NewFigure(200, 150)
		ax := fig.AddAxes(geom.Rect{
			Min: geom.Pt{X: 0.1, Y: 0.1},
			Max: geom.Pt{X: 0.9, Y: 0.9},
		})
		ax.Add(patternGradientSemanticArtist{})
		return fig
	}

	want := drawAndCount(t, backends.AGG, makeFigure())
	assertEffectRoutingCounts(t, backends.AGG, want)
	if want.filterStart == 0 || want.filterStop == 0 {
		t.Fatalf("AGG semantic probe did not route filtered path effects through render.FilterRenderer: %+v", want)
	}
	for _, backend := range []backends.Backend{backends.SVG, backends.PDF, backends.Skia} {
		info, ok := backends.DefaultRegistry.Get(backend)
		if !ok {
			t.Fatalf("backend %s not registered", backend)
		}
		if !info.Available {
			t.Run(string(backend), func(t *testing.T) {
				t.Skipf("backend %s is not available in this build", backend)
			})
			continue
		}
		got := drawAndCount(t, backend, makeFigure())
		assertEffectRoutingCounts(t, backend, got)
		if got.gradientPaint != want.gradientPaint || got.patternPaint != want.patternPaint {
			t.Fatalf("%s pattern/gradient semantic counts = gradient %d pattern %d, want gradient %d pattern %d",
				backend, got.gradientPaint, got.patternPaint, want.gradientPaint, want.patternPaint)
		}
		if backend == backends.SVG || backend == backends.PDF {
			if got.pathEffectFilterDraw == 0 || got.filterStart != 0 || got.filterStop != 0 {
				t.Fatalf("%s filtered path-effect routing = native draws %d filter starts/stops %d/%d, want native filter drawer only",
					backend, got.pathEffectFilterDraw, got.filterStart, got.filterStop)
			}
			continue
		}
		if got.pathEffectFilterDraw != 0 || got.filterStart == 0 || got.filterStop == 0 {
			t.Fatalf("%s filtered path-effect routing = native draws %d filter starts/stops %d/%d, want render.FilterRenderer",
				backend, got.pathEffectFilterDraw, got.filterStart, got.filterStop)
		}
	}
}

// drawAndCount instantiates the given backend, wraps it with a counting tee
// renderer, and drives the figure through it. It returns the counts of the
// semantic verbs we audit.
func drawAndCount(t *testing.T, backend backends.Backend, fig *core.Figure) semanticCounts {
	t.Helper()

	inner, err := backends.Create(backend, backends.Config{
		Width:      200,
		Height:     150,
		DPI:        72,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
	})
	if err != nil {
		t.Fatalf("create %s backend: %v", backend, err)
	}

	wrap := &countingRenderer{inner: inner}
	core.DrawFigure(fig, wrap)
	return wrap.counts
}

func compareCounts(t *testing.T, agg, svg *semanticCounts) {
	t.Helper()

	mismatches := []struct {
		name     string
		agg, svg int
	}{
		{"ClipRect", agg.clipRect, svg.clipRect},
		{"ClipPath", agg.clipPath, svg.clipPath},
		{"Path", agg.path, svg.path},
		{"Image", agg.image, svg.image},
		{"ImageTransformed", agg.imageTransformed, svg.imageTransformed},
		{"GlyphRun", agg.glyphRun, svg.glyphRun},
		{"GlyphsTotal", agg.glyphTotal, svg.glyphTotal},
		{"DrawText", agg.drawText, svg.drawText},
		{"DrawTextRotated", agg.drawTextRot, svg.drawTextRot},
		{"DrawTextVertical", agg.drawTextVert, svg.drawTextVert},
		{"DrawMarkers", agg.drawMarkers, svg.drawMarkers},
		{"MarkerItemsTotal", agg.markerItems, svg.markerItems},
		{"DrawPathCollection", agg.drawPathColl, svg.drawPathColl},
		{"PathCollectionItemsTotal", agg.pathCollItems, svg.pathCollItems},
		{"Save", agg.save, svg.save},
		{"Restore", agg.restore, svg.restore},
	}
	for _, m := range mismatches {
		if m.agg != m.svg {
			t.Errorf("backend semantic divergence: %s AGG=%d SVG=%d", m.name, m.agg, m.svg)
		}
	}
}

func assertEffectRoutingCounts(t *testing.T, backend backends.Backend, counts semanticCounts) {
	t.Helper()
	if counts.gradientPaint == 0 {
		t.Fatalf("%s semantic probe did not draw a gradient paint: %+v", backend, counts)
	}
	if counts.patternPaint == 0 {
		t.Fatalf("%s semantic probe did not draw a pattern paint: %+v", backend, counts)
	}
	if counts.pathEffectPaint == 0 {
		t.Fatalf("%s semantic probe did not draw a path-effect paint: %+v", backend, counts)
	}
	if counts.pathEffectDrawer == 0 {
		t.Fatalf("%s semantic probe did not route path effects through render.PathEffectDrawer: %+v", backend, counts)
	}
	if counts.pathEffectFilterSupports == 0 {
		t.Fatalf("%s semantic probe did not query render.PathEffectFilterDrawer support: %+v", backend, counts)
	}
}

type patternGradientSemanticArtist struct{}

func (patternGradientSemanticArtist) Draw(r render.Renderer, ctx *core.DrawContext) {
	gradientPath := semanticRectPath(30, 30, 92, 94)
	r.Path(gradientPath, &render.Paint{
		FillGradient: render.GradientFill{
			Kind:  render.LinearGradient,
			Start: geom.Pt{X: 30, Y: 62},
			End:   geom.Pt{X: 92, Y: 62},
			Stops: []render.GradientStop{
				{Offset: 0, Color: render.Color{R: 1, A: 1}},
				{Offset: 1, Color: render.Color{B: 1, A: 1}},
			},
		},
		Stroke:    render.Color{A: 1},
		LineWidth: 1,
	})

	var tile geom.Path
	tile.MoveTo(geom.Pt{X: 0, Y: 10})
	tile.LineTo(geom.Pt{X: 10, Y: 0})
	patternPath := semanticRectPath(108, 38, 170, 102)
	r.Path(patternPath, &render.Paint{
		FillPattern: render.PatternFill{
			ID:         "semantic-diagonal",
			Cell:       geom.Rect{Max: geom.Pt{X: 10, Y: 10}},
			Path:       tile,
			Foreground: render.Color{R: 0.1, G: 0.2, B: 0.8, A: 1},
			Background: render.Color{R: 0.9, G: 0.9, B: 0.95, A: 1},
			LineWidth:  1,
		},
		Stroke:    render.Color{A: 1},
		LineWidth: 1,
	})

	effectPath := semanticRectPath(42, 104, 92, 132)
	r.Path(effectPath, &render.Paint{
		Fill:      render.Color{G: 0.45, B: 0.2, A: 1},
		Stroke:    render.Color{A: 1},
		LineWidth: 2,
		PathEffects: []render.PathEffect{
			render.StrokePathEffect(render.Color{R: 1, G: 0.55, A: 1}, 5, geom.Pt{X: 3, Y: -3}),
			render.NormalPathEffect(),
		},
	})

	filterPath := semanticRectPath(112, 108, 164, 136)
	r.Path(filterPath, &render.Paint{
		PathEffects: []render.PathEffect{
			render.FilterPathEffect(render.Color{R: 0.15, B: 0.75, A: 0.8}, render.Color{}, 0, "blur", 3, geom.Pt{X: 3, Y: -2}),
			render.NormalPathEffect(),
		},
	})
	_ = ctx
}

func (patternGradientSemanticArtist) Z() float64 { return 1 }

func (patternGradientSemanticArtist) Bounds(*core.DrawContext) geom.Rect { return geom.Rect{} }

func semanticRectPath(x0, y0, x1, y1 float64) geom.Path {
	var p geom.Path
	p.MoveTo(geom.Pt{X: x0, Y: y0})
	p.LineTo(geom.Pt{X: x1, Y: y0})
	p.LineTo(geom.Pt{X: x1, Y: y1})
	p.LineTo(geom.Pt{X: x0, Y: y1})
	p.Close()
	return p
}

type semanticCounts struct {
	clipRect, clipPath                            int
	path, image                                   int
	gradientPaint, patternPaint, pathEffectPaint  int
	pathEffectDrawer, pathEffectFilterSupports    int
	pathEffectFilterDraw, filterStart, filterStop int
	imageTransformed                              int
	glyphRun, glyphTotal                          int
	drawText, drawTextRot                         int
	drawTextVert                                  int
	drawMarkers, markerItems                      int
	drawPathColl, pathCollItems                   int
	save, restore                                 int
}

// countingRenderer wraps a render.Renderer and counts the semantic verbs it
// receives, then forwards every call to the wrapped renderer. It implements
// every optional capability interface that AGG and SVG share, so capability
// checks in core code resolve identically for both backends.
type countingRenderer struct {
	inner  render.Renderer
	counts semanticCounts
}

// Compile-time guards: countingRenderer must satisfy the shared subset of
// capabilities AGG and SVG both implement, so type-assertion-based dispatch in
// core code routes identically for both backends.
var (
	_ render.Renderer               = (*countingRenderer)(nil)
	_ render.DPIAware               = (*countingRenderer)(nil)
	_ render.TextDrawer             = (*countingRenderer)(nil)
	_ render.RotatedTextDrawer      = (*countingRenderer)(nil)
	_ render.VerticalTextDrawer     = (*countingRenderer)(nil)
	_ render.TextPather             = (*countingRenderer)(nil)
	_ render.TeXMetricer            = (*countingRenderer)(nil)
	_ render.TeXDrawer              = (*countingRenderer)(nil)
	_ render.RotatedTeXDrawer       = (*countingRenderer)(nil)
	_ render.ImageTransformer       = (*countingRenderer)(nil)
	_ render.MarkerDrawer           = (*countingRenderer)(nil)
	_ render.PathCollectionDrawer   = (*countingRenderer)(nil)
	_ render.NativeHatcher          = (*countingRenderer)(nil)
	_ render.PatternFiller          = (*countingRenderer)(nil)
	_ render.GradientFiller         = (*countingRenderer)(nil)
	_ render.PathEffectDrawer       = (*countingRenderer)(nil)
	_ render.PathEffectFilterDrawer = (*countingRenderer)(nil)
	_ render.FilterRenderer         = (*countingRenderer)(nil)
)

func (r *countingRenderer) Begin(vp geom.Rect) error { return r.inner.Begin(vp) }
func (r *countingRenderer) End() error               { return r.inner.End() }

func (r *countingRenderer) Save() {
	r.counts.save++
	r.inner.Save()
}

func (r *countingRenderer) Restore() {
	r.counts.restore++
	r.inner.Restore()
}

func (r *countingRenderer) ClipRect(rect geom.Rect) {
	r.counts.clipRect++
	r.inner.ClipRect(rect)
}

func (r *countingRenderer) ClipPath(p geom.Path) {
	r.counts.clipPath++
	r.inner.ClipPath(p)
}

func (r *countingRenderer) Path(p geom.Path, paint *render.Paint) {
	r.counts.path++
	if paint != nil {
		if paint.FillGradient.Kind != render.GradientNone && len(paint.FillGradient.Stops) > 0 {
			r.counts.gradientPaint++
		}
		if paint.FillPattern.ID != "" || len(paint.FillPattern.Path.V) > 0 {
			r.counts.patternPaint++
		}
		if len(paint.PathEffects) > 0 {
			r.counts.pathEffectPaint++
			if r.DrawPathWithEffects(p, paint) {
				return
			}
		}
	}
	r.inner.Path(p, paint)
}

func (r *countingRenderer) Image(img render.Image, dst geom.Rect) {
	r.counts.image++
	r.inner.Image(img, dst)
}

func (r *countingRenderer) GlyphRun(run render.GlyphRun, c render.Color) {
	r.counts.glyphRun++
	r.counts.glyphTotal += len(run.Glyphs)
	r.inner.GlyphRun(run, c)
}

func (r *countingRenderer) MeasureText(text string, size float64, fontKey string) render.TextMetrics {
	return r.inner.MeasureText(text, size, fontKey)
}

func (r *countingRenderer) SetResolution(dpi uint) {
	if d, ok := r.inner.(render.DPIAware); ok {
		d.SetResolution(dpi)
	}
}

func (r *countingRenderer) DrawText(text string, origin geom.Pt, size float64, c render.Color) {
	r.counts.drawText++
	if d, ok := r.inner.(render.TextDrawer); ok {
		d.DrawText(text, origin, size, c)
	}
}

func (r *countingRenderer) DrawTextRotated(text string, anchor geom.Pt, size, angle float64, c render.Color) {
	r.counts.drawTextRot++
	if d, ok := r.inner.(render.RotatedTextDrawer); ok {
		d.DrawTextRotated(text, anchor, size, angle, c)
	}
}

func (r *countingRenderer) DrawTextVertical(text string, center geom.Pt, size float64, c render.Color) {
	r.counts.drawTextVert++
	if d, ok := r.inner.(render.VerticalTextDrawer); ok {
		d.DrawTextVertical(text, center, size, c)
	}
}

func (r *countingRenderer) TextPath(text string, origin geom.Pt, size float64, fontKey string) (geom.Path, bool) {
	if d, ok := r.inner.(render.TextPather); ok {
		return d.TextPath(text, origin, size, fontKey)
	}
	return geom.Path{}, false
}

func (r *countingRenderer) MeasureTeX(text string, size float64, fontKey string) (render.TextMetrics, bool) {
	if d, ok := r.inner.(render.TeXMetricer); ok {
		return d.MeasureTeX(text, size, fontKey)
	}
	return render.TextMetrics{}, false
}

func (r *countingRenderer) DrawTeX(text string, origin geom.Pt, size float64, c render.Color, fontKey string) bool {
	if d, ok := r.inner.(render.TeXDrawer); ok {
		return d.DrawTeX(text, origin, size, c, fontKey)
	}
	return false
}

func (r *countingRenderer) DrawTeXRotated(text string, anchor geom.Pt, size, angle float64, c render.Color, fontKey string) bool {
	if d, ok := r.inner.(render.RotatedTeXDrawer); ok {
		return d.DrawTeXRotated(text, anchor, size, angle, c, fontKey)
	}
	return false
}

func (r *countingRenderer) ImageTransformed(img render.Image, dst geom.Rect, transform geom.Affine) {
	r.counts.imageTransformed++
	if d, ok := r.inner.(render.ImageTransformer); ok {
		d.ImageTransformed(img, dst, transform)
	}
}

func (r *countingRenderer) DrawMarkers(batch render.MarkerBatch) bool {
	r.counts.drawMarkers++
	r.counts.markerItems += len(batch.Items)
	if d, ok := r.inner.(render.MarkerDrawer); ok {
		return d.DrawMarkers(batch)
	}
	return false
}

func (r *countingRenderer) DrawPathCollection(batch render.PathCollectionBatch) bool {
	r.counts.drawPathColl++
	r.counts.pathCollItems += len(batch.Items)
	if d, ok := r.inner.(render.PathCollectionDrawer); ok {
		return d.DrawPathCollection(batch)
	}
	return false
}

func (r *countingRenderer) SupportsNativeHatch() bool {
	if d, ok := r.inner.(render.NativeHatcher); ok {
		return d.SupportsNativeHatch()
	}
	return false
}

func (r *countingRenderer) SupportsPatternFill() bool {
	if d, ok := r.inner.(render.PatternFiller); ok {
		return d.SupportsPatternFill()
	}
	return false
}

func (r *countingRenderer) SupportsGradientFill() bool {
	if d, ok := r.inner.(render.GradientFiller); ok {
		return d.SupportsGradientFill()
	}
	return false
}

func (r *countingRenderer) DrawPathWithEffects(p geom.Path, paint *render.Paint) bool {
	r.counts.pathEffectDrawer++
	return render.DrawPathWithEffects(r, p, paint, r.Path)
}

func (r *countingRenderer) SupportsPathEffectFilter(effect render.PathEffect) bool {
	r.counts.pathEffectFilterSupports++
	if d, ok := r.inner.(render.PathEffectFilterDrawer); ok {
		return d.SupportsPathEffectFilter(effect)
	}
	return false
}

func (r *countingRenderer) DrawPathEffectFilter(path geom.Path, paint render.Paint, effect render.PathEffect, draw func(geom.Path, *render.Paint)) bool {
	r.counts.pathEffectFilterDraw++
	if d, ok := r.inner.(render.PathEffectFilterDrawer); ok {
		return d.DrawPathEffectFilter(path, paint, effect, draw)
	}
	return false
}

func (r *countingRenderer) StartFilter() {
	r.counts.filterStart++
	if d, ok := r.inner.(render.FilterRenderer); ok {
		d.StartFilter()
	}
}

func (r *countingRenderer) StopFilter(postProcess func(*image.RGBA, float64) (*image.RGBA, geom.Pt)) {
	r.counts.filterStop++
	if d, ok := r.inner.(render.FilterRenderer); ok {
		d.StopFilter(postProcess)
	}
}
