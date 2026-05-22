package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/transform"
)

const defaultAutoScaleMargin = 0.05

// PlotOptions holds optional parameters for plotting functions.
type PlotOptions struct {
	Color       *render.Color // if nil, uses automatic color cycling
	EdgeColor   *render.Color
	LineWidth   *float64 // if nil, uses default
	EdgeWidth   *float64
	Dashes      []float64 // dash pattern
	DrawStyle   *LineDrawStyle
	Label       string   // series label for legend
	Alpha       *float64 // alpha transparency
	LevelCount  int      // contour level count for contour-like plot types
	Levels      []float64
	ZDir        string
	Offset      *float64 // fixed projection offset for contour-like plot types
	RStride     *int     // row stride for 3D surface/wireframe sampling
	CStride     *int     // column stride for 3D surface/wireframe sampling
	RCount      *int     // maximum sampled row count for 3D surface/wireframe sampling
	CCount      *int     // maximum sampled column count for 3D surface/wireframe sampling
	FaceColors  []render.Color
	Shade       *bool
	Antialiased *bool
	Colormap    *string // scalar colormap for mappable plot types
	Norm        ScalarNormalizer
	VMin        *float64
	VMax        *float64
	AxLimClip   bool
}

// Plot creates a line plot with automatic color cycling if no color is specified.
func (a *Axes) Plot(x, y []float64, opts ...PlotOptions) *Line2D {
	if len(x) == 0 || len(y) == 0 {
		return nil
	}

	// Create points
	n := len(x)
	if len(y) < n {
		n = len(y)
	}
	points := make([]geom.Pt, n)
	for i := 0; i < n; i++ {
		points[i] = geom.Pt{X: x[i], Y: y[i]}
	}

	// Default options
	var opt PlotOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	// Get color (automatic cycling if not specified)
	color := a.NextColor()
	if opt.Color != nil {
		color = *opt.Color
	}

	// Get line width
	lineWidth := 2.0
	if opt.LineWidth != nil {
		lineWidth = *opt.LineWidth
	}

	// Create line
	line := &Line2D{
		XY:        points,
		W:         lineWidth,
		Col:       color,
		Dashes:    opt.Dashes,
		DrawStyle: LineDrawStyleDefault,
		Label:     opt.Label,
	}
	if opt.DrawStyle != nil {
		line.DrawStyle = *opt.DrawStyle
	}

	// Apply alpha if specified
	if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
		line.Col.A = *opt.Alpha
	}

	a.Add(line)
	a.autoScaleIfEnabled(defaultAutoScaleMargin)
	return line
}

// SemilogX is a convenience wrapper for creating a line plot on a logarithmic
// x-axis.
func (a *Axes) SemilogX(x, y []float64, opts ...PlotOptions) *Line2D {
	line := a.Plot(x, y, opts...)
	if line == nil {
		return nil
	}
	setLogScaleFromData(a, x, true)
	return line
}

// SemilogY is a convenience wrapper for creating a line plot on a logarithmic
// y-axis.
func (a *Axes) SemilogY(x, y []float64, opts ...PlotOptions) *Line2D {
	line := a.Plot(x, y, opts...)
	if line == nil {
		return nil
	}
	setLogScaleFromData(a, y, false)
	return line
}

// LogLog is a convenience wrapper for creating a line plot on logarithmic x/y
// axes.
func (a *Axes) LogLog(x, y []float64, opts ...PlotOptions) *Line2D {
	line := a.Plot(x, y, opts...)
	if line == nil {
		return nil
	}
	setLogScaleFromData(a, x, true)
	setLogScaleFromData(a, y, false)
	return line
}

