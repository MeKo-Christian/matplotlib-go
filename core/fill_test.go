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
	// Test empty fill
	fill := &Fill2D{}
	bounds := fill.Bounds(nil)
	expected := geom.Rect{}
	if bounds != expected {
		t.Errorf("Expected Bounds() = %v, got %v", expected, bounds)
	}

	// Test fill with data
	fill = &Fill2D{
		X:  []float64{0, 2, 1},
		Y1: []float64{1, 3, 2},
		Y2: []float64{0, 1, 0.5},
	}
	bounds = fill.Bounds(nil)
	
	// Should include all X and Y values
	if bounds.Min.X != 0 || bounds.Max.X != 2 {
		t.Errorf("Expected X bounds [0, 2], got [%v, %v]", bounds.Min.X, bounds.Max.X)
	}
	if bounds.Min.Y != 0 || bounds.Max.Y != 3 {
		t.Errorf("Expected Y bounds [0, 3], got [%v, %v]", bounds.Min.Y, bounds.Max.Y)
	}

	// Test with baseline
	fill = &Fill2D{
		X:        []float64{0, 1, 2},
		Y1:       []float64{2, 3, 1},
		Y2:       nil,
		Baseline: -1.0,
	}
	bounds = fill.Bounds(nil)
	
	// Should include baseline in Y bounds
	if bounds.Min.Y != -1.0 || bounds.Max.Y != 3 {
		t.Errorf("Expected Y bounds [-1, 3], got [%v, %v]", bounds.Min.Y, bounds.Max.Y)
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

func TestFill2D_EdgeColors(t *testing.T) {
	// Test with edge colors and width
	fill := &Fill2D{
		X:         []float64{0, 1, 2},
		Y1:        []float64{1, 2, 1},
		Color:     render.Color{R: 0.5, G: 0.5, B: 0.5, A: 1},
		EdgeColor: render.Color{R: 1, G: 0, B: 0, A: 1},
		EdgeWidth: 2.0,
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic with edge colors
	fill.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestFill2D_AlphaEdgeCases(t *testing.T) {
	testCases := []struct {
		name  string
		alpha float64
	}{
		{"Zero alpha", 0.0},
		{"Negative alpha", -0.5},
		{"Greater than 1 alpha", 1.5},
		{"Valid alpha", 0.5},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fill := &Fill2D{
				X:         []float64{0, 1, 2},
				Y1:        []float64{1, 2, 1},
				Color:     render.Color{R: 1, G: 0, B: 0, A: 1},
				EdgeColor: render.Color{R: 0, G: 1, B: 0, A: 1},
				EdgeWidth: 1.0,
				Alpha:     tc.alpha,
			}

			renderer := &render.NullRenderer{}
			ctx := createTestDrawContext()

			err := renderer.Begin(geom.Rect{})
			if err != nil {
				t.Fatalf("Failed to begin rendering: %v", err)
			}

			// Should not panic with edge case alpha values
			fill.Draw(renderer, ctx)

			err = renderer.End()
			if err != nil {
				t.Fatalf("Failed to end rendering: %v", err)
			}
		})
	}
}

func TestFill2D_LargeDataset(t *testing.T) {
	// Test performance with many points
	const numPoints = 10000
	x := make([]float64, numPoints)
	y1 := make([]float64, numPoints)
	y2 := make([]float64, numPoints)

	for i := 0; i < numPoints; i++ {
		x[i] = float64(i)
		y1[i] = float64(i % 100)
		y2[i] = float64((i % 50))
	}

	fill := &Fill2D{
		X:         x,
		Y1:        y1,
		Y2:        y2,
		Color:     render.Color{R: 0.3, G: 0.7, B: 0.9, A: 0.6},
		EdgeColor: render.Color{R: 0.1, G: 0.3, B: 0.5, A: 1},
		EdgeWidth: 1.0,
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should handle large dataset without issues
	fill.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}

	// Test bounds calculation with large dataset
	bounds := fill.Bounds(nil)
	if bounds == (geom.Rect{}) {
		t.Error("Expected non-empty bounds for large dataset")
	}

	// Check that bounds are reasonable
	if bounds.Min.X != 0 || bounds.Max.X != float64(numPoints-1) {
		t.Errorf("Expected X bounds [0, %d], got [%v, %v]", numPoints-1, bounds.Min.X, bounds.Max.X)
	}
}

func TestFill2D_NegativeValues(t *testing.T) {
	// Test with negative values
	fill := &Fill2D{
		X:         []float64{-2, -1, 0, 1, 2},
		Y1:        []float64{-1, 2, -0.5, 3, -2},
		Y2:        []float64{-3, -1, -2, 0, -4},
		Color:     render.Color{R: 1, G: 0.5, B: 0.2, A: 0.8},
		EdgeColor: render.Color{R: 0.7, G: 0.3, B: 0.1, A: 1},
		EdgeWidth: 1.5,
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should handle negative values correctly
	fill.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}

	// Test bounds with negative values
	bounds := fill.Bounds(nil)
	if bounds.Min.X != -2 || bounds.Max.X != 2 {
		t.Errorf("Expected X bounds [-2, 2], got [%v, %v]", bounds.Min.X, bounds.Max.X)
	}
	if bounds.Min.Y != -4 || bounds.Max.Y != 3 {
		t.Errorf("Expected Y bounds [-4, 3], got [%v, %v]", bounds.Min.Y, bounds.Max.Y)
	}
}
