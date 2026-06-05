package core

import (
	"fmt"
	"math"

	"github.com/cwbudde/matplotlib-go/transform"
)

func (a *Axes) TwinX() *Axes {
	twin := a.newOverlayAxes()
	if twin == nil {
		return nil
	}
	twin.shareX = a.xScaleRoot()
	if twin.XAxis != nil {
		twin.XAxis.ShowSpine = false
		twin.XAxis.ShowTicks = false
		twin.XAxis.ShowLabels = false
	}
	if twin.YAxis != nil {
		twin.YAxis.ShowSpine = false
		twin.YAxis.ShowTicks = false
		twin.YAxis.ShowLabels = false
	}
	twin.ShowFrame = false
	twin.RightAxis()
	return twin
}

func (a *Axes) TwinY() *Axes {
	twin := a.newOverlayAxes()
	if twin == nil {
		return nil
	}
	twin.shareY = a.yScaleRoot()
	if twin.YAxis != nil {
		twin.YAxis.ShowSpine = false
		twin.YAxis.ShowTicks = false
		twin.YAxis.ShowLabels = false
	}
	if twin.XAxis != nil {
		twin.XAxis.ShowSpine = false
		twin.XAxis.ShowTicks = false
		twin.XAxis.ShowLabels = false
	}
	twin.ShowFrame = false
	twin.TopAxis()
	return twin
}

func (a *Axes) SecondaryXAxis(side AxisSide, forward func(float64) float64, inverse func(float64) (float64, bool)) (*Axes, error) {
	if side != AxisTop && side != AxisBottom {
		return nil, fmt.Errorf("secondary x-axis side must be top or bottom")
	}
	return a.newSecondaryAxes(true, side, forward, inverse)
}

func (a *Axes) SecondaryYAxis(side AxisSide, forward func(float64) float64, inverse func(float64) (float64, bool)) (*Axes, error) {
	if side != AxisLeft && side != AxisRight {
		return nil, fmt.Errorf("secondary y-axis side must be left or right")
	}
	return a.newSecondaryAxes(false, side, forward, inverse)
}

func (a *Axes) newOverlayAxes() *Axes {
	if a == nil || a.figure == nil {
		return nil
	}
	overlay := a.figure.addAxesWithProjection(a.RectFraction, cloneProjection(a.projection))
	overlay.RC = a.RC
	overlay.aspectMode = a.aspectMode
	overlay.aspectValue = a.aspectValue
	overlay.boxAspect = a.boxAspect
	a.childAxes = append(a.childAxes, overlay)
	return overlay
}

func (a *Axes) newSecondaryAxes(isX bool, side AxisSide, forward func(float64) float64, inverse func(float64) (float64, bool)) (*Axes, error) {
	if a == nil || a.figure == nil {
		return nil, fmt.Errorf("secondary axes require a figure-backed axes")
	}
	if forward == nil || inverse == nil {
		return nil, fmt.Errorf("secondary axes require forward and inverse functions")
	}
	overlay := a.newOverlayAxes()
	if overlay == nil {
		return nil, fmt.Errorf("could not create overlay axes")
	}
	overlay.ShowFrame = false

	if isX {
		overlay.XScale = linkedSecondaryScale{parent: a, isX: true, forward: forward, inverse: inverse}
		overlay.shareY = a.yScaleRoot()
		if overlay.YAxis != nil {
			overlay.YAxis.ShowSpine = false
			overlay.YAxis.ShowTicks = false
			overlay.YAxis.ShowLabels = false
		}
		if overlay.XAxis != nil {
			overlay.XAxis.ShowSpine = false
			overlay.XAxis.ShowTicks = false
			overlay.XAxis.ShowLabels = false
		}
		var secondaryAxis *Axis
		if side == AxisTop {
			secondaryAxis = overlay.TopAxis()
		} else {
			overlay.XAxis = cloneAxisForSide(a.XAxis, AxisBottom)
			secondaryAxis = overlay.XAxis
		}
		if secondaryAxis != nil {
			secondaryAxis.ShowSpine = false
		}
	} else {
		overlay.YScale = linkedSecondaryScale{parent: a, isX: false, forward: forward, inverse: inverse}
		overlay.shareX = a.xScaleRoot()
		if overlay.XAxis != nil {
			overlay.XAxis.ShowSpine = false
			overlay.XAxis.ShowTicks = false
			overlay.XAxis.ShowLabels = false
		}
		if overlay.YAxis != nil {
			overlay.YAxis.ShowSpine = false
			overlay.YAxis.ShowTicks = false
			overlay.YAxis.ShowLabels = false
		}
		var secondaryAxis *Axis
		if side == AxisRight {
			secondaryAxis = overlay.RightAxis()
		} else {
			overlay.YAxis = cloneAxisForSide(a.YAxis, AxisLeft)
			secondaryAxis = overlay.YAxis
		}
		if secondaryAxis != nil {
			secondaryAxis.ShowSpine = false
		}
	}
	return overlay, nil
}

type linkedSecondaryScale struct {
	parent  *Axes
	isX     bool
	forward func(float64) float64
	inverse func(float64) (float64, bool)
}

func (s linkedSecondaryScale) primaryScale() transform.Scale {
	if s.parent == nil {
		return nil
	}
	if s.isX {
		return s.parent.effectiveXScale()
	}
	return s.parent.effectiveYScale()
}

func (s linkedSecondaryScale) Domain() (float64, float64) {
	base := s.primaryScale()
	if base == nil || s.forward == nil {
		return 0, 1
	}
	minVal, maxVal := base.Domain()
	return s.forward(minVal), s.forward(maxVal)
}

func (s linkedSecondaryScale) Fwd(x float64) float64 {
	base := s.primaryScale()
	if base == nil || s.inverse == nil {
		return 0
	}
	primary, ok := s.inverse(x)
	if !ok {
		return math.NaN()
	}
	return base.Fwd(primary)
}

func (s linkedSecondaryScale) Inv(u float64) (float64, bool) {
	base := s.primaryScale()
	if base == nil || s.forward == nil {
		return 0, false
	}
	primary, ok := base.Inv(u)
	if !ok {
		return 0, false
	}
	return s.forward(primary), true
}
