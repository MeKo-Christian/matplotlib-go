package gobasic

import (
	"errors"
	"image"
	"image/color"
	"image/png"
	"os"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"golang.org/x/image/vector"
)

// New creates a new GoBasic renderer with the specified dimensions and background color.
func New(w, h int, bg render.Color) *Renderer {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))

	// Fill with background color using premultiplied alpha
	red, green, blue, alpha := bg.ToPremultipliedRGBA()
	bgColor := color.RGBA{R: red, G: green, B: blue, A: alpha}

	// Fill the entire image with background color
	for y := 0; y < h; y++ {
		row := dst.PixOffset(0, y)
		for x := 0; x < w; x++ {
			i := row + x*4
			dst.Pix[i] = bgColor.R
			dst.Pix[i+1] = bgColor.G
			dst.Pix[i+2] = bgColor.B
			dst.Pix[i+3] = bgColor.A
		}
	}

	return &Renderer{
		dst:         dst,
		rasterizer:  vector.NewRasterizer(w, h),
		resolution:  72,
		clipMaskMap: make(map[clipMaskKey]*image.Alpha),
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
	r.clipPaths = r.clipPaths[:0]
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
	r.clipPaths = r.clipPaths[:0]
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
		clipRect:  clipCopy,
		clipPaths: clonePaths(r.clipPaths),
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
	r.clipPaths = clonePaths(s.clipPaths)
}

// ClipRect sets a rectangular clip region. The incoming rect is in y-up display
// space; it is flipped to the y-down device buffer so it matches the device-space
// pixel comparisons in fillPath/drawTargetRect.
func (r *Renderer) ClipRect(rect geom.Rect) {
	rect = r.devRect(rect)
	if r.clipRect == nil {
		r.clipRect = &rect
	} else {
		// Intersect with existing clip
		intersected := r.clipRect.Intersect(rect)
		r.clipRect = &intersected
	}
}

// ClipPath adds a path-based clip region to the current graphics state. The
// incoming path is in y-up display space; it is flipped to device space so the
// rasterized clip mask aligns with device pixels.
func (r *Renderer) ClipPath(p geom.Path) {
	if len(p.C) == 0 || !p.Validate() {
		return
	}
	r.clipPaths = append(r.clipPaths, clonePath(r.devPath(p)))
}

// SetResolution sets the text rendering resolution used for point-sized fonts.
func (r *Renderer) SetResolution(dpi uint) {
	if dpi > 0 {
		r.resolution = dpi
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
