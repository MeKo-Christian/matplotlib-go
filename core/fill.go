package core

import (
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

// Fill2D creates filled areas between curves or from curves to baselines.
type Fill2D struct {
	X        []float64    // x coordinates
	Y1       []float64    // first y curve (top boundary)
	Y2       []float64    // second y curve (bottom boundary), if nil uses Baseline
	Baseline float64      // baseline value when Y2 is nil
	Color    render.Color // fill color
	Alpha    float64      // alpha transparency override (0-1), if 0 uses Color.A
	z        float64      // z-order
}

// Draw renders the filled area by creating a closed path.
func (f *Fill2D) Draw(r render.Renderer, ctx *DrawContext) {
	if len(f.X) == 0 || len(f.Y1) == 0 {
		return // nothing to draw
	}

	// Use minimum length across all arrays
	n := len(f.X)
	if len(f.Y1) < n {
		n = len(f.Y1)
	}
	if f.Y2 != nil && len(f.Y2) < n {
		n = len(f.Y2)
	}

	if n < 2 {
		return // need at least 2 points for area
	}

	// Create the fill path
	fillPath := f.createFillPath(n, ctx)
	if len(fillPath.C) == 0 {
		return // invalid path
	}

	// Get fill color with alpha
	fillColor := f.Color
	if f.Alpha > 0 && f.Alpha <= 1 {
		fillColor.A = f.Alpha
	}

	// Draw filled area
	paint := render.Paint{
		Fill: fillColor,
	}
	r.Path(fillPath, &paint)
}

// createFillPath creates a closed path for the fill area.
func (f *Fill2D) createFillPath(n int, ctx *DrawContext) geom.Path {
	path := geom.Path{}

	// Draw the top boundary (Y1) from left to right
	for i := 0; i < n; i++ {
		pt := ctx.DataToPixel.Apply(geom.Pt{X: f.X[i], Y: f.Y1[i]})
		if i == 0 {
			path.C = append(path.C, geom.MoveTo)
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, pt)
	}

	// Draw the bottom boundary from right to left
	for i := n - 1; i >= 0; i-- {
		var y float64
		if f.Y2 != nil {
			y = f.Y2[i]
		} else {
			y = f.Baseline
		}

		pt := ctx.DataToPixel.Apply(geom.Pt{X: f.X[i], Y: y})
		path.C = append(path.C, geom.LineTo)
		path.V = append(path.V, pt)
	}

	// Close the path
	path.C = append(path.C, geom.ClosePath)

	return path
}

// Z returns the z-order for sorting.
func (f *Fill2D) Z() float64 {
	return f.z
}

// Bounds returns an empty rect for now (will be enhanced in later phases).
func (f *Fill2D) Bounds(*DrawContext) geom.Rect {
	return geom.Rect{}
}

// FillBetween creates a Fill2D for the area between two curves.
func FillBetween(x, y1, y2 []float64, color render.Color) *Fill2D {
	return &Fill2D{
		X:     x,
		Y1:    y1,
		Y2:    y2,
		Color: color,
	}
}

// FillToBaseline creates a Fill2D for the area from a curve to a baseline.
func FillToBaseline(x, y []float64, baseline float64, color render.Color) *Fill2D {
	return &Fill2D{
		X:        x,
		Y1:       y,
		Y2:       nil,
		Baseline: baseline,
		Color:    color,
	}
}
