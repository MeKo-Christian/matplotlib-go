package render

import (
	"math"
	"sort"

	"github.com/cwbudde/matplotlib-go/internal/geom"
)

// DrawHatchFallback expands Paint hatch metadata into ordinary clipped stroke
// paths for renderers that do not implement NativeHatcher.
func DrawHatchFallback(r Renderer, clipPath geom.Path, paint Paint) bool {
	if r == nil || paint.Hatch == "" {
		return false
	}
	bounds, ok := hatchPathBounds(clipPath)
	if !ok {
		return false
	}

	color := paint.HatchColor
	if color.A <= 0 {
		return false
	}

	counts := hatchCounts(paint.Hatch)
	if len(counts) == 0 {
		return false
	}

	clipPolygon := hatchClipPolygon(clipPath)
	if len(clipPolygon) < 3 {
		return false
	}

	lineWidth := paint.HatchLineWidth
	if lineWidth <= 0 {
		lineWidth = 1
	}
	spacingBase := paint.HatchSpacing
	if spacingBase <= 0 {
		spacingBase = 32
	}

	drew := false
	for pattern, count := range counts {
		spacing := math.Max(2, spacingBase/float64(count))
		hatchPaint := Paint{
			Stroke:    color,
			LineWidth: lineWidth,
			LineJoin:  JoinRound,
			LineCap:   CapRound,
		}
		for _, path := range hatchPatternPaths(pattern, bounds, spacing) {
			clipped := clipHatchPathToPolygon(path, clipPolygon)
			if len(clipped.C) == 0 {
				continue
			}
			if hatchPatternIsFilled(pattern) {
				hatchPaint.Fill = color
			} else {
				hatchPaint.Fill = Color{}
			}
			r.Path(clipped, &hatchPaint)
			drew = true
		}
	}
	return drew
}

