package core

import (
	"math"
	"sort"

	"github.com/cwbudde/matplotlib-go/geom"
)

func contourPolylines(tri Triangulation, values, levels []float64) ([][]geom.Pt, []float64) {
	var polylines [][]geom.Pt
	var polylineLevels []float64
	for _, level := range levels {
		segments := contourSegmentsForLevel(tri, values, level)
		for _, polyline := range stitchContourSegments(segments) {
			if len(polyline) < 2 {
				continue
			}
			polylines = append(polylines, polyline)
			polylineLevels = append(polylineLevels, level)
		}
	}
	return polylines, polylineLevels
}

func contourGridPolylines(x, y []float64, data [][]float64, levels []float64) ([][]geom.Pt, []float64) {
	return contourGridPolylinesCornerMask(x, y, data, levels, false)
}

func contourGridPolylinesCornerMask(x, y []float64, data [][]float64, levels []float64, cornerMask bool) ([][]geom.Pt, []float64) {
	rows := len(data)
	if rows < 2 || len(x) < 2 || len(y) < 2 {
		return nil, nil
	}
	cols := len(data[0])
	if cols < 2 || len(x) < cols || len(y) < rows {
		return nil, nil
	}
	for row := 1; row < rows; row++ {
		if len(data[row]) != cols {
			return nil, nil
		}
	}

	var polylines [][]geom.Pt
	var polylineLevels []float64
	for _, level := range levels {
		var segments [][]geom.Pt
		for row := 0; row+1 < rows; row++ {
			for col := 0; col+1 < cols; col++ {
				cellSegments := contourCellSegmentsForLevelCornerMask(
					[4]geom.Pt{
						{X: x[col], Y: y[row]},
						{X: x[col+1], Y: y[row]},
						{X: x[col+1], Y: y[row+1]},
						{X: x[col], Y: y[row+1]},
					},
					[4]float64{
						data[row][col],
						data[row][col+1],
						data[row+1][col+1],
						data[row+1][col],
					},
					level,
					cornerMask,
				)
				segments = append(segments, cellSegments...)
			}
		}
		levelPolylines := stitchContourSegments(segments)
		sort.SliceStable(levelPolylines, func(i, j int) bool {
			return contourPolylineClosed(levelPolylines[j]) && !contourPolylineClosed(levelPolylines[i])
		})
		for _, polyline := range levelPolylines {
			if len(polyline) < 2 {
				continue
			}
			polyline = rotateClosedLoopToContourpyStart(polyline, x, y)
			polyline = orientStructuredOpenBoundaryPolyline(polyline, x, y)
			polylines = append(polylines, polyline)
			polylineLevels = append(polylineLevels, level)
		}
	}
	return polylines, polylineLevels
}

func contourCellSegmentsForLevelCornerMask(points [4]geom.Pt, values [4]float64, level float64, cornerMask bool) [][]geom.Pt {
	finite := 0
	var trianglePoints [3]geom.Pt
	var triangleValues [3]float64
	for i, value := range values {
		if !isFinite(value) {
			continue
		}
		if finite < len(trianglePoints) {
			trianglePoints[finite] = points[i]
			triangleValues[finite] = value
		}
		finite++
	}
	if finite == 4 {
		return contourCellSegmentsForLevel(points, values, level)
	}
	if !cornerMask || finite != 3 {
		return nil
	}
	segment, ok := triangleContourSegment(trianglePoints, triangleValues, level)
	if !ok {
		return nil
	}
	return [][]geom.Pt{segment}
}

type contourBoundarySide uint8

const (
	contourBoundaryNone contourBoundarySide = iota
	contourBoundaryLeft
	contourBoundaryRight
	contourBoundaryBottom
	contourBoundaryTop
)

func orientStructuredOpenBoundaryPolyline(polyline []geom.Pt, x, y []float64) []geom.Pt {
	if len(polyline) < 2 || contourPolylineClosed(polyline) || len(x) == 0 || len(y) == 0 {
		return polyline
	}
	first := structuredBoundarySide(polyline[0], x, y)
	last := structuredBoundarySide(polyline[len(polyline)-1], x, y)
	if first == contourBoundaryNone || last == contourBoundaryNone {
		return polyline
	}
	if first == last {
		if contourSameBoundaryStartComesAfterEnd(polyline[0], polyline[len(polyline)-1], first) {
			return reversePoints(polyline)
		}
		return polyline
	}
	if structuredBoundarySideCount(polyline, x, y) != 2 {
		return polyline
	}
	desired, ok := contourpyOpenBoundaryStartSide(first, last)
	if !ok || first == desired {
		return polyline
	}
	if last == desired {
		return reversePoints(polyline)
	}
	return polyline
}

