package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// ErrorBar renders symmetric horizontal and/or vertical error bars for points.
type ErrorBar struct {
	XY              []geom.Pt    // data-space points
	XErr            []float64    // symmetric x errors (same length as XY or broadcast scalar)
	YErr            []float64    // symmetric y errors (same length as XY or broadcast scalar)
	XErrLower       []float64    // asymmetric lower x errors
	XErrUpper       []float64    // asymmetric upper x errors
	YErrLower       []float64    // asymmetric lower y errors
	YErrUpper       []float64    // asymmetric upper y errors
	LoLimits        []bool       // y value is a lower limit
	UpLimits        []bool       // y value is an upper limit
	XLoLimits       []bool       // x value is a lower limit
	XUpLimits       []bool       // x value is an upper limit
	Color           render.Color // stroke color
	LineWidth       float64      // stroke width in pixels
	CapSize         float64      // cap size in pixels
	Marker          MarkerType   // optional data marker, matching Matplotlib fmt markers
	MarkerSet       bool
	MarkerSize      float64 // marker size in points
	Alpha           float64 // alpha transparency (0-1), if 0 uses 1.0
	Label           string  // series label for legend
	NoDataLine      bool    // true matches Matplotlib fmt="none" data-line suppression
	ErrorEvery      int     // draw error bars every N points, default 1
	ErrorEveryStart int     // starting point for ErrorEvery
	z               float64 // z-order
}

// Draw renders each error bar from XY to XY with symmetric offsets.
func (e *ErrorBar) Draw(r render.Renderer, ctx *DrawContext) {
	if len(e.XY) == 0 {
		return
	}

	lineWidth := e.LineWidth
	if lineWidth <= 0 {
		lineWidth = 1.0
	}

	capSizePx := e.CapSize
	if capSizePx < 0 {
		capSizePx = 0
	}
	limitMarkerSizePx := capSizePx * 2
	capHalf := capSizePx * 0.5

	alpha := e.Alpha
	if alpha <= 0 {
		alpha = 1.0
	}
	if alpha > 1 {
		alpha = 1.0
	}

	color := e.Color
	color.A *= alpha
	if color.A <= 0 {
		return
	}

	paint := render.Paint{
		Stroke:    color,
		LineWidth: lineWidth,
		LineJoin:  render.JoinMiter,
		LineCap:   render.CapButt,
		Snap:      render.SnapAuto,
	}

	for i, pt := range e.XY {
		if !e.errorEveryApplies(i) {
			continue
		}
		xLow, xHigh := resolveErrorRange(e.XErr, e.XErrLower, e.XErrUpper, i)
		yLow, yHigh := resolveErrorRange(e.YErr, e.YErrLower, e.YErrUpper, i)
		xLoLimit := resolveBool(e.XLoLimits, i)
		xUpLimit := resolveBool(e.XUpLimits, i)
		loLimit := resolveBool(e.LoLimits, i)
		upLimit := resolveBool(e.UpLimits, i)
		if xLoLimit {
			xLow = 0
		}
		if xUpLimit {
			xHigh = 0
		}
		if loLimit {
			yLow = 0
		}
		if upLimit {
			yHigh = 0
		}

		if xLow > 0 || xHigh > 0 {
			left := geom.Pt{X: pt.X - xLow, Y: pt.Y}
			right := geom.Pt{X: pt.X + xHigh, Y: pt.Y}
			if left != right {
				r.Path(linePath(ctx, left, right), &paint)
			}

			if capHalf > 0 {
				if xLow > 0 && !xUpLimit {
					leftTop := addPixelOffset(ctx, left, 0, -capHalf)
					leftBottom := addPixelOffset(ctx, left, 0, capHalf)
					r.Path(linePath(ctx, leftTop, leftBottom), &paint)
				}
				if xHigh > 0 && !xLoLimit {
					rightTop := addPixelOffset(ctx, right, 0, -capHalf)
					rightBottom := addPixelOffset(ctx, right, 0, capHalf)
					r.Path(linePath(ctx, rightTop, rightBottom), &paint)
				}
				if xLoLimit && xHigh > 0 {
					drawLimitCaret(r, ctx, right, 1, 0, limitMarkerSizePx, &paint)
					drawErrorbarCapMarker(r, ctx, pt, true, capHalf, &paint)
				}
				if xUpLimit && xLow > 0 {
					drawLimitCaret(r, ctx, left, -1, 0, limitMarkerSizePx, &paint)
					drawErrorbarCapMarker(r, ctx, pt, true, capHalf, &paint)
				}
			}
		}

		if yLow > 0 || yHigh > 0 {
			lower := geom.Pt{X: pt.X, Y: pt.Y - yLow}
			upper := geom.Pt{X: pt.X, Y: pt.Y + yHigh}
			if lower != upper {
				r.Path(linePath(ctx, lower, upper), &paint)
			}

			if capHalf > 0 {
				if yLow > 0 && !upLimit {
					lowerLeft := addPixelOffset(ctx, lower, -capHalf, 0)
					lowerRight := addPixelOffset(ctx, lower, capHalf, 0)
					r.Path(linePath(ctx, lowerLeft, lowerRight), &paint)
				}
				if yHigh > 0 && !loLimit {
					upperLeft := addPixelOffset(ctx, upper, -capHalf, 0)
					upperRight := addPixelOffset(ctx, upper, capHalf, 0)
					r.Path(linePath(ctx, upperLeft, upperRight), &paint)
				}
				if loLimit && yHigh > 0 {
					drawLimitCaret(r, ctx, upper, 0, 1, limitMarkerSizePx, &paint)
					drawErrorbarCapMarker(r, ctx, pt, false, capHalf, &paint)
				}
				if upLimit && yLow > 0 {
					drawLimitCaret(r, ctx, lower, 0, -1, limitMarkerSizePx, &paint)
					drawErrorbarCapMarker(r, ctx, pt, false, capHalf, &paint)
				}
			}
		}
	}

	if len(e.XY) > 1 && !e.NoDataLine {
		line := Line2D{
			XY:  append([]geom.Pt(nil), e.XY...),
			W:   lineWidth,
			Col: color,
		}
		if path := line.displayPath(ctx); len(path.C) > 0 {
			linePaint := render.Paint{
				Stroke:    color,
				LineWidth: lineWidth,
				LineJoin:  render.JoinRound,
				LineCap:   render.CapButt,
				Snap:      render.SnapAuto,
				Simplify:  ctx != nil && ctx.RC.PathSimplify,
			}
			if ctx != nil {
				linePaint.SimplifyThreshold = ctx.RC.PathSimplifyThreshold
				linePaint.MaxChunkVertices = ctx.RC.AggPathChunkSize
			}
			r.Path(path, &linePaint)
		}
	}

	if e.MarkerSet && e.MarkerSize > 0 {
		scatter := &Scatter2D{
			XY:        append([]geom.Pt(nil), e.XY...),
			Size:      e.MarkerSize * e.MarkerSize,
			Color:     color,
			EdgeColor: color,
			EdgeWidth: lineWidth,
			Alpha:     alpha,
			Marker:    e.Marker,
			z:         e.Z() + 0.05,
		}
		scatter.Draw(r, ctx)
	}
}

