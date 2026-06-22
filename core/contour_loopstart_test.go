package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
)

// TestRotateClosedLoopToContourpyStart locks in the loop-start convention that
// matches Matplotlib's mpl2014 contour generator: an interior loop starts at the
// leftmost vertical-grid-edge crossing in its bottommost row-band, while a loop
// tangent to the bottom grid boundary starts at its leftmost boundary vertex.
// Aligning the start makes dashed-contour phase match Matplotlib.
func TestRotateClosedLoopToContourpyStart(t *testing.T) {
	x := []float64{0, 1, 2, 3, 4}
	y := []float64{0, 1, 2, 3, 4}

	// Interior rectangle loop with vertical-edge crossings at x=1 and x=3,
	// in row-bands [1,2] (y=1.5) and [2,3] (y=2.5). The bottommost band is
	// [1,2]; its leftmost vertical crossing is (1,1.5).
	interior := []geom.Pt{{X: 1, Y: 1.5}, {X: 3, Y: 1.5}, {X: 3, Y: 2.5}, {X: 1, Y: 2.5}, {X: 1, Y: 1.5}}
	got := rotateClosedLoopToContourpyStart(interior, x, y)
	if !pointsApprox(got[0], geom.Pt{X: 1, Y: 1.5}, 1e-9) {
		t.Fatalf("interior loop start = %+v, want leftmost vertical crossing in bottom band (1,1.5)", got[0])
	}
	if !contourPolylineClosed(got) {
		t.Fatal("rotated interior loop must stay closed")
	}

	// Loop tangent to the bottom boundary (y=0) starts at the leftmost
	// bottom-boundary vertex.
	boundary := []geom.Pt{{X: 3, Y: 1}, {X: 2, Y: 0}, {X: 1, Y: 1}, {X: 2, Y: 2}, {X: 3, Y: 1}}
	got = rotateClosedLoopToContourpyStart(boundary, x, y)
	if !pointsApprox(got[0], geom.Pt{X: 2, Y: 0}, 1e-9) {
		t.Fatalf("boundary-tangent loop start = %+v, want bottom-boundary vertex (2,0)", got[0])
	}

	// Open polylines are left untouched.
	open := []geom.Pt{{X: 0, Y: 1}, {X: 2, Y: 1}, {X: 4, Y: 1}}
	got = rotateClosedLoopToContourpyStart(open, x, y)
	if !pointsApprox(got[0], geom.Pt{X: 0, Y: 1}, 1e-9) {
		t.Fatalf("open polyline must be unchanged, got start %+v", got[0])
	}
}