func setLogScaleFromData(ax *Axes, values []float64, isX bool) {
	minVal, maxVal := finiteRange(values)
	if minVal <= 0 || maxVal <= 0 {
		return
	}
	if minVal == maxVal {
		minVal *= 0.95
		maxVal *= 1.05
		if minVal <= 0 {
			minVal = math.SmallestNonzeroFloat64
		}
	}
	if isX {
		_ = ax.SetXScale("log", transform.WithScaleDomain(minVal, maxVal))
		return
	}
	_ = ax.SetYScale("log", transform.WithScaleDomain(minVal, maxVal))
}

// ScatterOptions holds optional parameters for scatter plots.
type ScatterOptions struct {
	Color       *render.Color // if nil, uses automatic color cycling
	Size        *float64      // marker area in points^2
	Marker      *MarkerType   // marker type
	MarkerStyle *MarkerStyle  // marker style; overrides Marker when non-nil
	MarkerPath  *geom.Path    // custom marker path (overrides Marker when non-nil)
	EdgeColor   *render.Color // edge color
	EdgeWidth   *float64      // edge width
	Alpha       *float64      // alpha transparency
	Label       string        // series label for legend
}

// Scatter creates a scatter plot with automatic shape/fill color cycling if no color is specified.
func (a *Axes) Scatter(x, y []float64, opts ...ScatterOptions) *Scatter2D {
	if len(x) == 0 || len(y) == 0 {
		return nil
	}

	// Create points
	n := len(x)
	if len(y) < n {
		n = len(y)
	}
	points := make([]geom.Pt, n)
	for i := 0; i < n; i++ {
		points[i] = geom.Pt{X: x[i], Y: y[i]}
	}

	// Default options
	var opt ScatterOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	// Get color (automatic shape/fill cycling if not specified)
	color := a.NextPatchColor()
	if opt.Color != nil {
		color = *opt.Color
	}

	// Get size. Matplotlib's scatter "s" parameter is marker area in points^2;
	// the default is lines.markersize^2 with lines.markersize = 6 pt.
	size := 36.0
	if opt.Size != nil {
		size = *opt.Size
	}

	// Get marker type
	marker := MarkerCircle
	if opt.Marker != nil {
		marker = *opt.Marker
	}

	// Get edge properties
	edgeColor := render.Color{R: 0, G: 0, B: 0, A: 0} // transparent by default
	if opt.EdgeColor != nil {
		edgeColor = *opt.EdgeColor
	}

	edgeWidth := 0.0
	if opt.EdgeWidth != nil {
		edgeWidth = *opt.EdgeWidth
	}

	// Get alpha
	alpha := 1.0
	if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
		alpha = *opt.Alpha
	}

	// Create scatter
	scatter := &Scatter2D{
		XY:        points,
		Size:      size,
		Color:     color,
		EdgeColor: edgeColor,
		EdgeWidth: edgeWidth,
		Alpha:     alpha,
		Marker:    marker,
		Label:     opt.Label,
	}
	if opt.MarkerStyle != nil {
		scatter.MarkerStyle = *opt.MarkerStyle
	}
	if opt.MarkerPath != nil {
		scatter.MarkerPath = *opt.MarkerPath
	}

	a.Add(scatter)
	return scatter
}

// BarOptions holds optional parameters for bar plots.
type BarOptions struct {
	Color       *render.Color   // if nil, uses automatic color cycling
	Width       *float64        // bar width
	EdgeColor   *render.Color   // edge color
	EdgeWidth   *float64        // edge width
	Alpha       *float64        // alpha transparency
	Baseline    *float64        // baseline value
	Baselines   []float64       // per-bar baseline/left values
	Orientation *BarOrientation // vertical or horizontal
	Label       string          // series label for legend
}

