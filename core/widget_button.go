package core

import (
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// ButtonCallback receives an active button widget.
type ButtonCallback func(*Button)

// ButtonOptions configures a Button widget artist.
type ButtonOptions struct {
	FaceColor render.Color
	EdgeColor render.Color
	TextColor render.Color
	Pressed   *bool
	Disabled  *bool
	FontSize  float64
}

// Button draws a static button-style control inside its owning axes.
type Button struct {
	Label     string
	FaceColor render.Color
	EdgeColor render.Color
	TextColor render.Color
	Pressed   bool
	Enabled   bool
	FontSize  float64
	Hovered   bool

	onClicked widgetCallbackRegistry[ButtonCallback]

	z float64
}

// Button adds a button widget artist to the axes.
func (a *Axes) Button(label string, opts ...ButtonOptions) *Button {
	if a == nil {
		return nil
	}
	cfg := ButtonOptions{
		FaceColor: render.Color{R: 0.94, G: 0.95, B: 0.97, A: 1},
		EdgeColor: render.Color{R: 0.74, G: 0.76, B: 0.80, A: 1},
		TextColor: render.Color{R: 0.12, G: 0.13, B: 0.16, A: 1},
	}
	if len(opts) > 0 {
		cfg = mergeButtonOptions(cfg, opts[0])
	}
	prepareWidgetAxes(a)
	enabled := true
	if cfg.Disabled != nil {
		enabled = !*cfg.Disabled
	}
	w := &Button{
		Label:     label,
		FaceColor: cfg.FaceColor,
		EdgeColor: cfg.EdgeColor,
		TextColor: cfg.TextColor,
		Pressed:   boolValue(cfg.Pressed, false),
		Enabled:   enabled,
		FontSize:  cfg.FontSize,
		z:         1200,
	}
	a.AddWidget(w)
	return w
}

func (b *Button) OnClicked(cb ButtonCallback) WidgetCallbackID {
	if b == nil || any(cb) == nil {
		return 0
	}
	return b.onClicked.add(cb)
}

func (b *Button) RemoveOnClicked(id WidgetCallbackID) {
	if b == nil {
		return
	}
	b.onClicked.remove(id)
}

func (b *Button) triggerOnClicked() {
	if b == nil {
		return
	}
	b.onClicked.each(func(cb ButtonCallback) { cb(b) })
}

// Click triggers all registered click callbacks.
func (b *Button) Click() {
	if b == nil || !b.Enabled {
		return
	}
	b.triggerOnClicked()
}

func (b *Button) Draw(r render.Renderer, ctx *DrawContext) {
	if b == nil || r == nil || ctx == nil {
		return
	}
	bounds := insetRect(ctx.Clip, 6)
	fill := b.FaceColor
	if !b.Enabled {
		fill = mixColor(fill, render.Color{R: 1, G: 1, B: 1, A: 1}, 0.45)
		edge := mixColor(b.EdgeColor, render.Color{R: 1, G: 1, B: 1, A: 1}, 0.6)
		drawWidgetPanel(r, bounds, fill, edge, 1.25, 10)
		drawCenteredWidgetText(r, ctx, geom.Pt{
			X: bounds.Min.X + bounds.W()/2,
			Y: bounds.Min.Y + bounds.H()/2,
		}, b.Label, b.FontSize, mixColor(b.TextColor, render.Color{R: 1, G: 1, B: 1, A: 1}, 0.35))
		return
	}
	if b.Pressed {
		fill = mixColor(fill, render.Color{R: 0, G: 0, B: 0, A: 1}, 0.12)
	}
	if b.Hovered && !b.Pressed {
		fill = mixColor(fill, render.Color{R: 1, G: 1, B: 1, A: 1}, 0.06)
	}
	drawWidgetPanel(r, bounds, fill, b.EdgeColor, 1.25, 10)
	labelColor := b.TextColor
	if b.Pressed {
		labelColor = mixColor(b.TextColor, render.Color{R: 0, G: 0, B: 0, A: 1}, 0.3)
	}
	drawCenteredWidgetText(r, ctx, geom.Pt{
		X: bounds.Min.X + bounds.W()/2,
		Y: bounds.Min.Y + bounds.H()/2,
	}, b.Label, b.FontSize, labelColor)
}

func (b *Button) Bounds(ctx *DrawContext) geom.Rect {
	if b == nil || ctx == nil {
		return geom.Rect{}
	}
	return insetRect(ctx.Clip, 6)
}

func (b *Button) Contains(p geom.Pt, ctx *DrawContext) (bool, PickInfo) {
	if b == nil || ctx == nil {
		return false, PickInfo{}
	}
	if !b.Enabled {
		return false, PickInfo{}
	}
	if b.Bounds(ctx).Contains(p) {
		return true, PickInfo{}
	}
	return false, PickInfo{}
}
func (b *Button) Z() float64   { return b.z }
func (b *Button) WidgetLayer() {}

func mergeButtonOptions(base, override ButtonOptions) ButtonOptions {
	if override.FaceColor != (render.Color{}) {
		base.FaceColor = override.FaceColor
	}
	if override.EdgeColor != (render.Color{}) {
		base.EdgeColor = override.EdgeColor
	}
	if override.TextColor != (render.Color{}) {
		base.TextColor = override.TextColor
	}
	if override.Pressed != nil {
		base.Pressed = override.Pressed
	}
	if override.Disabled != nil {
		base.Disabled = override.Disabled
	}
	if override.FontSize > 0 {
		base.FontSize = override.FontSize
	}
	return base
}
