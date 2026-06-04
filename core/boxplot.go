package core

import (
	"math"
	"sort"
	"strconv"

	matcolor "github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// BoxPlot2D renders a single statistical box plot for one dataset.
type BoxPlot2D struct {
	Data               []float64    // raw sample values
	Position           float64      // x position of the box center in data units
	Width              float64      // box width in data units
	Color              render.Color // box fill color
	EdgeColor          render.Color // box outline color
	MedianColor        render.Color // median line color
	WhiskerColor       render.Color // whisker and cap color
	CapColor           render.Color // whisker cap color
	FlierColor         render.Color // outlier marker color
	FlierEdgeColor     render.Color
	EdgeWidth          float64 // box outline width in pixels
	WhiskerWidth       float64 // whisker/cap line width in pixels
	MedianWidth        float64 // median line width in pixels
	CapWidth           float64 // cap length in data units
	FlierSize          float64 // outlier marker size in points
	FlierEdgeWidth     float64
	Alpha              float64 // alpha transparency (0-1, 0 means 1.0)
	ShowFliers         bool    // whether to draw outliers
	Notch              bool    // whether to draw a notched median confidence interval
	Bootstrap          int     // stored for API parity; deterministic CI fallback is used
	ConfidenceInterval *[2]float64
	CustomMedian       *float64
	WhiskerPercentiles *[2]float64
	FlierMarker        MarkerType
	Label              string  // series label for legend
	z                  float64 // z-order

	computed bool
	hasData  bool
	stats    boxPlotStats
}

type boxPlotStats struct {
	min          float64
	max          float64
	q1           float64
	median       float64
	q3           float64
	lowerWhisker float64
	upperWhisker float64
	ciLow        float64
	ciHigh       float64
	outliers     []float64
}

// BxpStat contains precomputed statistics for Axes.Bxp.
type BxpStat struct {
	Med    float64
	Q1     float64
	Q3     float64
	Whislo float64
	Whishi float64

	Mean   *float64
	Cilo   *float64
	Cihi   *float64
	Fliers []float64
	Label  string
}

// BxpOptions configures Axes.Bxp.
type BxpOptions struct {
	Positions   []float64
	Widths      []float64
	CapWidths   []float64
	Orientation string

	ShowNotches *bool
	ShowMeans   *bool
	ShowCaps    *bool
	ShowBox     *bool
	ShowFliers  *bool
	MeanLine    bool
	ManageTicks *bool

	Color       *render.Color
	MedianColor *render.Color
	MeanColor   *render.Color
	FlierColor  *render.Color
	LineWidth   *float64
	MarkerSize  *float64

	Label  string
	Labels []string
}

// BxpContainer groups the Line2D artists created by Axes.Bxp.
type BxpContainer struct {
	Whiskers []*Line2D
	Caps     []*Line2D
	Boxes    []*Line2D
	Medians  []*Line2D
	Fliers   []*Line2D
	Means    []*Line2D
}

// Bxp draws box plots from precomputed statistics, matching Matplotlib's
// low-level Axes.bxp surface while returning typed artist groups.
func (a *Axes) Bxp(stats []BxpStat, opts ...BxpOptions) *BxpContainer {
	if a == nil || len(stats) == 0 {
		return nil
	}
	var opt BxpOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	n := len(stats)
	if !validOptionalList(opt.Positions, n) || !validOptionalScalarList(opt.Widths, n) || !validOptionalScalarList(opt.CapWidths, n) {
		return nil
	}
	if len(opt.Labels) > 0 && len(opt.Labels) != n {
		return nil
	}
	orientation := normalizeViolinOrientation(opt.Orientation)
	positions := expandFloatOption(opt.Positions, n, func(i int) float64 { return float64(i + 1) })
	widths := expandFloatOption(opt.Widths, n, func(int) float64 {
		return matplotlibBoxPlotDefaultWidth(n, positions)
	})
	capWidths := expandFloatOption(opt.CapWidths, n, func(i int) float64 {
		return math.Abs(widths[i]) * 0.5
	})

	showNotches := specialtyBool(opt.ShowNotches, false)
	showMeans := specialtyBool(opt.ShowMeans, false)
	showCaps := specialtyBool(opt.ShowCaps, true)
	showBox := specialtyBool(opt.ShowBox, true)
	showFliers := specialtyBool(opt.ShowFliers, true)
	manageTicks := specialtyBool(opt.ManageTicks, true)

	color := render.Color{R: 0, G: 0, B: 0, A: 1}
	if opt.Color != nil {
		color = *opt.Color
	}
	medianColor := matcolor.Tab10[1]
	if opt.MedianColor != nil {
		medianColor = *opt.MedianColor
	}
	meanColor := matcolor.Tab10[2]
	if opt.MeanColor != nil {
		meanColor = *opt.MeanColor
	}
	flierColor := color
	if opt.FlierColor != nil {
		flierColor = *opt.FlierColor
	}
	lineWidth := 1.0
	if opt.LineWidth != nil && *opt.LineWidth > 0 {
		lineWidth = *opt.LineWidth
	}
	markerSize := 3.5
	if opt.MarkerSize != nil && *opt.MarkerSize > 0 {
		markerSize = *opt.MarkerSize
	}

	container := &BxpContainer{}
	tickLabels := make([]string, n)
	for i, stat := range stats {
		if !validBxpStat(stat) {
			return nil
		}
		pos := positions[i]
		width := math.Abs(widths[i])
		capWidth := math.Abs(capWidths[i])
		left := pos - width*0.5
		right := pos + width*0.5
		capLeft := pos - capWidth*0.5
		capRight := pos + capWidth*0.5

		tickLabels[i] = stat.Label
		if tickLabels[i] == "" {
			tickLabels[i] = trimFloatLabel(pos)
		}

		if showBox {
			boxPoints := bxpBoxPoints(stat, pos, left, right, showNotches, orientation)
			container.Boxes = append(container.Boxes, a.addBxpLine(boxPoints, color, lineWidth, ""))
		}
		container.Whiskers = append(container.Whiskers,
			a.addBxpLine([]geom.Pt{violinPoint(pos, stat.Q1, orientation), violinPoint(pos, stat.Whislo, orientation)}, color, lineWidth, ""),
			a.addBxpLine([]geom.Pt{violinPoint(pos, stat.Q3, orientation), violinPoint(pos, stat.Whishi, orientation)}, color, lineWidth, ""),
		)
		if showCaps {
			container.Caps = append(container.Caps,
				a.addBxpLine([]geom.Pt{violinPoint(capLeft, stat.Whislo, orientation), violinPoint(capRight, stat.Whislo, orientation)}, color, lineWidth, ""),
				a.addBxpLine([]geom.Pt{violinPoint(capLeft, stat.Whishi, orientation), violinPoint(capRight, stat.Whishi, orientation)}, color, lineWidth, ""),
			)
		}
		medianLabel := ""
		if len(opt.Labels) > 0 {
			medianLabel = opt.Labels[i]
		} else if opt.Label != "" && i == 0 {
			medianLabel = opt.Label
		}
		medLeft, medRight := left, right
		if showNotches {
			medLeft, medRight = pos-width*0.25, pos+width*0.25
		}
		container.Medians = append(container.Medians, a.addBxpLine(
			[]geom.Pt{violinPoint(medLeft, stat.Med, orientation), violinPoint(medRight, stat.Med, orientation)},
			medianColor, lineWidth, medianLabel,
		))
		if showMeans && stat.Mean != nil && isFinite(*stat.Mean) {
			if opt.MeanLine {
				container.Means = append(container.Means, a.addBxpLine(
					[]geom.Pt{violinPoint(left, *stat.Mean, orientation), violinPoint(right, *stat.Mean, orientation)},
					meanColor, lineWidth, "",
				))
			} else {
				container.Means = append(container.Means, a.addBxpMarker(violinPoint(pos, *stat.Mean, orientation), meanColor, markerSize))
			}
		}
		if showFliers && len(stat.Fliers) > 0 {
			points := make([]geom.Pt, 0, len(stat.Fliers))
			for _, flier := range stat.Fliers {
				if isFinite(flier) {
					points = append(points, violinPoint(pos, flier, orientation))
				}
			}
			if len(points) > 0 {
				container.Fliers = append(container.Fliers, a.addBxpMarkers(points, flierColor, markerSize))
			}
		}
	}
	if manageTicks {
		if orientation == "horizontal" {
			a.YAxis.Locator = FixedLocator{TicksList: positions}
			a.YAxis.Formatter = FixedFormatter{Labels: tickLabels}
		} else {
			a.XAxis.Locator = FixedLocator{TicksList: positions}
			a.XAxis.Formatter = FixedFormatter{Labels: tickLabels}
		}
	}
	return container
}

func (a *Axes) addBxpLine(points []geom.Pt, color render.Color, width float64, label string) *Line2D {
	line := &Line2D{XY: points, Col: color, W: width, Label: label, z: 2}
	a.Add(line)
	return line
}

func (a *Axes) addBxpMarker(point geom.Pt, color render.Color, size float64) *Line2D {
	return a.addBxpMarkers([]geom.Pt{point}, color, size)
}

func (a *Axes) addBxpMarkers(points []geom.Pt, color render.Color, size float64) *Line2D {
	line := &Line2D{
		XY:              points,
		Col:             color,
		Marker:          MarkerCircle,
		MarkerSet:       true,
		MarkerSize:      size,
		MarkerFaceColor: color,
		MarkerEdgeColor: color,
		z:               2.1,
	}
	a.Add(line)
	return line
}

func bxpBoxPoints(stat BxpStat, pos, left, right float64, notched bool, orientation string) []geom.Pt {
	if !notched {
		return []geom.Pt{
			violinPoint(left, stat.Q1, orientation),
			violinPoint(right, stat.Q1, orientation),
			violinPoint(right, stat.Q3, orientation),
			violinPoint(left, stat.Q3, orientation),
			violinPoint(left, stat.Q1, orientation),
		}
	}
	cilo, cihi := stat.Med, stat.Med
	if stat.Cilo != nil {
		cilo = *stat.Cilo
	}
	if stat.Cihi != nil {
		cihi = *stat.Cihi
	}
	notchLeft := pos - (right-left)*0.25
	notchRight := pos + (right-left)*0.25
	return []geom.Pt{
		violinPoint(left, stat.Q1, orientation),
		violinPoint(right, stat.Q1, orientation),
		violinPoint(right, cilo, orientation),
		violinPoint(notchRight, stat.Med, orientation),
		violinPoint(right, cihi, orientation),
		violinPoint(right, stat.Q3, orientation),
		violinPoint(left, stat.Q3, orientation),
		violinPoint(left, cihi, orientation),
		violinPoint(notchLeft, stat.Med, orientation),
		violinPoint(left, cilo, orientation),
		violinPoint(left, stat.Q1, orientation),
	}
}

func validBxpStat(stat BxpStat) bool {
	return isFinite(stat.Med) && isFinite(stat.Q1) && isFinite(stat.Q3) && isFinite(stat.Whislo) && isFinite(stat.Whishi)
}

func validOptionalScalarList(values []float64, n int) bool {
	return len(values) == 0 || len(values) == 1 || len(values) == n
}

func validOptionalList(values []float64, n int) bool {
	return len(values) == 0 || len(values) == n
}

func expandFloatOption(values []float64, n int, fallback func(int) float64) []float64 {
	out := make([]float64, n)
	for i := range out {
		switch {
		case len(values) == 1:
			out[i] = values[0]
		case len(values) == n:
			out[i] = values[i]
		default:
			out[i] = fallback(i)
		}
	}
	return out
}

func trimFloatLabel(value float64) string {
	return strconv.FormatFloat(value, 'g', -1, 64)
}

func (b *BoxPlot2D) compute() {
	if b.computed {
		return
	}
	b.computed = true

	finite := make([]float64, 0, len(b.Data))
	for _, v := range b.Data {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			continue
		}
		finite = append(finite, v)
	}
	if len(finite) == 0 {
		return
	}

	sort.Float64s(finite)
	b.stats = b.computeBoxPlotStats(finite)
	b.hasData = true
}

