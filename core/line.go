package core

import (
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/transform"
)

// LineDrawStyle controls how consecutive data points are connected.
type LineDrawStyle uint8

const (
	LineDrawStyleDefault LineDrawStyle = iota
	LineDrawStyleStepsPre
	LineDrawStyleStepsMid
	LineDrawStyleStepsPost
)

// LineStyle is a Matplotlib linestyle spec for line plots. The constants
// cover the named rcParams values; full-name spellings ("dashed", "dotted",
// "dashdot", "solid") are accepted too. The empty string means "unset" (solid
// unless the property cycle carries a linestyle), matching PlotOptions'
// zero-value semantics.
type LineStyle string

const (
	LineStyleSolid   LineStyle = "-"
	LineStyleDashed  LineStyle = "--"
	LineStyleDashDot LineStyle = "-."
	LineStyleDotted  LineStyle = ":"
	// LineStyleNone suppresses the line segments entirely (markers-only plot),
	// matplotlib's linestyle "none".
	LineStyleNone LineStyle = "none"
)

// isNone reports whether the style suppresses line drawing (matplotlib
// accepts "none"/"None" and blank-space spellings; the empty string is
// "unset", not none).
func (s LineStyle) isNone() bool {
	if s == "" {
		return false // unset
	}
	trimmed := strings.TrimSpace(strings.ToLower(string(s)))
	return trimmed == "none" || trimmed == "" // " " also means none in matplotlib
}

// DashUnits describes the unit system used by Line2D.Dashes.
type DashUnits uint8

const (
	// DashUnitsPixels interprets Dashes as renderer pixel lengths. This keeps
	// existing direct field assignments and backend-facing paint unchanged.
	DashUnitsPixels DashUnits = iota
	// DashUnitsMatplotlib interprets Dashes like Matplotlib Line2D.set_dashes:
	// input lengths are in points and, with the default scale_dashes=True, are
	// scaled by the line width before rasterization.
	DashUnitsMatplotlib
)

// MarkEveryMode identifies a marker subsampling strategy for Line2D.
type MarkEveryMode uint8

const (
	MarkEveryNone MarkEveryMode = iota
	MarkEveryStartStep
	MarkEveryIndices
	MarkEverySlice
)

// MarkEverySpec describes Matplotlib-style Line2D marker subsampling.
//
// StartStep matches markevery=(start, step). Slice matches a Go-style
// start:stop:step range over data point indices. Indices draws markers at the
// explicit zero-based data point indices.
type MarkEverySpec struct {
	Mode    MarkEveryMode
	Start   int
	Stop    int
	Step    int
	Indices []int
}

// MarkerColorMode describes Matplotlib-style marker color sentinels.
type MarkerColorMode uint8

const (
	MarkerColorDefault MarkerColorMode = iota
	MarkerColorExplicit
	MarkerColorAuto
	MarkerColorNone
)

// MarkerColorSpec stores either an explicit marker color or a sentinel such as
// Matplotlib's "auto" and "none" values.
type MarkerColorSpec struct {
	Mode  MarkerColorMode
	Color render.Color
}

// ExplicitMarkerColor returns a marker color spec for an explicit RGBA color.
func ExplicitMarkerColor(color render.Color) MarkerColorSpec {
	return MarkerColorSpec{Mode: MarkerColorExplicit, Color: color}
}

// AutoMarkerColor returns a marker color spec that resolves from the line color.
func AutoMarkerColor() MarkerColorSpec {
	return MarkerColorSpec{Mode: MarkerColorAuto}
}

// NoMarkerColor returns a marker color spec equivalent to Matplotlib's "none".
func NoMarkerColor() MarkerColorSpec {
	return MarkerColorSpec{Mode: MarkerColorNone}
}

// EveryNMarkers returns a markevery spec that starts at the first point.
func EveryNMarkers(step int) MarkEverySpec {
	return StartStepMarkers(0, step)
}

// StartStepMarkers returns a markevery=(start, step) equivalent.
func StartStepMarkers(start, step int) MarkEverySpec {
	return MarkEverySpec{Mode: MarkEveryStartStep, Start: start, Step: step}
}

// IndexedMarkers returns a markevery spec for explicit data point indices.
func IndexedMarkers(indices ...int) MarkEverySpec {
	return MarkEverySpec{Mode: MarkEveryIndices, Indices: append([]int(nil), indices...)}
}

