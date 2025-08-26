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
	X           []float64      // x positions (centers of bars for vertical, positions for horizontal)
	Heights     []float64      // heights/lengths of bars (Y values for vertical, X values for horizontal)
	Widths      []float64      // bar widths, if nil uses Width
	Colors      []render.Color // bar fill colors, if nil uses Color
	EdgeColors  []render.Color // edge colors for bar outlines, if nil uses EdgeColor
	Width       float64        // default bar width in data units
	Color       render.Color   // default bar fill color
	EdgeColor   render.Color   // default edge color for bar outlines
	EdgeWidth   float64        // edge width in pixels (0 means no edge)
	Alpha       float64        // alpha transparency (0-1), applied to both fill and edge
	Baseline    float64        // baseline value (0 for most cases)
	Orientation BarOrientation // vertical or horizontal bars
	Label       string         // series label for legend
	z           float64        // z-order
}

// Draw renders bars by creating filled rectangles for each bar.
func (b *Bar2D) Draw(r render.Renderer, ctx *DrawContext) {
	if len(b.X) == 0 || len(b.Heights) == 0 {
		return // nothing to draw
	}

	// Determine the number of bars to draw
	numBars := len(b.X)
	if len(b.Heights) < numBars {
		numBars = len(b.Heights)
	}

	for i := 0; i < numBars; i++ {
		x := b.X[i]
		height := b.Heights[i]

		// Skip bars with zero or negative height
		if height <= 0 {
			continue
		}

		// Get width for this bar
		width := b.Width
		if b.Widths != nil && i < len(b.Widths) {
			width = b.Widths[i]
		}

		// Get fill color for this bar
		fillColor := b.Color
		if b.Colors != nil && i < len(b.Colors) {
			fillColor = b.Colors[i]
		}

		// Get edge color for this bar
		edgeColor := b.EdgeColor
		if b.EdgeColors != nil && i < len(b.EdgeColors) {
			edgeColor = b.EdgeColors[i]
		}

		// Apply alpha transparency
		alpha := b.Alpha
		if alpha <= 0 {
			alpha = 1.0 // default to fully opaque
		}
		if alpha > 1 {
			alpha = 1.0 // clamp to maximum opacity
		}

		// Apply alpha to colors
		fillColor.A *= alpha
		edgeColor.A *= alpha

		// Create rectangle path based on orientation
		var rectPath geom.Path
		if b.Orientation == BarVertical {
			rectPath = b.createVerticalBarPath(x, height, width, ctx)
		} else {
			rectPath = b.createHorizontalBarPath(x, height, width, ctx)
		}

		if len(rectPath.C) == 0 {
			continue // skip invalid bars
		}

		// Create paint for bar
		paint := render.Paint{
			Fill: fillColor,
		}

		// Add stroke if edge width is specified
		if b.EdgeWidth > 0 && edgeColor.A > 0 {
			paint.Stroke = edgeColor
			paint.LineWidth = b.EdgeWidth
			paint.LineJoin = render.JoinMiter
			paint.LineCap = render.CapSquare
		}

		// Draw bar
		r.Path(rectPath, &paint)
	}
}

// createVerticalBarPath creates a rectangle for a vertical bar.
func (b *Bar2D) createVerticalBarPath(x, height, width float64, ctx *DrawContext) geom.Path {
	path := geom.Path{}

	// Calculate rectangle corners in data space
	halfWidth := width / 2
	left := x - halfWidth
	right := x + halfWidth
	bottom := b.Baseline
	top := b.Baseline + height

	// Handle negative heights (bars extending below baseline)
	if height < 0 {
		bottom = b.Baseline + height
		top = b.Baseline
	}

	// Define rectangle corners
	corners := []geom.Pt{
		{X: left, Y: bottom},  // bottom-left
		{X: right, Y: bottom}, // bottom-right
		{X: right, Y: top},    // top-right
		{X: left, Y: top},     // top-left
	}

	// Transform to pixel coordinates and create path
	for i, corner := range corners {
		pixelPt := ctx.DataToPixel.Apply(corner)
		if i == 0 {
			path.C = append(path.C, geom.MoveTo)
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, pixelPt)
	}
	path.C = append(path.C, geom.ClosePath)

	return path
}

