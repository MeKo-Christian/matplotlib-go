package core

import (
	"testing"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

func TestFill2D_Draw_FillBetween(t *testing.T) {
	// Create a fill between two curves
	fill := &Fill2D{
		X:     []float64{0, 1, 2, 3, 4},
		Y1:    []float64{1, 3, 2, 4, 1},
		Y2:    []float64{0, 1, 0.5, 2, 0},
		Color: render.Color{R: 0.5, G: 0.7, B: 0.9, A: 0.6},
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic
	fill.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestFill2D_Draw_FillToBaseline(t *testing.T) {
	// Create a fill to baseline
	fill := &Fill2D{
		X:        []float64{0, 1, 2, 3, 4},
		Y1:       []float64{2, 4, 3, 5, 2},
		Y2:       nil, // use baseline
		Baseline: 1.0,
		Color:    render.Color{R: 1, G: 0.5, B: 0.2, A: 0.8},
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic
	fill.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestFill2D_EmptyData(t *testing.T) {
	// Test with empty data
	fill := &Fill2D{
		X:     []float64{},
		Y1:    []float64{},
		Color: render.Color{R: 1, G: 0, B: 0, A: 1},
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic with empty data
	fill.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestFill2D_SinglePoint(t *testing.T) {
	// Test with single point (should not draw)
	fill := &Fill2D{
		X:     []float64{1},
		Y1:    []float64{2},
		Color: render.Color{R: 1, G: 0, B: 0, A: 1},
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic but also should not draw
	fill.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestFill2D_MismatchedLengths(t *testing.T) {
	// Test with mismatched array lengths
	fill := &Fill2D{
		X:     []float64{0, 1, 2, 3, 4}, // 5 elements
		Y1:    []float64{1, 3, 2},       // 3 elements
		Y2:    []float64{0, 1},          // 2 elements
		Color: render.Color{R: 0.5, G: 0.5, B: 1, A: 0.5},
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should only use min length (2 in this case)
	fill.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestFill2D_AlphaOverride(t *testing.T) {
	// Test alpha override
	fill := &Fill2D{
		X:     []float64{0, 1, 2},
		Y1:    []float64{1, 2, 1},
		Color: render.Color{R: 1, G: 0, B: 0, A: 1}, // original alpha = 1
		Alpha: 0.3,                                  // override to 0.3
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should use alpha override
	fill.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestFill2D_ZOrder(t *testing.T) {
	fill := &Fill2D{z: 1.5}

	if got := fill.Z(); got != 1.5 {
		t.Errorf("Expected Z() = 1.5, got %v", got)
	}
}

func TestFill2D_Bounds(t *testing.T) {
	fill := &Fill2D{}

	// Currently returns empty rect
	bounds := fill.Bounds(nil)
	expected := geom.Rect{}

	if bounds != expected {
		t.Errorf("Expected Bounds() = %v, got %v", expected, bounds)
	}
}

func TestFillBetween(t *testing.T) {
	x := []float64{0, 1, 2}
	y1 := []float64{1, 2, 1}
	y2 := []float64{0, 1, 0}
	color := render.Color{R: 1, G: 0, B: 0, A: 0.5}

	fill := FillBetween(x, y1, y2, color)

	if len(fill.X) != len(x) {
		t.Errorf("Expected X length %d, got %d", len(x), len(fill.X))
	}
	if len(fill.Y1) != len(y1) {
		t.Errorf("Expected Y1 length %d, got %d", len(y1), len(fill.Y1))
	}
	if len(fill.Y2) != len(y2) {
		t.Errorf("Expected Y2 length %d, got %d", len(y2), len(fill.Y2))
	}
	if fill.Color != color {
		t.Errorf("Expected color %v, got %v", color, fill.Color)
	}
}

func TestFillToBaseline(t *testing.T) {
	x := []float64{0, 1, 2}
	y := []float64{1, 2, 1}
	baseline := 0.5
	color := render.Color{R: 0, G: 1, B: 0, A: 0.7}

	fill := FillToBaseline(x, y, baseline, color)

	if len(fill.X) != len(x) {
		t.Errorf("Expected X length %d, got %d", len(x), len(fill.X))
	}
	if len(fill.Y1) != len(y) {
		t.Errorf("Expected Y1 length %d, got %d", len(y), len(fill.Y1))
	}
	if fill.Y2 != nil {
		t.Errorf("Expected Y2 to be nil, got %v", fill.Y2)
	}
	if fill.Baseline != baseline {
		t.Errorf("Expected baseline %v, got %v", baseline, fill.Baseline)
	}
	if fill.Color != color {
		t.Errorf("Expected color %v, got %v", color, fill.Color)
	}
}
