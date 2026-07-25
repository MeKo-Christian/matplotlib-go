package widgets

import (
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// LassoSelectorCallback receives the lasso path in data coordinates.
type LassoSelectorCallback func(*LassoSelector, []geom.Pt)

// LassoSelectorOptions configures a lasso selector.
type LassoSelectorOptions struct {
	LineColor render.Color
	LineWidth float64
}

// LassoSelector stores a drawn freehand path in data coordinates.
type LassoSelector struct {
	Points    []geom.Pt
	LineColor render.Color
	LineWidth float64
	Active    bool
	Tracking  bool

	onSelect widgetCallbackRegistry[LassoSelectorCallback]
	z        float64
}

// NewLassoSelector creates a lasso selector bound to the axes.
func NewLassoSelector(a *core.Axes, opts ...LassoSelectorOptions) *LassoSelector {
	if a == nil {
		return nil
	}
	config := LassoSelectorOptions{
		LineColor: render.Color{R: 0.16, G: 0.42, B: 0.76, A: 1},
		LineWidth: 1.2,
	}
	if len(opts) > 0 {
		config = mergeLassoSelectorOptions(config, opts[0])
	}
	sel := &LassoSelector{
		LineColor: config.LineColor,
		LineWidth: config.LineWidth,
		z:         1200,
	}
	a.Add(sel)
	return sel
}

func (l *LassoSelector) OnSelect(cb LassoSelectorCallback) WidgetCallbackID {
	if l == nil || cb == nil {
		return 0
	}
	return l.onSelect.add(cb)
}

func (l *LassoSelector) RemoveOnSelect(id WidgetCallbackID) {
	if l == nil {
		return
	}
	l.onSelect.remove(id)
}

func (l *LassoSelector) Begin(point geom.Pt) bool {
	if l == nil || !isFinite(point.X) || !isFinite(point.Y) {
		return false
	}
	l.Active = true
	l.Tracking = true
	l.Points = []geom.Pt{point}
	return true
}

func (l *LassoSelector) AddPoint(point geom.Pt) bool {
	if l == nil || !l.Tracking || !isFinite(point.X) || !isFinite(point.Y) {
		return false
	}
	l.Points = append(l.Points, point)
	return true
}

func (l *LassoSelector) Finish() bool {
	if l == nil || !l.Tracking {
		return false
	}
	l.Tracking = false
	if len(l.Points) == 0 {
		l.Active = false
		return false
	}
	points := append([]geom.Pt(nil), l.Points...)
	l.Active = true
	l.onSelect.each(func(cb LassoSelectorCallback) {
		cb(l, points)
	})
	return true
}

func (l *LassoSelector) Clear() bool {
	if l == nil || (!l.Active && !l.Tracking) {
		return false
	}
	l.Active = false
	l.Tracking = false
	l.Points = nil
	return true
}

func (l *LassoSelector) Draw(r render.Renderer, ctx *core.DrawContext) {
	if l == nil || r == nil || ctx == nil || len(l.Points) < 2 {
		return
	}
	path := polygonSelectorPathFromData(l.Points, ctx)
	if len(path.V) == 0 {
		return
	}
	r.Path(path, &render.Paint{Stroke: l.LineColor, LineWidth: pointsToPixels(&ctx.RC, l.LineWidth)})
}

func (l *LassoSelector) Bounds(ctx *core.DrawContext) geom.Rect {
	if l == nil || ctx == nil || len(l.Points) < 2 {
		return geom.Rect{}
	}
	path := polygonSelectorPathFromData(l.Points, ctx)
	bounds, ok := pathBounds(path)
	if !ok {
		return geom.Rect{}
	}
	return bounds
}

func (l *LassoSelector) Contains(point geom.Pt, ctx *core.DrawContext) (bool, core.PickInfo) {
	if l == nil || ctx == nil || len(l.Points) < 2 {
		return false, core.PickInfo{}
	}
	path := polygonSelectorPathFromData(l.Points, ctx)
	for i := 1; i < len(path.V); i++ {
		if distancePointToSegment(path.V[i-1], path.V[i], point) <= core.DefaultPickRadius {
			return true, core.PickInfo{Index: i - 1}
		}
	}
	return false, core.PickInfo{}
}

func (l *LassoSelector) Z() float64   { return l.z }
func (l *LassoSelector) WidgetLayer() {}

func mergeLassoSelectorOptions(base, override LassoSelectorOptions) LassoSelectorOptions {
	if override.LineColor != (render.Color{}) {
		base.LineColor = override.LineColor
	}
	if override.LineWidth > 0 {
		base.LineWidth = override.LineWidth
	}
	return base
}
