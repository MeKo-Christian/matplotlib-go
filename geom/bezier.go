package geom

import "math"

// This file ports matplotlib's `lib/matplotlib/bezier.py` Bézier toolkit:
// De Casteljau subdivision, arc length, parallel/offset curves, and the small
// geometric helpers (normals, intersections, parallel tests) that fancy arrows
// and annotation connectors rely on. Keeping these in geom avoids the ad-hoc
// reimplementations that previously lived in core/.

// BezierSegment is a 2-D Bézier segment defined by its control points. The
// degree is one less than the number of control points (so two points describe
// a line, three a quadratic, four a cubic). Mirrors matplotlib's BezierSegment.
type BezierSegment struct {
	Points []Pt
}

// NewBezierSegment builds a segment from its control points.
func NewBezierSegment(points ...Pt) BezierSegment {
	return BezierSegment{Points: points}
}

// Degree is the polynomial degree, one less than the number of control points.
func (b BezierSegment) Degree() int { return len(b.Points) - 1 }

// PointAt evaluates the curve at parameter t in [0, 1] using De Casteljau's
// algorithm (numerically stable, matching matplotlib's point_at_t result).
func (b BezierSegment) PointAt(t float64) Pt {
	if len(b.Points) == 0 {
		return Pt{}
	}
	cur := append([]Pt(nil), b.Points...)
	for len(cur) > 1 {
		for i := 0; i < len(cur)-1; i++ {
			cur[i] = lerpPt(cur[i], cur[i+1], t)
		}
		cur = cur[:len(cur)-1]
	}
	return cur[0]
}

// ArcLength returns the length of the curve, computed by adaptive De Casteljau
// subdivision: a sub-segment is treated as straight once its control polygon
// and chord agree to within tolerance, and the per-level tolerance is halved so
// the accumulated error stays below the requested tolerance.
func (b BezierSegment) ArcLength(tolerance float64) float64 {
	if len(b.Points) < 2 {
		return 0
	}
	if tolerance <= 0 {
		tolerance = 1e-9
	}
	return arcLength(b.Points, tolerance)
}

func arcLength(pts []Pt, tol float64) float64 {
	poly := 0.0
	for i := 1; i < len(pts); i++ {
		poly += dist(pts[i-1], pts[i])
	}
	chord := dist(pts[0], pts[len(pts)-1])
	if poly-chord <= tol || len(pts) == 2 {
		return 0.5 * (poly + chord)
	}
	left, right := SplitDeCasteljau(pts, 0.5)
	return arcLength(left, tol/2) + arcLength(right, tol/2)
}

// PolynomialCoefficients returns the curve's coefficients C_j such that the
// curve equals sum_j C_j t^j. Follows matplotlib's formula
// C_j = comb(n, j) * sum_{i=0..j} (-1)^{i+j} comb(j, i) P_i.
func (b BezierSegment) PolynomialCoefficients() []Pt {
	n := b.Degree()
	if n < 0 {
		return nil
	}
	out := make([]Pt, n+1)
	for j := 0; j <= n; j++ {
		var acc Pt
		for i := 0; i <= j; i++ {
			sign := 1.0
			if (i+j)%2 != 0 {
				sign = -1.0
			}
			w := sign * binomial(j, i)
			acc.X += w * b.Points[i].X
			acc.Y += w * b.Points[i].Y
		}
		c := binomial(n, j)
		out[j] = Pt{X: c * acc.X, Y: c * acc.Y}
	}
	return out
}