// Bar creates a bar plot with automatic color cycling if no color is specified.
func (a *Axes) Bar(x, heights []float64, opts ...BarOptions) *Bar2D {
	if len(x) == 0 || len(heights) == 0 {
		return nil
	}

	// Default options
	var opt BarOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	// Get color (automatic cycling if not specified)
	color := a.NextColor()
	if opt.Color != nil {
		color = *opt.Color
	}

	// Get width
	width := 0.8
	if opt.Width != nil {
		width = *opt.Width
	}

	// Get edge properties
	edgeColor := render.Color{R: 0, G: 0, B: 0, A: 0} // transparent by default
	if opt.EdgeColor != nil {
		edgeColor = *opt.EdgeColor
	}

	edgeWidth := 0.0
	if opt.EdgeWidth != nil {
		edgeWidth = *opt.EdgeWidth
	}

	// Get alpha
	alpha := 0.0
	if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
		alpha = *opt.Alpha
	}

	// Get baseline
	baseline := 0.0
	if opt.Baseline != nil {
		baseline = *opt.Baseline
	}

	// Get orientation
	orientation := BarVertical
	if opt.Orientation != nil {
		orientation = *opt.Orientation
	}

	// Create bar chart
	bar := &Bar2D{
		X:           x,
		Heights:     heights,
		Width:       width,
		Baselines:   append([]float64(nil), opt.Baselines...),
		Color:       color,
		EdgeColor:   edgeColor,
		EdgeWidth:   edgeWidth,
		Alpha:       alpha,
		Baseline:    baseline,
		Orientation: orientation,
		Label:       opt.Label,
	}

	a.Add(bar)
	return bar
}

// BarH creates a horizontal bar chart and sets orientation to horizontal.
func (a *Axes) BarH(y, widths []float64, opts ...BarOptions) *Bar2D {
	var opt BarOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	orientation := BarHorizontal
	opt.Orientation = &orientation
	return a.Bar(y, widths, opt)
}

// FillBetween is a convenience alias for FillBetweenPlot.
func (a *Axes) FillBetween(x, y1, y2 []float64, opts ...FillOptions) *Fill2D {
	return a.FillBetweenPlot(x, y1, y2, opts...)
}

// Fill creates an arbitrary closed polygon fill using data-space coordinates.
func (a *Axes) Fill(x, y []float64, opts ...FillOptions) *PolyCollection {
	if len(x) == 0 || len(y) == 0 {
		return nil
	}
	n := minInt(len(x), len(y))
	if n < 3 {
		return nil
	}

	var opt FillOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	points := make([]geom.Pt, n)
	for i := 0; i < n; i++ {
		points[i] = geom.Pt{X: x[i], Y: y[i]}
	}

	color := a.NextColor()
	if opt.Color != nil {
		color = *opt.Color
	}

	edgeColor := render.Color{R: 0, G: 0, B: 0, A: 0}
	if opt.EdgeColor != nil {
		edgeColor = *opt.EdgeColor
	}

	edgeWidth := 0.0
	if opt.EdgeWidth != nil {
		edgeWidth = *opt.EdgeWidth
	}

	alpha := 0.0
	if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
		alpha = *opt.Alpha
	}

	fill := &PolyCollection{
		PatchCollection: PatchCollection{
			Collection: Collection{
				Label: opt.Label,
				Alpha: 1,
				z:     2,
			},
			FaceColors: []render.Color{color},
			EdgeColor:  edgeColor,
			EdgeWidth:  edgeWidth,
		},
		Polygons: [][]geom.Pt{points},
	}

	if alpha > 0 {
		fill.FaceColors[0].A *= alpha
	}
	a.Add(fill)
	a.autoScaleIfEnabled(defaultAutoScaleMargin)
	return fill
}

// FillToBaseline is a convenience alias for FillToBaselinePlot.
func (a *Axes) FillToBaseline(x, y []float64, opts ...FillOptions) *Fill2D {
	return a.FillToBaselinePlot(x, y, opts...)
}

