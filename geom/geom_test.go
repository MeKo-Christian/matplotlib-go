package geom

import (
	"math"
	"math/rand"
	"testing"
)

func TestRectBasics(t *testing.T) {
	r := Rect{Min: Pt{0, 0}, Max: Pt{10, 5}}
	if r.W() != 10 || r.H() != 5 {
		t.Fatalf("W/H mismatch: got %v/%v", r.W(), r.H())
	}
	if r.Empty() {
		t.Fatalf("rect should not be empty: %+v", r)
	}
	if !r.Intersect(Rect{Min: Pt{20, 20}, Max: Pt{21, 21}}).Empty() {
		t.Fatal("disjoint intersection should be empty")
	}

	// Contains: max exclusive
	if !r.Contains(Pt{0, 0}) || !r.Contains(Pt{9.999, 4.999}) {
		t.Fatalf("expected points inside")
	}
	if r.Contains(Pt{10, 0}) || r.Contains(Pt{0, 5}) {
		t.Fatalf("max edge should be exclusive")
	}
	if !r.ContainsInclusive(Pt{10, 5}) {
		t.Fatal("inclusive contains should include max corner")
	}

	// Inflate
	r2 := r.Inflate(1, 2)
	if r2.Min.X != -1 || r2.Min.Y != -2 || r2.Max.X != 11 || r2.Max.Y != 7 {
		t.Fatalf("inflate mismatch: %+v", r2)
	}
	if padded := r.Padded(2); padded != (Rect{Min: Pt{-2, -2}, Max: Pt{12, 7}}) {
		t.Fatalf("padded rect = %+v", padded)
	}
	if expanded := r.Expanded(2, 0.5); expanded != (Rect{Min: Pt{-5, 1.25}, Max: Pt{15, 3.75}}) {
		t.Fatalf("expanded rect = %+v", expanded)
	}
	if translated := r.Translated(2, -3); translated != (Rect{Min: Pt{2, -3}, Max: Pt{12, 2}}) {
		t.Fatalf("translated rect = %+v", translated)
	}

	// Intersect
	a := Rect{Min: Pt{0, 0}, Max: Pt{5, 5}}
	b := Rect{Min: Pt{3, 2}, Max: Pt{8, 4}}
	x := a.Intersect(b)
	exp := Rect{Min: Pt{3, 2}, Max: Pt{5, 4}}
	if x != exp {
		t.Fatalf("intersection mismatch: got %+v want %+v", x, exp)
	}

	// Disjoint -> empty
	c := Rect{Min: Pt{10, 10}, Max: Pt{12, 12}}
	e := a.Intersect(c)
	if e.W() != 0 || e.H() != 0 {
		t.Fatalf("expected empty intersection, got %+v", e)
	}
}

func TestRectUnionAndTransforms(t *testing.T) {
	a := Rect{Min: Pt{1, 2}, Max: Pt{3, 4}}
	b := Rect{Min: Pt{-2, 3}, Max: Pt{2, 8}}
	if got, want := a.Union(b), (Rect{Min: Pt{-2, 2}, Max: Pt{3, 8}}); got != want {
		t.Fatalf("union = %+v, want %+v", got, want)
	}

	union, ok := UnionRects(Rect{}, a, b)
	if !ok || union != (Rect{Min: Pt{-2, 2}, Max: Pt{3, 8}}) {
		t.Fatalf("UnionRects = %+v ok=%v", union, ok)
	}
	if _, ok := UnionRects(Rect{}); ok {
		t.Fatal("UnionRects of empty rectangles should report ok=false")
	}

	fromPoints, ok := RectFromPoints(Pt{3, 1}, Pt{-1, 4}, Pt{2, -2})
	if !ok || fromPoints != (Rect{Min: Pt{-1, -2}, Max: Pt{3, 4}}) {
		t.Fatalf("RectFromPoints = %+v ok=%v", fromPoints, ok)
	}
	if _, ok := RectFromPoints(); ok {
		t.Fatal("RectFromPoints without points should report ok=false")
	}

	transformed := a.Transformed(Affine{A: 2, D: -1, E: 5, F: 10})
	if transformed != (Rect{Min: Pt{7, 6}, Max: Pt{11, 8}}) {
		t.Fatalf("transformed rect = %+v", transformed)
	}
	inverse, ok := transformed.InverseTransformed(Affine{A: 2, D: -1, E: 5, F: 10})
	if !ok || inverse != a {
		t.Fatalf("inverse transformed rect = %+v ok=%v, want %+v", inverse, ok, a)
	}
	if _, ok := a.InverseTransformed(Affine{}); ok {
		t.Fatal("singular inverse transform should fail")
	}
}