func computeBoxPlotStats(sorted []float64) boxPlotStats {
	stats := boxPlotStats{
		min:    sorted[0],
		max:    sorted[len(sorted)-1],
		q1:     percentileSorted(sorted, 25),
		median: percentileSorted(sorted, 50),
		q3:     percentileSorted(sorted, 75),
	}

	iqr := stats.q3 - stats.q1
	lowerFence := stats.q1 - 1.5*iqr
	upperFence := stats.q3 + 1.5*iqr

	stats.lowerWhisker = stats.q1
	for _, v := range sorted {
		if v >= lowerFence {
			stats.lowerWhisker = v
			break
		}
	}

	stats.upperWhisker = stats.q3
	for i := len(sorted) - 1; i >= 0; i-- {
		if sorted[i] <= upperFence {
			stats.upperWhisker = sorted[i]
			break
		}
	}

	for _, v := range sorted {
		if v < lowerFence || v > upperFence {
			stats.outliers = append(stats.outliers, v)
		}
	}

	return stats
}

func (b *BoxPlot2D) computeBoxPlotStats(sorted []float64) boxPlotStats {
	stats := computeBoxPlotStats(sorted)
	if b.CustomMedian != nil && isFinite(*b.CustomMedian) {
		stats.median = *b.CustomMedian
	}
	if b.WhiskerPercentiles != nil {
		lo := math.Min(b.WhiskerPercentiles[0], b.WhiskerPercentiles[1])
		hi := math.Max(b.WhiskerPercentiles[0], b.WhiskerPercentiles[1])
		lo = math.Max(0, math.Min(100, lo))
		hi = math.Max(0, math.Min(100, hi))
		lowerFence := percentileSorted(sorted, lo)
		upperFence := percentileSorted(sorted, hi)
		stats.lowerWhisker = stats.q1
		for _, v := range sorted {
			if v >= lowerFence {
				stats.lowerWhisker = v
				break
			}
		}
		stats.upperWhisker = stats.q3
		for i := len(sorted) - 1; i >= 0; i-- {
			if sorted[i] <= upperFence {
				stats.upperWhisker = sorted[i]
				break
			}
		}
		stats.outliers = stats.outliers[:0]
		for _, v := range sorted {
			if v < stats.lowerWhisker || v > stats.upperWhisker {
				stats.outliers = append(stats.outliers, v)
			}
		}
	}
	stats.ciLow, stats.ciHigh = boxPlotMedianCI(sorted, stats)
	if b.ConfidenceInterval != nil && isFinite(b.ConfidenceInterval[0]) && isFinite(b.ConfidenceInterval[1]) {
		stats.ciLow = math.Min(b.ConfidenceInterval[0], b.ConfidenceInterval[1])
		stats.ciHigh = math.Max(b.ConfidenceInterval[0], b.ConfidenceInterval[1])
	}
	return stats
}