// FillBetweenX creates a horizontal fill between x-curves across y values.
func (a *Axes) FillBetweenX(y, x1, x2 []float64, opts ...FillOptions) *Fill2D {
	if len(y) == 0 || len(x1) == 0 || len(x2) == 0 {
		return nil
	}

	var opt FillOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	color := a.NextColor()
	if opt.Color != nil {
		color = *opt.Color
	}

	edgeColor := render.Color{R: 0, G: 0, B: 0, A: 0}
	if opt.EdgeColor != nil {
		edgeColor = *opt.EdgeColor
	}

	edgeWidth := 0.0
	if opt.EdgeWidth != nil {
		edgeWidth = *opt.EdgeWidth
	}

	alpha := 0.0
	if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
		alpha = *opt.Alpha
	}

	fill := &Fill2D{
		X:           y,
		Y1:          x1,
		Y2:          x2,
		Orientation: FillHorizontal,
		Color:       color,
		EdgeColor:   edgeColor,
		EdgeWidth:   edgeWidth,
		Alpha:       alpha,
		Label:       opt.Label,
	}

	a.Add(fill)
	a.autoScaleIfEnabled(defaultAutoScaleMargin)
	return fill
}

// FillOptions holds optional parameters for fill plots.
type FillOptions struct {
	Color     *render.Color // if nil, uses automatic color cycling
	EdgeColor *render.Color // edge color
	EdgeWidth *float64      // edge width
	Alpha     *float64      // alpha transparency
	Baseline  *float64      // baseline value
	Label     string        // series label for legend
}

// FillBetweenPlot creates a fill between two curves with automatic color cycling.
func (a *Axes) FillBetweenPlot(x, y1, y2 []float64, opts ...FillOptions) *Fill2D {
	if len(x) == 0 || len(y1) == 0 || len(y2) == 0 {
		return nil
	}

	// Default options
	var opt FillOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	// Get color (automatic cycling if not specified)
	color := a.NextColor()
	if opt.Color != nil {
		color = *opt.Color
	}

	// Get edge properties
	edgeColor := render.Color{R: 0, G: 0, B: 0, A: 0} // transparent by default
	if opt.EdgeColor != nil {
		edgeColor = *opt.EdgeColor
	}

	edgeWidth := 0.0
	if opt.EdgeWidth != nil {
		edgeWidth = *opt.EdgeWidth
	}

	// Get alpha. When omitted, preserve the color's own alpha, matching
	// Matplotlib's fill_between behavior.
	alpha := 0.0
	if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
		alpha = *opt.Alpha
	}

	// Create fill
	fill := &Fill2D{
		X:         x,
		Y1:        y1,
		Y2:        y2,
		Color:     color,
		EdgeColor: edgeColor,
		EdgeWidth: edgeWidth,
		Alpha:     alpha,
		Label:     opt.Label,
	}

	a.Add(fill)
	a.autoScaleIfEnabled(defaultAutoScaleMargin)
	return fill
}

// HistOptions holds optional parameters for histogram plots.
type HistOptions struct {
	Bins       int         // number of bins (0 = auto)
	BinEdges   []float64   // explicit bin edges (overrides Bins)
	BinStrat   BinStrategy // automatic binning strategy
	Norm       HistNorm    // normalization mode
	Cumulative bool        // accumulate bin heights from left to right
	HistType   HistType    // bar, step, or filled step presentation
	Baselines  []float64   // optional per-bin baselines for stacked histograms
	Color      *render.Color
	EdgeColor  *render.Color
	EdgeWidth  *float64
	Alpha      *float64
	Label      string
}

