package widgets

import (
	"math"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// RectangleSelectorCallback receives the committed rectangle bounds in data coordinates.
type RectangleSelectorCallback func(*RectangleSelector, geom.Rect)

// RectangleSelectorOptions configures a rectangle selector.
type RectangleSelectorOptions struct {
	EdgeColor render.Color
	FillColor render.Color
	LineWidth float64
}

// RectangleSelector stores a selected rectangular region in data coordinates.
type RectangleSelector struct {
	Min       geom.Pt
	Max       geom.Pt
	EdgeColor render.Color
	FillColor render.Color
	LineWidth float64
	Active    bool

	onSelect widgetCallbackRegistry[RectangleSelectorCallback]
	z        float64
}

// NewRectangleSelector creates a rectangle selector bound to the axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func NewRectangleSelector(a *core.Axes, opt RectangleSelectorOptions) *RectangleSelector {
	if a == nil {
		return nil
	}
	config := RectangleSelectorOptions{
		EdgeColor: render.Color{R: 0.16, G: 0.42, B: 0.76, A: 1},
		FillColor: render.Color{R: 0.16, G: 0.42, B: 0.76, A: 0.18},
		LineWidth: 1.2,
	}
	config = mergeRectangleSelectorOptions(config, opt)
	sel := &RectangleSelector{
		EdgeColor: config.EdgeColor,
		FillColor: config.FillColor,
		LineWidth: config.LineWidth,
		z:         1200,
	}
	a.Add(sel)
	return sel
}

func (s *RectangleSelector) OnSelect(cb RectangleSelectorCallback) WidgetCallbackID {
	if s == nil || cb == nil {
		return 0
	}
	return s.onSelect.add(cb)
}

func (s *RectangleSelector) RemoveOnSelect(id WidgetCallbackID) {
	if s == nil {
		return
	}
	s.onSelect.remove(id)
}

func (s *RectangleSelector) SetBounds(minV, maxV geom.Pt) bool {
	if s == nil || !isFinite(minV.X) || !isFinite(minV.Y) || !isFinite(maxV.X) || !isFinite(maxV.Y) {
		return false
	}
	minV, maxV = normalizedRect(minV, maxV)
	changed := !s.Active || s.Min != minV || s.Max != maxV
	s.Active = true
	s.Min = minV
	s.Max = maxV
	return changed
}

func (s *RectangleSelector) MoveBy(delta geom.Pt) bool {
	if s == nil || !s.Active || !isFinite(delta.X) || !isFinite(delta.Y) {
		return false
	}
	s.Min.X += delta.X
	s.Min.Y += delta.Y
	s.Max.X += delta.X
	s.Max.Y += delta.Y
	return true
}

func (s *RectangleSelector) Clear() bool {
	if s == nil || !s.Active {
		return false
	}
	s.Active = false
	s.Min = geom.Pt{}
	s.Max = geom.Pt{}
	return true
}

func (s *RectangleSelector) TriggerOnSelect() {
	if s == nil || !s.Active {
		return
	}
	s.onSelect.each(func(cb RectangleSelectorCallback) {
		cb(s, s.BoundsRect())
	})
}

func (s *RectangleSelector) BoundsRect() geom.Rect {
	if s == nil || !s.Active {
		return geom.Rect{}
	}
	return geom.Rect{Min: s.Min, Max: s.Max}
}

func (s *RectangleSelector) Draw(r render.Renderer, ctx *core.DrawContext) {
	if s == nil || r == nil || ctx == nil || !s.Active {
		return
	}
	rect := s.displayRectFromData(ctx)
	if rect.W() <= 0 || rect.H() <= 0 {
		return
	}
	r.Path(pixelRectPath(rect), &render.Paint{Fill: s.FillColor, Stroke: s.EdgeColor, LineWidth: pointsToPixels(&ctx.RC, s.LineWidth)})
}

func (s *RectangleSelector) Bounds(ctx *core.DrawContext) geom.Rect {
	if s == nil || !s.Active || ctx == nil {
		return geom.Rect{}
	}
	return s.displayRectFromData(ctx)
}

func (s *RectangleSelector) Contains(p geom.Pt, ctx *core.DrawContext) (bool, core.PickInfo) {
	if s == nil || !s.Active || ctx == nil {
		return false, core.PickInfo{}
	}
	b := s.Bounds(ctx)
	if b.W() <= 0 || b.H() <= 0 {
		return false, core.PickInfo{}
	}
	if p.X < b.Min.X || p.X > b.Max.X || p.Y < b.Min.Y || p.Y > b.Max.Y {
		return false, core.PickInfo{}
	}
	return true, core.PickInfo{}
}

func (s *RectangleSelector) displayRectFromData(ctx *core.DrawContext) geom.Rect {
	if s == nil || ctx == nil || !s.Active {
		return geom.Rect{}
	}
	p1 := ctx.DataToPixel.Apply(s.Min)
	p2 := ctx.DataToPixel.Apply(s.Max)
	return geom.Rect{Min: geom.Pt{X: math.Min(p1.X, p2.X), Y: math.Min(p1.Y, p2.Y)}, Max: geom.Pt{X: math.Max(p1.X, p2.X), Y: math.Max(p1.Y, p2.Y)}}
}

func (s *RectangleSelector) Z() float64   { return s.z }
func (s *RectangleSelector) WidgetLayer() {}

func mergeRectangleSelectorOptions(base, override RectangleSelectorOptions) RectangleSelectorOptions {
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
