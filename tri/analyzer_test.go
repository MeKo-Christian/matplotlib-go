package tri

import (
	"math"
	"testing"
)

// mplMesh is the fixed 10-point mesh with matplotlib's (Qhull) connectivity,
// used to validate neighbours, circle ratios and refinement against
// matplotlib 3.10.9 reference values.
func mplMesh() Triangulation {
	return Triangulation{
		X: []float64{0.0, 1.0, 2.0, 0.3, 1.4, 0.8, 1.9, 2.5, 0.1, 2.2},
		Y: []float64{0.0, 0.2, 0.0, 1.1, 0.9, 1.8, 1.6, 1.0, 2.0, 2.1},
		Triangles: [][3]int{
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
		},
	}
}

func TestNeighborsMatchesMatplotlib(t *testing.T) {
	tr := mplMesh()
	want := [][3]int{
		{-1, 12, 4},
		{3, 10, -1},
		{7, 3, -1},
		{8, 1, 2},
		{0, 5, 6},
		{4, 9, 7},
		{-1, 4, 7},
		{5, 2, 6},
		{10, 3, 9},
		{5, 12, 8},
		{1, 8, 11},
		{12, -1, 10},
		{9, 0, 11},
	}
	got := tr.Neighbors()
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("neighbors[%d] = %v, matplotlib %v", i, got[i], want[i])
		}
	}
}

func TestScaleFactorsMatchesMatplotlib(t *testing.T) {
	a := NewTriAnalyzer(mplMesh())
	kx, ky := a.ScaleFactors()
	if math.Abs(kx-0.4) > 1e-7 || math.Abs(ky-0.47619048) > 1e-7 {
		t.Errorf("ScaleFactors = (%v,%v), want (0.4, 0.47619048)", kx, ky)
	}
}

func TestCircleRatiosMatchesMatplotlib(t *testing.T) {
	a := NewTriAnalyzer(mplMesh())
	wantRescale := []float64{0.07063912, 0.05290857, 0.49130659, 0.46470936, 0.3006142, 0.4803553, 0.26565049, 0.48651002, 0.46193468, 0.4803553, 0.47506481, 0.04049327, 0.44975972}
	wantNo := []float64{0.05141207, 0.03808443, 0.49958542, 0.47054814, 0.25448968, 0.47554645, 0.31761288, 0.45710652, 0.45881124, 0.47554645, 0.49442719, 0.05612392, 0.48035552}
	got := a.CircleRatios(true)
	for i := range wantRescale {
		if math.Abs(got[i]-wantRescale[i]) > 1e-7 {
			t.Errorf("rescale ratio[%d] = %.8f, matplotlib %.8f", i, got[i], wantRescale[i])
		}
	}
	gotNo := a.CircleRatios(false)
	for i := range wantNo {
		if math.Abs(gotNo[i]-wantNo[i]) > 1e-7 {
			t.Errorf("norescale ratio[%d] = %.8f, matplotlib %.8f", i, gotNo[i], wantNo[i])
		}
	}
}

func TestUniformRefineCounts(t *testing.T) {
	r := NewUniformTriRefiner(mplMesh())
	r1, _ := r.RefineTriangulation(1, false)
	if len(r1.X) != 32 || len(r1.Triangles) != 52 {
		t.Errorf("subdiv=1: npts=%d ntri=%d, want 32, 52", len(r1.X), len(r1.Triangles))
	}
	r2, _ := r.RefineTriangulation(2, false)
	if len(r2.X) != 115 || len(r2.Triangles) != 208 {
		t.Errorf("subdiv=2: npts=%d ntri=%d, want 115, 208", len(r2.X), len(r2.Triangles))
	}
}

func TestRefineFieldLinearExact(t *testing.T) {
	tr := mplMesh()
	z := make([]float64, len(tr.X))
	for i := range z {
		z[i] = 2*tr.X[i] - tr.Y[i] + 0.5
	}
	r := NewUniformTriRefiner(tr)
	refi, refiZ, err := r.RefineField(z, nil, 2)
	if err != nil {
		t.Fatalf("RefineField: %v", err)
	}
	// A linear field must be reproduced exactly at every refined node.
	for i := range refi.X {
		want := 2*refi.X[i] - refi.Y[i] + 0.5
		if math.Abs(refiZ[i]-want) > 1e-6 {
			t.Errorf("refined node %d: z=%v, want %v", i, refiZ[i], want)
		}
	}
}

func TestGetFlatTriMaskRemovesSlivers(t *testing.T) {
	// A mesh with one extremely flat border triangle.
	tr := Triangulation{
		X: []float64{0, 1, 2, 1, 3},
		Y: []float64{0, 1, 0, 0.01, 0.005},
		Triangles: [][3]int{
			{0, 1, 3}, {0, 3, 2}, {3, 1, 2}, {2, 4, 3},
		},
	}
	a := NewTriAnalyzer(tr)
	mask := a.GetFlatTriMask(0.1, true)
	if len(mask) != len(tr.Triangles) {
		t.Fatalf("mask length = %d, want %d", len(mask), len(tr.Triangles))
	}
	// The very flat border triangle {2,4,3} must be masked.
	if !mask[3] {
		t.Errorf("expected flat border triangle 3 to be masked")
	}
}
