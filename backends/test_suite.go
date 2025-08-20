package backends

import (
	"testing"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

// BackendTestSuite runs comprehensive tests against any backend.
// This ensures all backends behave consistently and correctly.
type BackendTestSuite struct {
	backend Backend
	config  Config
}

// NewTestSuite creates a test suite for the given backend.
func NewTestSuite(backend Backend, config Config) *BackendTestSuite {
	return &BackendTestSuite{
		backend: backend,
		config:  config,
	}
}

// RunAll executes all backend tests.
func (s *BackendTestSuite) RunAll(t *testing.T) {
	t.Run("BasicOperations", s.TestBasicOperations)
	t.Run("StateManagement", s.TestStateManagement)
	t.Run("Clipping", s.TestClipping)
	t.Run("PathDrawing", s.TestPathDrawing)
	t.Run("ErrorHandling", s.TestErrorHandling)
}

// TestBasicOperations verifies Begin/End lifecycle.
func (s *BackendTestSuite) TestBasicOperations(t *testing.T) {
	renderer, err := Create(s.backend, s.config)
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}

	viewport := geom.Rect{
		Min: geom.Pt{X: 0, Y: 0},
		Max: geom.Pt{X: float64(s.config.Width), Y: float64(s.config.Height)},
	}

	// Test Begin
	err = renderer.Begin(viewport)
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	// Test double Begin should fail
	err = renderer.Begin(viewport)
	if err == nil {
		t.Error("Double Begin should fail")
	}

	// Test End
	err = renderer.End()
	if err != nil {
		t.Fatalf("End failed: %v", err)
	}

	// Test End without Begin should fail
	err = renderer.End()
	if err == nil {
		t.Error("End without Begin should fail")
	}
}

// TestStateManagement verifies Save/Restore stack behavior.
func (s *BackendTestSuite) TestStateManagement(t *testing.T) {
	renderer, err := Create(s.backend, s.config)
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}

	viewport := geom.Rect{
		Min: geom.Pt{X: 0, Y: 0},
		Max: geom.Pt{X: float64(s.config.Width), Y: float64(s.config.Height)},
	}

	err = renderer.Begin(viewport)
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	defer renderer.End()

	// Test Save/Restore balance
	renderer.Save()
	renderer.Save()
	renderer.Save()

	renderer.Restore()
	renderer.Restore()
	renderer.Restore()

	// Extra Restore should not crash
	renderer.Restore()
	renderer.Restore()
}

// TestClipping verifies clipping operations.
func (s *BackendTestSuite) TestClipping(t *testing.T) {
	renderer, err := Create(s.backend, s.config)
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}

	viewport := geom.Rect{
		Min: geom.Pt{X: 0, Y: 0},
		Max: geom.Pt{X: float64(s.config.Width), Y: float64(s.config.Height)},
	}

	err = renderer.Begin(viewport)
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	defer renderer.End()

	// Test rectangular clipping
	clipRect := geom.Rect{
		Min: geom.Pt{X: 10, Y: 10},
		Max: geom.Pt{X: 100, Y: 100},
	}
	renderer.ClipRect(clipRect)

	// Test path clipping (may be no-op in some backends)
	clipPath := geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo, geom.LineTo, geom.ClosePath},
		V: []geom.Pt{{X: 0, Y: 0}, {X: 50, Y: 0}, {X: 50, Y: 50}},
	}
	renderer.ClipPath(clipPath)
}

// TestPathDrawing verifies basic path rendering.
func (s *BackendTestSuite) TestPathDrawing(t *testing.T) {
	renderer, err := Create(s.backend, s.config)
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}

	viewport := geom.Rect{
		Min: geom.Pt{X: 0, Y: 0},
		Max: geom.Pt{X: float64(s.config.Width), Y: float64(s.config.Height)},
	}

	err = renderer.Begin(viewport)
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	defer renderer.End()

	// Test simple line path
	linePath := geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo},
		V: []geom.Pt{{X: 10, Y: 10}, {X: 100, Y: 100}},
	}

	paint := &render.Paint{
		LineWidth: 2.0,
		LineJoin:  render.JoinMiter,
		LineCap:   render.CapRound,
		Stroke:    render.Color{R: 0, G: 0, B: 0, A: 1},
		Fill:      render.Color{R: 0, G: 0, B: 0, A: 0}, // No fill
	}

	renderer.Path(linePath, paint)

	// Test filled rectangle
	rectPath := geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo, geom.LineTo, geom.LineTo, geom.ClosePath},
		V: []geom.Pt{
			{X: 20, Y: 20}, {X: 80, Y: 20},
			{X: 80, Y: 80}, {X: 20, Y: 80},
		},
	}

	fillPaint := &render.Paint{
		LineWidth: 0,
		Fill:      render.Color{R: 0.5, G: 0.5, B: 0.5, A: 1},
		Stroke:    render.Color{R: 0, G: 0, B: 0, A: 0}, // No stroke
	}

	renderer.Path(rectPath, fillPaint)

	// Test invalid path (should not crash)
	invalidPath := geom.Path{
		C: []geom.Cmd{geom.LineTo}, // LineTo without MoveTo
		V: []geom.Pt{{X: 0, Y: 0}},
	}

	renderer.Path(invalidPath, paint)
}

// TestErrorHandling verifies graceful error handling.
func (s *BackendTestSuite) TestErrorHandling(t *testing.T) {
	renderer, err := Create(s.backend, s.config)
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}

	// Operations before Begin should either be ignored or return errors
	// (backend-specific behavior, but should not crash)
	renderer.Save()
	renderer.Restore()
	
	emptyPath := geom.Path{}
	emptyPaint := &render.Paint{}
	renderer.Path(emptyPath, emptyPaint)

	viewport := geom.Rect{
		Min: geom.Pt{X: 0, Y: 0},
		Max: geom.Pt{X: float64(s.config.Width), Y: float64(s.config.Height)},
	}

	err = renderer.Begin(viewport)
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	// Test nil paint (should not crash)
	testPath := geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo},
		V: []geom.Pt{{X: 0, Y: 0}, {X: 10, Y: 10}},
	}
	renderer.Path(testPath, nil)

	renderer.End()
}

// TestDefaultConfig creates a standard test configuration.
func TestDefaultConfig(width, height int) Config {
	return Config{
		Width:      width,
		Height:     height,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1}, // White background
		DPI:        72.0,
	}
}