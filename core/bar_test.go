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
		Y:           []float64{5, 8, 3},
		BarWidth:    0.8,
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
		X:           []float64{5, 8, 3},
		Y:           []float64{1, 2, 3},
		BarWidth:    0.8,
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
		Y:           []float64{},
		BarWidth:    0.8,
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
	// Test with mismatched X and Y lengths
	bar := &Bar2D{
		X:           []float64{1, 2, 3, 4, 5}, // 5 elements
		Y:           []float64{5, 8},          // 2 elements
		BarWidth:    0.8,
		Color:       render.Color{R: 1, G: 0, B: 0, A: 1},
		Orientation: BarVertical,
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should only draw min(len(X), len(Y)) bars
	bar.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestBar2D_VariableWidthsAndColors(t *testing.T) {
	// Test with variable widths and colors
	bar := &Bar2D{
		X:     []float64{1, 2, 3},
		Y:     []float64{5, 8, 3},
		Width: []float64{0.5, 1.0, 0.3},
		Colors: []render.Color{
			{R: 1, G: 0, B: 0, A: 1},
			{R: 0, G: 1, B: 0, A: 1},
			{R: 0, G: 0, B: 1, A: 1},
		},
		BarWidth:    0.8,                                        // fallback
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
		Y:           []float64{-2, 5, -1},
		BarWidth:    0.8,
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

func TestBar2D_ZOrder(t *testing.T) {
	bar := &Bar2D{z: 2.5}

	if got := bar.Z(); got != 2.5 {
		t.Errorf("Expected Z() = 2.5, got %v", got)
	}
}

func TestBar2D_Bounds(t *testing.T) {
	bar := &Bar2D{}

	// Currently returns empty rect
	bounds := bar.Bounds(nil)
	expected := geom.Rect{}

	if bounds != expected {
		t.Errorf("Expected Bounds() = %v, got %v", expected, bounds)
	}
}
