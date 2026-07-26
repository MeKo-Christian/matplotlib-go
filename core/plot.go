package core

import (
	"fmt"
	"math"
	"strings"

	mplcolor "github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/ticker"
	"github.com/cwbudde/matplotlib-go/transform"
)

const defaultAutoScaleMargin = 0.05

// PlotOptions holds optional parameters for plotting functions.
type PlotOptions struct {
	Color           optional.Value[render.Color] // unset uses automatic color cycling
	EdgeColor       optional.Value[render.Color]
	LineWidth       optional.Value[float64] // unset uses the cycled or rc width
	LineCap         optional.Value[render.LineCap]
	LineJoin        optional.Value[render.LineJoin]
	EdgeWidth       optional.Value[float64]
	Dashes          []float64     // dash pattern in pixels; overrides LineStyle
	LineStyle       LineStyle     // typed matplotlib linestyle ("-", "--", "-.", ":", "none"); "" = unset
	DrawStyle       LineDrawStyle // zero value connects points directly
	Marker          optional.Value[MarkerType]
	MarkerStyle     optional.Value[MarkerStyle]
	MarkerPath      optional.Value[geom.Path]
	MarkerSize      optional.Value[float64]
	MarkerFaceColor optional.Value[render.Color]
	MarkerEdgeColor optional.Value[render.Color]
	MarkerFaceSpec  optional.Value[MarkerColorSpec]
	MarkerEdgeSpec  optional.Value[MarkerColorSpec]
	MarkerFaceAlt   MarkerColorSpec // zero value leaves the alternate face color unset
	MarkerEdgeWidth optional.Value[float64]
	MarkEvery       int
	MarkEverySpec   MarkEverySpec // zero value draws a marker at every point
	Label           string        // series label for legend
	Alpha           optional.Value[float64]
	LevelCount      int // contour level count for contour-like plot types
	Levels          []float64
	ZDir            string
	Offset          optional.Value[float64] // fixed projection offset for contour-like plot types
	RStride         optional.Value[int]     // row stride for 3D surface/wireframe sampling
	CStride         optional.Value[int]     // column stride for 3D surface/wireframe sampling
	RCount          optional.Value[int]     // maximum sampled row count for 3D surface/wireframe sampling
	CCount          optional.Value[int]     // maximum sampled column count for 3D surface/wireframe sampling
	FaceColors      []render.Color
	Shade           optional.Value[bool]
	Antialiased     optional.Value[bool]
	Colormap        optional.Value[string] // scalar colormap for mappable plot types
	Norm            ScalarNormalizer
	VMin            optional.Value[float64]
	VMax            optional.Value[float64]
	AxLimClip       bool
}

// ScalarMapConfig returns the scalar-map portion of PlotOptions.
//
// PlotOptions is shared by the core and plot3d plotting surfaces. Keeping this
// conversion beside the option definition ensures both surfaces pass the same
// colormap, normalizer, and explicit limits to ResolveScalarMapValues.
func (o PlotOptions) ScalarMapConfig() ScalarMapConfig {
	return ScalarMapConfig{
		Colormap: o.Colormap.OrZero(),
		Norm:     o.Norm,
		VMin:     o.VMin,
		VMax:     o.VMax,
	}
}

// Plot converts x and y through the axes units machinery and creates a line
// plot with automatic color cycling if no color is specified.
//
// Rejected input leaves the axes, its unit configuration, and its property
// cycle unchanged.
//
//nolint:gocritic // PlotOptions is an immutable snapshot retained by the line artist.
func (a *Axes) Plot(xVals, yVals any, opt PlotOptions) (*Line2D, error) {
	if a == nil {
		return nil, fmt.Errorf("plot axes cannot be nil")
	}
	tx := a.beginUnitConversion()
	x, err := a.convertValues(xVals, true)
	if err != nil {
		tx.rollback()
		return nil, fmt.Errorf("plot x values: %w", err)
	}
	y, err := a.convertValues(yVals, false)
	if err != nil {
		tx.rollback()
		return nil, fmt.Errorf("plot y values: %w", err)
	}
	if len(x) == 0 || len(y) == 0 {
		tx.rollback()
		return nil, fmt.Errorf("plot x and y values cannot be empty")
	}
	if len(x) != len(y) {
		tx.rollback()
		return nil, fmt.Errorf("plot x and y must have the same length (got %d and %d)", len(x), len(y))
	}

	line := a.plot(x, y, opt)
	tx.commit()
	return line, nil
}