// SliceMarkers returns a Go-style start:stop:step markevery range. Use stop <= 0
// to select through the end of the data.
func SliceMarkers(start, stop, step int) MarkEverySpec {
	return MarkEverySpec{Mode: MarkEverySlice, Start: start, Stop: stop, Step: step}
}

// Line2D is a polyline artist with optional Matplotlib-style data markers.
type Line2D struct {
	ArtistRasterization
	XY              []geom.Pt    // data space points
	W               float64      // stroke width in points
	Col             render.Color // stroke color
	Dashes          []float64    // dash pattern (on/off pairs)
	DashUnits       DashUnits    // unit system for Dashes
	LineCap         render.LineCap
	LineCapSet      bool
	GapColor        render.Color // optional dashed-line gap color
	GapColorSet     bool
	PathEffects     []render.PathEffect
	DrawStyle       LineDrawStyle // optional step-style connection mode
	Marker          MarkerType    // optional data marker
	MarkerSet       bool          // true when Marker should be drawn
	MarkerStyle     MarkerStyle   // optional rich marker style
	MarkerPath      geom.Path     // optional custom marker path in normalized marker space
	MarkerSize      float64       // marker size in points, 0 uses Matplotlib's 6 pt default
	MarkerFaceColor render.Color  // marker fill, 0 alpha falls back to line color
	MarkerEdgeColor render.Color  // marker edge, 0 alpha falls back to line color
	MarkerEdgeWidth float64       // marker edge width in points, 0 uses Matplotlib's 1 pt default
	MarkerFaceSpec  MarkerColorSpec
	MarkerEdgeSpec  MarkerColorSpec
	MarkerFaceAlt   MarkerColorSpec
	MarkEvery       int                 // optional every-N marker subset; <=1 draws every point
	MarkEverySpec   MarkEverySpec       // optional richer marker subset; overrides MarkEvery when set
	Label           string              // series label for legend
	Sketch          render.SketchParams // per-artist sketch/xkcd override; zero inherits the figure default
	z               float64             // z-order
	pickRadius      float64             // pick tolerance in pixels (0 = default)

	// Persistent affine/non-affine cache for the data-coordinate draw path
	// (Phase 13). transformedPath caches the non-affine projection of the source
	// points; an affine-only redraw (axes resize/pan/zoom) reuses it. cacheTransform
	// and cachePoints track the inputs the cache was built from so it is rebuilt
	// only when the transform identity or the source point backing changes.
	transformedPath *transform.TransformedPath
	cacheTransform  transform.T
	cachePoints     []geom.Pt
}

// Data returns cloned x and y data slices.
func (l *Line2D) Data() ([]float64, []float64) {
	if l == nil || len(l.XY) == 0 {
		return nil, nil
	}
	x := make([]float64, len(l.XY))
	y := make([]float64, len(l.XY))
	for i, pt := range l.XY {
		x[i] = pt.X
		y[i] = pt.Y
	}
	return x, y
}

// SetData replaces x and y data, truncating to the shorter slice like Plot.
func (l *Line2D) SetData(x, y []float64) {
	if l == nil {
		return
	}
	n := min(len(x), len(y))
	l.XY = make([]geom.Pt, n)
	for i := 0; i < n; i++ {
		l.XY[i] = geom.Pt{X: x[i], Y: y[i]}
	}
	l.SetStale(true)
}

// SetXData replaces the x data while preserving existing y values.
func (l *Line2D) SetXData(x []float64) {
	if l == nil {
		return
	}
	n := min(len(x), len(l.XY))
	next := make([]geom.Pt, n)
	for i := 0; i < n; i++ {
		next[i] = geom.Pt{X: x[i], Y: l.XY[i].Y}
	}
	l.XY = next
	l.SetStale(true)
}

// SetYData replaces the y data while preserving existing x values.
func (l *Line2D) SetYData(y []float64) {
	if l == nil {
		return
	}
	n := min(len(y), len(l.XY))
	next := make([]geom.Pt, n)
	for i := 0; i < n; i++ {
		next[i] = geom.Pt{X: l.XY[i].X, Y: y[i]}
	}
	l.XY = next
	l.SetStale(true)
}

