package core

import (
	"testing"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
	"matplotlib-go/style"
	"matplotlib-go/transform"
)

func TestLine2D_EmptyData(t *testing.T) {
	line := &Line2D{
		XY:  []geom.Pt{}, // empty data
		W:   2.0,
		Col: render.Color{R: 1, G: 0, B: 0, A: 1},
	}

	// Should not panic with empty data
	var r render.NullRenderer
	ctx := &DrawContext{
		DataToPixel: Transform2D{
			XScale:      transform.NewLinear(0, 10),
			YScale:      transform.NewLinear(0, 1),
			AxesToPixel: transform.NewAffine(geom.Identity()),
		},
		RC:   style.Default,
		Clip: geom.Rect{},
	}

	// This should not panic
	line.Draw(&r, ctx)
}

func TestLine2D_SingletonData(t *testing.T) {
	line := &Line2D{
		XY:  []geom.Pt{{X: 5, Y: 0.5}}, // single point
		W:   2.0,
		Col: render.Color{R: 1, G: 0, B: 0, A: 1},
	}

	// Should not panic with singleton data
	var r render.NullRenderer
	ctx := &DrawContext{
		DataToPixel: Transform2D{
			XScale:      transform.NewLinear(0, 10),
			YScale:      transform.NewLinear(0, 1),
			AxesToPixel: transform.NewAffine(geom.Identity()),
		},
		RC:   style.Default,
		Clip: geom.Rect{},
	}

	// This should not panic
	line.Draw(&r, ctx)
}

func TestLine2D_BasicFunctionality(t *testing.T) {
	line := &Line2D{
		XY: []geom.Pt{
			{X: 0, Y: 0},
			{X: 1, Y: 0.2},
			{X: 3, Y: 0.9},
			{X: 6, Y: 0.4},
			{X: 10, Y: 0.8},
		},
		W:   2.0,
		Col: render.Color{R: 0, G: 0, B: 0, A: 1},
		z:   1.0,
	}

	// Test Z() method
	if line.Z() != 1.0 {
		t.Errorf("Expected Z() = 1.0, got %f", line.Z())
	}

	// Test Bounds() method (should return empty rect for now)
	bounds := line.Bounds(nil)
	if bounds.Min.X != 0 || bounds.Min.Y != 0 || bounds.Max.X != 0 || bounds.Max.Y != 0 {
		t.Errorf("Expected empty bounds, got %+v", bounds)
	}

	// Test Draw() method doesn't panic
	var r render.NullRenderer
	ctx := &DrawContext{
		DataToPixel: Transform2D{
			XScale:      transform.NewLinear(0, 10),
			YScale:      transform.NewLinear(0, 1),
			AxesToPixel: transform.NewAffine(geom.Identity()),
		},
		RC:   style.Default,
		Clip: geom.Rect{},
	}

	// This should not panic
	line.Draw(&r, ctx)
}

func TestLine2D_AsArtist(t *testing.T) {
	line := &Line2D{
		XY:  []geom.Pt{{X: 0, Y: 0}, {X: 1, Y: 1}},
		W:   1.0,
		Col: render.Color{R: 1, G: 1, B: 1, A: 1},
	}

	// Test that Line2D implements Artist interface
	var _ Artist = line

	// Test integration with Axes
	fig := NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.15}, Max: geom.Pt{X: 0.95, Y: 0.9}})
	ax.XScale = transform.NewLinear(0, 10)
	ax.YScale = transform.NewLinear(0, 1)
	ax.Add(line)

	// Test that the figure can be drawn without panic
	var r render.NullRenderer
	DrawFigure(fig, &r)
}
