package core

import (
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// RadioButtonsCallback receives the active radio widget and selected index.
type RadioButtonsCallback func(*RadioButtons, int)

// RadioButtonsOptions configures a RadioButtons widget artist.
type RadioButtonsOptions struct {
	FaceColor render.Color
	EdgeColor render.Color
	TextColor render.Color
	DotColor  render.Color
	FontSize  float64
}

// RadioButtons draws a static radio-button control.
type RadioButtons struct {
	Labels    []string
	Active    int
	FaceColor render.Color
	EdgeColor render.Color
	TextColor render.Color
	DotColor  render.Color
	FontSize  float64
	focus     int

	onChanged widgetCallbackRegistry[RadioButtonsCallback]

	z float64
}

// RadioButtons adds a radio-button widget artist to the axes.
func (a *Axes) RadioButtons(labels []string, active int, opts ...RadioButtonsOptions) *RadioButtons {
	if a == nil || len(labels) == 0 {
		return nil
	}
	cfg := RadioButtonsOptions{
		FaceColor: render.Color{R: 0.96, G: 0.97, B: 0.98, A: 1},
		EdgeColor: render.Color{R: 0.74, G: 0.76, B: 0.80, A: 1},
		TextColor: render.Color{R: 0.12, G: 0.13, B: 0.16, A: 1},
		DotColor:  render.Color{R: 0.85, G: 0.32, B: 0.17, A: 1},
	}
	if len(opts) > 0 {
		cfg = mergeRadioButtonsOptions(cfg, opts[0])
	}
	prepareWidgetAxes(a)
	w := &RadioButtons{
		Labels:    append([]string(nil), labels...),
		Active:    clampInt(active, 0, len(labels)-1),
		FaceColor: cfg.FaceColor,
		EdgeColor: cfg.EdgeColor,
		TextColor: cfg.TextColor,
		DotColor:  cfg.DotColor,
		focus:     clampInt(active, 0, len(labels)-1),
		FontSize:  cfg.FontSize,
		z:         1200,
	}
	a.AddWidget(w)
	return w
}

func (r *RadioButtons) OnChanged(cb RadioButtonsCallback) WidgetCallbackID {
	if r == nil || any(cb) == nil {
		return 0
	}
	return r.onChanged.add(cb)
}

func (r *RadioButtons) RemoveOnChanged(id WidgetCallbackID) {
	if r == nil {
		return
	}
	r.onChanged.remove(id)
}

func (r *RadioButtons) triggerOnChanged(active int) {
	if r == nil {
		return
	}
	r.onChanged.each(func(cb RadioButtonsCallback) { cb(r, active) })
}

// SetActive updates the selected radio index and emits an on-change event when
// the value mutates.
func (r *RadioButtons) SetActive(index int) {
	if r == nil || len(r.Labels) == 0 {
		return
	}
	index = clampInt(index, 0, len(r.Labels)-1)
	if r.Active == index {
		return
	}
	r.Active = index
	r.triggerOnChanged(index)
}

// Next decrements or increments the active radio index by one and emits a
// change event when the index changes.
func (r *RadioButtons) Next(delta int) {
	if r == nil || len(r.Labels) == 0 {
		return
	}
	if delta == 0 {
		return
	}
	r.SetActive(r.Active + delta)
}

func (rdo *RadioButtons) Contains(p geom.Pt, ctx *DrawContext) (bool, PickInfo) {
	if rdo == nil || ctx == nil {
		return false, PickInfo{}
	}
	if len(rdo.Labels) == 0 {
		return false, PickInfo{}
	}
	panel := rdo.Bounds(ctx)
	if !panel.Contains(p) {
		return false, PickInfo{}
	}
	rowHeight := panel.H() / float64(len(rdo.Labels))
	if rowHeight <= 0 {
		return false, PickInfo{}
	}
	row := int((p.Y - panel.Min.Y) / rowHeight)
	if row < 0 || row >= len(rdo.Labels) {
		return false, PickInfo{}
	}
	return true, PickInfo{Index: row}
}

func (rdo *RadioButtons) Draw(r render.Renderer, ctx *DrawContext) {
	if rdo == nil || r == nil || ctx == nil {
		return
	}
	panel := insetRect(ctx.Clip, 4)
	drawWidgetPanel(r, panel, rdo.FaceColor, rdo.EdgeColor, 1.1, 12)
	if len(rdo.Labels) == 0 {
		return
	}
	rowHeight := panel.H() / float64(len(rdo.Labels))
	fontSize := resolvedFontSize(rdo.FontSize, ctx)
	for i, label := range rdo.Labels {
		center := geom.Pt{X: panel.Min.X + 24, Y: panel.Min.Y + rowHeight*float64(i) + rowHeight/2}
		outer := ellipsePath(16, 16)
		outerPath := applyAffinePath(outer, patchAffine(center, 0))
		r.Path(outerPath, &render.Paint{
			Fill:      render.Color{R: 1, G: 1, B: 1, A: 1},
			Stroke:    rdo.EdgeColor,
			LineWidth: 1,
			LineJoin:  render.JoinRound,
			LineCap:   render.CapRound,
		})
		if i == clampInt(rdo.Active, 0, len(rdo.Labels)-1) {
			inner := ellipsePath(8, 8)
			innerPath := applyAffinePath(inner, patchAffine(center, 0))
			r.Path(innerPath, &render.Paint{Fill: rdo.DotColor})
		}
		drawWidgetText(r, ctx, geom.Pt{X: center.X + 16, Y: center.Y}, label, fontSize, rdo.TextColor, TextAlignLeft, textLayoutVAlignCenter)
	}
}

func (rdo *RadioButtons) Bounds(ctx *DrawContext) geom.Rect {
	if rdo == nil || ctx == nil {
		return geom.Rect{}
	}
	return insetRect(ctx.Clip, 4)
}
func (rdo *RadioButtons) Z() float64   { return rdo.z }
func (rdo *RadioButtons) WidgetLayer() {}

func mergeRadioButtonsOptions(base, override RadioButtonsOptions) RadioButtonsOptions {
	if override.FaceColor != (render.Color{}) {
		base.FaceColor = override.FaceColor
	}
	if override.EdgeColor != (render.Color{}) {
		base.EdgeColor = override.EdgeColor
	}
	if override.TextColor != (render.Color{}) {
		base.TextColor = override.TextColor
	}
	if override.DotColor != (render.Color{}) {
		base.DotColor = override.DotColor
	}
	if override.FontSize > 0 {
		base.FontSize = override.FontSize
	}
	return base
}