// SetDashes sets the dash sequence using Matplotlib Line2D.set_dashes units.
func (l *Line2D) SetDashes(seq ...float64) {
	if len(seq) == 0 {
		l.Dashes = nil
		l.DashUnits = DashUnitsPixels
		l.SetStale(true)
		return
	}
	l.Dashes = append([]float64(nil), seq...)
	l.DashUnits = DashUnitsMatplotlib
	l.SetStale(true)
}

// SetGapColor sets the color used to paint dashed-line gaps.
func (l *Line2D) SetGapColor(color render.Color) {
	if l == nil {
		return
	}
	l.GapColor = color
	l.GapColorSet = true
	l.SetStale(true)
}

// SetMarkerFaceColor sets an explicit marker face color.
func (l *Line2D) SetMarkerFaceColor(color render.Color) {
	if l == nil {
		return
	}
	l.MarkerFaceColor = color
	l.MarkerFaceSpec = ExplicitMarkerColor(color)
	l.SetStale(true)
}

// SetMarkerFaceColorAuto makes the marker face color follow the line color.
func (l *Line2D) SetMarkerFaceColorAuto() {
	if l == nil {
		return
	}
	l.MarkerFaceSpec = AutoMarkerColor()
	l.SetStale(true)
}

// SetMarkerFaceColorNone disables marker face filling.
func (l *Line2D) SetMarkerFaceColorNone() {
	if l == nil {
		return
	}
	l.MarkerFaceSpec = NoMarkerColor()
	l.SetStale(true)
}

// SetMarkerFaceColorAlt sets the alternate face color used by half-filled markers.
func (l *Line2D) SetMarkerFaceColorAlt(color render.Color) {
	if l == nil {
		return
	}
	l.MarkerFaceAlt = ExplicitMarkerColor(color)
	l.SetStale(true)
}

// SetMarkerFaceColorAltNone disables the alternate half fill.
func (l *Line2D) SetMarkerFaceColorAltNone() {
	if l == nil {
		return
	}
	l.MarkerFaceAlt = NoMarkerColor()
	l.SetStale(true)
}

// SetMarkerEdgeColor sets an explicit marker edge color.
func (l *Line2D) SetMarkerEdgeColor(color render.Color) {
	if l == nil {
		return
	}
	l.MarkerEdgeColor = color
	l.MarkerEdgeSpec = ExplicitMarkerColor(color)
	l.SetStale(true)
}

// SetMarkerEdgeColorAuto makes the marker edge color follow Matplotlib's auto
// behavior: line RGB, inheriting the face alpha when the face is filled.
func (l *Line2D) SetMarkerEdgeColorAuto() {
	if l == nil {
		return
	}
	l.MarkerEdgeSpec = AutoMarkerColor()
	l.SetStale(true)
}

// SetMarkerEdgeColorNone disables marker edge stroking.
func (l *Line2D) SetMarkerEdgeColorNone() {
	if l == nil {
		return
	}
	l.MarkerEdgeSpec = NoMarkerColor()
	l.SetStale(true)
}

// ClearGapColor disables dashed-line gap painting.
func (l *Line2D) ClearGapColor() {
	if l == nil {
		return
	}
	l.GapColor = render.Color{}
	l.GapColorSet = false
	l.SetStale(true)
}

// SetMarkEvery sets a rich marker subsampling spec.
func (l *Line2D) SetMarkEvery(spec MarkEverySpec) {
	if l == nil {
		return
	}
	spec.Indices = append([]int(nil), spec.Indices...)
	l.MarkEverySpec = spec
	l.SetStale(true)
}

// Draw renders the line by transforming points to pixel space and drawing a path.
func (l *Line2D) Draw(r render.Renderer, ctx *DrawContext) {
	p := l.displayPath(ctx)
	if len(p.C) == 0 {
		return // nothing to draw
	}

	lineRC := style.Default
	if ctx != nil {
		lineRC = ctx.RC
	}
	widthPx := pointsToPixels(lineRC, l.W)
	dashes := lineDashesForPaint(l.Dashes, widthPx, l.DashUnits)
	lineCap := render.CapSquare
	if len(dashes) > 0 {
		lineCap = render.CapButt
	}
	if l.LineCapSet {
		lineCap = l.LineCap
	}
	paint := render.Paint{
		LineWidth:   widthPx,
		LineJoin:    render.JoinRound, // Default to round joins
		LineCap:     lineCap,
		MiterLimit:  10.0, // Standard miter limit
		Stroke:      l.ApplyArtistAlpha(l.Col),
		Dashes:      dashes,
		PathEffects: linePathEffects(ctx, l.PathEffects),
		Snap:        render.SnapAuto,
		Simplify:    ctx != nil && ctx.RC.PathSimplify,
	}
	if ctx != nil {
		paint.SimplifyThreshold = ctx.RC.PathSimplifyThreshold
		paint.MaxChunkVertices = ctx.RC.AggPathChunkSize
	}
	paint.Sketch = l.Sketch
	if l.GapColorSet && l.GapColor.A > 0 && l.W > 0 && len(dashes) >= 2 {
		if gapPath := dashGapPath(p, dashes); len(gapPath.C) > 0 {
			gapPaint := paint
			gapPaint.Stroke = l.ApplyArtistAlpha(l.GapColor)
			gapPaint.Dashes = nil
			r.Path(gapPath, &gapPaint)
		}
	}
	r.Path(p, &paint)
	l.drawMarkers(r, ctx)
}

