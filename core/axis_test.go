package core

import (
	"testing"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

func TestAxis_Draw(t *testing.T) {
	// Test drawing a basic X axis
	axis := NewXAxis()

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic
	axis.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestAxis_YAxis(t *testing.T) {
	// Test drawing a basic Y axis
	axis := NewYAxis()

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic
	axis.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestAxis_CustomSettings(t *testing.T) {
	// Test axis with custom settings
	axis := &Axis{
		Side:       AxisTop,
		Locator:    LinearLocator{},
		Formatter:  ScalarFormatter{Prec: 2},
		Color:      render.Color{R: 1, G: 0, B: 0, A: 1}, // red
		LineWidth:  2.0,
		TickSize:   10.0,
		ShowSpine:  true,
		ShowTicks:  true,
		ShowLabels: false,
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic
	axis.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestAxis_DisabledComponents(t *testing.T) {
	// Test axis with components disabled
	axis := NewXAxis()
	axis.ShowSpine = false
	axis.ShowTicks = false
	axis.ShowLabels = false

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic even with everything disabled
	axis.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestAxes_SetLimits(t *testing.T) {
	// Test the convenience methods for setting limits
	axes := &Axes{
		XScale: nil,
		YScale: nil,
		XAxis:  NewXAxis(),
		YAxis:  NewYAxis(),
	}

	// Test SetXLim
	axes.SetXLim(-5, 10)
	xMin, xMax := axes.XScale.Domain()
	if xMin != -5 || xMax != 10 {
		t.Errorf("SetXLim failed: expected (-5, 10), got (%v, %v)", xMin, xMax)
	}

	// Test SetYLim
	axes.SetYLim(0, 100)
	yMin, yMax := axes.YScale.Domain()
	if yMin != 0 || yMax != 100 {
		t.Errorf("SetYLim failed: expected (0, 100), got (%v, %v)", yMin, yMax)
	}

	// Test SetXLimLog
	axes.SetXLimLog(1, 1000, 10)
	xMin, xMax = axes.XScale.Domain()
	if xMin != 1 || xMax != 1000 {
		t.Errorf("SetXLimLog failed: expected (1, 1000), got (%v, %v)", xMin, xMax)
	}

	// Check that locator was updated to logarithmic
	if logLoc, ok := axes.XAxis.Locator.(LogLocator); !ok || logLoc.Base != 10 {
		t.Errorf("SetXLimLog should update locator to LogLocator with base 10")
	}
}

func TestGrid_Draw(t *testing.T) {
	// Test grid drawing
	grid := NewGrid(AxisBottom)

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic
	grid.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestGrid_Disabled(t *testing.T) {
	// Test grid with major disabled
	grid := NewGrid(AxisLeft)
	grid.Major = false

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not draw anything
	grid.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestAxes_AddGrid(t *testing.T) {
	// Test adding grids to axes
	axes := &Axes{
		Artists: []Artist{},
		XAxis:   NewXAxis(),
		YAxis:   NewYAxis(),
	}

	initialCount := len(axes.Artists)

	// Add X grid
	xGrid := axes.AddXGrid()
	if len(axes.Artists) != initialCount+1 {
		t.Errorf("AddXGrid should add one artist, got %d artists", len(axes.Artists))
	}
	if xGrid.Axis != AxisBottom {
		t.Errorf("AddXGrid should create grid for AxisBottom, got %v", xGrid.Axis)
	}

	// Add Y grid
	yGrid := axes.AddYGrid()
	if len(axes.Artists) != initialCount+2 {
		t.Errorf("AddYGrid should add second artist, got %d artists", len(axes.Artists))
	}
	if yGrid.Axis != AxisLeft {
		t.Errorf("AddYGrid should create grid for AxisLeft, got %v", yGrid.Axis)
	}
}
