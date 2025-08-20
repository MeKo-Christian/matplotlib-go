package gobasic

import (
	"errors"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	"golang.org/x/image/vector"
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

// quantizationEpsilon is the precision limit for float values to ensure determinism.
// All floating point coordinates and measurements are snapped to this precision.
const quantizationEpsilon = 1e-6

// quantize snaps a float64 value to quantizationEpsilon precision to eliminate
// tiny differences that could lead to cross-platform rendering variations.
func quantize(v float64) float64 {
	return math.Round(v/quantizationEpsilon) * quantizationEpsilon
}

// quantizePt quantizes both X and Y coordinates of a point.
func quantizePt(p geom.Pt) geom.Pt {
	return geom.Pt{
		X: quantize(p.X),
		Y: quantize(p.Y),
	}
}

// quantizePath quantizes all vertices in a path for deterministic rendering.
func quantizePath(p geom.Path) geom.Path {
	result := geom.Path{
		C: make([]geom.Cmd, len(p.C)),
		V: make([]geom.Pt, len(p.V)),
	}

	copy(result.C, p.C)
	for i, v := range p.V {
		result.V[i] = quantizePt(v)
	}

	return result
}

// state represents a saved graphics state.
type state struct {
	clipRect *geom.Rect
}

// Renderer implements render.Renderer using pure Go dependencies.
type Renderer struct {
	dst        *image.RGBA
	viewport   geom.Rect
	began      bool
	stack      []state
	clipRect   *geom.Rect
	rasterizer *vector.Rasterizer
}

var _ render.Renderer = (*Renderer)(nil)

// New creates a new GoBasic renderer with the specified dimensions and background color.
func New(w, h int, bg render.Color) *Renderer {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))

	// Fill with background color using premultiplied alpha
	red, green, blue, alpha := bg.ToPremultipliedRGBA()
	bgColor := color.RGBA{R: red, G: green, B: blue, A: alpha}

	// Fill the entire image with background color
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dst.Set(x, y, bgColor)
		}
	}

	return &Renderer{
		dst:        dst,
		rasterizer: vector.NewRasterizer(w, h),
	}
}

// Begin starts a drawing session with the given viewport.
func (r *Renderer) Begin(viewport geom.Rect) error {
	if r.began {
		return errors.New("Begin called twice")
	}
	r.began = true
	r.viewport = viewport
	r.stack = r.stack[:0]
	r.clipRect = nil
	return nil
}

// End finishes the drawing session.
func (r *Renderer) End() error {
	if !r.began {
		return errors.New("End called before Begin")
	}
	r.began = false
	r.stack = r.stack[:0]
	r.clipRect = nil
	return nil
}

// Save pushes the current graphics state onto the stack.
func (r *Renderer) Save() {
	var clipCopy *geom.Rect
	if r.clipRect != nil {
		rectCopy := *r.clipRect
		clipCopy = &rectCopy
	}
	r.stack = append(r.stack, state{
		clipRect: clipCopy,
	})
}

// Restore pops the graphics state from the stack.
func (r *Renderer) Restore() {
	if len(r.stack) == 0 {
		return // No state to restore
	}

	// Pop the last state
	s := r.stack[len(r.stack)-1]
	r.stack = r.stack[:len(r.stack)-1]

	// Restore state
	r.clipRect = s.clipRect
}

// ClipRect sets a rectangular clip region.
func (r *Renderer) ClipRect(rect geom.Rect) {
	if r.clipRect == nil {
		r.clipRect = &rect
	} else {
		// Intersect with existing clip
		intersected := r.clipRect.Intersect(rect)
		r.clipRect = &intersected
	}
}

// ClipPath sets a path-based clip region (stub implementation for Phase B).
func (r *Renderer) ClipPath(p geom.Path) {
	// For Phase B, we only support rectangular clipping
	// This is a no-op for now
}