func lineDashesForPaint(dashes []float64, lineWidth float64, units DashUnits) []float64 {
	if len(dashes) == 0 {
		return nil
	}
	out := append([]float64(nil), dashes...)
	if units == DashUnitsMatplotlib {
		for i := range out {
			out[i] *= lineWidth
		}
	}
	return out
}

// Z returns the z-order for sorting.
func (l *Line2D) Z() float64 {
	return zOrDefault(l.z, defaultLineZ)
}

// Bounds returns the bounding box of all points in data space.
func (l *Line2D) Bounds(*DrawContext) geom.Rect {
	if len(l.XY) == 0 || !artistUsesDataCoords(l, Coords(CoordData)) {
		return geom.Rect{}
	}
	var r geom.Rect
	ok := false
	for _, pt := range l.XY {
		if !finitePoint(pt) {
			continue
		}
		if !ok {
			r = geom.Rect{Min: pt, Max: pt}
			ok = true
			continue
		}
		r = expandRect(r, pt)
	}
	if !ok {
		return geom.Rect{}
	}
	return r
}

func (l *Line2D) pathPoints() []geom.Pt {
	if len(l.XY) < 2 {
		return l.XY
	}

	switch l.DrawStyle {
	case LineDrawStyleStepsPre:
		out := make([]geom.Pt, 0, 2*len(l.XY)-1)
		out = append(out, l.XY[0])
		for i := 1; i < len(l.XY); i++ {
			out = append(
				out,
				geom.Pt{X: l.XY[i-1].X, Y: l.XY[i].Y},
				l.XY[i],
			)
		}
		return out
	case LineDrawStyleStepsMid:
		out := make([]geom.Pt, 0, 3*len(l.XY)-2)
		out = append(out, l.XY[0])
		for i := 1; i < len(l.XY); i++ {
			midX := (l.XY[i-1].X + l.XY[i].X) / 2
			out = append(
				out,
				geom.Pt{X: midX, Y: l.XY[i-1].Y},
				geom.Pt{X: midX, Y: l.XY[i].Y},
				l.XY[i],
			)
		}
		return out
	case LineDrawStyleStepsPost:
		out := make([]geom.Pt, 0, 2*len(l.XY)-1)
		out = append(out, l.XY[0])
		for i := 1; i < len(l.XY); i++ {
			out = append(
				out,
				geom.Pt{X: l.XY[i].X, Y: l.XY[i-1].Y},
				l.XY[i],
			)
		}
		return out
	default:
		return l.XY
	}
}

func (l *Line2D) displayPath(ctx *DrawContext) geom.Path {
	points := l.pathPoints()
	if len(points) == 0 {
		return geom.Path{}
	}
	if ctx != nil && isGeoProjection(ctx.Projection) && artistUsesDataCoords(l, Coords(CoordData)) {
		points = interpolateFiniteLineSegments(points, geoGridSegments)
	}
	tr := artistTransformFor(ctx, l, Coords(CoordData))

	// Cached fast path: when the line draws in data coordinates through the
	// persistent axes transform graph, route the projection through a
	// transform.TransformedPath so an affine-only redraw (axes resize/pan/zoom)
	// reuses the cached non-affine vertex pass instead of re-projecting per
	// vertex. Applying the trailing affine to the cached non-affine points and
	// then segmenting on the final points is byte-equivalent to the direct loop:
	// the (non-)affine of a non-finite point stays non-finite, so a single
	// finiteness check on the final point matches the direct pre+post checks.
	if final, ok := l.cachedDisplayPoints(ctx, points, tr); ok {
		return segmentFinitePath(final)
	}

	return transformAndSegment(points, tr)
}

