package render

import (
	"errors"

	"matplotlib-go/internal/geom"
)

// Paint configures drawing style for paths.
type Paint struct {
	LineWidth  float64
	LineJoin   LineJoin
	LineCap    LineCap
	MiterLimit float64
	Stroke     Color
	Fill       Color
	Dashes     []float64 // on/off pairs, in user space units
}

// LineJoin controls how path joins are rendered.
type LineJoin uint8

const (
	JoinMiter LineJoin = iota
	JoinRound
	JoinBevel
)

// LineCap controls how path endpoints are rendered.
type LineCap uint8

const (
	CapButt LineCap = iota
	CapRound
	CapSquare
)

// Color is a simple RGBA in linear [0..1].
type Color struct{ R, G, B, A float64 }

// Premultiply returns a color with RGB components premultiplied by alpha.
// This ensures consistent blending behavior across backends.
func (c Color) Premultiply() Color {
	return Color{
		R: c.R * c.A,
		G: c.G * c.A,
		B: c.B * c.A,
		A: c.A,
	}
}

// ToPremultipliedRGBA converts to image/color.RGBA with premultiplied alpha.
// This is the preferred conversion for deterministic rendering.
func (c Color) ToPremultipliedRGBA() (uint8, uint8, uint8, uint8) {
	premul := c.Premultiply()
	return uint8(premul.R*255 + 0.5),
		uint8(premul.G*255 + 0.5),
		uint8(premul.B*255 + 0.5),
		uint8(premul.A*255 + 0.5)
}

// Glyph represents a single shaped glyph.
type Glyph struct {
	ID      uint32
	Advance float64
	Offset  geom.Pt
}

// GlyphRun represents a run of glyphs to render at a baseline origin.
type GlyphRun struct {
	Glyphs  []Glyph
	Origin  geom.Pt
	Size    float64
	FontKey string
}

// TextMetrics provides basic text measurements.
type TextMetrics struct{ W, H, Ascent, Descent float64 }

// Image is a minimal interface for raster images passed to renderers.
type Image interface {
	Size() (w, h int)
}

// Renderer defines the core drawing verbs.
type Renderer interface {
	Begin(viewport geom.Rect) error
	End() error

	// State stack
	Save()
	Restore()

	// Clipping
	ClipRect(r geom.Rect)
	ClipPath(p geom.Path)

	// Drawing
	Path(p geom.Path, paint *Paint)
	Image(img Image, dst geom.Rect)
	GlyphRun(run GlyphRun, color Color)
	MeasureText(text string, size float64, fontKey string) TextMetrics
}

// NullRenderer is a no-op renderer used for traversal/tests.
type NullRenderer struct {
	began  bool
	stack  int
	cstack int
}

var _ Renderer = (*NullRenderer)(nil)

// Begin starts a drawing session for the given viewport.
func (n *NullRenderer) Begin(_ geom.Rect) error {
	if n.began {
		return errors.New("Begin called twice")
	}
	n.began = true
	return nil
}

// End ends a drawing session.
func (n *NullRenderer) End() error {
	if !n.began {
		return errors.New("End called before Begin")
	}
	n.began = false
	n.stack = 0
	n.cstack = 0
	return nil
}

// Save pushes state.
func (n *NullRenderer) Save() { n.stack++ }

// Restore pops state; underflow is clamped to zero.
func (n *NullRenderer) Restore() {
	if n.stack > 0 {
		n.stack--
	}
}

// ClipRect pushes a rectangular clip.
func (n *NullRenderer) ClipRect(_ geom.Rect) { n.cstack++ }

// ClipPath pushes a path clip.
func (n *NullRenderer) ClipPath(_ geom.Path) { n.cstack++ }

// Path draws a path using the provided paint; no-op here.
func (n *NullRenderer) Path(_ geom.Path, _ *Paint) {}

// Image draws an image in the destination rectangle; no-op here.
func (n *NullRenderer) Image(_ Image, _ geom.Rect) {}

// GlyphRun draws a run of glyphs with the given color; no-op here.
func (n *NullRenderer) GlyphRun(_ GlyphRun, _ Color) {}

// MeasureText returns zero metrics in the null renderer.
func (n *NullRenderer) MeasureText(_ string, _ float64, _ string) TextMetrics { return TextMetrics{} }

// depth returns the current state stack depth (for tests).
func (n *NullRenderer) depth() int { return n.stack }
