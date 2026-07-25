package core

import (
	"fmt"
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// StepWhere controls where step transitions happen between sample points.
type StepWhere uint8

const (
	StepWherePre StepWhere = iota
	StepWhereMid
	StepWherePost
)

// StepOptions configures Axes.Step.
type StepOptions struct {
	Color     *render.Color
	LineWidth *float64
	Dashes    []float64
	Where     *StepWhere
	Label     string
	Alpha     *float64
}

// StairsOptions configures Axes.Stairs.
type StairsOptions struct {
	Color     *render.Color
	EdgeColor *render.Color
	LineWidth *float64
	Alpha     *float64
	Baseline  *float64
	Fill      *bool
	Label     string
}

// ReferenceLineOptions configures reference lines such as axhline/axvline/axline.
type ReferenceLineOptions struct {
	Color     *render.Color
	LineWidth *float64
	Dashes    []float64
	Alpha     *float64
}

// HLineOptions configures AxHLine.
type HLineOptions struct {
	Color     *render.Color
	LineWidth *float64
	Dashes    []float64
	Alpha     *float64
	XMin      *float64
	XMax      *float64
}

// VLineOptions configures AxVLine.
type VLineOptions struct {
	Color     *render.Color
	LineWidth *float64
	Dashes    []float64
	Alpha     *float64
	YMin      *float64
	YMax      *float64
}

// SpanOptions configures span helpers such as axhspan and axvspan.
type SpanOptions struct {
	Color       *render.Color
	EdgeColor   *render.Color
	EdgeWidth   *float64
	Alpha       *float64
	Antialiased *bool
}

// HSpanOptions configures AxHSpan.
type HSpanOptions struct {
	Color       *render.Color
	EdgeColor   *render.Color
	EdgeWidth   *float64
	Alpha       *float64
	Antialiased *bool
	XMin        *float64
	XMax        *float64
}

// VSpanOptions configures AxVSpan.
type VSpanOptions struct {
	Color       *render.Color
	EdgeColor   *render.Color
	EdgeWidth   *float64
	Alpha       *float64
	Antialiased *bool
	YMin        *float64
	YMax        *float64
}

// BarLabelOptions configures bar-label placement.
type BarLabelOptions struct {
	FontSize float64
	Color    render.Color
	Padding  float64
	Format   string
	Position string
}

// Stairs2D renders pre-binned histogram-style steps, optionally filled to a baseline.
type Stairs2D struct {
	Edges     []float64
	Values    []float64
	Baseline  float64
	Fill      bool
	Color     render.Color
	EdgeColor render.Color
	LineWidth float64
	Alpha     float64
	Label     string
	z         float64
}

// Segment2D renders a straight line segment in arbitrary coordinate spaces.
type Segment2D struct {
	Start     geom.Pt
	End       geom.Pt
	Coords    CoordinateSpec
	Color     render.Color
	LineWidth float64
	Dashes    []float64
	z         float64
}

// Span2D renders a filled rectangle in arbitrary coordinate spaces.
type Span2D struct {
	Start     geom.Pt
	End       geom.Pt
	Coords    CoordinateSpec
	Color     render.Color
	EdgeColor render.Color
	EdgeWidth float64
	Antialias render.AntialiasMode
	z         float64
}

// InfiniteLine2D renders an infinite data-space line clipped to the current axes view.
type InfiniteLine2D struct {
	Point     geom.Pt
	Direction geom.Pt
	Color     render.Color
	LineWidth float64
	Dashes    []float64
	z         float64
}

// Step draws a step-connected line through the provided samples.
func (a *Axes) Step(x, y []float64, opts ...StepOptions) *Line2D {
	var opt StepOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	where := StepWherePre
	if opt.Where != nil {
		where = *opt.Where
	}
	drawStyle := lineDrawStyleFromStepWhere(where)

	return a.plot(x, y, PlotOptions{
		Color:     opt.Color,
		LineWidth: opt.LineWidth,
		Dashes:    opt.Dashes,
		DrawStyle: &drawStyle,
		Label:     opt.Label,
		Alpha:     opt.Alpha,
	})
}

// Stairs draws pre-binned histogram-style steps from explicit bin edges.
func (a *Axes) Stairs(values, edges []float64, opts ...StairsOptions) *Stairs2D {
	if len(values) == 0 || len(edges) < 2 {
		return nil
	}

	var opt StairsOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	lineWidth := 1.5
	if opt.LineWidth != nil {
		lineWidth = *opt.LineWidth
	}

	alpha := 1.0
	if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
		alpha = *opt.Alpha
	}

	baseline := 0.0
	if opt.Baseline != nil {
		baseline = *opt.Baseline
	}

	fill := false
	if opt.Fill != nil {
		fill = *opt.Fill
	}

	color := a.NextColor()
	if opt.Color != nil {
		color = *opt.Color
	}

	edgeColor := render.Color{}
	if opt.EdgeColor != nil {
		edgeColor = *opt.EdgeColor
	}

	stairs := &Stairs2D{
		Edges:     append([]float64(nil), edges...),
		Values:    append([]float64(nil), values...),
		Baseline:  baseline,
		Fill:      fill,
		Color:     color,
		EdgeColor: edgeColor,
		LineWidth: lineWidth,
		Alpha:     alpha,
		Label:     opt.Label,
	}
	a.Add(stairs)
	return stairs
}

