package widgets

import (
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// CursorOptions configures a hover cursor helper.
type CursorOptions struct {
	Color      render.Color
	LineWidth  float64
	HorizOn    *bool
	VerticalOn *bool
}

// Cursor draws cursor cross-lines and is updated by hover events.
type Cursor struct {
	X, Y       float64
	Color      render.Color
	LineWidth  float64
	Horizontal bool
	Vertical   bool

	axis *core.Axes
	show bool

	z float64
}

// NewCursor creates a hover cursor for this axis.
func NewCursor(a *core.Axes, opts ...CursorOptions) *Cursor {
	if a == nil {
		return nil
	}
	config := CursorOptions{
		Color:      render.Color{R: 0.24, G: 0.24, B: 0.24, A: 1},
		LineWidth:  0.9,
		HorizOn:    boolPtr(true),
		VerticalOn: boolPtr(true),
	}
	if len(opts) > 0 {
		config = mergeCursorOptions(config, opts[0])
	}
	c := &Cursor{
		axis:       a,
		Color:      config.Color,
		LineWidth:  config.LineWidth,
		Horizontal: boolValue(config.HorizOn, true),
		Vertical:   boolValue(config.VerticalOn, true),
		z:          1200,
	}
	a.Add(c)
	return c
}

func (c *Cursor) SetData(x, y float64) bool {
	if c == nil || !isFinite(x) || !isFinite(y) {
		return false
	}
	changed := c.X != x || c.Y != y || !c.show
	c.X = x
	c.Y = y
	c.show = true
	return changed
}

func (c *Cursor) Hide() bool {
	if c == nil || !c.show {
		return false
	}
	c.show = false
	return true
}

func (c *Cursor) Position() (float64, float64, bool) {
	if c == nil || !c.show {
		return 0, 0, false
	}
	return c.X, c.Y, true
}

func (c *Cursor) Draw(r render.Renderer, ctx *core.DrawContext) {
	if c == nil || r == nil || ctx == nil || c.axis == nil || !c.show || ctx.Axes != c.axis {
		return
	}
	pt := ctx.DataToPixel.Apply(geom.Pt{X: c.X, Y: c.Y})
	paint := render.Paint{Stroke: c.Color, LineWidth: pointsToPixels(&ctx.RC, c.LineWidth)}
	if c.Vertical {
		r.Path(pixelLinePath(geom.Pt{X: pt.X, Y: ctx.Clip.Min.Y}, geom.Pt{X: pt.X, Y: ctx.Clip.Max.Y}), &paint)
	}
	if c.Horizontal {
		r.Path(pixelLinePath(geom.Pt{X: ctx.Clip.Min.X, Y: pt.Y}, geom.Pt{X: ctx.Clip.Max.X, Y: pt.Y}), &paint)
	}
}

func (c *Cursor) Bounds(ctx *core.DrawContext) geom.Rect {
	if c == nil || !c.show {
		return geom.Rect{}
	}
	return ctx.Clip
}

func (c *Cursor) Contains(geom.Pt, *core.DrawContext) (bool, core.PickInfo) {
	return false, core.PickInfo{}
}

func (c *Cursor) Z() float64   { return c.z }
func (c *Cursor) WidgetLayer() {}

func mergeCursorOptions(base, override CursorOptions) CursorOptions {
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