// Hist creates a histogram from raw data with automatic color cycling.
func (a *Axes) Hist(data []float64, opts ...HistOptions) *Hist2D {
	if len(data) == 0 {
		return nil
	}

	var opt HistOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	color := a.NextColor()
	if opt.Color != nil {
		color = *opt.Color
	}

	edgeColor := render.Color{R: 0, G: 0, B: 0, A: 0}
	if opt.EdgeColor != nil {
		edgeColor = *opt.EdgeColor
	} else if opt.HistType != HistTypeBar {
		edgeColor = color
	}

	edgeWidth := 0.0
	if opt.EdgeWidth != nil {
		edgeWidth = *opt.EdgeWidth
	} else if opt.HistType != HistTypeBar {
		edgeWidth = 1.5
	}

	alpha := 0.0
	if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
		alpha = *opt.Alpha
	}

	hist := &Hist2D{
		Data:       data,
		Bins:       opt.Bins,
		BinEdges:   opt.BinEdges,
		BinStrat:   opt.BinStrat,
		Norm:       opt.Norm,
		Cumulative: opt.Cumulative,
		HistType:   opt.HistType,
		Baselines:  append([]float64(nil), opt.Baselines...),
		Color:      color,
		EdgeColor:  edgeColor,
		EdgeWidth:  edgeWidth,
		Alpha:      alpha,
		Label:      opt.Label,
	}

	a.Add(hist)
	return hist
}

// ErrorBarOptions holds optional parameters for error bar plots.
type ErrorBarOptions struct {
	Color      *render.Color // if nil, uses automatic color cycling
	LineWidth  *float64      // error bar line width (px)
	CapSize    *float64      // cap size in pixels
	Marker     *MarkerType   // optional data marker equivalent to Matplotlib fmt markers
	MarkerSize *float64      // marker size in points
	Alpha      *float64      // alpha transparency
	Label      string        // series label for legend

	XErrLower []float64 // optional asymmetric lower x errors
	XErrUpper []float64 // optional asymmetric upper x errors
	YErrLower []float64 // optional asymmetric lower y errors
	YErrUpper []float64 // optional asymmetric upper y errors
	LoLimits  []bool    // y value is a lower limit; draw upward limit marker
	UpLimits  []bool    // y value is an upper limit; draw downward limit marker
	XLoLimits []bool    // x value is a lower limit; draw rightward limit marker
	XUpLimits []bool    // x value is an upper limit; draw leftward limit marker
}

// ErrorBar renders symmetric or asymmetric error bars for x and/or y values.
func (a *Axes) ErrorBar(x, y, xErr, yErr []float64, opts ...ErrorBarOptions) *ErrorBar {
	if len(x) == 0 || len(y) == 0 {
		return nil
	}

	var opt ErrorBarOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	color := a.NextColor()
	if opt.Color != nil {
		color = *opt.Color
	}

	lineWidth := 1.0
	if opt.LineWidth != nil {
		lineWidth = *opt.LineWidth
	}

	capSize := 6.0
	if opt.CapSize != nil {
		capSize = *opt.CapSize
	}

	alpha := 1.0
	if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
		alpha = *opt.Alpha
	}

	n := len(x)
	if len(y) < n {
		n = len(y)
	}
	if !validErrorValues(xErr, n) || !validErrorValues(yErr, n) ||
		!validErrorValues(opt.XErrLower, n) || !validErrorValues(opt.XErrUpper, n) ||
		!validErrorValues(opt.YErrLower, n) || !validErrorValues(opt.YErrUpper, n) ||
		!validBoolValues(opt.LoLimits, n) || !validBoolValues(opt.UpLimits, n) ||
		!validBoolValues(opt.XLoLimits, n) || !validBoolValues(opt.XUpLimits, n) {
		return nil
	}

	pts := make([]geom.Pt, n)
	for i := 0; i < n; i++ {
		pts[i] = geom.Pt{X: x[i], Y: y[i]}
	}

	bar := &ErrorBar{
		XY:        pts,
		XErr:      xErr,
		YErr:      yErr,
		XErrLower: append([]float64(nil), opt.XErrLower...),
		XErrUpper: append([]float64(nil), opt.XErrUpper...),
		YErrLower: append([]float64(nil), opt.YErrLower...),
		YErrUpper: append([]float64(nil), opt.YErrUpper...),
		LoLimits:  append([]bool(nil), opt.LoLimits...),
		UpLimits:  append([]bool(nil), opt.UpLimits...),
		XLoLimits: append([]bool(nil), opt.XLoLimits...),
		XUpLimits: append([]bool(nil), opt.XUpLimits...),
		Color:     color,
		LineWidth: lineWidth,
		CapSize:   capSize,
		Alpha:     alpha,
		Label:     opt.Label,
	}
	if opt.Marker != nil {
		bar.Marker = *opt.Marker
		bar.MarkerSet = true
	}
	if opt.MarkerSize != nil {
		bar.MarkerSize = *opt.MarkerSize
	}
	a.Add(bar)
	return bar
}

