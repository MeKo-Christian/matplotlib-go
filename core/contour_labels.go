package core

import (
	"math"
	"sort"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func contourLabels(polylines [][]geom.Pt, levels []float64, colors []render.Color, formatter Formatter) []contourLabel {
	type candidate struct {
		polyline []geom.Pt
		color    render.Color
	}
	best := map[float64]candidate{}
	bestLen := map[float64]float64{}
	for i, polyline := range polylines {
		length := polylineLength(polyline)
		level := levels[i]
		if length <= bestLen[level] {
			continue
		}
		bestLen[level] = length
		best[level] = candidate{polyline: polyline, color: colors[i]}
	}

	labels := make([]contourLabel, 0, len(best))
	for _, level := range dedupeFloat64(levels) {
		candidate, ok := best[level]
		if !ok {
			continue
		}
		position, angle := polylineLabelPlacement(candidate.polyline)
		labels = append(labels, contourLabel{
			Text:     formatter.Format(level),
			Position: position,
			Angle:    normalizeLabelAngle(angle),
			Color:    candidate.color,
			Level:    level,
		})
	}
	return labels
}

func (c *ContourSet) clabelLineIndices(levels []float64) ([]int, bool) {
	if c == nil || len(c.lineLevels) == 0 {
		return nil, false
	}
	if len(levels) == 0 {
		indices := make([]int, len(c.lineLevels))
		for i := range c.lineLevels {
			indices[i] = i
		}
		return indices, true
	}
	for _, level := range levels {
		if !contourLevelAvailable(c.Levels, level) {
			return nil, false
		}
	}
	indices := []int{}
	for i, level := range c.lineLevels {
		if contourLevelAvailable(levels, level) {
			indices = append(indices, i)
		}
	}
	return indices, true
}

func (c *ContourSet) clabelPlaceAutomatic(indices []int, opt ClabelOptions) []contourLabel {
	polylines := make([][]geom.Pt, 0, len(indices))
	levels := make([]float64, 0, len(indices))
	colors := make([]render.Color, 0, len(indices))
	for _, idx := range indices {
		if idx < 0 || c.Lines == nil || idx >= len(c.Lines.Segments) || idx >= len(c.lineLevels) {
			continue
		}
		level := c.lineLevels[idx]
		polylines = append(polylines, c.Lines.Segments[idx])
		levels = append(levels, level)
		colors = append(colors, c.clabelColor(idx, level, len(levels)-1, opt))
	}
	return contourLabels(polylines, levels, colors, c.LabelFormatter)
}

func (c *ContourSet) clabelPlaceManual(indices []int, opt ClabelOptions) []contourLabel {
	labels := make([]contourLabel, 0, len(opt.ManualPositions))
	for _, point := range opt.ManualPositions {
		idx, projection, angle, ok := c.nearestContourLabelPoint(indices, point)
		if !ok {
			continue
		}
		level := c.lineLevels[idx]
		labels = append(labels, contourLabel{
			Text:     c.LabelFormatter.Format(level),
			Position: projection,
			Angle:    angle,
			Color:    c.clabelColor(idx, level, len(labels), opt),
			Level:    level,
		})
	}
	return labels
}

func (c *ContourSet) nearestContourLabelPoint(indices []int, point geom.Pt) (int, geom.Pt, float64, bool) {
	bestIdx := -1
	bestPoint := geom.Pt{}
	bestAngle := 0.0
	bestDist := math.Inf(1)
	for _, idx := range indices {
		if c == nil || c.Lines == nil || idx < 0 || idx >= len(c.Lines.Segments) {
			continue
		}
		segment := c.Lines.Segments[idx]
		for i := 1; i < len(segment); i++ {
			projection, dist := projectPointToSegment(point, segment[i-1], segment[i])
			if dist >= bestDist {
				continue
			}
			bestDist = dist
			bestIdx = idx
			bestPoint = projection
			bestAngle = normalizeLabelAngle(math.Atan2(segment[i].Y-segment[i-1].Y, segment[i].X-segment[i-1].X))
		}
	}
	return bestIdx, bestPoint, bestAngle, bestIdx >= 0
}

func (c *ContourSet) clabelColor(segmentIndex int, level float64, labelIndex int, opt ClabelOptions) render.Color {
	if opt.Color != nil {
		return *opt.Color
	}
	if len(opt.Colors) > 0 {
		return opt.Colors[labelIndex%len(opt.Colors)]
	}
	if c != nil && c.Lines != nil {
		return colorAt(c.Lines.Color, c.Lines.Colors, segmentIndex)
	}
	return render.Color{}
}

func publicContourLabels(labels []contourLabel) []ContourLabel {
	out := make([]ContourLabel, len(labels))
	for i, label := range labels {
		out[i] = ContourLabel{
			Text:     label.Text,
			Level:    label.Level,
			Position: label.Position,
			Angle:    label.Angle,
			Color:    label.Color,
		}
	}
	return out
}

func uniqueLevelsForIndices(levels []float64, indices []int) []float64 {
	out := []float64{}
	for _, idx := range indices {
		if idx < 0 || idx >= len(levels) || contourLevelAvailable(out, levels[idx]) {
			continue
		}
		out = append(out, levels[idx])
	}
	return out
}

func contourLevelAvailable(levels []float64, level float64) bool {
	for _, existing := range levels {
		if math.Abs(existing-level) <= 1e-12 {
			return true
		}
	}
	return false
}

func firstFormatter(primary, fallback Formatter) Formatter {
	if primary != nil {
		return primary
	}
	return fallback
}

func projectPointToSegment(point, a, b geom.Pt) (geom.Pt, float64) {
	dx := b.X - a.X
	dy := b.Y - a.Y
	den := dx*dx + dy*dy
	if den <= 0 {
		return a, pointDistanceSquared(point, a)
	}
	t := ((point.X-a.X)*dx + (point.Y-a.Y)*dy) / den
	t = clampFloat(t, 0, 1)
	projection := geom.Pt{X: a.X + t*dx, Y: a.Y + t*dy}
	return projection, pointDistanceSquared(point, projection)
}

func contourInlineLabelSegments(lines *LineCollection, levels []float64, formatter Formatter, fontSize float64, r render.Renderer, ctx *DrawContext) ([][]geom.Pt, []render.Color, []float64, []contourLabel) {
	return contourInlineLabelSegmentsForLevels(lines, levels, nil, formatter, fontSize, 5, r, ctx)
}

func contourInlineLabelSegmentsForLevels(lines *LineCollection, levels, selectedLevels []float64, formatter Formatter, fontSize, inlineSpacing float64, r render.Renderer, ctx *DrawContext) ([][]geom.Pt, []render.Color, []float64, []contourLabel) {
	segments := make([][]geom.Pt, 0, len(lines.Segments))
	colors := make([]render.Color, 0, len(lines.Segments))
	widths := make([]float64, 0, len(lines.Segments))
	labels := []contourLabel{}
	placed := []geom.Pt{}

	for i, segment := range lines.Segments {
		color := colorAt(lines.Color, lines.Colors, i)
		width := widthAt(lines.LineWidth, lines.LineWidths, i)
		appendSegment := func(part []geom.Pt) {
			if len(part) < 2 {
				return
			}
			segments = append(segments, part)
			colors = append(colors, color)
			widths = append(widths, width)
		}

		if len(segment) < 2 || i >= len(levels) {
			appendSegment(segment)
			continue
		}
		if len(selectedLevels) > 0 && !contourLevelAvailable(selectedLevels, levels[i]) {
			appendSegment(segment)
			continue
		}

		text := formatter.Format(levels[i])
		if displayTextIsEmpty(text) {
			appendSegment(segment)
			continue
		}
		labelWidth := contourLabelWidth(text, fontSize, r, ctx)
		screen := contourDisplayPolyline(segment, ctx)
		if !contourPrintLabel(screen, labelWidth) {
			appendSegment(segment)
			continue
		}

		screenPos, labelIdx := contourLocateLabel(screen, labelWidth, placed)
		if labelIdx < 0 || labelIdx >= len(segment) {
			appendSegment(segment)
			continue
		}

		angle, parts := splitContourPolylineForLabel(segment, screen, labelIdx, labelWidth, inlineSpacing)
		if len(parts) == 0 {
			appendSegment(segment)
			continue
		}
		for _, part := range parts {
			appendSegment(part)
		}
		placed = append(placed, screenPos)
		labels = append(labels, contourLabel{
			Text:     text,
			Position: segment[labelIdx],
			Angle:    angle,
			Color:    color,
			Level:    levels[i],
		})
	}

	return segments, colors, widths, labels
}

func contourLabelWidth(text string, fontSize float64, r render.Renderer, ctx *DrawContext) float64 {
	fontKey := ""
	useTeX := false
	if ctx != nil {
		fontKey = ctx.RC.FontKey
		useTeX = ctx.RC.UseTeX
	}
	layout := measureSingleLineTextLayout(r, text, fontSize, fontKey, useTeX)
	if layout.Width > 0 {
		return layout.Width
	}
	return math.Max(fontSize, 1)
}

func contourDisplayPolyline(polyline []geom.Pt, ctx *DrawContext) []geom.Pt {
	out := make([]geom.Pt, len(polyline))
	for i, pt := range polyline {
		out[i] = ctx.DataToPixel.Apply(pt)
	}
	return out
}

func contourPrintLabel(line []geom.Pt, labelWidth float64) bool {
	if len(line) == 0 || labelWidth <= 0 {
		return false
	}
	if float64(len(line)) > 10*labelWidth {
		return true
	}
	minX, maxX := line[0].X, line[0].X
	minY, maxY := line[0].Y, line[0].Y
	for _, pt := range line[1:] {
		minX = math.Min(minX, pt.X)
		maxX = math.Max(maxX, pt.X)
		minY = math.Min(minY, pt.Y)
		maxY = math.Max(maxY, pt.Y)
	}
	return maxX-minX > 1.2*labelWidth || maxY-minY > 1.2*labelWidth
}

func contourLocateLabel(line []geom.Pt, labelWidth float64, placed []geom.Pt) (geom.Pt, int) {
	if len(line) == 0 {
		return geom.Pt{}, -1
	}
	ctrSize := len(line)
	nBlocks := 1
	if labelWidth > 1 {
		nBlocks = int(math.Ceil(float64(ctrSize) / labelWidth))
		if nBlocks < 1 {
			nBlocks = 1
		}
	}
	blockSize := ctrSize
	if nBlocks != 1 {
		blockSize = int(labelWidth)
		if blockSize < 1 {
			blockSize = 1
		}
	}

	type candidate struct {
		block    int
		distance float64
	}
	candidates := make([]candidate, nBlocks)
	for block := 0; block < nBlocks; block++ {
		first := contourResizedPoint(line, block, blockSize, 0)
		last := contourResizedPoint(line, block, blockSize, blockSize-1)
		length := math.Hypot(last.X-first.X, last.Y-first.Y)
		distance := math.Inf(1)
		if length > 0 {
			distance = 0
			for j := 0; j < blockSize; j++ {
				pt := contourResizedPoint(line, block, blockSize, j)
				cross := (first.Y-pt.Y)*(last.X-first.X) - (first.X-pt.X)*(last.Y-first.Y)
				distance += math.Abs(cross) / length
			}
		}
		candidates[block] = candidate{block: block, distance: distance}
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].distance < candidates[j].distance
	})

	halfBlock := blockSize / 2
	chosen := candidates[0].block
	for _, candidate := range candidates {
		idx := (candidate.block*blockSize + halfBlock) % ctrSize
		pt := line[idx]
		if !contourLabelTooClose(pt, labelWidth, placed) {
			chosen = candidate.block
			break
		}
	}
	idx := (chosen*blockSize + halfBlock) % ctrSize
	return line[idx], idx
}

