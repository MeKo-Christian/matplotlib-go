package geom

import "testing"

// Reference values produced by matplotlib 3.10.9 matplotlib.path.Path
// (unit_circle, unit_circle_righthalf, arc, wedge, unit_rectangle,
// unit_regular_polygon, unit_regular_star, unit_regular_asterisk).

func cmdSeq(p Path) []Cmd { return p.C }

func eqCmds(a, b []Cmd) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestUnitCircle(t *testing.T) {
	p := UnitCircle()
	if !p.Validate() {
		t.Fatalf("unit circle path failed validation")
	}
	wantCmds := []Cmd{MoveTo, CubicTo, CubicTo, CubicTo, CubicTo, CubicTo, CubicTo, CubicTo, CubicTo, ClosePath}
	if !eqCmds(cmdSeq(p), wantCmds) {
		t.Fatalf("unit circle cmds = %v, want %v", p.C, wantCmds)
	}
	want := []Pt{
		{0.0, -1.0},
		{0.2652031, -1.0},
		{0.5195798707848535, -0.8946336915882417},
		{0.7071067811865476, -0.7071067811865476},
		{0.8946336915882417, -0.5195798707848535},
		{1.0, -0.2652031},
		{1.0, 0.0},
		{1.0, 0.2652031},
		{0.8946336915882417, 0.5195798707848535},
		{0.7071067811865476, 0.7071067811865476},
		{0.5195798707848535, 0.8946336915882417},
		{0.2652031, 1.0},
		{0.0, 1.0},
		{-0.2652031, 1.0},
		{-0.5195798707848535, 0.8946336915882417},
		{-0.7071067811865476, 0.7071067811865476},
		{-0.8946336915882417, 0.5195798707848535},
		{-1.0, 0.2652031},
		{-1.0, 0.0},
		{-1.0, -0.2652031},
		{-0.8946336915882417, -0.5195798707848535},
		{-0.7071067811865476, -0.7071067811865476},
		{-0.5195798707848535, -0.8946336915882417},
		{-0.2652031, -1.0},
		{0.0, -1.0},
	}
	wantPts(t, "unit_circle", p.V, want)
}

func TestCircleScaleTranslate(t *testing.T) {
	c := Pt{X: 3, Y: -2}
	r := 5.0
	p := Circle(c, r)
	u := UnitCircle()
	if len(p.V) != len(u.V) {
		t.Fatalf("circle len = %d, want %d", len(p.V), len(u.V))
	}
	scaled := make([]Pt, len(u.V))
	for i, v := range u.V {
		scaled[i] = Pt{X: v.X*r + c.X, Y: v.Y*r + c.Y}
	}
	wantPts(t, "circle", p.V, scaled)
}

func TestUnitCircleRightHalf(t *testing.T) {
	p := UnitCircleRightHalf()
	if !p.Validate() {
		t.Fatalf("right-half path failed validation")
	}
	wantCmds := []Cmd{MoveTo, CubicTo, CubicTo, CubicTo, CubicTo, ClosePath}
	if !eqCmds(cmdSeq(p), wantCmds) {
		t.Fatalf("right-half cmds = %v, want %v", p.C, wantCmds)
	}
	want := []Pt{
		{0.0, -1.0},
		{0.2652031, -1.0},
		{0.5195798707848535, -0.8946336915882417},
		{0.7071067811865476, -0.7071067811865476},
		{0.8946336915882417, -0.5195798707848535},
		{1.0, -0.2652031},
		{1.0, 0.0},
		{1.0, 0.2652031},
		{0.8946336915882417, 0.5195798707848535},
		{0.7071067811865476, 0.7071067811865476},
		{0.5195798707848535, 0.8946336915882417},
		{0.2652031, 1.0},
		{0.0, 1.0},
	}
	wantPts(t, "unit_circle_righthalf", p.V, want)
}

func TestArc(t *testing.T) {
	p := Arc(0, 90, 0)
	if !p.Validate() {
		t.Fatalf("arc path failed validation")
	}
	wantCmds := []Cmd{MoveTo, CubicTo, CubicTo}
	if !eqCmds(cmdSeq(p), wantCmds) {
		t.Fatalf("arc cmds = %v, want %v", p.C, wantCmds)
	}
	want := []Pt{
		{1.0, 0.0},
		{1.0, 0.26511477349130247},
		{0.8945712353149831, 0.5196423270581119},
		{0.7071067811865476, 0.7071067811865475},
		{0.5196423270581121, 0.8945712353149831},
		{0.2651147734913025, 1.0},
		{6.123233995736766e-17, 1.0},
	}
	wantPts(t, "arc_0_90", p.V, want)
}

func TestArcSegmentCount(t *testing.T) {
	// n is auto-selected as 2**ceil((eta2-eta1)/(pi/2)): 170 degrees -> 4 segments.
	p := Arc(30, 200, 0)
	wantCmds := []Cmd{MoveTo, CubicTo, CubicTo, CubicTo, CubicTo}
	if !eqCmds(cmdSeq(p), wantCmds) {
		t.Fatalf("arc_30_200 cmds = %v, want %v", p.C, wantCmds)
	}
	want := []Pt{
		{0.8660254037844387, 0.49999999999999994},
		{0.7409990613186983, 0.7165519774351686},
		{0.5391852837035322, 0.8785246582077157},
		{0.3007057995042733, 0.9537169507482268},
		{0.06222631530501438, 1.028909243288738},
		{-0.1959935662178763, 1.0119846180859873},
		{-0.42261826174069933, 0.90630778703665},
		{-0.6492429572635223, 0.8006309559873129},
		{-0.8281885127696013, 0.6137019900227784},
		{-0.9238795325112867, 0.3826834323650899},
		{-1.019570552252972, 0.1516648747074014},
		{-1.0252156759251418, -0.10704748048785248},
		{-0.9396926207859084, -0.34202014332566866},
	}
	wantPts(t, "arc_30_200", p.V, want)
}

