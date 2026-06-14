package core

import (
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// FillBetweenPolyCollection is a polygon-collection primitive specialized for
// fill_between-style regions.
type FillBetweenPolyCollection struct {
	PatchCollection
	X           []float64
	Y1          []float64
	Y2          []float64
	Baseline    float64
	Orientation FillOrientation
}

// Draw renders the fill-between poly collection.
func (c *FillBetweenPolyCollection) Draw(r render.Renderer, ctx *DrawContext) {
	if c == nil {
		return
	}
	c.asPatchCollection().Draw(r, ctx)
}

// Bounds returns the fill-between collection's data-space bounds when applicable.
func (c *FillBetweenPolyCollection) Bounds(ctx *DrawContext) geom.Rect {
	if c == nil {
		return geom.Rect{}
	}
	return c.asPatchCollection().Bounds(ctx)
}

func (c *FillBetweenPolyCollection) legendEntry() (legendEntry, bool) {
	if c == nil {
		return legendEntry{}, false
	}
	return c.asPatchCollection().legendEntry()
}

func (c *FillBetweenPolyCollection) asPatchCollection() *PatchCollection {
	if c == nil {
		return nil
	}
	paths := []geom.Path{}
	if poly := fillBetweenPolygon(c.X, c.Y1, c.Y2, c.Baseline, c.Orientation); len(poly) > 0 {
		paths = append(paths, polygonPath(poly, true))
	}
	patches := c.PatchCollection
	patches.Paths = paths
	return &patches
}

func fillBetweenPolygon(x, y1, y2 []float64, baseline float64, orientation FillOrientation) []geom.Pt {
	if len(x) == 0 || len(y1) == 0 {
		return nil
	}
	n := len(x)
	if len(y1) < n {
		n = len(y1)
	}
	if len(y2) > 0 && len(y2) < n {
		n = len(y2)
	}
	if n < 2 {
		return nil
	}

	poly := make([]geom.Pt, 0, 2*n)
	for i := 0; i < n; i++ {
		poly = append(poly, fillBetweenPoint(orientation, x[i], y1[i]))
	}
	for i := n - 1; i >= 0; i-- {
		dep := baseline
		if len(y2) > 0 {
			dep = y2[i]
		}
		poly = append(poly, fillBetweenPoint(orientation, x[i], dep))
	}
	return poly
}

func fillBetweenPoint(orientation FillOrientation, primary, dependent float64) geom.Pt {
	if orientation == FillHorizontal {
		return geom.Pt{X: dependent, Y: primary}
	}
	return geom.Pt{X: primary, Y: dependent}
}