// Path draws a path with the given paint style.
func (r *Renderer) Path(p geom.Path, paint *render.Paint) {
	if !p.Validate() {
		return // Invalid path
	}

	// Quantize path coordinates for deterministic rendering
	p = quantizePath(p)

	// Quantize paint parameters for consistency
	quantizedPaint := &render.Paint{
		LineWidth:  quantize(paint.LineWidth),
		LineJoin:   paint.LineJoin,
		LineCap:    paint.LineCap,
		MiterLimit: quantize(paint.MiterLimit),
		Stroke:     paint.Stroke,
		Fill:       paint.Fill,
		Dashes:     make([]float64, len(paint.Dashes)),
	}

	// Quantize dash pattern
	for i, dash := range paint.Dashes {
		quantizedPaint.Dashes[i] = quantize(dash)
	}

	// Fill first if requested
	if quantizedPaint.Fill.A > 0 {
		r.fillPath(p, quantizedPaint.Fill)
	}

	// Then stroke if requested
	if quantizedPaint.Stroke.A > 0 && quantizedPaint.LineWidth > 0 {
		r.drawStroke(p, quantizedPaint)
	}
}

// fillPath fills a path with the given color.
func (r *Renderer) fillPath(p geom.Path, fillColor render.Color) {
	// Reset and rebuild path for filling
	r.rasterizer.Reset(r.dst.Bounds().Dx(), r.dst.Bounds().Dy())

	vi := 0 // vertex index

	for _, cmd := range p.C {
		switch cmd {
		case geom.MoveTo:
			pt := p.V[vi]
			// Apply explicit rounding to ensure consistent float32 conversion
			r.rasterizer.MoveTo(float32(math.Round(pt.X*1e6)/1e6), float32(math.Round(pt.Y*1e6)/1e6))
			vi++
		case geom.LineTo:
			pt := p.V[vi]
			r.rasterizer.LineTo(float32(math.Round(pt.X*1e6)/1e6), float32(math.Round(pt.Y*1e6)/1e6))
			vi++
		case geom.QuadTo:
			ctrl := p.V[vi]
			to := p.V[vi+1]
			r.rasterizer.QuadTo(
				float32(math.Round(ctrl.X*1e6)/1e6), float32(math.Round(ctrl.Y*1e6)/1e6),
				float32(math.Round(to.X*1e6)/1e6), float32(math.Round(to.Y*1e6)/1e6))
			vi += 2
		case geom.CubicTo:
			c1 := p.V[vi]
			c2 := p.V[vi+1]
			to := p.V[vi+2]
			r.rasterizer.CubeTo(
				float32(math.Round(c1.X*1e6)/1e6), float32(math.Round(c1.Y*1e6)/1e6),
				float32(math.Round(c2.X*1e6)/1e6), float32(math.Round(c2.Y*1e6)/1e6),
				float32(math.Round(to.X*1e6)/1e6), float32(math.Round(to.Y*1e6)/1e6))
			vi += 3
		case geom.ClosePath:
			r.rasterizer.ClosePath()
		}
	}

	// Draw the filled path using premultiplied alpha
	red, green, blue, alpha := fillColor.ToPremultipliedRGBA()
	c := color.RGBA{R: red, G: green, B: blue, A: alpha}

	// Apply clipping if set
	bounds := r.dst.Bounds()
	if r.clipRect != nil {
		clipBounds := image.Rect(
			int(math.Floor(r.clipRect.Min.X)),
			int(math.Floor(r.clipRect.Min.Y)),
			int(math.Ceil(r.clipRect.Max.X)),
			int(math.Ceil(r.clipRect.Max.Y)),
		)
		bounds = bounds.Intersect(clipBounds)
	}

	r.rasterizer.Draw(r.dst, bounds, image.NewUniform(c), image.Point{})
}

// drawStroke handles stroke drawing for paths using proper stroke geometry.
func (r *Renderer) drawStroke(p geom.Path, paint *render.Paint) {
	// Convert stroke to filled path with proper joins, caps, and dashes
	strokePath := strokeToPath(p, paint)
	if len(strokePath.C) == 0 {
		return // No stroke geometry generated
	}

	// Fill the stroke geometry with the stroke color
	r.fillPath(strokePath, paint.Stroke)
}

