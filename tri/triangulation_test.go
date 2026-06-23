package tri

import (
	"math"
	"testing"
)

// unitSquare returns the four corners of the unit square.
func unitSquare() ([]float64, []float64) {
	return []float64{0, 1, 1, 0}, []float64{0, 0, 1, 1}
}

func signedArea(t Triangulation, tri [3]int) float64 {
	ax, ay := t.X[tri[0]], t.Y[tri[0]]
	bx, by := t.X[tri[1]], t.Y[tri[1]]
	cx, cy := t.X[tri[2]], t.Y[tri[2]]
	return 0.5 * ((bx-ax)*(cy-ay) - (by-ay)*(cx-ax))
}

func TestNewDelaunaySquare(t *testing.T) {
	x, y := unitSquare()
	tr, err := New(x, y)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if len(tr.Triangles) != 2 {
		t.Fatalf("triangle count = %d, want 2", len(tr.Triangles))
	}
	// Every triangle must be wound anticlockwise (positive signed area).
	for i, tri := range tr.Triangles {
		if a := signedArea(tr, tri); a <= 0 {
			t.Errorf("triangle %d %v signed area = %v, want > 0", i, tri, a)
		}
	}
}

func TestNewErrors(t *testing.T) {
	if _, err := New([]float64{0, 1}, []float64{0, 1}); err == nil {
		t.Error("expected error for fewer than 3 points")
	}
	// Collinear points have zero area -> no triangulation.
	if _, err := New([]float64{0, 1, 2}, []float64{0, 1, 2}); err == nil {
		t.Error("expected error for collinear points")
	}
}

func TestEdges(t *testing.T) {
	x, y := unitSquare()
	tr, err := New(x, y)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	edges := tr.Edges()
	// (3*ntri + boundary)/2 = (6 + 4)/2 = 5 unique edges.
	if len(edges) != 5 {
		t.Fatalf("edge count = %d, want 5", len(edges))
	}
	for _, e := range edges {
		if e[0] >= e[1] {
			t.Errorf("edge %v not ordered (min,max)", e)
		}
	}
}

func TestNeighbors(t *testing.T) {
	x, y := unitSquare()
	tr, err := New(x, y)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	nb := tr.Neighbors()
	if len(nb) != len(tr.Triangles) {
		t.Fatalf("neighbors length = %d, want %d", len(nb), len(tr.Triangles))
	}
	// Two triangles sharing the diagonal: each has exactly one non-boundary
	// neighbour, and that relationship is symmetric.
	boundary, interior := 0, 0
	for tIdx, row := range nb {
		for apex, other := range row {
			if other == -1 {
				boundary++
				continue
			}
			interior++
			// Symmetry: the neighbour must reference this triangle back.
			found := false
			for _, back := range nb[other] {
				if back == tIdx {
					found = true
				}
			}
			if !found {
				t.Errorf("neighbour of triangle %d apex %d = %d is not symmetric", tIdx, apex, other)
			}
		}
	}
	if interior != 2 {
		t.Errorf("interior neighbour links = %d, want 2", interior)
	}
	if boundary != 4 {
		t.Errorf("boundary edges = %d, want 4", boundary)
	}
}

func TestPlaneCoefficientsLinearField(t *testing.T) {
	x, y := unitSquare()
	tr, err := New(x, y)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	// z = 2x + 3y + 1 on every node; every triangle's plane must recover it.
	z := make([]float64, len(x))
	for i := range z {
		z[i] = 2*x[i] + 3*y[i] + 1
	}
	coeffs, err := tr.PlaneCoefficients(z)
	if err != nil {
		t.Fatalf("PlaneCoefficients: %v", err)
	}
	for i, c := range coeffs {
		if math.Abs(c[0]-2) > 1e-9 || math.Abs(c[1]-3) > 1e-9 || math.Abs(c[2]-1) > 1e-9 {
			t.Errorf("triangle %d coeffs = %v, want (2,3,1)", i, c)
		}
	}
}

func TestPlaneCoefficientsLengthMismatch(t *testing.T) {
	x, y := unitSquare()
	tr, _ := New(x, y)
	if _, err := tr.PlaneCoefficients([]float64{1, 2}); err == nil {
		t.Error("expected error for mismatched z length")
	}
}

func TestEnsureTrianglesPassThrough(t *testing.T) {
	// A pre-set triangle list is returned unchanged.
	in := Triangulation{
		X:         []float64{0, 1, 0},
		Y:         []float64{0, 0, 1},
		Triangles: [][3]int{{0, 1, 2}},
	}
	out, ok := in.EnsureTriangles()
	if !ok || len(out.Triangles) != 1 {
		t.Fatalf("EnsureTriangles pass-through failed: ok=%v n=%d", ok, len(out.Triangles))
	}
}
