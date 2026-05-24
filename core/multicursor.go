package core

import (
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// MultiCursorOptions configures a multi-axis hover cursor helper.
type MultiCursorOptions struct {
	Color      render.Color
	LineWidth  float64
	HorizOn    *bool
	VerticalOn *bool
}

// MultiCursor draws shared cross-lines across multiple axes.
type MultiCursor struct {
	Axes       []*Axes
	Color      render.Color
	LineWidth  float64
	Horizontal bool
	Vertical   bool

	FigureX, FigureY float64
	show             bool

	z float64
}

// MultiCursor creates a synchronized cross-axis cursor helper.
func (a *Axes) MultiCursor(axes ...*Axes) *MultiCursor {
	if a == nil {
		return nil
	}
	config := MultiCursorOptions{
		Color:      render.Color{R: 0.24, G: 0.24, B: 0.24, A: 1},
		LineWidth:  0.9,
		HorizOn:    boolPtr(true),
		VerticalOn: boolPtr(true),
	}
	selector := &MultiCursor{
		Axes:       dedupeAxes(append([]*Axes{a}, axes...)),
		Color:      config.Color,
		LineWidth:  config.LineWidth,
		Horizontal: boolValue(config.HorizOn, true),
		Vertical:   boolValue(config.VerticalOn, true),
		z:          1200,
	}
	a.AddWidget(selector)
	return selector
}

func (a *Axes) MultiCursorWithOptions(opts []MultiCursorOptions, axes ...*Axes) *MultiCursor {
	if a == nil {
		return nil
	}
	config := MultiCursorOptions{
		Color:      render.Color{R: 0.24, G: 0.24, B: 0.24, A: 1},
		LineWidth:  0.9,
		HorizOn:    boolPtr(true),
		VerticalOn: boolPtr(true),
	}
	if len(opts) > 0 {
		config = mergeMultiCursorOptions(config, opts[0])
	}
	mc := &MultiCursor{
		Axes:       dedupeAxes(append([]*Axes{a}, axes...)),
		Color:      config.Color,
		LineWidth:  config.LineWidth,
		Horizontal: boolValue(config.HorizOn, true),
		Vertical:   boolValue(config.VerticalOn, true),
		z:          1200,
	}
	a.AddWidget(mc)
	return mc
}

func (c *MultiCursor) SetFigurePoint(point geom.Pt) bool {
	if c == nil || !isFinite(point.X) || !isFinite(point.Y) {
		return false
	}
	changed := c.FigureX != point.X || c.FigureY != point.Y || !c.show
	c.FigureX = point.X
	c.FigureY = point.Y
	c.show = true
	return changed
}

func (c *MultiCursor) Hide() bool {
	if c == nil || !c.show {
		return false
	}
	c.show = false
	return true
}

func (c *MultiCursor) Position() (float64, float64, bool) {
	if c == nil || !c.show {
		return 0, 0, false
	}
	return c.FigureX, c.FigureY, true
}

func (c *MultiCursor) Draw(r render.Renderer, ctx *DrawContext) {
	if c == nil || r == nil || ctx == nil || !c.show || len(c.Axes) == 0 || !axesContains(c.Axes, ctx.Axes) {
		return
	}
	paint := render.Paint{Stroke: c.Color, LineWidth: c.LineWidth}
	if c.Vertical {
		r.Path(pixelLinePath(geom.Pt{X: c.FigureX, Y: ctx.Clip.Min.Y}, geom.Pt{X: c.FigureX, Y: ctx.Clip.Max.Y}), &paint)
	}
	if c.Horizontal {
		r.Path(pixelLinePath(geom.Pt{X: ctx.Clip.Min.X, Y: c.FigureY}, geom.Pt{X: ctx.Clip.Max.X, Y: c.FigureY}), &paint)
	}
}

func (c *MultiCursor) Bounds(ctx *DrawContext) geom.Rect {
	if c == nil || ctx == nil || !c.show || len(c.Axes) == 0 {
		return geom.Rect{}
	}
	fig := ctx.Axes.figure
	first := true
	var out geom.Rect
	for _, ax := range c.Axes {
		if ax == nil || ax.figure != fig {
			continue
		}
		clip := ax.adjustedLayout(fig)
		if first {
			out = clip
			first = false
			continue
		}
		out = unionRect(out, clip)
	}
	if first {
		return geom.Rect{}
	}
	return out
}

func (c *MultiCursor) Contains(geom.Pt, *DrawContext) (bool, PickInfo) {
	return false, PickInfo{}
}

func (c *MultiCursor) Z() float64   { return c.z }
func (c *MultiCursor) WidgetLayer() {}

func mergeMultiCursorOptions(base, override MultiCursorOptions) MultiCursorOptions {
	if override.Color != (render.Color{}) {
		base.Color = override.Color
	}
	if override.LineWidth > 0 {
		base.LineWidth = override.LineWidth
	}
	if override.HorizOn != nil {
		base.HorizOn = override.HorizOn
	}
	if override.VerticalOn != nil {
		base.VerticalOn = override.VerticalOn
	}
	return base
}