func contourSameBoundaryStartComesAfterEnd(first, last geom.Pt, side contourBoundarySide) bool {
	switch side {
	case contourBoundaryBottom:
		return first.X > last.X
	case contourBoundaryRight:
		return first.Y < last.Y
	case contourBoundaryTop:
		return first.X < last.X
	case contourBoundaryLeft:
		return first.Y > last.Y
	default:
		return false
	}
}

func structuredBoundarySideCount(polyline []geom.Pt, x, y []float64) int {
	sides := map[contourBoundarySide]bool{}
	for _, pt := range polyline {
		side := structuredBoundarySide(pt, x, y)
		if side != contourBoundaryNone {
			sides[side] = true
		}
	}
	return len(sides)
}

func contourpyOpenBoundaryStartSide(a, b contourBoundarySide) (contourBoundarySide, bool) {
	switch {
	case boundarySidePair(a, b, contourBoundaryBottom, contourBoundaryLeft):
		return contourBoundaryBottom, true
	case boundarySidePair(a, b, contourBoundaryBottom, contourBoundaryRight):
		return contourBoundaryRight, true
	case boundarySidePair(a, b, contourBoundaryLeft, contourBoundaryTop):
		return contourBoundaryLeft, true
	case boundarySidePair(a, b, contourBoundaryTop, contourBoundaryRight):
		return contourBoundaryTop, true
	default:
		return contourBoundaryNone, false
	}
}

func boundarySidePair(a, b, c, d contourBoundarySide) bool {
	return (a == c && b == d) || (a == d && b == c)
}

func structuredBoundarySide(pt geom.Pt, x, y []float64) contourBoundarySide {
	minX, maxX := math.Min(x[0], x[len(x)-1]), math.Max(x[0], x[len(x)-1])
	minY, maxY := math.Min(y[0], y[len(y)-1]), math.Max(y[0], y[len(y)-1])
	switch {
	case math.Abs(pt.X-minX) <= 1e-9:
		return contourBoundaryLeft
	case math.Abs(pt.X-maxX) <= 1e-9:
		return contourBoundaryRight
	case math.Abs(pt.Y-minY) <= 1e-9:
		return contourBoundaryBottom
	case math.Abs(pt.Y-maxY) <= 1e-9:
		return contourBoundaryTop
	default:
		return contourBoundaryNone
	}
}

func contourCellSegmentsForLevel(points [4]geom.Pt, values [4]float64, level float64) [][]geom.Pt {
	for _, value := range values {
		if !isFinite(value) {
			return nil
		}
	}

	above := [4]bool{}
	for i, value := range values {
		above[i] = value >= level
	}

	edgePairs := [4][2]int{{0, 1}, {1, 2}, {2, 3}, {3, 0}}
	edgePoints := make([]geom.Pt, 4)
	edgeHit := [4]bool{}
	for edgeIdx, pair := range edgePairs {
		aIdx, bIdx := pair[0], pair[1]
		aValue, bValue := values[aIdx], values[bIdx]
		if aValue == bValue {
			continue
		}
		minValue, maxValue := math.Min(aValue, bValue), math.Max(aValue, bValue)
		if level < minValue || level > maxValue {
			continue
		}
		t := (level - aValue) / (bValue - aValue)
		if t < 0 || t > 1 {
			continue
		}
		edgePoints[edgeIdx] = interpolatePoint(points[aIdx], points[bIdx], t)
		edgeHit[edgeIdx] = true
	}

	if above[0] == above[2] && above[1] == above[3] && above[0] != above[1] {
		// Saddle cell: opposite corners share a sign, adjacent corners differ, so
		// all four edges are crossed and the level splits the cell into two
		// separate segments. Matplotlib/contourpy disambiguates with the bilinear
		// value at the cell centre (the mean of the four corners): the diagonal
		// corner-pair whose sign differs from the centre is isolated, each corner
		// joined across its two incident edges. The centre uses a strict compare
		// so an exact centre==level tie resolves as "below" (matches contourpy).
		// Edge indices: 0=bottom(0-1), 1=right(1-2), 2=top(2-3), 3=left(3-0).
		centerAbove := (values[0]+values[1]+values[2]+values[3])/4 > level
		pairs := [2][2]int{{0, 1}, {2, 3}} // isolate corners 1 and 3
		if above[0] != centerAbove {
			pairs = [2][2]int{{0, 3}, {1, 2}} // isolate corners 0 and 2
		}
		segments := make([][]geom.Pt, 0, 2)
		for _, pair := range pairs {
			if edgeHit[pair[0]] && edgeHit[pair[1]] {
				segments = append(segments, []geom.Pt{edgePoints[pair[0]], edgePoints[pair[1]]})
			}
		}
		if len(segments) == 2 {
			return segments
		}
	}

	var hits []geom.Pt
	for edgeIdx := range edgeHit {
		if edgeHit[edgeIdx] && !containsPoint(hits, edgePoints[edgeIdx]) {
			hits = append(hits, edgePoints[edgeIdx])
		}
	}
	if len(hits) == 2 {
		return [][]geom.Pt{hits}
	}
	if len(hits) == 4 {
		return [][]geom.Pt{{hits[0], hits[1]}, {hits[2], hits[3]}}
	}
	return nil
}