func boxPlotMedianCI(sorted []float64, stats boxPlotStats) (float64, float64) {
	iqr := stats.q3 - stats.q1
	if len(sorted) == 0 || iqr <= 0 || !isFinite(iqr) {
		return stats.median, stats.median
	}
	delta := 1.57 * iqr / math.Sqrt(float64(len(sorted)))
	return stats.median - delta, stats.median + delta
}

func percentileSorted(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return math.NaN()
	}
	if len(sorted) == 1 {
		return sorted[0]
	}
	if p <= 0 {
		return sorted[0]
	}
	if p >= 100 {
		return sorted[len(sorted)-1]
	}

	pos := (p / 100) * float64(len(sorted)-1)
	lo := int(math.Floor(pos))
	hi := int(math.Ceil(pos))
	if lo == hi {
		return sorted[lo]
	}

	frac := pos - float64(lo)
	return sorted[lo]*(1-frac) + sorted[hi]*frac
}

func (b *BoxPlot2D) ensureComputed() {
	b.compute()
}

func (b *BoxPlot2D) Draw(r render.Renderer, ctx *DrawContext) {
	if len(b.Data) == 0 {
		return
	}
	b.ensureComputed()
	if !b.hasData {
		return
	}

	boxWidth := b.Width
	if boxWidth <= 0 {
		boxWidth = 0.6
	}
	boxWidth = math.Abs(boxWidth)

	capWidth := b.CapWidth
	if capWidth <= 0 {
		capWidth = boxWidth * 0.5
	}
	capWidth = math.Abs(capWidth)

	edgeWidth := b.EdgeWidth
	if edgeWidth <= 0 {
		edgeWidth = 1.0
	}
	whiskerWidth := b.WhiskerWidth
	if whiskerWidth <= 0 {
		whiskerWidth = edgeWidth
	}
	medianWidth := b.MedianWidth
	if medianWidth <= 0 {
		medianWidth = math.Max(edgeWidth, 1.5)
	}
	flierSize := b.FlierSize
	if flierSize <= 0 {
		flierSize = 3.5
	}
	flierEdgeWidth := b.FlierEdgeWidth
	if flierEdgeWidth <= 0 {
		flierEdgeWidth = math.Max(1.0, whiskerWidth*0.6)
	}
	alpha := b.Alpha
	if alpha <= 0 {
		alpha = 1.0
	}
	if alpha > 1 {
		alpha = 1.0
	}

	boxColor := applyAlpha(b.Color, alpha)
	edgeColor := applyAlpha(b.EdgeColor, alpha)
	medianColor := applyAlpha(b.MedianColor, alpha)
	whiskerColor := applyAlpha(b.WhiskerColor, alpha)
	capColor := applyAlpha(b.CapColor, alpha)
	flierColor := applyAlpha(b.FlierColor, alpha)
	flierEdgeColor := applyAlpha(b.FlierEdgeColor, alpha)

	xLeft := b.Position - boxWidth/2
	xRight := b.Position + boxWidth/2

	boxPath := b.boxPath(ctx, xLeft, xRight)
	if len(boxPath.C) > 0 {
		paint := render.Paint{
			Fill:     boxColor,
			LineJoin: render.JoinMiter,
			LineCap:  render.CapButt,
			Snap:     render.SnapAuto,
		}
		if edgeWidth > 0 && edgeColor.A > 0 {
			paint.Stroke = edgeColor
			paint.LineWidth = edgeWidth
		}
		r.Path(boxPath, &paint)
	}

	if whiskerWidth > 0 && whiskerColor.A > 0 {
		whiskerPaint := render.Paint{
			Stroke:    whiskerColor,
			LineWidth: whiskerWidth,
			LineJoin:  render.JoinMiter,
			LineCap:   render.CapButt,
			Snap:      render.SnapAuto,
		}
		r.Path(linePath(ctx, geom.Pt{X: b.Position, Y: b.stats.lowerWhisker}, geom.Pt{X: b.Position, Y: b.stats.q1}), &whiskerPaint)
		r.Path(linePath(ctx, geom.Pt{X: b.Position, Y: b.stats.q3}, geom.Pt{X: b.Position, Y: b.stats.upperWhisker}), &whiskerPaint)

		capPaint := render.Paint{
			Stroke:    capColor,
			LineWidth: whiskerWidth,
			LineJoin:  render.JoinMiter,
			LineCap:   render.CapButt,
			Snap:      render.SnapAuto,
		}
		capLeft := b.Position - capWidth/2
		capRight := b.Position + capWidth/2
		r.Path(linePath(ctx, geom.Pt{X: capLeft, Y: b.stats.lowerWhisker}, geom.Pt{X: capRight, Y: b.stats.lowerWhisker}), &capPaint)
		r.Path(linePath(ctx, geom.Pt{X: capLeft, Y: b.stats.upperWhisker}, geom.Pt{X: capRight, Y: b.stats.upperWhisker}), &capPaint)
	}

	if medianWidth > 0 && medianColor.A > 0 {
		medianPaint := render.Paint{
			Stroke:    medianColor,
			LineWidth: medianWidth,
			LineJoin:  render.JoinMiter,
			LineCap:   render.CapButt,
			Snap:      render.SnapAuto,
		}
		medianLeft, medianRight := xLeft, xRight
		if b.Notch {
			notchInset := (xRight - xLeft) * 0.25
			medianLeft = b.Position - notchInset
			medianRight = b.Position + notchInset
		}
		r.Path(linePath(ctx, geom.Pt{X: medianLeft, Y: b.stats.median}, geom.Pt{X: medianRight, Y: b.stats.median}), &medianPaint)
	}

	if b.ShowFliers {
		if flierColor.A <= 0 && flierEdgeColor.A <= 0 {
			return
		}
		flierPaint := render.Paint{
			Fill:      flierColor,
			Stroke:    flierEdgeColor,
			LineWidth: flierEdgeWidth,
			LineJoin:  render.JoinRound,
			LineCap:   render.CapRound,
			Snap:      render.SnapAuto,
		}
		marker := b.FlierMarker
		if marker == 0 {
			marker = MarkerCircle
		}
		scatter := Scatter2D{Marker: marker}
		flierSizePx := pointsToPixels(ctx.RC, flierSize)
		for _, v := range b.stats.outliers {
			pt := ctx.DataToPixel.Apply(geom.Pt{X: b.Position, Y: v})
			r.Path(scaleAndTranslatePath(scatter.markerPrototypePath(), flierSizePx, pt), &flierPaint)
		}
	}
}

