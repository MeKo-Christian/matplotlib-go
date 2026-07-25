package core

import (
	"fmt"
	"strings"

	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/ticker"
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
		a.XAxisTop.ShowSpine = a.fallbackSpineVisible(AxisTop)
	}
	return a.XAxisTop
}

func (a *Axes) RightAxis() *Axis {
	if a == nil {
		return nil
	}
	if a.YAxisRight == nil {
		a.YAxisRight = cloneAxisForSide(a.YAxis, AxisRight)
		a.YAxisRight.ShowSpine = a.fallbackSpineVisible(AxisRight)
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
				axis.tickLineWidthSet = true
			case "minor":
				axis.MinorTickLineWidth = *params.Width
				axis.minorTickLineWidthSet = true
			case "both":
				axis.TickLineWidth = *params.Width
				axis.MinorTickLineWidth = *params.Width
				axis.tickLineWidthSet = true
				axis.minorTickLineWidthSet = true
			}
		}
		if params.ShowTicks != nil {
			switch which {
			case "major":
				axis.ShowTicks = *params.ShowTicks
			case "minor":
				axis.ShowMinorTicks = *params.ShowTicks
				axis.minorTickVisibilitySet = true
			case "both":
				axis.ShowTicks = *params.ShowTicks
				axis.ShowMinorTicks = *params.ShowTicks
				axis.minorTickVisibilitySet = false
			}
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
				axis.minorTickSizeSet = true
			case "both":
				axis.TickSize = *params.Length
				axis.MinorTickSize = *params.Length
				axis.minorTickSizeSet = true
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
	axis.tickLineWidthSet = defaults.tickLineWidthSet
	axis.minorTickLineWidthSet = defaults.minorTickLineWidthSet
	axis.TickSize = defaults.TickSize
	axis.MinorTickSize = defaults.MinorTickSize
	axis.minorTickSizeSet = defaults.minorTickSizeSet
	axis.MajorTickCount = defaults.MajorTickCount
	axis.MinorTickCount = defaults.MinorTickCount
	axis.majorTickCountFixed = defaults.majorTickCountFixed
	axis.TickDirection = defaults.TickDirection
	axis.ShowTicks = defaults.ShowTicks
	axis.ShowMinorTicks = defaults.ShowMinorTicks
	axis.minorTickVisibilitySet = defaults.minorTickVisibilitySet
	axis.ShowLabels = defaults.ShowLabels
	axis.ShowMinorLabels = defaults.ShowMinorLabels
	axis.MajorLabelStyle = defaults.MajorLabelStyle
	axis.MinorLabelStyle = defaults.MinorLabelStyle
}

func (a *Axes) applyTickSideParams(params TickParams) {
	if a == nil {
		return
	}
	which := normalizeTickWhich(params.Which)
	if axisAllowsX(params.Axis) {
		if params.Bottom != nil {
			if *params.Bottom && a.XAxis == nil {
				a.XAxis = NewXAxis()
			}
			setAxisTickVisibility(a.XAxis, which, *params.Bottom)
		}
		if params.Top != nil {
			var axis *Axis
			if *params.Top {
				axis = a.ensureTickSideAxis(AxisTop)
			} else {
				axis = a.XAxisTop
			}
			setAxisTickVisibility(axis, which, *params.Top)
		}
		if params.LabelBottom != nil {
			if *params.LabelBottom && a.XAxis == nil {
				a.XAxis = NewXAxis()
			}
			setAxisTickLabelVisibility(a.XAxis, which, *params.LabelBottom)
		}
		if params.LabelTop != nil {
			var axis *Axis
			if *params.LabelTop {
				axis = a.ensureTickSideAxis(AxisTop)
			} else {
				axis = a.XAxisTop
			}
			setAxisTickLabelVisibility(axis, which, *params.LabelTop)
		}
	}
	if axisAllowsY(params.Axis) {
		if params.Left != nil {
			if *params.Left && a.YAxis == nil {
				a.YAxis = NewYAxis()
			}
			setAxisTickVisibility(a.YAxis, which, *params.Left)
		}
		if params.Right != nil {
			var axis *Axis
			if *params.Right {
				axis = a.ensureTickSideAxis(AxisRight)
			} else {
				axis = a.YAxisRight
			}
			setAxisTickVisibility(axis, which, *params.Right)
		}
		if params.LabelLeft != nil {
			if *params.LabelLeft && a.YAxis == nil {
				a.YAxis = NewYAxis()
			}
			setAxisTickLabelVisibility(a.YAxis, which, *params.LabelLeft)
		}
		if params.LabelRight != nil {
			var axis *Axis
			if *params.LabelRight {
				axis = a.ensureTickSideAxis(AxisRight)
			} else {
				axis = a.YAxisRight
			}
			setAxisTickLabelVisibility(axis, which, *params.LabelRight)
		}
	}
}