// AxisAlignedExtrema returns the interior extrema of the curve: the dimensions
// (0 for x, 1 for y) and parameter values t in [0, 1] where a partial
// derivative vanishes. Matches matplotlib's axis_aligned_extrema for curves up
// to cubic degree (the only degrees matplotlib itself produces); higher-degree
// derivatives are not solved.
func (b BezierSegment) AxisAlignedExtrema() (dims []int, roots []float64) {
	n := b.Degree()
	if n <= 1 {
		return nil, nil
	}
	c := b.PolynomialCoefficients()
	for dim := range 2 {
		// Derivative coefficients (ascending powers): d/dt sum C_j t^j.
		dc := make([]float64, n)
		for j := 1; j <= n; j++ {
			v := c[j].X
			if dim == 1 {
				v = c[j].Y
			}
			dc[j-1] = float64(j) * v
		}
		for _, r := range realRootsInUnit(dc) {
			dims = append(dims, dim)
			roots = append(roots, r)
		}
	}
	return dims, roots
}

// SplitDeCasteljau splits a Bézier segment given by control points beta into two
// segments meeting at parameter t, returning the control points of each. Ports
// matplotlib's split_de_casteljau.
func SplitDeCasteljau(beta []Pt, t float64) (left, right []Pt) {
	cur := append([]Pt(nil), beta...)
	left = []Pt{cur[0]}
	right = []Pt{cur[len(cur)-1]}
	for len(cur) > 1 {
		next := make([]Pt, len(cur)-1)
		for i := range next {
			next[i] = lerpPt(cur[i], cur[i+1], t)
		}
		cur = next
		left = append(left, cur[0])
		right = append(right, cur[len(cur)-1])
	}
	// right is collected outermost-first; matplotlib reverses it.
	for i, j := 0, len(right)-1; i < j; i, j = i+1, j-1 {
		right[i], right[j] = right[j], right[i]
	}
	return left, right
}

// CosSin returns the cosine and sine of the angle of the line from p0 to p1.
// Coincident points return (0, 0). Ports get_cos_sin.
func CosSin(p0, p1 Pt) (cos, sin float64) {
	dx, dy := p1.X-p0.X, p1.Y-p0.Y
	d := math.Hypot(dx, dy)
	if d == 0 {
		return 0, 0
	}
	return dx / d, dy / d
}

// NormalPoints returns the two points a perpendicular distance length from
// (c) along the line through c at angle t (given by its cosine/sine). The first
// return is the "left" point, the second the "right". Ports get_normal_points.
func NormalPoints(c Pt, cosT, sinT, length float64) (left, right Pt) {
	if length == 0 {
		return c, c
	}
	left = Pt{X: length*sinT + c.X, Y: length*(-cosT) + c.Y}
	right = Pt{X: length*(-sinT) + c.X, Y: length*cosT + c.Y}
	return left, right
}

// Intersection returns the intersection of the line through c1 at angle t1
// and the line through c2 at angle t2 (each angle given by its cosine/sine).
// The boolean is false when the lines are (near-)parallel and do not intersect.
// Ports get_intersection (which raises ValueError in that case).
func Intersection(c1 Pt, cosT1, sinT1 float64, c2 Pt, cosT2, sinT2 float64) (Pt, bool) {
	line1RHS := sinT1*c1.X - cosT1*c1.Y
	line2RHS := sinT2*c2.X - cosT2*c2.Y

	a, b := sinT1, -cosT1
	cc, d := sinT2, -cosT2

	adBC := a*d - b*cc
	if math.Abs(adBC) < 1e-12 {
		return Pt{}, false
	}

	aInv := d / adBC
	bInv := -b / adBC
	cInv := -cc / adBC
	dInv := a / adBC

	return Pt{
		X: aInv*line1RHS + bInv*line2RHS,
		Y: cInv*line1RHS + dInv*line2RHS,
	}, true
}

// CheckIfParallel reports whether two direction vectors are parallel: 1 if they
// point the same way, -1 if opposite, 0 otherwise. tolerance is the angular
// tolerance in radians. Ports check_if_parallel.
func CheckIfParallel(d1, d2 Pt, tolerance float64) int {
	// matplotlib uses arctan2(dx, dy) (x first) here.
	theta1 := math.Atan2(d1.X, d1.Y)
	theta2 := math.Atan2(d2.X, d2.Y)
	dtheta := math.Abs(theta1 - theta2)
	switch {
	case dtheta < tolerance:
		return 1
	case math.Abs(dtheta-math.Pi) < tolerance:
		return -1
	default:
		return 0
	}
}

