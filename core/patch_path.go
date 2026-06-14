package core

import (
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// PathPatch draws an arbitrary path in data, axes, or figure coordinates.
type PathPatch struct {
	Patch
	Path   geom.Path
	Coords CoordinateSpec
}

// Draw renders the path patch using the embedded patch styling.
func (p *PathPatch) Draw(ren render.Renderer, ctx *DrawContext) {
	if p == nil || ctx == nil || ren == nil || len(p.Path.C) == 0 {
		return
	}
	path := buildArtistDisplayPath(ctx, p, p.Coords, p.Path, geom.Identity())
	p.drawStyledPath(ren, path, geom.Path{})
}

// Bounds returns the path's data-space bounding box when applicable.
func (p *PathPatch) Bounds(*DrawContext) geom.Rect {
	if p == nil || len(p.Path.C) == 0 || !artistUsesDataCoords(p, p.Coords) {
		return geom.Rect{}
	}
	bounds, _ := pathBounds(p.Path)
	return bounds
}
