package tri

// TriFinder locates the triangle containing a query point.
type TriFinder interface {
	// Find returns the index of the triangle containing (x, y), or -1 if the
	// point lies outside the triangulation (or only inside masked triangles).
	Find(x, y float64) int
}

// TriFinder returns the default point-location structure for the
// triangulation, a [TrapezoidMapTriFinder]. It mirrors matplotlib's
// Triangulation.get_trifinder.
func (t Triangulation) TriFinder() TriFinder {
	return NewTrapezoidMapTriFinder(t)
}

// BruteForceTriFinder locates points by testing every (unmasked) triangle. It
// is O(n) per query but trivially correct, and serves as the reference oracle
// for the trapezoid-map finder.
type BruteForceTriFinder struct {
	tri Triangulation
}

// NewBruteForceTriFinder returns a brute-force point locator for t.
func NewBruteForceTriFinder(t Triangulation) *BruteForceTriFinder {
	return &BruteForceTriFinder{tri: t}
}

// Find returns the index of the first unmasked triangle containing (x, y), or
// -1 if none.
func (f *BruteForceTriFinder) Find(x, y float64) int {
	for i, tr := range f.tri.Triangles {
		if f.tri.Masked(i) {
			continue
		}
		if pointInTriangle(x, y, f.tri, tr) {
			return i
		}
	}
	return -1
}

// pointInTriangle reports whether (x, y) lies within the triangle (inclusive of
// edges), using barycentric sign tests. Works for either winding order.
func pointInTriangle(x, y float64, t Triangulation, tri [3]int) bool {
	ax, ay := t.X[tri[0]], t.Y[tri[0]]
	bx, by := t.X[tri[1]], t.Y[tri[1]]
	cx, cy := t.X[tri[2]], t.Y[tri[2]]

	d1 := edgeSign(x, y, ax, ay, bx, by)
	d2 := edgeSign(x, y, bx, by, cx, cy)
	d3 := edgeSign(x, y, cx, cy, ax, ay)

	hasNeg := d1 < 0 || d2 < 0 || d3 < 0
	hasPos := d1 > 0 || d2 > 0 || d3 > 0
	// Inside if all signs agree (allowing zeros on edges).
	return !hasNeg || !hasPos
}

// edgeSign returns the cross product (p - a) x (b - a), whose sign indicates
// which side of edge a->b the point p lies on.
func edgeSign(px, py, ax, ay, bx, by float64) float64 {
	return (px-bx)*(ay-by) - (ax-bx)*(py-by)
}
