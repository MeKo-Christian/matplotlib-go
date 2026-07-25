package geom

import (
	"math"
	"testing"
)

const bezEps = 1e-9

func approxF(a, b float64) bool { return math.Abs(a-b) <= bezEps }

func approxPtEq(a, b Pt) bool { return approxF(a.X, b.X) && approxF(a.Y, b.Y) }

func wantPts(t *testing.T, label string, got, want []Pt) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: len = %d, want %d (%v)", label, len(got), len(want), got)
	}
	for i := range want {
		if !approxPtEq(got[i], want[i]) {
			t.Fatalf("%s[%d] = %v, want %v", label, i, got[i], want[i])
		}
	}
}

// Reference values produced by matplotlib 3.10.9 matplotlib.bezier.

func TestBezierSegmentPointAt(t *testing.T) {
	seg := NewBezierSegment(Pt{0, 0}, Pt{1, 2}, Pt{3, 3}, Pt{4, 0})
	if seg.Degree() != 3 {
		t.Fatalf("degree = %d, want 3", seg.Degree())
	}
	cases := []struct {
		t    float64
		want Pt
	}{
		{0.0, Pt{0, 0}},
		{0.25, Pt{0.90625, 1.265625}},
		{0.5, Pt{2.0, 1.875}},
		{0.75, Pt{3.09375, 1.546875}},
		{1.0, Pt{4, 0}},
	}
	for _, c := range cases {
		if got := seg.PointAt(c.t); !approxPtEq(got, c.want) {
			t.Errorf("PointAt(%g) = %v, want %v", c.t, got, c.want)
		}
	}
}

func TestBezierSegmentArcLength(t *testing.T) {
	cubic := NewBezierSegment(Pt{0, 0}, Pt{1, 2}, Pt{3, 3}, Pt{4, 0})
	if got := cubic.ArcLength(1e-9); math.Abs(got-5.829927224626297) > 1e-6 {
		t.Errorf("cubic arc length = %.12f, want 5.829927224626297", got)
	}
	quad := NewBezierSegment(Pt{0, 0}, Pt{2, 4}, Pt{6, 0})
	if got := quad.ArcLength(1e-9); math.Abs(got-7.509272824633937) > 1e-6 {
		t.Errorf("quad arc length = %.12f, want 7.509272824633937", got)
	}
	line := NewBezierSegment(Pt{0, 0}, Pt{3, 4})
	if got := line.ArcLength(1e-9); math.Abs(got-5.0) > 1e-9 {
		t.Errorf("line arc length = %.12f, want 5", got)
	}
}

func TestBezierSegmentPolynomialCoefficients(t *testing.T) {
	seg := NewBezierSegment(Pt{0, 0}, Pt{1, 2}, Pt{3, 3}, Pt{4, 0})
	want := []Pt{{0, 0}, {3, 6}, {3, -3}, {-2, -3}}
	wantPts(t, "polycoeffs", seg.PolynomialCoefficients(), want)
}

func TestBezierSegmentAxisAlignedExtrema(t *testing.T) {
	seg := NewBezierSegment(Pt{0, 0}, Pt{1, 2}, Pt{3, 3}, Pt{4, 0})
	dims, roots := seg.AxisAlignedExtrema()
	if len(dims) != 1 || dims[0] != 1 {
		t.Fatalf("dims = %v, want [1]", dims)
	}
	if len(roots) != 1 || math.Abs(roots[0]-0.5485837703548636) > 1e-9 {
		t.Fatalf("roots = %v, want [0.5485837703548636]", roots)
	}
}

func TestSplitDeCasteljauCubic(t *testing.T) {
	cps := []Pt{{0, 0}, {1, 2}, {3, 3}, {4, 0}}
	left, right := SplitDeCasteljau(cps, 0.5)
	wantPts(t, "left", left, []Pt{{0, 0}, {0.5, 1}, {1.25, 1.75}, {2, 1.875}})
	wantPts(t, "right", right, []Pt{{2, 1.875}, {2.75, 2}, {3.5, 1.5}, {4, 0}})
}

func TestSplitDeCasteljauQuadratic(t *testing.T) {
	cps := []Pt{{0, 0}, {2, 4}, {6, 0}}
	left, right := SplitDeCasteljau(cps, 0.3)
	wantPts(t, "left", left, []Pt{{0, 0}, {0.6, 1.2}, {1.38, 1.68}})
	wantPts(t, "right", right, []Pt{{1.38, 1.68}, {3.2, 2.8}, {6, 0}})
}

func TestGetParallels(t *testing.T) {
	left, right := Parallels([3]Pt{{0, 0}, {2, 4}, {6, 0}}, 0.5)
	wantPts(t, "left", left[:], []Pt{
		{0.4472135954999579, -0.22360679774997896},
		{2.136975735854449, 3.155917482959003},
		{5.646446609406726, -0.35355339059327373},
	})
	wantPts(t, "right", right[:], []Pt{
		{-0.4472135954999579, 0.22360679774997896},
		{1.8630242641455512, 4.844082517040997},
		{6.353553390593274, 0.35355339059327373},
	})
}

