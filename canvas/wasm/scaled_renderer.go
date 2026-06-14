//go:build js && wasm

package wasm

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

type scaledRenderer struct {
	base  render.Renderer
	scale float64
}

var _ render.Renderer = (*scaledRenderer)(nil)

func newScaledRenderer(base render.Renderer, scale float64) render.Renderer {
	if scale <= 0 || scale == 1 {
		return base
	}
	return &scaledRenderer{base: base, scale: scale}
}

func (r *scaledRenderer) Begin(viewport geom.Rect) error {
	return r.base.Begin(scaleRect(viewport, r.scale))
}

func (r *scaledRenderer) End() error {
	return r.base.End()
}

func (r *scaledRenderer) Save() {
	r.base.Save()
}

func (r *scaledRenderer) Restore() {
	r.base.Restore()
}

func (r *scaledRenderer) ClipRect(rect geom.Rect) {
	r.base.ClipRect(scaleRect(rect, r.scale))
}

func (r *scaledRenderer) ClipPath(path geom.Path) {
	r.base.ClipPath(scalePath(path, r.scale))
}

func (r *scaledRenderer) Path(path geom.Path, paint *render.Paint) {
	r.base.Path(scalePath(path, r.scale), scalePaint(paint, r.scale))
}

func (r *scaledRenderer) Image(img render.Image, dst geom.Rect) {
	r.base.Image(img, scaleRect(dst, r.scale))
}

func (r *scaledRenderer) GlyphRun(run render.GlyphRun, color render.Color) {
	scaled := run
	scaled.Origin = scalePt(run.Origin, r.scale)
	scaled.Size = run.Size * r.scale
	scaled.Glyphs = make([]render.Glyph, len(run.Glyphs))
	for i, glyph := range run.Glyphs {
		scaled.Glyphs[i] = render.Glyph{
			ID:      glyph.ID,
			Advance: glyph.Advance * r.scale,
			Offset:  scalePt(glyph.Offset, r.scale),
		}
	}
	r.base.GlyphRun(scaled, color)
}

func (r *scaledRenderer) MeasureText(text string, size float64, fontKey string) render.TextMetrics {
	metrics := r.base.MeasureText(text, size*r.scale, fontKey)
	return render.TextMetrics{
		W:       metrics.W / r.scale,
		H:       metrics.H / r.scale,
		Ascent:  metrics.Ascent / r.scale,
		Descent: metrics.Descent / r.scale,
	}
}

func (r *scaledRenderer) SetResolution(dpi uint) {
	if setter, ok := r.base.(render.DPIAware); ok {
		setter.SetResolution(uint(math.Round(float64(dpi) * r.scale)))
	}
}

func (r *scaledRenderer) DrawText(text string, origin geom.Pt, size float64, textColor render.Color) {
	if drawer, ok := r.base.(render.TextDrawer); ok {
		drawer.DrawText(text, scalePt(origin, r.scale), size*r.scale, textColor)
	}
}

func (r *scaledRenderer) DrawTextRotated(text string, anchor geom.Pt, size, angle float64, textColor render.Color) {
	if drawer, ok := r.base.(render.RotatedTextDrawer); ok {
		drawer.DrawTextRotated(text, scalePt(anchor, r.scale), size*r.scale, angle, textColor)
		return
	}
	r.DrawText(text, anchor, size, textColor)
}

func (r *scaledRenderer) DrawTextVertical(text string, center geom.Pt, size float64, textColor render.Color) {
	if drawer, ok := r.base.(render.VerticalTextDrawer); ok {
		drawer.DrawTextVertical(text, scalePt(center, r.scale), size*r.scale, textColor)
		return
	}
	r.DrawText(text, center, size, textColor)
}

func scalePaint(paint *render.Paint, scale float64) *render.Paint {
	if paint == nil {
		return nil
	}
	scaled := *paint
	scaled.LineWidth *= scale
	scaled.MiterLimit *= scale
	if len(paint.Dashes) > 0 {
		scaled.Dashes = make([]float64, len(paint.Dashes))
		for i, dash := range paint.Dashes {
			scaled.Dashes[i] = dash * scale
		}
	}
	return &scaled
}

func scalePath(path geom.Path, scale float64) geom.Path {
	scaled := geom.Path{
		C: make([]geom.Cmd, len(path.C)),
		V: make([]geom.Pt, len(path.V)),
	}
	copy(scaled.C, path.C)
	for i, pt := range path.V {
		scaled.V[i] = scalePt(pt, scale)
	}
	return scaled
}

func scaleRect(rect geom.Rect, scale float64) geom.Rect {
	return geom.Rect{
		Min: scalePt(rect.Min, scale),
		Max: scalePt(rect.Max, scale),
	}
}

func scalePt(pt geom.Pt, scale float64) geom.Pt {
	return geom.Pt{X: pt.X * scale, Y: pt.Y * scale}
}