// plot draws an already validated and unit-converted line. It takes its options
// as a single value so the "at most one" rule is settled by the exported entry
// point that called it.
//
//nolint:gocritic // PlotOptions is an immutable snapshot retained by the line artist.
func (a *Axes) plot(x, y []float64, opt PlotOptions) *Line2D {
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

	// Pull one step from the property cycle (color plus any
	// linestyle/marker/linewidth carried by axes.prop_cycle).
	cycle := a.NextLineProps()
	color := opt.Color.Or(cycle.Color)

	// Get line width: explicit option, else cycled linewidth, else the
	// lines.linewidth rc value (matplotlib default 1.5 points); converted to
	// device pixels at the Line2D Paint sink.
	lineWidth := 1.5
	if rcW := a.resolvedRC().LineWidth; rcW > 0 {
		lineWidth = rcW
	}
	if cycle.HasLineWidth {
		lineWidth = cycle.LineWidth
	}
	lineWidth = opt.LineWidth.Or(lineWidth)

	// Resolve dashes: an explicit dash pattern wins, then the typed LineStyle
	// option, then a cycled linestyle, then a non-default lines.linestyle rc
	// value. LineStyleNone (or rc "none") suppresses the line entirely
	// (markers-only), like matplotlib's linestyle "none".
	resolvedRC := a.resolvedRC()
	rcLines := resolvedRC.Lines
	widthPx := pointsToPixels(resolvedRC, lineWidth)
	pointPx := pointsToPixels(resolvedRC, 1)
	dashes := opt.Dashes
	switch {
	case dashes != nil:
	case opt.LineStyle.isNone():
		lineWidth = 0
	case opt.LineStyle != "":
		dashes = lineStyleToDashesRC(string(opt.LineStyle), widthPx, pointPx, &rcLines)
	case cycle.HasLineStyle:
		dashes = lineStyleToDashesRC(cycle.LineStyle, widthPx, pointPx, &rcLines)
	case rcLines.LineStyle != "" && rcLines.LineStyle != "-":
		if LineStyle(rcLines.LineStyle).isNone() {
			lineWidth = 0
		} else {
			dashes = lineStyleToDashesRC(rcLines.LineStyle, widthPx, pointPx, &rcLines)
		}
	}

	// Create line
	line := &Line2D{
		XY:                points,
		W:                 lineWidth,
		Col:               color,
		Dashes:            dashes,
		DrawStyle:         opt.DrawStyle,
		Label:             opt.Label,
		DashCap:           rcLines.DashCap,
		DashJoin:          rcLines.DashJoin,
		SolidCap:          rcLines.SolidCap,
		SolidJoin:         rcLines.SolidJoin,
		RCStrokeStylesSet: true,
	}
	if lineCap, ok := opt.LineCap.Get(); ok {
		line.LineCap = lineCap
		line.LineCapSet = true
	}
	if lineJoin, ok := opt.LineJoin.Get(); ok {
		line.LineJoin = lineJoin
		line.LineJoinSet = true
	}
	if marker, ok := opt.Marker.Get(); ok {
		line.Marker = marker
		line.MarkerSet = true
	} else if cycle.HasMarker {
		if marker, ok := MarkerTypeFromString(cycle.Marker); ok && marker != MarkerNone {
			line.Marker = marker
			line.MarkerSet = true
		}
	} else if rcMarker := rcLines.Marker; rcMarker != "" && !strings.EqualFold(rcMarker, "none") {
		// A non-default lines.marker rc value seeds the default marker.
		if marker, ok := MarkerTypeFromString(rcMarker); ok && marker != MarkerNone {
			line.Marker = marker
			line.MarkerSet = true
		}
	}
	if markerStyle, ok := opt.MarkerStyle.Get(); ok {
		line.MarkerStyle = markerStyle
	} else if line.MarkerSet {
		line.MarkerStyle = NewMarkerStyle(line.Marker)
		line.MarkerStyle.FillStyle = markerFillStyleFromRC(rcLines.MarkerFillStyle)
	}
	if markerPath, ok := opt.MarkerPath.Get(); ok {
		line.MarkerPath = markerPath
	}
	if markerSize, ok := opt.MarkerSize.Get(); ok {
		line.MarkerSize = markerSize
	} else if rcLines.MarkerSize > 0 && rcLines.MarkerSize != 6 {
		// Line2D treats 0 as "use the 6 pt default", so only a non-default
		// lines.markersize rc value needs seeding.
		line.MarkerSize = rcLines.MarkerSize
	}
	markerFaceColor, markerFaceColorSet := opt.MarkerFaceColor.Get()
	line.MarkerFaceColor = color
	if markerFaceColorSet {
		line.MarkerFaceColor = markerFaceColor
		line.MarkerFaceSpec = ExplicitMarkerColor(markerFaceColor)
	} else if !opt.MarkerFaceSpec.IsSet() {
		line.MarkerFaceSpec = markerColorSpecFromRC(rcLines.MarkerFaceColor, &resolvedRC)
	}
	markerEdgeColor, markerEdgeColorSet := opt.MarkerEdgeColor.Get()
	line.MarkerEdgeColor = color
	if markerEdgeColorSet {
		line.MarkerEdgeColor = markerEdgeColor
		line.MarkerEdgeSpec = ExplicitMarkerColor(markerEdgeColor)
	} else if !opt.MarkerEdgeSpec.IsSet() {
		line.MarkerEdgeSpec = markerColorSpecFromRC(rcLines.MarkerEdgeColor, &resolvedRC)
	}
	if spec, ok := opt.MarkerFaceSpec.Get(); ok {
		line.MarkerFaceSpec = spec
	}
	if spec, ok := opt.MarkerEdgeSpec.Get(); ok {
		line.MarkerEdgeSpec = spec
	}
	line.Antialiased = opt.Antialiased.Or(rcLines.Antialiased)
	line.AntialiasedSet = true
	line.MarkerFaceAlt = opt.MarkerFaceAlt
	if markerEdgeWidth, ok := opt.MarkerEdgeWidth.Get(); ok {
		line.MarkerEdgeWidth = markerEdgeWidth
	} else if rcLines.MarkerEdgeWidth > 0 && rcLines.MarkerEdgeWidth != 1 {
		// Line2D treats 0 as "use the 1 pt default", so only a non-default
		// lines.markeredgewidth rc value needs seeding.
		line.MarkerEdgeWidth = rcLines.MarkerEdgeWidth
	}
	line.MarkEvery = opt.MarkEvery
	line.SetMarkEverySpec(opt.MarkEverySpec)

	// Apply alpha if specified. An alpha outside [0, 1] is ignored rather than
	// clamped, matching the pointer-model behavior this replaced.
	if alpha, ok := opt.Alpha.Get(); ok && alpha >= 0 && alpha <= 1 {
		line.Col.A = alpha
		if !markerFaceColorSet {
			line.MarkerFaceColor.A = alpha
		}
		if !markerEdgeColorSet {
			line.MarkerEdgeColor.A = alpha
		}
		if line.MarkerFaceSpec.Mode == MarkerColorExplicit {
			line.MarkerFaceSpec.Color.A = alpha
		}
		if line.MarkerEdgeSpec.Mode == MarkerColorExplicit {
			line.MarkerEdgeSpec.Color.A = alpha
		}
		if line.MarkerFaceAlt.Mode == MarkerColorExplicit {
			line.MarkerFaceAlt.Color.A = alpha
		}
	}

	a.Add(line)
	a.autoScaleIfEnabled(defaultAutoScaleMargin)
	return line
}

func markerFillStyleFromRC(fill style.MarkerFillStyle) MarkerFillStyle {
	switch fill {
	case style.MarkerFillLeft:
		return MarkerFillLeft
	case style.MarkerFillRight:
		return MarkerFillRight
	case style.MarkerFillBottom:
		return MarkerFillBottom
	case style.MarkerFillTop:
		return MarkerFillTop
	case style.MarkerFillNone:
		return MarkerFillNone
	default:
		return MarkerFillFull
	}
}

func markerColorSpecFromRC(spec style.MarkerColorRC, rc *style.RC) MarkerColorSpec {
	switch spec.Mode {
	case style.MarkerColorNone:
		return NoMarkerColor()
	case style.MarkerColorExplicit:
		if spec.Raw != "" {
			if resolved, err := mplcolor.ToRGBA(spec.Raw, mplcolor.WithColorCycle(rc.Palette()), mplcolor.WithBareHex()); err == nil {
				return ExplicitMarkerColor(resolved)
			}
		}
		return ExplicitMarkerColor(spec.Color)
	default:
		return AutoMarkerColor()
	}
}

// applyLineRCDefaults seeds Line2D constructor defaults for artists built by
// helpers other than Axes.Plot. Explicit common cap/join fields and marker
// colors/specs remain dominant at draw time.
func applyLineRCDefaults(line *Line2D, rc *style.RC) {
	if line == nil || rc == nil {
		return
	}
	rcLines := rc.Lines
	if !line.RCStrokeStylesSet {
		line.DashCap = rcLines.DashCap
		line.DashJoin = rcLines.DashJoin
		line.SolidCap = rcLines.SolidCap
		line.SolidJoin = rcLines.SolidJoin
		line.RCStrokeStylesSet = true
	}
	if !line.AntialiasedSet {
		line.Antialiased = rcLines.Antialiased
		line.AntialiasedSet = true
	}
	if !line.hasMarkers() {
		return
	}
	if line.MarkerSize <= 0 && rcLines.MarkerSize > 0 {
		line.MarkerSize = rcLines.MarkerSize
	}
	if line.MarkerEdgeWidth <= 0 && rcLines.MarkerEdgeWidth > 0 {
		line.MarkerEdgeWidth = rcLines.MarkerEdgeWidth
	}
	if line.MarkerStyle.Tuple == nil && line.MarkerStyle.MathText == "" &&
		len(line.MarkerStyle.Path.C) == 0 && line.MarkerStyle.Type == 0 &&
		line.MarkerStyle.FillStyle == 0 {
		line.MarkerStyle = NewMarkerStyle(line.Marker)
		line.MarkerStyle.FillStyle = markerFillStyleFromRC(rcLines.MarkerFillStyle)
	}
	if line.MarkerFaceSpec.Mode == MarkerColorDefault && line.MarkerFaceColor == (render.Color{}) {
		line.MarkerFaceSpec = markerColorSpecFromRC(rcLines.MarkerFaceColor, rc)
	}
	if line.MarkerEdgeSpec.Mode == MarkerColorDefault && line.MarkerEdgeColor == (render.Color{}) {
		line.MarkerEdgeSpec = markerColorSpecFromRC(rcLines.MarkerEdgeColor, rc)
	}
}

// SemilogX is a convenience wrapper for creating a line plot on a logarithmic
// x-axis.
//
//nolint:gocritic // PlotOptions is an immutable snapshot retained by the line artist.
func (a *Axes) SemilogX(x, y []float64, opt PlotOptions) *Line2D {
	line := a.plot(x, y, opt)
	if line == nil {
		return nil
	}
	setLogScaleFromData(a, x, true)
	return line
}

// SemilogY is a convenience wrapper for creating a line plot on a logarithmic
// y-axis.
//
//nolint:gocritic // PlotOptions is an immutable snapshot retained by the line artist.
func (a *Axes) SemilogY(x, y []float64, opt PlotOptions) *Line2D {
	line := a.plot(x, y, opt)
	if line == nil {
		return nil
	}
	setLogScaleFromData(a, y, false)
	return line
}

