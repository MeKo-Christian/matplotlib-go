package core

import "github.com/cwbudde/matplotlib-go/render"

// AxisSide specifies which side of the plot area an axis is on.
type AxisSide uint8

const (
	AxisBottom AxisSide = iota // x-axis at bottom
	AxisTop                    // x-axis at top
	AxisLeft                   // y-axis at left
	AxisRight                  // y-axis at right
)

// Matplotlib defaults axes.linewidth to 0.8 pt and major ticks to 3.5 pt.
// The default figure DPI is 100, so store the pixel equivalent for axes
// constructed without a figure-backed RC.
const (
	defaultAxisLineWidth      = 0.8 * 100.0 / 72.0
	defaultMinorTickLineWidth = 0.6 * 100.0 / 72.0
	defaultTickSizePt         = 3.5
	defaultTickSizePx         = defaultTickSizePt * 100.0 / 72.0
	defaultTickPadPt          = 3.5
	offsetTextPadPt           = 3.0
)

// TickLabelStyle captures axis-owned label placement and orientation.
type TickLabelStyle struct {
	Rotation  float64
	Pad       float64
	HAlign    TextAlign
	VAlign    TextVerticalAlign
	FontSize  float64
	FontKey   string
	AutoAlign bool
}

// TickLevel adds an optional additional tick/label row to an axis.
type TickLevel struct {
	Locator    Locator
	Formatter  Formatter
	Size       float64
	ShowTicks  bool
	ShowLabels bool
	LabelStyle TickLabelStyle
}

// TickDirection controls whether ticks point outward, inward, or straddle the spine.
type TickDirection uint8

const (
	TickDirectionOut TickDirection = iota
	TickDirectionIn
	TickDirectionInOut
)

// AxisSpinePositionMode controls where the axis spine is drawn.
type AxisSpinePositionMode uint8

const (
	AxisSpinePositionBoundary AxisSpinePositionMode = iota
	AxisSpinePositionData
)

// Axis renders axis spines, ticks, and labels for a single dimension.
type Axis struct {
	Side                AxisSide     // which side of the plot
	Locator             Locator      // major tick position calculator
	MinorLocator        Locator      // minor tick position calculator (nil = no minor ticks)
	Formatter           Formatter    // major tick label formatter
	MinorFormatter      Formatter    // optional minor tick label formatter
	Color               render.Color // axis spine color, and tick/label color unless overridden
	TickColor           *render.Color
	TickLabelColor      *render.Color
	MinorTickColor      *render.Color // minor tick mark color (nil falls back to TickColor)
	MinorTickLabelColor *render.Color // minor tick label color (nil falls back to MinorTickColor)
	LineWidth           float64       // width of axis spine
	LineCap             render.LineCap
	LineJoin            render.LineJoin
	TickLineCap         render.LineCap
	TickLineJoin        render.LineJoin
	TickLineWidth       float64
	MinorTickLineWidth  float64
	Dashes              []float64
	TickSize            float64 // length of major tick marks (in pixels)
	MinorTickSize       float64 // length of minor tick marks (in pixels); 0 uses TickSize*0.6
	MajorTickCount      int     // target major tick count for automatic locators
	MinorTickCount      int     // target minor tick count for automatic locators
	majorTickCountFixed bool
	TickDirection       TickDirection
	SpinePositionMode   AxisSpinePositionMode
	SpinePosition       float64
	ShowSpine           bool // whether to draw the axis line
	ShowTicks           bool // whether to draw major/minor tick marks
	ShowLabels          bool // whether to draw major tick labels
	ShowMinorLabels     bool // whether to draw minor tick labels
	MajorLabelStyle     TickLabelStyle
	MinorLabelStyle     TickLabelStyle
	ExtraTickLevels     []TickLevel
	z                   float64 // z-order
}

// NewXAxis creates an axis for the bottom (x-axis).
func NewXAxis() *Axis {
	return &Axis{
		Side:               AxisBottom,
		Locator:            AutoLocator{},
		Formatter:          ScalarFormatter{Prec: 3},
		Color:              render.Color{R: 0, G: 0, B: 0, A: 1}, // black
		LineWidth:          defaultAxisLineWidth,
		LineCap:            render.CapSquare,
		LineJoin:           render.JoinMiter,
		TickLineCap:        render.CapButt,
		TickLineJoin:       render.JoinMiter,
		MinorTickLineWidth: defaultMinorTickLineWidth,
		TickSize:           defaultTickSizePx,
		MajorTickCount:     9,
		MinorTickCount:     30,
		TickDirection:      TickDirectionOut,
		SpinePositionMode:  AxisSpinePositionBoundary,
		ShowSpine:          true,
		ShowTicks:          true,
		ShowLabels:         true,
		MajorLabelStyle:    defaultTickLabelStyle(),
		MinorLabelStyle:    defaultTickLabelStyle(),
	}
}

// NewYAxis creates an axis for the left (y-axis).
func NewYAxis() *Axis {
	return &Axis{
		Side:               AxisLeft,
		Locator:            AutoLocator{},
		Formatter:          ScalarFormatter{Prec: 3},
		Color:              render.Color{R: 0, G: 0, B: 0, A: 1}, // black
		LineWidth:          defaultAxisLineWidth,
		LineCap:            render.CapSquare,
		LineJoin:           render.JoinMiter,
		TickLineCap:        render.CapButt,
		TickLineJoin:       render.JoinMiter,
		MinorTickLineWidth: defaultMinorTickLineWidth,
		TickSize:           defaultTickSizePx,
		MajorTickCount:     9,
		MinorTickCount:     30,
		TickDirection:      TickDirectionOut,
		SpinePositionMode:  AxisSpinePositionBoundary,
		ShowSpine:          true,
		ShowTicks:          true,
		ShowLabels:         true,
		MajorLabelStyle:    defaultTickLabelStyle(),
		MinorLabelStyle:    defaultTickLabelStyle(),
	}
}