// Z returns the z-order for sorting.
func (e *ErrorBar) Z() float64 {
	return zOrDefault(e.z, defaultLineZ)
}

// Bounds returns the data-space bounding box for bars and error extents.
func (e *ErrorBar) Bounds(*DrawContext) geom.Rect {
	if len(e.XY) == 0 {
		return geom.Rect{}
	}

	bounds := geom.Rect{Min: e.XY[0], Max: e.XY[0]}
	for i, pt := range e.XY {
		if math.IsNaN(pt.X) || math.IsNaN(pt.Y) || math.IsInf(pt.X, 0) || math.IsInf(pt.Y, 0) {
			continue
		}

		if pt.X < bounds.Min.X {
			bounds.Min.X = pt.X
		}
		if pt.Y < bounds.Min.Y {
			bounds.Min.Y = pt.Y
		}
		if pt.X > bounds.Max.X {
			bounds.Max.X = pt.X
		}
		if pt.Y > bounds.Max.Y {
			bounds.Max.Y = pt.Y
		}

		if !e.errorEveryApplies(i) {
			continue
		}
		xLow, xHigh := resolveErrorRange(e.XErr, e.XErrLower, e.XErrUpper, i)
		yLow, yHigh := resolveErrorRange(e.YErr, e.YErrLower, e.YErrUpper, i)
		if resolveBool(e.XLoLimits, i) {
			xLow = 0
		}
		if resolveBool(e.XUpLimits, i) {
			xHigh = 0
		}
		if resolveBool(e.LoLimits, i) {
			yLow = 0
		}
		if resolveBool(e.UpLimits, i) {
			yHigh = 0
		}
		if xLow > 0 || xHigh > 0 {
			left := pt.X - xLow
			right := pt.X + xHigh
			if left < bounds.Min.X {
				bounds.Min.X = left
			}
			if right > bounds.Max.X {
				bounds.Max.X = right
			}
		}
		if yLow > 0 || yHigh > 0 {
			lower := pt.Y - yLow
			upper := pt.Y + yHigh
			if lower < bounds.Min.Y {
				bounds.Min.Y = lower
			}
			if upper > bounds.Max.Y {
				bounds.Max.Y = upper
			}
		}
	}

	return bounds
}

