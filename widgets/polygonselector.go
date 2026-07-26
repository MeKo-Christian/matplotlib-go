package widgets

import (
	"math"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// PolygonSelectorCallback receives the polygon vertices in data coordinates.
type PolygonSelectorCallback func(*PolygonSelector, []geom.Pt)

// PolygonSelectorOptions configures a polygon selector.
type PolygonSelectorOptions struct {
	EdgeColor render.Color
	FillColor render.Color
	LineWidth float64
}

// PolygonSelector stores a selected polygon path in data coordinates.
type PolygonSelector struct {
	Points    []geom.Pt
	EdgeColor render.Color
	FillColor render.Color
	LineWidth float64
	Closed    bool
	Active    bool

	onSelect widgetCallbackRegistry[PolygonSelectorCallback]
	z        float64
}

// NewPolygonSelector creates a polygon selector bound to the axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func NewPolygonSelector(a *core.Axes, opt PolygonSelectorOptions) *PolygonSelector {
	if a == nil {
		return nil
	}
	config := PolygonSelectorOptions{
		EdgeColor: render.Color{R: 0.16, G: 0.42, B: 0.76, A: 1},
		FillColor: render.Color{R: 0.16, G: 0.42, B: 0.76, A: 0.18},
		LineWidth: 1.2,
	}
	config = mergePolygonSelectorOptions(config, opt)
	sel := &PolygonSelector{
		EdgeColor: config.EdgeColor,
		FillColor: config.FillColor,
		LineWidth: config.LineWidth,
		z:         1200,
	}
	a.Add(sel)
	return sel
}

func (p *PolygonSelector) OnSelect(cb PolygonSelectorCallback) WidgetCallbackID {
	if p == nil || cb == nil {
		return 0
	}
	return p.onSelect.add(cb)
}

func (p *PolygonSelector) RemoveOnSelect(id WidgetCallbackID) {
	if p == nil {
		return
	}
	p.onSelect.remove(id)
}

func (p *PolygonSelector) AppendPoint(point geom.Pt) bool {
	if p == nil || !isFinite(point.X) || !isFinite(point.Y) {
		return false
	}
	p.Points = append(p.Points, point)
	p.Active = true
	return true
}

func (p *PolygonSelector) SetPoint(index int, point geom.Pt) bool {
	if p == nil || !isFinite(point.X) || !isFinite(point.Y) || index < 0 || index >= len(p.Points) {
		return false
	}
	p.Points[index] = point
	p.Active = true
	return true
}

func (p *PolygonSelector) MovePointBy(index int, delta geom.Pt) bool {
	if p == nil || !isFinite(delta.X) || !isFinite(delta.Y) || index < 0 || index >= len(p.Points) {
		return false
	}
	p.Points[index].X += delta.X
	p.Points[index].Y += delta.Y
	p.Active = true
	return true
}

func (p *PolygonSelector) MoveBy(delta geom.Pt) bool {
	if p == nil || !p.Active || len(p.Points) == 0 || !isFinite(delta.X) || !isFinite(delta.Y) {
		return false
	}
	for i := range p.Points {
		p.Points[i].X += delta.X
		p.Points[i].Y += delta.Y
	}
	return true
}

func (p *PolygonSelector) Close() bool {
	if p == nil || len(p.Points) < 3 {
		return false
	}
	p.Closed = true
	p.Active = true
	return true
}

func (p *PolygonSelector) Clear() bool {
	if p == nil || !p.Active {
		return false
	}
	p.Active = false
	p.Closed = false
	p.Points = nil
	return true
}

func (p *PolygonSelector) TriggerOnSelect() {
	if p == nil || !p.Active || len(p.Points) < 3 {
		return
	}
	points := append([]geom.Pt(nil), p.Points...)
	p.onSelect.each(func(cb PolygonSelectorCallback) {
		cb(p, points)
	})
}

func (p *PolygonSelector) Draw(r render.Renderer, ctx *core.DrawContext) {
	if p == nil || r == nil || ctx == nil || len(p.Points) == 0 {
		return
	}
	path := polygonSelectorPathFromData(p.Points, ctx)
	if len(path.V) == 0 {
		return
	}
	paint := &render.Paint{Stroke: p.EdgeColor, LineWidth: pointsToPixels(&ctx.RC, p.LineWidth)}
	if p.Closed {
		path.Close()
		paint.Fill = p.FillColor
	}
	r.Path(path, paint)
}

func (p *PolygonSelector) Bounds(ctx *core.DrawContext) geom.Rect {
	if p == nil || ctx == nil || len(p.Points) == 0 {
		return geom.Rect{}
	}
	path := polygonSelectorPathFromData(p.Points, ctx)
	bounds, ok := pathBounds(path)
	if !ok {
		return geom.Rect{}
	}
	return bounds
}

func (p *PolygonSelector) Contains(point geom.Pt, ctx *core.DrawContext) (bool, core.PickInfo) {
	if p == nil || ctx == nil || len(p.Points) == 0 {
		return false, core.PickInfo{}
	}
	pix := polygonSelectorPathFromData(p.Points, ctx)
	if len(pix.V) < 2 {
		return false, core.PickInfo{}
	}
	for i, vertex := range pix.V {
		if math.Hypot(point.X-vertex.X, point.Y-vertex.Y) <= core.DefaultPickRadius {
			return true, core.PickInfo{Index: i}
		}
	}
	for i := 1; i < len(pix.V); i++ {
		if distancePointToSegment(pix.V[i-1], pix.V[i], point) <= core.DefaultPickRadius {
			return true, core.PickInfo{Index: i - 1}
		}
	}
	if p.Closed && len(pix.V) >= 3 && pointInPolygon(point, pix.V) {
		return true, core.PickInfo{Index: 0}
	}
	return false, core.PickInfo{}
}

func (p *PolygonSelector) Z() float64   { return p.z }
func (p *PolygonSelector) WidgetLayer() {}

func mergePolygonSelectorOptions(base, override PolygonSelectorOptions) PolygonSelectorOptions {
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
