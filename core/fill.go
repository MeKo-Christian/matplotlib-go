package core

import (
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

// Fill2D creates filled areas between curves or from curves to baselines.
type Fill2D struct {
	X         []float64    // x coordinates
	Y1        []float64    // first y curve (top boundary)
	Y2        []float64    // second y curve (bottom boundary), if nil uses Baseline
	Baseline  float64      // baseline value when Y2 is nil
	Color     render.Color // fill color
	EdgeColor render.Color // edge color for outline (0 alpha means no edge)
	EdgeWidth float64      // edge width in pixels (0 means no edge)
	Alpha     float64      // alpha transparency override (0-1), if 0 uses Color.A
	Label     string       // series label for legend
	z         float64      // z-order
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

	// Get edge color with alpha
	edgeColor := f.EdgeColor
	if f.Alpha > 0 && f.Alpha <= 1 {
		edgeColor.A *= f.Alpha
	}

	// Create paint for fill area
	paint := render.Paint{
		Fill: fillColor,
	}

	// Add stroke if edge width is specified and edge color has alpha > 0
	if f.EdgeWidth > 0 && edgeColor.A > 0 {
		paint.Stroke = edgeColor
		paint.LineWidth = f.EdgeWidth
		paint.LineJoin = render.JoinRound
		paint.LineCap = render.CapRound
	}

	// Draw fill area with optional edge
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

// Bounds returns the bounding box of the fill area.
func (f *Fill2D) Bounds(*DrawContext) geom.Rect {
	if len(f.X) == 0 || len(f.Y1) == 0 {
		return geom.Rect{}
	}

	// Use minimum length across all arrays
	n := len(f.X)
	if len(f.Y1) < n {
		n = len(f.Y1)
	}
	if f.Y2 != nil && len(f.Y2) < n {
		n = len(f.Y2)
	}

	if n == 0 {
		return geom.Rect{}
	}

	// Initialize bounds with first point
	bounds := geom.Rect{
		Min: geom.Pt{X: f.X[0], Y: f.Y1[0]},
		Max: geom.Pt{X: f.X[0], Y: f.Y1[0]},
	}

	// Expand bounds to include all X and Y1 points
	for i := 0; i < n; i++ {
		x := f.X[i]
		y1 := f.Y1[i]

		if x < bounds.Min.X {
			bounds.Min.X = x
		}
		if x > bounds.Max.X {
			bounds.Max.X = x
		}
		if y1 < bounds.Min.Y {
			bounds.Min.Y = y1
		}
		if y1 > bounds.Max.Y {
			bounds.Max.Y = y1
		}
	}

	// Include Y2 values or baseline in bounds
	if f.Y2 != nil {
		// Include all Y2 points
		for i := 0; i < n; i++ {
			y2 := f.Y2[i]
			if y2 < bounds.Min.Y {
				bounds.Min.Y = y2
			}
			if y2 > bounds.Max.Y {
				bounds.Max.Y = y2
			}
		}
	} else {
		// Include baseline
		if f.Baseline < bounds.Min.Y {
			bounds.Min.Y = f.Baseline
		}
		if f.Baseline > bounds.Max.Y {
			bounds.Max.Y = f.Baseline
		}
	}

	return bounds
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
