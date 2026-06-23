package tri

import (
	"math"
	"math/rand"
	"testing"
)

func centroid(t Triangulation, tri [3]int) (float64, float64) {
	return (t.X[tri[0]] + t.X[tri[1]] + t.X[tri[2]]) / 3,
		(t.Y[tri[0]] + t.Y[tri[1]] + t.Y[tri[2]]) / 3
}

func TestTrapezoidFindsCentroids(t *testing.T) {
	x, y := unitSquare()
	tr, err := New(x, y)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	f := NewTrapezoidMapTriFinder(tr)
	if f.fallback != nil {
		t.Fatal("trapezoid map fell back to brute force unexpectedly")
	}
	// The centroid of each triangle is strictly interior, so it must resolve to
	// exactly that triangle.
	for i, tri := range tr.Triangles {
		cx, cy := centroid(tr, tri)
		if got := f.Find(cx, cy); got != i {
			t.Errorf("Find(centroid of triangle %d) = %d, want %d", i, got, i)
		}
	}
}

func TestTrapezoidOutsideHull(t *testing.T) {
	x, y := unitSquare()
	tr, _ := New(x, y)
	f := NewTrapezoidMapTriFinder(tr)
	for _, p := range [][2]float64{{-5, -5}, {5, 5}, {-1, 0.5}, {2, 0.5}} {
		if got := f.Find(p[0], p[1]); got != -1 {
			t.Errorf("Find(%v) = %d, want -1 (outside hull)", p, got)
		}
	}
}

// randomMesh builds a Delaunay triangulation of n pseudo-random points.
func randomMesh(seed int64, n int) (Triangulation, *rand.Rand) {
	rng := rand.New(rand.NewSource(seed))
	x := make([]float64, n)
	y := make([]float64, n)
	for i := range x {
		x[i] = rng.Float64() * 10
		y[i] = rng.Float64() * 10
	}
	tr, err := New(x, y)
	if err != nil {
		panic(err)
	}
	return tr, rng
}

func TestTrapezoidMatchesBruteForce(t *testing.T) {
	for _, seed := range []int64{1, 2, 3, 7, 42} {
		tr, rng := randomMesh(seed, 40)
		tm := NewTrapezoidMapTriFinder(tr)
		if tm.fallback != nil {
			t.Fatalf("seed %d: trapezoid map fell back to brute force", seed)
		}
		bf := NewBruteForceTriFinder(tr)

		// Centroids must resolve exactly.
		for i, tri := range tr.Triangles {
			cx, cy := centroid(tr, tri)
			if got := tm.Find(cx, cy); got != i {
				t.Fatalf("seed %d: trapezoid Find(centroid %d) = %d, want %d", seed, i, got, i)
			}
		}

		// Random points: -1 status must agree, and any triangle reported must
		// actually contain the point (avoids on-edge tie ambiguity).
		for k := 0; k < 500; k++ {
			px := rng.Float64() * 10
			py := rng.Float64() * 10
			gotTM := tm.Find(px, py)
			gotBF := bf.Find(px, py)
			if (gotTM == -1) != (gotBF == -1) {
				// Allow disagreement only if the point sits essentially on an edge.
				if !nearAnyEdge(tr, px, py) {
					t.Fatalf("seed %d: containment disagreement at (%g,%g): trapezoid=%d brute=%d",
						seed, px, py, gotTM, gotBF)
				}
				continue
			}
			if gotTM >= 0 && !pointInTriangle(px, py, tr, tr.Triangles[gotTM]) {
				t.Fatalf("seed %d: trapezoid returned triangle %d not containing (%g,%g)",
					seed, gotTM, px, py)
			}
		}
	}
}

// nearAnyEdge reports whether (px,py) lies within a small distance of any
// triangulation edge, used to tolerate boundary ties in tests.
func nearAnyEdge(t Triangulation, px, py float64) bool {
	const eps = 1e-6
	for _, e := range t.Edges() {
		ax, ay := t.X[e[0]], t.Y[e[0]]
		bx, by := t.X[e[1]], t.Y[e[1]]
		if pointSegDist(px, py, ax, ay, bx, by) < eps {
			return true
		}
	}
	return false
}

func pointSegDist(px, py, ax, ay, bx, by float64) float64 {
	dx, dy := bx-ax, by-ay
	l2 := dx*dx + dy*dy
	if l2 == 0 {
		return math.Hypot(px-ax, py-ay)
	}
	tt := ((px-ax)*dx + (py-ay)*dy) / l2
	tt = math.Max(0, math.Min(1, tt))
	return math.Hypot(px-(ax+tt*dx), py-(ay+tt*dy))
}

func TestTrapezoidMaskedRegion(t *testing.T) {
	tr, _ := randomMesh(11, 30)
	// Mask the first few triangles.
	tr.Mask = make([]bool, len(tr.Triangles))
	for i := 0; i < len(tr.Mask) && i < 3; i++ {
		tr.Mask[i] = true
	}
	tm := NewTrapezoidMapTriFinder(tr)
	for i := 0; i < 3 && i < len(tr.Triangles); i++ {
		cx, cy := centroid(tr, tr.Triangles[i])
		// A masked triangle's centroid must not resolve to that masked triangle.
		if got := tm.Find(cx, cy); got == i {
			t.Errorf("masked triangle %d returned for its own centroid", i)
		}
	}
}
