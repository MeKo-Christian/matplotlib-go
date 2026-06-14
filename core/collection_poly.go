package core

import (
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// PolyCollection draws many polygons with patch semantics.
type PolyCollection struct {
	PatchCollection
	Polygons [][]geom.Pt
}

// Draw renders the polygon collection.
func (c *PolyCollection) Draw(r render.Renderer, ctx *DrawContext) {
	if c == nil {
		return
	}
	c.asPatchCollection().Draw(r, ctx)
}

// Bounds returns the polygon collection's data-space bounds when applicable.
func (c *PolyCollection) Bounds(ctx *DrawContext) geom.Rect {
	if c == nil {
		return geom.Rect{}
	}
	return c.asPatchCollection().Bounds(ctx)
}

func (c *PolyCollection) legendEntry() (legendEntry, bool) {
	if c == nil {
		return legendEntry{}, false
	}
	return c.asPatchCollection().legendEntry()
}

func (c *PolyCollection) asPatchCollection() *PatchCollection {
	if c == nil {
		return nil
	}
	paths := make([]geom.Path, 0, len(c.Polygons)+len(c.Paths))
	for _, poly := range c.Polygons {
		if len(poly) == 0 {
			continue
		}
		paths = append(paths, polygonPath(poly, true))
	}
	paths = append(paths, c.Paths...)
	patches := c.PatchCollection
	patches.Paths = paths
	return &patches
}
