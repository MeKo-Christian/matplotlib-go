package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/diag"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/transform"
)

const defaultAutoScaleMargin = 0.05

// PlotOptions holds optional parameters for plotting functions.
type PlotOptions struct {
	Color           *render.Color // if nil, uses automatic color cycling
	EdgeColor       *render.Color
	LineWidth       *float64 // if nil, uses default
	LineCap         *render.LineCap
	EdgeWidth       *float64
	Dashes          []float64 // dash pattern
	DrawStyle       *LineDrawStyle
	Marker          *MarkerType
	MarkerStyle     *MarkerStyle
	MarkerPath      *geom.Path
	MarkerSize      *float64
	MarkerFaceColor *render.Color
	MarkerEdgeColor *render.Color
	MarkerFaceSpec  *MarkerColorSpec
	MarkerEdgeSpec  *MarkerColorSpec
	MarkerFaceAlt   *MarkerColorSpec
	MarkerEdgeWidth *float64
	MarkEvery       int
	MarkEverySpec   *MarkEverySpec
	Label           string   // series label for legend
	Alpha           *float64 // alpha transparency
	LevelCount      int      // contour level count for contour-like plot types
	Levels          []float64
	ZDir            string
	Offset          *float64 // fixed projection offset for contour-like plot types
	RStride         *int     // row stride for 3D surface/wireframe sampling
	CStride         *int     // column stride for 3D surface/wireframe sampling
	RCount          *int     // maximum sampled row count for 3D surface/wireframe sampling
	CCount          *int     // maximum sampled column count for 3D surface/wireframe sampling
	FaceColors      []render.Color
	Shade           *bool
	Antialiased     *bool
	Colormap        *string // scalar colormap for mappable plot types
	Norm            ScalarNormalizer
	VMin            *float64
	VMax            *float64
	AxLimClip       bool
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

	// Pull one step from the property cycle (color plus any
	// linestyle/marker/linewidth carried by axes.prop_cycle).
	cycle := a.NextLineProps()
	color := cycle.Color
	if opt.Color != nil {
		color = *opt.Color
	}

	// Get line width: explicit option, else cycled linewidth, else default.
	lineWidth := 2.0
	if cycle.HasLineWidth {
		lineWidth = cycle.LineWidth
	}
	if opt.LineWidth != nil {
		lineWidth = *opt.LineWidth
	}

	// Resolve dashes: explicit option wins, otherwise honor a cycled linestyle.
	dashes := opt.Dashes
	if dashes == nil && cycle.HasLineStyle {
		dashes = lineStyleToDashes(cycle.LineStyle, lineWidth)
	}

	// Create line
	line := &Line2D{
		XY:        points,
		W:         lineWidth,
		Col:       color,
		Dashes:    dashes,
		DrawStyle: LineDrawStyleDefault,
		Label:     opt.Label,
	}
	if opt.LineCap != nil {
		line.LineCap = *opt.LineCap
		line.LineCapSet = true
	}
	if opt.DrawStyle != nil {
		line.DrawStyle = *opt.DrawStyle
	}
	if opt.Marker != nil {
		line.Marker = *opt.Marker
		line.MarkerSet = true
	} else if cycle.HasMarker {
		if marker, ok := MarkerTypeFromString(cycle.Marker); ok && marker != MarkerNone {
			line.Marker = marker
			line.MarkerSet = true
		}
	}
	if opt.MarkerStyle != nil {
		line.MarkerStyle = *opt.MarkerStyle
	}
	if opt.MarkerPath != nil {
		line.MarkerPath = *opt.MarkerPath
	}
	if opt.MarkerSize != nil {
		line.MarkerSize = *opt.MarkerSize
	}
	line.MarkerFaceColor = color
	if opt.MarkerFaceColor != nil {
		line.MarkerFaceColor = *opt.MarkerFaceColor
	}
	line.MarkerEdgeColor = color
	if opt.MarkerEdgeColor != nil {
		line.MarkerEdgeColor = *opt.MarkerEdgeColor
	}
	if opt.MarkerFaceSpec != nil {
		line.MarkerFaceSpec = *opt.MarkerFaceSpec
	}
	if opt.MarkerEdgeSpec != nil {
		line.MarkerEdgeSpec = *opt.MarkerEdgeSpec
	}
	if opt.MarkerFaceAlt != nil {
		line.MarkerFaceAlt = *opt.MarkerFaceAlt
	}
	if opt.MarkerEdgeWidth != nil {
		line.MarkerEdgeWidth = *opt.MarkerEdgeWidth
	}
	line.MarkEvery = opt.MarkEvery
	if opt.MarkEverySpec != nil {
		line.SetMarkEvery(*opt.MarkEverySpec)
	}

	// Apply alpha if specified
	if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
		line.Col.A = *opt.Alpha
		if opt.MarkerFaceColor == nil {
			line.MarkerFaceColor.A = *opt.Alpha
		}
		if opt.MarkerEdgeColor == nil {
			line.MarkerEdgeColor.A = *opt.Alpha
		}
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
	Color        *render.Color    // if nil, uses automatic color cycling
	Colors       []render.Color   // per-point marker face colors
	ScalarValues []float64        // per-point scalar values mapped through Colormap/Norm
	Colormap     string           // colormap name for ScalarValues
	Norm         ScalarNormalizer // normalizer for ScalarValues
	VMin         *float64         // lower color limit for ScalarValues
	VMax         *float64         // upper color limit for ScalarValues
	Size         *float64         // marker area in points^2
	Sizes        []float64        // per-point marker areas in points^2
	Marker       *MarkerType      // marker type
	MarkerStyle  *MarkerStyle     // marker style; overrides Marker when non-nil
	MarkerPath   *geom.Path       // custom marker path (overrides Marker when non-nil)
	EdgeColor    *render.Color    // edge color
	EdgeColors   []render.Color   // per-point marker edge colors
	EdgeWidth    *float64         // edge width
	Alpha        *float64         // alpha transparency
	Label        string           // series label for legend
	AxLimClip    bool             // 3D scatter: hide points outside explicit axes limits
	// PlotNonfinite mirrors Matplotlib's plotnonfinite kwarg. When false
	// (default), points with a non-finite x/y/size/scalar/color are masked out.
	// When true, points with non-finite color/scalar values are kept and ride
	// the colormap's "bad" color; only non-finite positions/sizes are dropped.
	PlotNonfinite bool
}

