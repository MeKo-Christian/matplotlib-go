package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// drawSpine draws the main axis line directly in pixel space, centered on the
// axes edge so the stroke can extend on both sides like Matplotlib's spines.
func (a *Axis) drawSpine(r render.Renderer, ctx *DrawContext) {
	px := ctx.Clip
	lw := pointsToPixels(ctx.RC, a.LineWidth)
	p1, p2 := axisSpinePixelEndpoints(a, ctx, px)

	paint := render.Paint{
		LineWidth: lw,
		Stroke:    a.Color,
		LineCap:   a.LineCap,
		LineJoin:  a.LineJoin,
		Dashes:    styleCloneDashes(a.Dashes),
	}
	path := geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo},
		V: []geom.Pt{p1, p2},
	}
	r.Path(path, &paint)
}

// spinePixelEndpoints returns the two pixel-space endpoints for a spine on the
// given side of px. Matplotlib's AGG backend appends the y-flip to the path
// transform before PathSnapper runs, so X can be snapped in display space but Y
// must emulate device-space snapping and then convert back to y-up display
// coordinates.
func spinePixelEndpoints(side AxisSide, px geom.Rect, contexts ...*DrawContext) (geom.Pt, geom.Pt) {
	var ctx *DrawContext
	if len(contexts) > 0 {
		ctx = contexts[0]
	}
	x1 := snapDisplayX(px.Min.X)
	y1 := snapDisplayY(px.Min.Y, figureSnapHeight(ctx))
	x2 := snapDisplayX(px.Max.X)
	y2 := snapDisplayY(px.Max.Y, figureSnapHeight(ctx))

	switch side {
	case AxisBottom:
		return geom.Pt{X: x1, Y: y1}, geom.Pt{X: x2, Y: y1}
	case AxisTop:
		return geom.Pt{X: x1, Y: y2}, geom.Pt{X: x2, Y: y2}
	case AxisLeft:
		return geom.Pt{X: x1, Y: y1}, geom.Pt{X: x1, Y: y2}
	case AxisRight:
		return geom.Pt{X: x2, Y: y1}, geom.Pt{X: x2, Y: y2}
	}
	return geom.Pt{}, geom.Pt{}
}

func snapDisplayX(x float64) float64 {
	return math.Floor(x+0.5) + 0.5
}

func snapDisplayY(y float64, heights ...float64) float64 {
	if len(heights) > 0 && heights[0] > 0 {
		height := heights[0]
		deviceY := height - y
		return height - (math.Floor(deviceY+0.5) + 0.5)
	}
	return math.Floor(y+0.5) + 0.5
}

// snapOffsetForWidth mirrors the AGG backend's PathSnapper parity rule: a stroke
// whose rounded pixel width is odd is centred on a half-pixel (+0.5), while an
// even width is centred on an integer pixel boundary (+0.0). Matplotlib applies
// this same offset uniformly to every coordinate of a snapped axis-aligned path.
func snapOffsetForWidth(lineWidthPx float64) float64 {
	if lineWidthPx > 0 && int(math.Round(lineWidthPx))%2 != 0 {
		return 0.5
	}
	return 0.0
}

// snapDisplayXForWidth snaps an x display coordinate with the width-parity offset
// so even-width tick marks land on the pixel boundary exactly like Matplotlib.
func snapDisplayXForWidth(x, lineWidthPx float64) float64 {
	return math.Floor(x+0.5) + snapOffsetForWidth(lineWidthPx)
}

// snapDisplayYForWidth is the y-up analogue of snapDisplayXForWidth, emulating
// device-space snapping ahead of the backend y-flip.
func snapDisplayYForWidth(y, lineWidthPx float64, heights ...float64) float64 {
	offset := snapOffsetForWidth(lineWidthPx)
	if len(heights) > 0 && heights[0] > 0 {
		height := heights[0]
		deviceY := height - y
		return height - (math.Floor(deviceY+0.5) + offset)
	}
	return math.Floor(y+0.5) + offset
}

func figureSnapHeight(ctx *DrawContext) float64 {
	if ctx == nil {
		return 0
	}
	if ctx.FigureRect.H() > 0 {
		return ctx.FigureRect.Max.Y
	}
	return 0
}