func contourResizedPoint(line []geom.Pt, block, blockSize, offset int) geom.Pt {
	return line[(block*blockSize+offset)%len(line)]
}

func contourLabelTooClose(pt geom.Pt, labelWidth float64, placed []geom.Pt) bool {
	threshold := 1.2 * labelWidth
	threshold *= threshold
	for _, existing := range placed {
		dx := pt.X - existing.X
		dy := pt.Y - existing.Y
		if dx*dx+dy*dy < threshold {
			return true
		}
	}
	return false
}

func splitContourPolylineForLabel(data, screen []geom.Pt, labelIdx int, labelWidth, spacing float64) (float64, [][]geom.Pt) {
	if len(data) < 2 || len(data) != len(screen) || labelIdx < 0 || labelIdx >= len(data) {
		return 0, nil
	}
	cpls := contourCumulativeDisplayLengths(screen)
	total := cpls[len(cpls)-1]
	if total <= 0 {
		return 0, nil
	}
	if contourPolylineClosed(screen) {
		return splitClosedContourPolylineForLabel(data, screen, cpls, labelIdx, labelWidth, spacing)
	}

	center := cpls[labelIdx]
	angleStart := clampFloat(center-labelWidth/2, 0, total)
	angleEnd := clampFloat(center+labelWidth/2, 0, total)
	p0, ok0 := contourInterpolateAtCPL(screen, cpls, angleStart)
	p1, ok1 := contourInterpolateAtCPL(screen, cpls, angleEnd)
	angle := 0.0
	if ok0 && ok1 {
		angle = normalizeLabelAngle(math.Atan2(p1.Y-p0.Y, p1.X-p0.X))
	}

	gapStart := center - labelWidth/2 - spacing
	gapEnd := center + labelWidth/2 + spacing
	parts := [][]geom.Pt{}
	if gapStart > 0 {
		before := contourPolylineBeforeCPL(data, cpls, gapStart)
		if len(before) >= 2 {
			parts = append(parts, before)
		}
	}
	if gapEnd < total {
		after := contourPolylineAfterCPL(data, cpls, gapEnd)
		if len(after) >= 2 {
			parts = append(parts, after)
		}
	}
	return angle, parts
}

