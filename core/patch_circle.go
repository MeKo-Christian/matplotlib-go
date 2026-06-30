package core

import (
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// Circle draws a circle centered at Center with Radius in the chosen coords.
type Circle struct {
	Patch
	Center geom.Pt
	Radius float64
	Coords CoordinateSpec
}

// Draw renders the circle path using the embedded patch styling.
func (c *Circle) Draw(ren render.Renderer, ctx *DrawContext) {
	if c == nil || ctx == nil || ren == nil || c.Radius <= 0 {
		return
	}
	local := ellipsePath(c.Radius*2, c.Radius*2)
	path := buildArtistDisplayPath(ctx, c, c.Coords, local, translateAffine(c.Center))
	c.drawStyledPath(ren, &ctx.RC, path, geom.Path{})
}

// Bounds returns the circle's data-space bounding box when applicable.
func (c *Circle) Bounds(*DrawContext) geom.Rect {
	if c == nil || c.Radius <= 0 || !artistUsesDataCoords(c, c.Coords) {
		return geom.Rect{}
	}
	return geom.Rect{
		Min: geom.Pt{X: c.Center.X - c.Radius, Y: c.Center.Y - c.Radius},
		Max: geom.Pt{X: c.Center.X + c.Radius, Y: c.Center.Y + c.Radius},
	}
}