// cachedDisplayPoints returns the data points fully transformed through the
// cached TransformedPath, reporting ok only when caching is eligible: the line
// draws in data coordinates with no explicit transform override and the
// persistent axes transform graph is available to drive invalidation. Other
// cases (explicit transforms, axes/figure/blended coordinates, no graph) fall
// back to the direct per-vertex transform.
func (l *Line2D) cachedDisplayPoints(ctx *DrawContext, points []geom.Pt, tr transform.T) ([]geom.Pt, bool) {
	if ctx == nil || tr == nil || !artistUsesDataCoords(l, Coords(CoordData)) {
		return nil, false
	}
	deps := ctx.dataTransformDeps()
	if len(deps) == 0 {
		return nil, false
	}
	if _, fullyAffine := transform.AsAffine(tr); fullyAffine {
		// A fully-affine transform (linear rectilinear axes) has no non-affine
		// projection to cache, and collapsing scale+bbox into a single matrix
		// would change the per-vertex arithmetic versus the direct two-step
		// chain (last-ULP differences). Use the direct path so output stays
		// byte-identical. The cache engages only when a genuine non-affine leg
		// (log/symlog/polar/geo) is present, where the trailing affine is the
		// single bbox matrix and matches the direct chain exactly.
		return nil, false
	}

	src := geom.Path{V: points}
	switch {
	case l.transformedPath == nil || l.cacheTransform != tr:
		l.transformedPath = transform.NewTransformedPath(src, tr, deps...)
		l.cacheTransform = tr
		l.cachePoints = points
	case !samePointsHeader(l.cachePoints, points):
		l.transformedPath.SetPath(src)
		l.cachePoints = points
	}

	nonAffine, affine := l.transformedPath.TransformedPointsAndAffine()
	final := make([]geom.Pt, len(nonAffine.V))
	for i, v := range nonAffine.V {
		final[i] = affine.Apply(v)
	}
	return final, true
}

// samePointsHeader reports whether two point slices share the same backing
// array and length, a cheap O(1) check for "the source points did not change".
// Default-style, non-geo lines reuse the stable l.XY backing across draws (so
// the cache is reused); steps/geo styles allocate fresh slices each draw and so
// reproject, which is correct if not optimal.
func samePointsHeader(a, b []geom.Pt) bool {
	if len(a) != len(b) {
		return false
	}
	if len(a) == 0 {
		return true
	}
	return &a[0] == &b[0]
}

// segmentFinitePath builds a polyline path from already-transformed points,
// breaking the path at every non-finite point (NaN/Inf gaps).
func segmentFinitePath(points []geom.Pt) geom.Path {
	p := geom.Path{}
	inSegment := false
	for _, q := range points {
		if !finitePoint(q) {
			inSegment = false
			continue
		}
		if !inSegment {
			p.C = append(p.C, geom.MoveTo)
			inSegment = true
		} else {
			p.C = append(p.C, geom.LineTo)
		}
		p.V = append(p.V, q)
	}
	return p
}

// transformAndSegment applies tr to each point and segments on finiteness,
// matching the historical Line2D draw path byte-for-byte (finiteness is checked
// both before and after the transform).
func transformAndSegment(points []geom.Pt, tr transform.T) geom.Path {
	p := geom.Path{}
	inSegment := false
	for _, v := range points {
		if !finitePoint(v) {
			inSegment = false
			continue
		}
		q := v
		if tr != nil {
			q = tr.Apply(v)
		}
		if !finitePoint(q) {
			inSegment = false
			continue
		}
		if !inSegment {
			p.C = append(p.C, geom.MoveTo)
			inSegment = true
		} else {
			p.C = append(p.C, geom.LineTo)
		}
		p.V = append(p.V, q)
	}
	return p
}

func interpolateFiniteLineSegments(points []geom.Pt, steps int) []geom.Pt {
	if len(points) < 2 || steps <= 1 {
		return points
	}
	out := make([]geom.Pt, 0, 1+(len(points)-1)*steps)
	var prev geom.Pt
	havePrev := false
	for _, pt := range points {
		if !finitePoint(pt) {
			out = append(out, pt)
			havePrev = false
			continue
		}
		if !havePrev {
			out = append(out, pt)
			prev = pt
			havePrev = true
			continue
		}
		for i := 1; i <= steps; i++ {
			t := float64(i) / float64(steps)
			out = append(out, interpolateLinePoint(prev, pt, t))
		}
		prev = pt
	}
	return out
}