func validErrorValues(values []float64, n int) bool {
	if len(values) == 0 || len(values) == 1 || len(values) == n {
		for _, value := range values {
			if value < 0 || !isFinite(value) {
				return false
			}
		}
		return true
	}
	return false
}

func validBoolValues(values []bool, n int) bool {
	return len(values) == 0 || len(values) == 1 || len(values) == n
}

// BoxPlotOptions holds optional parameters for box plots.
type BoxPlotOptions struct {
	Position     *float64      // x position of the box center
	Width        *float64      // box width in data units
	Color        *render.Color // box fill color
	EdgeColor    *render.Color // box outline color
	MedianColor  *render.Color // median line color
	WhiskerColor *render.Color // whisker and cap color
	CapColor     *render.Color // whisker cap color
	FlierColor   *render.Color // outlier marker color
	EdgeWidth    *float64      // box outline width in pixels
	WhiskerWidth *float64      // whisker line width in pixels
	MedianWidth  *float64      // median line width in pixels
	CapWidth     *float64      // cap length in data units
	FlierSize    *float64      // outlier marker radius in pixels
	Alpha        *float64      // alpha transparency
	ShowFliers   *bool         // whether to draw outliers
	Label        string        // series label for legend

	Notch              *bool       // draw a notched box using the confidence interval
	Bootstrap          int         // accepted for Matplotlib API parity; deterministic CI fallback is used
	ConfidenceInterval *[2]float64 // custom median confidence interval for notches
	CustomMedian       *float64    // override the computed median
	WhiskerPercentiles *[2]float64 // percentile whisker range, e.g. [5, 95]
	FlierMarker        *MarkerType // marker for outlier points
	FlierEdgeColor     *render.Color
	FlierEdgeWidth     *float64
}

// BoxPlotsOptions holds optional parameters for multi-series box plots.
type BoxPlotsOptions struct {
	Positions    []float64      // x positions for each box center
	Width        *float64       // box width in data units
	Colors       []render.Color // box fill colors, one per dataset
	EdgeColor    *render.Color  // box outline color
	MedianColor  *render.Color  // median line color
	WhiskerColor *render.Color  // whisker and cap color
	CapColor     *render.Color  // whisker cap color
	FlierColor   *render.Color  // outlier marker color
	EdgeWidth    *float64       // box outline width in pixels
	WhiskerWidth *float64       // whisker line width in pixels
	MedianWidth  *float64       // median line width in pixels
	CapWidth     *float64       // cap length in data units
	FlierSize    *float64       // outlier marker radius in pixels
	Alpha        *float64       // alpha transparency
	ShowFliers   *bool          // whether to draw outliers
	Labels       []string       // series labels for legend

	Notch               *bool
	Bootstrap           int
	ConfidenceIntervals [][2]float64
	CustomMedians       []float64
	WhiskerPercentiles  *[2]float64
	FlierMarker         *MarkerType
	FlierEdgeColor      *render.Color
	FlierEdgeWidth      *float64
}

