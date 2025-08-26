package core

import (
	"testing"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
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