func TestArcExplicitN(t *testing.T) {
	p := Arc(0, 90, 1)
	wantCmds := []Cmd{MoveTo, CubicTo}
	if !eqCmds(cmdSeq(p), wantCmds) {
		t.Fatalf("arc n=1 cmds = %v, want %v", p.C, wantCmds)
	}
}

func TestWedge(t *testing.T) {
	p := Wedge(0, 90, 0)
	if !p.Validate() {
		t.Fatalf("wedge path failed validation")
	}
	wantCmds := []Cmd{MoveTo, LineTo, CubicTo, CubicTo, LineTo, ClosePath}
	if !eqCmds(cmdSeq(p), wantCmds) {
		t.Fatalf("wedge cmds = %v, want %v", p.C, wantCmds)
	}
	want := []Pt{
		{0.0, 0.0},
		{1.0, 0.0},
		{1.0, 0.26511477349130247},
		{0.8945712353149831, 0.5196423270581119},
		{0.7071067811865476, 0.7071067811865475},
		{0.5196423270581121, 0.8945712353149831},
		{0.2651147734913025, 1.0},
		{6.123233995736766e-17, 1.0},
		{0.0, 0.0},
	}
	wantPts(t, "wedge_0_90", p.V, want)
}

func TestEllipseBezier(t *testing.T) {
	p := EllipseBezier(Pt{}, 2, 1)
	if !p.Validate() {
		t.Fatalf("ellipse path failed validation")
	}
	const kx = 2 * BezierCircleKappa
	const ky = 1 * BezierCircleKappa
	want := []Pt{
		{2, 0},
		{2, ky},
		{kx, 1},
		{0, 1},
		{-kx, 1},
		{-2, ky},
		{-2, 0},
		{-2, -ky},
		{-kx, -1},
		{0, -1},
		{kx, -1},
		{2, -ky},
		{2, 0},
	}
	wantPts(t, "ellipse", p.V, want)

	if got := EllipseBezier(Pt{}, 0, 1); len(got.V) != 0 {
		t.Fatalf("ellipse with rx=0 should be empty, got %v", got.V)
	}
}

func TestUnitRectangle(t *testing.T) {
	p := UnitRectangle()
	wantCmds := []Cmd{MoveTo, LineTo, LineTo, LineTo, ClosePath}
	if !eqCmds(cmdSeq(p), wantCmds) {
		t.Fatalf("rectangle cmds = %v, want %v", p.C, wantCmds)
	}
	wantPts(t, "unit_rectangle", p.V, []Pt{{0, 0}, {1, 0}, {1, 1}, {0, 1}})
}

func TestUnitRegularPolygon(t *testing.T) {
	p := UnitRegularPolygon(5)
	wantCmds := []Cmd{MoveTo, LineTo, LineTo, LineTo, LineTo, ClosePath}
	if !eqCmds(cmdSeq(p), wantCmds) {
		t.Fatalf("polygon cmds = %v, want %v", p.C, wantCmds)
	}
	want := []Pt{
		{6.123233995736766e-17, 1.0},
		{-0.9510565162951535, 0.3090169943749475},
		{-0.5877852522924732, -0.8090169943749473},
		{0.5877852522924729, -0.8090169943749476},
		{0.9510565162951536, 0.3090169943749472},
	}
	wantPts(t, "unit_regular_polygon_5", p.V, want)

	if got := UnitRegularPolygon(2); len(got.V) != 0 {
		t.Fatalf("polygon with 2 vertices should be empty, got %v", got.V)
	}
}

func TestUnitRegularStar(t *testing.T) {
	p := UnitRegularStar(5, 0.5)
	want := []Pt{
		{6.123233995736766e-17, 1.0},
		{-0.2938926261462365, 0.4045084971874737},
		{-0.9510565162951535, 0.3090169943749475},
		{-0.4755282581475768, -0.15450849718747364},
		{-0.5877852522924732, -0.8090169943749473},
		{-9.184850993605148e-17, -0.5},
		{0.5877852522924729, -0.8090169943749476},
		{0.47552825814757677, -0.1545084971874738},
		{0.9510565162951536, 0.3090169943749472},
		{0.2938926261462367, 0.4045084971874736},
	}
	wantPts(t, "unit_regular_star_5", p.V, want)
}

func TestUnitRegularAsterisk(t *testing.T) {
	p := UnitRegularAsterisk(3)
	want := []Pt{
		{6.123233995736766e-17, 1.0},
		{0.0, 0.0},
		{-0.8660254037844388, -0.4999999999999997},
		{0.0, 0.0},
		{0.8660254037844384, -0.5000000000000004},
		{0.0, 0.0},
	}
	wantPts(t, "unit_regular_asterisk_3", p.V, want)
}