func (b *BoxPlot2D) boxPath(ctx *DrawContext, xLeft, xRight float64) geom.Path {
	if !b.Notch {
		return rectPath(ctx, geom.Pt{X: xLeft, Y: b.stats.q1}, geom.Pt{X: xRight, Y: b.stats.q3})
	}
	xMid := b.Position
	notchInset := (xRight - xLeft) * 0.25
	ciLow := math.Max(b.stats.q1, math.Min(b.stats.q3, b.stats.ciLow))
	ciHigh := math.Max(b.stats.q1, math.Min(b.stats.q3, b.stats.ciHigh))
	points := []geom.Pt{
		{X: xLeft, Y: b.stats.q1},
		{X: xRight, Y: b.stats.q1},
		{X: xRight, Y: ciLow},
		{X: xMid + notchInset, Y: b.stats.median},
		{X: xRight, Y: ciHigh},
		{X: xRight, Y: b.stats.q3},
		{X: xLeft, Y: b.stats.q3},
		{X: xLeft, Y: ciHigh},
		{X: xMid - notchInset, Y: b.stats.median},
		{X: xLeft, Y: ciLow},
	}
	return polygonDisplayPath(ctx, points, true)
}

func polygonDisplayPath(ctx *DrawContext, points []geom.Pt, closed bool) geom.Path {
	path := geom.Path{}
	for i, pt := range points {
		if i == 0 {
			path.C = append(path.C, geom.MoveTo)
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, ctx.DataToPixel.Apply(pt))
	}
	if closed && len(points) > 0 {
		path.C = append(path.C, geom.ClosePath)
	}
	return path
}

