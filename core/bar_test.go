package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestBar2D_Draw_Vertical(t *testing.T) {
	// Create a vertical bar chart
	bar := &Bar2D{
		X:           []float64{1, 2, 3},
		Heights:     []float64{5, 8, 3},
		Width:       0.8,
		Color:       render.Color{R: 0.5, G: 0.5, B: 1, A: 1},
		Baseline:    0,
		Orientation: BarVertical,
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic
	bar.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestBar2D_Draw_Horizontal(t *testing.T) {
	// Create a horizontal bar chart
	bar := &Bar2D{
		X:           []float64{1, 2, 3},
		Heights:     []float64{5, 8, 3},
		Width:       0.8,
		Color:       render.Color{R: 1, G: 0.5, B: 0.5, A: 1},
		Baseline:    0,
		Orientation: BarHorizontal,
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic
	bar.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestBar2D_EmptyData(t *testing.T) {
	// Test with empty data
	bar := &Bar2D{
		X:           []float64{},
		Heights:     []float64{},
		Width:       0.8,
		Color:       render.Color{R: 1, G: 0, B: 0, A: 1},
		Orientation: BarVertical,
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic with empty data
	bar.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestBar2D_MismatchedLengths(t *testing.T) {
	// Test with mismatched X and Heights lengths
	bar := &Bar2D{
		X:           []float64{1, 2, 3, 4, 5}, // 5 elements
		Heights:     []float64{5, 8},          // 2 elements
		Width:       0.8,
		Color:       render.Color{R: 1, G: 0, B: 0, A: 1},
		Orientation: BarVertical,
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should only draw min(len(X), len(Heights)) bars
	bar.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestBarAutoScaleUsesStickyBaseline(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	ax.Bar([]float64{0, 1, 2, 3}, []float64{3, 8, 6, 4})
	ax.AutoScale(0.05)

	_, xMax := ax.XScale.Domain()
	yMin, yMax := ax.YScale.Domain()
	if math.Abs(yMin) > 1e-12 {
		t.Fatalf("bar autoscale y-min = %v, want sticky baseline 0", yMin)
	}
	if math.Abs(yMax-8.4) > 1e-12 {
		t.Fatalf("bar autoscale y-max = %v, want 8.4", yMax)
	}
	if math.Abs(xMax-3.59) > 1e-12 {
		t.Fatalf("bar autoscale x-max = %v, want 3.59", xMax)
	}
}

func TestAxesBarAutoScalesOnAddLikeMatplotlib(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	width := 0.72

	ax.Bar([]float64{0, 1, 2}, []float64{0.35, 0.75, 0.55}, BarOptions{Width: &width})

	xMin, xMax := ax.XScale.Domain()
	yMin, yMax := ax.YScale.Domain()
	if math.Abs(xMin-(-0.496)) > 1e-12 || math.Abs(xMax-2.496) > 1e-12 {
		t.Fatalf("bar autoscale x = [%v, %v], want [-0.496, 2.496]", xMin, xMax)
	}
	if math.Abs(yMin) > 1e-12 || math.Abs(yMax-0.7875) > 1e-12 {
		t.Fatalf("bar autoscale y = [%v, %v], want [0, 0.7875]", yMin, yMax)
	}
}

func TestHorizontalBarAutoScaleUsesStickyBaseline(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	orientation := BarHorizontal

	ax.Bar([]float64{0, 1, 2}, []float64{4, 7, 5}, BarOptions{Orientation: &orientation})
	ax.AutoScale(0.05)

	xMin, xMax := ax.XScale.Domain()
	yMin, yMax := ax.YScale.Domain()
	if math.Abs(xMin) > 1e-12 {
		t.Fatalf("horizontal bar autoscale x-min = %v, want sticky baseline 0", xMin)
	}
	if math.Abs(xMax-7.35) > 1e-12 {
		t.Fatalf("horizontal bar autoscale x-max = %v, want 7.35", xMax)
	}
	if math.Abs(yMin-(-0.54)) > 1e-12 || math.Abs(yMax-2.54) > 1e-12 {
		t.Fatalf("horizontal bar autoscale y = [%v, %v], want [-0.54, 2.54]", yMin, yMax)
	}
}

func TestAxesBarHForcesHorizontalOrientation(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	vertical := BarVertical

	bar := ax.BarH([]float64{0, 1}, []float64{4, 7}, BarOptions{
		Orientation: &vertical,
		Label:       "horizontal",
	})
	if bar == nil {
		t.Fatal("BarH() returned nil")
	}
	if bar.Orientation != BarHorizontal {
		t.Fatalf("BarH orientation = %v, want BarHorizontal", bar.Orientation)
	}
	if bar.Label != "horizontal" {
		t.Fatalf("BarH label = %q, want horizontal", bar.Label)
	}
}

func TestAxesBarEdgeAlignConvertsPositionsToCenters(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})
	align := BarAlignEdge
	width := 0.5

	bar := ax.Bar([]float64{1, 2}, []float64{3, 4}, BarOptions{
		Align: &align,
		Width: &width,
	})

	if bar == nil {
		t.Fatal("Bar returned nil")
	}
	if got, want := bar.X, []float64{1.25, 2.25}; len(got) != len(want) {
		t.Fatalf("bar centers = %v, want %v", got, want)
	}
	for i, want := range []float64{1.25, 2.25} {
		if got := bar.X[i]; got != want {
			t.Fatalf("bar center %d = %v, want %v", i, got, want)
		}
	}
	if got, want := bar.Bounds(nil), (geom.Rect{Min: geom.Pt{X: 1, Y: 0}, Max: geom.Pt{X: 2.5, Y: 4}}); got != want {
		t.Fatalf("bar bounds = %+v, want %+v", got, want)
	}
}

func TestAxesBarAppliesPerBarWidthsColorsAndBaselines(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})
	widths := []float64{0.4, 0.8}
	colors := []render.Color{{R: 1, A: 1}, {G: 1, A: 1}}
	edgeColors := []render.Color{{B: 1, A: 1}, {R: 0.5, G: 0.5, A: 1}}
	baselines := []float64{1, 3}

	bar := ax.Bar([]float64{1, 2}, []float64{2, -1}, BarOptions{
		Widths:     widths,
		Colors:     colors,
		EdgeColors: edgeColors,
		Baselines:  baselines,
	})

	if bar == nil {
		t.Fatal("Bar returned nil")
	}
	for i, want := range widths {
		if got := bar.Widths[i]; got != want {
			t.Fatalf("width %d = %v, want %v", i, got, want)
		}
	}
	for i, want := range colors {
		if got := bar.Colors[i]; got != want {
			t.Fatalf("color %d = %+v, want %+v", i, got, want)
		}
	}
	for i, want := range edgeColors {
		if got := bar.EdgeColors[i]; got != want {
			t.Fatalf("edge color %d = %+v, want %+v", i, got, want)
		}
	}
	if got, want := bar.Bounds(nil), (geom.Rect{Min: geom.Pt{X: 0.8, Y: 1}, Max: geom.Pt{X: 2.4, Y: 3}}); got != want {
		t.Fatalf("bar bounds = %+v, want %+v", got, want)
	}
}

func TestBar2D_VariableWidthsAndColors(t *testing.T) {
	// Test with variable widths and colors
	bar := &Bar2D{
		X:       []float64{1, 2, 3},
		Heights: []float64{5, 8, 3},
		Widths:  []float64{0.5, 1.0, 0.3},
		Colors: []render.Color{
			{R: 1, G: 0, B: 0, A: 1},
			{R: 0, G: 1, B: 0, A: 1},
			{R: 0, G: 0, B: 1, A: 1},
		},
		Width:       0.8,                                        // fallback
		Color:       render.Color{R: 0.5, G: 0.5, B: 0.5, A: 1}, // fallback
		Orientation: BarVertical,
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic
	bar.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestBar2D_NegativeValues(t *testing.T) {
	// Test with negative values (bars below baseline)
	bar := &Bar2D{
		X:           []float64{1, 2, 3},
		Heights:     []float64{-2, 5, -1},
		Width:       0.8,
		Color:       render.Color{R: 1, G: 0.5, B: 0, A: 1},
		Baseline:    0,
		Orientation: BarVertical,
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should handle negative values correctly
	bar.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestAxesBarPreservesEdgeWidthPixelsForMatplotlibRefHelpers(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})
	edgeWidth := 1.0

	bar := ax.Bar([]float64{1}, []float64{3}, BarOptions{EdgeWidth: &edgeWidth})
	if bar == nil {
		t.Fatal("Bar returned nil")
	}
	if got := bar.EdgeWidth; got != edgeWidth {
		t.Fatalf("bar edge width = %v, want Go renderer pixel width %v", got, edgeWidth)
	}
}

func TestBar2D_EdgeColors(t *testing.T) {
	// Test with edge colors and width
	bar := &Bar2D{
		X:       []float64{1, 2, 3},
		Heights: []float64{5, 8, 3},
		Width:   0.8,
		Color:   render.Color{R: 0.8, G: 0.8, B: 0.8, A: 1},
		EdgeColors: []render.Color{
			{R: 1, G: 0, B: 0, A: 1},
			{R: 0, G: 1, B: 0, A: 1},
			{R: 0, G: 0, B: 1, A: 1},
		},
		EdgeColor:   render.Color{R: 0, G: 0, B: 0, A: 1}, // fallback
		EdgeWidth:   2.0,
		Orientation: BarVertical,
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic with edge colors
	bar.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestBar2D_AlphaTransparency(t *testing.T) {
	// Test with alpha transparency
	bar := &Bar2D{
		X:           []float64{1, 2, 3},
		Heights:     []float64{5, 8, 3},
		Width:       0.8,
		Color:       render.Color{R: 1, G: 0, B: 0, A: 1},
		Alpha:       0.5, // 50% transparent
		Orientation: BarVertical,
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic with alpha transparency
	bar.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestBar2D_ZOrder(t *testing.T) {
	bar := &Bar2D{z: 2.5}

	if got := bar.Z(); got != 2.5 {
		t.Errorf("Expected Z() = 2.5, got %v", got)
	}
}

func TestBar2D_Bounds_Empty(t *testing.T) {
	bar := &Bar2D{}

	bounds := bar.Bounds(nil)
	expected := geom.Rect{}

	if bounds != expected {
		t.Errorf("Expected Bounds() = %v, got %v", expected, bounds)
	}
}

func TestBar2D_Bounds_Vertical(t *testing.T) {
	// Test vertical bar chart bounds
	bar := &Bar2D{
		X:           []float64{1, 2, 3},
		Heights:     []float64{5, 8, 3},
		Width:       1.0,
		Orientation: BarVertical,
		Baseline:    0,
	}
	bounds := bar.Bounds(nil)

	// Should include all bar positions and heights
	expectedMinX := 1.0 - 0.5 // first bar left edge
	expectedMaxX := 3.0 + 0.5 // last bar right edge
	expectedMinY := 0.0       // baseline
	expectedMaxY := 8.0       // tallest bar

	if bounds.Min.X != expectedMinX {
		t.Errorf("Expected MinX = %v, got %v", expectedMinX, bounds.Min.X)
	}
	if bounds.Max.X != expectedMaxX {
		t.Errorf("Expected MaxX = %v, got %v", expectedMaxX, bounds.Max.X)
	}
	if bounds.Min.Y != expectedMinY {
		t.Errorf("Expected MinY = %v, got %v", expectedMinY, bounds.Min.Y)
	}
	if bounds.Max.Y != expectedMaxY {
		t.Errorf("Expected MaxY = %v, got %v", expectedMaxY, bounds.Max.Y)
	}
}

func TestBarErrorBarsPlacementLikeMatplotlib(t *testing.T) {
	// Vertical bars: error anchors sit at the bar center (x) and top (baseline +
	// height), matching matplotlib bar()'s ex=left+0.5w, ey=bottom+h.
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	x := []float64{0, 1, 2}
	heights := []float64{3, 8, 6}
	yErr := []float64{0.5, 1.0, 0.75}

	bar := ax.Bar(x, heights, BarOptions{YErr: yErr})
	if bar == nil {
		t.Fatal("Bar() returned nil")
	}
	if bar.errorbar == nil {
		t.Fatal("expected an attached error bar when YErr is set")
	}
	eb := bar.errorbar
	if len(eb.XY) != len(x) {
		t.Fatalf("error bar point count = %d, want %d", len(eb.XY), len(x))
	}
	for i := range x {
		wantX, wantY := x[i], heights[i]
		if math.Abs(eb.XY[i].X-wantX) > 1e-12 || math.Abs(eb.XY[i].Y-wantY) > 1e-12 {
			t.Fatalf("vertical error anchor[%d] = (%v, %v), want (%v, %v)",
				i, eb.XY[i].X, eb.XY[i].Y, wantX, wantY)
		}
	}
	// ecolor defaults to black ('k').
	if (eb.Color != render.Color{R: 0, G: 0, B: 0, A: 1}) {
		t.Fatalf("default error color = %+v, want black", eb.Color)
	}
	// fmt="none": no connecting data line.
	if !eb.NoDataLine {
		t.Fatal("error bar should suppress its data line (NoDataLine)")
	}
	// The container surfaces the error bar.
	if c := bar.Container(); c == nil || c.Errorbar == nil || c.Errorbar.Errorbar != eb {
		t.Fatal("BarContainer should expose the attached error bar")
	}
	// Autoscale includes the upward error reach: top bar 8 + err 1.0 = 9.
	if _, yMax := ax.YScale.Domain(); yMax < 9.0-1e-9 {
		t.Fatalf("autoscale y-max = %v, want >= 9 (error reach included)", yMax)
	}

	// Horizontal bars: anchors at the bar end (x) and center (y).
	figH := NewFigure(800, 600)
	axH := figH.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	xErr := []float64{0.4, 0.6}
	barH := axH.BarH([]float64{0, 1}, []float64{4, 7}, BarOptions{XErr: xErr})
	if barH == nil || barH.errorbar == nil {
		t.Fatal("BarH with XErr should attach an error bar")
	}
	wantAnchors := []geom.Pt{{X: 4, Y: 0}, {X: 7, Y: 1}}
	for i, want := range wantAnchors {
		got := barH.errorbar.XY[i]
		if math.Abs(got.X-want.X) > 1e-12 || math.Abs(got.Y-want.Y) > 1e-12 {
			t.Fatalf("horizontal error anchor[%d] = %+v, want %+v", i, got, want)
		}
	}
}

func TestBarWithoutErrorDataHasNoErrorBar(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	bar := ax.Bar([]float64{0, 1, 2}, []float64{3, 8, 6})
	if bar.errorbar != nil {
		t.Fatal("bar without error data should not create an error bar")
	}
	if c := bar.Container(); c.Errorbar != nil {
		t.Fatal("container should have a nil Errorbar when no error data was given")
	}
}
