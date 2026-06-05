package core

import (
	"fmt"
	"strings"

	"github.com/cwbudde/matplotlib-go/render"
)

type TickParams struct {
	Reset         bool
	Axis          string
	Which         string
	Color         *render.Color
	Length        *float64
	Width         *float64
	Direction     *string
	ShowTicks     *bool
	ShowLabels    *bool
	Top           *bool
	Bottom        *bool
	Left          *bool
	Right         *bool
	LabelTop      *bool
	LabelBottom   *bool
	LabelLeft     *bool
	LabelRight    *bool
	LabelRotation *float64
	LabelSize     *float64
	LabelPad      *float64
	LabelHAlign   *TextAlign
	LabelVAlign   *TextVerticalAlign
	GridVisible   *bool
	GridColor     *render.Color
	GridAlpha     *float64
	GridLineWidth *float64
	GridDashes    []float64
}

type LocatorParams struct {
	Axis       string
	MajorCount int
	MinorCount int
}

func (a *Axes) TopAxis() *Axis {
	if a == nil {
		return nil
	}
	if a.XAxisTop == nil {
		a.XAxisTop = cloneAxisForSide(a.XAxis, AxisTop)
	}
	return a.XAxisTop
}

func (a *Axes) RightAxis() *Axis {
	if a == nil {
		return nil
	}
	if a.YAxisRight == nil {
		a.YAxisRight = cloneAxisForSide(a.YAxis, AxisRight)
	}
	return a.YAxisRight
}

func (a *Axes) HideTopAxis() {
	if a == nil {
		return
	}
	a.XAxisTop = nil
}

func (a *Axes) HideRightAxis() {
	if a == nil {
		return
	}
	a.YAxisRight = nil
}

func (a *Axes) SetXAxisSide(side AxisSide) error {
	if a == nil {
		return nil
	}
	if side != AxisBottom && side != AxisTop {
		return fmt.Errorf("x-axis side must be bottom or top")
	}
	if a.XAxis == nil {
		a.XAxis = cloneAxisForSide(nil, side)
	} else {
		a.XAxis.Side = side
	}
	if side == AxisTop {
		a.XAxisTop = nil
	}
	return nil
}

func (a *Axes) SetYAxisSide(side AxisSide) error {
	if a == nil {
		return nil
	}
	if side != AxisLeft && side != AxisRight {
		return fmt.Errorf("y-axis side must be left or right")
	}
	if a.YAxis == nil {
		a.YAxis = cloneAxisForSide(nil, side)
	} else {
		a.YAxis.Side = side
	}
	if side == AxisRight {
		a.YAxisRight = nil
	}
	return nil
}

func (a *Axes) MoveXAxisToTop() error { return a.SetXAxisSide(AxisTop) }

func (a *Axes) MoveXAxisToBottom() error { return a.SetXAxisSide(AxisBottom) }

func (a *Axes) MoveYAxisToLeft() error { return a.SetYAxisSide(AxisLeft) }

func (a *Axes) MoveYAxisToRight() error { return a.SetYAxisSide(AxisRight) }

func (a *Axes) SetXTickLabelPosition(position string) error {
	if a == nil {
		return nil
	}
	switch strings.ToLower(strings.TrimSpace(position)) {
	case "", "bottom":
		if a.XAxis != nil {
			a.XAxis.ShowLabels = true
		}
		if a.XAxisTop != nil {
			a.XAxisTop.ShowLabels = false
		}
	case "top":
		if a.XAxis != nil {
			a.XAxis.ShowLabels = false
		}
		a.TopAxis().ShowLabels = true
	case "both":
		if a.XAxis != nil {
			a.XAxis.ShowLabels = true
		}
		a.TopAxis().ShowLabels = true
	case "none":
		if a.XAxis != nil {
			a.XAxis.ShowLabels = false
		}
		if a.XAxisTop != nil {
			a.XAxisTop.ShowLabels = false
		}
	default:
		return fmt.Errorf("unsupported x tick label position %q", position)
	}
	return nil
}

func (a *Axes) SetYTickLabelPosition(position string) error {
	if a == nil {
		return nil
	}
	switch strings.ToLower(strings.TrimSpace(position)) {
	case "", "left":
		if a.YAxis != nil {
			a.YAxis.ShowLabels = true
		}
		if a.YAxisRight != nil {
			a.YAxisRight.ShowLabels = false
		}
	case "right":
		if a.YAxis != nil {
			a.YAxis.ShowLabels = false
		}
		a.RightAxis().ShowLabels = true
	case "both":
		if a.YAxis != nil {
			a.YAxis.ShowLabels = true
		}
		a.RightAxis().ShowLabels = true
	case "none":
		if a.YAxis != nil {
			a.YAxis.ShowLabels = false
		}
		if a.YAxisRight != nil {
			a.YAxisRight.ShowLabels = false
		}
	default:
		return fmt.Errorf("unsupported y tick label position %q", position)
	}
	return nil
}