func applyAlpha(c render.Color, alpha float64) render.Color {
	if alpha <= 0 {
		alpha = 1
	}
	if alpha > 1 {
		alpha = 1
	}
	c.A *= alpha
	return c
}

func linePath(ctx *DrawContext, p1, p2 geom.Pt) geom.Path {
	path := geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo},
		V: []geom.Pt{
			ctx.DataToPixel.Apply(p1),
			ctx.DataToPixel.Apply(p2),
		},
	}
	return path
}

func rectPath(ctx *DrawContext, minPt, maxPt geom.Pt) geom.Path {
	corners := []geom.Pt{
		{X: minPt.X, Y: minPt.Y},
		{X: maxPt.X, Y: minPt.Y},
		{X: maxPt.X, Y: maxPt.Y},
		{X: minPt.X, Y: maxPt.Y},
	}
	path := geom.Path{}
	for i, corner := range corners {
		if i == 0 {
			path.C = append(path.C, geom.MoveTo)
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, ctx.DataToPixel.Apply(corner))
	}
	path.C = append(path.C, geom.ClosePath)
	return path
}

func circlePath(center geom.Pt, radius float64) geom.Path {
	if radius <= 0 {
		return geom.Path{}
	}

	const segments = 16
	path := geom.Path{}
	for i := 0; i < segments; i++ {
		angle := 2 * math.Pi * float64(i) / segments
		x := center.X + radius*math.Cos(angle)
		y := center.Y + radius*math.Sin(angle)
		if i == 0 {
			path.C = append(path.C, geom.MoveTo)
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, geom.Pt{X: x, Y: y})
	}
	path.C = append(path.C, geom.ClosePath)
	return path
}

func (b *BoxPlot2D) Z() float64 {
	return b.z
}

func (b *BoxPlot2D) Bounds(_ *DrawContext) geom.Rect {
	if len(b.Data) == 0 {
		return geom.Rect{}
	}
	b.ensureComputed()
	if !b.hasData {
		return geom.Rect{}
	}

	boxWidth := b.Width
	if boxWidth <= 0 {
		boxWidth = 0.6
	}
	boxWidth = math.Abs(boxWidth)
	capWidth := b.CapWidth
	if capWidth <= 0 {
		capWidth = boxWidth * 0.5
	}
	capWidth = math.Abs(capWidth)
	halfSpan := math.Max(boxWidth, capWidth) / 2

	return geom.Rect{
		Min: geom.Pt{X: b.Position - halfSpan, Y: b.stats.min},
		Max: geom.Pt{X: b.Position + halfSpan, Y: b.stats.max},
	}
}