// AxHLine draws a horizontal reference line using axes-fraction x coordinates and a data-space y value.
func (a *Axes) AxHLine(y float64, opts ...HLineOptions) *Segment2D {
	var opt HLineOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	xMin := 0.0
	xMax := 1.0
	if opt.XMin != nil {
		xMin = *opt.XMin
	}
	if opt.XMax != nil {
		xMax = *opt.XMax
	}
	line := a.newSegment(
		geom.Pt{X: xMin, Y: y},
		geom.Pt{X: xMax, Y: y},
		BlendCoords(CoordAxes, CoordData),
		ReferenceLineOptions{Color: opt.Color, LineWidth: opt.LineWidth, Dashes: opt.Dashes, Alpha: opt.Alpha},
	)
	a.Add(line)
	return line
}

// AxVLine draws a vertical reference line using a data-space x value and axes-fraction y coordinates.
func (a *Axes) AxVLine(x float64, opts ...VLineOptions) *Segment2D {
	var opt VLineOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	yMin := 0.0
	yMax := 1.0
	if opt.YMin != nil {
		yMin = *opt.YMin
	}
	if opt.YMax != nil {
		yMax = *opt.YMax
	}
	line := a.newSegment(
		geom.Pt{X: x, Y: yMin},
		geom.Pt{X: x, Y: yMax},
		BlendCoords(CoordData, CoordAxes),
		ReferenceLineOptions{Color: opt.Color, LineWidth: opt.LineWidth, Dashes: opt.Dashes, Alpha: opt.Alpha},
	)
	a.Add(line)
	return line
}

// AxLine draws an infinite data-space line through two points, clipped to the current axes view.
func (a *Axes) AxLine(p1, p2 geom.Pt, opts ...ReferenceLineOptions) *InfiniteLine2D {
	var opt ReferenceLineOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	dir := geom.Pt{X: p2.X - p1.X, Y: p2.Y - p1.Y}
	line := a.newInfiniteLine(p1, dir, opt)
	a.Add(line)
	return line
}

// AxLineSlope draws an infinite data-space line through a point with the provided slope.
func (a *Axes) AxLineSlope(point geom.Pt, slope float64, opts ...ReferenceLineOptions) *InfiniteLine2D {
	var opt ReferenceLineOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	line := a.newInfiniteLine(point, geom.Pt{X: 1, Y: slope}, opt)
	a.Add(line)
	return line
}