func (a *Axes) MinorticksOn(axis string) error {
	if err := validateAxisSpec(axis); err != nil {
		return err
	}
	for _, ax := range a.axesForSpec(axis) {
		enableMinorTicks(ax)
	}
	return nil
}

func (a *Axes) MinorticksOff(axis string) error {
	if err := validateAxisSpec(axis); err != nil {
		return err
	}
	for _, ax := range a.axesForSpec(axis) {
		if ax != nil {
			ax.MinorLocator = nil
		}
	}
	return nil
}

func (a *Axes) LocatorParams(params LocatorParams) error {
	if err := validateAxisSpec(params.Axis); err != nil {
		return err
	}
	for _, axis := range a.axesForSpec(params.Axis) {
		if axis == nil {
			continue
		}
		if params.MajorCount > 0 {
			axis.MajorTickCount = params.MajorCount
			axis.majorTickCountFixed = true
		}
		if params.MinorCount > 0 {
			axis.MinorTickCount = params.MinorCount
		}
	}
	return nil
}

func (a *Axes) TickParams(params TickParams) error {
	if err := validateAxisSpec(params.Axis); err != nil {
		return err
	}
	which := normalizeTickWhich(params.Which)
	if which == "" {
		return fmt.Errorf("unsupported tick selection %q", params.Which)
	}

	for _, axis := range a.axesForSpec(params.Axis) {
		if axis == nil {
			continue
		}
		if params.Reset {
			resetAxisTickParams(axis)
		}
		if params.Color != nil {
			// matplotlib's tick_params(colors=...) sets both the tick mark and
			// label color, scoped to the selected major/minor tick set.
			color := *params.Color
			switch which {
			case "minor":
				minorColor := color
				minorLabel := color
				axis.MinorTickColor = &minorColor
				axis.MinorTickLabelColor = &minorLabel
			case "both":
				tickColor := color
				labelColor := color
				minorColor := color
				minorLabel := color
				axis.TickColor = &tickColor
				axis.TickLabelColor = &labelColor
				axis.MinorTickColor = &minorColor
				axis.MinorTickLabelColor = &minorLabel
			default: // major
				tickColor := color
				labelColor := color
				axis.TickColor = &tickColor
				axis.TickLabelColor = &labelColor
			}
		}
		if params.Width != nil {
			switch which {
			case "major":
				axis.TickLineWidth = *params.Width
			case "minor":
				axis.MinorTickLineWidth = *params.Width
			case "both":
				axis.TickLineWidth = *params.Width
				axis.MinorTickLineWidth = *params.Width
			}
		}
		if params.ShowTicks != nil {
			axis.ShowTicks = *params.ShowTicks
		}
		if params.Direction != nil {
			if err := axis.SetTickDirection(*params.Direction); err != nil {
				return err
			}
		}
		if params.ShowLabels != nil {
			switch which {
			case "major":
				axis.ShowLabels = *params.ShowLabels
			case "minor":
				axis.ShowMinorLabels = *params.ShowLabels
			case "both":
				axis.ShowLabels = *params.ShowLabels
				axis.ShowMinorLabels = *params.ShowLabels
			}
		}
		if params.Length != nil {
			switch which {
			case "major":
				axis.TickSize = *params.Length
			case "minor":
				axis.MinorTickSize = *params.Length
			case "both":
				axis.TickSize = *params.Length
				axis.MinorTickSize = *params.Length
			}
		}
		switch which {
		case "major":
			applyTickLabelParams(&axis.MajorLabelStyle, params)
		case "minor":
			applyTickLabelParams(&axis.MinorLabelStyle, params)
		case "both":
			applyTickLabelParams(&axis.MajorLabelStyle, params)
			applyTickLabelParams(&axis.MinorLabelStyle, params)
		}
	}
	a.applyTickSideParams(params)
	a.applyTickGridParams(params, which)
	return nil
}

