package core

import (
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// Draw renders the axis spine on the axes edge.
func (a *Axis) Draw(r render.Renderer, ctx *DrawContext) {
	if isPolarProjection(ctx.Projection) {
		a.drawPolarSpine(r, ctx)
		return
	}
	if framePath, ok := projectionFramePath(ctx.Projection, ctx.Clip); ok {
		if a.ShowSpine && (a.Side == AxisBottom || a.Side == AxisTop) {
			paint := axisStrokePaint(a, false)
			paint.LineCap = render.CapButt
			r.Path(framePath, &paint)
		}
		return
	}
	if a.ShowSpine {
		a.drawSpine(r, ctx)
	}
}

// Z returns the z-order for sorting.
func (a *Axis) Z() float64 {
	return a.z
}

// Bounds returns an empty rect for now.
func (a *Axis) Bounds(*DrawContext) geom.Rect {
	return geom.Rect{}
}
