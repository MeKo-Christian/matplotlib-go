package core

import (
	"math"
	"sort"

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
