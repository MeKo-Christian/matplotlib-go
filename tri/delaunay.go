package tri

import qhull "github.com/MeKo-Christian/qhull-go"

// delaunayTriangles computes a Delaunay triangulation of the supplied points,
// delegating to the qhull package. The triangle set (connectivity) is the true
// Delaunay triangulation, which for general position is identical to matplotlib's
// Qhull backend. For cocircular inputs the triangulation is non-unique;
// qhull.Delaunay resolves the diagonal to match Qhull's where the computed
// construction order determines it, and otherwise returns a valid Delaunay
// triangulation (see the qhull package docs).
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