func splitClosedContourPolylineForLabel(data, screen []geom.Pt, cpls []float64, labelIdx int, labelWidth, spacing float64) (float64, [][]geom.Pt) {
	total := cpls[len(cpls)-1]
	gap := labelWidth + 2*spacing
	if total <= 0 || gap >= total {
		return 0, nil
	}

	center := cpls[labelIdx]
	p0, ok0 := contourInterpolateAtClosedCPL(screen, cpls, center-labelWidth/2)
	p1, ok1 := contourInterpolateAtClosedCPL(screen, cpls, center+labelWidth/2)
	angle := 0.0
	if ok0 && ok1 {
		angle = normalizeLabelAngle(math.Atan2(p1.Y-p0.Y, p1.X-p0.X))
	}

	gapStart := center - labelWidth/2 - spacing
	gapEnd := center + labelWidth/2 + spacing
	part := contourClosedPolylineComplement(data, cpls, gapEnd, gapStart+total)
	if len(part) < 2 {
		return angle, nil
	}
	return angle, [][]geom.Pt{part}
}

func contourPolylineClosed(points []geom.Pt) bool {
	return len(points) > 2 && sameContourPoint(points[0], points[len(points)-1])
}

func contourClosedPolylineComplement(data []geom.Pt, cpls []float64, start, end float64) []geom.Pt {
	total := cpls[len(cpls)-1]
	if total <= 0 {
		return nil
	}
	for start < 0 {
		start += total
		end += total
	}
	for start >= total {
		start -= total
		end -= total
	}
	if end <= start {
		end += total
	}
	if end-start <= 0 || end-start >= total {
		return nil
	}

	startPt, ok := contourInterpolateAtClosedCPL(data, cpls, start)
	if !ok {
		return nil
	}
	out := []geom.Pt{startPt}

	type closedVertex struct {
		cpl float64
		pt  geom.Pt
	}
	vertices := []closedVertex{}
	base := math.Floor(start/total) - 1
	for copyIdx := 0; copyIdx < 4; copyIdx++ {
		offset := (base + float64(copyIdx)) * total
		for i := 0; i < len(data); i++ {
			vertexCPL := cpls[i] + offset
			if vertexCPL <= start || vertexCPL >= end {
				continue
			}
			vertices = append(vertices, closedVertex{cpl: vertexCPL, pt: data[i]})
		}
	}
	sort.SliceStable(vertices, func(i, j int) bool {
		return vertices[i].cpl < vertices[j].cpl
	})
	for _, vertex := range vertices {
		out = appendContourPoint(out, vertex.pt)
	}

	endPt, ok := contourInterpolateAtClosedCPL(data, cpls, end)
	if !ok {
		return nil
	}
	out = appendContourPoint(out, endPt)
	return out
}