func contourSegmentsForLevel(tri Triangulation, values []float64, level float64) [][]geom.Pt {
	segments := make([][]geom.Pt, 0, len(tri.Triangles))
	for triIdx, triangle := range tri.Triangles {
		if tri.Masked(triIdx) {
			continue
		}
		segment, ok := triangleContourSegment(
			[3]geom.Pt{tri.Point(triangle[0]), tri.Point(triangle[1]), tri.Point(triangle[2])},
			[3]float64{values[triangle[0]], values[triangle[1]], values[triangle[2]]},
			level,
		)
		if ok {
			segments = append(segments, segment)
		}
	}
	return segments
}

func triangleContourSegment(points [3]geom.Pt, values [3]float64, level float64) ([]geom.Pt, bool) {
	intersections := []geom.Pt{}
	edges := [][2]int{{0, 1}, {1, 2}, {2, 0}}
	for _, edge := range edges {
		aIdx, bIdx := edge[0], edge[1]
		aValue := values[aIdx]
		bValue := values[bIdx]
		if !isFinite(aValue) || !isFinite(bValue) {
			return nil, false
		}
		if aValue == bValue {
			continue
		}
		minValue := math.Min(aValue, bValue)
		maxValue := math.Max(aValue, bValue)
		if level < minValue || level > maxValue {
			continue
		}
		t := (level - aValue) / (bValue - aValue)
		if t < 0 || t > 1 {
			continue
		}
		point := interpolatePoint(points[aIdx], points[bIdx], t)
		if !containsPoint(intersections, point) {
			intersections = append(intersections, point)
		}
	}
	if len(intersections) != 2 {
		return nil, false
	}
	return intersections, true
}

func stitchContourSegments(segments [][]geom.Pt) [][]geom.Pt {
	remaining := make([][]geom.Pt, 0, len(segments))
	for _, segment := range segments {
		if len(segment) >= 2 {
			remaining = append(remaining, append([]geom.Pt(nil), segment...))
		}
	}
	out := [][]geom.Pt{}
	for len(remaining) > 0 {
		polyline := append([]geom.Pt(nil), remaining[0]...)
		remaining = remaining[1:]
		progress := true
		for progress {
			progress = false
			for i := 0; i < len(remaining); i++ {
				segment := remaining[i]
				switch {
				case sameContourPoint(polyline[len(polyline)-1], segment[0]):
					polyline = append(polyline, segment[1:]...)
				case sameContourPoint(polyline[len(polyline)-1], segment[len(segment)-1]):
					polyline = append(polyline, reversePoints(segment[:len(segment)-1])...)
				case sameContourPoint(polyline[0], segment[len(segment)-1]):
					polyline = append(segment[:len(segment)-1], polyline...)
				case sameContourPoint(polyline[0], segment[0]):
					polyline = append(reversePoints(segment[1:]), polyline...)
				default:
					continue
				}
				remaining = append(remaining[:i], remaining[i+1:]...)
				progress = true
				break
			}
		}
		out = append(out, rotateClosedContourPolylineToMatplotlibStart(polyline))
	}
	return out
}