// AxHSpan draws a horizontal span using axes-fraction x coordinates and data-space y values.
func (a *Axes) AxHSpan(yMin, yMax float64, opts ...HSpanOptions) *Span2D {
	var opt HSpanOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	xMin := 0.0
	xMax := 1.0
	if opt.XMin != nil {
		xMin = *opt.XMin
	}
	if opt.XMax != nil {
		xMax = *opt.XMax
	}
	span := a.newSpan(
		geom.Pt{X: xMin, Y: yMin},
		geom.Pt{X: xMax, Y: yMax},
		BlendCoords(CoordAxes, CoordData),
		SpanOptions{Color: opt.Color, EdgeColor: opt.EdgeColor, EdgeWidth: opt.EdgeWidth, Alpha: opt.Alpha, Antialiased: opt.Antialiased},
	)
	a.Add(span)
	return span
}

// AxVSpan draws a vertical span using data-space x values and axes-fraction y coordinates.
func (a *Axes) AxVSpan(xMin, xMax float64, opts ...VSpanOptions) *Span2D {
	var opt VSpanOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	yMin := 0.0
	yMax := 1.0
	if opt.YMin != nil {
		yMin = *opt.YMin
	}
	if opt.YMax != nil {
		yMax = *opt.YMax
	}
	span := a.newSpan(
		geom.Pt{X: xMin, Y: yMin},
		geom.Pt{X: xMax, Y: yMax},
		BlendCoords(CoordData, CoordAxes),
		SpanOptions{Color: opt.Color, EdgeColor: opt.EdgeColor, EdgeWidth: opt.EdgeWidth, Alpha: opt.Alpha, Antialiased: opt.Antialiased},
	)
	a.Add(span)
	return span
}

// BrokenBarH draws one or more horizontal rectangles sharing a common y-range.
func (a *Axes) BrokenBarH(xRanges [][2]float64, yRange [2]float64, opts ...BarOptions) *Bar2D {
	if len(xRanges) == 0 {
		return nil
	}

	centers := make([]float64, len(xRanges))
	widths := make([]float64, len(xRanges))
	baselines := make([]float64, len(xRanges))
	yCenter := yRange[0] + yRange[1]/2
	for i, xr := range xRanges {
		centers[i] = yCenter
		baselines[i] = xr[0]
		widths[i] = xr[1]
	}

	var opt BarOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	orientation := BarHorizontal
	opt.Orientation = &orientation
	opt.Width = float64Ptr(yRange[1])
	opt.Baselines = baselines

	return a.Bar(centers, widths, opt)
}

// BarLabel adds text labels to bars, either centered inside the bar or at the bar edge.
func (a *Axes) BarLabel(bar *Bar2D, labels []string, opts ...BarLabelOptions) []*Text {
	if a == nil || bar == nil {
		return nil
	}

	var opt BarLabelOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	position := strings.ToLower(strings.TrimSpace(opt.Position))
	if position == "" {
		position = "edge"
	}
	padding := pointsToPixels(a.resolvedRC(), opt.Padding)
	format := opt.Format
	if format == "" {
		format = "%g"
	}

	n := len(bar.X)
	if len(bar.Heights) < n {
		n = len(bar.Heights)
	}

	texts := make([]*Text, 0, n)
	for i := 0; i < n; i++ {
		label := barLabelText(i, labels, barLabelValue(bar, i, position), format)
		if label == "" {
			continue
		}

		textOpt, anchorX, anchorY := barLabelPlacement(bar, i, position, padding)
		textOpt.FontSize = opt.FontSize
		textOpt.Color = opt.Color
		texts = append(texts, a.Text(anchorX, anchorY, label, textOpt))
	}

	return texts
}

