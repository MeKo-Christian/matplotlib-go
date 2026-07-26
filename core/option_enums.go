package core

// Typed spellings for the option fields that used to be raw strings.
//
// Each is a defined string type, so an untyped string constant at a call site
// still compiles (`Orientation: "horizontal"` keeps working) while the named
// constants document the accepted set and make a typo a compile error wherever
// the value is passed as a variable. The zero value is always the empty string,
// which every consumer already reads as "use the default".

// PlotOrientation selects the axis a plot type grows along.
type PlotOrientation string

const (
	// OrientationVertical grows along y, the Matplotlib default.
	OrientationVertical PlotOrientation = "vertical"
	// OrientationHorizontal grows along x.
	OrientationHorizontal PlotOrientation = "horizontal"
)

// ColorbarExtend selects which ends of a colorbar grow an out-of-range wedge.
type ColorbarExtend string

const (
	// ExtendNeither draws no out-of-range wedges, the Matplotlib default.
	ExtendNeither ColorbarExtend = "neither"
	// ExtendMin draws a wedge below the mapped range.
	ExtendMin ColorbarExtend = "min"
	// ExtendMax draws a wedge above the mapped range.
	ExtendMax ColorbarExtend = "max"
	// ExtendBoth draws wedges at both ends.
	ExtendBoth ColorbarExtend = "both"
)

// ImageAspect selects how an image's pixels map onto the axes box.
type ImageAspect string

const (
	// AspectEqual keeps pixels square by adjusting the axes box.
	AspectEqual ImageAspect = "equal"
	// AspectAuto stretches pixels to fill the axes box.
	AspectAuto ImageAspect = "auto"
)

// VectorPivot selects which part of an arrow sits on its anchor point.
type VectorPivot string

const (
	// PivotTail anchors the arrow's tail, the Matplotlib default.
	PivotTail VectorPivot = "tail"
	// PivotMiddle anchors the arrow's midpoint.
	PivotMiddle VectorPivot = "middle"
	// PivotTip anchors the arrow's tip.
	PivotTip VectorPivot = "tip"
)

// ViolinSide selects which half of a violin is drawn.
type ViolinSide string

const (
	// ViolinSideBoth draws the full violin, the Matplotlib default.
	ViolinSideBoth ViolinSide = "both"
	// ViolinSideLow draws only the low (left/bottom) half.
	ViolinSideLow ViolinSide = "low"
	// ViolinSideHigh draws only the high (right/top) half.
	ViolinSideHigh ViolinSide = "high"
)

// ColorbarLocation selects which side of the parent axes a colorbar sits on.
type ColorbarLocation string

const (
	// ColorbarRight places the colorbar to the right, the Matplotlib default.
	ColorbarRight ColorbarLocation = "right"
	// ColorbarLeft places the colorbar to the left.
	ColorbarLeft ColorbarLocation = "left"
	// ColorbarTop places the colorbar above.
	ColorbarTop ColorbarLocation = "top"
	// ColorbarBottom places the colorbar below.
	ColorbarBottom ColorbarLocation = "bottom"
)