func (a *Axes) ensureTickSideAxis(side AxisSide) *Axis {
	var axis *Axis
	switch side {
	case AxisTop:
		if a.XAxisTop != nil {
			return a.XAxisTop
		}
		axis = a.TopAxis()
	case AxisRight:
		if a.YAxisRight != nil {
			return a.YAxisRight
		}
		axis = a.RightAxis()
	default:
		return nil
	}
	axis.ShowTicks = false
	axis.ShowMinorTicks = false
	axis.minorTickVisibilitySet = true
	axis.ShowLabels = false
	axis.ShowMinorLabels = false
	return axis
}

func setAxisTickVisibility(axis *Axis, which string, visible bool) {
	if axis == nil {
		return
	}
	switch which {
	case "major":
		axis.ShowTicks = visible
	case "minor":
		axis.ShowMinorTicks = visible
		axis.minorTickVisibilitySet = true
	case "both":
		axis.ShowTicks = visible
		axis.ShowMinorTicks = visible
		axis.minorTickVisibilitySet = false
	}
}

func setAxisTickLabelVisibility(axis *Axis, which string, visible bool) {
	if axis == nil {
		return
	}
	switch which {
	case "major":
		axis.ShowLabels = visible
	case "minor":
		axis.ShowMinorLabels = visible
	case "both":
		axis.ShowLabels = visible
		axis.ShowMinorLabels = visible
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

func (a *Axes) applyRCTickDefaults(rc *style.RC) {
	if a == nil || rc == nil {
		return
	}
	a.applyRCTickAxisDefaults(rc, &rc.XTick, true)
	a.applyRCTickAxisDefaults(rc, &rc.YTick, false)
}

func (a *Axes) applyRCTickAxisDefaults(rc *style.RC, cfg *style.TickAxisRC, isX bool) {
	if cfg == nil {
		return
	}

	var primary *Axis
	if isX {
		primary = a.XAxis
	} else {
		primary = a.YAxis
	}
	majorPrimary := cfg.Primary && cfg.Major.Primary
	minorPrimary := cfg.Primary && cfg.Minor.Primary
	labelMajorPrimary := cfg.LabelPrimary && cfg.Major.Primary
	labelMinorPrimary := cfg.Minor.Visible && cfg.LabelPrimary && cfg.Minor.Primary
	configureAxisFromTickRC(primary, rc, cfg, isX,
		majorPrimary, minorPrimary, labelMajorPrimary, labelMinorPrimary)

	majorSecondary := cfg.Secondary && cfg.Major.Secondary
	minorSecondary := cfg.Secondary && cfg.Minor.Secondary
	labelMajorSecondary := cfg.LabelSecondary && cfg.Major.Secondary
	labelMinorSecondary := cfg.Minor.Visible && cfg.LabelSecondary && cfg.Minor.Secondary
	if !majorSecondary && !minorSecondary && !labelMajorSecondary && !labelMinorSecondary {
		return
	}

	side := AxisRight
	if isX {
		side = AxisTop
	}
	secondary := a.ensureTickSideAxis(side)
	configureAxisFromTickRC(secondary, rc, cfg, isX,
		majorSecondary, minorSecondary, labelMajorSecondary, labelMinorSecondary)
}

func configureAxisFromTickRC(axis *Axis, rc *style.RC, cfg *style.TickAxisRC, isX, majorTicks, minorTicks, majorLabels, minorLabels bool) {
	if axis == nil || rc == nil || cfg == nil {
		return
	}
	axis.TickSize = pointsToPixels(*rc, cfg.Major.Size)
	axis.MinorTickSize = pointsToPixels(*rc, cfg.Minor.Size)
	axis.minorTickSizeSet = true
	axis.TickLineWidth = cfg.Major.Width
	axis.MinorTickLineWidth = cfg.Minor.Width
	axis.tickLineWidthSet = true
	axis.minorTickLineWidthSet = true
	axis.MajorLabelStyle.PadPt = cfg.Major.Pad
	axis.MajorLabelStyle.padPtSet = true
	axis.MinorLabelStyle.PadPt = cfg.Minor.Pad
	axis.MinorLabelStyle.padPtSet = true
	axis.ShowTicks = majorTicks
	axis.ShowMinorTicks = minorTicks
	axis.minorTickVisibilitySet = majorTicks != minorTicks
	axis.ShowLabels = majorLabels
	axis.ShowMinorLabels = minorLabels
	_ = axis.SetTickDirection(cfg.Direction)
	applyRCTickAlignment(axis, cfg.Alignment, isX)

	if cfg.Minor.Visible {
		enableMinorTicks(axis)
		if auto, ok := axis.MinorLocator.(ticker.AutoMinorLocator); ok {
			auto.N = cfg.Minor.NDivs
			axis.MinorLocator = auto
		}
	}
}

func applyRCTickAlignment(axis *Axis, alignment string, isX bool) {
	if axis == nil {
		return
	}
	if isX {
		if alignment == style.Default.XTick.Alignment {
			return
		}
		hAlign := TextAlignCenter
		switch alignment {
		case "left":
			hAlign = TextAlignLeft
		case "right":
			hAlign = TextAlignRight
		}
		vAlign := TextVAlignTop
		if axis.Side == AxisTop {
			vAlign = TextVAlignBottom
		}
		setTickLevelAlignment(&axis.MajorLabelStyle, hAlign, vAlign)
		setTickLevelAlignment(&axis.MinorLabelStyle, hAlign, vAlign)
		return
	}
	if alignment == style.Default.YTick.Alignment {
		return
	}
	vAlign := TextVAlignMiddle
	switch alignment {
	case "baseline":
		vAlign = TextVAlignBaseline
	case "bottom":
		vAlign = TextVAlignBottom
	case "center_baseline":
		vAlign = TextVAlignCenterBaseline
	case "top":
		vAlign = TextVAlignTop
	}
	hAlign := TextAlignRight
	if axis.Side == AxisRight {
		hAlign = TextAlignLeft
	}
	setTickLevelAlignment(&axis.MajorLabelStyle, hAlign, vAlign)
	setTickLevelAlignment(&axis.MinorLabelStyle, hAlign, vAlign)
}

func setTickLevelAlignment(labelStyle *TickLabelStyle, hAlign TextAlign, vAlign TextVerticalAlign) {
	labelStyle.HAlign = hAlign
	labelStyle.VAlign = vAlign
	labelStyle.AutoAlign = false
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
	case ticker.LogLocator:
		axis.MinorLocator = ticker.LogLocator{Base: loc.Base, Minor: true, Subs: loc.Subs}
	case ticker.AutoLocator:
		axis.MinorLocator = ticker.AutoMinorLocator{Major: loc}
	case ticker.MaxNLocator:
		axis.MinorLocator = ticker.AutoMinorLocator{Major: loc}
	case ticker.MultipleLocator:
		axis.MinorLocator = ticker.AutoMinorLocator{Major: loc}
	default:
		axis.MinorLocator = ticker.MinorLinearLocator{}
	}
}