// Draw renders the step outline, optionally with a filled baseline patch.
func (s *Stairs2D) Draw(r render.Renderer, ctx *DrawContext) {
	n := s.validCount()
	if n == 0 {
		return
	}

	fillColor, strokeColor := s.resolvedColors()
	if s.Fill && fillColor.A > 0 {
		paint := render.Paint{Fill: fillColor, Snap: render.SnapAuto}
		if s.LineWidth > 0 && strokeColor.A > 0 {
			paint.Stroke = strokeColor
			paint.LineWidth = pointsToPixels(ctx.RC, s.LineWidth)
			paint.LineJoin = render.JoinMiter
			paint.LineCap = render.CapButt
		}
		r.Path(s.fillPath(n, ctx), &paint)
		return
	}

	if s.LineWidth <= 0 || strokeColor.A <= 0 {
		return
	}
	r.Path(s.stepPath(n, ctx), &render.Paint{
		Stroke:    strokeColor,
		LineWidth: pointsToPixels(ctx.RC, s.LineWidth),
		LineJoin:  render.JoinMiter,
		LineCap:   render.CapButt,
		Snap:      render.SnapAuto,
	})
}

// Bounds returns the data bounds of the stairs outline, including the baseline when filled.
func (s *Stairs2D) Bounds(*DrawContext) geom.Rect {
	n := s.validCount()
	if n == 0 {
		return geom.Rect{}
	}

	minX, maxX := s.Edges[0], s.Edges[n]
	if maxX < minX {
		minX, maxX = maxX, minX
	}
	minY, maxY := s.Values[0], s.Values[0]
	for i := 1; i < n; i++ {
		if s.Values[i] < minY {
			minY = s.Values[i]
		}
		if s.Values[i] > maxY {
			maxY = s.Values[i]
		}
	}
	if s.Fill {
		if s.Baseline < minY {
			minY = s.Baseline
		}
		if s.Baseline > maxY {
			maxY = s.Baseline
		}
	}

	return geom.Rect{
		Min: geom.Pt{X: minX, Y: minY},
		Max: geom.Pt{X: maxX, Y: maxY},
	}
}

// Z returns the stairs z-order.
func (s *Stairs2D) Z() float64 { return s.z }

func (s *Stairs2D) legendEntry() (legendEntry, bool) {
	if s == nil || s.Label == "" {
		return legendEntry{}, false
	}
	fillColor, strokeColor := s.resolvedColors()
	if s.Fill {
		return legendEntryFromPatchStyle(s.Label, fillColor, strokeColor, s.LineWidth, "", render.Color{}, 0), true
	}
	return legendEntryFromLine(s.Label, strokeColor, s.LineWidth, nil), true
}

// Draw renders the line segment in its configured coordinate spaces.
func (s *Segment2D) Draw(r render.Renderer, ctx *DrawContext) {
	if s == nil || ctx == nil || s.LineWidth <= 0 || s.Color.A <= 0 {
		return
	}
	path := geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo},
		V: []geom.Pt{
			transformedPoint(ctx, s.Coords, s.Start, 0, 0),
			transformedPoint(ctx, s.Coords, s.End, 0, 0),
		},
	}
	segWidthPx := pointsToPixels(ctx.RC, s.LineWidth)
	r.Path(path, &render.Paint{
		Stroke:    s.Color,
		LineWidth: segWidthPx,
		LineJoin:  render.JoinRound,
		LineCap:   render.CapButt,
		Dashes:    lineDashesForPaint(s.Dashes, segWidthPx, DashUnitsMatplotlib),
		Snap:      render.SnapAuto,
	})
}

// Bounds returns an empty rect so reference helpers do not affect autoscaling.
func (s *Segment2D) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }

// Z returns the line z-order.
func (s *Segment2D) Z() float64 { return s.z }

// Draw renders the span rectangle in its configured coordinate spaces.
func (s *Span2D) Draw(r render.Renderer, ctx *DrawContext) {
	if s == nil || ctx == nil || s.Color.A <= 0 {
		return
	}

	p0 := transformedPoint(ctx, s.Coords, s.Start, 0, 0)
	p1 := transformedPoint(ctx, s.Coords, s.End, 0, 0)
	rect, ok := rectFromPoints(p0, p1)
	if !ok {
		return
	}

	paint := render.Paint{Fill: s.Color, Antialias: s.Antialias}
	if s.EdgeWidth > 0 && s.EdgeColor.A > 0 {
		paint.Stroke = s.EdgeColor
		paint.LineWidth = pointsToPixels(ctx.RC, s.EdgeWidth)
		paint.LineJoin = render.JoinMiter
		paint.LineCap = render.CapButt
	}
	r.Path(snappedFillRectPath(rect), &paint)
}

