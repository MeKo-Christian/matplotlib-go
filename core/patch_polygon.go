package core

import (
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// Polygon draws a closed polygon by default.
type Polygon struct {
	Patch
	XY     []geom.Pt
	Open   bool
	Coords CoordinateSpec
}

// Draw renders the polygon path using the embedded patch styling.
func (p *Polygon) Draw(ren render.Renderer, ctx *DrawContext) {
	if p == nil || ctx == nil || ren == nil || len(p.XY) < 2 {
		return
	}
	path := buildArtistDisplayPath(ctx, p, p.Coords, polygonPath(p.XY, !p.Open), geom.Identity())
	p.drawStyledPath(ren, &ctx.RC, path, geom.Path{})
}

// Bounds returns the polygon's data-space bounding box when applicable.
func (p *Polygon) Bounds(*DrawContext) geom.Rect {
	if p == nil || len(p.XY) == 0 || !artistUsesDataCoords(p, p.Coords) {
		return geom.Rect{}
	}
	bounds := geom.Rect{Min: p.XY[0], Max: p.XY[0]}
	for _, pt := range p.XY[1:] {
		bounds = expandRect(bounds, pt)
	}
	return bounds
}
