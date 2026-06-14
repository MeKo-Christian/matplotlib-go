package core

import (
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// Ellipse draws a rotated ellipse centered at Center.
type Ellipse struct {
	Patch
	Center geom.Pt
	Width  float64
	Height float64
	Angle  float64
	Coords CoordinateSpec
}

// Draw renders the ellipse path using the embedded patch styling.
func (e *Ellipse) Draw(ren render.Renderer, ctx *DrawContext) {
	if e == nil || ctx == nil || ren == nil || e.Width == 0 || e.Height == 0 {
		return
	}
	local := ellipsePath(e.Width, e.Height)
	path := buildArtistDisplayPath(ctx, e, e.Coords, local, patchAffine(e.Center, e.Angle))
	e.drawStyledPath(ren, path, geom.Path{})
}

// Bounds returns the ellipse's data-space bounding box when applicable.
func (e *Ellipse) Bounds(*DrawContext) geom.Rect {
	if e == nil || e.Width == 0 || e.Height == 0 || !artistUsesDataCoords(e, e.Coords) {
		return geom.Rect{}
	}
	path := applyAffinePath(ellipsePath(e.Width, e.Height), patchAffine(e.Center, e.Angle))
	bounds, _ := pathBounds(path)
	return bounds
}