func (l *Line2D) drawMarkers(r render.Renderer, ctx *DrawContext) {
	if l == nil || r == nil || ctx == nil || !l.hasMarkers() {
		return
	}
	points := l.markerPoints()
	if len(points) == 0 {
		return
	}
	spec := l.markerPathSpec(r, ctx)
	if len(spec.Path.C) == 0 {
		return
	}
	markerSize := l.resolvedMarkerSize(ctx)
	if markerSize <= 0 {
		return
	}

	newCollection := func(path geom.Path, face, edge render.Color, edgeWidth float64, lineOnly bool) *PathCollection {
		return &PathCollection{
			Collection: Collection{
				ArtistRasterization: l.ArtistRasterization,
				Coords:              Coords(CoordData),
				Alpha:               1,
				PathEffects:         cloneRenderPathEffects(l.PathEffects),
			},
			Path:          path,
			Offsets:       points,
			Size:          markerSize,
			PathInDisplay: true,
			FaceColor:     face,
			EdgeColor:     edge,
			EdgeWidth:     edgeWidth,
			LineJoin:      (&Scatter2D{Marker: l.Marker, MarkerStyle: l.MarkerStyle, MarkerPath: l.MarkerPath}).markerLineJoin(),
			LineJoinSet:   true,
			LineCap:       render.CapButt,
			LineCapSet:    true,
			LineOnly:      lineOnly,
		}
	}

	if spec.HasAlt {
		edge := l.resolvedMarkerEdgeColor()
		edgeWidth := l.resolvedMarkerEdgeWidth(ctx)
		newCollection(spec.Path, l.resolvedMarkerFaceColor(), edge, edgeWidth, false).Draw(r, ctx)
		newCollection(spec.AltPath, l.resolvedMarkerFaceColorAlt(), edge, edgeWidth, false).Draw(r, ctx)
		return
	}

	markers := newCollection(
		spec.Path,
		l.resolvedMarkerFaceColor(),
		l.resolvedMarkerEdgeColor(),
		l.resolvedMarkerEdgeWidth(ctx),
		markerLineOnly(l.resolvedMarkerStyle()),
	)
	markers.Draw(r, ctx)
}

type markerPathSpec struct {
	Path     geom.Path
	AltPath  geom.Path
	EdgePath geom.Path
	HasAlt   bool
}

func (l *Line2D) markerPathSpec(r render.Renderer, ctx *DrawContext) markerPathSpec {
	base := l.markerPrototypePath(r, ctx)
	spec := markerPathSpec{Path: base, EdgePath: base}
	style := l.resolvedMarkerStyle()
	if markerLineOnly(style) {
		return spec
	}
	if primary, alt, ok := splitMarkerFillPaths(style, base); ok {
		spec.Path = primary
		spec.AltPath = alt
		spec.HasAlt = len(primary.C) > 0 && len(alt.C) > 0
	}
	return spec
}

func (l *Line2D) hasMarkers() bool {
	if l == nil {
		return false
	}
	return l.MarkerSet || len(l.MarkerPath.C) > 0 || l.MarkerStyle.Type != 0 || l.MarkerStyle.FillStyle != 0 ||
		l.MarkerStyle.Tuple != nil || l.MarkerStyle.MathText != "" || len(l.MarkerStyle.Path.C) > 0
}

func (l *Line2D) markerPoints() []geom.Pt {
	if l == nil || len(l.XY) == 0 {
		return nil
	}
	if l.MarkEverySpec.Mode != MarkEveryNone {
		return l.markerPointsForSpec(l.MarkEverySpec)
	}
	if l.MarkEvery <= 1 {
		return finitePoints(l.XY)
	}
	out := make([]geom.Pt, 0, (len(l.XY)+l.MarkEvery-1)/l.MarkEvery)
	for i, pt := range l.XY {
		if i%l.MarkEvery == 0 && finitePoint(pt) {
			out = append(out, pt)
		}
	}
	return out
}

