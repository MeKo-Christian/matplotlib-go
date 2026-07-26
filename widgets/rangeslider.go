package widgets

import (
	"fmt"
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// RangeSliderCallback receives the active range slider and updated low/high values.
type RangeSliderCallback func(*RangeSlider, float64, float64)

// RangeSliderOptions configures a RangeSlider widget artist.
type RangeSliderOptions = SliderOptions

// RangeSlider draws a two-handle slider control inside its owning axes.
type RangeSlider struct {
	Label   string
	Min     float64
	Max     float64
	Enabled bool
	// low and high are setter-owned: the setters keep the endpoints ordered
	// and route through SetRange so the on-changed callback fires once.
	low         float64
	high        float64
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

// NewRangeSlider adds a two-handle range slider widget artist to the axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func NewRangeSlider(a *core.Axes, label string, minValue, maxValue, low, high float64, opt RangeSliderOptions) *RangeSlider {
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
	mergeSliderOptions(&cfg, &opt)
	step := (maxValue - minValue) / 100
	if v, ok := cfg.ValueStep.Get(); ok && v > 0 {
		step = cfg.ValueStep.OrZero()
	}
	if step == 0 {
		step = 1
	}
	valueFormat := "%.2f"
	if cfg.ValueFormat.IsSet() && strings.TrimSpace(cfg.ValueFormat.OrZero()) != "" {
		valueFormat = cfg.ValueFormat.OrZero()
	}
	low = normalizeSliderValue(minValue, maxValue, math.Abs(step), low)
	high = normalizeSliderValue(minValue, maxValue, math.Abs(step), high)
	if low > high {
		low, high = high, low
	}
	prepareWidgetAxes(a)
	w := &RangeSlider{
		Label:       label,
		Min:         minValue,
		Max:         maxValue,
		Enabled:     true,
		low:         low,
		high:        high,
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
	a.Add(w)
	return w
}

func (s *RangeSlider) OnChanged(cb RangeSliderCallback) WidgetCallbackID {
	if s == nil || cb == nil {
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
	s.onChanged.each(func(cb RangeSliderCallback) { cb(s, s.low, s.high) })
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
	if low == s.low && high == s.high {
		return
	}
	s.low = low
	s.high = high
	s.triggerOnChanged()
}

// SetLow updates the lower range endpoint without crossing the high endpoint.
func (s *RangeSlider) SetLow(low float64) {
	if s == nil {
		return
	}
	low = math.Min(low, s.high)
	s.SetRange(low, s.high)
}

// SetHigh updates the upper range endpoint without crossing the low endpoint.
func (s *RangeSlider) SetHigh(high float64) {
	if s == nil {
		return
	}
	high = math.Max(high, s.low)
	s.SetRange(s.low, high)
}

func (s *RangeSlider) Draw(r render.Renderer, ctx *core.DrawContext) {
	if s == nil || r == nil || ctx == nil {
		return
	}
	defaults := widgetDefaultsForRC(&ctx.RC)
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
	drawWidgetText(r, ctx, valueAnchor, rangeSliderDisplayValue(s, &defaults), fontSize, textColor, defaults.SliderValueAlign, defaults.SliderTextVAlign)

	track := widgetStyledSliderTrack(panel, &defaults)
	trackRadius := widgetSliderTrackRadius(track, &defaults)
	drawWidgetPanel(r, track, s.TrackColor, render.Color{A: 0}, 0, trackRadius)
	lowFraction := sliderFraction(s.Min, s.Max, s.low)
	highFraction := sliderFraction(s.Min, s.Max, s.high)
	fill := track
	fill.Min.X = track.Min.X + track.W()*lowFraction
	fill.Max.X = track.Min.X + track.W()*highFraction
	drawWidgetPanel(r, fill, s.FillColor, render.Color{A: 0}, 0, trackRadius)

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

func (s *RangeSlider) Bounds(ctx *core.DrawContext) geom.Rect {
	if s == nil || ctx == nil {
		return geom.Rect{}
	}
	defaults := widgetDefaultsForRC(&ctx.RC)
	return widgetStyledPanelRect(ctx.Clip, defaults.SliderPanelPad)
}

func (s *RangeSlider) Contains(p geom.Pt, ctx *core.DrawContext) (bool, core.PickInfo) {
	if s == nil || ctx == nil {
		return false, core.PickInfo{}
	}
	bounds := s.Bounds(ctx)
	if !bounds.Contains(p) {
		return false, core.PickInfo{}
	}
	defaults := widgetDefaultsForRC(&ctx.RC)
	track := widgetStyledSliderTrack(bounds, &defaults)
	if track.W() <= 0 {
		return true, core.PickInfo{}
	}
	lowX := track.Min.X + track.W()*sliderFraction(s.Min, s.Max, s.low)
	highX := track.Min.X + track.W()*sliderFraction(s.Min, s.Max, s.high)
	if math.Abs(p.X-lowX) <= math.Abs(p.X-highX) {
		return true, core.PickInfo{Index: 0}
	}
	return true, core.PickInfo{Index: 1}
}

// ValueForPoint maps a figure-pixel point to the range-slider value using the
// same visual-style track geometry used for drawing.
func (s *RangeSlider) ValueForPoint(p geom.Pt, ctx *core.DrawContext) (float64, bool) {
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

func (s *RangeSlider) Z() float64   { return s.z }
func (s *RangeSlider) WidgetLayer() {}

func rangeSliderDisplayValue(s *RangeSlider, defaults *widgetVisualDefaults) string {
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
		return "(" + formatOne(s.low) + ", " + formatOne(s.high) + ")"
	}
	return formatOne(s.low) + " - " + formatOne(s.high)
}

// Low reports the lower range endpoint.
func (s *RangeSlider) Low() float64 {
	if s == nil {
		return 0
	}
	return s.low
}

// High reports the upper range endpoint.
func (s *RangeSlider) High() float64 {
	if s == nil {
		return 0
	}
	return s.high
}
