package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func (l *Legend) titleLayout(r render.Renderer, ctx *DrawContext, fallbackFontSize float64) (singleLineTextLayout, float64, float64) {
	if l == nil || l.Title == "" {
		return singleLineTextLayout{}, 0, fallbackFontSize
	}
	fontSize := l.TitleFontSize
	if fontSize <= 0 {
		fontSize = fallbackFontSize
	}
	layout := measureSingleLineTextLayout(r, l.Title, fontSize, ctx.RC.FontKey, ctx.RC.UseTeX)
	return layout, legendRowHeight(layout, fontSize, ctx), fontSize
}

type legendLayout struct {
	Columns       []legendColumnLayout
	Width         float64
	Height        float64
	ColumnSpacing float64
}

type legendColumnLayout struct {
	Start int
	End   int
	Width float64
}

func (l *Legend) layoutEntries(labelLayouts []singleLineTextLayout, rowHeights []float64, ctx *DrawContext, fontSize float64) legendLayout {
	count := len(labelLayouts)
	columns := l.effectiveNumColumns(count)
	if columns <= 0 {
		return legendLayout{}
	}
	columnSpacing := l.ColumnSpacing
	if columnSpacing <= 0 && !l.defaultsSet && ctx != nil {
		columnSpacing = 2.0 * pointsToPixels(ctx.RC, fontSize)
	}
	layout := legendLayout{
		Columns:       make([]legendColumnLayout, 0, columns),
		ColumnSpacing: columnSpacing,
	}
	start := 0
	for col := 0; col < columns; col++ {
		size := count / columns
		if col < count%columns {
			size++
		}
		if size <= 0 {
			continue
		}
		end := start + size
		maxLabelWidth := 0.0
		columnHeight := 0.0
		for i := start; i < end; i++ {
			labelWidth := legendLabelWidth(labelLayouts[i])
			if labelWidth > maxLabelWidth {
				maxLabelWidth = labelWidth
			}
			columnHeight += rowHeights[i]
		}
		if size > 1 {
			columnHeight += l.RowGap * float64(size-1)
		}
		columnWidth := l.SampleWidth + l.SampleTextGap + maxLabelWidth
		layout.Columns = append(layout.Columns, legendColumnLayout{Start: start, End: end, Width: columnWidth})
		layout.Width += columnWidth
		if columnHeight > layout.Height {
			layout.Height = columnHeight
		}
		start = end
	}
	if len(layout.Columns) > 1 {
		layout.Width += columnSpacing * float64(len(layout.Columns)-1)
	}
	return layout
}

func (l *Legend) handleMetrics(fontPx float64) (height, descent float64) {
	handleHeight := l.HandleHeight
	if handleHeight <= 0 && !l.defaultsSet {
		handleHeight = 0.7
	}
	descent = 0.35 * fontPx * (handleHeight - 0.7)
	height = fontPx*handleHeight - descent
	return height, descent
}

func (l *Legend) effectiveNumColumns(entryCount int) int {
	if entryCount <= 0 {
		return 0
	}
	columns := l.NumColumns
	if columns <= 0 {
		columns = 1
	}
	if columns > entryCount {
		columns = entryCount
	}
	return columns
}

func legendRowHeight(layout singleLineTextLayout, fontSize float64, ctx *DrawContext) float64 {
	fontPx := pointsToPixels(ctx.RC, fontSize)
	rowHeight := layout.RunAscent + layout.RunDescent
	if layout.MathLayout != nil && rowHeight < fontPx*1.28 {
		rowHeight = fontPx * 1.28
	}
	if rowHeight < fontPx {
		rowHeight = fontPx
	}
	if rowHeight <= 0 {
		rowHeight = layout.Height
	}
	return rowHeight
}

func legendLabelWidth(layout singleLineTextLayout) float64 {
	return layout.Width
}

