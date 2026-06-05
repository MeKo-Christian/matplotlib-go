package canvas

import (
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
)

func (w *WidgetInteraction) handleSliderKey(slider *core.Slider, ev KeyEvent, key string) bool {
	if slider == nil || !slider.Enabled {
		return false
	}
	step := slider.Step
	if step <= 0 {
		step = 1
	}
	delta := 0.0
	switch key {
	case "left", "down":
		delta = -step
	case "right", "up":
		delta = step
	default:
		return false
	}
	if ev.Modifiers&ModifierControl != 0 {
		delta *= 10
	}
	before := slider.Value
	slider.SetValue(slider.Value + delta)
	return slider.Value != before
}

func (w *WidgetInteraction) handleRangeSliderKey(slider *core.RangeSlider, handle int, ev KeyEvent, key string) bool {
	if slider == nil || !slider.Enabled {
		return false
	}
	step := slider.Step
	if step <= 0 {
		step = 1
	}
	delta := 0.0
	switch key {
	case "left", "down":
		delta = -step
	case "right", "up":
		delta = step
	default:
		return false
	}
	if ev.Modifiers&ModifierControl != 0 {
		delta *= 10
	}
	beforeLow, beforeHigh := slider.Low, slider.High
	if handle <= 0 {
		slider.SetLow(slider.Low + delta)
	} else {
		slider.SetHigh(slider.High + delta)
	}
	return slider.Low != beforeLow || slider.High != beforeHigh
}

func (w *WidgetInteraction) setSliderValueFromPointLocked(slider *core.Slider, ax *Axes, position geom.Pt) {
	if slider == nil || ax == nil {
		return
	}
	ctx := core.AxesDrawContext(ax, w.figure)
	if ctx == nil {
		return
	}
	panel := widgetInsetRect(ctx.Clip, 4)
	if panel.W() <= 0 {
		return
	}
	track := geom.Rect{
		Min: geom.Pt{X: panel.Min.X + 14, Y: panel.Max.Y - 26},
		Max: geom.Pt{X: panel.Max.X - 14, Y: panel.Max.Y - 14},
	}
	if track.W() <= 0 {
		return
	}
	v := slider.Min + (slider.Max-slider.Min)*((position.X-track.Min.X)/track.W())
	slider.SetValue(v)
}

func (w *WidgetInteraction) setRangeSliderValueFromPointLocked(slider *core.RangeSlider, ax *Axes, position geom.Pt, handle int) {
	if slider == nil || ax == nil {
		return
	}
	ctx := core.AxesDrawContext(ax, w.figure)
	if ctx == nil {
		return
	}
	panel := widgetInsetRect(ctx.Clip, 4)
	if panel.W() <= 0 {
		return
	}
	track := geom.Rect{
		Min: geom.Pt{X: panel.Min.X + 14, Y: panel.Max.Y - 26},
		Max: geom.Pt{X: panel.Max.X - 14, Y: panel.Max.Y - 14},
	}
	if track.W() <= 0 {
		return
	}
	v := slider.Min + (slider.Max-slider.Min)*((position.X-track.Min.X)/track.W())
	if handle <= 0 {
		slider.SetLow(v)
		return
	}
	slider.SetHigh(v)
}