// LogLog is a convenience wrapper for creating a line plot on logarithmic x/y
// axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes) LogLog(x, y []float64, opt PlotOptions) *Line2D {
	line := a.plot(x, y, opt)
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
	Color        optional.Value[render.Color] // if nil, uses automatic color cycling
	Colors       []render.Color               // per-point marker face colors
	ScalarValues []float64                    // per-point scalar values mapped through Colormap/Norm
	Colormap     string                       // colormap name for ScalarValues
	Norm         ScalarNormalizer             // normalizer for ScalarValues
	VMin         optional.Value[float64]      // lower color limit for ScalarValues
	VMax         optional.Value[float64]      // upper color limit for ScalarValues
	Size         optional.Value[float64]      // marker area in points^2
	Sizes        []float64                    // per-point marker areas in points^2
	Marker       optional.Value[MarkerType]   // marker type
	MarkerStyle  optional.Value[MarkerStyle]  // marker style; overrides Marker when non-nil
	MarkerPath   optional.Value[geom.Path]    // custom marker path (overrides Marker when non-nil)
	EdgeColor    optional.Value[render.Color] // edge color
	EdgeColors   []render.Color               // per-point marker edge colors
	EdgeWidth    optional.Value[float64]      // edge width
	Alpha        optional.Value[float64]      // alpha transparency
	Label        string                       // series label for legend
	AxLimClip    bool                         // 3D scatter: hide points outside explicit axes limits
	// PlotNonfinite mirrors Matplotlib's plotnonfinite kwarg. When false
	// (default), points with a non-finite x/y/size/scalar/color are masked out.
	// When true, points with non-finite color/scalar values are kept and ride
	// the colormap's "bad" color; only non-finite positions/sizes are dropped.
	PlotNonfinite bool
}

// Scatter converts x and y through the axes units machinery and creates a
// scatter plot with automatic shape/fill color cycling if no color is
// specified.
//
// Rejected input leaves the axes, its unit configuration, and its property
// cycle unchanged.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes) Scatter(xVals, yVals any, opt ScatterOptions) (*Scatter2D, error) {
	if a == nil {
		return nil, fmt.Errorf("scatter axes cannot be nil")
	}

	tx := a.beginUnitConversion()
	x, err := a.convertValues(xVals, true)
	if err != nil {
		tx.rollback()
		return nil, fmt.Errorf("scatter x values: %w", err)
	}
	y, err := a.convertValues(yVals, false)
	if err != nil {
		tx.rollback()
		return nil, fmt.Errorf("scatter y values: %w", err)
	}
	if err := validateScatterInput(x, y, opt); err != nil {
		tx.rollback()
		return nil, err
	}

	scatter := a.scatter(x, y, opt)
	tx.commit()
	return scatter, nil
}

//nolint:gocritic // ScatterOptions is read-only here; a copy keeps validation free of caller side effects.
func validateScatterInput(x, y []float64, opt ScatterOptions) error {
	if len(x) == 0 || len(y) == 0 {
		return fmt.Errorf("scatter x and y values cannot be empty")
	}
	if len(x) != len(y) {
		return fmt.Errorf("scatter x and y must have the same length (got %d and %d)", len(x), len(y))
	}

	n := len(x)
	optionLengths := []struct {
		name   string
		length int
	}{
		{name: "Sizes", length: len(opt.Sizes)},
		{name: "Colors", length: len(opt.Colors)},
		{name: "ScalarValues", length: len(opt.ScalarValues)},
		{name: "EdgeColors", length: len(opt.EdgeColors)},
	}
	for _, option := range optionLengths {
		if option.length > 0 && !validScatterOptionLength(option.length, n) {
			return fmt.Errorf("scatter %s must have length 1 or %d (got %d)", option.name, n, option.length)
		}
	}
	if len(opt.ScalarValues) > 0 {
		if _, err := ResolveScalarMapValues(opt.ScalarValues, ScalarMapConfig{
			Colormap: opt.Colormap,
			Norm:     opt.Norm,
			VMin:     opt.VMin,
			VMax:     opt.VMax,
		}); err != nil {
			return fmt.Errorf("scatter scalar values: %w", err)
		}
	}
	return nil
}