// BoxPlot creates a box plot from raw sample data with automatic color cycling.
func (a *Axes) BoxPlot(data []float64, opts ...BoxPlotOptions) *BoxPlot2D {
	if len(data) == 0 {
		return nil
	}

	var opt BoxPlotOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	position := 1.0
	if opt.Position != nil {
		position = *opt.Position
	}

	width := 0.6
	if opt.Width != nil {
		width = *opt.Width
	}

	color := a.NextColor()
	if opt.Color != nil {
		color = *opt.Color
	}

	edgeColor := render.Color{R: 0, G: 0, B: 0, A: 1}
	if opt.EdgeColor != nil {
		edgeColor = *opt.EdgeColor
	}

	medianColor := edgeColor
	if opt.MedianColor != nil {
		medianColor = *opt.MedianColor
	}

	whiskerColor := edgeColor
	if opt.WhiskerColor != nil {
		whiskerColor = *opt.WhiskerColor
	}

	capColor := whiskerColor
	if opt.CapColor != nil {
		capColor = *opt.CapColor
	}

	flierColor := edgeColor
	if opt.FlierColor != nil {
		flierColor = *opt.FlierColor
	}

	edgeWidth := 1.0
	if opt.EdgeWidth != nil {
		edgeWidth = *opt.EdgeWidth
	}

	whiskerWidth := 1.0
	if opt.WhiskerWidth != nil {
		whiskerWidth = *opt.WhiskerWidth
	}

	medianWidth := 1.5
	if opt.MedianWidth != nil {
		medianWidth = *opt.MedianWidth
	}

	capWidth := width * 0.5
	if opt.CapWidth != nil {
		capWidth = *opt.CapWidth
	}

	flierSize := 3.5
	if opt.FlierSize != nil {
		flierSize = *opt.FlierSize
	}

	alpha := 1.0
	if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
		alpha = *opt.Alpha
	}

	showFliers := true
	if opt.ShowFliers != nil {
		showFliers = *opt.ShowFliers
	}
	notch := false
	if opt.Notch != nil {
		notch = *opt.Notch
	}
	flierMarker := MarkerCircle
	if opt.FlierMarker != nil {
		flierMarker = *opt.FlierMarker
	}
	flierEdgeColor := flierColor
	if opt.FlierEdgeColor != nil {
		flierEdgeColor = *opt.FlierEdgeColor
	}
	flierEdgeWidth := 1.0
	if opt.FlierEdgeWidth != nil {
		flierEdgeWidth = *opt.FlierEdgeWidth
	}

	box := &BoxPlot2D{
		Data:               data,
		Position:           position,
		Width:              width,
		Color:              color,
		EdgeColor:          edgeColor,
		MedianColor:        medianColor,
		WhiskerColor:       whiskerColor,
		CapColor:           capColor,
		FlierColor:         flierColor,
		FlierEdgeColor:     flierEdgeColor,
		EdgeWidth:          edgeWidth,
		WhiskerWidth:       whiskerWidth,
		MedianWidth:        medianWidth,
		CapWidth:           capWidth,
		FlierSize:          flierSize,
		FlierEdgeWidth:     flierEdgeWidth,
		Alpha:              alpha,
		ShowFliers:         showFliers,
		Notch:              notch,
		Bootstrap:          opt.Bootstrap,
		ConfidenceInterval: opt.ConfidenceInterval,
		CustomMedian:       opt.CustomMedian,
		WhiskerPercentiles: opt.WhiskerPercentiles,
		FlierMarker:        flierMarker,
		Label:              opt.Label,
	}

	a.Add(box)
	return box
}