// FindControlPoints returns the control points of the quadratic Bézier that
// passes through c1, mm, c2 at parameters 0, 0.5, 1. Ports find_control_points.
func FindControlPoints(c1, mm, c2 Pt) [3]Pt {
	cm := Pt{
		X: 0.5 * (4*mm.X - (c1.X + c2.X)),
		Y: 0.5 * (4*mm.Y - (c1.Y + c2.Y)),
	}
	return [3]Pt{c1, cm, c2}
}

// parallelTolerance is matplotlib's default angular tolerance for the parallel
// test inside get_parallels.
const parallelTolerance = 1e-5

// Parallels returns the control points of two quadratic Bézier curves
// roughly parallel to the given quadratic (bezier2), offset by width on either
// side. Ports get_parallels.
func Parallels(bezier2 [3]Pt, width float64) (left, right [3]Pt) {
	c1, cm, c2 := bezier2[0], bezier2[1], bezier2[2]

	parallel := CheckIfParallel(
		Pt{X: c1.X - cm.X, Y: c1.Y - cm.Y},
		Pt{X: cm.X - c2.X, Y: cm.Y - c2.Y},
		parallelTolerance,
	)

	var cosT1, sinT1, cosT2, sinT2 float64
	if parallel == -1 {
		// Lines do not intersect; fall back to a straight line.
		cosT1, sinT1 = CosSin(c1, c2)
		cosT2, sinT2 = cosT1, sinT1
	} else {
		cosT1, sinT1 = CosSin(c1, cm)
		cosT2, sinT2 = CosSin(cm, c2)
	}

	c1Left, c1Right := NormalPoints(c1, cosT1, sinT1, width)
	c2Left, c2Right := NormalPoints(c2, cosT2, sinT2, width)

	cmLeft, okL := Intersection(c1Left, cosT1, sinT1, c2Left, cosT2, sinT2)
	cmRight, okR := Intersection(c1Right, cosT1, sinT1, c2Right, cosT2, sinT2)
	if !okL || !okR {
		// Near-straight line: use midpoints (matplotlib's except branch).
		cmLeft = midpoint(c1Left, c2Left)
		cmRight = midpoint(c1Right, c2Right)
	}

	left = [3]Pt{c1Left, cmLeft, c2Left}
	right = [3]Pt{c1Right, cmRight, c2Right}
	return left, right
}

// MakeWedgedBezier2 is like Parallels but produces a wedge: the offset width
// is scaled by w1, wm, w2 at the start, middle, and end. Ports
// make_wedged_bezier2.
func MakeWedgedBezier2(bezier2 [3]Pt, width, w1, wm, w2 float64) (left, right [3]Pt) {
	c1, cm, c3 := bezier2[0], bezier2[1], bezier2[2]

	cosT1, sinT1 := CosSin(c1, cm)
	cosT2, sinT2 := CosSin(cm, c3)

	c1Left, c1Right := NormalPoints(c1, cosT1, sinT1, width*w1)
	c3Left, c3Right := NormalPoints(c3, cosT2, sinT2, width*w2)

	c12 := midpoint(c1, cm)
	c23 := midpoint(cm, c3)
	c123 := midpoint(c12, c23)

	cosT123, sinT123 := CosSin(c12, c23)
	c123Left, c123Right := NormalPoints(c123, cosT123, sinT123, width*wm)

	left = FindControlPoints(c1Left, c123Left, c3Left)
	right = FindControlPoints(c1Right, c123Right, c3Right)
	return left, right
}

// InsideCircle returns a predicate reporting whether a point lies strictly
// inside the circle with centre (cx, cy) and radius r. Ports inside_circle.
func InsideCircle(cx, cy, r float64) func(Pt) bool {
	r2 := r * r
	return func(p Pt) bool {
		dx, dy := p.X-cx, p.Y-cy
		return dx*dx+dy*dy < r2
	}
}