func (e *ErrorBar) errorEveryApplies(i int) bool {
	every := e.ErrorEvery
	if every <= 0 {
		every = 1
	}
	start := e.ErrorEveryStart
	if start < 0 || i < start {
		return false
	}
	return (i-start)%every == 0
}

func resolveErrorRange(symmetric, lower, upper []float64, i int) (float64, float64) {
	low := resolveError(symmetric, i)
	high := low
	if len(lower) > 0 {
		low = resolveError(lower, i)
	}
	if len(upper) > 0 {
		high = resolveError(upper, i)
	}
	return low, high
}

// resolveError returns scalar or broadcasted symmetric error magnitude at index i.
func resolveError(err []float64, i int) float64 {
	if len(err) == 0 {
		return 0
	}
	if len(err) == 1 {
		return math.Abs(err[0])
	}
	if i < len(err) {
		return math.Abs(err[i])
	}
	return 0
}

func resolveBool(values []bool, i int) bool {
	if len(values) == 0 {
		return false
	}
	if len(values) == 1 {
		return values[0]
	}
	return i < len(values) && values[i]
}

func drawLimitCaret(r render.Renderer, ctx *DrawContext, basePt geom.Pt, dirX, dirY, markerSize float64, paint *render.Paint) {
	if r == nil || ctx == nil || markerSize <= 0 {
		return
	}
	base := ctx.DataToPixel.Apply(basePt)
	length := markerSize * 0.75
	half := markerSize * 0.5
	var p1, apex, p2 geom.Pt
	switch {
	case dirX > 0:
		p1 = geom.Pt{X: base.X, Y: base.Y - half}
		apex = geom.Pt{X: base.X + length, Y: base.Y}
		p2 = geom.Pt{X: base.X, Y: base.Y + half}
	case dirX < 0:
		p1 = geom.Pt{X: base.X, Y: base.Y - half}
		apex = geom.Pt{X: base.X - length, Y: base.Y}
		p2 = geom.Pt{X: base.X, Y: base.Y + half}
	case dirY > 0:
		p1 = geom.Pt{X: base.X - half, Y: base.Y}
		apex = geom.Pt{X: base.X, Y: base.Y - length}
		p2 = geom.Pt{X: base.X + half, Y: base.Y}
	default:
		p1 = geom.Pt{X: base.X - half, Y: base.Y}
		apex = geom.Pt{X: base.X, Y: base.Y + length}
		p2 = geom.Pt{X: base.X + half, Y: base.Y}
	}
	markerPaint := *paint
	markerPaint.Fill = render.Color{}
	r.Path(geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo, geom.LineTo},
		V: []geom.Pt{p1, apex, p2},
	}, &markerPaint)
}

func drawErrorbarCapMarker(r render.Renderer, ctx *DrawContext, dataPt geom.Pt, vertical bool, halfSize float64, paint *render.Paint) {
	if r == nil || ctx == nil || halfSize <= 0 {
		return
	}
	pt := ctx.DataToPixel.Apply(dataPt)
	var p1, p2 geom.Pt
	if vertical {
		p1 = geom.Pt{X: pt.X, Y: pt.Y - halfSize}
		p2 = geom.Pt{X: pt.X, Y: pt.Y + halfSize}
	} else {
		p1 = geom.Pt{X: pt.X - halfSize, Y: pt.Y}
		p2 = geom.Pt{X: pt.X + halfSize, Y: pt.Y}
	}
	r.Path(geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo},
		V: []geom.Pt{p1, p2},
	}, paint)
}

func addPixelOffset(ctx *DrawContext, dataPt geom.Pt, dxPx, dyPx float64) geom.Pt {
	basePx := ctx.DataToPixel.Apply(dataPt)
	targetPx := geom.Pt{
		X: basePx.X + dxPx,
		Y: basePx.Y + dyPx,
	}

	dataSpace, ok := invertToData(ctx, targetPx)
	if !ok {
		return dataPt
	}
	return dataSpace
}

func invertToData(ctx *DrawContext, px geom.Pt) (geom.Pt, bool) {
	if ctx == nil {
		return geom.Pt{}, false
	}
	return ctx.DataToPixel.Invert(px)
}