func contourInterpolateAtClosedCPL(points []geom.Pt, cpls []float64, target float64) (geom.Pt, bool) {
	total := cpls[len(cpls)-1]
	if total <= 0 {
		return geom.Pt{}, false
	}
	target = math.Mod(target, total)
	if target < 0 {
		target += total
	}
	return contourInterpolateAtCPL(points, cpls, target)
}

func contourRotatedTextAnchor(center geom.Pt, layout singleLineTextLayout, angle float64) geom.Pt {
	return rotatedTextBackendAnchorFromP(center, layout, TextAlignCenter, textLayoutVAlignCenter, angle, false)
}

func contourCumulativeDisplayLengths(points []geom.Pt) []float64 {
	cpls := make([]float64, len(points))
	for i := 1; i < len(points); i++ {
		cpls[i] = cpls[i-1] + math.Hypot(points[i].X-points[i-1].X, points[i].Y-points[i-1].Y)
	}
	return cpls
}

func contourPolylineBeforeCPL(data []geom.Pt, cpls []float64, target float64) []geom.Pt {
	out := []geom.Pt{}
	for i, pt := range data {
		if cpls[i] <= target {
			out = append(out, pt)
		}
	}
	if pt, ok := contourInterpolateAtCPL(data, cpls, target); ok {
		out = appendContourPoint(out, pt)
	}
	return out
}