func TestMakeWedgedBezier2(t *testing.T) {
	left, right := MakeWedgedBezier2([3]Pt{{0, 0}, {2, 4}, {6, 0}}, 0.5, 1.0, 0.5, 0.0)
	wantPts(t, "left", left[:], []Pt{
		{0.4472135954999579, -0.22360679774997896},
		{1.7763932022500208, 3.6118033988749896},
		{6, 0},
	})
	wantPts(t, "right", right[:], []Pt{
		{-0.4472135954999579, 0.22360679774997896},
		{2.223606797749979, 4.388196601125011},
		{6, 0},
	})
}

func TestGetIntersection(t *testing.T) {
	// horizontal line through origin, vertical line through (0,2) -> (0,0)
	p, ok := Intersection(Pt{0, 0}, math.Cos(0), math.Sin(0), Pt{0, 2}, math.Cos(math.Pi/2), math.Sin(math.Pi/2))
	if !ok {
		t.Fatal("expected intersection")
	}
	if !approxPtEq(p, Pt{0, 0}) {
		t.Fatalf("intersection = %v, want (0,0)", p)
	}
	// parallel lines do not intersect
	if _, ok := Intersection(Pt{0, 0}, 1, 0, Pt{0, 1}, 1, 0); ok {
		t.Fatal("parallel lines should not intersect")
	}
}

func TestGetNormalPoints(t *testing.T) {
	left, right := NormalPoints(Pt{1, 1}, math.Cos(math.Pi/4), math.Sin(math.Pi/4), 2.0)
	if !approxPtEq(left, Pt{2.414213562373095, -0.41421356237309515}) {
		t.Fatalf("left = %v", left)
	}
	if !approxPtEq(right, Pt{-0.4142135623730949, 2.414213562373095}) {
		t.Fatalf("right = %v", right)
	}
	// zero length returns the center twice.
	l0, r0 := NormalPoints(Pt{3, 4}, 1, 0, 0)
	if l0 != (Pt{3, 4}) || r0 != (Pt{3, 4}) {
		t.Fatalf("zero length = %v,%v", l0, r0)
	}
}

func TestGetCosSin(t *testing.T) {
	cos, sin := CosSin(Pt{0, 0}, Pt{3, 4})
	if !approxF(cos, 0.6) || !approxF(sin, 0.8) {
		t.Fatalf("cos,sin = %g,%g want 0.6,0.8", cos, sin)
	}
	// coincident points return 0,0.
	if c, s := CosSin(Pt{1, 1}, Pt{1, 1}); c != 0 || s != 0 {
		t.Fatalf("coincident = %g,%g", c, s)
	}
}

func TestCheckIfParallel(t *testing.T) {
	if got := CheckIfParallel(Pt{1, 1}, Pt{2, 2}, 1e-5); got != 1 {
		t.Errorf("same direction = %d, want 1", got)
	}
	if got := CheckIfParallel(Pt{1, 1}, Pt{-2, -2}, 1e-5); got != -1 {
		t.Errorf("opposite direction = %d, want -1", got)
	}
	if got := CheckIfParallel(Pt{1, 0}, Pt{0, 1}, 1e-5); got != 0 {
		t.Errorf("perpendicular = %d, want 0", got)
	}
}

func TestFindControlPoints(t *testing.T) {
	got := FindControlPoints(Pt{0, 0}, Pt{2, 4}, Pt{6, 0})
	wantPts(t, "find_cp", got[:], []Pt{{0, 0}, {1, 8}, {6, 0}})
}

func TestInsideCircle(t *testing.T) {
	f := InsideCircle(0, 0, 5)
	if !f(Pt{3, 3}) {
		t.Error("(3,3) should be inside radius 5")
	}
	if f(Pt{4, 4}) {
		t.Error("(4,4) should be outside radius 5")
	}
	if f(Pt{5, 0}) {
		t.Error("boundary point should be excluded (strict <)")
	}
}

func TestSplitBezierIntersectingWithClosedPath(t *testing.T) {
	// A horizontal quadratic from (0,0) (inside the unit circle) out to (2,0)
	// (outside it); split where it leaves the circle, near (1,0).
	bezier := []Pt{{0, 0}, {1, 0}, {2, 0}}
	left, right := SplitBezierIntersectingWithClosedPath(bezier, InsideCircle(0, 0, 1), 0.001)
	if len(left) != 3 || len(right) != 3 {
		t.Fatalf("expected quadratic halves, got %d/%d", len(left), len(right))
	}
	// Join point continuity: last of left == first of right.
	if !approxPtEq(left[len(left)-1], right[0]) {
		t.Fatalf("discontinuous split: %v vs %v", left[len(left)-1], right[0])
	}
	// Split crosses the circle boundary near x=1.
	if math.Abs(right[0].X-1) > 0.01 || math.Abs(right[0].Y) > 0.01 {
		t.Fatalf("split point = %v, want ~(1,0)", right[0])
	}
	// No-intersection case: a curve entirely inside returns the whole as left.
	if l, r := SplitBezierIntersectingWithClosedPath(
		[]Pt{{0, 0}, {0.1, 0}, {0.2, 0}}, InsideCircle(0, 0, 1), 0.001,
	); len(l) != 3 || r != nil {
		t.Fatalf("non-crossing split = %d/%v, want whole/nil", len(l), r)
	}
}
