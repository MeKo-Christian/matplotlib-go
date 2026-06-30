package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// EllipseSelectorCallback receives the ellipse bounding box in data coordinates.
type EllipseSelectorCallback func(*EllipseSelector, geom.Rect)

// EllipseSelectorOptions configures an ellipse selector.
type EllipseSelectorOptions struct {
	EdgeColor render.Color
	FillColor render.Color
	LineWidth float64
}

// EllipseSelector stores a selected ellipse region by data bounding box.
type EllipseSelector struct {
	Min       geom.Pt
	Max       geom.Pt
	EdgeColor render.Color
	FillColor render.Color
	LineWidth float64
	Active    bool

	onSelect widgetCallbackRegistry[EllipseSelectorCallback]
	z        float64
}

// EllipseSelector creates an ellipse selector bound to the axes.
func (a *Axes) EllipseSelector(opts ...EllipseSelectorOptions) *EllipseSelector {
	if a == nil {
		return nil
	}
	config := EllipseSelectorOptions{
		EdgeColor: render.Color{R: 0.16, G: 0.42, B: 0.76, A: 1},
		FillColor: render.Color{R: 0.16, G: 0.42, B: 0.76, A: 0.18},
		LineWidth: 1.2,
	}
	if len(opts) > 0 {
		config = mergeEllipseSelectorOptions(config, opts[0])
	}
	sel := &EllipseSelector{
		EdgeColor: config.EdgeColor,
		FillColor: config.FillColor,
		LineWidth: config.LineWidth,
		z:         1200,
	}
	a.AddWidget(sel)
	return sel
}

func (e *EllipseSelector) OnSelect(cb EllipseSelectorCallback) WidgetCallbackID {
	if e == nil || cb == nil {
		return 0
	}
	return e.onSelect.add(cb)
}

func (e *EllipseSelector) RemoveOnSelect(id WidgetCallbackID) {
	if e == nil {
		return
	}
	e.onSelect.remove(id)
}

func (e *EllipseSelector) SetBounds(minV, maxV geom.Pt) bool {
	if e == nil || !isFinite(minV.X) || !isFinite(minV.Y) || !isFinite(maxV.X) || !isFinite(maxV.Y) {
		return false
	}
	minV, maxV = normalizedRect(minV, maxV)
	changed := !e.Active || e.Min != minV || e.Max != maxV
	e.Active = true
	e.Min = minV
	e.Max = maxV
	return changed
}

func (e *EllipseSelector) MoveBy(delta geom.Pt) bool {
	if e == nil || !e.Active || !isFinite(delta.X) || !isFinite(delta.Y) {
		return false
	}
	e.Min.X += delta.X
	e.Min.Y += delta.Y
	e.Max.X += delta.X
	e.Max.Y += delta.Y
	return true
}

func (e *EllipseSelector) Clear() bool {
	if e == nil || !e.Active {
		return false
	}
	e.Active = false
	e.Min = geom.Pt{}
	e.Max = geom.Pt{}
	return true
}

func (e *EllipseSelector) TriggerOnSelect() {
	if e == nil || !e.Active {
		return
	}
	e.onSelect.each(func(cb EllipseSelectorCallback) {
		cb(e, geom.Rect{Min: e.Min, Max: e.Max})
	})
}

func (e *EllipseSelector) Draw(r render.Renderer, ctx *DrawContext) {
	if e == nil || r == nil || ctx == nil || !e.Active {
		return
	}
	if e.Min.X == e.Max.X || e.Min.Y == e.Max.Y {
		return
	}
	min := geom.Pt{X: e.Min.X, Y: e.Min.Y}
	max := geom.Pt{X: e.Max.X, Y: e.Max.Y}
	centerData := geom.Pt{X: (min.X + max.X) / 2, Y: (min.Y + max.Y) / 2}
	center := ctx.DataToPixel.Apply(centerData)
	p1 := ctx.DataToPixel.Apply(geom.Pt{X: min.X, Y: centerData.Y})
	p2 := ctx.DataToPixel.Apply(geom.Pt{X: max.X, Y: centerData.Y})
	q1 := ctx.DataToPixel.Apply(geom.Pt{X: centerData.X, Y: min.Y})
	q2 := ctx.DataToPixel.Apply(geom.Pt{X: centerData.X, Y: max.Y})
	width := math.Abs(p2.X - p1.X)
	height := math.Abs(q2.Y - q1.Y)
	if width <= 0 || height <= 0 {
		return
	}
	path := ellipsePath(width, height)
	path = applyAffinePath(path, translateAffine(center))
	r.Path(path, &render.Paint{Fill: e.FillColor, Stroke: e.EdgeColor, LineWidth: pointsToPixels(ctx.RC, e.LineWidth)})
}

func (e *EllipseSelector) Bounds(ctx *DrawContext) geom.Rect {
	if e == nil || !e.Active || ctx == nil {
		return geom.Rect{}
	}
	return e.displayRect(ctx)
}

func (e *EllipseSelector) Contains(p geom.Pt, ctx *DrawContext) (bool, PickInfo) {
	if e == nil || !e.Active || ctx == nil {
		return false, PickInfo{}
	}
	if e.Min.X == e.Max.X || e.Min.Y == e.Max.Y {
		return false, PickInfo{}
	}
	data, ok := (&ctx.DataToPixel).Invert(p)
	if !ok {
		return false, PickInfo{}
	}
	cx := (e.Min.X + e.Max.X) / 2
	cy := (e.Min.Y + e.Max.Y) / 2
	rx := (e.Max.X - e.Min.X) / 2
	ry := (e.Max.Y - e.Min.Y) / 2
	nx := (data.X - cx) / rx
	ny := (data.Y - cy) / ry
	return nx*nx+ny*ny <= 1, PickInfo{}
}

func (e *EllipseSelector) displayRect(ctx *DrawContext) geom.Rect {
	p1 := ctx.DataToPixel.Apply(e.Min)
	p2 := ctx.DataToPixel.Apply(e.Max)
	return geom.Rect{Min: geom.Pt{X: math.Min(p1.X, p2.X), Y: math.Min(p1.Y, p2.Y)}, Max: geom.Pt{X: math.Max(p1.X, p2.X), Y: math.Max(p1.Y, p2.Y)}}
}

func (e *EllipseSelector) Z() float64   { return e.z }
func (e *EllipseSelector) WidgetLayer() {}

func mergeEllipseSelectorOptions(base, override EllipseSelectorOptions) EllipseSelectorOptions {
	if override.EdgeColor != (render.Color{}) {
		base.EdgeColor = override.EdgeColor
	}
	if override.FillColor != (render.Color{}) {
		base.FillColor = override.FillColor
	}
	if override.LineWidth > 0 {
		base.LineWidth = override.LineWidth
	}
	return base
}
