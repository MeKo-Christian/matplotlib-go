package qhull

import (
	"math"
	"testing"
)

// Ground truth captured from the vendored Qhull 8.0.2 via the introspection
// tools (third_party/qhull-8.0.2/{dump_state,introspect}.c): projected/scaled
// coordinates, tolerances, and the initial-simplex append order (= first vertex
// ids = start of the point creation order).

func TestProjectSquareMatchesQhull(t *testing.T) {
	// sq_ccw: (0,0)(1,0)(1,1)(0,1).
	q := project([]float64{0, 1, 1, 0}, []float64{0, 0, 1, 1})

	wantPts := [][3]float64{
		{-0.5, -0.5, 0}, {0.5, -0.5, 0}, {0.5, 0.5, 0}, {-0.5, 0.5, 0}, // real
		{0, 0, 0.5}, // Qz infinity point
	}
	if len(q.pts) != len(wantPts) {
		t.Fatalf("got %d projected points, want %d", len(q.pts), len(wantPts))
	}
	for i, w := range wantPts {
		for k := 0; k < 3; k++ {
			if q.pts[i][k] != w[k] {
				t.Errorf("pts[%d][%d] = %v, Qhull %v", i, k, q.pts[i][k], w[k])
			}
		}
	}
	if q.maxAbs != 0.5 || q.maxWidth != 1 || q.maxSum != 1.5 {
		t.Errorf("maxAbs=%v maxWidth=%v maxSum=%v, want 0.5/1/1.5", q.maxAbs, q.maxWidth, q.maxSum)
	}
	if math.Abs(q.distRound-6.9367999643673557e-16) > 1e-30 {
		t.Errorf("distRound=%.17g, Qhull 6.9367999643673557e-16", q.distRound)
	}
}

func TestMaxSimplexMatchesQhull(t *testing.T) {
	cases := []struct {
		name string
		x, y []float64
		want []int // simplex append order (point indices; infinity == n)
	}{
		{"sq_ccw", []float64{0, 1, 1, 0}, []float64{0, 0, 1, 1}, []int{0, 1, 2, 4}},
		{"grid2x2", []float64{0, 1, 0, 1}, []float64{0, 0, 1, 1}, []int{0, 1, 2, 4}},
		{
			"reg6",
			regX(6), regY(6),
			[]int{3, 0, 5, 6},
		},
	}
	for _, tc := range cases {
		q := project(tc.x, tc.y)
		got := q.maxsimplex()
		if len(got) != len(tc.want) {
			t.Errorf("%s: simplex %v, Qhull %v", tc.name, got, tc.want)
			continue
		}
		for i := range tc.want {
			if got[i] != tc.want[i] {
				t.Errorf("%s: simplex %v, Qhull %v", tc.name, got, tc.want)
				break
			}
		}
	}
}

func regX(n int) []float64 {
	out := make([]float64, n)
	for i := range out {
		out[i] = math.Cos(float64(i) * 2 * math.Pi / float64(n))
	}
	return out
}

func regY(n int) []float64 {
	out := make([]float64, n)
	for i := range out {
		out[i] = math.Sin(float64(i) * 2 * math.Pi / float64(n))
	}
	return out
}