func TestRectNullAccumulation(t *testing.T) {
	null := NullRect()
	if !null.Null() || !null.Empty() {
		t.Fatalf("NullRect = %+v, want null and empty", null)
	}
	if !math.IsInf(float64(null.Min.X), 1) || !math.IsInf(float64(null.Max.X), -1) {
		t.Fatalf("NullRect should use infinite sentinels, got %+v", null)
	}

	withPoint := null.AddPoint(Pt{3, -2})
	if withPoint.Null() || withPoint != (Rect{Min: Pt{3, -2}, Max: Pt{3, -2}}) {
		t.Fatalf("NullRect.AddPoint = %+v", withPoint)
	}
	grown := withPoint.AddPoint(Pt{-1, 5})
	if grown != (Rect{Min: Pt{-1, -2}, Max: Pt{3, 5}}) {
		t.Fatalf("AddPoint grown rect = %+v", grown)
	}

	rect := Rect{Min: Pt{10, 10}, Max: Pt{20, 20}}
	if got := null.AddRect(rect); got != rect {
		t.Fatalf("NullRect.AddRect = %+v, want %+v", got, rect)
	}
	if got := rect.AddRect(null); got != rect {
		t.Fatalf("Rect.AddRect(NullRect) = %+v, want %+v", got, rect)
	}
	if got := null.Union(rect); got != rect {
		t.Fatalf("NullRect.Union = %+v, want %+v", got, rect)
	}
}

func TestRectAnchored(t *testing.T) {
	container := Rect{Min: Pt{10, 20}, Max: Pt{110, 70}}
	size := Pt{20, 10}
	cases := []struct {
		anchor string
		want   Rect
	}{
		{anchor: "C", want: Rect{Min: Pt{50, 40}, Max: Pt{70, 50}}},
		{anchor: "SW", want: Rect{Min: Pt{10, 20}, Max: Pt{30, 30}}},
		{anchor: "NE", want: Rect{Min: Pt{90, 60}, Max: Pt{110, 70}}},
		{anchor: "upper right", want: Rect{Min: Pt{90, 60}, Max: Pt{110, 70}}},
		{anchor: "center left", want: Rect{Min: Pt{10, 40}, Max: Pt{30, 50}}},
	}
	for _, tc := range cases {
		got, ok := container.Anchored(size, tc.anchor)
		if !ok || got != tc.want {
			t.Fatalf("Anchored(%q) = %+v ok=%v, want %+v", tc.anchor, got, ok, tc.want)
		}
	}
	if _, ok := container.Anchored(size, "outside"); ok {
		t.Fatal("unknown anchor should report ok=false")
	}
}

func TestAffineBasicsAndInvert(t *testing.T) {
	id := Identity()
	p := Pt{3, 4}
	if id.Apply(p) != p {
		t.Fatalf("identity should not change point")
	}

	// translation then scale
	s := Affine{A: 2, D: 3}                // scale x2, y3
	tr := Affine{A: 1, D: 1, E: 10, F: -5} // translate
	comb := tr.Mul(s)                      // apply s, then tr
	got := comb.Apply(Pt{1, 2})
	want := Pt{X: 2*1 + 10, Y: 3*2 - 5}
	if got != want {
		t.Fatalf("compose/apply mismatch: got %+v want %+v", got, want)
	}

	inv, ok := comb.Invert()
	if !ok {
		t.Fatalf("expected invertible")
	}
	back := inv.Mul(comb)
	q := back.Apply(Pt{7.3, -2.1})
	if !approxPt(q, Pt{7.3, -2.1}, 1e-12) {
		t.Fatalf("inverse*matrix not identity: got %+v", q)
	}
}