// BoxPlots creates a group of box plots from raw sample datasets.
func (a *Axes) BoxPlots(datasets [][]float64, opts ...BoxPlotsOptions) []*BoxPlot2D {
	if len(datasets) == 0 {
		return nil
	}

	var opt BoxPlotsOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	width := opt.Width
	if width == nil {
		defaultWidth := matplotlibBoxPlotDefaultWidth(len(datasets), opt.Positions)
		width = &defaultWidth
	}

	boxes := make([]*BoxPlot2D, 0, len(datasets))
	for i, data := range datasets {
		position := float64(i + 1)
		if i < len(opt.Positions) {
			position = opt.Positions[i]
		}

		boxOpt := BoxPlotOptions{
			Position:           &position,
			Width:              width,
			EdgeColor:          opt.EdgeColor,
			MedianColor:        opt.MedianColor,
			WhiskerColor:       opt.WhiskerColor,
			CapColor:           opt.CapColor,
			FlierColor:         opt.FlierColor,
			EdgeWidth:          opt.EdgeWidth,
			WhiskerWidth:       opt.WhiskerWidth,
			MedianWidth:        opt.MedianWidth,
			CapWidth:           opt.CapWidth,
			FlierSize:          opt.FlierSize,
			Alpha:              opt.Alpha,
			ShowFliers:         opt.ShowFliers,
			Notch:              opt.Notch,
			Bootstrap:          opt.Bootstrap,
			WhiskerPercentiles: opt.WhiskerPercentiles,
			FlierMarker:        opt.FlierMarker,
			FlierEdgeColor:     opt.FlierEdgeColor,
			FlierEdgeWidth:     opt.FlierEdgeWidth,
		}
		if i < len(opt.ConfidenceIntervals) {
			ci := opt.ConfidenceIntervals[i]
			boxOpt.ConfidenceInterval = &ci
		}
		if i < len(opt.CustomMedians) && isFinite(opt.CustomMedians[i]) {
			median := opt.CustomMedians[i]
			boxOpt.CustomMedian = &median
		}
		if i < len(opt.Colors) {
			boxOpt.Color = &opt.Colors[i]
		}
		if i < len(opt.Labels) {
			boxOpt.Label = opt.Labels[i]
		}

		if box := a.BoxPlot(data, boxOpt); box != nil {
			boxes = append(boxes, box)
		}
	}
	return boxes
}

func matplotlibBoxPlotDefaultWidth(n int, positions []float64) float64 {
	if n <= 1 {
		return 0.15
	}
	minPos := math.Inf(1)
	maxPos := math.Inf(-1)
	for i := 0; i < n; i++ {
		position := float64(i + 1)
		if i < len(positions) && isFinite(positions[i]) {
			position = positions[i]
		}
		minPos = math.Min(minPos, position)
		maxPos = math.Max(maxPos, position)
	}
	if !isFinite(minPos) || !isFinite(maxPos) {
		return 0.15
	}
	return math.Min(0.5, math.Max(0.15, 0.15*(maxPos-minPos)))
}

// FillToBaselinePlot creates a fill from a curve to baseline with automatic color cycling.
func (a *Axes) FillToBaselinePlot(x, y []float64, opts ...FillOptions) *Fill2D {
	if len(x) == 0 || len(y) == 0 {
		return nil
	}

	// Default options
	var opt FillOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	// Get color (automatic cycling if not specified)
	color := a.NextColor()
	if opt.Color != nil {
		color = *opt.Color
	}

	// Get edge properties
	edgeColor := render.Color{R: 0, G: 0, B: 0, A: 0} // transparent by default
	if opt.EdgeColor != nil {
		edgeColor = *opt.EdgeColor
	}

	edgeWidth := 0.0
	if opt.EdgeWidth != nil {
		edgeWidth = *opt.EdgeWidth
	}

	// Get alpha. When omitted, preserve the color's own alpha, matching
	// Matplotlib's fill_between behavior.
	alpha := 0.0
	if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
		alpha = *opt.Alpha
	}

	// Get baseline
	baseline := 0.0
	if opt.Baseline != nil {
		baseline = *opt.Baseline
	}

	// Create fill
	fill := &Fill2D{
		X:         x,
		Y1:        y,
		Y2:        nil,
		Baseline:  baseline,
		Color:     color,
		EdgeColor: edgeColor,
		EdgeWidth: edgeWidth,
		Alpha:     alpha,
		Label:     opt.Label,
	}

	a.Add(fill)
	a.autoScaleIfEnabled(defaultAutoScaleMargin)
	return fill
}