func (l *Line2D) markerPointsForSpec(spec MarkEverySpec) []geom.Pt {
	switch spec.Mode {
	case MarkEveryStartStep:
		return l.markerPointsStartStep(spec.Start, spec.Step)
	case MarkEveryIndices:
		out := make([]geom.Pt, 0, len(spec.Indices))
		for _, idx := range spec.Indices {
			if idx < 0 {
				idx += len(l.XY)
			}
			if idx >= 0 && idx < len(l.XY) && finitePoint(l.XY[idx]) {
				out = append(out, l.XY[idx])
			}
		}
		return out
	case MarkEverySlice:
		stop := spec.Stop
		if stop <= 0 || stop > len(l.XY) {
			stop = len(l.XY)
		}
		return l.markerPointsRange(spec.Start, stop, spec.Step)
	default:
		return finitePoints(l.XY)
	}
}

func (l *Line2D) markerPointsStartStep(start, step int) []geom.Pt {
	if step <= 0 {
		return finitePoints(l.XY)
	}
	if start < 0 {
		start += len(l.XY)
	}
	return l.markerPointsRange(start, len(l.XY), step)
}

func (l *Line2D) markerPointsRange(start, stop, step int) []geom.Pt {
	if step <= 0 {
		return finitePoints(l.XY)
	}
	if start < 0 {
		start += len(l.XY)
	}
	if stop < 0 {
		stop += len(l.XY)
	}
	if start < 0 {
		start = 0
	}
	if stop > len(l.XY) {
		stop = len(l.XY)
	}
	if start >= stop {
		return nil
	}
	out := make([]geom.Pt, 0, (stop-start+step-1)/step)
	for i := start; i < stop; i += step {
		if finitePoint(l.XY[i]) {
			out = append(out, l.XY[i])
		}
	}
	return out
}

func finitePoints(points []geom.Pt) []geom.Pt {
	out := make([]geom.Pt, 0, len(points))
	for _, pt := range points {
		if finitePoint(pt) {
			out = append(out, pt)
		}
	}
	return out
}

func finitePoint(pt geom.Pt) bool {
	return !math.IsNaN(pt.X) && !math.IsInf(pt.X, 0) && !math.IsNaN(pt.Y) && !math.IsInf(pt.Y, 0)
}

func dashGapPath(path geom.Path, dashes []float64) geom.Path {
	if len(path.C) == 0 || len(dashes) < 2 {
		return geom.Path{}
	}
	pattern := make([]float64, 0, len(dashes))
	for _, d := range dashes {
		if d > 0 {
			pattern = append(pattern, d)
		}
	}
	if len(pattern) < 2 {
		return geom.Path{}
	}
	if len(pattern)%2 == 1 {
		pattern = append(pattern, pattern...)
	}

	var out geom.Path
	var cur geom.Pt
	haveCur := false
	dashIndex := 0
	dashRemaining := pattern[0]
	vi := 0
	const epsilon = 1e-10

	for _, cmd := range path.C {
		switch cmd {
		case geom.MoveTo:
			if vi >= len(path.V) {
				return out
			}
			cur = path.V[vi]
			vi++
			haveCur = true
			dashIndex = 0
			dashRemaining = pattern[0]
		case geom.LineTo:
			if vi >= len(path.V) {
				return out
			}
			next := path.V[vi]
			vi++
			if !haveCur {
				cur = next
				haveCur = true
				continue
			}
			segLen := pointDistance(cur, next)
			if segLen <= epsilon {
				cur = next
				continue
			}
			consumed := 0.0
			for consumed < segLen-epsilon {
				available := segLen - consumed
				step := math.Min(available, dashRemaining)
				if dashIndex%2 == 1 && step > epsilon {
					t0 := consumed / segLen
					t1 := (consumed + step) / segLen
					start := interpolateLinePoint(cur, next, t0)
					end := interpolateLinePoint(cur, next, t1)
					out.MoveTo(start)
					out.LineTo(end)
				}
				consumed += step
				dashRemaining -= step
				if dashRemaining <= epsilon {
					dashIndex = (dashIndex + 1) % len(pattern)
					dashRemaining = pattern[dashIndex]
				}
			}
			cur = next
		case geom.QuadTo:
			vi += 2
			haveCur = false
		case geom.CubicTo:
			vi += 3
			haveCur = false
		case geom.ClosePath:
			haveCur = false
		}
	}
	return out
}

func pointDistance(a, b geom.Pt) float64 {
	return math.Hypot(b.X-a.X, b.Y-a.Y)
}

