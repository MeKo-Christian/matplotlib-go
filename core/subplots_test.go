package core

import (
	"math"
	"testing"
)

func TestFigureSubplotsGridLayout(t *testing.T) {
	fig := NewFigure(1000, 800)
	grid := fig.Subplots(
		2,
		2,
		WithSubplotPadding(0.1, 0.9, 0.1, 0.9),
		WithSubplotSpacing(0.1, 0.1),
	)
	if len(grid) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(grid))
	}
	if len(grid[0]) != 2 || len(grid[1]) != 2 {
		t.Fatalf("expected 2 columns, got %d and %d", len(grid[0]), len(grid[1]))
	}

	got := grid[0][0].RectFraction
	if !floatApprox(got.Min.X, 0.1, 1e-12) ||
		!floatApprox(got.Max.X, 0.45, 1e-12) ||
		!floatApprox(got.Min.Y, 0.55, 1e-12) ||
		!floatApprox(got.Max.Y, 0.9, 1e-12) {
		t.Fatalf("unexpected top-left rect: %+v", got)
	}

	got = grid[0][1].RectFraction
	if !floatApprox(got.Min.X, 0.55, 1e-12) ||
		!floatApprox(got.Max.X, 0.9, 1e-12) ||
		!floatApprox(got.Min.Y, 0.55, 1e-12) ||
		!floatApprox(got.Max.Y, 0.9, 1e-12) {
		t.Fatalf("unexpected top-right rect: %+v", got)
	}

	got = grid[1][0].RectFraction
	if !floatApprox(got.Min.X, 0.1, 1e-12) ||
		!floatApprox(got.Max.X, 0.45, 1e-12) ||
		!floatApprox(got.Min.Y, 0.1, 1e-12) ||
		!floatApprox(got.Max.Y, 0.45, 1e-12) {
		t.Fatalf("unexpected bottom-left rect: %+v", got)
	}

	got = grid[1][1].RectFraction
	if !floatApprox(got.Min.X, 0.55, 1e-12) ||
		!floatApprox(got.Max.X, 0.9, 1e-12) ||
		!floatApprox(got.Min.Y, 0.1, 1e-12) ||
		!floatApprox(got.Max.Y, 0.45, 1e-12) {
		t.Fatalf("unexpected bottom-right rect: %+v", got)
	}
}

func TestFigureSubplotsLayoutMapsTopRowToTopPixels(t *testing.T) {
	fig := NewFigure(1000, 800)
	grid := fig.Subplots(
		2,
		2,
		WithSubplotPadding(0.1, 0.9, 0.1, 0.9),
		WithSubplotSpacing(0.1, 0.1),
	)

	// Display space is y-up: the top row maps to the high-Y (top) region.
	topLeft := grid[0][0].layout(fig)
	if !floatApprox(topLeft.Min.X, 100, 1e-12) ||
		!floatApprox(topLeft.Max.X, 450, 1e-12) ||
		!floatApprox(topLeft.Min.Y, 440, 1e-12) ||
		!floatApprox(topLeft.Max.Y, 720, 1e-12) {
		t.Fatalf("unexpected top-left pixel rect: %+v", topLeft)
	}

	bottomLeft := grid[1][0].layout(fig)
	if !floatApprox(bottomLeft.Min.Y, 80, 1e-12) ||
		!floatApprox(bottomLeft.Max.Y, 360, 1e-12) {
		t.Fatalf("unexpected bottom-left pixel rect: %+v", bottomLeft)
	}
}

func TestFigureSubplotsAxisSharing(t *testing.T) {
	fig := NewFigure(1000, 800)
	grid := fig.Subplots(2, 2, WithSubplotShareX(), WithSubplotShareY())
	if len(grid) != 2 || len(grid[0]) != 2 {
		t.Fatalf("expected 2x2 grid, got %dx%d", len(grid), len(grid[0]))
	}

	grid[1][0].SetXLim(-2, 2)
	grid[1][1].SetYLim(-5, 5)

	xMin, xMax := grid[0][0].effectiveXScale().Domain()
	axXMin, axXMax := grid[1][0].effectiveXScale().Domain()
	if xMin != axXMin || xMax != axXMax {
		t.Fatalf("x scale not shared: got base=%v..%v and shared=%v..%v", xMin, xMax, axXMin, axXMax)
	}

	yMin, yMax := grid[0][0].effectiveYScale().Domain()
	axYMin, axYMax := grid[0][1].effectiveYScale().Domain()
	if yMin != axYMin || yMax != axYMax {
		t.Fatalf("y scale not shared: got base=%v..%v and shared=%v..%v", yMin, yMax, axYMin, axYMax)
	}

	if grid[1][1].XAxis != grid[0][0].XAxis {
		t.Fatalf("x-axis object not shared")
	}
	if grid[1][1].YAxis != grid[0][0].YAxis {
		t.Fatalf("y-axis object not shared")
	}
}

func TestFigureSubplotsShareModes(t *testing.T) {
	fig := NewFigure(1000, 800)
	grid := fig.Subplots(
		2,
		2,
		WithSubplotShareXMode(ShareRow),
		WithSubplotShareYMode(ShareCol),
	)

	grid[0][1].SetXLim(-4, 4)
	xMin, xMax := grid[0][0].effectiveXScale().Domain()
	if xMin != -4 || xMax != 4 {
		t.Fatalf("expected row-shared x scale, got %v..%v", xMin, xMax)
	}
	otherXMin, otherXMax := grid[1][0].effectiveXScale().Domain()
	if otherXMin == -4 || otherXMax == 4 {
		t.Fatalf("expected bottom row x scale to remain independent, got %v..%v", otherXMin, otherXMax)
	}

	grid[1][0].SetYLim(-2, 2)
	yMin, yMax := grid[0][0].effectiveYScale().Domain()
	if yMin != -2 || yMax != 2 {
		t.Fatalf("expected column-shared y scale, got %v..%v", yMin, yMax)
	}
	otherYMin, otherYMax := grid[0][1].effectiveYScale().Domain()
	if otherYMin == -2 || otherYMax == 2 {
		t.Fatalf("expected right column y scale to remain independent, got %v..%v", otherYMin, otherYMax)
	}

	if grid[0][1].XAxis != grid[0][0].XAxis {
		t.Fatalf("expected row peers to share x-axis object")
	}
	if grid[1][1].XAxis != grid[1][0].XAxis {
		t.Fatalf("expected bottom row peers to share x-axis object")
	}
	if grid[0][0].XAxis == grid[1][0].XAxis {
		t.Fatalf("did not expect x-axis sharing across rows")
	}

	if grid[1][0].YAxis != grid[0][0].YAxis {
		t.Fatalf("expected column peers to share y-axis object")
	}
	if grid[1][1].YAxis != grid[0][1].YAxis {
		t.Fatalf("expected right column peers to share y-axis object")
	}
	if grid[0][0].YAxis == grid[0][1].YAxis {
		t.Fatalf("did not expect y-axis sharing across columns")
	}
}

func floatApprox(a, b, eps float64) bool {
	if math.IsNaN(a) || math.IsNaN(b) {
		return false
	}
	return math.Abs(a-b) <= eps
}
