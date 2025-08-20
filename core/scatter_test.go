package core

import (
	"testing"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
	"matplotlib-go/style"
	"matplotlib-go/transform"
)

// createTestDrawContext creates a valid DrawContext for testing
func createTestDrawContext() *DrawContext {
	return &DrawContext{
		DataToPixel: Transform2D{
			XScale:      transform.NewLinear(0, 10),
			YScale:      transform.NewLinear(0, 10),
			AxesToPixel: transform.NewAffine(geom.Affine{A: 100, D: -100, E: 50, F: 450}), // 500x500 viewport
		},
		RC:   style.Default,
		Clip: geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 500, Y: 500}},
	}
}

func TestScatter2D_Draw(t *testing.T) {
	// Create a scatter with basic data
	scatter := &Scatter2D{
		XY: []geom.Pt{
			{X: 0, Y: 0},
			{X: 1, Y: 1},
			{X: 2, Y: 2},
		},
		Size:   5.0,
		Color:  render.Color{R: 1, G: 0, B: 0, A: 1},
		Marker: MarkerCircle,
	}

	// Test that Draw doesn't panic with null renderer
	renderer := &render.NullRenderer{}

	// Create a proper DrawContext with transforms
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic
	scatter.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestScatter2D_EmptyData(t *testing.T) {
	// Test with empty data
	scatter := &Scatter2D{
		XY:     []geom.Pt{},
		Size:   5.0,
		Color:  render.Color{R: 1, G: 0, B: 0, A: 1},
		Marker: MarkerCircle,
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic with empty data
	scatter.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestScatter2D_VariableSizesAndColors(t *testing.T) {
	// Test with variable sizes and colors
	scatter := &Scatter2D{
		XY: []geom.Pt{
			{X: 0, Y: 0},
			{X: 1, Y: 1},
		},
		Sizes: []float64{3.0, 7.0},
		Colors: []render.Color{
			{R: 1, G: 0, B: 0, A: 1},
			{R: 0, G: 1, B: 0, A: 1},
		},
		Size:   5.0,                                  // fallback size
		Color:  render.Color{R: 0, G: 0, B: 1, A: 1}, // fallback color
		Marker: MarkerSquare,
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic
	scatter.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestScatter2D_AllMarkerTypes(t *testing.T) {
	markerTypes := []MarkerType{
		MarkerCircle, MarkerSquare, MarkerTriangle,
		MarkerDiamond, MarkerPlus, MarkerCross,
	}

	for _, markerType := range markerTypes {
		scatter := &Scatter2D{
			XY:     []geom.Pt{{X: 0, Y: 0}},
			Size:   5.0,
			Color:  render.Color{R: 1, G: 0, B: 0, A: 1},
			Marker: markerType,
		}

		renderer := &render.NullRenderer{}
		ctx := createTestDrawContext()

		err := renderer.Begin(geom.Rect{})
		if err != nil {
			t.Fatalf("Failed to begin rendering for marker %v: %v", markerType, err)
		}

		// Should not panic for any marker type
		scatter.Draw(renderer, ctx)

		err = renderer.End()
		if err != nil {
			t.Fatalf("Failed to end rendering for marker %v: %v", markerType, err)
		}
	}
}

func TestScatter2D_ZOrder(t *testing.T) {
	scatter := &Scatter2D{z: 3.5}

	if got := scatter.Z(); got != 3.5 {
		t.Errorf("Expected Z() = 3.5, got %v", got)
	}
}

func TestScatter2D_Bounds(t *testing.T) {
	scatter := &Scatter2D{}

	// Currently returns empty rect
	bounds := scatter.Bounds(nil)
	expected := geom.Rect{}

	if bounds != expected {
		t.Errorf("Expected Bounds() = %v, got %v", expected, bounds)
	}
}