func interpolateLinePoint(a, b geom.Pt, t float64) geom.Pt {
	return geom.Pt{
		X: a.X + (b.X-a.X)*t,
		Y: a.Y + (b.Y-a.Y)*t,
	}
}

func (l *Line2D) resolvedMarkerStyle() MarkerStyle {
	if l == nil {
		return MarkerStyle{}
	}
	if l.MarkerStyle.Tuple != nil || l.MarkerStyle.MathText != "" || len(l.MarkerStyle.Path.C) > 0 || l.MarkerStyle.Type != 0 || l.MarkerStyle.FillStyle != 0 {
		style := l.MarkerStyle
		if style.FillStyle == 0 {
			style.FillStyle = MarkerFillFull
		}
		return style
	}
	return MarkerStyle{Type: l.Marker, FillStyle: MarkerFillFull}
}

func (l *Line2D) markerPrototypePath(r render.Renderer, ctx *DrawContext) geom.Path {
	if l == nil {
		return geom.Path{}
	}
	scatter := Scatter2D{
		Marker:      l.Marker,
		MarkerStyle: l.resolvedMarkerStyle(),
		MarkerPath:  l.MarkerPath,
	}
	return scatter.markerPrototypePathForContext(r, ctx)
}

func (l *Line2D) resolvedMarkerSize(ctx *DrawContext) float64 {
	size := 6.0
	if l != nil && l.MarkerSize > 0 {
		size = l.MarkerSize
	}
	rc := style.Default
	if ctx != nil {
		rc = ctx.RC
	}
	return pointsToPixels(rc, size)
}

func (l *Line2D) markerFaceNone() bool {
	if l == nil {
		return true
	}
	return l.resolvedMarkerStyle().FillStyle == MarkerFillNone || l.MarkerFaceSpec.Mode == MarkerColorNone
}

func (l *Line2D) resolvedMarkerFaceColor() render.Color {
	if l == nil {
		return render.Color{}
	}
	if l.resolvedMarkerStyle().FillStyle == MarkerFillNone {
		return render.Color{}
	}
	var color render.Color
	switch l.MarkerFaceSpec.Mode {
	case MarkerColorExplicit:
		color = l.MarkerFaceSpec.Color
	case MarkerColorNone:
		return render.Color{}
	case MarkerColorAuto:
		color = l.Col
	default:
		color = l.MarkerFaceColor
		if color.A <= 0 {
			color = l.Col
		}
	}
	return l.ApplyArtistAlpha(color)
}

func (l *Line2D) resolvedMarkerFaceColorAlt() render.Color {
	if l == nil {
		return render.Color{}
	}
	if l.resolvedMarkerStyle().FillStyle == MarkerFillNone {
		return render.Color{}
	}
	switch l.MarkerFaceAlt.Mode {
	case MarkerColorExplicit:
		return l.ApplyArtistAlpha(l.MarkerFaceAlt.Color)
	case MarkerColorAuto:
		return l.ApplyArtistAlpha(l.Col)
	default:
		return render.Color{}
	}
}

func (l *Line2D) resolvedMarkerEdgeColor() render.Color {
	if l == nil {
		return render.Color{}
	}
	switch l.MarkerEdgeSpec.Mode {
	case MarkerColorExplicit:
		return l.ApplyArtistAlpha(l.MarkerEdgeSpec.Color)
	case MarkerColorNone:
		return render.Color{}
	case MarkerColorAuto:
		edge := l.ApplyArtistAlpha(l.Col)
		if !l.markerFaceNone() {
			edge.A = l.resolvedMarkerFaceColor().A
		}
		return edge
	default:
		if l.MarkerEdgeColor.A > 0 {
			return l.ApplyArtistAlpha(l.MarkerEdgeColor)
		}
		edge := l.ApplyArtistAlpha(l.Col)
		if !l.markerFaceNone() {
			edge.A = l.resolvedMarkerFaceColor().A
		}
		return edge
	}
}

// resolvedMarkerEdgeWidth returns the marker edge width in points (matplotlib
// markeredgewidth defaults to 1.0 pt). Callers convert to device pixels at the
// Paint sink.
func (l *Line2D) resolvedMarkerEdgeWidth(_ *DrawContext) float64 {
	width := 1.0
	if l != nil && l.MarkerEdgeWidth > 0 {
		width = l.MarkerEdgeWidth
	}
	return width
}