func contourPolylineAfterCPL(data []geom.Pt, cpls []float64, target float64) []geom.Pt {
	out := []geom.Pt{}
	if pt, ok := contourInterpolateAtCPL(data, cpls, target); ok {
		out = append(out, pt)
	}
	for i, pt := range data {
		if cpls[i] >= target {
			out = appendContourPoint(out, pt)
		}
	}
	return out
}

func contourInterpolateAtCPL(points []geom.Pt, cpls []float64, target float64) (geom.Pt, bool) {
	if len(points) == 0 || len(points) != len(cpls) {
		return geom.Pt{}, false
	}
	if target <= cpls[0] {
		return points[0], true
	}
	last := len(cpls) - 1
	if target >= cpls[last] {
		return points[last], true
	}
	for i := 1; i < len(cpls); i++ {
		if cpls[i] < target {
			continue
		}
		if cpls[i] == cpls[i-1] {
			return points[i], true
		}
		t := (target - cpls[i-1]) / (cpls[i] - cpls[i-1])
		return interpolatePoint(points[i-1], points[i], t), true
	}
	return points[last], true
}

func appendContourPoint(points []geom.Pt, point geom.Pt) []geom.Pt {
	if len(points) > 0 && sameContourPoint(points[len(points)-1], point) {
		return points
	}
	return append(points, point)
}

func contourFormatter(formatter Formatter) Formatter {
	if formatter != nil {
		return formatter
	}
	return ScalarFormatter{Prec: 3}
}

func polylineLength(polyline []geom.Pt) float64 {
	total := 0.0
	for i := 1; i < len(polyline); i++ {
		total += math.Hypot(polyline[i].X-polyline[i-1].X, polyline[i].Y-polyline[i-1].Y)
	}
	return total
}

func polylineLabelPlacement(polyline []geom.Pt) (geom.Pt, float64) {
	total := polylineLength(polyline)
	if total == 0 || len(polyline) < 2 {
		if len(polyline) == 0 {
			return geom.Pt{}, 0
		}
		return polyline[0], 0
	}

	target := total * 0.5
	run := 0.0
	for i := 1; i < len(polyline); i++ {
		segLen := math.Hypot(polyline[i].X-polyline[i-1].X, polyline[i].Y-polyline[i-1].Y)
		if run+segLen >= target {
			t := (target - run) / segLen
			point := interpolatePoint(polyline[i-1], polyline[i], t)
			return point, math.Atan2(polyline[i].Y-polyline[i-1].Y, polyline[i].X-polyline[i-1].X)
		}
		run += segLen
	}

	last := polyline[len(polyline)-1]
	prev := polyline[len(polyline)-2]
	return last, math.Atan2(last.Y-prev.Y, last.X-prev.X)
}

func normalizeLabelAngle(angle float64) float64 {
	for angle > math.Pi/2 {
		angle -= math.Pi
	}
	for angle < -math.Pi/2 {
		angle += math.Pi
	}
	return angle
}
