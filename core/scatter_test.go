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
	// Test empty scatter
	scatter := &Scatter2D{}
	bounds := scatter.Bounds(nil)
	expected := geom.Rect{}
	if bounds != expected {
		t.Errorf("Expected Bounds() = %v, got %v", expected, bounds)
	}

	// Test scatter with points
	scatter = &Scatter2D{
		XY: []geom.Pt{
			{X: 0, Y: 0},
			{X: 2, Y: 1},
			{X: 1, Y: 2},
		},
		Size: 5.0,
	}
	bounds = scatter.Bounds(nil)
	
	// Should expand bounds by approximately the marker size
	if bounds.Min.X > 0 || bounds.Min.Y > 0 {
		t.Errorf("Expected bounds to expand below data minimum, got %v", bounds)
	}
	if bounds.Max.X < 2 || bounds.Max.Y < 2 {
		t.Errorf("Expected bounds to expand above data maximum, got %v", bounds)
	}

	// Test with variable sizes
	scatter = &Scatter2D{
		XY: []geom.Pt{
			{X: 0, Y: 0},
			{X: 1, Y: 1},
		},
		Sizes: []float64{3.0, 10.0}, // Max size is 10.0
		Size:  5.0,                  // fallback size
	}
	bounds = scatter.Bounds(nil)
	
	// Should use the maximum size for bounds expansion
	if bounds.Min.X >= 0 || bounds.Min.Y >= 0 {
		t.Errorf("Expected bounds to expand below origin with max size, got %v", bounds)
	}
}

func TestScatter2D_EdgeColors(t *testing.T) {
	// Test with edge colors and width
	scatter := &Scatter2D{
		XY: []geom.Pt{
			{X: 0, Y: 0},
			{X: 1, Y: 1},
		},
		EdgeColors: []render.Color{
			{R: 1, G: 0, B: 0, A: 1},
			{R: 0, G: 1, B: 0, A: 1},
		},
		EdgeColor: render.Color{R: 0, G: 0, B: 1, A: 1}, // fallback
		EdgeWidth: 2.0,
		Size:      5.0,
		Color:     render.Color{R: 0.5, G: 0.5, B: 0.5, A: 1},
		Marker:    MarkerCircle,
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic with edge colors
	scatter.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestScatter2D_AlphaTransparency(t *testing.T) {
	// Test with alpha transparency
	scatter := &Scatter2D{
		XY:     []geom.Pt{{X: 0, Y: 0}},
		Size:   5.0,
		Color:  render.Color{R: 1, G: 0, B: 0, A: 1},
		Alpha:  0.5, // 50% transparent
		Marker: MarkerCircle,
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic with alpha transparency
	scatter.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestScatter2D_AlphaEdgeCases(t *testing.T) {
	testCases := []struct {
		name  string
		alpha float64
	}{
		{"Zero alpha", 0.0},
		{"Negative alpha", -0.5},
		{"Greater than 1 alpha", 1.5},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			scatter := &Scatter2D{
				XY:     []geom.Pt{{X: 0, Y: 0}},
				Size:   5.0,
				Color:  render.Color{R: 1, G: 0, B: 0, A: 1},
				Alpha:  tc.alpha,
				Marker: MarkerCircle,
			}

			renderer := &render.NullRenderer{}
			ctx := createTestDrawContext()

			err := renderer.Begin(geom.Rect{})
			if err != nil {
				t.Fatalf("Failed to begin rendering: %v", err)
			}

			// Should not panic with edge case alpha values
			scatter.Draw(renderer, ctx)

			err = renderer.End()
			if err != nil {
				t.Fatalf("Failed to end rendering: %v", err)
			}
		})
	}
}

func TestScatter2D_LargeDataset(t *testing.T) {
	// Test performance with many points
	const numPoints = 10000
	points := make([]geom.Pt, numPoints)
	sizes := make([]float64, numPoints)
	colors := make([]render.Color, numPoints)

	for i := 0; i < numPoints; i++ {
		points[i] = geom.Pt{X: float64(i), Y: float64(i % 100)}
		sizes[i] = float64(3 + (i % 10))
		colors[i] = render.Color{
			R: float64(i%256) / 255.0,
			G: float64((i*2)%256) / 255.0,
			B: float64((i*3)%256) / 255.0,
			A: 1.0,
		}
	}

	scatter := &Scatter2D{
		XY:     points,
		Sizes:  sizes,
		Colors: colors,
		Marker: MarkerCircle,
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should handle large dataset without issues
	scatter.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}

	// Test bounds calculation with large dataset
	bounds := scatter.Bounds(nil)
	if bounds == (geom.Rect{}) {
		t.Error("Expected non-empty bounds for large dataset")
	}
}
