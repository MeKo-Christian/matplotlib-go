package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
)

func countPathCmd(path geom.Path, want geom.Cmd) int {
	count := 0
	for _, cmd := range path.C {
		if cmd == want {
			count++
		}
	}
	return count
}

func assertApproxPathBounds(t *testing.T, path geom.Path, want geom.Rect) {
	t.Helper()
	got, ok := pathBounds(path)
	if !ok {
		t.Fatal("path has no bounds")
	}
	if !approxPt(got.Min, want.Min, 1e-9) || !approxPt(got.Max, want.Max, 1e-9) {
		t.Fatalf("bounds = %+v, want %+v", got, want)
	}
}

func assertPathVerticesApprox(t *testing.T, got geom.Path, want []geom.Pt, tol float64) {
	t.Helper()
	if len(got.V) != len(want) {
		t.Fatalf("path vertices = %d, want %d: %+v", len(got.V), len(want), got.V)
	}
	for i := range want {
		if !approxPt(got.V[i], want[i], tol) {
			t.Fatalf("vertex %d = %+v, want %+v (all vertices %+v)", i, got.V[i], want[i], got.V)
		}
	}
}

func styleMatrixTestContext() *DrawContext {
	fig := NewFigure(720, 420)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.03, Y: 0.06}, Max: geom.Pt{X: 0.97, Y: 0.96}})
	ax.SetXLim(0, 12)
	ax.SetYLim(0, 8)
	return AxesDrawContext(ax, fig)
}

func approxPt(a, b geom.Pt, tol float64) bool {
	return math.Abs(a.X-b.X) <= tol && math.Abs(a.Y-b.Y) <= tol
}

func containsPointForPatchTest(path geom.Path, want geom.Pt) bool {
	for _, pt := range path.V {
		if approx(pt.X, want.X, 1e-9) && approx(pt.Y, want.Y, 1e-9) {
			return true
		}
	}
	return false
}

func maxPathYForPatchTest(path geom.Path) float64 {
	maxY := math.Inf(-1)
	for _, pt := range path.V {
		maxY = math.Max(maxY, pt.Y)
	}
	return maxY
}
