package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
)

func TestGridSpecRatiosAndSpans(t *testing.T) {
	fig := NewFigure(1000, 800)
	gs := fig.GridSpec(
		2,
		3,
		WithGridSpecPadding(0.1, 0.9, 0.2, 0.8),
		WithGridSpecSpacing(0.1, 0.2),
		WithGridSpecWidthRatios(1, 2, 1),
		WithGridSpecHeightRatios(1, 3),
	)
	if gs == nil {
		t.Fatal("expected gridspec")
	}

	middle := gs.Cell(0, 1).Rect()
	assertRectApprox(t, middle, geom.Rect{
		Min: geom.Pt{X: 0.34, Y: 0.68},
		Max: geom.Pt{X: 0.66, Y: 0.8},
	})

	leftSpan := gs.Span(0, 0, 2, 1).Rect()
	assertRectApprox(t, leftSpan, geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.2},
		Max: geom.Pt{X: 0.26, Y: 0.8},
	})
}

func TestNestedGridSpecUsesParentRect(t *testing.T) {
	fig := NewFigure(1000, 800)
	outer := fig.GridSpec(
		1,
		2,
		WithGridSpecPadding(0.1, 0.9, 0.1, 0.9),
		WithGridSpecSpacing(0.1, 0),
	)
	parent := outer.Cell(0, 1)
	inner := parent.GridSpec(
		2,
		1,
		WithGridSpecPadding(0.1, 0.9, 0.2, 0.8),
		WithGridSpecSpacing(0, 0.1),
	)
	if inner == nil {
		t.Fatal("expected nested gridspec")
	}

	top := inner.Cell(0, 0).Rect()
	assertRectApprox(t, top, geom.Rect{
		Min: geom.Pt{X: 0.576, Y: 0.524},
		Max: geom.Pt{X: 0.864, Y: 0.74},
	})
}

func TestFigureAddSubplotAndSubplotCode(t *testing.T) {
	fig := NewFigure(1000, 800)

	ax := fig.AddSubplot(2, 2, 3)
	if ax == nil {
		t.Fatal("expected subplot axes")
	}
	assertRectApprox(t, ax.RectFraction, geom.Rect{
		Min: geom.Pt{X: 0.125, Y: 0.11},
		Max: geom.Pt{X: 0.4875, Y: 0.465},
	})

	code := fig.AddSubplotCode(224)
	if code == nil {
		t.Fatal("expected subplot axes from code")
	}
	assertRectApprox(t, code.RectFraction, geom.Rect{
		Min: geom.Pt{X: 0.5375, Y: 0.11},
		Max: geom.Pt{X: 0.9, Y: 0.465},
	})
}

func TestFigureSubplot2GridSpans(t *testing.T) {
	fig := NewFigure(1000, 800)
	ax := fig.Subplot2Grid([2]int{3, 3}, [2]int{1, 1}, 2, 2)
	if ax == nil {
		t.Fatal("expected subplot2grid axes")
	}
	assertRectApprox(t, ax.RectFraction, geom.Rect{
		Min: geom.Pt{X: 0.4, Y: 0.11},
		Max: geom.Pt{X: 0.9, Y: 0.6033333333333333},
	})
}

func TestFigureSubplotMosaic(t *testing.T) {
	fig := NewFigure(1000, 800)
	axes, err := fig.SubplotMosaic([][]string{
		{"A", "A", "B"},
		{"C", ".", "B"},
	}, WithGridSpecPadding(0, 1, 0, 1), WithGridSpecSpacing(0, 0))
	if err != nil {
		t.Fatalf("subplot mosaic failed: %v", err)
	}

	if len(axes) != 3 {
		t.Fatalf("expected 3 mosaic axes, got %d", len(axes))
	}
	assertRectApprox(t, axes["A"].RectFraction, geom.Rect{
		Min: geom.Pt{X: 0, Y: 0.5},
		Max: geom.Pt{X: 2.0 / 3.0, Y: 1},
	})
	assertRectApprox(t, axes["B"].RectFraction, geom.Rect{
		Min: geom.Pt{X: 2.0 / 3.0, Y: 0},
		Max: geom.Pt{X: 1, Y: 1},
	})
	assertRectApprox(t, axes["C"].RectFraction, geom.Rect{
		Min: geom.Pt{X: 0, Y: 0},
		Max: geom.Pt{X: 1.0 / 3.0, Y: 0.5},
	})
}

func TestFigureSubplotMosaicRejectsNonRectangularRegion(t *testing.T) {
	fig := NewFigure(1000, 800)
	_, err := fig.SubplotMosaic([][]string{
		{"A", "A"},
		{"A", "B"},
	}, WithGridSpecPadding(0, 1, 0, 1), WithGridSpecSpacing(0, 0))
	if err == nil {
		t.Fatal("expected non-rectangular mosaic error")
	}
}

func TestSubFigureComposition(t *testing.T) {
	fig := NewFigure(1000, 800)
	sub := fig.AddSubFigure(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.2},
		Max: geom.Pt{X: 0.9, Y: 0.8},
	})
	nested := sub.AddSubFigure(geom.Rect{
		Min: geom.Pt{X: 0.25, Y: 0.5},
		Max: geom.Pt{X: 0.75, Y: 1},
	})
	ax := nested.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.2},
		Max: geom.Pt{X: 0.9, Y: 0.8},
	})
	if ax == nil {
		t.Fatal("expected axes inside nested subfigure")
	}

	assertRectApprox(t, nested.RectFraction, geom.Rect{
		Min: geom.Pt{X: 0.3, Y: 0.5},
		Max: geom.Pt{X: 0.7, Y: 0.8},
	})
	assertRectApprox(t, ax.RectFraction, geom.Rect{
		Min: geom.Pt{X: 0.34, Y: 0.56},
		Max: geom.Pt{X: 0.66, Y: 0.74},
	})
}

