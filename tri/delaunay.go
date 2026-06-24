package tri

import "github.com/cwbudde/matplotlib-go/tri/qhull"

// delaunayTriangles computes a Delaunay triangulation of the supplied points,
// delegating to the qhull package's robust exact-predicate engine. The triangle
// set (connectivity) is the true Delaunay triangulation, which for general
// position is identical to matplotlib's Qhull backend. For cocircular inputs the
// triangulation is non-unique and the chosen diagonal may differ from Qhull's
// (it is always a valid Delaunay triangulation); see the qhull package docs.
//
// Triangles are wound anticlockwise and returned in a deterministic
// (lexicographic) order.
func delaunayTriangles(x, y []float64) ([][3]int, bool) {
	tris, _, err := qhull.Delaunay(x, y)
	if err != nil || len(tris) == 0 {
		return nil, false
	}
	return tris, true
}

// sortedEdge returns the undirected edge (a,b) as an ordered pair (min,max).
func sortedEdge(a, b int) [2]int {
	if a < b {
		return [2]int{a, b}
	}
	return [2]int{b, a}
}