func TestAffineRandomInvertibility(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for i := 0; i < 200; i++ {
		// Ensure invertible linear part by avoiding near-zero det
		var m Affine
		for {
			m = Affine{
				A: r.Float64()*4 - 2,
				B: r.Float64()*4 - 2,
				C: r.Float64()*4 - 2,
				D: r.Float64()*4 - 2,
				E: r.Float64()*20 - 10,
				F: r.Float64()*20 - 10,
			}
			det := m.A*m.D - m.C*m.B
			if det > 1e-6 || det < -1e-6 { // not near singular
				break
			}
		}
		inv, ok := m.Invert()
		if !ok {
			t.Fatalf("expected invertible")
		}
		// random point
		p := Pt{r.Float64()*10 - 5, r.Float64()*10 - 5}
		got := inv.Apply(m.Apply(p))
		if !approxPt(got, p, 1e-9) {
			t.Fatalf("roundtrip mismatch: got %+v want %+v", got, p)
		}
	}
}

func TestPathValidate(t *testing.T) {
	var pth Path
	pth.MoveTo(Pt{0, 0})
	pth.LineTo(Pt{1, 0})
	pth.QuadTo(Pt{1, 1}, Pt{2, 1})
	pth.CubicTo(Pt{2, 2}, Pt{3, 2}, Pt{3, 3})
	pth.Close()
	if !pth.Validate() {
		t.Fatalf("expected path to validate")
	}

	// Break validation: remove one vertex
	bad := pth
	bad.V = bad.V[:len(bad.V)-1]
	if bad.Validate() {
		t.Fatalf("expected invalid path")
	}
}

func TestPathCloneAndTransform(t *testing.T) {
	var path Path
	path.MoveTo(Pt{1, 2})
	path.LineTo(Pt{3, 4})

	clone := path.Clone()
	clone.V[0].X = 99
	if path.V[0].X == 99 {
		t.Fatal("Path.Clone should deep-copy vertices")
	}
	clone.C[0] = LineTo
	if path.C[0] == LineTo {
		t.Fatal("Path.Clone should deep-copy commands")
	}

	transformed := path.Transformed(Affine{A: 2, D: 3, E: 5, F: -1})
	if transformed.V[0] != (Pt{7, 5}) || transformed.V[1] != (Pt{11, 11}) {
		t.Fatalf("transformed path vertices = %+v", transformed.V)
	}
	if !transformed.Validate() {
		t.Fatal("transformed path should preserve command/vertex structure")
	}

	bounds, ok := path.Bounds()
	if !ok || bounds != (Rect{Min: Pt{1, 2}, Max: Pt{3, 4}}) {
		t.Fatalf("path bounds = %+v ok=%v", bounds, ok)
	}
	transformedBounds, ok := path.TransformedBounds(Affine{A: 2, D: -1, E: 1, F: 10})
	if !ok || transformedBounds != (Rect{Min: Pt{3, 6}, Max: Pt{7, 8}}) {
		t.Fatalf("transformed bounds = %+v ok=%v", transformedBounds, ok)
	}
	if _, ok := (Path{}).Bounds(); ok {
		t.Fatal("empty path bounds should report ok=false")
	}
}

func TestPathInterpolatedSubdividesCurves(t *testing.T) {
	var path Path
	path.MoveTo(Pt{0, 0})
	path.LineTo(Pt{2, 0})
	path.QuadTo(Pt{3, 2}, Pt{4, 0})
	path.CubicTo(Pt{5, -2}, Pt{6, 2}, Pt{7, 0})
	path.Close()

	interp := path.Interpolated(2)
	if !interp.Validate() {
		t.Fatalf("interpolated path should validate: %+v", interp)
	}
	wantCmds := []Cmd{MoveTo, LineTo, LineTo, LineTo, LineTo, LineTo, LineTo, ClosePath}
	if len(interp.C) != len(wantCmds) {
		t.Fatalf("interpolated command count = %d, want %d (%v)", len(interp.C), len(wantCmds), interp.C)
	}
	for i, want := range wantCmds {
		if interp.C[i] != want {
			t.Fatalf("interpolated command %d = %v, want %v", i, interp.C[i], want)
		}
	}
	wantPts := []Pt{
		{0, 0},
		{1, 0},
		{2, 0},
		{3, 1},
		{4, 0},
		{5.5, 0},
		{7, 0},
	}
	if len(interp.V) != len(wantPts) {
		t.Fatalf("interpolated vertex count = %d, want %d (%v)", len(interp.V), len(wantPts), interp.V)
	}
	for i, want := range wantPts {
		if !approxPt(interp.V[i], want, 1e-12) {
			t.Fatalf("interpolated vertex %d = %+v, want %+v", i, interp.V[i], want)
		}
	}

	clone := path.Interpolated(1)
	if len(clone.C) != len(path.C) || len(clone.V) != len(path.V) {
		t.Fatalf("steps <= 1 should clone original path, got %+v", clone)
	}
	clone.V[0].X = 99
	if path.V[0].X == 99 {
		t.Fatal("Interpolated(1) should deep-copy vertices")
	}
}