// Bounds returns an empty rect so spans do not affect autoscaling.
func (s *Span2D) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }

// Z returns the span z-order.
func (s *Span2D) Z() float64 { return s.z }

// Draw renders the clipped infinite line.
func (l *InfiniteLine2D) Draw(r render.Renderer, ctx *DrawContext) {
	if l == nil || ctx == nil || l.LineWidth <= 0 || l.Color.A <= 0 {
		return
	}
	start, end, ok := l.segmentWithin(ctx)
	if !ok {
		return
	}
	lineCap := render.CapSquare
	if len(l.Dashes) > 0 {
		lineCap = render.CapButt
	}
	r.Path(geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo},
		V: []geom.Pt{start, end},
	}, &render.Paint{
		Stroke:    l.Color,
		LineWidth: pointsToPixels(ctx.RC, l.LineWidth),
		LineJoin:  render.JoinRound,
		LineCap:   lineCap,
		Dashes:    lineDashesForPaint(l.Dashes, pointsToPixels(ctx.RC, l.LineWidth), DashUnitsMatplotlib),
		Snap:      render.SnapAuto,
	})
}

// Bounds returns an empty rect so infinite reference lines do not affect autoscaling.
func (l *InfiniteLine2D) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }

// Z returns the infinite-line z-order.
func (l *InfiniteLine2D) Z() float64 { return l.z }

func (s *Stairs2D) validCount() int {
	n := len(s.Values)
	if len(s.Edges)-1 < n {
		n = len(s.Edges) - 1
	}
	if n < 1 {
		return 0
	}
	return n
}

func (s *Stairs2D) resolvedColors() (render.Color, render.Color) {
	fillColor := s.Color
	strokeColor := s.Color
	if s.EdgeColor != (render.Color{}) {
		strokeColor = s.EdgeColor
	}
	if s.Alpha > 0 && s.Alpha <= 1 {
		fillColor = fillColor.WithAlphaMultiplier(s.Alpha)
		strokeColor = strokeColor.WithAlphaMultiplier(s.Alpha)
	}
	return fillColor, strokeColor
}

func (s *Stairs2D) stepPath(n int, ctx *DrawContext) geom.Path {
	path := geom.Path{}
	path.MoveTo(ctx.DataToPixel.Apply(geom.Pt{X: s.Edges[0], Y: s.Values[0]}))
	for i := 0; i < n; i++ {
		path.LineTo(ctx.DataToPixel.Apply(geom.Pt{X: s.Edges[i+1], Y: s.Values[i]}))
		if i+1 < n {
			path.LineTo(ctx.DataToPixel.Apply(geom.Pt{X: s.Edges[i+1], Y: s.Values[i+1]}))
		}
	}
	return path
}

func (s *Stairs2D) fillPath(n int, ctx *DrawContext) geom.Path {
	path := geom.Path{}
	path.MoveTo(ctx.DataToPixel.Apply(geom.Pt{X: s.Edges[0], Y: s.Baseline}))
	path.LineTo(ctx.DataToPixel.Apply(geom.Pt{X: s.Edges[0], Y: s.Values[0]}))
	for i := 0; i < n; i++ {
		path.LineTo(ctx.DataToPixel.Apply(geom.Pt{X: s.Edges[i+1], Y: s.Values[i]}))
		if i+1 < n {
			path.LineTo(ctx.DataToPixel.Apply(geom.Pt{X: s.Edges[i+1], Y: s.Values[i+1]}))
		}
	}
	path.LineTo(ctx.DataToPixel.Apply(geom.Pt{X: s.Edges[n], Y: s.Baseline}))
	return path
}