func (l *Legend) legendBoxRect(ctx *DrawContext, width, height float64) geom.Rect {
	if ctx == nil {
		return geom.Rect{}
	}
	if l.Location == LegendBest && l.Locator == nil && l.Axes != nil {
		return l.bestLegendBoxRect(ctx, width, height)
	}
	return resolveAnchoredBoxRect(l.Locator, ctx.Clip, width, height, l.Location, l.Inset)
}

func (l *Legend) bestLegendBoxRect(ctx *DrawContext, width, height float64) geom.Rect {
	candidates := []LegendLocation{
		LegendUpperRight,
		LegendUpperLeft,
		LegendLowerLeft,
		LegendLowerRight,
		LegendRight,
		LegendCenterLeft,
		LegendCenterRight,
		LegendLowerCenter,
		LegendUpperCenter,
		LegendCenter,
	}
	data := l.legendAvoidanceData(ctx)

	best := anchoredBoxRect(ctx.Clip, width, height, candidates[0], l.Inset)
	bestBadness := legendPlacementBadness(best, data)
	if bestBadness == 0 {
		return best
	}

	for _, location := range candidates[1:] {
		box := anchoredBoxRect(ctx.Clip, width, height, location, l.Inset)
		badness := legendPlacementBadness(box, data)
		if badness == 0 {
			return box
		}
		if badness < bestBadness {
			best = box
			bestBadness = badness
		}
	}
	return best
}

type legendAvoidanceData struct {
	points []geom.Pt
	lines  []geom.Path
	boxes  []geom.Rect
}

func (l *Legend) legendAvoidanceData(ctx *DrawContext) legendAvoidanceData {
	if l == nil || l.Axes == nil || ctx == nil {
		return legendAvoidanceData{}
	}
	data := legendAvoidanceData{}
	appendPoints := func(spec CoordinateSpec, pts []geom.Pt) {
		tr := ctx.TransformFor(spec)
		for _, pt := range pts {
			if tr != nil {
				pt = tr.Apply(pt)
			}
			data.points = append(data.points, pt)
		}
	}
	appendRect := func(spec CoordinateSpec, rect geom.Rect) {
		if rect == (geom.Rect{}) {
			return
		}
		tr := ctx.TransformFor(spec)
		if tr == nil {
			data.boxes = append(data.boxes, rect)
			return
		}
		min := tr.Apply(rect.Min)
		max := tr.Apply(rect.Max)
		data.boxes = append(data.boxes, geom.Rect{
			Min: geom.Pt{X: math.Min(min.X, max.X), Y: math.Min(min.Y, max.Y)},
			Max: geom.Pt{X: math.Max(min.X, max.X), Y: math.Max(min.Y, max.Y)},
		})
	}
	appendLine := func(path geom.Path) {
		if len(path.V) == 0 {
			return
		}
		data.lines = append(data.lines, path)
		data.points = append(data.points, path.V...)
	}

	for _, art := range l.Axes.Artists {
		switch a := art.(type) {
		case *Legend:
			continue
		case *Scatter2D:
			appendPoints(Coords(CoordData), a.XY)
		case *Line2D:
			appendLine(a.displayPath(ctx))
		case *PathCollection:
			appendPoints(a.Coords, a.Offsets)
		case *LineCollection:
			for _, segment := range a.Segments {
				appendLine(displayPolylineForLegend(ctx, a.Coords, segment))
			}
		case *Rectangle:
			appendRect(a.Coords, a.Bounds(ctx))
		case *PathPatch:
			appendLine(buildArtistDisplayPath(ctx, a, a.Coords, a.Path, geom.Identity()))
		case *Image2D:
			appendRect(Coords(CoordData), a.Bounds(ctx))
		case *Text:
			data.points = append(data.points, transformedPoint(ctx, a.Coords, a.Position, a.OffsetX, a.OffsetY))
		case *Annotation:
			target := transformedPoint(ctx, a.Coords, a.Point, 0, 0)
			text := a.textAnchor(ctx)
			data.points = append(data.points, target, text)
		case *AnnotationBbox:
			target := transformedPoint(ctx, a.XYCoords, a.Point, 0, 0)
			box := target
			if a.BoxPosition != nil {
				box = transformedPoint(ctx, a.BoxCoords, *a.BoxPosition, 0, 0)
			}
			data.points = append(data.points, target, box)
		}
	}
	return data
}

