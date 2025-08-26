package core

import (
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

// AxisSide specifies which side of the plot area an axis is on.
type AxisSide uint8

const (
	AxisBottom AxisSide = iota // x-axis at bottom
	AxisTop                    // x-axis at top
	AxisLeft                   // y-axis at left
	AxisRight                  // y-axis at right
)

// Axis renders axis spines, ticks, and labels for a single dimension.
type Axis struct {
	Side       AxisSide     // which side of the plot
	Locator    Locator      // tick position calculator
	Formatter  Formatter    // tick label formatter
	Color      render.Color // axis line and tick color
	LineWidth  float64      // width of axis line and ticks
	TickSize   float64      // length of tick marks (in pixels)
	ShowSpine  bool         // whether to draw the axis line
	ShowTicks  bool         // whether to draw tick marks
	ShowLabels bool         // whether to draw tick labels (stub for now)
	z          float64      // z-order
}

// NewXAxis creates an axis for the bottom (x-axis).
func NewXAxis() *Axis {
	return &Axis{
		Side:       AxisBottom,
		Locator:    LinearLocator{},
		Formatter:  ScalarFormatter{Prec: 3},
		Color:      render.Color{R: 0, G: 0, B: 0, A: 1}, // black
		LineWidth:  1.0,
		TickSize:   5.0,
		ShowSpine:  true,
		ShowTicks:  true,
		ShowLabels: true,
	}
}

// NewYAxis creates an axis for the left (y-axis).
func NewYAxis() *Axis {
	return &Axis{
		Side:       AxisLeft,
		Locator:    LinearLocator{},
		Formatter:  ScalarFormatter{Prec: 3},
		Color:      render.Color{R: 0, G: 0, B: 0, A: 1}, // black
		LineWidth:  1.0,
		TickSize:   5.0,
		ShowSpine:  true,
		ShowTicks:  true,
		ShowLabels: true,
	}
}

// Draw renders the axis spine and ticks.
func (a *Axis) Draw(r render.Renderer, ctx *DrawContext) {
	// Get the axis domain from the appropriate scale
	var min, max float64
	var isXAxis bool

	switch a.Side {
	case AxisBottom, AxisTop:
		min, max = ctx.DataToPixel.XScale.Domain()
		isXAxis = true
	case AxisLeft, AxisRight:
		min, max = ctx.DataToPixel.YScale.Domain()
		isXAxis = false
	}

	// Calculate tick positions
	ticks := a.Locator.Ticks(min, max, 8) // aim for ~8 ticks

	// Draw spine (axis line)
	if a.ShowSpine {
		a.drawSpine(r, ctx, isXAxis)
	}

	// Draw tick marks
	if a.ShowTicks && len(ticks) > 0 {
		a.drawTicks(r, ctx, ticks, isXAxis)
	}

	// Draw tick labels if supported by the renderer
	if a.ShowLabels && len(ticks) > 0 {
		a.drawTickLabels(r, ctx, ticks, isXAxis)
	}
}

// drawSpine draws the main axis line.
func (a *Axis) drawSpine(r render.Renderer, ctx *DrawContext, isXAxis bool) {
	var p1, p2 geom.Pt

	if isXAxis {
		// Horizontal spine
		min, max := ctx.DataToPixel.XScale.Domain()
		y := getSpinePosition(a.Side, ctx)

		p1 = ctx.DataToPixel.Apply(geom.Pt{X: min, Y: y})
		p2 = ctx.DataToPixel.Apply(geom.Pt{X: max, Y: y})
	} else {
		// Vertical spine
		min, max := ctx.DataToPixel.YScale.Domain()
		x := getSpinePosition(a.Side, ctx)

		p1 = ctx.DataToPixel.Apply(geom.Pt{X: x, Y: min})
		p2 = ctx.DataToPixel.Apply(geom.Pt{X: x, Y: max})
	}

	// Create line path
	path := geom.Path{}
	path.C = append(path.C, geom.MoveTo)
	path.V = append(path.V, p1)
	path.C = append(path.C, geom.LineTo)
	path.V = append(path.V, p2)

	// Draw the spine
	paint := render.Paint{
		LineWidth: a.LineWidth,
		Stroke:    a.Color,
		LineCap:   render.CapButt,
		LineJoin:  render.JoinMiter,
	}
	r.Path(path, &paint)
}

// drawTicks draws tick marks at the specified positions.
func (a *Axis) drawTicks(r render.Renderer, ctx *DrawContext, ticks []float64, isXAxis bool) {
	for _, tickValue := range ticks {
		a.drawSingleTick(r, ctx, tickValue, isXAxis)
	}
}

// drawSingleTick draws a single tick mark.
func (a *Axis) drawSingleTick(r render.Renderer, ctx *DrawContext, tickValue float64, isXAxis bool) {
	var p1, p2 geom.Pt

	if isXAxis {
		// Vertical tick mark
		spineY := getSpinePosition(a.Side, ctx)
		tickX := tickValue

		// Transform spine position to pixel coordinates
		spinePixel := ctx.DataToPixel.Apply(geom.Pt{X: tickX, Y: spineY})

		// Calculate tick endpoints in pixel space
		// Note: With Y-flipped coordinates, positive Y is up in data space but down in pixel space
		switch a.Side {
		case AxisBottom:
			p1 = spinePixel
			p2 = geom.Pt{X: spinePixel.X, Y: spinePixel.Y - a.TickSize} // Ticks point down (more positive Y in screen coords)
		case AxisTop:
			p1 = spinePixel
			p2 = geom.Pt{X: spinePixel.X, Y: spinePixel.Y + a.TickSize} // Ticks point up (less Y in screen coords)
		}
	} else {
		// Horizontal tick mark
		spineX := getSpinePosition(a.Side, ctx)
		tickY := tickValue

		// Transform spine position to pixel coordinates
		spinePixel := ctx.DataToPixel.Apply(geom.Pt{X: spineX, Y: tickY})

		// Calculate tick endpoints in pixel space
		switch a.Side {
		case AxisLeft:
			p1 = spinePixel
			p2 = geom.Pt{X: spinePixel.X - a.TickSize, Y: spinePixel.Y}
		case AxisRight:
			p1 = spinePixel
			p2 = geom.Pt{X: spinePixel.X + a.TickSize, Y: spinePixel.Y}
		}
	}

	// Create tick path
	path := geom.Path{}
	path.C = append(path.C, geom.MoveTo)
	path.V = append(path.V, p1)
	path.C = append(path.C, geom.LineTo)
	path.V = append(path.V, p2)

	// Draw the tick
	paint := render.Paint{
		LineWidth: a.LineWidth,
		Stroke:    a.Color,
		LineCap:   render.CapButt,
		LineJoin:  render.JoinMiter,
	}
	r.Path(path, &paint)
}

// getSpinePosition returns the data coordinate where the spine should be drawn.
func getSpinePosition(side AxisSide, ctx *DrawContext) float64 {
	switch side {
	case AxisBottom, AxisTop:
		// For x-axis, spine is at y coordinate
		yMin, yMax := ctx.DataToPixel.YScale.Domain()
		if side == AxisBottom {
			return yMin // bottom of plot
		}
		return yMax // top of plot
	case AxisLeft, AxisRight:
		// For y-axis, spine is at x coordinate
		xMin, xMax := ctx.DataToPixel.XScale.Domain()
		if side == AxisLeft {
			return xMin // left of plot
		}
		return xMax // right of plot
	}
	return 0
}

// Z returns the z-order for sorting.
func (a *Axis) Z() float64 {
	return a.z
}

// Bounds returns an empty rect for now.
func (a *Axis) Bounds(*DrawContext) geom.Rect {
	return geom.Rect{}
}

// drawTickLabels draws text labels for the ticks if the renderer supports text.
func (a *Axis) drawTickLabels(r render.Renderer, ctx *DrawContext, ticks []float64, isXAxis bool) {
	// Check if renderer supports text drawing (gobasic.Renderer has DrawText method)
	type textRenderer interface {
		DrawText(text string, origin geom.Pt, size float64, textColor render.Color)
	}
	
	textRen, ok := r.(textRenderer)
	if !ok {
		return // Renderer doesn't support text
	}
	
	fontSize := 12.0 // Default font size
	
	for _, tickValue := range ticks {
		// Format the tick value using the formatter
		label := a.Formatter.Format(tickValue)
		if label == "" {
			continue
		}
		
		// Calculate label position
		var labelPos geom.Pt
		
		if isXAxis {
			// X-axis labels go below the ticks
			spineY := getSpinePosition(a.Side, ctx)
			tickPos := ctx.DataToPixel.Apply(geom.Pt{X: tickValue, Y: spineY})
			
			switch a.Side {
			case AxisBottom:
				labelPos = geom.Pt{X: tickPos.X, Y: tickPos.Y - a.TickSize - 5} // Below tick
			case AxisTop:
				labelPos = geom.Pt{X: tickPos.X, Y: tickPos.Y + a.TickSize + fontSize + 5} // Above tick
			}
		} else {
			// Y-axis labels go to the left of the ticks
			spineX := getSpinePosition(a.Side, ctx)
			tickPos := ctx.DataToPixel.Apply(geom.Pt{X: spineX, Y: tickValue})
			
			switch a.Side {
			case AxisLeft:
				labelPos = geom.Pt{X: tickPos.X - a.TickSize - 50, Y: tickPos.Y + fontSize/2} // Left of tick
			case AxisRight:
				labelPos = geom.Pt{X: tickPos.X + a.TickSize + 5, Y: tickPos.Y + fontSize/2} // Right of tick
			}
		}
		
		// Draw the label
		textRen.DrawText(label, labelPos, fontSize, a.Color)
	}
}