func (l *InfiniteLine2D) segmentWithin(ctx *DrawContext) (geom.Pt, geom.Pt, bool) {
	xMin, xMax := currentScaleDomain(ctx.DataToPixel.XScale)
	yMin, yMax := currentScaleDomain(ctx.DataToPixel.YScale)
	if xMin > xMax {
		xMin, xMax = xMax, xMin
	}
	if yMin > yMax {
		yMin, yMax = yMax, yMin
	}

	candidates := lineRectCandidates(l.Point, l.Direction, xMin, xMax, yMin, yMax)
	if len(candidates) < 2 {
		return geom.Pt{}, geom.Pt{}, false
	}

	bestAxisX := math.Abs(l.Direction.X) >= math.Abs(l.Direction.Y)
	sortCandidatesAlongDirection(candidates, l.Point, l.Direction, bestAxisX)

	return ctx.DataToPixel.Apply(candidates[0]), ctx.DataToPixel.Apply(candidates[len(candidates)-1]), true
}

func lineRectCandidates(point, direction geom.Pt, xMin, xMax, yMin, yMax float64) []geom.Pt {
	if direction.X == 0 && direction.Y == 0 {
		return nil
	}

	candidates := make([]geom.Pt, 0, 4)
	if direction.X != 0 {
		for _, x := range []float64{xMin, xMax} {
			t := (x - point.X) / direction.X
			y := point.Y + t*direction.Y
			if y >= yMin && y <= yMax {
				candidates = appendUniquePoint(candidates, geom.Pt{X: x, Y: y})
			}
		}
	}
	if direction.Y != 0 {
		for _, y := range []float64{yMin, yMax} {
			t := (y - point.Y) / direction.Y
			x := point.X + t*direction.X
			if x >= xMin && x <= xMax {
				candidates = appendUniquePoint(candidates, geom.Pt{X: x, Y: y})
			}
		}
	}
	return candidates
}

func appendUniquePoint(points []geom.Pt, candidate geom.Pt) []geom.Pt {
	for _, existing := range points {
		if math.Abs(existing.X-candidate.X) < 1e-9 && math.Abs(existing.Y-candidate.Y) < 1e-9 {
			return points
		}
	}
	return append(points, candidate)
}

func sortCandidatesAlongDirection(points []geom.Pt, point, direction geom.Pt, useX bool) {
	if len(points) < 2 {
		return
	}
	sortFn := func(i, j int) bool {
		return lineParam(points[i], point, direction, useX) < lineParam(points[j], point, direction, useX)
	}
	for i := 0; i < len(points)-1; i++ {
		for j := i + 1; j < len(points); j++ {
			if !sortFn(i, j) {
				points[i], points[j] = points[j], points[i]
			}
		}
	}
}

func lineParam(candidate, point, direction geom.Pt, useX bool) float64 {
	if useX && direction.X != 0 {
		return (candidate.X - point.X) / direction.X
	}
	if direction.Y != 0 {
		return (candidate.Y - point.Y) / direction.Y
	}
	return 0
}

func lineDrawStyleFromStepWhere(where StepWhere) LineDrawStyle {
	switch where {
	case StepWhereMid:
		return LineDrawStyleStepsMid
	case StepWherePost:
		return LineDrawStyleStepsPost
	default:
		return LineDrawStyleStepsPre
	}
}

func (a *Axes) newSegment(start, end geom.Pt, coords CoordinateSpec, opt ReferenceLineOptions) *Segment2D {
	color, width := a.referenceLineStyle(opt)
	return &Segment2D{
		Start:     start,
		End:       end,
		Coords:    coords,
		Color:     color,
		LineWidth: width,
		Dashes:    append([]float64(nil), opt.Dashes...),
		z:         40,
	}
}

func (a *Axes) newInfiniteLine(point, direction geom.Pt, opt ReferenceLineOptions) *InfiniteLine2D {
	color, width := a.referenceLineStyle(opt)
	return &InfiniteLine2D{
		Point:     point,
		Direction: direction,
		Color:     color,
		LineWidth: width,
		Dashes:    append([]float64(nil), opt.Dashes...),
		z:         40,
	}
}