func displayPolylineForLegend(ctx *DrawContext, spec CoordinateSpec, pts []geom.Pt) geom.Path {
	tr := ctx.TransformFor(spec)
	path := geom.Path{}
	inSegment := false
	for _, pt := range pts {
		if !finitePoint(pt) {
			inSegment = false
			continue
		}
		if tr != nil {
			pt = tr.Apply(pt)
		}
		if !finitePoint(pt) {
			inSegment = false
			continue
		}
		if !inSegment {
			path.C = append(path.C, geom.MoveTo)
			inSegment = true
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, pt)
	}
	return path
}

func legendPlacementBadness(box geom.Rect, data legendAvoidanceData) int {
	badness := 0
	for _, pt := range data.points {
		if pointInRect(pt, box) {
			badness++
		}
	}
	for _, other := range data.boxes {
		if rectsOverlap(box, other) {
			badness++
		}
	}
	for _, line := range data.lines {
		if pathIntersectsRect(line, box) {
			badness++
		}
	}
	return badness
}

func pointInRect(pt geom.Pt, rect geom.Rect) bool {
	return pt.X >= rect.Min.X && pt.X <= rect.Max.X && pt.Y >= rect.Min.Y && pt.Y <= rect.Max.Y
}

func rectsOverlap(a, b geom.Rect) bool {
	return a.Min.X < b.Max.X && a.Max.X > b.Min.X && a.Min.Y < b.Max.Y && a.Max.Y > b.Min.Y
}

func pathIntersectsRect(path geom.Path, rect geom.Rect) bool {
	var cur geom.Pt
	haveCur := false
	vi := 0
	for _, cmd := range path.C {
		switch cmd {
		case geom.MoveTo:
			if vi >= len(path.V) {
				return false
			}
			cur = path.V[vi]
			haveCur = true
			vi++
		case geom.LineTo:
			if vi >= len(path.V) {
				return false
			}
			next := path.V[vi]
			if haveCur && segmentIntersectsRect(cur, next, rect) {
				return true
			}
			cur = next
			haveCur = true
			vi++
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
	return false
}

func segmentIntersectsRect(a, b geom.Pt, rect geom.Rect) bool {
	corners := []geom.Pt{
		rect.Min,
		{X: rect.Max.X, Y: rect.Min.Y},
		rect.Max,
		{X: rect.Min.X, Y: rect.Max.Y},
	}
	for i := range corners {
		if segmentsIntersect(a, b, corners[i], corners[(i+1)%len(corners)]) {
			return true
		}
	}
	return false
}

func segmentsIntersect(a, b, c, d geom.Pt) bool {
	const eps = 1e-9
	orient := func(p, q, r geom.Pt) float64 {
		return (q.X-p.X)*(r.Y-p.Y) - (q.Y-p.Y)*(r.X-p.X)
	}
	onSegment := func(p, q, r geom.Pt) bool {
		return math.Min(p.X, r.X)-eps <= q.X && q.X <= math.Max(p.X, r.X)+eps &&
			math.Min(p.Y, r.Y)-eps <= q.Y && q.Y <= math.Max(p.Y, r.Y)+eps
	}
	o1 := orient(a, b, c)
	o2 := orient(a, b, d)
	o3 := orient(c, d, a)
	o4 := orient(c, d, b)
	if math.Abs(o1) <= eps && onSegment(a, c, b) {
		return true
	}
	if math.Abs(o2) <= eps && onSegment(a, d, b) {
		return true
	}
	if math.Abs(o3) <= eps && onSegment(c, a, d) {
		return true
	}
	if math.Abs(o4) <= eps && onSegment(c, b, d) {
		return true
	}
	return (o1 > 0) != (o2 > 0) && (o3 > 0) != (o4 > 0)
}
