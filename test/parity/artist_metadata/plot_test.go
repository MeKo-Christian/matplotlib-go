package artist_metadata

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
)

func TestDataToDisplayMatchesMatplotlibYUpDisplay(t *testing.T) {
	got := dataToDisplay(geom.Pt{X: 2.2, Y: 4.45})
	want := geom.Pt{X: 155.088, Y: 169.74}
	if !approxPt(got, want, 1e-9) {
		t.Fatalf("dataToDisplay(2.2, 4.45) = %+v, want Matplotlib display %+v", got, want)
	}
}

func approxPt(a, b geom.Pt, tol float64) bool {
	return abs(a.X-b.X) <= tol && abs(a.Y-b.Y) <= tol
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