// Image draws an image within the destination rectangle (stub implementation).
func (r *Renderer) Image(img render.Image, dst geom.Rect) {
	// Stub implementation for Phase B
	// Will be implemented in later phases
}

// GlyphRun draws a run of glyphs using basicfont.
// Note: This is a basic implementation that works with the available glyph information.
func (r *Renderer) GlyphRun(run render.GlyphRun, textColor render.Color) {
	if len(run.Glyphs) == 0 {
		return
	}

	// For a basic implementation, we can't easily map glyph IDs back to characters
	// without additional font metadata. This method provides the interface but
	// requires higher-level text drawing to work through other means.
	// A complete implementation would need a glyph ID to character mapping.
	
	// For now, this is a stub that maintains the interface contract
	// Real text rendering would happen through higher-level text drawing functions
}

// MeasureText measures text dimensions using basicfont.
func (r *Renderer) MeasureText(text string, size float64, fontKey string) render.TextMetrics {
	if text == "" {
		return render.TextMetrics{}
	}

	// Use basicfont.Face7x13 as the default font for now
	face := basicfont.Face7x13
	metrics := face.Metrics()

	// Calculate text width by summing character advances
	width := 0
	for _, ch := range text {
		advance, ok := face.GlyphAdvance(ch)
		if ok {
			width += int(advance >> 6) // Convert from fixed.Int26_6 to pixels
		} else {
			// Use advance for missing character replacement
			width += face.Advance
		}
	}

	// Convert fixed.Int26_6 metrics to float64
	height := float64(metrics.Height >> 6)
	ascent := float64(metrics.Ascent >> 6)
	descent := float64(metrics.Descent >> 6)

	// Apply size scaling - basicfont.Face7x13 is a fixed-size font,
	// so we scale the measurements by the requested size relative to the font's natural size
	scale := size / 13.0 // Face7x13 is 13 pixels tall
	
	return render.TextMetrics{
		W:       quantize(float64(width) * scale),
		H:       quantize(height * scale),
		Ascent:  quantize(ascent * scale),
		Descent: quantize(descent * scale),
	}
}

// GetImage returns the underlying image.RGBA for PNG export.
func (r *Renderer) GetImage() *image.RGBA {
	return r.dst
}

// SavePNG saves the rendered image to a PNG file.
func (r *Renderer) SavePNG(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, r.dst)
}

// DrawText is a helper method to draw text directly (not part of the Renderer interface).
// This provides a practical way to render text using basicfont.
func (r *Renderer) DrawText(text string, origin geom.Pt, size float64, textColor render.Color) {
	if text == "" {
		return
	}

	// Use basicfont.Face7x13 as the default font
	face := basicfont.Face7x13

	// Convert text color to image.Image for font.Drawer
	red, green, blue, alpha := textColor.ToPremultipliedRGBA()
	src := image.NewUniform(color.RGBA{R: red, G: green, B: blue, A: alpha})

	// Create font drawer
	drawer := &font.Drawer{
		Dst:  r.dst,
		Src:  src,
		Face: face,
	}

	// Quantize origin for deterministic rendering
	origin = geom.Pt{X: quantize(origin.X), Y: quantize(origin.Y)}

	// Convert to fixed.Point26_6 and set as drawer dot
	// Note: font coordinates have Y increasing downward, but we expect Y increasing upward
	// So we need to flip the Y coordinate
	bounds := r.dst.Bounds()
	drawer.Dot = fixed.Point26_6{
		X: fixed.Int26_6(origin.X * 64), // Convert to fixed point
		Y: fixed.Int26_6((float64(bounds.Max.Y) - origin.Y) * 64), // Flip Y coordinate
	}

	// Apply clipping if set
	if r.clipRect != nil {
		// Simple clipping check - only draw if the text origin is within clip bounds
		if origin.X < r.clipRect.Min.X || origin.X > r.clipRect.Max.X ||
			origin.Y < r.clipRect.Min.Y || origin.Y > r.clipRect.Max.Y {
			return
		}
	}

	// Draw the text
	drawer.DrawString(text)
}