func hatchPathBounds(path geom.Path) (geom.Rect, bool) {
	if len(path.V) == 0 {
		return geom.Rect{}, false
	}
	minX, maxX := path.V[0].X, path.V[0].X
	minY, maxY := path.V[0].Y, path.V[0].Y
	for _, p := range path.V[1:] {
		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	if minX == maxX || minY == maxY {
		return geom.Rect{}, false
	}
	return geom.Rect{Min: geom.Pt{X: minX, Y: minY}, Max: geom.Pt{X: maxX, Y: maxY}}, true
}

func hatchClipPolygon(path geom.Path) []geom.Pt {
	const curveSteps = 16
	polygon := make([]geom.Pt, 0, len(path.V))
	var current, start geom.Pt
	haveCurrent := false
	vi := 0
	for _, cmd := range path.C {
		switch cmd {
		case geom.MoveTo:
			if vi >= len(path.V) {
				return polygon
			}
			current = path.V[vi]
			start = current
			haveCurrent = true
			polygon = append(polygon, current)
			vi++
		case geom.LineTo:
			if vi >= len(path.V) || !haveCurrent {
				return polygon
			}
			current = path.V[vi]
			polygon = append(polygon, current)
			vi++
		case geom.QuadTo:
			if vi+1 >= len(path.V) || !haveCurrent {
				return polygon
			}
			ctrl := path.V[vi]
			to := path.V[vi+1]
			for step := 1; step <= curveSteps; step++ {
				t := float64(step) / curveSteps
				polygon = append(polygon, quadPoint(current, ctrl, to, t))
			}
			current = to
			vi += 2
		case geom.CubicTo:
			if vi+2 >= len(path.V) || !haveCurrent {
				return polygon
			}
			c1 := path.V[vi]
			c2 := path.V[vi+1]
			to := path.V[vi+2]
			for step := 1; step <= curveSteps; step++ {
				t := float64(step) / curveSteps
				polygon = append(polygon, cubicPoint(current, c1, c2, to, t))
			}
			current = to
			vi += 3
		case geom.ClosePath:
			if haveCurrent && !samePoint(current, start) {
				polygon = append(polygon, start)
			}
		}
	}
	return dedupeAdjacentPoints(polygon)
}

func quadPoint(p0, p1, p2 geom.Pt, t float64) geom.Pt {
	mt := 1 - t
	return geom.Pt{
		X: mt*mt*p0.X + 2*mt*t*p1.X + t*t*p2.X,
		Y: mt*mt*p0.Y + 2*mt*t*p1.Y + t*t*p2.Y,
	}
}

func cubicPoint(p0, p1, p2, p3 geom.Pt, t float64) geom.Pt {
	mt := 1 - t
	return geom.Pt{
		X: mt*mt*mt*p0.X + 3*mt*mt*t*p1.X + 3*mt*t*t*p2.X + t*t*t*p3.X,
		Y: mt*mt*mt*p0.Y + 3*mt*mt*t*p1.Y + 3*mt*t*t*p2.Y + t*t*t*p3.Y,
	}
}

func clipHatchPathToPolygon(path geom.Path, polygon []geom.Pt) geom.Path {
	if hatchPathFullyInsidePolygon(path, polygon) {
		return path
	}
	out := geom.Path{}
	var current geom.Pt
	haveCurrent := false
	vi := 0
	for _, cmd := range path.C {
		switch cmd {
		case geom.MoveTo:
			if vi >= len(path.V) {
				return out
			}
			current = path.V[vi]
			haveCurrent = true
			vi++
		case geom.LineTo:
			if vi >= len(path.V) || !haveCurrent {
				return out
			}
			to := path.V[vi]
			appendClippedSegment(&out, current, to, polygon)
			current = to
			vi++
		case geom.QuadTo:
			vi += 2
		case geom.CubicTo:
			vi += 3
		}
	}
	return out
}

func hatchPathFullyInsidePolygon(path geom.Path, polygon []geom.Pt) bool {
	if len(path.V) == 0 || len(polygon) < 3 {
		return false
	}
	for _, pt := range path.V {
		if !hatchPointInPolygon(pt, polygon) && !pointOnPolygon(pt, polygon) {
			return false
		}
	}
	return true
}

func appendClippedSegment(out *geom.Path, a, b geom.Pt, polygon []geom.Pt) {
	ts := []float64{0, 1}
	for i := range polygon {
		c := polygon[i]
		d := polygon[(i+1)%len(polygon)]
		if samePoint(c, d) {
			continue
		}
		if t, ok := segmentIntersectionParameter(a, b, c, d); ok && t > -1e-9 && t < 1+1e-9 {
			ts = append(ts, clamp(t, 0, 1))
		}
	}
	sort.Float64s(ts)
	ts = uniqueSortedParameters(ts)
	for i := 0; i < len(ts)-1; i++ {
		t0 := ts[i]
		t1 := ts[i+1]
		if t1-t0 <= 1e-9 {
			continue
		}
		mid := lerpPoint(a, b, (t0+t1)/2)
		if !hatchPointInPolygon(mid, polygon) && !pointOnPolygon(mid, polygon) {
			continue
		}
		p0 := lerpPoint(a, b, t0)
		p1 := lerpPoint(a, b, t1)
		out.MoveTo(p0)
		out.LineTo(p1)
	}
}

func segmentIntersectionParameter(a, b, c, d geom.Pt) (float64, bool) {
	r := geom.Pt{X: b.X - a.X, Y: b.Y - a.Y}
	s := geom.Pt{X: d.X - c.X, Y: d.Y - c.Y}
	denom := cross(r, s)
	if math.Abs(denom) < 1e-9 {
		return 0, false
	}
	qp := geom.Pt{X: c.X - a.X, Y: c.Y - a.Y}
	t := cross(qp, s) / denom
	u := cross(qp, r) / denom
	if t < -1e-9 || t > 1+1e-9 || u < -1e-9 || u > 1+1e-9 {
		return 0, false
	}
	return t, true
}

func pointOnPolygon(p geom.Pt, polygon []geom.Pt) bool {
	for i := range polygon {
		if hatchPointOnSegment(p, polygon[i], polygon[(i+1)%len(polygon)]) {
			return true
		}
	}
	return false
}

func hatchPointOnSegment(p, a, b geom.Pt) bool {
	ab := geom.Pt{X: b.X - a.X, Y: b.Y - a.Y}
	ap := geom.Pt{X: p.X - a.X, Y: p.Y - a.Y}
	if math.Abs(cross(ab, ap)) > 1e-7 {
		return false
	}
	dot := ap.X*ab.X + ap.Y*ab.Y
	if dot < -1e-7 {
		return false
	}
	len2 := ab.X*ab.X + ab.Y*ab.Y
	return dot <= len2+1e-7
}

func hatchPointInPolygon(p geom.Pt, polygon []geom.Pt) bool {
	if len(polygon) < 3 {
		return false
	}
	inside := false
	j := len(polygon) - 1
	for i := range polygon {
		pi := polygon[i]
		pj := polygon[j]
		if ((pi.Y > p.Y) != (pj.Y > p.Y)) &&
			(p.X < (pj.X-pi.X)*(p.Y-pi.Y)/(pj.Y-pi.Y)+pi.X) {
			inside = !inside
		}
		j = i
	}
	return inside
}

func cross(a, b geom.Pt) float64 {
	return a.X*b.Y - a.Y*b.X
}

func lerpPoint(a, b geom.Pt, t float64) geom.Pt {
	return geom.Pt{
		X: a.X + (b.X-a.X)*t,
		Y: a.Y + (b.Y-a.Y)*t,
	}
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func uniqueSortedParameters(values []float64) []float64 {
	if len(values) == 0 {
		return values
	}
	out := values[:1]
	for _, v := range values[1:] {
		if math.Abs(v-out[len(out)-1]) > 1e-8 {
			out = append(out, v)
		}
	}
	return out
}

func dedupeAdjacentPoints(points []geom.Pt) []geom.Pt {
	if len(points) == 0 {
		return points
	}
	out := points[:1]
	for _, p := range points[1:] {
		if !samePoint(p, out[len(out)-1]) {
			out = append(out, p)
		}
	}
	if len(out) > 1 && samePoint(out[0], out[len(out)-1]) {
		out = out[:len(out)-1]
	}
	return out
}

func samePoint(a, b geom.Pt) bool {
	return math.Abs(a.X-b.X) < 1e-9 && math.Abs(a.Y-b.Y) < 1e-9
}

func hatchCounts(pattern string) map[rune]int {
	counts := map[rune]int{}
	for _, ch := range pattern {
		switch ch {
		case '|', '-', '/', '\\', '+', 'x', 'X', 'o', 'O', '.', '*':
			counts[ch]++
		}
	}
	return counts
}

func hatchPatternPaths(pattern rune, bounds geom.Rect, spacing float64) []geom.Path {
	switch pattern {
	case '|':
		return []geom.Path{verticalHatchPath(bounds, spacing)}
	case '-':
		return []geom.Path{horizontalHatchPath(bounds, spacing)}
	case '/':
		return []geom.Path{slashHatchPath(bounds, spacing)}
	case '\\':
		return []geom.Path{backslashHatchPath(bounds, spacing)}
	case '+':
		return []geom.Path{
			verticalHatchPath(bounds, spacing),
			horizontalHatchPath(bounds, spacing),
		}
	case 'x', 'X':
		return []geom.Path{
			slashHatchPath(bounds, spacing),
			backslashHatchPath(bounds, spacing),
		}
	case 'o':
		return circleHatchPaths(bounds, spacing, 0.20)
	case 'O':
		return circleHatchPaths(bounds, spacing, 0.35)
	case '.':
		return circleHatchPaths(bounds, spacing, 0.10)
	case '*':
		return starHatchPaths(bounds, spacing)
	default:
		return nil
	}
}

func hatchPatternIsFilled(pattern rune) bool {
	return pattern == '.' || pattern == '*'
}

func verticalHatchPath(bounds geom.Rect, spacing float64) geom.Path {
	if bounds.W() <= 0 || bounds.H() <= 0 {
		return geom.Path{}
	}
	path := geom.Path{}
	for x := bounds.Min.X; x <= bounds.Max.X+spacing*0.5; x += spacing {
		path.MoveTo(geom.Pt{X: x, Y: bounds.Min.Y})
		path.LineTo(geom.Pt{X: x, Y: bounds.Max.Y})
	}
	return path
}

func horizontalHatchPath(bounds geom.Rect, spacing float64) geom.Path {
	if bounds.W() <= 0 || bounds.H() <= 0 {
		return geom.Path{}
	}
	path := geom.Path{}
	for y := bounds.Min.Y; y <= bounds.Max.Y+spacing*0.5; y += spacing {
		path.MoveTo(geom.Pt{X: bounds.Min.X, Y: y})
		path.LineTo(geom.Pt{X: bounds.Max.X, Y: y})
	}
	return path
}

func slashHatchPath(bounds geom.Rect, spacing float64) geom.Path {
	if bounds.W() <= 0 || bounds.H() <= 0 {
		return geom.Path{}
	}
	path := geom.Path{}
	span := bounds.W() + bounds.H() + 2*spacing
	x0 := bounds.Min.X - span
	x1 := bounds.Max.X + span
	// Matplotlib AGG draws the hatch path into an origin-anchored DPI-sized
	// tile and repeats it, so diagonal hatch phase is global, not patch-local.
	minC := bounds.Min.X + bounds.Min.Y - span
	maxC := bounds.Max.X + bounds.Max.Y + span
	start := math.Floor(minC/spacing) * spacing
	for c := start; c <= maxC; c += spacing {
		path.MoveTo(geom.Pt{X: x0, Y: c - x0})
		path.LineTo(geom.Pt{X: x1, Y: c - x1})
	}
	return path
}

func backslashHatchPath(bounds geom.Rect, spacing float64) geom.Path {
	if bounds.W() <= 0 || bounds.H() <= 0 {
		return geom.Path{}
	}
	path := geom.Path{}
	span := bounds.W() + bounds.H() + 2*spacing
	x0 := bounds.Min.X - span
	x1 := bounds.Max.X + span
	minC := bounds.Min.Y - bounds.Max.X - span
	maxC := bounds.Max.Y - bounds.Min.X + span
	start := math.Floor(minC/spacing) * spacing
	for c := start; c <= maxC; c += spacing {
		path.MoveTo(geom.Pt{X: x0, Y: x0 + c})
		path.LineTo(geom.Pt{X: x1, Y: x1 + c})
	}
	return path
}

func circleHatchPaths(bounds geom.Rect, spacing, radiusFactor float64) []geom.Path {
	if bounds.W() <= 0 || bounds.H() <= 0 || spacing <= 0 {
		return nil
	}
	radius := math.Max(0.5, spacing*radiusFactor)
	return shapeHatchPaths(bounds, spacing, func(center geom.Pt) geom.Path {
		return circleHatchPath(center, radius)
	})
}

func starHatchPaths(bounds geom.Rect, spacing float64) []geom.Path {
	if bounds.W() <= 0 || bounds.H() <= 0 || spacing <= 0 {
		return nil
	}
	outer := math.Max(0.75, spacing/3)
	inner := outer * 0.5
	return shapeHatchPaths(bounds, spacing, func(center geom.Pt) geom.Path {
		return starHatchPath(center, outer, inner)
	})
}

func shapeHatchPaths(bounds geom.Rect, spacing float64, build func(geom.Pt) geom.Path) []geom.Path {
	radiusPad := spacing * 0.5
	minX := bounds.Min.X + radiusPad
	maxX := bounds.Max.X - radiusPad
	minY := bounds.Min.Y + radiusPad
	maxY := bounds.Max.Y - radiusPad
	if minX > maxX {
		minX, maxX = (bounds.Min.X+bounds.Max.X)/2, (bounds.Min.X+bounds.Max.X)/2
	}
	if minY > maxY {
		minY, maxY = (bounds.Min.Y+bounds.Max.Y)/2, (bounds.Min.Y+bounds.Max.Y)/2
	}

	var paths []geom.Path
	row := 0
	for y := minY; y <= maxY+1e-9; y += spacing {
		offset := 0.0
		if row%2 == 1 {
			offset = spacing / 2
		}
		for x := minX + offset; x <= maxX+1e-9; x += spacing {
			paths = append(paths, build(geom.Pt{X: x, Y: y}))
		}
		row++
	}
	return paths
}

func circleHatchPath(center geom.Pt, radius float64) geom.Path {
	k := radius * 0.5522847498307936
	path := geom.Path{}
	path.MoveTo(geom.Pt{X: center.X + radius, Y: center.Y})
	path.CubicTo(
		geom.Pt{X: center.X + radius, Y: center.Y + k},
		geom.Pt{X: center.X + k, Y: center.Y + radius},
		geom.Pt{X: center.X, Y: center.Y + radius},
	)
	path.CubicTo(
		geom.Pt{X: center.X - k, Y: center.Y + radius},
		geom.Pt{X: center.X - radius, Y: center.Y + k},
		geom.Pt{X: center.X - radius, Y: center.Y},
	)
	path.CubicTo(
		geom.Pt{X: center.X - radius, Y: center.Y - k},
		geom.Pt{X: center.X - k, Y: center.Y - radius},
		geom.Pt{X: center.X, Y: center.Y - radius},
	)
	path.CubicTo(
		geom.Pt{X: center.X + k, Y: center.Y - radius},
		geom.Pt{X: center.X + radius, Y: center.Y - k},
		geom.Pt{X: center.X + radius, Y: center.Y},
	)
	path.Close()
	return path
}

func starHatchPath(center geom.Pt, outer, inner float64) geom.Path {
	points := make([]geom.Pt, 0, 10)
	for i := 0; i < 10; i++ {
		radius := outer
		if i%2 == 1 {
			radius = inner
		}
		angle := -math.Pi/2 + float64(i)*math.Pi/5
		points = append(points, geom.Pt{
			X: center.X + radius*math.Cos(angle),
			Y: center.Y + radius*math.Sin(angle),
		})
	}
	path := geom.Path{}
	for i, pt := range points {
		if i == 0 {
			path.MoveTo(pt)
		} else {
			path.LineTo(pt)
		}
	}
	path.Close()
	return path
}