func resetAxisTickParams(axis *Axis) {
	if axis == nil {
		return
	}
	var defaults *Axis
	switch axis.Side {
	case AxisLeft, AxisRight:
		defaults = NewYAxis()
	default:
		defaults = NewXAxis()
	}
	axis.TickColor = nil
	axis.TickLabelColor = nil
	axis.MinorTickColor = nil
	axis.MinorTickLabelColor = nil
	axis.TickLineCap = defaults.TickLineCap
	axis.TickLineJoin = defaults.TickLineJoin
	axis.TickLineWidth = defaults.TickLineWidth
	axis.MinorTickLineWidth = defaults.MinorTickLineWidth
	axis.TickSize = defaults.TickSize
	axis.MinorTickSize = defaults.MinorTickSize
	axis.MajorTickCount = defaults.MajorTickCount
	axis.MinorTickCount = defaults.MinorTickCount
	axis.majorTickCountFixed = defaults.majorTickCountFixed
	axis.TickDirection = defaults.TickDirection
	axis.ShowTicks = defaults.ShowTicks
	axis.ShowLabels = defaults.ShowLabels
	axis.ShowMinorLabels = defaults.ShowMinorLabels
	axis.MajorLabelStyle = defaults.MajorLabelStyle
	axis.MinorLabelStyle = defaults.MinorLabelStyle
}

func (a *Axes) applyTickSideParams(params TickParams) {
	if a == nil {
		return
	}
	if axisAllowsX(params.Axis) {
		if params.Bottom != nil {
			if *params.Bottom && a.XAxis == nil {
				a.XAxis = NewXAxis()
			}
			if a.XAxis != nil {
				a.XAxis.ShowTicks = *params.Bottom
			}
		}
		if params.Top != nil {
			if *params.Top {
				a.TopAxis().ShowTicks = true
			} else if a.XAxisTop != nil {
				a.XAxisTop.ShowTicks = false
			}
		}
		if params.LabelBottom != nil {
			if *params.LabelBottom && a.XAxis == nil {
				a.XAxis = NewXAxis()
			}
			if a.XAxis != nil {
				a.XAxis.ShowLabels = *params.LabelBottom
			}
		}
		if params.LabelTop != nil {
			if *params.LabelTop {
				a.TopAxis().ShowLabels = true
			} else if a.XAxisTop != nil {
				a.XAxisTop.ShowLabels = false
			}
		}
	}
	if axisAllowsY(params.Axis) {
		if params.Left != nil {
			if *params.Left && a.YAxis == nil {
				a.YAxis = NewYAxis()
			}
			if a.YAxis != nil {
				a.YAxis.ShowTicks = *params.Left
			}
		}
		if params.Right != nil {
			if *params.Right {
				a.RightAxis().ShowTicks = true
			} else if a.YAxisRight != nil {
				a.YAxisRight.ShowTicks = false
			}
		}
		if params.LabelLeft != nil {
			if *params.LabelLeft && a.YAxis == nil {
				a.YAxis = NewYAxis()
			}
			if a.YAxis != nil {
				a.YAxis.ShowLabels = *params.LabelLeft
			}
		}
		if params.LabelRight != nil {
			if *params.LabelRight {
				a.RightAxis().ShowLabels = true
			} else if a.YAxisRight != nil {
				a.YAxisRight.ShowLabels = false
			}
		}
	}
}

func (a *Axes) applyTickGridParams(params TickParams, which string) {
	if a == nil || (params.GridVisible == nil && params.GridColor == nil && params.GridAlpha == nil && params.GridLineWidth == nil && params.GridDashes == nil) {
		return
	}
	for _, artist := range a.Artists {
		grid, ok := artist.(*Grid)
		if !ok || !gridMatchesAxisSpec(grid, params.Axis) {
			continue
		}
		if params.GridVisible != nil {
			switch which {
			case "major":
				grid.Major = *params.GridVisible
			case "minor":
				grid.Minor = *params.GridVisible
			case "both":
				grid.Major = *params.GridVisible
				grid.Minor = *params.GridVisible
			}
		}
		if params.GridColor != nil {
			switch which {
			case "major":
				grid.Color = *params.GridColor
			case "minor":
				grid.MinorColor = *params.GridColor
			case "both":
				grid.Color = *params.GridColor
				grid.MinorColor = *params.GridColor
			}
		}
		if params.GridAlpha != nil {
			alpha := *params.GridAlpha
			if alpha < 0 {
				alpha = 0
			}
			if alpha > 1 {
				alpha = 1
			}
			switch which {
			case "major":
				grid.Alpha = alpha
			case "minor":
				grid.MinorColor.A = alpha
			case "both":
				grid.Alpha = alpha
				grid.MinorColor.A = alpha
			}
		}
		if params.GridLineWidth != nil {
			switch which {
			case "major":
				grid.LineWidth = *params.GridLineWidth
			case "minor":
				grid.MinorLineWidth = *params.GridLineWidth
			case "both":
				grid.LineWidth = *params.GridLineWidth
				grid.MinorLineWidth = *params.GridLineWidth
			}
		}
		if params.GridDashes != nil {
			dashes := styleCloneDashes(params.GridDashes)
			switch which {
			case "major":
				grid.Dashes = dashes
			case "minor":
				grid.MinorDashes = dashes
			case "both":
				grid.Dashes = styleCloneDashes(dashes)
				grid.MinorDashes = styleCloneDashes(dashes)
			}
		}
	}
}

