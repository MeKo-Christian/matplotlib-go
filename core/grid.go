package core

import (
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

// Grid renders grid lines at tick positions.
type Grid struct {
	Axis      AxisSide     // which axis to use for tick positions
	Color     render.Color // grid line color
	LineWidth float64      // width of grid lines
	Alpha     float64      // alpha override (0-1), if 0 uses Color.A
	Major     bool         // draw grid at major ticks
	Minor     bool         // draw grid at minor ticks (not implemented yet)
	z         float64      // z-order (should be behind data)
}

// NewGrid creates a new grid for the specified axis.
func NewGrid(axis AxisSide) *Grid {
	return &Grid{
		Axis:      axis,
		Color:     render.Color{R: 0.8, G: 0.8, B: 0.8, A: 1}, // light gray
		LineWidth: 0.5,
		Alpha:     0, // use Color.A
		Major:     true,
		Minor:     false,
		z:         -1000, // behind everything else
	}
}

// Draw renders grid lines at tick positions.
func (g *Grid) Draw(r render.Renderer, ctx *DrawContext) {
	if !g.Major {
		return // nothing to draw
	}

	// Get the axis domain and locator
	var min, max float64
	var locator Locator
	var isXAxis bool

	switch g.Axis {
	case AxisBottom, AxisTop:
		min, max = ctx.DataToPixel.XScale.Domain()
		isXAxis = true
		// Use a default locator since we don't have direct access to axis
		locator = LinearLocator{}
	case AxisLeft, AxisRight:
		min, max = ctx.DataToPixel.YScale.Domain()
		isXAxis = false
		locator = LinearLocator{}
	}

	// Calculate tick positions
	ticks := locator.Ticks(min, max, 8)

	// Get grid color
	gridColor := g.Color
	if g.Alpha > 0 && g.Alpha <= 1 {
		gridColor.A = g.Alpha
	}

	// Draw grid lines
	for _, tickValue := range ticks {
		g.drawGridLine(r, ctx, tickValue, isXAxis, gridColor)
	}
}

// drawGridLine draws a single grid line.
func (g *Grid) drawGridLine(r render.Renderer, ctx *DrawContext, tickValue float64, isXAxis bool, color render.Color) {
	var p1, p2 geom.Pt

	if isXAxis {
		// Vertical grid line (for x-axis ticks)
		yMin, yMax := ctx.DataToPixel.YScale.Domain()
		p1 = ctx.DataToPixel.Apply(geom.Pt{X: tickValue, Y: yMin})
		p2 = ctx.DataToPixel.Apply(geom.Pt{X: tickValue, Y: yMax})
	} else {
		// Horizontal grid line (for y-axis ticks)
		xMin, xMax := ctx.DataToPixel.XScale.Domain()
		p1 = ctx.DataToPixel.Apply(geom.Pt{X: xMin, Y: tickValue})
		p2 = ctx.DataToPixel.Apply(geom.Pt{X: xMax, Y: tickValue})
	}

	// Create line path
	path := geom.Path{}
	path.C = append(path.C, geom.MoveTo)
	path.V = append(path.V, p1)
	path.C = append(path.C, geom.LineTo)
	path.V = append(path.V, p2)

	// Draw the grid line
	paint := render.Paint{
		LineWidth: g.LineWidth,
		Stroke:    color,
		LineCap:   render.CapButt,
		LineJoin:  render.JoinMiter,
	}
	r.Path(path, &paint)
}

// Z returns the z-order for sorting.
func (g *Grid) Z() float64 {
	return g.z
}

// Bounds returns an empty rect for now.
func (g *Grid) Bounds(*DrawContext) geom.Rect {
	return geom.Rect{}
}
