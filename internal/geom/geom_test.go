package geom

import (
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