func TestPathClippedToRectClipsLineSegments(t *testing.T) {
	var path Path
	path.MoveTo(Pt{-1, 0.5})
	path.LineTo(Pt{0.5, 0.5})
	path.LineTo(Pt{2, 0.5})
	path.MoveTo(Pt{-1, -1})
	path.LineTo(Pt{-2, -2})

	rect := Rect{Min: Pt{0, 0}, Max: Pt{1, 1}}
	clipped := path.ClippedToRect(rect, 2)
	if !clipped.Validate() {
		t.Fatalf("clipped path should validate: %+v", clipped)
	}
	if len(clipped.V) == 0 {
		t.Fatal("clipped path should retain the segment crossing the rect")
	}
	for i, pt := range clipped.V {
		if !rect.ContainsInclusive(pt) {
			t.Fatalf("clipped vertex %d outside rect: %+v", i, pt)
		}
	}
	if !pathContainsPt(clipped, Pt{0, 0.5}, 1e-12) {
		t.Fatalf("clipped path should include left boundary intersection: %+v", clipped.V)
	}
	if !pathContainsPt(clipped, Pt{1, 0.5}, 1e-12) {
		t.Fatalf("clipped path should include right boundary intersection: %+v", clipped.V)
	}
	if pathContainsPt(clipped, Pt{-1, -1}, 1e-12) || pathContainsPt(clipped, Pt{-2, -2}, 1e-12) {
		t.Fatalf("clipped path should drop wholly outside segment: %+v", clipped.V)
	}
}

func TestPathClippedToRectFlattensCurves(t *testing.T) {
	var path Path
	path.MoveTo(Pt{-1, -1})
	path.CubicTo(Pt{0, 2}, Pt{1, 2}, Pt{2, -1})

	rect := Rect{Min: Pt{0, 0}, Max: Pt{1, 1}}
	clipped := path.ClippedToRect(rect, 8)
	if !clipped.Validate() {
		t.Fatalf("clipped curve path should validate: %+v", clipped)
	}
	if len(clipped.V) == 0 {
		t.Fatal("clipped curve should produce visible line segments")
	}
	for i, cmd := range clipped.C {
		if cmd != MoveTo && cmd != LineTo {
			t.Fatalf("clipped flattened command %d = %v, want MoveTo/LineTo only", i, cmd)
		}
	}
	for i, pt := range clipped.V {
		if !rect.ContainsInclusive(pt) {
			t.Fatalf("clipped curve vertex %d outside rect: %+v", i, pt)
		}
	}
}

func TestPathClippedToRectEmptyRect(t *testing.T) {
	var path Path
	path.MoveTo(Pt{0, 0})
	path.LineTo(Pt{1, 1})

	if clipped := path.ClippedToRect(Rect{}, 8); len(clipped.C) != 0 || len(clipped.V) != 0 {
		t.Fatalf("empty rect clip = %+v, want empty path", clipped)
	}
	if clipped := path.ClippedToRect(NullRect(), 8); len(clipped.C) != 0 || len(clipped.V) != 0 {
		t.Fatalf("null rect clip = %+v, want empty path", clipped)
	}
}

func approxPt(a, b Pt, eps float64) bool {
	dx := a.X - b.X
	if dx < 0 {
		dx = -dx
	}
	dy := a.Y - b.Y
	if dy < 0 {
		dy = -dy
	}
	return dx <= eps && dy <= eps
}

func pathContainsPt(path Path, want Pt, eps float64) bool {
	for _, got := range path.V {
		if approxPt(got, want, eps) {
			return true
		}
	}
	return false
}
