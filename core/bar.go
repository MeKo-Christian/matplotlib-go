package core

import (
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// BarOrientation specifies the direction of bars.
type BarOrientation uint8

const (
	BarVertical   BarOrientation = iota // bars extend upward from baseline
	BarHorizontal                       // bars extend rightward from baseline
)

// Bar2D renders bar charts using filled rectangles.
type Bar2D struct {
	ArtistRasterization
	X           []float64      // x positions (centers of bars for vertical, positions for horizontal)
	Heights     []float64      // heights/lengths of bars (Y values for vertical, X values for horizontal)
	Widths      []float64      // bar widths, if nil uses Width
	Baselines   []float64      // per-bar baseline/left values, if nil uses Baseline
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
		baseline := b.baselineAt(i)

		// Skip zero-sized bars.
		if height == 0 {
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

		// Create rectangle path based on orientation.
		var path geom.Path
		if b.Orientation == BarVertical {
			path = b.createVerticalBarPath(x, height, width, baseline, ctx)
		} else {
			path = b.createHorizontalBarPath(x, height, width, baseline, ctx)
		}

		if len(path.C) == 0 {
			continue // skip invalid bars
		}

		paint := render.Paint{
			Snap: render.SnapAuto,
		}
		if fillColor.A > 0 {
			paint.Fill = fillColor
		}
		if b.EdgeWidth > 0 && edgeColor.A > 0 {
			paint.Stroke = edgeColor
			paint.LineWidth = b.EdgeWidth
			paint.LineJoin = render.JoinMiter
			paint.LineCap = render.CapButt
		}
		if paint.Fill.A > 0 || paint.Stroke.A > 0 {
			r.Path(path, &paint)
		}
	}
}

func (b *Bar2D) createVerticalBarPath(x, height, width, baseline float64, ctx *DrawContext) geom.Path {
	// Calculate rectangle corners in data space
	halfWidth := width / 2
	left := x - halfWidth
	right := x + halfWidth
	bottom := baseline
	top := baseline + height

	// Handle negative heights (bars extending below baseline)
	if height < 0 {
		bottom = baseline + height
		top = baseline
	}

	px0 := ctx.DataToPixel.Apply(geom.Pt{X: left, Y: bottom})
	px1 := ctx.DataToPixel.Apply(geom.Pt{X: right, Y: top})
	rect, ok := rectFromPoints(px0, px1)
	if !ok {
		return geom.Path{}
	}
	return pixelRectPath(rect)
}

func (b *Bar2D) createHorizontalBarPath(y, height, width, baseline float64, ctx *DrawContext) geom.Path {
	// For horizontal bars:
	// y is the y-position (center)
	// height is the length (width) of the bar
	// width is the thickness (height) of the bar
	halfWidth := width / 2
	left := baseline
	right := baseline + height
	bottom := y - halfWidth
	top := y + halfWidth

	// Handle negative heights (bars extending left from baseline)
	if height < 0 {
		left = baseline + height
		right = baseline
	}

	px0 := ctx.DataToPixel.Apply(geom.Pt{X: left, Y: bottom})
	px1 := ctx.DataToPixel.Apply(geom.Pt{X: right, Y: top})
	rect, ok := rectFromPoints(px0, px1)
	if !ok {
		return geom.Path{}
	}
	return pixelRectPath(rect)
}

// Z returns the z-order for sorting.
func (b *Bar2D) Z() float64 {
	return zOrDefault(b.z, defaultPatchZ)
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

// StickyEdges returns the bar baseline edge used by autoscaling. Matplotlib
// bars suppress margins across the baseline so positive bars start at the
// spine while the far data edge still receives the configured margin.
func (b *Bar2D) StickyEdges() ([]float64, []float64) {
	n := len(b.X)
	if len(b.Heights) < n {
		n = len(b.Heights)
	}
	if n <= 0 {
		return nil, nil
	}

	edges := make([]float64, 0, n)
	for i := 0; i < n; i++ {
		edges = append(edges, b.baselineAt(i))
	}
	if b.Orientation == BarHorizontal {
		return edges, nil
	}
	return nil, edges
}

// verticalBounds calculates bounds for vertical bars.
func (b *Bar2D) verticalBounds(numBars int) geom.Rect {
	// Initialize bounds with first bar
	x0 := b.X[0]
	height0 := b.Heights[0]
	baseline0 := b.baselineAt(0)
	left0, right0 := barSpanAroundCenter(x0, barWidthAt(b.Width, b.Widths, 0))
	minX := left0
	maxX := right0
	minY := baseline0
	maxY := baseline0 + height0

	if height0 < 0 {
		minY = baseline0 + height0
		maxY = baseline0
	}

	// Expand bounds to include all bars
	for i := 1; i < numBars; i++ {
		x := b.X[i]
		height := b.Heights[i]
		baseline := b.baselineAt(i)

		// X bounds (bar positions and width)
		left, right := barSpanAroundCenter(x, barWidthAt(b.Width, b.Widths, i))
		if left < minX {
			minX = left
		}
		if right > maxX {
			maxX = right
		}

		// Y bounds (bar heights)
		if height >= 0 {
			bottom := baseline
			top := baseline + height
			if bottom < minY {
				minY = bottom
			}
			if top > maxY {
				maxY = top
			}
		} else {
			bottom := baseline + height
			top := baseline
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
	// Initialize bounds with first bar
	y0 := b.X[0] // In horizontal bars, X represents Y positions
	height0 := b.Heights[0]
	baseline0 := b.baselineAt(0)
	minX := baseline0
	maxX := baseline0 + height0
	bottom0, top0 := barSpanAroundCenter(y0, barWidthAt(b.Width, b.Widths, 0))
	minY := bottom0
	maxY := top0

	if height0 < 0 {
		minX = baseline0 + height0
		maxX = baseline0
	}

	// Expand bounds to include all bars
	for i := 1; i < numBars; i++ {
		y := b.X[i] // In horizontal bars, X represents Y positions
		height := b.Heights[i]
		baseline := b.baselineAt(i)

		// X bounds (bar lengths)
		if height >= 0 {
			left := baseline
			right := baseline + height
			if left < minX {
				minX = left
			}
			if right > maxX {
				maxX = right
			}
		} else {
			left := baseline + height
			right := baseline
			if left < minX {
				minX = left
			}
			if right > maxX {
				maxX = right
			}
		}

		// Y bounds (bar positions and width)
		bottom, top := barSpanAroundCenter(y, barWidthAt(b.Width, b.Widths, i))
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

func (b *Bar2D) baselineAt(i int) float64 {
	if len(b.Baselines) > 0 && i < len(b.Baselines) {
		return b.Baselines[i]
	}
	return b.Baseline
}

func barWidthAt(fallback float64, widths []float64, i int) float64 {
	if len(widths) > 0 && i < len(widths) {
		return widths[i]
	}
	return fallback
}

func barSpanAroundCenter(center, width float64) (float64, float64) {
	half := width / 2
	a := center - half
	b := center + half
	if a <= b {
		return a, b
	}
	return b, a
}