// Scatter creates a scatter plot with automatic shape/fill color cycling if no color is specified.
func (a *Axes) Scatter(x, y []float64, opts ...ScatterOptions) *Scatter2D {
	if len(x) == 0 || len(y) == 0 {
		return nil
	}
	if len(x) != len(y) {
		diag.Warnf("Scatter: x and y must be the same size (got %d and %d); skipping", len(x), len(y))
		return nil
	}

	// Create points
	n := len(x)
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
	var sizes []float64
	if len(opt.Sizes) > 0 {
		if !validScatterOptionLength(len(opt.Sizes), n) {
			return nil
		}
		if len(opt.Sizes) == 1 {
			size = opt.Sizes[0]
		} else {
			sizes = cloneFloat64s(opt.Sizes)
		}
	}

	// Get marker type
	marker := MarkerCircle
	if opt.Marker != nil {
		marker = *opt.Marker
	}

	// Matplotlib defaults scatter marker edges to "face" with linewidth 1.
	edgeColor := color
	if opt.EdgeColor != nil {
		edgeColor = *opt.EdgeColor
	}
	var colors []render.Color
	if len(opt.Colors) > 0 {
		if !validScatterOptionLength(len(opt.Colors), n) {
			return nil
		}
		if len(opt.Colors) == 1 {
			color = opt.Colors[0]
			if opt.EdgeColor == nil {
				edgeColor = color
			}
		} else {
			colors = cloneRenderColors(opt.Colors)
		}
	}
	var scalarValues []float64
	var scalarMap ScalarMapInfo
	scalarMapSet := false
	if len(opt.ScalarValues) > 0 {
		if !validScatterOptionLength(len(opt.ScalarValues), n) {
			return nil
		}
		scalarValues = cloneFloat64s(opt.ScalarValues)
		if len(opt.ScalarValues) == 1 && n > 1 {
			scalarValues = make([]float64, n)
			for i := range scalarValues {
				scalarValues[i] = opt.ScalarValues[0]
			}
		}
		mapping, err := ResolveScalarMapValues(scalarValues, ScalarMapConfig{
			Colormap: opt.Colormap,
			Norm:     opt.Norm,
			VMin:     opt.VMin,
			VMax:     opt.VMax,
		})
		if err != nil {
			return nil
		}
		scalarMap = mapping
		scalarMapSet = true
		colors = nil
	}
	var edgeColors []render.Color
	if len(opt.EdgeColors) > 0 {
		if !validScatterOptionLength(len(opt.EdgeColors), n) {
			return nil
		}
		if len(opt.EdgeColors) == 1 {
			edgeColor = opt.EdgeColors[0]
		} else {
			edgeColors = cloneRenderColors(opt.EdgeColors)
		}
	}

	edgeWidth := 1.0
	if opt.EdgeWidth != nil {
		edgeWidth = *opt.EdgeWidth
	}

	// Get alpha
	alpha := 1.0
	if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
		alpha = *opt.Alpha
	}

	// Drop points Matplotlib would mask for being non-finite (cbook._combine_masks
	// in axes._axes.scatter). vmin/vmax above were autoscaled via finiteRange, so
	// the scalar map already ignores the masked values.
	points, sizes, colors, edgeColors, scalarValues = filterScatterFinite(
		points, sizes, colors, edgeColors, scalarValues, opt.PlotNonfinite,
	)

	// Create scatter
	scatter := &Scatter2D{
		XY:           points,
		Sizes:        sizes,
		Colors:       colors,
		EdgeColors:   edgeColors,
		ScalarValues: scalarValues,
		Size:         size,
		Color:        color,
		EdgeColor:    edgeColor,
		EdgeWidth:    edgeWidth,
		Alpha:        alpha,
		Marker:       marker,
		Label:        opt.Label,
	}
	if scalarMapSet {
		scatter.Colormap = scalarMap.Colormap
		scatter.Norm = scalarMap.Norm
		scatter.VMin = scalarMap.VMin
		scatter.VMax = scalarMap.VMax
		scatter.scalarCLimSet = true
		scatter.EdgeColorsFace = opt.EdgeColor == nil && len(opt.EdgeColors) == 0
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

func validScatterOptionLength(length, n int) bool {
	return length == 1 || length == n
}

// filterScatterFinite drops scatter points Matplotlib would mask for being
// non-finite, rebuilding every per-point array in lockstep. Position and size
// must always be finite to render; color/scalar finiteness only masks a point
// when plotNonfinite is false (otherwise non-finite color/scalar values ride the
// colormap's bad color). Per-point arrays that are nil (single-value fallbacks)
// stay nil. Mirrors cbook._combine_masks used by axes._axes.scatter.
func filterScatterFinite(points []geom.Pt, sizes []float64, colors, edgeColors []render.Color, scalars []float64, plotNonfinite bool) ([]geom.Pt, []float64, []render.Color, []render.Color, []float64) {
	anyDropped := false
	keep := make([]bool, len(points))
	for i := range points {
		ok := isFinite(points[i].X) && isFinite(points[i].Y)
		if ok && i < len(sizes) && !isFinite(sizes[i]) {
			ok = false
		}
		if ok && !plotNonfinite {
			if i < len(scalars) && !isFinite(scalars[i]) {
				ok = false
			}
			if ok && i < len(colors) && !renderColorFinite(colors[i]) {
				ok = false
			}
			if ok && i < len(edgeColors) && !renderColorFinite(edgeColors[i]) {
				ok = false
			}
		}
		keep[i] = ok
		if !ok {
			anyDropped = true
		}
	}
	if !anyDropped {
		return points, sizes, colors, edgeColors, scalars
	}

	fp := make([]geom.Pt, 0, len(points))
	var fs []float64
	var fc, fe []render.Color
	var fv []float64
	if len(sizes) > 0 {
		fs = make([]float64, 0, len(sizes))
	}
	if len(colors) > 0 {
		fc = make([]render.Color, 0, len(colors))
	}
	if len(edgeColors) > 0 {
		fe = make([]render.Color, 0, len(edgeColors))
	}
	if len(scalars) > 0 {
		fv = make([]float64, 0, len(scalars))
	}
	for i := range points {
		if !keep[i] {
			continue
		}
		fp = append(fp, points[i])
		if i < len(sizes) {
			fs = append(fs, sizes[i])
		}
		if i < len(colors) {
			fc = append(fc, colors[i])
		}
		if i < len(edgeColors) {
			fe = append(fe, edgeColors[i])
		}
		if i < len(scalars) {
			fv = append(fv, scalars[i])
		}
	}
	return fp, fs, fc, fe, fv
}

// renderColorFinite reports whether every component of a color is finite.
func renderColorFinite(c render.Color) bool {
	return isFinite(c.R) && isFinite(c.G) && isFinite(c.B) && isFinite(c.A)
}

// BarOptions holds optional parameters for bar plots.
type BarAlign uint8

const (
	BarAlignCenter BarAlign = iota
	BarAlignEdge
)

type BarOptions struct {
	Color       *render.Color   // if nil, uses automatic color cycling
	Colors      []render.Color  // per-bar fill colors
	Width       *float64        // bar width
	Widths      []float64       // per-bar widths
	EdgeColor   *render.Color   // edge color
	EdgeColors  []render.Color  // per-bar edge colors
	EdgeWidth   *float64        // edge width
	Alpha       *float64        // alpha transparency
	Baseline    *float64        // baseline value
	Baselines   []float64       // per-bar baseline/left values
	Orientation *BarOrientation // vertical or horizontal
	Align       *BarAlign       // center or edge alignment
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
	var widths []float64
	if len(opt.Widths) > 0 {
		if !validBarOptionLength(len(opt.Widths), len(x)) {
			return nil
		}
		if len(opt.Widths) == 1 {
			width = opt.Widths[0]
		} else {
			widths = cloneFloat64s(opt.Widths)
		}
	}

	// Get edge properties
	edgeColor := render.Color{R: 0, G: 0, B: 0, A: 0} // transparent by default
	if opt.EdgeColor != nil {
		edgeColor = *opt.EdgeColor
	}
	var colors []render.Color
	if len(opt.Colors) > 0 {
		if !validBarOptionLength(len(opt.Colors), len(x)) {
			return nil
		}
		if len(opt.Colors) == 1 {
			color = opt.Colors[0]
		} else {
			colors = cloneRenderColors(opt.Colors)
		}
	}
	var edgeColors []render.Color
	if len(opt.EdgeColors) > 0 {
		if !validBarOptionLength(len(opt.EdgeColors), len(x)) {
			return nil
		}
		if len(opt.EdgeColors) == 1 {
			edgeColor = opt.EdgeColors[0]
		} else {
			edgeColors = cloneRenderColors(opt.EdgeColors)
		}
	}

	edgeWidth := 0.0
	if opt.EdgeWidth != nil {
		edgeWidth = *opt.EdgeWidth
	}

	// Alpha override: an explicit value (including 0 for fully transparent) is
	// baked into the resolved colors so it is honored, while a nil alpha leaves
	// each color's own alpha intact.
	color = bakeExplicitAlpha(color, opt.Alpha)
	edgeColor = bakeExplicitAlpha(edgeColor, opt.Alpha)
	for i := range colors {
		colors[i] = bakeExplicitAlpha(colors[i], opt.Alpha)
	}
	for i := range edgeColors {
		edgeColors[i] = bakeExplicitAlpha(edgeColors[i], opt.Alpha)
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
	align := BarAlignCenter
	if opt.Align != nil {
		align = *opt.Align
	}
	positions := cloneFloat64s(x)
	if align == BarAlignEdge {
		for i := range positions {
			positions[i] += barWidthAt(width, widths, i) / 2
		}
	}

	// Create bar chart
	bar := &Bar2D{
		X:           positions,
		Heights:     heights,
		Width:       width,
		Widths:      widths,
		Baselines:   append([]float64(nil), opt.Baselines...),
		Color:       color,
		Colors:      colors,
		EdgeColor:   edgeColor,
		EdgeColors:  edgeColors,
		EdgeWidth:   edgeWidth,
		Baseline:    baseline,
		Orientation: orientation,
		Label:       opt.Label,
	}

	a.Add(bar)
	a.autoScaleIfEnabled(defaultAutoScaleMargin)
	return bar
}

func validBarOptionLength(length, n int) bool {
	return length == 1 || length == n
}

// bakeExplicitAlpha returns c with its alpha replaced by *alpha when alpha is an
// explicit value in [0,1]. An explicit 0 yields a fully transparent color; a nil
// alpha leaves the color (and its own alpha channel) untouched. Baking the
// override into the resolved color at construction lets artists honor an
// explicit alpha=0 — which a plain float64 "0 means unset" field cannot express.
func bakeExplicitAlpha(c render.Color, alpha *float64) render.Color {
	if alpha != nil && *alpha >= 0 && *alpha <= 1 {
		c.A = *alpha
	}
	return c
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

	color = bakeExplicitAlpha(color, opt.Alpha)
	edgeColor = bakeExplicitAlpha(edgeColor, opt.Alpha)

	fill := &PolyCollection{
		PatchCollection: PatchCollection{
			Collection: Collection{
				Label: opt.Label,
				Alpha: 1,
				// matplotlib's fill() returns Polygon patches (Patch.zorder=1),
				// which draw below gridlines (axisbelow='line' puts the axis at 1.5).
				z: 1,
			},
			FaceColors: []render.Color{color},
			EdgeColor:  edgeColor,
			EdgeWidth:  edgeWidth,
		},
		Polygons: [][]geom.Pt{points},
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
	if len(opt.Where) > 0 && len(opt.Where) != len(y) {
		diag.Warnf("FillBetween: where length %d must match y length %d; skipping", len(opt.Where), len(y))
		return nil
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

	color = bakeExplicitAlpha(color, opt.Alpha)
	edgeColor = bakeExplicitAlpha(edgeColor, opt.Alpha)

	fill := &Fill2D{
		X:           y,
		Y1:          x1,
		Y2:          x2,
		Where:       append([]bool(nil), opt.Where...),
		Interpolate: opt.Interpolate,
		Step:        opt.Step,
		Orientation: FillHorizontal,
		Color:       color,
		EdgeColor:   edgeColor,
		EdgeWidth:   edgeWidth,
		Label:       opt.Label,
	}

	a.Add(fill)
	a.autoScaleIfEnabled(defaultAutoScaleMargin)
	return fill
}

// FillOptions holds optional parameters for fill plots.
type FillOptions struct {
	Color       *render.Color // if nil, uses automatic color cycling
	EdgeColor   *render.Color // edge color
	EdgeWidth   *float64      // edge width
	Alpha       *float64      // alpha transparency
	Baseline    *float64      // baseline value
	Where       []bool        // fill only contiguous regions where adjacent points are true
	Interpolate bool          // interpolate region boundaries at curve crossings
	Step        FillStep      // optional step mode
	Label       string        // series label for legend
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
	if len(opt.Where) > 0 && len(opt.Where) != len(x) {
		diag.Warnf("FillBetweenX: where length %d must match x length %d; skipping", len(opt.Where), len(x))
		return nil
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

	// When alpha is omitted, preserve the color's own alpha, matching
	// Matplotlib's fill_between behavior; an explicit value (including 0) wins.
	color = bakeExplicitAlpha(color, opt.Alpha)
	edgeColor = bakeExplicitAlpha(edgeColor, opt.Alpha)

	// Create fill
	fill := &Fill2D{
		X:           x,
		Y1:          y1,
		Y2:          y2,
		Where:       append([]bool(nil), opt.Where...),
		Interpolate: opt.Interpolate,
		Step:        opt.Step,
		Color:       color,
		EdgeColor:   edgeColor,
		EdgeWidth:   edgeWidth,
		Label:       opt.Label,
	}

	a.Add(fill)
	a.autoScaleIfEnabled(defaultAutoScaleMargin)
	return fill
}

// HistOptions holds optional parameters for histogram plots.
type HistOptions struct {
	Bins              int         // number of bins (0 = auto)
	BinEdges          []float64   // explicit bin edges (overrides Bins)
	Range             *HistRange  // explicit histogram range; ignored when BinEdges is set
	Weights           []float64   // per-sample weights, same length as data when provided
	BinStrat          BinStrategy // automatic binning strategy
	Norm              HistNorm    // normalization mode
	Cumulative        bool        // accumulate bin heights from left to right
	ReverseCumulative bool        // accumulate from right to left, matching cumulative < 0
	HistType          HistType    // bar, step, or filled step presentation
	Baselines         []float64   // optional per-bin baselines for stacked histograms
	Color             *render.Color
	EdgeColor         *render.Color
	EdgeWidth         *float64
	Alpha             *float64
	Label             string
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
	if len(opt.Weights) > 0 && len(opt.Weights) != len(data) {
		diag.Warnf("Hist: weights length %d must match data length %d; skipping", len(opt.Weights), len(data))
		return nil
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

	// Bake an explicit alpha (including 0 for fully transparent) into the
	// resolved colors so it is honored; the Alpha field stays at the 0
	// "unset" sentinel and Draw's override is a no-op. A nil alpha preserves
	// each color's own channel.
	color = bakeExplicitAlpha(color, opt.Alpha)
	edgeColor = bakeExplicitAlpha(edgeColor, opt.Alpha)

	hist := &Hist2D{
		Data:              data,
		Weights:           append([]float64(nil), opt.Weights...),
		Bins:              opt.Bins,
		BinEdges:          opt.BinEdges,
		Range:             opt.Range,
		BinStrat:          opt.BinStrat,
		Norm:              opt.Norm,
		Cumulative:        opt.Cumulative,
		ReverseCumulative: opt.ReverseCumulative,
		HistType:          opt.HistType,
		Baselines:         append([]float64(nil), opt.Baselines...),
		Color:             color,
		EdgeColor:         edgeColor,
		EdgeWidth:         edgeWidth,
		Label:             opt.Label,
	}

	a.Add(hist)
	return hist
}

// ErrorBarOptions holds optional parameters for error bar plots.
type ErrorBarOptions struct {
	Color           *render.Color // if nil, uses automatic color cycling
	LineWidth       *float64      // error bar line width (px); nil uses Matplotlib's default
	CapSize         *float64      // Matplotlib capsize in points
	CapThick        *float64      // Matplotlib capthick in points (cap line thickness); nil uses the 1pt default
	Marker          *MarkerType   // optional data marker equivalent to Matplotlib fmt markers
	MarkerSize      *float64      // marker size in points
	Alpha           *float64      // alpha transparency
	Label           string        // series label for legend
	NoDataLine      bool          // true matches Matplotlib fmt="none"
	ErrorEvery      int           // draw error bars every N points, default 1
	ErrorEveryStart int           // starting point for ErrorEvery, matching errorevery=(start,N)

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

	lineWidth := 0.0
	if opt.LineWidth != nil {
		lineWidth = *opt.LineWidth
	}

	capSizePx := 0.0
	if opt.CapSize != nil {
		capSizePx = pointsToPixels(a.resolvedRC(), 2*(*opt.CapSize))
	}

	capThickPx := 0.0
	if opt.CapThick != nil && *opt.CapThick > 0 {
		capThickPx = pointsToPixels(a.resolvedRC(), *opt.CapThick)
	}

	// Bake an explicit alpha (including 0) into the stroke color; the Alpha
	// field stays unset so Draw's alpha multiplier is the identity.
	color = bakeExplicitAlpha(color, opt.Alpha)

	n := len(x)
	if len(y) < n {
		n = len(y)
	}
	if !validErrorValues(xErr, n) || !validErrorValues(yErr, n) ||
		!validErrorValues(opt.XErrLower, n) || !validErrorValues(opt.XErrUpper, n) ||
		!validErrorValues(opt.YErrLower, n) || !validErrorValues(opt.YErrUpper, n) ||
		!validBoolValues(opt.LoLimits, n) || !validBoolValues(opt.UpLimits, n) ||
		!validBoolValues(opt.XLoLimits, n) || !validBoolValues(opt.XUpLimits, n) {
		diag.Warnf("ErrorBar: error/limit arrays must each be empty or length %d; skipping", n)
		return nil
	}
	if opt.ErrorEvery < 0 || (opt.ErrorEvery == 0 && opt.ErrorEveryStart != 0) || opt.ErrorEveryStart < 0 {
		diag.Warnf("ErrorBar: invalid errorevery (every=%d, start=%d); skipping", opt.ErrorEvery, opt.ErrorEveryStart)
		return nil
	}
	errorEvery := opt.ErrorEvery
	if errorEvery == 0 {
		errorEvery = 1
	}

	pts := make([]geom.Pt, n)
	for i := 0; i < n; i++ {
		pts[i] = geom.Pt{X: x[i], Y: y[i]}
	}

	bar := &ErrorBar{
		XY:              pts,
		XErr:            xErr,
		YErr:            yErr,
		XErrLower:       append([]float64(nil), opt.XErrLower...),
		XErrUpper:       append([]float64(nil), opt.XErrUpper...),
		YErrLower:       append([]float64(nil), opt.YErrLower...),
		YErrUpper:       append([]float64(nil), opt.YErrUpper...),
		LoLimits:        append([]bool(nil), opt.LoLimits...),
		UpLimits:        append([]bool(nil), opt.UpLimits...),
		XLoLimits:       append([]bool(nil), opt.XLoLimits...),
		XUpLimits:       append([]bool(nil), opt.XUpLimits...),
		Color:           color,
		LineWidth:       lineWidth,
		CapSize:         capSizePx,
		CapThick:        capThickPx,
		Label:           opt.Label,
		NoDataLine:      opt.NoDataLine,
		ErrorEvery:      errorEvery,
		ErrorEveryStart: opt.ErrorEveryStart,
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
	FlierSize    *float64      // outlier marker size in points
	Alpha        *float64      // alpha transparency
	ShowFliers   *bool         // whether to draw outliers
	Label        string        // series label for legend

	PatchArtist *bool         // fill the box (Matplotlib patch_artist=True); default is unfilled
	Orientation *string       // "vertical" (default) or "horizontal"
	ShowBox     *bool         // whether to draw the box (default true)
	ShowCaps    *bool         // whether to draw the whisker caps (default true)
	ShowMeans   *bool         // whether to draw the mean
	MeanLine    *bool         // draw the mean as a line across the box instead of a marker
	MeanColor   *render.Color // mean line/marker color
	Whis        *float64      // IQR multiplier for whiskers (default 1.5)
	Sym         *string       // Matplotlib flier format string, e.g. "b+"; "" disables fliers

	Notch              *bool       // draw a notched box using the confidence interval
	Bootstrap          int         // number of bootstrap resamples for the notch CI (0 = analytic)
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
	FlierSize    *float64       // outlier marker size in points
	Alpha        *float64       // alpha transparency
	ShowFliers   *bool          // whether to draw outliers
	ManageTicks  *bool          // whether to place position ticks at box positions
	Labels       []string       // series labels for legend

	PatchArtist *bool         // fill the boxes (Matplotlib patch_artist=True); default is unfilled
	Orientation *string       // "vertical" (default) or "horizontal"
	ShowBox     *bool         // whether to draw the box (default true)
	ShowCaps    *bool         // whether to draw the whisker caps (default true)
	ShowMeans   *bool         // whether to draw the mean
	MeanLine    *bool         // draw the mean as a line across the box instead of a marker
	MeanColor   *render.Color // mean line/marker color
	Whis        *float64      // IQR multiplier for whiskers (default 1.5)
	Sym         *string       // Matplotlib flier format string, e.g. "b+"; "" disables fliers

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

	rc := a.resolvedRC()

	position := 1.0
	if opt.Position != nil {
		position = *opt.Position
	}

	width := 0.6
	if opt.Width != nil {
		width = *opt.Width
	}

	// Matplotlib's default boxplot is patch_artist=False: an unfilled box that
	// does not consume the color cycle. Only fill (and default the facecolor to
	// white) when patch_artist is requested.
	patchArtist := rc.Boxplot.PatchArtist
	if opt.PatchArtist != nil {
		patchArtist = *opt.PatchArtist
	}
	color := render.Color{}
	if patchArtist {
		color = render.Color{R: 1, G: 1, B: 1, A: 1}
	}
	if opt.Color != nil {
		color = *opt.Color
	}

	edgeColor := render.Color{R: 0, G: 0, B: 0, A: 1}
	if opt.EdgeColor != nil {
		edgeColor = *opt.EdgeColor
	}

	medianColor := rc.Boxplot.MedianColor
	if opt.MedianColor != nil {
		medianColor = *opt.MedianColor
	}

	meanColor := rc.Boxplot.MeanColor
	if opt.MeanColor != nil {
		meanColor = *opt.MeanColor
	}

	whiskerColor := edgeColor
	if opt.WhiskerColor != nil {
		whiskerColor = *opt.WhiskerColor
	}

	capColor := whiskerColor
	if opt.CapColor != nil {
		capColor = *opt.CapColor
	}

	flierColor := render.Color{}
	if opt.FlierColor != nil {
		flierColor = *opt.FlierColor
	}

	edgeWidth := pointsToPixels(rc, rc.Boxplot.BoxLineWidth)
	if opt.EdgeWidth != nil {
		edgeWidth = *opt.EdgeWidth
	}

	whiskerWidth := pointsToPixels(rc, rc.Boxplot.WhiskerLineWidth)
	if opt.WhiskerWidth != nil {
		whiskerWidth = *opt.WhiskerWidth
	}

	medianWidth := pointsToPixels(rc, rc.Boxplot.MedianLineWidth)
	if opt.MedianWidth != nil {
		medianWidth = *opt.MedianWidth
	}

	capWidth := width * 0.5
	if opt.CapWidth != nil {
		capWidth = *opt.CapWidth
	}

	flierSize := rc.Boxplot.FlierMarkerSize
	if opt.FlierSize != nil {
		flierSize = *opt.FlierSize
	}

	// Bake an explicit alpha (including 0) into the box fill and edge colors —
	// the only colors Draw applies alpha to — leaving the Alpha field unset so
	// its multiplier is the identity.
	color = bakeExplicitAlpha(color, opt.Alpha)
	edgeColor = bakeExplicitAlpha(edgeColor, opt.Alpha)

	showFliers := rc.Boxplot.ShowFliers
	if opt.ShowFliers != nil {
		showFliers = *opt.ShowFliers
	}
	showBox := rc.Boxplot.ShowBox
	if opt.ShowBox != nil {
		showBox = *opt.ShowBox
	}
	showCaps := rc.Boxplot.ShowCaps
	if opt.ShowCaps != nil {
		showCaps = *opt.ShowCaps
	}
	showMeans := rc.Boxplot.ShowMeans
	if opt.ShowMeans != nil {
		showMeans = *opt.ShowMeans
	}
	meanLine := rc.Boxplot.MeanLine
	if opt.MeanLine != nil {
		meanLine = *opt.MeanLine
	}
	orientation := ""
	if opt.Orientation != nil {
		orientation = *opt.Orientation
	}
	notch := rc.Boxplot.Notch
	if opt.Notch != nil {
		notch = *opt.Notch
	}
	flierMarker := MarkerCircle
	if opt.FlierMarker != nil {
		flierMarker = *opt.FlierMarker
	}
	flierEdgeColor := rc.Boxplot.FlierColor
	if opt.FlierColor != nil {
		flierEdgeColor = flierColor
	}
	if opt.FlierEdgeColor != nil {
		flierEdgeColor = *opt.FlierEdgeColor
	}
	flierEdgeWidth := pointsToPixels(rc, rc.Boxplot.FlierEdgeWidth)
	if opt.FlierEdgeWidth != nil {
		flierEdgeWidth = *opt.FlierEdgeWidth
	}

	// Matplotlib's sym shorthand overrides the flier marker/color; sym="" hides
	// fliers entirely. Structured FlierMarker/FlierColor options take precedence.
	if opt.Sym != nil {
		if *opt.Sym == "" {
			showFliers = false
		} else {
			marker, symColor, hasMarker, hasColor := parseBoxplotSym(*opt.Sym)
			if hasMarker && opt.FlierMarker == nil {
				flierMarker = marker
			}
			if hasColor && opt.FlierColor == nil {
				flierColor = symColor
				flierEdgeColor = symColor
			}
		}
	}

	box := &BoxPlot2D{
		Data:               data,
		Position:           position,
		Width:              width,
		Color:              color,
		EdgeColor:          edgeColor,
		MedianColor:        medianColor,
		MeanColor:          meanColor,
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
		PatchArtist:        patchArtist,
		Orientation:        orientation,
		ShowBox:            showBox,
		ShowCaps:           showCaps,
		ShowFliers:         showFliers,
		ShowMeans:          showMeans,
		MeanLine:           meanLine,
		Notch:              notch,
		Bootstrap:          opt.Bootstrap,
		Whis:               opt.Whis,
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
	positions := make([]float64, 0, len(datasets))
	for i, data := range datasets {
		position := float64(i + 1)
		if i < len(opt.Positions) {
			position = opt.Positions[i]
		}
		positions = append(positions, position)

		boxOpt := BoxPlotOptions{
			Position:           &position,
			Width:              width,
			EdgeColor:          opt.EdgeColor,
			MedianColor:        opt.MedianColor,
			MeanColor:          opt.MeanColor,
			WhiskerColor:       opt.WhiskerColor,
			CapColor:           opt.CapColor,
			FlierColor:         opt.FlierColor,
			EdgeWidth:          opt.EdgeWidth,
			WhiskerWidth:       opt.WhiskerWidth,
			MedianWidth:        opt.MedianWidth,
			CapWidth:           opt.CapWidth,
			FlierSize:          opt.FlierSize,
			Alpha:              opt.Alpha,
			PatchArtist:        opt.PatchArtist,
			Orientation:        opt.Orientation,
			ShowBox:            opt.ShowBox,
			ShowCaps:           opt.ShowCaps,
			ShowFliers:         opt.ShowFliers,
			ShowMeans:          opt.ShowMeans,
			MeanLine:           opt.MeanLine,
			Notch:              opt.Notch,
			Bootstrap:          opt.Bootstrap,
			Whis:               opt.Whis,
			WhiskerPercentiles: opt.WhiskerPercentiles,
			FlierMarker:        opt.FlierMarker,
			FlierEdgeColor:     opt.FlierEdgeColor,
			FlierEdgeWidth:     opt.FlierEdgeWidth,
			Sym:                opt.Sym,
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
	manageTicks := true
	if opt.ManageTicks != nil {
		manageTicks = *opt.ManageTicks
	}
	if manageTicks && len(positions) > 0 {
		// Matplotlib boxplot(..., manage_ticks=True) places the position-axis ticks
		// at the box positions by default — the y axis for horizontal orientation.
		horizontal := opt.Orientation != nil && normalizeViolinOrientation(*opt.Orientation) == "horizontal"
		if horizontal {
			if a.YAxis != nil {
				a.YAxis.Locator = FixedLocator{TicksList: positions}
			}
		} else if a.XAxis != nil {
			a.XAxis.Locator = FixedLocator{TicksList: positions}
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

// boxplotSymColors maps Matplotlib's single-letter color shorthands to RGBA.
var boxplotSymColors = map[byte]render.Color{
	'b': {R: 0, G: 0, B: 1, A: 1},
	'g': {R: 0, G: 0.5, B: 0, A: 1},
	'r': {R: 1, G: 0, B: 0, A: 1},
	'c': {R: 0, G: 0.75, B: 0.75, A: 1},
	'm': {R: 0.75, G: 0, B: 0.75, A: 1},
	'y': {R: 0.75, G: 0.75, B: 0, A: 1},
	'k': {R: 0, G: 0, B: 0, A: 1},
	'w': {R: 1, G: 1, B: 1, A: 1},
}

// parseBoxplotSym splits a Matplotlib flier format string (e.g. "b+", "ro")
// into a marker and color, mirroring _process_plot_format for the subset that
// boxplot's sym shorthand uses. A leading or trailing single-letter color code
// is recognized; the remainder is parsed as a marker via MarkerTypeFromString.
func parseBoxplotSym(sym string) (marker MarkerType, color render.Color, hasMarker, hasColor bool) {
	rest := sym
	if rest != "" {
		if c, ok := boxplotSymColors[rest[0]]; ok {
			color, hasColor = c, true
			rest = rest[1:]
		} else if c, ok := boxplotSymColors[rest[len(rest)-1]]; ok {
			color, hasColor = c, true
			rest = rest[:len(rest)-1]
		}
	}
	if rest != "" {
		if m, ok := MarkerTypeFromString(rest); ok {
			marker, hasMarker = m, true
		}
	}
	return marker, color, hasMarker, hasColor
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

	// When alpha is omitted, preserve the color's own alpha, matching
	// Matplotlib's fill_between behavior; an explicit value (including 0) wins.
	color = bakeExplicitAlpha(color, opt.Alpha)
	edgeColor = bakeExplicitAlpha(edgeColor, opt.Alpha)

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
		Label:     opt.Label,
	}

	a.Add(fill)
	a.autoScaleIfEnabled(defaultAutoScaleMargin)
	return fill
}
