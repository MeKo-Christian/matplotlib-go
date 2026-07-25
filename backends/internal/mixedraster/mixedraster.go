package mixedraster

import (
	"math"

	"github.com/cwbudde/matplotlib-go/backends/gobasic"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// Session captures one mixed raster/vector draw group on a transparent
// offscreen surface.
type Session struct {
	backing  *gobasic.Renderer
	renderer render.Renderer
	rect     geom.Rect
}

// Start creates a transparent raster surface sized to the vector page.
func Start(width, height int, viewport geom.Rect, options render.Rasterization, fallbackDPI uint, clipRect *geom.Rect, clipPaths []geom.Path) (*Session, bool) {
	if width <= 0 || height <= 0 {
		return nil, false
	}
	dpi := fallbackDPI
	if options.DPI > 0 {
		dpi = uint(math.Round(options.DPI))
	}
	if dpi == 0 {
		dpi = 72
	}
	scale := float64(dpi) / 72.0
	if scale <= 0 || math.IsNaN(scale) || math.IsInf(scale, 0) {
		scale = 1
	}
	surfaceWidth := int(math.Ceil(float64(width) * scale))
	surfaceHeight := int(math.Ceil(float64(height) * scale))
	if surfaceWidth <= 0 {
		surfaceWidth = 1
	}
	if surfaceHeight <= 0 {
		surfaceHeight = 1
	}

	r := gobasic.New(surfaceWidth, surfaceHeight, render.Color{})
	r.SetResolution(dpi)
	scaled := &scaledRenderer{
		inner: r,
		scale: scale,
	}
	if err := scaled.Begin(viewport); err != nil {
		return nil, false
	}
	if clipRect != nil {
		scaled.ClipRect(*clipRect)
	}
	for _, clipPath := range clipPaths {
		scaled.ClipPath(clipPath)
	}
	return &Session{
		backing:  r,
		renderer: scaled,
		rect: geom.Rect{
			Min: geom.Pt{},
			Max: geom.Pt{X: float64(width), Y: float64(height)},
		},
	}, true
}

// Renderer returns the active offscreen renderer.
func (s *Session) Renderer() render.Renderer {
	if s == nil {
		return nil
	}
	return s.renderer
}

// Stop finishes the offscreen draw and returns an image suitable for embedding.
func (s *Session) Stop() (*render.ImageData, geom.Rect, bool) {
	if s == nil || s.backing == nil {
		return nil, geom.Rect{}, false
	}
	_ = s.renderer.End()
	return render.NewImageData(s.backing.Image()), s.rect, true
}

type scaledRenderer struct {
	inner *gobasic.Renderer
	scale float64
}

var (
	_ render.Renderer           = (*scaledRenderer)(nil)
	_ render.DPIAware           = (*scaledRenderer)(nil)
	_ render.ImageTransformer   = (*scaledRenderer)(nil)
	_ render.TextDrawer         = (*scaledRenderer)(nil)
	_ render.TextPather         = (*scaledRenderer)(nil)
	_ render.RotatedTextDrawer  = (*scaledRenderer)(nil)
	_ render.VerticalTextDrawer = (*scaledRenderer)(nil)
)

func (r *scaledRenderer) Begin(viewport geom.Rect) error {
	return r.inner.Begin(r.scaleRect(viewport))
}

func (r *scaledRenderer) End() error { return r.inner.End() }

func (r *scaledRenderer) Save() { r.inner.Save() }

func (r *scaledRenderer) Restore() { r.inner.Restore() }

func (r *scaledRenderer) ClipRect(rect geom.Rect) { r.inner.ClipRect(r.scaleRect(rect)) }

func (r *scaledRenderer) ClipPath(path geom.Path) { r.inner.ClipPath(r.scalePath(path)) }

func (r *scaledRenderer) Path(path geom.Path, paint *render.Paint) {
	r.inner.Path(r.scalePath(path), scalePaint(paint, r.scale))
}

func (r *scaledRenderer) DrawImage(img render.Image, dst geom.Rect) {
	r.inner.DrawImage(img, r.scaleRect(dst))
}

func (r *scaledRenderer) ImageTransformed(img render.Image, dst geom.Rect, transform geom.Affine) {
	r.inner.ImageTransformed(img, r.scaleRect(dst), geom.Affine{A: r.scale, D: r.scale}.Mul(transform))
}

func (r *scaledRenderer) GlyphRun(run render.GlyphRun, color render.Color) {
	run.Origin = r.scalePt(run.Origin)
	for i := range run.Glyphs {
		run.Glyphs[i].Offset = r.scalePt(run.Glyphs[i].Offset)
		run.Glyphs[i].Advance *= r.scale
	}
	r.inner.GlyphRun(run, color)
}

func (r *scaledRenderer) MeasureText(text string, size float64, fontKey string) render.TextMetrics {
	return scaleTextMetrics(r.inner.MeasureText(text, size, fontKey), 1/r.scale)
}

func (r *scaledRenderer) SetResolution(dpi uint) { r.inner.SetResolution(dpi) }

func (r *scaledRenderer) DrawText(text string, origin geom.Pt, size float64, textColor render.Color) {
	r.inner.DrawText(text, r.scalePt(origin), size, textColor)
}

func (r *scaledRenderer) DrawTextRotated(text string, anchor geom.Pt, size, angle float64, textColor render.Color) {
	r.inner.DrawTextRotated(text, r.scalePt(anchor), size, angle, textColor)
}

func (r *scaledRenderer) DrawTextVertical(text string, center geom.Pt, size float64, textColor render.Color) {
	r.inner.DrawTextVertical(text, r.scalePt(center), size, textColor)
}

func (r *scaledRenderer) TextPath(text string, origin geom.Pt, size float64, fontKey string) (geom.Path, bool) {
	return r.inner.TextPath(text, origin, size, fontKey)
}

func (r *scaledRenderer) scalePt(pt geom.Pt) geom.Pt {
	return geom.Pt{X: pt.X * r.scale, Y: pt.Y * r.scale}
}

func (r *scaledRenderer) scaleRect(rect geom.Rect) geom.Rect {
	return geom.Rect{
		Min: r.scalePt(rect.Min),
		Max: r.scalePt(rect.Max),
	}
}

func (r *scaledRenderer) scalePath(path geom.Path) geom.Path {
	out := ClonePath(path)
	for i, pt := range out.V {
		out.V[i] = r.scalePt(pt)
	}
	return out
}

func scalePaint(paint *render.Paint, scale float64) *render.Paint {
	if paint == nil {
		return nil
	}
	out := *paint
	out.LineWidth *= scale
	out.MiterLimit *= scale
	out.SimplifyThreshold *= scale
	out.HatchLineWidth *= scale
	out.HatchSpacing *= scale
	out.Dashes = scaleFloats(paint.Dashes, scale)
	out.PathEffects = scalePathEffects(paint.PathEffects, scale)
	if out.HasClipPathTrans {
		out.ClipPathTransform = scaleAffine(out.ClipPathTransform, scale)
	}
	return &out
}

func scalePathEffects(effects []render.PathEffect, scale float64) []render.PathEffect {
	if len(effects) == 0 {
		return nil
	}
	out := make([]render.PathEffect, len(effects))
	for i, effect := range effects {
		effect.Offset.X *= scale
		effect.Offset.Y *= scale
		effect.LineWidth *= scale
		effect.MiterLimit *= scale
		effect.FilterRadius *= scale
		effect.TickSpacing *= scale
		effect.TickLength *= scale
		effect.Dashes = scaleFloats(effect.Dashes, scale)
		out[i] = effect
	}
	return out
}

func scaleFloats(values []float64, scale float64) []float64 {
	if len(values) == 0 {
		return nil
	}
	out := make([]float64, len(values))
	for i, value := range values {
		out[i] = value * scale
	}
	return out
}

func scaleAffine(affine geom.Affine, scale float64) geom.Affine {
	return geom.Affine{A: scale, D: scale}.Mul(affine)
}

func scaleTextMetrics(metrics render.TextMetrics, scale float64) render.TextMetrics {
	metrics.W *= scale
	metrics.H *= scale
	metrics.Ascent *= scale
	metrics.Descent *= scale
	return metrics
}

// ClonePaths returns a deep copy of paths suitable for renderer state stacks.
func ClonePaths(paths []geom.Path) []geom.Path {
	if len(paths) == 0 {
		return nil
	}
	out := make([]geom.Path, len(paths))
	for i, path := range paths {
		out[i] = ClonePath(path)
	}
	return out
}

// ClonePath returns a deep copy of path.
func ClonePath(path geom.Path) geom.Path {
	if len(path.C) == 0 {
		return geom.Path{}
	}
	return geom.Path{
		C: append([]geom.Cmd(nil), path.C...),
		V: append([]geom.Pt(nil), path.V...),
	}
}

// ApplyAffine returns a copy of path transformed by affine.
func ApplyAffine(path geom.Path, affine geom.Affine) geom.Path {
	out := ClonePath(path)
	for i, pt := range out.V {
		out.V[i] = affine.Apply(pt)
	}
	return out
}