func TestSubplotSpecSubFigureUsesMatplotlibRawGridCell(t *testing.T) {
	fig := NewFigure(960, 640)
	outer := fig.GridSpec(
		2,
		2,
		WithGridSpecPadding(0.08, 0.96, 0.10, 0.92),
		WithGridSpecSpacing(0.06/(2+0.06), 0.28/(2+0.28)),
		WithGridSpecWidthRatios(2, 1),
	)

	sub := outer.Cell(1, 1).SubFigure()
	if sub == nil {
		t.Fatal("expected subfigure")
	}
	assertRectApprox(t, sub.RectFraction, geom.Rect{
		Min: geom.Pt{X: 2.0 / 3.0, Y: 0},
		Max: geom.Pt{X: 1, Y: 0.5},
	})

	ax := sub.AddSubplot(1, 1, 1)
	if ax == nil {
		t.Fatal("expected subfigure subplot")
	}
	assertRectApprox(t, ax.RectFraction, geom.Rect{
		Min: geom.Pt{X: 0.7083333333333334, Y: 0.055},
		Max: geom.Pt{X: 0.9666666666666667, Y: 0.44},
	})
}

func TestSubplotSpecExplicitPeerSharing(t *testing.T) {
	fig := NewFigure(1000, 800)
	gs := fig.GridSpec(1, 2, WithGridSpecPadding(0, 1, 0, 1), WithGridSpecSpacing(0, 0))
	left := gs.Cell(0, 0).AddAxes()
	right := gs.Cell(0, 1).AddAxes(WithSharedAxes(left))

	right.SetXLim(-2, 2)
	right.SetYLim(-3, 3)

	xMin, xMax := left.effectiveXScale().Domain()
	yMin, yMax := left.effectiveYScale().Domain()
	if xMin != -2 || xMax != 2 {
		t.Fatalf("shared x scale mismatch: got %v..%v", xMin, xMax)
	}
	if yMin != -3 || yMax != 3 {
		t.Fatalf("shared y scale mismatch: got %v..%v", yMin, yMax)
	}
	if right.XAxis != left.XAxis {
		t.Fatal("expected shared x-axis object")
	}
	if right.YAxis != left.YAxis {
		t.Fatal("expected shared y-axis object")
	}
}

func TestFigureSubplotsAdjustUpdatesManagedGrid(t *testing.T) {
	fig := NewFigure(1000, 800)
	grid := fig.Subplots(2, 2)

	left := 0.2
	right := 0.85
	bottom := 0.15
	top := 0.88
	fig.SubplotsAdjust(SubplotAdjust{
		Left:   &left,
		Right:  &right,
		Bottom: &bottom,
		Top:    &top,
	})

	if !floatApprox(grid[0][0].RectFraction.Min.X, left, 1e-12) {
		t.Fatalf("left margin = %v, want %v", grid[0][0].RectFraction.Min.X, left)
	}
	if !floatApprox(grid[0][0].RectFraction.Max.Y, top, 1e-12) {
		t.Fatalf("top margin = %v, want %v", grid[0][0].RectFraction.Max.Y, top)
	}
	if !floatApprox(grid[1][1].RectFraction.Max.X, right, 1e-12) {
		t.Fatalf("right margin = %v, want %v", grid[1][1].RectFraction.Max.X, right)
	}
	if !floatApprox(grid[1][1].RectFraction.Min.Y, bottom, 1e-12) {
		t.Fatalf("bottom margin = %v, want %v", grid[1][1].RectFraction.Min.Y, bottom)
	}
}

func assertRectApprox(t *testing.T, got, want geom.Rect) {
	t.Helper()
	if !floatApprox(got.Min.X, want.Min.X, 1e-12) ||
		!floatApprox(got.Min.Y, want.Min.Y, 1e-12) ||
		!floatApprox(got.Max.X, want.Max.X, 1e-12) ||
		!floatApprox(got.Max.Y, want.Max.Y, 1e-12) {
		t.Fatalf("unexpected rect: got %+v want %+v", got, want)
	}
}

func TestLabelOuterSuppressesInnerLabels(t *testing.T) {
	fig := NewFigure(800, 600)
	gs := fig.GridSpec(2, 2)
	// Top-left (0,0): first row, first col.
	tl := gs.Cell(0, 0).AddAxes()
	// Bottom-right (1,1): last row, last col.
	br := gs.Cell(1, 1).AddAxes()
	for _, ax := range []*Axes{tl, br} {
		ax.SetXLabel("x")
		ax.SetYLabel("y")
		ax.LabelOuter()
	}
	// Top-left: inner bottom (not last row) -> hide x tick labels + xlabel.
	if !tl.hideXTickLabels || !tl.hideXLabel {
		t.Fatalf("top-left should hide x labels (not last row): %+v", tl.hideXTickLabels)
	}
	// Top-left is first col -> keep y labels.
	if tl.hideYTickLabels || tl.hideYLabel {
		t.Fatalf("top-left (first col) should keep y labels")
	}
	// Bottom-right is last row -> keep x labels.
	if br.hideXTickLabels || br.hideXLabel {
		t.Fatalf("bottom-right (last row) should keep x labels")
	}
	// Bottom-right not first col -> hide y labels.
	if !br.hideYTickLabels || !br.hideYLabel {
		t.Fatalf("bottom-right should hide y labels (not first col)")
	}
}