// DrawFrame draws fallback top/right border lines when those sides are not
// represented by explicit axes.
func DrawFrame(r render.Renderer, ctx *DrawContext, ref *Axis, drawTop, drawRight bool) {
	if ref == nil || !ref.ShowSpine {
		return
	}
	if !drawTop && !drawRight {
		return
	}
	paint := render.Paint{
		LineWidth: pointsToPixels(ctx.RC, ref.LineWidth),
		Stroke:    ref.Color,
		LineCap:   ref.LineCap,
		LineJoin:  ref.LineJoin,
		Dashes:    styleCloneDashes(ref.Dashes),
	}
	drawLine := func(p1, p2 geom.Pt) {
		path := geom.Path{
			C: []geom.Cmd{geom.MoveTo, geom.LineTo},
			V: []geom.Pt{p1, p2},
		}
		r.Path(path, &paint)
	}

	if drawTop {
		p1, p2 := spinePixelEndpoints(AxisTop, ctx.Clip, ctx)
		drawLine(p1, p2)
	}
	if drawRight {
		p1, p2 := spinePixelEndpoints(AxisRight, ctx.Clip, ctx)
		drawLine(p1, p2)
	}
}

// getSpinePosition returns the data coordinate where the spine should be drawn.
func getSpinePosition(axis *Axis, ctx *DrawContext) float64 {
	if axis == nil || ctx == nil {
		return 0
	}
	if axis.SpinePositionMode == AxisSpinePositionData {
		return axis.SpinePosition
	}
	switch axis.Side {
	case AxisBottom, AxisTop:
		// For x-axis, spine is at y coordinate
		yMin, yMax := ctx.DataToPixel.YScale.Domain()
		if axis.Side == AxisBottom {
			return yMin // bottom of plot
		}
		return yMax // top of plot
	case AxisLeft, AxisRight:
		// For y-axis, spine is at x coordinate
		xMin, xMax := ctx.DataToPixel.XScale.Domain()
		if axis.Side == AxisLeft {
			return xMin // left of plot
		}
		return xMax // right of plot
	}
	return 0
}

func axisSpinePixelEndpoints(axis *Axis, ctx *DrawContext, px geom.Rect) (geom.Pt, geom.Pt) {
	if axis == nil {
		return geom.Pt{}, geom.Pt{}
	}
	if ctx == nil || axis.SpinePositionMode == AxisSpinePositionBoundary {
		return spinePixelEndpoints(axis.Side, px, ctx)
	}

	switch axis.Side {
	case AxisBottom, AxisTop:
		xMin, xMax := ctx.DataToPixel.XScale.Domain()
		y := getSpinePosition(axis, ctx)
		p1 := ctx.DataToPixel.Apply(geom.Pt{X: xMin, Y: y})
		p2 := ctx.DataToPixel.Apply(geom.Pt{X: xMax, Y: y})
		p1.X = math.Round(p1.X) + 0.5
		p2.X = math.Round(p2.X) + 0.5
		p1.Y = snapDisplayY(p1.Y, figureSnapHeight(ctx))
		p2.Y = snapDisplayY(p2.Y, figureSnapHeight(ctx))
		return p1, p2
	case AxisLeft, AxisRight:
		yMin, yMax := ctx.DataToPixel.YScale.Domain()
		x := getSpinePosition(axis, ctx)
		p1 := ctx.DataToPixel.Apply(geom.Pt{X: x, Y: yMin})
		p2 := ctx.DataToPixel.Apply(geom.Pt{X: x, Y: yMax})
		p1.X = math.Round(p1.X) + 0.5
		p2.X = math.Round(p2.X) + 0.5
		p1.Y = snapDisplayY(p1.Y, figureSnapHeight(ctx))
		p2.Y = snapDisplayY(p2.Y, figureSnapHeight(ctx))
		return p1, p2
	default:
		return geom.Pt{}, geom.Pt{}
	}
}

// SetLineStyle configures cap/join/dash styling for the axis spine and ticks.
func (a *Axis) SetLineStyle(lineCap render.LineCap, join render.LineJoin, dashes ...float64) {
	if a == nil {
		return
	}
	a.LineCap = lineCap
	a.LineJoin = join
	a.TickLineCap = lineCap
	a.TickLineJoin = join
	a.Dashes = styleCloneDashes(dashes)
}

// SetSpinePositionData places the axis spine at a data coordinate instead of the plot boundary.
func (a *Axis) SetSpinePositionData(value float64) {
	if a == nil {
		return
	}
	a.SpinePositionMode = AxisSpinePositionData
	a.SpinePosition = value
}

// ResetSpinePosition restores the axis spine to its default plot boundary.
func (a *Axis) ResetSpinePosition() {
	if a == nil {
		return
	}
	a.SpinePositionMode = AxisSpinePositionBoundary
	a.SpinePosition = 0
}
