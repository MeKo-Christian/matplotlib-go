package agg

import (
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func transformMarkerPath(path geom.Path, affine geom.Affine, offset geom.Pt) geom.Path {
	if len(path.C) == 0 {
		return geom.Path{}
	}
	out := geom.Path{
		C: append([]geom.Cmd(nil), path.C...),
		V: make([]geom.Pt, len(path.V)),
	}
	for i, pt := range path.V {
		pt = affine.Apply(pt)
		out.V[i] = geom.Pt{X: pt.X + offset.X, Y: pt.Y + offset.Y}
	}
	return out
}

// markerShapeDevice builds the marker shape in y-down device orientation,
// centred at the origin (no offset applied). It mirrors Matplotlib's
// marker_trans *= scale(1,-1): the marker affine is applied and the Y axis is
// negated so the shape is oriented for the device buffer. The result reuses
// markerScratch and is intended to be snapped (around the origin) and then
// stamped at the rounded device offset, matching draw_markers in _backend_agg.h.
func (r *Renderer) markerShapeDevice(path geom.Path, affine geom.Affine) geom.Path {
	if r == nil || len(path.C) == 0 {
		return geom.Path{}
	}
	out := &r.markerScratch
	out.C = path.C
	if cap(out.V) < len(path.V) {
		out.V = make([]geom.Pt, len(path.V))
	} else {
		out.V = out.V[:len(path.V)]
	}
	for i, pt := range path.V {
		pt = affine.Apply(pt)
		out.V[i] = geom.Pt{X: pt.X, Y: -pt.Y}
	}
	return *out
}

func (r *Renderer) transformMarkerPathDevice(path geom.Path, affine geom.Affine, offset geom.Pt) geom.Path {
	if r == nil || len(path.C) == 0 {
		return geom.Path{}
	}
	out := &r.markerScratch
	out.C = path.C
	if cap(out.V) < len(path.V) {
		out.V = make([]geom.Pt, len(path.V))
	} else {
		out.V = out.V[:len(path.V)]
	}
	h := float64(r.height)
	for i, pt := range path.V {
		pt = affine.Apply(pt)
		out.V[i] = geom.Pt{X: pt.X + offset.X, Y: h - (pt.Y + offset.Y)}
	}
	return *out
}

func applyForcedAlpha(paint *render.Paint) {
	if paint == nil || !paint.ForceAlpha {
		return
	}
	alpha := clamp01(paint.Alpha)
	if paint.Stroke.A > 0 {
		paint.Stroke.A = alpha
	}
	if paint.Fill.A > 0 {
		paint.Fill.A = alpha
	}
	if paint.HatchColor.A > 0 {
		paint.HatchColor.A = alpha
	}
}

func colorWithForcedAlpha(c render.Color, paint *render.Paint) render.Color {
	if paint != nil && paint.ForceAlpha && c.A > 0 {
		c.A = clamp01(paint.Alpha)
	}
	return c
}

func (r *Renderer) applyAntialiasMode(mode render.AntialiasMode) func() {
	if r.ctx == nil {
		return func() {}
	}
	switch mode {
	case render.AntialiasOn:
		prev := r.ctx.GetAntiAliasGamma()
		prevAA := r.ctx.GetAntiAliased()
		r.ctx.SetAntiAliased(true)
		r.ctx.SetAntiAliasGamma(1.0)
		return func() {
			r.ctx.SetAntiAliasGamma(prev)
			r.ctx.SetAntiAliased(prevAA)
		}
	case render.AntialiasOff:
		// matplotlib's Agg backend switches to scanline_bin /
		// renderer_scanline_bin_solid when a graphics context has
		// antialiased=false: every touched cell becomes a fully covered
		// pixel (_backend_agg.h _draw_path).
		prevAA := r.ctx.GetAntiAliased()
		r.ctx.SetAntiAliased(false)
		return func() {
			r.ctx.SetAntiAliased(prevAA)
		}
	default:
		return func() {}
	}
}
