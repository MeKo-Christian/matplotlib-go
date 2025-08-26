package core

import (
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

// PlotOptions holds optional parameters for plotting functions.
type PlotOptions struct {
	Color      *render.Color // if nil, uses automatic color cycling
	LineWidth  *float64      // if nil, uses default
	Dashes     []float64     // dash pattern
	Label      string        // series label for legend
	Alpha      *float64      // alpha transparency
}

// Plot creates a line plot with automatic color cycling if no color is specified.
func (a *Axes) Plot(x, y []float64, opts ...PlotOptions) *Line2D {
	if len(x) == 0 || len(y) == 0 {
		return nil
	}

	// Create points
	points := make([]geom.Pt, len(x))
	n := len(x)
	if len(y) < n {
		n = len(y)
	}
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
		XY:     points,
		W:      lineWidth,
		Col:    color,
		Dashes: opt.Dashes,
		Label:  opt.Label,
	}

	// Apply alpha if specified
	if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
		line.Col.A = *opt.Alpha
	}

	a.Add(line)
	return line
}

// ScatterOptions holds optional parameters for scatter plots.
type ScatterOptions struct {
	Color      *render.Color // if nil, uses automatic color cycling
	Size       *float64      // marker size
	Marker     *MarkerType   // marker type
	EdgeColor  *render.Color // edge color
	EdgeWidth  *float64      // edge width
	Alpha      *float64      // alpha transparency
	Label      string        // series label for legend
}

// Scatter creates a scatter plot with automatic color cycling if no color is specified.
func (a *Axes) Scatter(x, y []float64, opts ...ScatterOptions) *Scatter2D {
	if len(x) == 0 || len(y) == 0 {
		return nil
	}

	// Create points
	points := make([]geom.Pt, len(x))
	n := len(x)
	if len(y) < n {
		n = len(y)
	}
	for i := 0; i < n; i++ {
		points[i] = geom.Pt{X: x[i], Y: y[i]}
	}

	// Default options
	var opt ScatterOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	// Get color (automatic cycling if not specified)
	color := a.NextColor()
	if opt.Color != nil {
		color = *opt.Color
	}

	// Get size
	size := 8.0
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
	alpha := 1.0
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

	// Get alpha
	alpha := 0.6 // default for fill areas
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
	return fill
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

	// Get alpha
	alpha := 0.6 // default for fill areas
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
	return fill
}