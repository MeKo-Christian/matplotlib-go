package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// CheckButtonsCallback receives the active check button, changed index and
// state after the change.
type CheckButtonsCallback func(*CheckButtons, int, bool)

// CheckButtonsOptions configures a CheckButtons widget artist.
type CheckButtonsOptions struct {
	FaceColor  render.Color
	EdgeColor  render.Color
	TextColor  render.Color
	CheckColor render.Color
	FontSize   float64
	Disabled   *bool
}

// CheckButtons draws a static checklist-style control.
type CheckButtons struct {
	Labels     []string
	Values     []bool
	Enabled    bool
	FaceColor  render.Color
	EdgeColor  render.Color
	TextColor  render.Color
	CheckColor render.Color
	FontSize   float64
	focus      int

	onChanged widgetCallbackRegistry[CheckButtonsCallback]

	z float64
}

// CheckButtons adds a check-button widget artist to the axes.
func (a *Axes) CheckButtons(labels []string, active []bool, opts ...CheckButtonsOptions) *CheckButtons {
	if a == nil || len(labels) == 0 {
		return nil
	}
	defaults := widgetDefaultsForAxes(a)
	cfg := CheckButtonsOptions{
		FaceColor:  defaults.PanelFace,
		EdgeColor:  defaults.PanelEdge,
		TextColor:  defaults.Text,
		CheckColor: defaults.Check,
	}
	if len(opts) > 0 {
		cfg = mergeCheckButtonsOptions(cfg, opts[0])
	}
	prepareWidgetAxes(a)
	enabled := true
	if cfg.Disabled != nil {
		enabled = !*cfg.Disabled
	}
	values := make([]bool, len(labels))
	copy(values, active)
	w := &CheckButtons{
		Labels:     append([]string(nil), labels...),
		Values:     values,
		Enabled:    enabled,
		FaceColor:  cfg.FaceColor,
		EdgeColor:  cfg.EdgeColor,
		TextColor:  cfg.TextColor,
		CheckColor: cfg.CheckColor,
		focus:      -1,
		FontSize:   cfg.FontSize,
		z:          1200,
	}
	a.AddWidget(w)
	return w
}

func (c *CheckButtons) OnChanged(cb CheckButtonsCallback) WidgetCallbackID {
	if c == nil || any(cb) == nil {
		return 0
	}
	return c.onChanged.add(cb)
}

func (c *CheckButtons) RemoveOnChanged(id WidgetCallbackID) {
	if c == nil {
		return
	}
	c.onChanged.remove(id)
}

func (c *CheckButtons) triggerOnChanged(index int, value bool) {
	if c == nil {
		return
	}
	c.onChanged.each(func(cb CheckButtonsCallback) { cb(c, index, value) })
}

// SetValue updates one check-button state and emits an on-change event when
// it mutates.
func (c *CheckButtons) SetValue(index int, checked bool) {
	if c == nil {
		return
	}
	if !c.Enabled {
		return
	}
	if index < 0 || index >= len(c.Values) {
		return
	}
	if c.Values[index] == checked {
		return
	}
	c.Values[index] = checked
	c.triggerOnChanged(index, checked)
}

// Toggle flips one check-button state.
func (c *CheckButtons) Toggle(index int) {
	if c == nil || index < 0 || index >= len(c.Values) {
		return
	}
	c.SetValue(index, !c.Values[index])
}

func (c *CheckButtons) Contains(p geom.Pt, ctx *DrawContext) (bool, PickInfo) {
	if c == nil || ctx == nil {
		return false, PickInfo{}
	}
	if !c.Enabled {
		return false, PickInfo{}
	}
	if len(c.Labels) == 0 {
		return false, PickInfo{}
	}
	panel := c.Bounds(ctx)
	if !panel.Contains(p) {
		return false, PickInfo{}
	}
	rowHeight := panel.H() / float64(len(c.Labels))
	if rowHeight <= 0 {
		return false, PickInfo{}
	}
	// Display space is y-up: row 0 is drawn at the top (panel.Max.Y).
	row := int((panel.Max.Y - p.Y) / rowHeight)
	if row < 0 || row >= len(c.Labels) {
		return false, PickInfo{}
	}
	return true, PickInfo{Index: row}
}

