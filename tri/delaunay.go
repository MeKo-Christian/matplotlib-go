package tri

import (
	"math"
	"sort"
)

// delaunayTriangle is an oriented triangle used during Bowyer–Watson insertion.
type delaunayTriangle struct {
	a int
	b int
	c int
}

// delaunayTriangles computes a Delaunay triangulation of the supplied points
// using the incremental Bowyer–Watson algorithm with a super-triangle.
//
// Note: this is a pure-Go triangulation and is NOT byte-for-byte identical to
// matplotlib's Qhull-based Delaunay. It is deterministic and matches the prior
// in-tree behaviour exactly (same tie-break ordering), so downstream rendering
// parity is preserved.
func delaunayTriangles(x, y []float64) ([][3]int, bool) {
	n := len(x)
	minX, maxX := x[0], x[0]
	minY, maxY := y[0], y[0]
	for i := 1; i < n; i++ {
		minX = math.Min(minX, x[i])
		maxX = math.Max(maxX, x[i])
		minY = math.Min(minY, y[i])
		maxY = math.Max(maxY, y[i])
	}
	dx := maxX - minX
	dy := maxY - minY
	delta := math.Max(dx, dy)
	if delta == 0 {
		return nil, false
	}
	midX := (minX + maxX) / 2
	midY := (minY + maxY) / 2

	px := append(append([]float64(nil), x...), midX-20*delta, midX, midX+20*delta)
	py := append(append([]float64(nil), y...), midY-delta, midY+20*delta, midY-delta)
	superA, superB, superC := n, n+1, n+2
	super, ok := orientedDelaunayTriangle(superA, superB, superC, px, py)
	if !ok {
		return nil, false
	}
	triangles := []delaunayTriangle{super}

	for p := 0; p < n; p++ {
		bad := make([]bool, len(triangles))
		boundary := make(map[[2]int]int)
		for i, tri := range triangles {
			if pointInCircumcircle(px[p], py[p], tri, px, py) {
				bad[i] = true
				boundary[sortedEdge(tri.a, tri.b)]++
				boundary[sortedEdge(tri.b, tri.c)]++
				boundary[sortedEdge(tri.c, tri.a)]++
			}
		}

		kept := triangles[:0]
		for i, tri := range triangles {
			if !bad[i] {
				kept = append(kept, tri)
			}
		}
		triangles = kept

		for edge, count := range boundary {
			if count != 1 {
				continue
			}
			tri, ok := orientedDelaunayTriangle(edge[0], edge[1], p, px, py)
			if ok {
				triangles = append(triangles, tri)
			}
		}
	}

	out := make([][3]int, 0, len(triangles))
	for _, tri := range triangles {
		if tri.a >= n || tri.b >= n || tri.c >= n {
			continue
		}
		out = append(out, [3]int{tri.a, tri.b, tri.c})
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		for k := 0; k < 3; k++ {
			if a[k] != b[k] {
				return a[k] < b[k]
			}
		}
		return false
	})
	return out, len(out) > 0
}

// orientedDelaunayTriangle returns a triangle whose vertices are wound
// anticlockwise (positive signed area). It reports false for degenerate
// (near-collinear) triangles.
func orientedDelaunayTriangle(a, b, c int, x, y []float64) (delaunayTriangle, bool) {
	area := (x[b]-x[a])*(y[c]-y[a]) - (y[b]-y[a])*(x[c]-x[a])
	if math.Abs(area) < 1e-14 {
		return delaunayTriangle{}, false
	}
	if area < 0 {
		a, b = b, a
	}
	return delaunayTriangle{a: a, b: b, c: c}, true
}

// pointInCircumcircle reports whether (px,py) lies strictly inside the
// circumcircle of the (anticlockwise) triangle tri.
func pointInCircumcircle(px, py float64, tri delaunayTriangle, x, y []float64) bool {
	ax := x[tri.a] - px
	ay := y[tri.a] - py
	bx := x[tri.b] - px
	by := y[tri.b] - py
	cx := x[tri.c] - px
	cy := y[tri.c] - py

	det := (ax*ax+ay*ay)*(bx*cy-cx*by) -
		(bx*bx+by*by)*(ax*cy-cx*ay) +
		(cx*cx+cy*cy)*(ax*by-bx*ay)
	return det > 1e-12
}

// sortedEdge returns the undirected edge (a,b) as an ordered pair (min,max).
func sortedEdge(a, b int) [2]int {
	if a < b {
		return [2]int{a, b}
	}
	return [2]int{b, a}
}
