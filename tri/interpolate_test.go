package tri

import (
	"math"
	"testing"
)

func TestLinearInterpolatorExactOnLinearField(t *testing.T) {
	tr, rng := randomMesh(5, 30)
	z := make([]float64, len(tr.X))
	for i := range z {
		z[i] = 2*tr.X[i] + 3*tr.Y[i] + 1
	}
	li, err := NewLinearTriInterpolator(tr, z)
	if err != nil {
		t.Fatalf("NewLinearTriInterpolator: %v", err)
	}
	// Interpolation of a linear field is exact at any interior point.
	for i, tri := range tr.Triangles {
		cx, cy := centroid(tr, tri)
		got, ok := li.Interpolate(cx, cy)
		if !ok {
			t.Fatalf("triangle %d centroid reported outside mesh", i)
		}
		want := 2*cx + 3*cy + 1
		if math.Abs(got-want) > 1e-9 {
			t.Errorf("triangle %d: Interpolate = %v, want %v", i, got, want)
		}
		dzdx, dzdy, ok := li.Gradient(cx, cy)
		if !ok || math.Abs(dzdx-2) > 1e-9 || math.Abs(dzdy-3) > 1e-9 {
			t.Errorf("triangle %d: Gradient = (%v,%v), want (2,3)", i, dzdx, dzdy)
		}
	}

	// A handful of random interior points.
	for k := 0; k < 200; k++ {
		px := rng.Float64() * 10
		py := rng.Float64() * 10
		got, ok := li.Interpolate(px, py)
		if !ok {
			continue
		}
		want := 2*px + 3*py + 1
		if math.Abs(got-want) > 1e-9 {
			t.Errorf("Interpolate(%g,%g) = %v, want %v", px, py, got, want)
		}
	}
}

func TestLinearInterpolatorOutsideHull(t *testing.T) {
	x, y := unitSquare()
	tr, _ := New(x, y)
	li, err := NewLinearTriInterpolator(tr, []float64{0, 1, 2, 3})
	if err != nil {
		t.Fatalf("NewLinearTriInterpolator: %v", err)
	}
	if _, ok := li.Interpolate(-5, -5); ok {
		t.Error("expected ok=false outside hull")
	}
}

func TestLinearInterpolatorBadLength(t *testing.T) {
	x, y := unitSquare()
	tr, _ := New(x, y)
	if _, err := NewLinearTriInterpolator(tr, []float64{1, 2}); err == nil {
		t.Error("expected error for mismatched z length")
	}
}