func (c *CheckButtons) Draw(r render.Renderer, ctx *DrawContext) {
	if c == nil || r == nil || ctx == nil {
		return
	}
	defaults := widgetDefaultsForRC(ctx.RC)
	panel := widgetStyledPanelRect(ctx.Clip, defaults.PanelPad)
	face := c.FaceColor
	edge := c.EdgeColor
	textColor := c.TextColor
	checkColor := c.CheckColor
	if !c.Enabled {
		face = mixColor(face, render.Color{R: 1, G: 1, B: 1, A: 1}, 0.45)
		edge = mixColor(edge, render.Color{R: 1, G: 1, B: 1, A: 1}, 0.6)
		textColor = mixColor(textColor, render.Color{R: 1, G: 1, B: 1, A: 1}, 0.35)
		checkColor = mixColor(checkColor, render.Color{R: 1, G: 1, B: 1, A: 1}, 0.35)
	}
	drawWidgetPanel(r, panel, face, edge, defaults.PanelLineWidth, defaults.PanelRadius)
	if len(c.Labels) == 0 {
		return
	}
	rowHeight := panel.H() / float64(len(c.Labels))
	fontSize := resolvedFontSize(c.FontSize, ctx)
	for i, label := range c.Labels {
		// y-up: stack row 0 at the top (panel.Max.Y) downward.
		rowMaxY := panel.Max.Y - rowHeight*float64(i)
		rowMinY := rowMaxY - rowHeight
		boxSize := math.Min(defaults.CheckBoxMaxSize, rowHeight*defaults.CheckBoxScale)
		box := geom.Rect{
			Min: geom.Pt{X: panel.Min.X + defaults.CheckBoxXPad, Y: rowMinY + (rowHeight-boxSize)/2},
			Max: geom.Pt{X: panel.Min.X + defaults.CheckBoxXPad + boxSize, Y: rowMaxY - (rowHeight-boxSize)/2},
		}
		drawWidgetPanel(r, box, render.Color{R: 1, G: 1, B: 1, A: 1}, edge, defaults.CheckBoxLineWidth, defaults.CheckBoxRadius)
		if i < len(c.Values) && c.Values[i] {
			path := geom.Path{}
			path.MoveTo(geom.Pt{X: box.Min.X + box.W()*0.18, Y: box.Min.Y + box.H()*0.56})
			path.LineTo(geom.Pt{X: box.Min.X + box.W()*0.42, Y: box.Max.Y - box.H()*0.20})
			path.LineTo(geom.Pt{X: box.Max.X - box.W()*0.16, Y: box.Min.Y + box.H()*0.22})
			r.Path(path, &render.Paint{
				Stroke:    checkColor,
				LineWidth: defaults.CheckMarkWidth,
				LineJoin:  render.JoinRound,
				LineCap:   render.CapRound,
			})
		}
		drawWidgetText(r, ctx, geom.Pt{X: box.Max.X + defaults.CheckLabelGap, Y: rowMinY + rowHeight/2}, label, fontSize, textColor, TextAlignLeft, textLayoutVAlignCenter)
	}
}

func (c *CheckButtons) Bounds(ctx *DrawContext) geom.Rect {
	if c == nil || ctx == nil {
		return geom.Rect{}
	}
	defaults := widgetDefaultsForRC(ctx.RC)
	return widgetStyledPanelRect(ctx.Clip, defaults.PanelPad)
}
func (c *CheckButtons) Z() float64   { return c.z }
func (c *CheckButtons) WidgetLayer() {}

func mergeCheckButtonsOptions(base, override CheckButtonsOptions) CheckButtonsOptions {
	if override.FaceColor != (render.Color{}) {
		base.FaceColor = override.FaceColor
	}
	if override.EdgeColor != (render.Color{}) {
		base.EdgeColor = override.EdgeColor
	}
	if override.TextColor != (render.Color{}) {
		base.TextColor = override.TextColor
	}
	if override.CheckColor != (render.Color{}) {
		base.CheckColor = override.CheckColor
	}
	if override.FontSize > 0 {
		base.FontSize = override.FontSize
	}
	if override.Disabled != nil {
		base.Disabled = override.Disabled
	}
	return base
}