// FindBezierTIntersectingWithClosedPath bisects to find parameters t0 <= t <= t1
// bracketing where the Bézier curve crosses the boundary of a closed region
// (described by the inside predicate). The search needs one endpoint inside and
// the other outside. The boolean is false when both endpoints lie on the same
// side. Ports find_bezier_t_intersecting_with_closedpath.
func FindBezierTIntersectingWithClosedPath(
	pointAt func(float64) Pt, inside func(Pt) bool, t0, t1, tolerance float64,
) (float64, float64, bool) {
	start := pointAt(t0)
	end := pointAt(t1)

	startInside := inside(start)
	endInside := inside(end)
	if startInside == endInside && start != end {
		return t0, t1, false
	}

	for {
		if math.Hypot(start.X-end.X, start.Y-end.Y) < tolerance {
			return t0, t1, true
		}
		middleT := 0.5 * (t0 + t1)
		middle := pointAt(middleT)
		middleInside := inside(middle)

		if startInside != middleInside {
			t1 = middleT
			if end == middle {
				return t0, t1, true
			}
			end = middle
		} else {
			t0 = middleT
			if start == middle {
				return t0, t1, true
			}
			start = middle
			startInside = middleInside
		}
	}
}

// SplitBezierIntersectingWithClosedPath splits the Bézier segment (control
// points bezier) into two halves at its crossing of the closed region described
// by inside. When the curve does not cross the boundary, the whole segment is
// returned as left and right is nil. Ports
// split_bezier_intersecting_with_closedpath.
func SplitBezierIntersectingWithClosedPath(
	bezier []Pt, inside func(Pt) bool, tolerance float64,
) (left, right []Pt) {
	if tolerance <= 0 {
		tolerance = 0.01
	}
	seg := NewBezierSegment(bezier...)
	t0, t1, ok := FindBezierTIntersectingWithClosedPath(seg.PointAt, inside, 0, 1, tolerance)
	if !ok {
		return append([]Pt(nil), bezier...), nil
	}
	return SplitDeCasteljau(bezier, (t0+t1)/2)
}

// --- small internal helpers ---

func midpoint(a, b Pt) Pt {
	return Pt{X: 0.5 * (a.X + b.X), Y: 0.5 * (a.Y + b.Y)}
}

func dist(a, b Pt) float64 {
	return math.Hypot(a.X-b.X, a.Y-b.Y)
}

// binomial returns the binomial coefficient C(n, k) as a float64.
func binomial(n, k int) float64 {
	if k < 0 || k > n {
		return 0
	}
	if k > n-k {
		k = n - k
	}
	res := 1.0
	for i := range k {
		res = res * float64(n-i) / float64(i+1)
	}
	return res
}

// realRootsInUnit returns the real roots in [0, 1] of the polynomial with the
// given ascending-power coefficients. Solves up to quadratic (degree-3 Bézier
// derivatives); higher degrees are not produced by matplotlib and return nil.
func realRootsInUnit(coef []float64) []float64 {
	m := len(coef) - 1
	for m > 0 && math.Abs(coef[m]) < 1e-14 {
		m--
	}
	switch m {
	case 1:
		r := -coef[0] / coef[1]
		if r >= 0 && r <= 1 {
			return []float64{r}
		}
		return nil
	case 2:
		a, b, c := coef[2], coef[1], coef[0]
		disc := b*b - 4*a*c
		if disc < 0 {
			return nil
		}
		s := math.Sqrt(disc)
		var out []float64
		for _, r := range [2]float64{(-b - s) / (2 * a), (-b + s) / (2 * a)} {
			if r >= 0 && r <= 1 {
				out = append(out, r)
			}
		}
		if len(out) == 2 && out[0] > out[1] {
			out[0], out[1] = out[1], out[0]
		}
		return out
	default:
		return nil
	}
}
