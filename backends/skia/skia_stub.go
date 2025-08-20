//go:build !skia

package skia

import (
	"errors"

	"matplotlib-go/backends"
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

// Stub implementation for when Skia is not available

// New returns an error when Skia support is not compiled in.
func New(config backends.Config) (*Renderer, error) {
	return nil, errors.New("skia backend not available: build with -tags skia")
}

// Renderer is a stub type for non-Skia builds.
type Renderer struct{}

// All methods return errors indicating Skia is not available

func (r *Renderer) Begin(_ geom.Rect) error {
	return errors.New("skia backend not available")
}

func (r *Renderer) End() error {
	return errors.New("skia backend not available")
}

func (r *Renderer) Save() {
	// No-op
}

func (r *Renderer) Restore() {
	// No-op  
}

func (r *Renderer) ClipRect(_ geom.Rect) {
	// No-op
}

func (r *Renderer) ClipPath(_ geom.Path) {
	// No-op
}

func (r *Renderer) Path(_ geom.Path, _ *render.Paint) {
	// No-op
}

func (r *Renderer) Image(_ render.Image, _ geom.Rect) {
	// No-op
}

func (r *Renderer) GlyphRun(_ render.GlyphRun, _ render.Color) {
	// No-op
}

func (r *Renderer) MeasureText(_ string, _ float64, _ string) render.TextMetrics {
	return render.TextMetrics{}
}

func (r *Renderer) SavePNG(_ string) error {
	return errors.New("skia backend not available")
}