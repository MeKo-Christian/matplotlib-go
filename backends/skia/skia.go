//go:build skia

package skia

import (
	"errors"
	"fmt"

	"matplotlib-go/backends"
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

// Renderer implements render.Renderer using Skia graphics library.
// This is a stub implementation - actual Skia integration pending.
type Renderer struct {
	width      int
	height     int
	background render.Color
	useGPU     bool
	samples    int
	began      bool
	stack      []state
}

type state struct {
	// Skia graphics state would go here
	// transform matrix, clip regions, etc.
}

var _ render.Renderer = (*Renderer)(nil)

// New creates a new Skia renderer with the given configuration.
func New(config backends.Config) (*Renderer, error) {
	skiaConfig, ok := config.Options.(backends.SkiaConfig)
	if !ok {
		// Use defaults if no Skia-specific config provided
		skiaConfig = backends.SkiaConfig{
			UseGPU:      false,
			SampleCount: 1,
			ColorType:   "RGBA8888",
		}
	}

	// TODO: Initialize Skia context here
	// - Create GrDirectContext for GPU rendering if requested
	// - Set up SkSurface with appropriate ColorType and SampleCount
	// - Configure anti-aliasing and text rendering

	return &Renderer{
		width:      config.Width,
		height:     config.Height,
		background: config.Background,
		useGPU:     skiaConfig.UseGPU,
		samples:    skiaConfig.SampleCount,
	}, nil
}

// Begin starts a drawing session with the given viewport.
func (r *Renderer) Begin(viewport geom.Rect) error {
	if r.began {
		return errors.New("Begin called twice")
	}
	r.began = true
	r.stack = r.stack[:0]

	// TODO: Skia-specific initialization
	// - Clear surface with background color
	// - Set up viewport transform
	// - Reset graphics state

	return nil
}

// End finishes the drawing session.
func (r *Renderer) End() error {
	if !r.began {
		return errors.New("End called before Begin")
	}
	r.began = false
	r.stack = r.stack[:0]

	// TODO: Skia-specific cleanup
	// - Flush pending operations
	// - Sync GPU if using GPU backend

	return nil
}

// Save pushes the current graphics state onto the stack.
func (r *Renderer) Save() {
	// TODO: Use SkCanvas::save()
	r.stack = append(r.stack, state{})
}

// Restore pops the graphics state from the stack.
func (r *Renderer) Restore() {
	if len(r.stack) == 0 {
		return
	}
	r.stack = r.stack[:len(r.stack)-1]

	// TODO: Use SkCanvas::restore()
}

// ClipRect sets a rectangular clip region.
func (r *Renderer) ClipRect(rect geom.Rect) {
	// TODO: Use SkCanvas::clipRect()
	// Convert geom.Rect to SkRect and apply clipping
}

// ClipPath sets a path-based clip region.
func (r *Renderer) ClipPath(p geom.Path) {
	// TODO: Use SkCanvas::clipPath()
	// Convert geom.Path to SkPath and apply clipping
}

// Path draws a path with the given paint style.
func (r *Renderer) Path(p geom.Path, paint *render.Paint) {
	if !p.Validate() {
		return
	}

	// TODO: Skia path rendering
	// - Convert geom.Path to SkPath
	// - Convert render.Paint to SkPaint
	// - Handle fills and strokes with proper anti-aliasing
	// - Apply dash patterns, line caps, joins
}

// Image draws an image within the destination rectangle.
func (r *Renderer) Image(img render.Image, dst geom.Rect) {
	// TODO: Skia image rendering
	// - Convert render.Image to SkImage
	// - Use SkCanvas::drawImageRect() with filtering
}

// GlyphRun draws a run of glyphs.
func (r *Renderer) GlyphRun(run render.GlyphRun, color render.Color) {
	// TODO: Skia text rendering
	// - Use SkTextBlob for shaped text
	// - Apply proper hinting and subpixel positioning
	// - Handle font loading and caching
}

// MeasureText measures text dimensions.
func (r *Renderer) MeasureText(text string, size float64, fontKey string) render.TextMetrics {
	// TODO: Skia text measurement
	// - Use SkFont for metrics calculation
	// - Return accurate bounds, ascent, descent
	return render.TextMetrics{}
}

// GetSurface returns the underlying Skia surface for advanced operations.
// This would return *skia.Surface or similar when implemented.
func (r *Renderer) GetSurface() interface{} {
	// TODO: Return actual SkSurface
	return nil
}

// SavePNG saves the rendered image to a PNG file.
func (r *Renderer) SavePNG(path string) error {
	// TODO: Use Skia's image encoding
	// - Create SkImage from surface
	// - Encode as PNG with appropriate compression
	return fmt.Errorf("Skia backend not implemented - use gobasic backend")
}

// FlushGPU flushes pending GPU operations (if using GPU backend).
func (r *Renderer) FlushGPU() {
	if !r.useGPU {
		return
	}
	// TODO: Call GrDirectContext::flushAndSubmit()
}

// GPU returns true if this renderer is using GPU acceleration.
func (r *Renderer) GPU() bool {
	return r.useGPU
}

// SampleCount returns the MSAA sample count.
func (r *Renderer) SampleCount() int {
	return r.samples
}