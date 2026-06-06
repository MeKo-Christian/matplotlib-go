package core

import (
	"fmt"
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

// SliderCallback receives the active slider and updated value.
type SliderCallback func(*Slider, float64)

// SliderOptions configures a Slider widget artist.
type SliderOptions struct {
	FaceColor   render.Color
	TrackColor  render.Color
	FillColor   render.Color
	HandleColor render.Color
	TextColor   render.Color
	ValueStep   *float64
	ValueFormat *string
	FontSize    float64
}

// Slider draws a static slider control inside its owning axes.
type Slider struct {
	Label       string
	Min         float64
	Max         float64
	Enabled     bool
	Value       float64
	Initial     float64
	FaceColor   render.Color
	TrackColor  render.Color
	FillColor   render.Color
	HandleColor render.Color
	TextColor   render.Color
	Step        float64
	ValueFormat string
	FontSize    float64
	Dragging    bool

	onChanged widgetCallbackRegistry[SliderCallback]

	z float64
}

// Slider adds a slider widget artist to the axes.
func (a *Axes) Slider(label string, min, max, value float64, opts ...SliderOptions) *Slider {
	if a == nil {
		return nil
	}
	defaults := widgetDefaultsForAxes(a)
	cfg := SliderOptions{
		FaceColor:   defaults.PanelFace,
		TrackColor:  defaults.Track,
		FillColor:   defaults.Fill,
		HandleColor: defaults.Handle,
		TextColor:   defaults.Text,
	}
	if len(opts) > 0 {
		cfg = mergeSliderOptions(cfg, opts[0])
	}
	step := (max - min) / 100
	if cfg.ValueStep != nil && *cfg.ValueStep > 0 {
		step = *cfg.ValueStep
	}
	if step == 0 {
		step = 1
	}
	valueFormat := "%.2f"
	if cfg.ValueFormat != nil && strings.TrimSpace(*cfg.ValueFormat) != "" {
		valueFormat = *cfg.ValueFormat
	}
	prepareWidgetAxes(a)
	w := &Slider{
		Label:       label,
		Min:         min,
		Max:         max,
		Enabled:     true,
		Value:       normalizeSliderValue(min, max, math.Abs(step), value),
		FaceColor:   cfg.FaceColor,
		TrackColor:  cfg.TrackColor,
		FillColor:   cfg.FillColor,
		HandleColor: cfg.HandleColor,
		TextColor:   cfg.TextColor,
		Step:        math.Abs(step),
		ValueFormat: valueFormat,
		FontSize:    cfg.FontSize,
		z:           1200,
	}
	w.Initial = w.Value
	a.AddWidget(w)
	return w
}

func (s *Slider) OnChanged(cb SliderCallback) WidgetCallbackID {
	if s == nil || any(cb) == nil {
		return 0
	}
	return s.onChanged.add(cb)
}

func (s *Slider) RemoveOnChanged(id WidgetCallbackID) {
	if s == nil {
		return
	}
	s.onChanged.remove(id)
}

func (s *Slider) triggerOnChanged() {
	if s == nil {
		return
	}
	s.onChanged.each(func(cb SliderCallback) { cb(s, s.Value) })
}

// SetValue updates the slider value and emits an on-changed event when the
// value changes.
func (s *Slider) SetValue(value float64) {
	if s == nil {
		return
	}
	clamped := normalizeSliderValue(s.Min, s.Max, s.Step, value)
	if clamped == s.Value {
		return
	}
	s.Value = clamped
	s.triggerOnChanged()
}

func (s *Slider) Draw(r render.Renderer, ctx *DrawContext) {
	if s == nil || r == nil || ctx == nil {
		return
	}
	defaults := widgetDefaultsForRC(ctx.RC)
	panel := widgetStyledPanelRect(ctx.Clip, defaults.SliderPanelPad)
	drawWidgetPanel(r, panel, s.FaceColor, render.Color{A: 0}, 0, defaults.SliderRadius)
	textColor := s.TextColor
	fontSize := resolvedFontSize(s.FontSize, ctx)
	labelAnchor := geom.Pt{
		X: widgetResolvedCoord(panel.Min.X, panel.Max.X, defaults.SliderLabelX),
		Y: widgetResolvedCoord(panel.Min.Y, panel.Max.Y, defaults.SliderLabelY),
	}
	valueAnchor := geom.Pt{
		X: widgetResolvedCoord(panel.Min.X, panel.Max.X, defaults.SliderValueX),
		Y: widgetResolvedCoord(panel.Min.Y, panel.Max.Y, defaults.SliderValueY),
	}
	drawWidgetText(r, ctx, labelAnchor, s.Label, fontSize, textColor, defaults.SliderLabelAlign, defaults.SliderTextVAlign)
	drawWidgetText(r, ctx, valueAnchor, sliderDisplayValue(s), fontSize, textColor, defaults.SliderValueAlign, defaults.SliderTextVAlign)

	track := widgetStyledSliderTrack(panel, defaults)
	trackRadius := widgetSliderTrackRadius(track, defaults)
	drawWidgetPanel(r, track, s.TrackColor, render.Color{A: 0}, 0, trackRadius)
	fraction := sliderFraction(s.Min, s.Max, s.Value)
	fill := track
	fill.Max.X = fill.Min.X + track.W()*fraction
	drawWidgetPanel(r, fill, s.FillColor, render.Color{A: 0}, 0, trackRadius)
	if defaults.SliderInitColor.A > 0 && defaults.SliderInitLine > 0 {
		initX := track.Min.X + track.W()*sliderFraction(s.Min, s.Max, s.Initial)
		r.Path(pixelLinePath(
			geom.Pt{X: initX, Y: track.Min.Y},
			geom.Pt{X: initX, Y: track.Max.Y},
		), &render.Paint{
			Stroke:    defaults.SliderInitColor,
			LineWidth: defaults.SliderInitLine,
		})
	}
	handleX := track.Min.X + track.W()*fraction
	handleSize := defaults.SliderHandleSize
	if handleSize <= 0 {
		handleSize = track.H()
	} else if handleSize < 5 {
		handleSize *= track.H()
	}
	handle := ellipsePath(handleSize, handleSize)
	handlePath := applyAffinePath(handle, patchAffine(geom.Pt{X: handleX, Y: track.Min.Y + track.H()/2}, 0))
	r.Path(handlePath, &render.Paint{
		Fill:      s.HandleColor,
		Stroke:    defaults.HandleEdge,
		LineWidth: defaults.SliderHandleLine,
		LineJoin:  render.JoinRound,
		LineCap:   render.CapRound,
	})
}

func (s *Slider) Bounds(ctx *DrawContext) geom.Rect {
	if s == nil || ctx == nil {
		return geom.Rect{}
	}
	defaults := widgetDefaultsForRC(ctx.RC)
	return widgetStyledPanelRect(ctx.Clip, defaults.SliderPanelPad)
}

func (s *Slider) Contains(p geom.Pt, ctx *DrawContext) (bool, PickInfo) {
	if s == nil || ctx == nil {
		return false, PickInfo{}
	}
	if s.Bounds(ctx).Contains(p) {
		return true, PickInfo{}
	}
	return false, PickInfo{}
}

// ValueForPoint maps a figure-pixel point to the slider value using the same
// visual-style track geometry used for drawing.
func (s *Slider) ValueForPoint(p geom.Pt, ctx *DrawContext) (float64, bool) {
	if s == nil || ctx == nil {
		return 0, false
	}
	track := widgetSliderTrackForContext(ctx)
	if track.W() <= 0 {
		return 0, false
	}
	value := s.Min + (s.Max-s.Min)*((p.X-track.Min.X)/track.W())
	return value, true
}

func (s *Slider) Z() float64   { return s.z }
func (s *Slider) WidgetLayer() {}

func sliderFraction(min, max, value float64) float64 {
	if max <= min {
		return 0
	}
	return clampFloat((value-min)/(max-min), 0, 1)
}

func widgetSliderTrack(panel geom.Rect) geom.Rect {
	return widgetStyledSliderTrack(panel, widgetDefaultsForRC(style.Default))
}

func sliderDisplayValue(s *Slider) string {
	if s == nil {
		return ""
	}
	format := strings.TrimSpace(s.ValueFormat)
	if format == "" {
		format = "%.2f"
	}
	formatted := func() string {
		defer func() {
			if recover() != nil {
				format = "%.2f"
			}
		}()
		return fmt.Sprintf(format, s.Value)
	}()
	if strings.HasPrefix(formatted, "%!") {
		return fmt.Sprintf("%.2f", s.Value)
	}
	return formatted
}

func clampSliderValue(min, max, value float64) float64 {
	if max <= min {
		return min
	}
	return math.Max(min, math.Min(max, value))
}

func normalizeSliderValue(min, max, step, value float64) float64 {
	clamped := clampSliderValue(min, max, value)
	if max <= min {
		return clamped
	}
	if step == 0 {
		return clamped
	}
	step = math.Abs(step)
	if !isFinite(step) || step == 0 {
		return clamped
	}
	based := (clamped - min) / step
	return clampSliderValue(min, max, min+math.Round(based)*step)
}

func mergeSliderOptions(base, override SliderOptions) SliderOptions {
	if override.FaceColor != (render.Color{}) {
		base.FaceColor = override.FaceColor
	}
	if override.TrackColor != (render.Color{}) {
		base.TrackColor = override.TrackColor
	}
	if override.FillColor != (render.Color{}) {
		base.FillColor = override.FillColor
	}
	if override.HandleColor != (render.Color{}) {
		base.HandleColor = override.HandleColor
	}
	if override.TextColor != (render.Color{}) {
		base.TextColor = override.TextColor
	}
	if override.FontSize > 0 {
		base.FontSize = override.FontSize
	}
	if override.ValueStep != nil {
		base.ValueStep = override.ValueStep
	}
	if override.ValueFormat != nil {
		base.ValueFormat = override.ValueFormat
	}
	return base
}
