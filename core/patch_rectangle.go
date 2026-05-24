package core

import (
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// Rectangle draws an axis-aligned or rotated rectangle.
type Rectangle struct {
	Patch
	XY     geom.Pt
	Width  float64
	Height float64
	Angle  float64
	Coords CoordinateSpec
}

// Draw renders the rectangle path using the embedded patch styling.
func (r *Rectangle) Draw(ren render.Renderer, ctx *DrawContext) {
	if r == nil || ctx == nil || ren == nil || r.Width == 0 || r.Height == 0 {
		return
	}
	path := buildArtistDisplayPath(ctx, r, r.Coords, rectanglePath(r.Width, r.Height), patchAffine(r.XY, r.Angle))
	r.drawStyledPath(ren, path, geom.Path{})
}

// Bounds returns the rectangle's data-space bounding box when applicable.
func (r *Rectangle) Bounds(*DrawContext) geom.Rect {
	if r == nil || r.Width == 0 || r.Height == 0 || !artistUsesDataCoords(r, r.Coords) {
		return geom.Rect{}
	}
	path := applyAffinePath(rectanglePath(r.Width, r.Height), patchAffine(r.XY, r.Angle))
	bounds, _ := pathBounds(path)
	return bounds
}
