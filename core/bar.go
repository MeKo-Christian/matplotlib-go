package core

import (
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

// BarOrientation specifies the direction of bars.
type BarOrientation uint8

const (
	BarVertical   BarOrientation = iota // bars extend upward from baseline
	BarHorizontal                       // bars extend rightward from baseline
)

// Bar2D renders bar charts using filled rectangles.
type Bar2D struct {
	X           []float64      // x positions (centers of bars for vertical, left edges for horizontal)
	Y           []float64      // y values (heights for vertical, x-extent for horizontal)
	Width       []float64      // bar widths, if nil uses Width
	Colors      []render.Color // bar colors, if nil uses Color
	BarWidth    float64        // default bar width
	Color       render.Color   // default bar color
	Baseline    float64        // baseline value (0 for most cases)
	Orientation BarOrientation // vertical or horizontal bars
	z           float64        // z-order
}

// Draw renders bars by creating filled rectangles for each bar.
func (b *Bar2D) Draw(r render.Renderer, ctx *DrawContext) {
	if len(b.X) == 0 || len(b.Y) == 0 {
		return // nothing to draw
	}

	// Use minimum of X and Y lengths
	n := len(b.X)
	if len(b.Y) < n {
		n = len(b.Y)
	}

	for i := 0; i < n; i++ {
		// Get width for this bar
		width := b.BarWidth
		if b.Width != nil && i < len(b.Width) {
			width = b.Width[i]
		}

		// Get color for this bar
		color := b.Color
		if b.Colors != nil && i < len(b.Colors) {
			color = b.Colors[i]
		}

		// Create rectangle path based on orientation
		var rectPath geom.Path
		if b.Orientation == BarVertical {
			rectPath = b.createVerticalBarPath(b.X[i], b.Y[i], width, ctx)
		} else {
			rectPath = b.createHorizontalBarPath(b.X[i], b.Y[i], width, ctx)
		}

		if len(rectPath.C) == 0 {
			continue // skip invalid bars
		}

		// Draw filled rectangle
		paint := render.Paint{
			Fill: color,
		}
		r.Path(rectPath, &paint)
	}
}

// createVerticalBarPath creates a rectangle for a vertical bar.
func (b *Bar2D) createVerticalBarPath(x, height, width float64, ctx *DrawContext) geom.Path {
	// Calculate rectangle corners in data space
	left := x - width/2
	right := x + width/2
	bottom := b.Baseline
	top := height

	// Ensure correct order (bottom <= top)
	if top < bottom {
		bottom, top = top, bottom
	}

	// Transform to pixel coordinates
	bl := ctx.DataToPixel.Apply(geom.Pt{X: left, Y: bottom})  // bottom-left
	br := ctx.DataToPixel.Apply(geom.Pt{X: right, Y: bottom}) // bottom-right
	tr := ctx.DataToPixel.Apply(geom.Pt{X: right, Y: top})    // top-right
	tl := ctx.DataToPixel.Apply(geom.Pt{X: left, Y: top})     // top-left

	path := geom.Path{}
	path.C = append(path.C, geom.MoveTo)
	path.V = append(path.V, bl)
	path.C = append(path.C, geom.LineTo)
	path.V = append(path.V, br)
	path.C = append(path.C, geom.LineTo)
	path.V = append(path.V, tr)
	path.C = append(path.C, geom.LineTo)
	path.V = append(path.V, tl)
	path.C = append(path.C, geom.ClosePath)

	return path
}

// createHorizontalBarPath creates a rectangle for a horizontal bar.
func (b *Bar2D) createHorizontalBarPath(y, width, height float64, ctx *DrawContext) geom.Path {
	// Calculate rectangle corners in data space
	left := b.Baseline
	right := width
	bottom := y - height/2
	top := y + height/2

	// Ensure correct order (left <= right)
	if right < left {
		left, right = right, left
	}

	// Transform to pixel coordinates
	bl := ctx.DataToPixel.Apply(geom.Pt{X: left, Y: bottom})  // bottom-left
	br := ctx.DataToPixel.Apply(geom.Pt{X: right, Y: bottom}) // bottom-right
	tr := ctx.DataToPixel.Apply(geom.Pt{X: right, Y: top})    // top-right
	tl := ctx.DataToPixel.Apply(geom.Pt{X: left, Y: top})     // top-left

	path := geom.Path{}
	path.C = append(path.C, geom.MoveTo)
	path.V = append(path.V, bl)
	path.C = append(path.C, geom.LineTo)
	path.V = append(path.V, br)
	path.C = append(path.C, geom.LineTo)
	path.V = append(path.V, tr)
	path.C = append(path.C, geom.LineTo)
	path.V = append(path.V, tl)
	path.C = append(path.C, geom.ClosePath)

	return path
}

// Z returns the z-order for sorting.
func (b *Bar2D) Z() float64 {
	return b.z
}

// Bounds returns an empty rect for now (will be enhanced in later phases).
func (b *Bar2D) Bounds(*DrawContext) geom.Rect {
	return geom.Rect{}
}