func gridMatchesAxisSpec(grid *Grid, spec string) bool {
	if grid == nil {
		return false
	}
	switch normalizeAxisSpec(spec) {
	case "both":
		return true
	case "x", "bottom", "top":
		return grid.Axis == AxisBottom || grid.Axis == AxisTop
	case "y", "left", "right":
		return grid.Axis == AxisLeft || grid.Axis == AxisRight
	default:
		return false
	}
}

func (a *Axes) SetAxisLineStyle(spec string, lineCap render.LineCap, join render.LineJoin, dashes ...float64) error {
	if err := validateAxisSpec(spec); err != nil {
		return err
	}
	for _, axis := range a.axesForSpec(spec) {
		if axis == nil {
			continue
		}
		axis.SetLineStyle(lineCap, join, dashes...)
	}
	return nil
}

func cloneAxisForSide(src *Axis, side AxisSide) *Axis {
	var axis Axis
	if src != nil {
		axis = *src
	} else {
		switch side {
		case AxisBottom, AxisTop:
			axis = *NewXAxis()
		case AxisLeft, AxisRight:
			axis = *NewYAxis()
		}
	}
	axis.Side = side
	axis.ShowSpine = true
	axis.ShowTicks = true
	axis.ShowLabels = true
	return &axis
}

func validateAxisSpec(spec string) error {
	switch normalizeAxisSpec(spec) {
	case "", "both", "x", "y", "bottom", "top", "left", "right":
		return nil
	default:
		return fmt.Errorf("unsupported axis selection %q", spec)
	}
}

func normalizeAxisSpec(spec string) string {
	spec = strings.ToLower(strings.TrimSpace(spec))
	if spec == "" {
		return "both"
	}
	return spec
}

func axisAllowsX(spec string) bool {
	switch normalizeAxisSpec(spec) {
	case "both", "x", "bottom", "top":
		return true
	default:
		return false
	}
}

func axisAllowsY(spec string) bool {
	switch normalizeAxisSpec(spec) {
	case "both", "y", "left", "right":
		return true
	default:
		return false
	}
}

func normalizeTickWhich(which string) string {
	switch strings.ToLower(strings.TrimSpace(which)) {
	case "", "both":
		return "both"
	case "major":
		return "major"
	case "minor":
		return "minor"
	default:
		return ""
	}
}

func applyTickLabelParams(style *TickLabelStyle, params TickParams) {
	if style == nil {
		return
	}
	*style = normalizeTickLabelStyle(*style)
	if params.LabelRotation != nil {
		style.Rotation = *params.LabelRotation
	}
	if params.LabelSize != nil {
		style.FontSize = *params.LabelSize
	}
	if params.LabelPad != nil {
		style.Pad = *params.LabelPad
	}
	if params.LabelHAlign != nil {
		style.HAlign = *params.LabelHAlign
		style.AutoAlign = false
	}
	if params.LabelVAlign != nil {
		style.VAlign = *params.LabelVAlign
		style.AutoAlign = false
	}
}

func (a *Axes) axesForSpec(spec string) []*Axis {
	switch normalizeAxisSpec(spec) {
	case "x":
		return []*Axis{a.XAxis, a.XAxisTop}
	case "y":
		return []*Axis{a.YAxis, a.YAxisRight}
	case "bottom":
		return []*Axis{a.XAxis}
	case "top":
		return []*Axis{a.XAxisTop}
	case "left":
		return []*Axis{a.YAxis}
	case "right":
		return []*Axis{a.YAxisRight}
	default:
		return []*Axis{a.XAxis, a.XAxisTop, a.YAxis, a.YAxisRight}
	}
}

func enableMinorTicks(axis *Axis) {
	if axis == nil || axis.MinorLocator != nil {
		return
	}
	switch loc := axis.Locator.(type) {
	case LogLocator:
		axis.MinorLocator = LogLocator{Base: loc.Base, Minor: true, Subs: loc.Subs}
	case AutoLocator:
		axis.MinorLocator = AutoMinorLocator{Major: loc}
	case MaxNLocator:
		axis.MinorLocator = AutoMinorLocator{Major: loc}
	case MultipleLocator:
		axis.MinorLocator = AutoMinorLocator{Major: loc}
	default:
		axis.MinorLocator = MinorLinearLocator{}
	}
}
