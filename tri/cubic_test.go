package tri

import (
	"math"
	"testing"
)

// A cubic HCT interpolator must reproduce a linear field exactly (linear patch
// test) for every gradient-estimation strategy.
func TestCubicLinearPatch(t *testing.T) {
	tr, rng := randomMesh(5, 30)
	z := make([]float64, len(tr.X))
	dzdx := make([]float64, len(tr.X))
	dzdy := make([]float64, len(tr.X))
	for i := range z {
		z[i] = 2*tr.X[i] + 3*tr.Y[i] + 1
		dzdx[i] = 2
		dzdy[i] = 3
	}

	kinds := []struct {
		name string
		kind CubicKind
	}{
		{"min_E", CubicMinE},
		{"geom", CubicGeom},
		{"user", CubicUser},
	}
	for _, kc := range kinds {
		var ci *CubicTriInterpolator
		var err error
		if kc.kind == CubicUser {
			ci, err = NewCubicTriInterpolator(tr, z, kc.kind, dzdx, dzdy)
		} else {
			ci, err = NewCubicTriInterpolator(tr, z, kc.kind, nil, nil)
		}
		if err != nil {
			t.Fatalf("%s: NewCubicTriInterpolator: %v", kc.name, err)
		}
		for i, tri := range tr.Triangles {
			cx, cy := centroid(tr, tri)
			got, ok := ci.Interpolate(cx, cy)
			if !ok {
				t.Fatalf("%s: triangle %d centroid outside mesh", kc.name, i)
			}
			want := 2*cx + 3*cy + 1
			if math.Abs(got-want) > 1e-7 {
				t.Errorf("%s: triangle %d Interpolate = %v, want %v", kc.name, i, got, want)
			}
			dx, dy, ok := ci.Gradient(cx, cy)
			if !ok || math.Abs(dx-2) > 1e-6 || math.Abs(dy-3) > 1e-6 {
				t.Errorf("%s: triangle %d Gradient = (%v,%v), want (2,3)", kc.name, i, dx, dy)
			}
		}
		_ = rng
	}
}

// At the nodes the interpolated value must equal the imposed nodal value.
func TestCubicReproducesNodalValues(t *testing.T) {
	tr, _ := randomMesh(9, 25)
	z := make([]float64, len(tr.X))
	for i := range z {
		z[i] = math.Sin(tr.X[i]) * math.Cos(tr.Y[i])
	}
	ci, err := NewCubicTriInterpolator(tr, z, CubicMinE, nil, nil)
	if err != nil {
		t.Fatalf("NewCubicTriInterpolator: %v", err)
	}
	// Evaluate slightly inside each triangle toward each vertex; the value near a
	// vertex should approach the nodal value.
	for _, tri := range tr.Triangles {
		for _, v := range tri {
			vx, vy := tr.X[v], tr.Y[v]
			cx, cy := centroid(tr, tri)
			// Point very close to the vertex but inside the triangle.
			px := vx + 1e-4*(cx-vx)
			py := vy + 1e-4*(cy-vy)
			got, ok := ci.Interpolate(px, py)
			if !ok {
				continue
			}
			if math.Abs(got-z[v]) > 1e-2 {
				t.Errorf("near node %d: Interpolate = %v, nodal z = %v", v, got, z[v])
			}
		}
	}
}

func TestCubicOutsideHull(t *testing.T) {
	x, y := unitSquare()
	tr, _ := New(x, y)
	ci, err := NewCubicTriInterpolator(tr, []float64{0, 1, 2, 3}, CubicGeom, nil, nil)
	if err != nil {
		t.Fatalf("NewCubicTriInterpolator: %v", err)
	}
	if _, ok := ci.Interpolate(-5, -5); ok {
		t.Error("expected ok=false outside hull")
	}
}

// TestCubicMatchesMatplotlib compares against values produced by matplotlib
// 3.10.9's CubicTriInterpolator on an identical mesh (connectivity captured
// from matplotlib's Qhull output so both interpolate over the same triangles).
func TestCubicMatchesMatplotlib(t *testing.T) {
	xs := []float64{0.0, 1.0, 2.0, 0.3, 1.4, 0.8, 1.9, 2.5, 0.1, 2.2}
	ys := []float64{0.0, 0.2, 0.0, 1.1, 0.9, 1.8, 1.6, 1.0, 2.0, 2.1}
	triangles := [][3]int{
		{9, 8, 5},
		{2, 1, 0},
		{7, 4, 2},
		{4, 1, 2},
		{9, 5, 6},
		{6, 5, 4},
		{7, 9, 6},
		{6, 4, 7},
		{3, 1, 4},
		{4, 5, 3},
		{0, 1, 3},
		{3, 8, 0},
		{3, 5, 8},
	}
	z := make([]float64, len(xs))
	for i := range z {
		z[i] = math.Sin(xs[i]) * math.Cos(ys[i])
	}
	tr := Triangulation{X: xs, Y: ys, Triangles: triangles}

	qx := []float64{0.6, 1.2, 1.5, 0.9, 1.7, 1.0}
	qy := []float64{0.5, 0.6, 1.0, 1.2, 1.3, 0.8}

	cases := []struct {
		kind CubicKind
		want []float64
	}{
		{CubicMinE, []float64{0.5169357306, 0.7643173917, 0.536585714, 0.2781565439, 0.2567751914, 0.5785778633}},
		{CubicGeom, []float64{0.5334162028, 0.7601190975, 0.54490775, 0.2889555599, 0.2633142657, 0.5882442846}},
	}
	for _, c := range cases {
		ci, err := NewCubicTriInterpolator(tr, z, c.kind, nil, nil)
		if err != nil {
			t.Fatalf("kind %d: %v", c.kind, err)
		}
		for i := range qx {
			got, ok := ci.Interpolate(qx[i], qy[i])
			if !ok {
				t.Fatalf("kind %d: point %d outside mesh", c.kind, i)
			}
			if math.Abs(got-c.want[i]) > 1e-6 {
				t.Errorf("kind %d point %d: got %.10f, matplotlib %.10f (diff %.2e)",
					c.kind, i, got, c.want[i], math.Abs(got-c.want[i]))
			}
		}
	}
}