// createHorizontalBarPath creates a rectangle for a horizontal bar.
func (b *Bar2D) createHorizontalBarPath(y, height, width float64, ctx *DrawContext) geom.Path {
	path := geom.Path{}

	// For horizontal bars:
	// y is the y-position (center)
	// height is the length (width) of the bar
	// width is the thickness (height) of the bar
	halfWidth := width / 2
	left := b.Baseline
	right := b.Baseline + height
	bottom := y - halfWidth
	top := y + halfWidth

	// Handle negative heights (bars extending left from baseline)
	if height < 0 {
		left = b.Baseline + height
		right = b.Baseline
	}

	// Define rectangle corners
	corners := []geom.Pt{
		{X: left, Y: bottom},  // bottom-left
		{X: right, Y: bottom}, // bottom-right
		{X: right, Y: top},    // top-right
		{X: left, Y: top},     // top-left
	}

	// Transform to pixel coordinates and create path
	for i, corner := range corners {
		pixelPt := ctx.DataToPixel.Apply(corner)
		if i == 0 {
			path.C = append(path.C, geom.MoveTo)
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, pixelPt)
	}
	path.C = append(path.C, geom.ClosePath)

	return path
}

// Z returns the z-order for sorting.
func (b *Bar2D) Z() float64 {
	return b.z
}

// Bounds returns the bounding box of all bars.
func (b *Bar2D) Bounds(*DrawContext) geom.Rect {
	if len(b.X) == 0 || len(b.Heights) == 0 {
		return geom.Rect{}
	}

	// Determine the number of bars
	numBars := len(b.X)
	if len(b.Heights) < numBars {
		numBars = len(b.Heights)
	}

	if numBars == 0 {
		return geom.Rect{}
	}

	// Calculate bounds based on orientation
	if b.Orientation == BarVertical {
		return b.verticalBounds(numBars)
	} else {
		return b.horizontalBounds(numBars)
	}
}

// verticalBounds calculates bounds for vertical bars.
func (b *Bar2D) verticalBounds(numBars int) geom.Rect {
	// Get maximum width for bounds calculation
	maxWidth := b.Width
	if b.Widths != nil {
		for _, width := range b.Widths {
			if width > maxWidth {
				maxWidth = width
			}
		}
	}
	halfMaxWidth := maxWidth / 2

	// Initialize bounds with first bar
	x0 := b.X[0]
	height0 := b.Heights[0]
	minX := x0 - halfMaxWidth
	maxX := x0 + halfMaxWidth
	minY := b.Baseline
	maxY := b.Baseline + height0

	if height0 < 0 {
		minY = b.Baseline + height0
		maxY = b.Baseline
	}

	// Expand bounds to include all bars
	for i := 1; i < numBars; i++ {
		x := b.X[i]
		height := b.Heights[i]

		// X bounds (bar positions and width)
		left := x - halfMaxWidth
		right := x + halfMaxWidth
		if left < minX {
			minX = left
		}
		if right > maxX {
			maxX = right
		}

		// Y bounds (bar heights)
		if height >= 0 {
			bottom := b.Baseline
			top := b.Baseline + height
			if bottom < minY {
				minY = bottom
			}
			if top > maxY {
				maxY = top
			}
		} else {
			bottom := b.Baseline + height
			top := b.Baseline
			if bottom < minY {
				minY = bottom
			}
			if top > maxY {
				maxY = top
			}
		}
	}

	return geom.Rect{
		Min: geom.Pt{X: minX, Y: minY},
		Max: geom.Pt{X: maxX, Y: maxY},
	}
}

// horizontalBounds calculates bounds for horizontal bars.
func (b *Bar2D) horizontalBounds(numBars int) geom.Rect {
	// Get maximum width for bounds calculation
	maxWidth := b.Width
	if b.Widths != nil {
		for _, width := range b.Widths {
			if width > maxWidth {
				maxWidth = width
			}
		}
	}
	halfMaxWidth := maxWidth / 2

	// Initialize bounds with first bar
	y0 := b.X[0] // In horizontal bars, X represents Y positions
	height0 := b.Heights[0]
	minX := b.Baseline
	maxX := b.Baseline + height0
	minY := y0 - halfMaxWidth
	maxY := y0 + halfMaxWidth

	if height0 < 0 {
		minX = b.Baseline + height0
		maxX = b.Baseline
	}

	// Expand bounds to include all bars
	for i := 1; i < numBars; i++ {
		y := b.X[i] // In horizontal bars, X represents Y positions
		height := b.Heights[i]

		// X bounds (bar lengths)
		if height >= 0 {
			left := b.Baseline
			right := b.Baseline + height
			if left < minX {
				minX = left
			}
			if right > maxX {
				maxX = right
			}
		} else {
			left := b.Baseline + height
			right := b.Baseline
			if left < minX {
				minX = left
			}
			if right > maxX {
				maxX = right
			}
		}

		// Y bounds (bar positions and width)
		bottom := y - halfMaxWidth
		top := y + halfMaxWidth
		if bottom < minY {
			minY = bottom
		}
		if top > maxY {
			maxY = top
		}
	}

	return geom.Rect{
		Min: geom.Pt{X: minX, Y: minY},
		Max: geom.Pt{X: maxX, Y: maxY},
	}
}