func (a *Axes) newSpan(start, end geom.Pt, coords CoordinateSpec, opt SpanOptions) *Span2D {
	rc := a.resolvedRC()
	color := rc.DefaultPatchFaceColor()
	if opt.Color != nil {
		color = *opt.Color
	}
	alpha := 1.0
	if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
		alpha = *opt.Alpha
	}
	color = color.WithAlphaMultiplier(alpha)
	edgeColor := render.Color{}
	if opt.Color != nil {
		// Matplotlib's Patch color= alias sets both facecolor and edgecolor.
		edgeColor = *opt.Color
		edgeColor = edgeColor.WithAlphaMultiplier(alpha)
	} else if rc.Patch.ForceEdgeColor {
		edgeColor = rc.Patch.EdgeColor
	}
	if opt.EdgeColor != nil {
		edgeColor = *opt.EdgeColor
		if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
			edgeColor = edgeColor.WithAlphaMultiplier(*opt.Alpha)
		}
	}
	edgeWidth := rc.Patch.LineWidth // points; converted at the Span2D Paint sink
	if opt.EdgeWidth != nil {
		edgeWidth = *opt.EdgeWidth
	}

	return &Span2D{
		Start:     start,
		End:       end,
		Coords:    coords,
		Color:     color,
		EdgeColor: edgeColor,
		EdgeWidth: edgeWidth,
		Antialias: patchAntialiasMode(&rc.Patch, opt.Antialiased),
		z:         defaultPatchZ,
	}
}

func (a *Axes) referenceLineStyle(opt ReferenceLineOptions) (render.Color, float64) {
	rc := a.resolvedRC()
	color := rc.AxesEdgeColor
	if opt.Color != nil {
		color = *opt.Color
	}
	if color.A == 0 {
		color.A = 1
	}
	if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
		color = color.WithAlphaMultiplier(*opt.Alpha)
	}
	width := rc.AxisLineWidth
	if width <= 0 {
		width = 1
	}
	if opt.LineWidth != nil {
		width = *opt.LineWidth
	}
	return color, width
}

func barLabelText(index int, labels []string, value float64, format string) string {
	if len(labels) > index {
		return labels[index]
	}
	return fmt.Sprintf(format, value)
}

func barLabelValue(bar *Bar2D, index int, position string) float64 {
	if position == "center" {
		return bar.Heights[index]
	}
	return bar.baselineAt(index) + bar.Heights[index]
}

func barLabelPlacement(bar *Bar2D, index int, position string, padding float64) (TextOptions, float64, float64) {
	height := bar.Heights[index]
	baseline := bar.baselineAt(index)

	if bar.Orientation == BarHorizontal {
		anchorY := bar.X[index]
		if position == "center" {
			return TextOptions{
				HAlign: TextAlignCenter,
				VAlign: TextVAlignMiddle,
			}, baseline + height/2, anchorY
		}
		end := baseline + height
		if height >= 0 {
			return TextOptions{
				HAlign:  TextAlignLeft,
				VAlign:  TextVAlignMiddle,
				OffsetX: padding,
			}, end, anchorY
		}
		return TextOptions{
			HAlign:  TextAlignRight,
			VAlign:  TextVAlignMiddle,
			OffsetX: -padding,
		}, end, anchorY
	}

	anchorX := bar.X[index]
	if position == "center" {
		return TextOptions{
			HAlign: TextAlignCenter,
			VAlign: TextVAlignMiddle,
		}, anchorX, baseline + height/2
	}
	end := baseline + height
	if height >= 0 {
		return TextOptions{
			HAlign:  TextAlignCenter,
			VAlign:  TextVAlignBottom,
			OffsetY: padding,
		}, anchorX, end
	}
	return TextOptions{
		HAlign:  TextAlignCenter,
		VAlign:  TextVAlignTop,
		OffsetY: -padding,
	}, anchorX, end
}

func float64Ptr(v float64) *float64 {
	out := v
	return &out
}