// rotateClosedLoopToContourpyStart rotates a closed structured-grid contour loop
// so it begins at the vertex Matplotlib's default contour generator (contourpy
// "mpl2014") emits first. Go's marching squares yields the same vertex sequence
// and winding as contourpy, only rotated to a different start; aligning the
// start makes dash phase (and manual-label tangents) match Matplotlib.
//
// contourpy traces boundary-touching loops during its boundary pass and other
// loops during its interior pass, so the start is:
//   - the leftmost vertex on the bottom grid boundary, if the loop touches it;
//   - otherwise the leftmost vertical-grid-edge crossing, ties broken by
//     minimum y (contourpy's column-major interior cell scan).
func rotateClosedLoopToContourpyStart(polyline []geom.Pt, x, y []float64) []geom.Pt {
	if !contourPolylineClosed(polyline) || len(x) < 2 || len(y) < 2 {
		return polyline
	}
	body := polyline[:len(polyline)-1]
	if len(body) == 0 {
		return polyline
	}

	yBottom := math.Min(y[0], y[len(y)-1])
	best := -1
	// Phase 1: a loop tangent to the bottom boundary starts at its leftmost
	// bottom-boundary vertex.
	for i, p := range body {
		if math.Abs(p.Y-yBottom) <= 1e-7*(1+math.Abs(yBottom)) {
			if best < 0 || p.X < body[best].X-1e-9 {
				best = i
			}
		}
	}
	if best < 0 {
		// Phase 2: interior loop — contourpy's row-major cell scan starts at the
		// leftmost vertical-edge crossing in the bottommost row-band.
		bestRow := 0
		for i, p := range body {
			if !onVerticalGridEdge(p.X, x) {
				continue
			}
			row := gridRowBand(p.Y, y)
			if best < 0 || row < bestRow ||
				(row == bestRow && p.X < body[best].X-1e-9) {
				best = i
				bestRow = row
			}
		}
	}
	if best <= 0 {
		return polyline
	}
	out := make([]geom.Pt, 0, len(polyline))
	out = append(out, body[best:]...)
	out = append(out, body[:best]...)
	out = append(out, out[0])
	return out
}

// onVerticalGridEdge reports whether px lies on one of the grid's vertical edges
// (x equals a grid coordinate): a vertical-edge contour crossing.
func onVerticalGridEdge(px float64, x []float64) bool {
	for _, gx := range x {
		if math.Abs(px-gx) <= 1e-7*(1+math.Abs(gx)) {
			return true
		}
	}
	return false
}

// gridRowBand returns the index of the grid row-band (cell row) bracketing py,
// numbered from the bottom of the grid upward regardless of y ordering.
func gridRowBand(py float64, y []float64) int {
	ascending := y[len(y)-1] >= y[0]
	for j := 0; j+1 < len(y); j++ {
		lo, hi := y[j], y[j+1]
		if !ascending {
			lo, hi = hi, lo
		}
		if py >= lo-1e-9 && py <= hi+1e-9 {
			if ascending {
				return j
			}
			return len(y) - 2 - j
		}
	}
	if py < math.Min(y[0], y[len(y)-1]) {
		return 0
	}
	return len(y) - 2
}

func rotateClosedContourPolylineToMatplotlibStart(polyline []geom.Pt) []geom.Pt {
	if !contourPolylineClosed(polyline) {
		return polyline
	}
	body := polyline[:len(polyline)-1]
	if len(body) == 0 {
		return polyline
	}
	start := 0
	const startTieTolerance = 1e-7
	for i := 1; i < len(body); i++ {
		if body[i].Y < body[start].Y-startTieTolerance || (math.Abs(body[i].Y-body[start].Y) <= startTieTolerance && body[i].X < body[start].X) {
			start = i
		}
	}
	next := (start + 1) % len(body)
	prev := (start + len(body) - 1) % len(body)
	if math.Abs(body[prev].Y-body[start].Y) <= startTieTolerance && body[next].Y > body[start].Y && body[start].X-body[next].X >= 0.75 {
		start = next
	}
	out := make([]geom.Pt, 0, len(polyline))
	out = append(out, body[start:]...)
	out = append(out, body[:start]...)
	out = append(out, out[0])
	return out
}

func reversePoints(points []geom.Pt) []geom.Pt {
	out := append([]geom.Pt(nil), points...)
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}
