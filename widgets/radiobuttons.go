package widgets

import (
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/optarg"
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
	Disabled  *bool
}

// RadioButtons draws a static radio-button control.
type RadioButtons struct {
	Labels    []string
	Active    int
	Enabled   bool
	FaceColor render.Color
	EdgeColor render.Color
	TextColor render.Color
	DotColor  render.Color
	FontSize  float64
	focus     int

	onChanged widgetCallbackRegistry[RadioButtonsCallback]

	z float64
}

// NewRadioButtons adds a radio-button widget artist to the axes.
func NewRadioButtons(a *core.Axes, labels []string, active int, opts ...RadioButtonsOptions) *RadioButtons {
	if a == nil || len(labels) == 0 {
		return nil
	}
	defaults := widgetDefaultsForAxes(a)
	cfg := RadioButtonsOptions{
		FaceColor: defaults.PanelFace,
		EdgeColor: defaults.PanelEdge,
		TextColor: defaults.Text,
		DotColor:  defaults.RadioDot,
	}
	if opt, ok := optarg.Optional("radiobuttons", opts); ok {
		mergeRadioButtonsOptions(&cfg, &opt)
	}
	prepareWidgetAxes(a)
	enabled := true
	if cfg.Disabled != nil {
		enabled = !*cfg.Disabled
	}
	w := &RadioButtons{
		Labels:    append([]string(nil), labels...),
		Active:    clampInt(active, 0, len(labels)-1),
		Enabled:   enabled,
		FaceColor: cfg.FaceColor,
		EdgeColor: cfg.EdgeColor,
		TextColor: cfg.TextColor,
		DotColor:  cfg.DotColor,
		focus:     clampInt(active, 0, len(labels)-1),
		FontSize:  cfg.FontSize,
		z:         1200,
	}
	a.Add(w)
	return w
}

func (r *RadioButtons) OnChanged(cb RadioButtonsCallback) WidgetCallbackID {
	if r == nil || cb == nil {
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
	if !r.Enabled {
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

func (rdo *RadioButtons) Contains(p geom.Pt, ctx *core.DrawContext) (bool, core.PickInfo) {
	if rdo == nil || ctx == nil {
		return false, core.PickInfo{}
	}
	if !rdo.Enabled {
		return false, core.PickInfo{}
	}
	if len(rdo.Labels) == 0 {
		return false, core.PickInfo{}
	}
	panel := rdo.Bounds(ctx)
	if !panel.Contains(p) {
		return false, core.PickInfo{}
	}
	rowHeight := panel.H() / float64(len(rdo.Labels))
	if rowHeight <= 0 {
		return false, core.PickInfo{}
	}
	// Display space is y-up: row 0 is drawn at the top (panel.Max.Y).
	row := int((panel.Max.Y - p.Y) / rowHeight)
	if row < 0 || row >= len(rdo.Labels) {
		return false, core.PickInfo{}
	}
	return true, core.PickInfo{Index: row}
}

func (rdo *RadioButtons) Draw(r render.Renderer, ctx *core.DrawContext) {
	if rdo == nil || r == nil || ctx == nil {
		return
	}
	defaults := widgetDefaultsForRC(&ctx.RC)
	panel := widgetStyledPanelRect(ctx.Clip, defaults.PanelPad)
	face := rdo.FaceColor
	edge := rdo.EdgeColor
	textColor := rdo.TextColor
	dotColor := rdo.DotColor
	if !rdo.Enabled {
		face = mixColor(face, render.Color{R: 1, G: 1, B: 1, A: 1}, 0.45)
		edge = mixColor(edge, render.Color{R: 1, G: 1, B: 1, A: 1}, 0.6)
		textColor = mixColor(textColor, render.Color{R: 1, G: 1, B: 1, A: 1}, 0.35)
		dotColor = mixColor(dotColor, render.Color{R: 1, G: 1, B: 1, A: 1}, 0.35)
	}
	drawWidgetPanel(r, panel, face, edge, defaults.PanelLineWidth, defaults.PanelRadius)
	if len(rdo.Labels) == 0 {
		return
	}
	fontSize := resolvedFontSize(rdo.FontSize, ctx)
	for i, label := range rdo.Labels {
		center := geom.Pt{
			X: widgetStyleCoord(panel.Min.X, panel.Max.X, defaults.RadioCenterXPad),
			Y: widgetButtonRowCenterY(panel, i, len(rdo.Labels), defaults.RadioCenterXPad),
		}
		outer := ellipsePath(defaults.RadioOuterSize, defaults.RadioOuterSize)
		outerPath := applyAffinePath(outer, patchAffine(center, 0))
		markerFace := defaults.RadioInactiveFace
		active := i == clampInt(rdo.Active, 0, len(rdo.Labels)-1)
		if active && defaults.RadioActiveOuter {
			markerFace = dotColor
		}
		r.Path(outerPath, &render.Paint{
			Fill:      markerFace,
			Stroke:    edge,
			LineWidth: defaults.RadioLineWidth,
			LineJoin:  render.JoinRound,
			LineCap:   render.CapRound,
		})
		if active && !defaults.RadioActiveOuter && defaults.RadioInnerSize > 0 {
			inner := ellipsePath(defaults.RadioInnerSize, defaults.RadioInnerSize)
			innerPath := applyAffinePath(inner, patchAffine(center, 0))
			r.Path(innerPath, &render.Paint{Fill: dotColor})
		}
		labelX := center.X + defaults.RadioLabelGap
		if defaults.RadioLabelGap >= 0 && defaults.RadioLabelGap <= 1 {
			labelX = widgetStyleCoord(panel.Min.X, panel.Max.X, defaults.RadioLabelGap)
		}
		drawWidgetText(r, ctx, geom.Pt{X: labelX, Y: center.Y}, label, fontSize, textColor, core.TextAlignLeft, core.TextVAlignMiddle)
	}
}

func (rdo *RadioButtons) Bounds(ctx *core.DrawContext) geom.Rect {
	if rdo == nil || ctx == nil {
		return geom.Rect{}
	}
	defaults := widgetDefaultsForRC(&ctx.RC)
	return widgetStyledPanelRect(ctx.Clip, defaults.PanelPad)
}
func (rdo *RadioButtons) Z() float64   { return rdo.z }
func (rdo *RadioButtons) WidgetLayer() {}

func mergeRadioButtonsOptions(base, override *RadioButtonsOptions) {
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
	if override.Disabled != nil {
		base.Disabled = override.Disabled
	}
}
