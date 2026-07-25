package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
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
	LineWidth       float64      // stroke width in points
	CapSize         float64      // cap size in pixels
	CapThick        float64      // cap line width in points (0 uses the 1pt default)
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

	rc := style.Default
	if ctx != nil {
		rc = ctx.RC
	}
	lineWidth := e.LineWidth // points
	if lineWidth <= 0 {
		lineWidth = 1.5
	}

	capSizePx := e.CapSize
	if capSizePx < 0 {
		capSizePx = 0
	}
	limitMarkerSizePx := capSizePx
	if limitMarkerSizePx <= 0 {
		limitMarkerSizePx = pointsToPixels(rc, 6)
	}
	capHalf := capSizePx / 2

	alpha := e.Alpha
	if alpha <= 0 {
		alpha = 1.0
	}
	if alpha > 1 {
		alpha = 1.0
	}

	color := e.Color
	color = color.WithAlphaMultiplier(alpha)
	if color.A <= 0 {
		return
	}

	paint := render.Paint{
		Stroke:    color,
		LineWidth: pointsToPixels(rc, lineWidth),
		LineJoin:  render.JoinMiter,
		LineCap:   render.CapButt,
		Snap:      render.SnapAuto,
	}
	capPaint := paint
	capThick := 1.0 // points
	if e.CapThick > 0 {
		capThick = e.CapThick
	}
	capPaint.LineWidth = pointsToPixels(rc, capThick)
	// Matplotlib draws the caps as Line2D markers ('|' for x-error, '_' for
	// y-error). Its draw_markers floors the marker centre to the device pixel
	// grid and snaps the marker half-extent to an integer separately, then sums
	// them — it does NOT snap a stroked segment's two endpoints independently.
	// faithfulCapPath reproduces that geometry directly, so the cap paths are
	// pre-snapped and must skip the generic per-vertex snapper.
	capLinePaint := capPaint
	capLinePaint.Snap = render.SnapOff

	for i, pt := range e.XY {
		if !e.errorEveryApplies(i) {
			continue
		}
		xLow, xHigh := resolveErrorRange(e.XErr, e.XErrLower, e.XErrUpper, i)
		yLow, yHigh := resolveErrorRange(e.YErr, e.YErrLower, e.YErrUpper, i)
		hasXErr := xLow > 0 || xHigh > 0
		hasYErr := yLow > 0 || yHigh > 0
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

		if hasXErr {
			left := geom.Pt{X: pt.X - xLow, Y: pt.Y}
			right := geom.Pt{X: pt.X + xHigh, Y: pt.Y}
			if left != right {
				r.Path(linePath(ctx, left, right), &paint)
			}

			if capHalf > 0 {
				if xLow > 0 && !xUpLimit {
					r.Path(faithfulCapPath(ctx, left, true, capHalf, capLinePaint.LineWidth), &capLinePaint)
				}
				if xHigh > 0 && !xLoLimit {
					r.Path(faithfulCapPath(ctx, right, true, capHalf, capLinePaint.LineWidth), &capLinePaint)
				}
			}
			if xLoLimit {
				drawLimitCaret(r, ctx, right, 1, 0, limitMarkerSizePx, &paint)
				if capHalf > 0 {
					drawErrorbarCapMarker(r, ctx, pt, true, capHalf, &capPaint)
				}
			}
			if xUpLimit {
				drawLimitCaret(r, ctx, left, -1, 0, limitMarkerSizePx, &paint)
				if capHalf > 0 {
					drawErrorbarCapMarker(r, ctx, pt, true, capHalf, &capPaint)
				}
			}
		}

		if hasYErr {
			lower := geom.Pt{X: pt.X, Y: pt.Y - yLow}
			upper := geom.Pt{X: pt.X, Y: pt.Y + yHigh}
			if lower != upper {
				r.Path(linePath(ctx, lower, upper), &paint)
			}

			if capHalf > 0 {
				if yLow > 0 && !upLimit {
					r.Path(faithfulCapPath(ctx, lower, false, capHalf, capLinePaint.LineWidth), &capLinePaint)
				}
				if yHigh > 0 && !loLimit {
					r.Path(faithfulCapPath(ctx, upper, false, capHalf, capLinePaint.LineWidth), &capLinePaint)
				}
			}
			if loLimit {
				drawLimitCaret(r, ctx, upper, 0, 1, limitMarkerSizePx, &paint)
				if capHalf > 0 {
					drawErrorbarCapMarker(r, ctx, pt, false, capHalf, &capPaint)
				}
			}
			if upLimit {
				drawLimitCaret(r, ctx, lower, 0, -1, limitMarkerSizePx, &paint)
				if capHalf > 0 {
					drawErrorbarCapMarker(r, ctx, pt, false, capHalf, &capPaint)
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
				LineWidth: pointsToPixels(rc, lineWidth),
				LineJoin:  rc.Lines.SolidJoin,
				LineCap:   rc.Lines.SolidCap,
				Snap:      render.SnapAuto,
				Simplify:  ctx != nil && ctx.RC.PathSimplify,
			}
			linePaint.Antialias = render.AntialiasOff
			if rc.Lines.Antialiased {
				linePaint.Antialias = render.AntialiasOn
			}
			if ctx != nil {
				linePaint.SimplifyThreshold = ctx.RC.PathSimplifyThreshold
				linePaint.MaxChunkVertices = ctx.RC.AggPathChunkSize
			}
			r.Path(path, &linePaint)
		}
	}

	if e.MarkerSet && e.MarkerSize > 0 {
		markerLine := &Line2D{
			XY:         append([]geom.Pt(nil), e.XY...),
			Col:        color,
			Marker:     e.Marker,
			MarkerSet:  true,
			MarkerSize: e.MarkerSize,
			z:          e.Z() + 0.05,
		}
		applyLineRCDefaults(markerLine, &rc)
		if markerLine.MarkerFaceSpec.Mode == MarkerColorExplicit {
			markerLine.MarkerFaceSpec.Color.A = color.A
		}
		if markerLine.MarkerEdgeSpec.Mode == MarkerColorExplicit {
			markerLine.MarkerEdgeSpec.Color.A = color.A
		}
		markerLine.drawMarkers(r, ctx)
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
	var marker geom.Path
	switch {
	case dirX > 0:
		marker = markerCaretPath(90, true)
	case dirX < 0:
		marker = markerCaretPath(270, true)
	case dirY > 0:
		marker = markerCaretPath(180, true)
	default:
		marker = markerCaretPath(0, true)
	}
	markerPaint := *paint
	markerPaint.Fill = paint.Stroke
	markerPaint.LineJoin = render.JoinMiter
	markerPaint.LineCap = render.CapButt
	r.Path(scaleAndTranslatePath(marker, markerSize, base), &markerPaint)
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

// ebSnap mirrors the AGG backend's PathSnapper rounding (Matplotlib's
// floor(v + 0.5)): vertices land on integer pixel boundaries.
func ebSnap(v float64) float64 {
	return math.Floor(v + 0.5 + 1e-9)
}

// faithfulCapPath builds a single error-bar cap segment in y-up display space,
// reproducing Matplotlib's cap-as-marker geometry exactly.
//
// Matplotlib renders caps as Line2D markers ('|' for x-error, '_' for
// y-error). In src/_backend_agg.h::draw_markers the marker offset (the cap
// centre) is floored to the device pixel grid, while the marker's own vertices
// are snapped to integers independently by the PathSnapper, and the two are then
// summed. A stroked segment drawn through the un-floored centre instead snaps its
// two endpoints jointly, which rounds asymmetrically whenever the centre sits on
// a half-pixel and the half-extent is non-integral — pushing the cap a pixel off.
//
// The (1,-1)+height device flip cancels algebraically, so the result is computed
// purely from the display-space data point: base is the device-floored centre
// mapped back to display, xSnap is the floored centre column, and hHi/hLo are the
// snapped marker half-extents. snapValue follows the same odd-stroke-width rule as
// the backend snapper. The path is pre-snapped, so callers draw it with Snap off.
func faithfulCapPath(ctx *DrawContext, center geom.Pt, vertical bool, half, strokeWidth float64) geom.Path {
	p := ctx.DataToPixel.Apply(center)
	snapValue := 0.0
	if int(math.Round(strokeWidth))%2 != 0 {
		snapValue = 0.5
	}
	xSnap := ebSnap(p.X)
	// height - floor(height - p.Y + 0.5): the height term cancels, leaving a
	// device-floored centre expressed back in y-up display space.
	base := -math.Floor(0.5 - p.Y + 1e-9)
	hHi := ebSnap(half)
	hLo := ebSnap(-half)

	var a, b geom.Pt
	if vertical {
		a = geom.Pt{X: xSnap + snapValue, Y: base - hHi - snapValue}
		b = geom.Pt{X: xSnap + snapValue, Y: base - hLo - snapValue}
	} else {
		a = geom.Pt{X: xSnap + hLo + snapValue, Y: base - snapValue}
		b = geom.Pt{X: xSnap + hHi + snapValue, Y: base - snapValue}
	}
	return geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo},
		V: []geom.Pt{a, b},
	}
}
