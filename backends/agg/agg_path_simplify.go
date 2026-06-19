package agg

import "github.com/cwbudde/matplotlib-go/geom"

func simplifyLinePath(path geom.Path, threshold float64) geom.Path {
	if threshold <= 0 || pathHasCurvesOrClose(path) {
		return path
	}
	out := geom.Path{}
	var current []geom.Pt
	flush := func() {
		if len(current) == 0 {
			return
		}
		points := simplifyPolyline(current, threshold)
		if len(points) > 0 {
			out.MoveTo(points[0])
			for _, pt := range points[1:] {
				out.LineTo(pt)
			}
		}
		current = current[:0]
	}

	vi := 0
	for _, cmd := range path.C {
		switch cmd {
		case geom.MoveTo:
			flush()
			current = append(current, path.V[vi])
			vi++
		case geom.LineTo:
			current = append(current, path.V[vi])
			vi++
		}
	}
	flush()
	return out
}

func pathHasCurvesOrClose(path geom.Path) bool {
	for _, cmd := range path.C {
		if cmd == geom.QuadTo || cmd == geom.CubicTo || cmd == geom.ClosePath {
			return true
		}
	}
	return false
}

func simplifyPolyline(points []geom.Pt, threshold float64) []geom.Pt {
	if len(points) <= 2 {
		return append([]geom.Pt(nil), points...)
	}
	keep := make([]bool, len(points))
	keep[0] = true
	keep[len(points)-1] = true
	simplifyPolylineRange(points, threshold*threshold, 0, len(points)-1, keep)
	out := make([]geom.Pt, 0, len(points))
	for i, pt := range points {
		if keep[i] {
			out = append(out, pt)
		}
	}
	return out
}

func simplifyPolylineRange(points []geom.Pt, threshold2 float64, first, last int, keep []bool) {
	if last <= first+1 {
		return
	}
	maxDist2 := -1.0
	maxIndex := -1
	for i := first + 1; i < last; i++ {
		dist2 := pointSegmentDistanceSquared(points[i], points[first], points[last])
		if dist2 > maxDist2 {
			maxDist2 = dist2
			maxIndex = i
		}
	}
	if maxDist2 > threshold2 && maxIndex >= 0 {
		keep[maxIndex] = true
		simplifyPolylineRange(points, threshold2, first, maxIndex, keep)
		simplifyPolylineRange(points, threshold2, maxIndex, last, keep)
	}
}

func pointSegmentDistanceSquared(p, a, b geom.Pt) float64 {
	dx := b.X - a.X
	dy := b.Y - a.Y
	if dx == 0 && dy == 0 {
		return squaredDistance(p, a)
	}
	t := ((p.X-a.X)*dx + (p.Y-a.Y)*dy) / (dx*dx + dy*dy)
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	proj := geom.Pt{X: a.X + t*dx, Y: a.Y + t*dy}
	return squaredDistance(p, proj)
}

func squaredDistance(a, b geom.Pt) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return dx*dx + dy*dy
}
