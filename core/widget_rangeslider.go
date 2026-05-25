package core

import (
	"fmt"
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// RangeSliderCallback receives the active range slider and updated low/high values.
type RangeSliderCallback func(*RangeSlider, float64, float64)

// RangeSliderOptions configures a RangeSlider widget artist.
type RangeSliderOptions = SliderOptions

// RangeSlider draws a two-handle slider control inside its owning axes.
type RangeSlider struct {
	Label       string
	Min         float64
	Max         float64
	Enabled     bool
	Low         float64
	High        float64
	FaceColor   render.Color
	TrackColor  render.Color
	FillColor   render.Color
	HandleColor render.Color
	TextColor   render.Color
	Step        float64
	ValueFormat string
	FontSize    float64
	Dragging    bool

	onChanged widgetCallbackRegistry[RangeSliderCallback]

	z float64
}

// RangeSlider adds a two-handle range slider widget artist to the axes.
func (a *Axes) RangeSlider(label string, min, max, low, high float64, opts ...RangeSliderOptions) *RangeSlider {
	if a == nil {
		return nil
	}
	defaults := widgetDefaultsForAxes(a)
	cfg := RangeSliderOptions{
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
	low = normalizeSliderValue(min, max, math.Abs(step), low)
	high = normalizeSliderValue(min, max, math.Abs(step), high)
	if low > high {
		low, high = high, low
	}
	prepareWidgetAxes(a)
	w := &RangeSlider{
		Label:       label,
		Min:         min,
		Max:         max,
		Enabled:     true,
		Low:         low,
		High:        high,
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
	a.AddWidget(w)
	return w
}

func (s *RangeSlider) OnChanged(cb RangeSliderCallback) WidgetCallbackID {
	if s == nil || any(cb) == nil {
		return 0
	}
	return s.onChanged.add(cb)
}

func (s *RangeSlider) RemoveOnChanged(id WidgetCallbackID) {
	if s == nil {
		return
	}
	s.onChanged.remove(id)
}

func (s *RangeSlider) triggerOnChanged() {
	if s == nil {
		return
	}
	s.onChanged.each(func(cb RangeSliderCallback) { cb(s, s.Low, s.High) })
}

// SetRange updates the selected interval and emits an on-changed event when
// either endpoint changes.
func (s *RangeSlider) SetRange(low, high float64) {
	if s == nil {
		return
	}
	low = normalizeSliderValue(s.Min, s.Max, s.Step, low)
	high = normalizeSliderValue(s.Min, s.Max, s.Step, high)
	if low > high {
		low, high = high, low
	}
	if low == s.Low && high == s.High {
		return
	}
	s.Low = low
	s.High = high
	s.triggerOnChanged()
}

// SetLow updates the lower range endpoint without crossing the high endpoint.
func (s *RangeSlider) SetLow(low float64) {
	if s == nil {
		return
	}
	low = math.Min(low, s.High)
	s.SetRange(low, s.High)
}

// SetHigh updates the upper range endpoint without crossing the low endpoint.
func (s *RangeSlider) SetHigh(high float64) {
	if s == nil {
		return
	}
	high = math.Max(high, s.Low)
	s.SetRange(s.Low, high)
}

func (s *RangeSlider) Draw(r render.Renderer, ctx *DrawContext) {
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
	drawWidgetText(r, ctx, valueAnchor, rangeSliderDisplayValue(s, defaults), fontSize, textColor, defaults.SliderValueAlign, defaults.SliderTextVAlign)

	track := widgetStyledSliderTrack(panel, defaults)
	drawWidgetPanel(r, track, s.TrackColor, render.Color{A: 0}, 0, track.H()/2)
	lowFraction := sliderFraction(s.Min, s.Max, s.Low)
	highFraction := sliderFraction(s.Min, s.Max, s.High)
	fill := track
	fill.Min.X = track.Min.X + track.W()*lowFraction
	fill.Max.X = track.Min.X + track.W()*highFraction
	drawWidgetPanel(r, fill, s.FillColor, render.Color{A: 0}, 0, fill.H()/2)

	for _, fraction := range []float64{lowFraction, highFraction} {
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
}

func (s *RangeSlider) Bounds(ctx *DrawContext) geom.Rect {
	if s == nil || ctx == nil {
		return geom.Rect{}
	}
	defaults := widgetDefaultsForRC(ctx.RC)
	return widgetStyledPanelRect(ctx.Clip, defaults.SliderPanelPad)
}

func (s *RangeSlider) Contains(p geom.Pt, ctx *DrawContext) (bool, PickInfo) {
	if s == nil || ctx == nil {
		return false, PickInfo{}
	}
	bounds := s.Bounds(ctx)
	if !bounds.Contains(p) {
		return false, PickInfo{}
	}
	defaults := widgetDefaultsForRC(ctx.RC)
	track := widgetStyledSliderTrack(bounds, defaults)
	if track.W() <= 0 {
		return true, PickInfo{}
	}
	lowX := track.Min.X + track.W()*sliderFraction(s.Min, s.Max, s.Low)
	highX := track.Min.X + track.W()*sliderFraction(s.Min, s.Max, s.High)
	if math.Abs(p.X-lowX) <= math.Abs(p.X-highX) {
		return true, PickInfo{Index: 0}
	}
	return true, PickInfo{Index: 1}
}

func (s *RangeSlider) Z() float64   { return s.z }
func (s *RangeSlider) WidgetLayer() {}

func rangeSliderDisplayValue(s *RangeSlider, defaults widgetVisualDefaults) string {
	if s == nil {
		return ""
	}
	format := strings.TrimSpace(s.ValueFormat)
	if format == "" {
		format = "%.2f"
	}
	formatOne := func(value float64) (out string) {
		defer func() {
			if recover() != nil {
				out = fmt.Sprintf("%.2f", value)
			}
		}()
		out = fmt.Sprintf(format, value)
		if strings.HasPrefix(out, "%!") {
			out = fmt.Sprintf("%.2f", value)
		}
		return out
	}
	if defaults.RangeTupleText {
		return "(" + formatOne(s.Low) + ", " + formatOne(s.High) + ")"
	}
	return formatOne(s.Low) + " - " + formatOne(s.High)
}