// scatter draws already validated and unit-converted points. Like plot, it
// takes a single options value rather than a variadic tail.
//
//nolint:gocritic // ScatterOptions is an immutable snapshot retained by the scatter artist.
func (a *Axes) scatter(x, y []float64, opt ScatterOptions) *Scatter2D {
	// Create points
	n := len(x)
	points := make([]geom.Pt, n)
	for i := 0; i < n; i++ {
		points[i] = geom.Pt{X: x[i], Y: y[i]}
	}

	// Get color (automatic shape/fill cycling if not specified)
	color := a.NextPatchColor()
	if v, ok := opt.Color.Get(); ok {
		color = v
	}

	// Get size. Matplotlib's scatter "s" parameter is marker area in points^2;
	// the default is lines.markersize^2 with lines.markersize = 6 pt.
	rcScatter := a.resolvedRC()
	size := 36.0
	if ms := rcScatter.Lines.MarkerSize; ms > 0 && ms != 6 {
		size = ms * ms
	}
	if v, ok := opt.Size.Get(); ok {
		size = v
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

	// Get marker type: explicit option, else a non-default scatter.marker rc
	// value, else matplotlib's circle default.
	marker := MarkerCircle
	if rcMarker := rcScatter.Scatter.Marker; rcMarker != "" && rcMarker != "o" {
		if m, ok := MarkerTypeFromString(rcMarker); ok && m != MarkerNone {
			marker = m
		}
	}
	if v, ok := opt.Marker.Get(); ok {
		marker = v
	}

	// Matplotlib defaults scatter marker edges to "face" with linewidth 1;
	// the scatter.edgecolors rcParam overrides that default ("none" or a
	// color), and an explicit EdgeColor option overrides both.
	edgeColor := color
	edgeFollowsFace := true
	if ec := rcScatter.Scatter.EdgeColors; ec != "" && !strings.EqualFold(ec, "face") {
		edgeFollowsFace = false
		if strings.EqualFold(ec, "none") {
			edgeColor = render.Color{}
		} else if parsed, err := mplcolor.ToRGBA(ec); err == nil {
			edgeColor = parsed
		}
	}
	if v, ok := opt.EdgeColor.Get(); ok {
		edgeColor = v
	}
	var colors []render.Color
	if len(opt.Colors) > 0 {
		if !validScatterOptionLength(len(opt.Colors), n) {
			return nil
		}
		if len(opt.Colors) == 1 {
			color = opt.Colors[0]
			if !opt.EdgeColor.IsSet() && edgeFollowsFace {
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
	if v, ok := opt.EdgeWidth.Get(); ok {
		edgeWidth = v
	}

	// Get alpha
	alpha := 1.0
	if v, ok := opt.Alpha.Get(); ok && v >= 0 && opt.Alpha.OrZero() <= 1 {
		alpha = opt.Alpha.OrZero()
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
		scatter.EdgeColorsFace = !opt.EdgeColor.IsSet() && len(opt.EdgeColors) == 0
	}
	if v, ok := opt.MarkerStyle.Get(); ok {
		scatter.MarkerStyle = v
	}
	if v, ok := opt.MarkerPath.Get(); ok {
		scatter.MarkerPath = v
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
	Color       optional.Value[render.Color]   // if nil, uses automatic color cycling
	Colors      []render.Color                 // per-bar fill colors
	Width       optional.Value[float64]        // bar width
	Widths      []float64                      // per-bar widths
	EdgeColor   optional.Value[render.Color]   // edge color
	EdgeColors  []render.Color                 // per-bar edge colors
	EdgeWidth   optional.Value[float64]        // edge width
	Alpha       optional.Value[float64]        // alpha transparency
	Antialiased optional.Value[bool]           // nil uses patch.antialiased
	Baseline    optional.Value[float64]        // baseline value
	Baselines   []float64                      // per-bar baseline/left values
	Orientation optional.Value[BarOrientation] // vertical or horizontal
	Align       optional.Value[BarAlign]       // center or edge alignment
	Label       string                         // series label for legend

	// Error bars (matplotlib bar(yerr=/xerr=)). When any error data is present,
	// bars draw error bars anchored at the bar top (vertical) or end
	// (horizontal), matching matplotlib's placement.
	XErr     []float64                       // symmetric x errors (per-bar or scalar broadcast)
	YErr     []float64                       // symmetric y errors (per-bar or scalar broadcast)
	ECol     optional.Value[render.Color]    // error-bar color; nil = matplotlib default black ('k')
	CapSize  optional.Value[float64]         // cap size in points; nil = errorbar.capsize rc value
	CapThick optional.Value[float64]         // cap line thickness in points; nil = 1pt default
	ErrorKw  optional.Value[ErrorBarOptions] // passthrough for asymmetric errors, errorevery, etc.
}

// Bar converts positions and heights through the axes units machinery and
// creates a bar plot with automatic color cycling if no color is specified.
//
// Rejected input leaves the axes, its unit configuration, and its property
// cycle unchanged.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes) Bar(posVals, heightVals any, opt BarOptions) (*Bar2D, error) {
	return a.barWithOrientation(posVals, heightVals, nil, opt)
}

//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes) barWithOrientation(
	posVals, heightVals any,
	forcedOrientation *BarOrientation,
	opt BarOptions,
) (*Bar2D, error) {
	if a == nil {
		return nil, fmt.Errorf("bar axes cannot be nil")
	}
	if forcedOrientation != nil {
		orientation := *forcedOrientation
		opt.Orientation = optional.Of(orientation)
	}
	orientation := BarVertical
	if v, ok := opt.Orientation.Get(); ok {
		orientation = v
	}
	if orientation != BarVertical && orientation != BarHorizontal {
		return nil, fmt.Errorf("bar orientation must be BarVertical or BarHorizontal (got %d)", orientation)
	}
	if v, ok := opt.Align.Get(); ok && v != BarAlignCenter && opt.Align.OrZero() != BarAlignEdge {
		return nil, fmt.Errorf("bar alignment must be BarAlignCenter or BarAlignEdge (got %d)", opt.Align.OrZero())
	}

	tx := a.beginUnitConversion()
	categoryIsX := orientation == BarVertical
	positions, err := a.convertValues(posVals, categoryIsX)
	if err != nil {
		tx.rollback()
		return nil, fmt.Errorf("bar positions: %w", err)
	}
	heights, err := a.convertValues(heightVals, !categoryIsX)
	if err != nil {
		tx.rollback()
		return nil, fmt.Errorf("bar heights: %w", err)
	}
	if err := validateBarInput(positions, heights, &opt); err != nil {
		tx.rollback()
		return nil, err
	}

	a.applyCategoricalBarValueLocator(categoryIsX)
	bar := a.bar(positions, heights, opt)
	tx.commit()
	return bar, nil
}

func validateBarInput(positions, heights []float64, opt *BarOptions) error {
	if len(positions) == 0 || len(heights) == 0 {
		return fmt.Errorf("bar positions and heights cannot be empty")
	}
	if len(positions) != len(heights) {
		return fmt.Errorf(
			"bar positions and heights must have the same length (got %d and %d)",
			len(positions),
			len(heights),
		)
	}

	n := len(positions)
	optionLengths := []struct {
		name   string
		length int
	}{
		{name: "Widths", length: len(opt.Widths)},
		{name: "Colors", length: len(opt.Colors)},
		{name: "EdgeColors", length: len(opt.EdgeColors)},
		{name: "Baselines", length: len(opt.Baselines)},
	}
	for _, option := range optionLengths {
		if option.length > 0 && !validBarOptionLength(option.length, n) {
			return fmt.Errorf("bar %s must have length 1 or %d (got %d)", option.name, n, option.length)
		}
	}
	if !validErrorValues(opt.XErr, n) || !validErrorValues(opt.YErr, n) {
		return fmt.Errorf("bar XErr and YErr must each be empty or length 1 or %d, with finite non-negative values", n)
	}
	kw, ok := opt.ErrorKw.Get()
	if !ok {
		return nil
	}
	if !validErrorValues(kw.XErrLower, n) || !validErrorValues(kw.XErrUpper, n) ||
		!validErrorValues(kw.YErrLower, n) || !validErrorValues(kw.YErrUpper, n) {
		return fmt.Errorf("bar ErrorKw error arrays must each be empty or length 1 or %d, with finite non-negative values", n)
	}
	if !validBoolValues(kw.LoLimits, n) || !validBoolValues(kw.UpLimits, n) ||
		!validBoolValues(kw.XLoLimits, n) || !validBoolValues(kw.XUpLimits, n) {
		return fmt.Errorf("bar ErrorKw limit arrays must each be empty or length 1 or %d", n)
	}
	if kw.ErrorEvery < 0 || (kw.ErrorEvery == 0 && kw.ErrorEveryStart != 0) || kw.ErrorEveryStart < 0 {
		return fmt.Errorf(
			"bar ErrorKw errorevery is invalid (every=%d, start=%d)",
			kw.ErrorEvery,
			kw.ErrorEveryStart,
		)
	}
	return nil
}

// bar draws already validated and unit-converted bars. Like plot and scatter,
// it takes a single options value rather than a variadic tail.
//
//nolint:gocritic // BarOptions is an immutable snapshot retained by the bar artist.
func (a *Axes) bar(x, heights []float64, opt BarOptions) *Bar2D {
	if len(x) == 0 || len(heights) == 0 {
		return nil
	}

	// Get color (automatic cycling if not specified)
	color := a.NextColor()
	if v, ok := opt.Color.Get(); ok {
		color = v
	}

	// Get width
	width := 0.8
	if v, ok := opt.Width.Get(); ok {
		width = v
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
	rcPatch := a.resolvedRC().Patch
	edgeColor := render.Color{}
	if rcPatch.ForceEdgeColor {
		edgeColor = rcPatch.EdgeColor
	}
	if v, ok := opt.EdgeColor.Get(); ok {
		edgeColor = v
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

	edgeWidth := rcPatch.LineWidth
	if v, ok := opt.EdgeWidth.Get(); ok {
		edgeWidth = v
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
	if v, ok := opt.Baseline.Get(); ok {
		baseline = v
	}

	// Get orientation
	orientation := BarVertical
	if v, ok := opt.Orientation.Get(); ok {
		orientation = v
	}
	align := BarAlignCenter
	if v, ok := opt.Align.Get(); ok {
		align = v
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
		Antialias:   patchAntialiasMode(&rcPatch, opt.Antialiased),
		Baseline:    baseline,
		Orientation: orientation,
		Label:       opt.Label,
	}

	a.Add(bar)
	// Error bars, matplotlib bar(yerr=/xerr=): anchor at the bar top (vertical)
	// or end (horizontal), drawn with fmt="none" and ecolor default black. Added
	// before autoscale so the error extents widen the data limits.
	bar.addErrorBars(a, &opt)
	a.autoScaleIfEnabled(defaultAutoScaleMargin)
	return bar
}

// barHasErrorData reports whether the bar options carry any error information,
// either the direct symmetric fields or asymmetric arrays via ErrorKw.
func barHasErrorData(opt *BarOptions) bool {
	if len(opt.XErr) > 0 || len(opt.YErr) > 0 {
		return true
	}
	if kw, ok := opt.ErrorKw.Get(); ok {
		if len(kw.XErrLower) > 0 || len(kw.XErrUpper) > 0 ||
			len(kw.YErrLower) > 0 || len(kw.YErrUpper) > 0 {
			return true
		}
	}
	return false
}

// addErrorBars constructs and attaches matplotlib-faithful error bars for a bar
// chart. The anchor points are the bar tops (vertical) or ends (horizontal); the
// heavy lifting (array validation, errorbar.capsize rc resolution, fmt="none"
// data-line suppression) is reused from Axes.ErrorBar.
func (b *Bar2D) addErrorBars(a *Axes, opt *BarOptions) {
	if b == nil || !barHasErrorData(opt) {
		return
	}
	n := min(len(b.X), len(b.Heights))
	ex := make([]float64, n)
	ey := make([]float64, n)
	for i := range n {
		end := b.baselineAt(i) + b.Heights[i]
		if b.Orientation == BarHorizontal {
			ex[i], ey[i] = end, b.X[i]
		} else {
			ex[i], ey[i] = b.X[i], end
		}
	}

	// Start from the ErrorKw passthrough (asymmetric errors, errorevery, …); the
	// bar-level fields then apply with fmt="none" data-line suppression.
	var eb ErrorBarOptions
	if v, ok := opt.ErrorKw.Get(); ok {
		eb = v
	}
	eb.NoDataLine = true
	eb.Label = ""
	// ecolor precedence mirrors matplotlib's error_kw.setdefault: an explicit
	// ECol wins, else an ErrorKw.Color passthrough is preserved, else the default
	// black ('k'). Never clobber a caller-supplied ErrorKw color.
	switch {
	case opt.ECol.IsSet():
		ecolor := opt.ECol.OrZero()
		eb.Color = optional.Of(ecolor)
	case eb.Color.IsSet():
		// keep the ErrorKw passthrough color
	default:
		eb.Color = optional.Of(render.Color{R: 0, G: 0, B: 0, A: 1})
	}
	if opt.CapSize.IsSet() {
		eb.CapSize = opt.CapSize
	}
	if opt.CapThick.IsSet() {
		eb.CapThick = opt.CapThick
	}

	// Bar error bars are matplotlib's hidden fmt="none" helper and must not
	// consume the axes color cycle. Axes.ErrorBar advances the cycle
	// unconditionally, so snapshot and restore the index around the call.
	savedCycleIndex := 0
	if a.ColorCycle != nil {
		savedCycleIndex = a.ColorCycle.Index()
	}
	// validateBarInput already rejected empty bars and every malformed error or
	// limit array, so this call cannot fail; the container simply stays nil if
	// it ever does.
	b.errorbar, _ = a.ErrorBar(ex, ey, opt.XErr, opt.YErr, eb)
	if a.ColorCycle != nil {
		a.ColorCycle.SetIndex(savedCycleIndex)
	}
}

func validBarOptionLength(length, n int) bool {
	return length == 1 || length == n
}

// bakeExplicitAlpha returns c with its alpha replaced by *alpha when alpha is an
// explicit value in [0,1]. An explicit 0 yields a fully transparent color; a nil
// alpha leaves the color (and its own alpha channel) untouched. Baking the
// override into the resolved color at construction lets artists honor an
// explicit alpha=0 — which a plain float64 "0 means unset" field cannot express.
func bakeExplicitAlpha(c render.Color, alpha optional.Value[float64]) render.Color {
	if v, ok := alpha.Get(); ok && v >= 0 && v <= 1 {
		c.A = v
	}
	return c
}

// BarH converts positions and widths through the axes units machinery and
// creates a horizontal bar chart.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes) BarH(yVals, widthVals any, opt BarOptions) (*Bar2D, error) {
	orientation := BarHorizontal
	return a.barWithOrientation(yVals, widthVals, &orientation, opt)
}

// FillBetween converts x, y1, and y2 through the axes units machinery and fills
// the area between the two dependent curves.
//
// Rejected input leaves the axes, its unit configuration, and its property
// cycle unchanged.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes) FillBetween(xVals, y1Vals, y2Vals any, opt FillOptions) (*Fill2D, error) {
	if a == nil {
		return nil, fmt.Errorf("fill between axes cannot be nil")
	}

	// Convert every input before touching the artist list or the property
	// cycle: a rejected y2 must not leave x-axis units configured.
	tx := a.beginUnitConversion()
	x, err := a.convertValues(xVals, true)
	if err != nil {
		tx.rollback()
		return nil, fmt.Errorf("fill between x values: %w", err)
	}
	y1, err := a.convertValues(y1Vals, false)
	if err != nil {
		tx.rollback()
		return nil, fmt.Errorf("fill between y1 values: %w", err)
	}
	y2, err := a.convertValues(y2Vals, false)
	if err != nil {
		tx.rollback()
		return nil, fmt.Errorf("fill between y2 values: %w", err)
	}
	if err := fillBetweenNames.validate(x, y1, y2, &opt); err != nil {
		tx.rollback()
		return nil, err
	}

	fill := a.fillBetween(x, y1, y2, &opt)
	tx.commit()
	return fill, nil
}

// fillInputNames labels the three coordinate slices of a fill-between call so
// the shared shape check can report the argument the caller actually passed.
type fillInputNames struct {
	call        string
	independent string
	first       string
	second      string
}

var (
	fillBetweenNames  = fillInputNames{call: "fill between", independent: "x", first: "y1", second: "y2"}
	fillBetweenXNames = fillInputNames{call: "fill between x", independent: "y", first: "x1", second: "x2"}
)

// validate enforces the fill-between shape contract: three non-empty,
// equal-length coordinate slices and a Where mask that is either empty or the
// same length as the independent variable.
func (n fillInputNames) validate(independent, first, second []float64, opt *FillOptions) error {
	if len(independent) == 0 || len(first) == 0 || len(second) == 0 {
		return fmt.Errorf("%s %s, %s, and %s values cannot be empty", n.call, n.independent, n.first, n.second)
	}
	if len(independent) != len(first) || len(independent) != len(second) {
		return fmt.Errorf(
			"%s %s, %s, and %s must have the same length (got %d, %d, and %d)",
			n.call,
			n.independent,
			n.first,
			n.second,
			len(independent),
			len(first),
			len(second),
		)
	}
	if len(opt.Where) > 0 && len(opt.Where) != len(independent) {
		return fmt.Errorf("%s Where must have length %d (got %d)", n.call, len(independent), len(opt.Where))
	}
	return nil
}

// Fill creates an arbitrary closed polygon fill using data-space coordinates.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes) Fill(x, y []float64, opt FillOptions) *PolyCollection {
	if len(x) == 0 || len(y) == 0 {
		return nil
	}
	n := minInt(len(x), len(y))
	if n < 3 {
		return nil
	}

	points := make([]geom.Pt, n)
	for i := 0; i < n; i++ {
		points[i] = geom.Pt{X: x[i], Y: y[i]}
	}

	color := a.NextColor()
	if v, ok := opt.Color.Get(); ok {
		color = v
	}

	rcPatch := a.resolvedRC().Patch
	edgeColor := render.Color{}
	if rcPatch.ForceEdgeColor {
		edgeColor = rcPatch.EdgeColor
	}
	if v, ok := opt.EdgeColor.Get(); ok {
		edgeColor = v
	}

	edgeWidth := rcPatch.LineWidth
	if v, ok := opt.EdgeWidth.Get(); ok {
		edgeWidth = v
	}

	color = bakeExplicitAlpha(color, opt.Alpha)
	edgeColor = bakeExplicitAlpha(edgeColor, opt.Alpha)

	fill := &PolyCollection{
		PatchCollection: PatchCollection{
			Collection: Collection{
				Label:     opt.Label,
				Alpha:     1,
				Antialias: patchAntialiasMode(&rcPatch, opt.Antialiased),
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
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes) FillToBaseline(x, y []float64, opt FillOptions) *Fill2D {
	return a.FillToBaselinePlot(x, y, opt)
}

// FillBetweenX creates a horizontal fill between x-curves across y values.
//
// Rejected input leaves the axes and its property cycle unchanged.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes) FillBetweenX(y, x1, x2 []float64, opt FillOptions) (*Fill2D, error) {
	if a == nil {
		return nil, fmt.Errorf("fill between x axes cannot be nil")
	}
	if err := fillBetweenXNames.validate(y, x1, x2, &opt); err != nil {
		return nil, err
	}

	color := a.NextColor()
	if v, ok := opt.Color.Get(); ok {
		color = v
	}

	rcPatch := a.resolvedRC().Patch
	edgeColor := render.Color{}
	if rcPatch.ForceEdgeColor {
		edgeColor = rcPatch.EdgeColor
	}
	if v, ok := opt.EdgeColor.Get(); ok {
		edgeColor = v
	}

	edgeWidth := rcPatch.LineWidth
	if v, ok := opt.EdgeWidth.Get(); ok {
		edgeWidth = v
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
		Antialias:   patchAntialiasMode(&rcPatch, opt.Antialiased),
		Label:       opt.Label,
	}

	a.Add(fill)
	a.autoScaleIfEnabled(defaultAutoScaleMargin)
	return fill, nil
}

// FillOptions holds optional parameters for fill plots.
type FillOptions struct {
	Color       optional.Value[render.Color] // if nil, uses automatic color cycling
	EdgeColor   optional.Value[render.Color] // edge color
	EdgeWidth   optional.Value[float64]      // edge width
	Alpha       optional.Value[float64]      // alpha transparency
	Antialiased optional.Value[bool]         // nil uses patch.antialiased
	Baseline    optional.Value[float64]      // baseline value
	Where       []bool                       // fill only contiguous regions where adjacent points are true
	Interpolate bool                         // interpolate region boundaries at curve crossings
	Step        FillStep                     // optional step mode
	Label       string                       // series label for legend
}

func patchAntialiasMode(rc *style.PatchRC, explicit optional.Value[bool]) render.AntialiasMode {
	enabled := explicit.Or(rc == nil || rc.Antialiased)
	if enabled {
		return render.AntialiasOn
	}
	return render.AntialiasOff
}

// FillBetweenPlot creates a fill between two curves with automatic color
// cycling. It is the numeric-only entry point; FillBetween additionally
// converts unit-carrying values.
//
// Rejected input leaves the axes and its property cycle unchanged.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes) FillBetweenPlot(x, y1, y2 []float64, opt FillOptions) (*Fill2D, error) {
	if a == nil {
		return nil, fmt.Errorf("fill between axes cannot be nil")
	}
	if err := fillBetweenNames.validate(x, y1, y2, &opt); err != nil {
		return nil, err
	}

	return a.fillBetween(x, y1, y2, &opt), nil
}

func (a *Axes) fillBetween(x, y1, y2 []float64, opt *FillOptions) *Fill2D {
	// Get color (automatic cycling if not specified)
	color := a.NextColor()
	if v, ok := opt.Color.Get(); ok {
		color = v
	}

	// Get edge properties
	rcPatch := a.resolvedRC().Patch
	edgeColor := render.Color{}
	if rcPatch.ForceEdgeColor {
		edgeColor = rcPatch.EdgeColor
	}
	if v, ok := opt.EdgeColor.Get(); ok {
		edgeColor = v
	}

	edgeWidth := rcPatch.LineWidth
	if v, ok := opt.EdgeWidth.Get(); ok {
		edgeWidth = v
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
		Antialias:   patchAntialiasMode(&rcPatch, opt.Antialiased),
		Label:       opt.Label,
	}

	a.Add(fill)
	a.autoScaleIfEnabled(defaultAutoScaleMargin)
	return fill
}

// HistOptions holds optional parameters for histogram plots.
type HistOptions struct {
	Bins              int                       // number of bins (0 = hist.bins rc default, 10)
	BinEdges          []float64                 // explicit bin edges (overrides Bins)
	Range             optional.Value[HistRange] // explicit histogram range; ignored when BinEdges is set
	Weights           []float64                 // per-sample weights, same length as data when provided
	BinStrat          BinStrategy               // automatic binning strategy
	Norm              HistNorm                  // normalization mode
	Cumulative        bool                      // accumulate bin heights from left to right
	ReverseCumulative bool                      // accumulate from right to left, matching cumulative < 0
	Log               bool                      // set the count (y) axis to log scale, matching matplotlib hist(log=True)
	HistType          HistType                  // bar, step, or filled step presentation
	Baselines         []float64                 // optional per-bin baselines for stacked histograms
	Color             optional.Value[render.Color]
	EdgeColor         optional.Value[render.Color]
	EdgeWidth         optional.Value[float64]
	Alpha             optional.Value[float64]
	Antialiased       optional.Value[bool]
	Label             string
}

// Hist creates a histogram from raw data with automatic color cycling.
//
// Rejected input leaves the axes and its property cycle unchanged.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes) Hist(data []float64, opt HistOptions) (*Hist2D, error) {
	if a == nil {
		return nil, fmt.Errorf("hist axes cannot be nil")
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("hist data cannot be empty")
	}
	if len(opt.Weights) > 0 && len(opt.Weights) != len(data) {
		return nil, fmt.Errorf("hist Weights must have length %d (got %d)", len(data), len(opt.Weights))
	}

	// With no explicit bins, edges, or strategy, honor the hist.bins rcParam
	// (matplotlib default: 10 fixed bins; 'auto' selects numpy's estimator).
	bins := opt.Bins
	binStrat := opt.BinStrat
	if bins <= 0 && len(opt.BinEdges) < 2 && binStrat == BinStrategyDefault {
		switch rcBins := a.resolvedRC().HistBins; {
		case rcBins == style.HistBinsAuto:
			binStrat = BinStrategyAuto
		case rcBins > 0:
			bins = rcBins
		}
	}

	color := a.NextColor()
	if v, ok := opt.Color.Get(); ok {
		color = v
	}

	rcPatch := a.resolvedRC().Patch
	edgeColor := render.Color{}
	if rcPatch.ForceEdgeColor {
		edgeColor = rcPatch.EdgeColor
	}
	if v, ok := opt.EdgeColor.Get(); ok {
		edgeColor = v
	} else if opt.HistType != HistTypeBar {
		edgeColor = color
	}

	edgeWidth := rcPatch.LineWidth
	if v, ok := opt.EdgeWidth.Get(); ok {
		edgeWidth = v
	} else if opt.HistType != HistTypeBar {
		edgeWidth = rcPatch.LineWidth
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
		Bins:              bins,
		BinEdges:          opt.BinEdges,
		Range:             opt.Range.Ptr(),
		BinStrat:          binStrat,
		Norm:              opt.Norm,
		Cumulative:        opt.Cumulative,
		ReverseCumulative: opt.ReverseCumulative,
		HistType:          opt.HistType,
		Baselines:         append([]float64(nil), opt.Baselines...),
		Color:             color,
		EdgeColor:         edgeColor,
		EdgeWidth:         edgeWidth,
		Antialias:         patchAntialiasMode(&rcPatch, opt.Antialiased),
		Label:             opt.Label,
	}

	a.Add(hist)

	// matplotlib's hist(log=True) sets the count axis to a log scale with
	// nonpositive='clip', so the zero bar baseline clips to the axis floor
	// instead of masking to NaN. Histograms here are vertical-only (no
	// orientation field), so this always targets the y axis.
	if opt.Log {
		_ = a.SetYScale("log", transform.WithScaleNonPositive(transform.NonPositiveClip))
	}

	return hist, nil
}

// ErrorBarOptions holds optional parameters for error bar plots.
type ErrorBarOptions struct {
	Color           optional.Value[render.Color] // if nil, uses automatic color cycling
	LineWidth       optional.Value[float64]      // error bar line width (px); nil uses Matplotlib's default
	CapSize         optional.Value[float64]      // Matplotlib capsize in points
	CapThick        optional.Value[float64]      // Matplotlib capthick in points (cap line thickness); nil uses the 1pt default
	Marker          optional.Value[MarkerType]   // optional data marker equivalent to Matplotlib fmt markers
	MarkerSize      optional.Value[float64]      // marker size in points
	Alpha           optional.Value[float64]      // alpha transparency
	Label           string                       // series label for legend
	NoDataLine      bool                         // true matches Matplotlib fmt="none"
	ErrorEvery      int                          // draw error bars every N points, default 1
	ErrorEveryStart int                          // starting point for ErrorEvery, matching errorevery=(start,N)

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
//
// Rejected input leaves the axes and its property cycle unchanged.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes) ErrorBar(x, y, xErr, yErr []float64, opt ErrorBarOptions) (*ErrorBar, error) {
	if a == nil {
		return nil, fmt.Errorf("errorbar axes cannot be nil")
	}

	// Validate before the property cycle advances so a rejected call leaves the
	// axes untouched.
	n := minInt(len(x), len(y))
	if err := validateErrorBarInput(n, xErr, yErr, &opt); err != nil {
		return nil, err
	}

	color := a.NextColor()
	if v, ok := opt.Color.Get(); ok {
		color = v
	}

	lineWidth := 0.0
	if v, ok := opt.LineWidth.Get(); ok {
		lineWidth = v
	}

	// Cap size: explicit option, else the errorbar.capsize rc value
	// (matplotlib default 0 — no caps).
	capSizePx := 0.0
	switch {
	case opt.CapSize.IsSet():
		capSizePx = pointsToPixels(a.resolvedRC(), 2*opt.CapSize.OrZero())
	default:
		if rcCap := a.resolvedRC().Errorbar.CapSize; rcCap > 0 {
			capSizePx = pointsToPixels(a.resolvedRC(), 2*rcCap)
		}
	}

	capThick := 0.0 // points; ErrorBar converts at its cap Paint sink
	if v, ok := opt.CapThick.Get(); ok && v > 0 {
		capThick = opt.CapThick.OrZero()
	}

	// Bake an explicit alpha (including 0) into the stroke color; the Alpha
	// field stays unset so Draw's alpha multiplier is the identity.
	color = bakeExplicitAlpha(color, opt.Alpha)

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
		CapThick:        capThick,
		Label:           opt.Label,
		NoDataLine:      opt.NoDataLine,
		ErrorEvery:      errorEvery,
		ErrorEveryStart: opt.ErrorEveryStart,
	}
	if opt.Marker.IsSet() {
		bar.Marker = opt.Marker.OrZero()
		bar.MarkerSet = true
	}
	if v, ok := opt.MarkerSize.Get(); ok {
		bar.MarkerSize = v
	}
	a.Add(bar)
	return bar, nil
}

// validateErrorBarInput rejects empty coordinates, error/limit arrays that are
// neither scalar nor per-point, and out-of-range errorevery settings.
func validateErrorBarInput(n int, xErr, yErr []float64, opt *ErrorBarOptions) error {
	if n == 0 {
		return fmt.Errorf("errorbar x and y values cannot be empty")
	}
	if !validErrorValues(xErr, n) || !validErrorValues(yErr, n) ||
		!validErrorValues(opt.XErrLower, n) || !validErrorValues(opt.XErrUpper, n) ||
		!validErrorValues(opt.YErrLower, n) || !validErrorValues(opt.YErrUpper, n) ||
		!validBoolValues(opt.LoLimits, n) || !validBoolValues(opt.UpLimits, n) ||
		!validBoolValues(opt.XLoLimits, n) || !validBoolValues(opt.XUpLimits, n) {
		return fmt.Errorf(
			"errorbar error and limit arrays must each be empty, scalar, or length %d with finite non-negative errors",
			n,
		)
	}
	if opt.ErrorEvery < 0 || (opt.ErrorEvery == 0 && opt.ErrorEveryStart != 0) || opt.ErrorEveryStart < 0 {
		return fmt.Errorf(
			"errorbar has an invalid errorevery (every=%d, start=%d)",
			opt.ErrorEvery,
			opt.ErrorEveryStart,
		)
	}
	return nil
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
	Position     optional.Value[float64]      // x position of the box center
	Width        optional.Value[float64]      // box width in data units
	Color        optional.Value[render.Color] // box fill color
	EdgeColor    optional.Value[render.Color] // box outline color
	MedianColor  optional.Value[render.Color] // median line color
	WhiskerColor optional.Value[render.Color] // whisker and cap color
	CapColor     optional.Value[render.Color] // whisker cap color
	FlierColor   optional.Value[render.Color] // outlier marker color
	EdgeWidth    optional.Value[float64]      // box outline width in pixels
	WhiskerWidth optional.Value[float64]      // whisker line width in pixels
	MedianWidth  optional.Value[float64]      // median line width in pixels
	CapWidth     optional.Value[float64]      // cap length in data units
	FlierSize    optional.Value[float64]      // outlier marker size in points
	Alpha        optional.Value[float64]      // alpha transparency
	ShowFliers   optional.Value[bool]         // whether to draw outliers
	Label        string                       // series label for legend

	PatchArtist optional.Value[bool]            // fill the box (Matplotlib patch_artist=True); default is unfilled
	Orientation optional.Value[PlotOrientation] // "vertical" (default) or "horizontal"
	ShowBox     optional.Value[bool]            // whether to draw the box (default true)
	ShowCaps    optional.Value[bool]            // whether to draw the whisker caps (default true)
	ShowMeans   optional.Value[bool]            // whether to draw the mean
	MeanLine    optional.Value[bool]            // draw the mean as a line across the box instead of a marker
	MeanColor   optional.Value[render.Color]    // mean line/marker color
	Whis        optional.Value[float64]         // IQR multiplier for whiskers (default 1.5)
	Sym         optional.Value[string]          // Matplotlib flier format string, e.g. "b+"; "" disables fliers

	Notch              optional.Value[bool]       // draw a notched box using the confidence interval
	Bootstrap          int                        // number of bootstrap resamples for the notch CI (0 = analytic)
	ConfidenceInterval optional.Value[[2]float64] // custom median confidence interval for notches
	CustomMedian       optional.Value[float64]    // override the computed median
	WhiskerPercentiles optional.Value[[2]float64] // percentile whisker range, e.g. [5, 95]
	FlierMarker        optional.Value[MarkerType] // marker for outlier points
	FlierEdgeColor     optional.Value[render.Color]
	FlierEdgeWidth     optional.Value[float64]
}

// BoxPlotsOptions holds optional parameters for multi-series box plots.
type BoxPlotsOptions struct {
	Positions    []float64                    // x positions for each box center
	Width        optional.Value[float64]      // box width in data units
	Colors       []render.Color               // box fill colors, one per dataset
	EdgeColor    optional.Value[render.Color] // box outline color
	MedianColor  optional.Value[render.Color] // median line color
	WhiskerColor optional.Value[render.Color] // whisker and cap color
	CapColor     optional.Value[render.Color] // whisker cap color
	FlierColor   optional.Value[render.Color] // outlier marker color
	EdgeWidth    optional.Value[float64]      // box outline width in pixels
	WhiskerWidth optional.Value[float64]      // whisker line width in pixels
	MedianWidth  optional.Value[float64]      // median line width in pixels
	CapWidth     optional.Value[float64]      // cap length in data units
	FlierSize    optional.Value[float64]      // outlier marker size in points
	Alpha        optional.Value[float64]      // alpha transparency
	ShowFliers   optional.Value[bool]         // whether to draw outliers
	ManageTicks  optional.Value[bool]         // whether to place position ticks at box positions
	Labels       []string                     // series labels for legend

	PatchArtist optional.Value[bool]            // fill the boxes (Matplotlib patch_artist=True); default is unfilled
	Orientation optional.Value[PlotOrientation] // "vertical" (default) or "horizontal"
	ShowBox     optional.Value[bool]            // whether to draw the box (default true)
	ShowCaps    optional.Value[bool]            // whether to draw the whisker caps (default true)
	ShowMeans   optional.Value[bool]            // whether to draw the mean
	MeanLine    optional.Value[bool]            // draw the mean as a line across the box instead of a marker
	MeanColor   optional.Value[render.Color]    // mean line/marker color
	Whis        optional.Value[float64]         // IQR multiplier for whiskers (default 1.5)
	Sym         optional.Value[string]          // Matplotlib flier format string, e.g. "b+"; "" disables fliers

	Notch               optional.Value[bool]
	Bootstrap           int
	ConfidenceIntervals [][2]float64
	CustomMedians       []float64
	WhiskerPercentiles  optional.Value[[2]float64]
	FlierMarker         optional.Value[MarkerType]
	FlierEdgeColor      optional.Value[render.Color]
	FlierEdgeWidth      optional.Value[float64]
}

// BoxPlot creates a box plot from raw sample data with automatic color cycling.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes) BoxPlot(data []float64, opt BoxPlotOptions) *BoxPlot2D {
	if len(data) == 0 {
		return nil
	}

	rc := a.resolvedRC()

	position := 1.0
	if v, ok := opt.Position.Get(); ok {
		position = v
	}

	// Matplotlib's bxp default is clip(0.15*ptp(positions), 0.15, 0.5).
	// A direct BoxPlot has exactly one position, so ptp is zero and the
	// resulting default width is 0.15.
	width := 0.15
	if v, ok := opt.Width.Get(); ok {
		width = v
	}

	// Matplotlib's default boxplot is patch_artist=False: an unfilled box that
	// does not consume the color cycle. Only fill (and default the facecolor to
	// white) when patch_artist is requested.
	patchArtist := rc.Boxplot.PatchArtist
	if v, ok := opt.PatchArtist.Get(); ok {
		patchArtist = v
	}
	color := render.Color{}
	if patchArtist {
		color = render.Color{R: 1, G: 1, B: 1, A: 1}
	}
	if v, ok := opt.Color.Get(); ok {
		color = v
	}

	edgeColor := render.Color{R: 0, G: 0, B: 0, A: 1}
	if v, ok := opt.EdgeColor.Get(); ok {
		edgeColor = v
	}

	medianColor := rc.Boxplot.MedianColor
	if v, ok := opt.MedianColor.Get(); ok {
		medianColor = v
	}

	meanColor := rc.Boxplot.MeanColor
	if v, ok := opt.MeanColor.Get(); ok {
		meanColor = v
	}

	whiskerColor := edgeColor
	if v, ok := opt.WhiskerColor.Get(); ok {
		whiskerColor = v
	}

	capColor := whiskerColor
	if v, ok := opt.CapColor.Get(); ok {
		capColor = v
	}

	flierColor := render.Color{}
	if v, ok := opt.FlierColor.Get(); ok {
		flierColor = v
	}

	edgeWidth := rc.Boxplot.BoxLineWidth // points; BoxPlot2D converts at its sinks
	if v, ok := opt.EdgeWidth.Get(); ok {
		edgeWidth = v
	}

	whiskerWidth := rc.Boxplot.WhiskerLineWidth // points
	if v, ok := opt.WhiskerWidth.Get(); ok {
		whiskerWidth = v
	}

	medianWidth := rc.Boxplot.MedianLineWidth // points
	if v, ok := opt.MedianWidth.Get(); ok {
		medianWidth = v
	}

	capWidth := width * 0.5
	if v, ok := opt.CapWidth.Get(); ok {
		capWidth = v
	}

	flierSize := rc.Boxplot.FlierMarkerSize
	if v, ok := opt.FlierSize.Get(); ok {
		flierSize = v
	}

	// Bake an explicit alpha (including 0) into the box fill and edge colors —
	// the only colors Draw applies alpha to — leaving the Alpha field unset so
	// its multiplier is the identity.
	color = bakeExplicitAlpha(color, opt.Alpha)
	edgeColor = bakeExplicitAlpha(edgeColor, opt.Alpha)

	showFliers := rc.Boxplot.ShowFliers
	if v, ok := opt.ShowFliers.Get(); ok {
		showFliers = v
	}
	showBox := rc.Boxplot.ShowBox
	if v, ok := opt.ShowBox.Get(); ok {
		showBox = v
	}
	showCaps := rc.Boxplot.ShowCaps
	if v, ok := opt.ShowCaps.Get(); ok {
		showCaps = v
	}
	showMeans := rc.Boxplot.ShowMeans
	if v, ok := opt.ShowMeans.Get(); ok {
		showMeans = v
	}
	meanLine := rc.Boxplot.MeanLine
	if v, ok := opt.MeanLine.Get(); ok {
		meanLine = v
	}
	var orientation PlotOrientation
	if v, ok := opt.Orientation.Get(); ok {
		orientation = v
	}
	notch := rc.Boxplot.Notch
	if v, ok := opt.Notch.Get(); ok {
		notch = v
	}
	flierMarker := MarkerCircle
	if v, ok := opt.FlierMarker.Get(); ok {
		flierMarker = v
	}
	flierEdgeColor := rc.Boxplot.FlierColor
	if opt.FlierColor.IsSet() {
		flierEdgeColor = flierColor
	}
	if v, ok := opt.FlierEdgeColor.Get(); ok {
		flierEdgeColor = v
	}
	flierEdgeWidth := rc.Boxplot.FlierEdgeWidth // points; BoxPlot2D converts at its sinks
	if v, ok := opt.FlierEdgeWidth.Get(); ok {
		flierEdgeWidth = v
	}

	// Matplotlib's sym shorthand overrides the flier marker/color; sym="" hides
	// fliers entirely. Structured FlierMarker/FlierColor options take precedence.
	if opt.Sym.IsSet() {
		if opt.Sym.OrZero() == "" {
			showFliers = false
		} else {
			marker, symColor, hasMarker, hasColor := parseBoxplotSym(opt.Sym.OrZero())
			if hasMarker && !opt.FlierMarker.IsSet() {
				flierMarker = marker
			}
			if hasColor && !opt.FlierColor.IsSet() {
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
		Whis:               opt.Whis.Ptr(),
		ConfidenceInterval: opt.ConfidenceInterval.Ptr(),
		CustomMedian:       opt.CustomMedian.Ptr(),
		WhiskerPercentiles: opt.WhiskerPercentiles.Ptr(),
		FlierMarker:        flierMarker,
		Label:              opt.Label,
		// Matplotlib draws boxplot artists at Line2D.zorder (2) so they render
		// above default grids (axisbelow z=0.5/1.5); medians/means use 2.1.
		z: 2,
	}

	a.Add(box)
	return box
}

// BoxPlots creates a group of box plots from raw sample datasets.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes) BoxPlots(datasets [][]float64, opt BoxPlotsOptions) []*BoxPlot2D {
	if len(datasets) == 0 {
		return nil
	}

	width := opt.Width
	if !width.IsSet() {
		defaultWidth := matplotlibBoxPlotDefaultWidth(len(datasets), opt.Positions)
		width = optional.Of(defaultWidth)
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
			Position:           optional.Of(position),
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
			boxOpt.ConfidenceInterval = optional.Of(ci)
		}
		if i < len(opt.CustomMedians) && isFinite(opt.CustomMedians[i]) {
			median := opt.CustomMedians[i]
			boxOpt.CustomMedian = optional.Of(median)
		}
		if i < len(opt.Colors) {
			boxOpt.Color = optional.Of(opt.Colors[i])
		}
		if i < len(opt.Labels) {
			boxOpt.Label = opt.Labels[i]
		}

		if box := a.BoxPlot(data, boxOpt); box != nil {
			boxes = append(boxes, box)
		}
	}
	manageTicks := true
	if v, ok := opt.ManageTicks.Get(); ok {
		manageTicks = v
	}
	if manageTicks && len(positions) > 0 {
		// Matplotlib boxplot(..., manage_ticks=True) places the position-axis ticks
		// at the box positions by default — the y axis for horizontal orientation.
		horizontal := opt.Orientation.IsSet() && normalizeViolinOrientation(PlotOrientation(opt.Orientation.OrZero())) == "horizontal"
		if horizontal {
			if a.YAxis != nil {
				a.YAxis.Locator = ticker.FixedLocator{TicksList: positions}
			}
		} else if a.XAxis != nil {
			a.XAxis.Locator = ticker.FixedLocator{TicksList: positions}
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
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes) FillToBaselinePlot(x, y []float64, opt FillOptions) *Fill2D {
	if len(x) == 0 || len(y) == 0 {
		return nil
	}

	// Default options

	// Get color (automatic cycling if not specified)
	color := a.NextColor()
	if v, ok := opt.Color.Get(); ok {
		color = v
	}

	// Get edge properties
	rcPatch := a.resolvedRC().Patch
	edgeColor := render.Color{}
	if rcPatch.ForceEdgeColor {
		edgeColor = rcPatch.EdgeColor
	}
	if v, ok := opt.EdgeColor.Get(); ok {
		edgeColor = v
	}

	edgeWidth := rcPatch.LineWidth
	if v, ok := opt.EdgeWidth.Get(); ok {
		edgeWidth = v
	}

	// When alpha is omitted, preserve the color's own alpha, matching
	// Matplotlib's fill_between behavior; an explicit value (including 0) wins.
	color = bakeExplicitAlpha(color, opt.Alpha)
	edgeColor = bakeExplicitAlpha(edgeColor, opt.Alpha)

	// Get baseline
	baseline := 0.0
	if v, ok := opt.Baseline.Get(); ok {
		baseline = v
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
		Antialias: patchAntialiasMode(&rcPatch, opt.Antialiased),
		Label:     opt.Label,
	}

	a.Add(fill)
	a.autoScaleIfEnabled(defaultAutoScaleMargin)
	return fill
}
