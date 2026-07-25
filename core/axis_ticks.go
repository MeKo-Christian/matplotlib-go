package core

import (
	"fmt"
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/ticker"
)

// DrawTicks renders tick marks pointing outward from the plot area.
// Called outside the clip region so ticks are visible.
func (a *Axis) DrawTicks(r render.Renderer, ctx *DrawContext) {
	if isPolarProjection(ctx.Projection) {
		a.drawPolarTicks(r, ctx)
		return
	}
	var domainMin, domainMax float64
	var isXAxis bool

	switch a.Side {
	case AxisBottom, AxisTop:
		domainMin, domainMax = ctx.DataToPixel.XScale.Domain()
		isXAxis = true
	case AxisLeft, AxisRight:
		domainMin, domainMax = ctx.DataToPixel.YScale.Domain()
	}

	// Minor ticks first.
	showMinorTicks := a.ShowMinorTicks
	if !a.minorTickVisibilitySet {
		// Preserve the historical direct-field contract: ShowTicks=false hides
		// both levels unless TickParams/rcParams explicitly split them.
		showMinorTicks = a.ShowTicks
	}
	if showMinorTicks && a.MinorLocator != nil {
		minorLoc := ticker.WithMajorContext(a.MinorLocator, a.Locator)
		minorTicks := visibleTicks(minorLoc.Ticks(domainMin, domainMax, a.minorTickTargetCountForContext(ctx, isXAxis)), domainMin, domainMax)
		if len(minorTicks) > 0 {
			a.drawMinorTicks(r, ctx, minorTicks, isXAxis)
		}
	}

	if a.ShowTicks {
		ticks := visibleTicks(a.Locator.Ticks(domainMin, domainMax, a.majorTickTargetCountForContext(ctx, isXAxis)), domainMin, domainMax)
		if len(ticks) > 0 {
			a.drawTicks(r, ctx, ticks, isXAxis)
		}
	}

	for _, level := range a.ExtraTickLevels {
		if !level.ShowTicks || level.Locator == nil {
			continue
		}
		ticks := visibleTicks(level.Locator.Ticks(domainMin, domainMax, a.majorTickTargetCountForContext(ctx, isXAxis)), domainMin, domainMax)
		if len(ticks) == 0 {
			continue
		}
		size := tickLevelSize(level, a.TickSize)
		for _, tickValue := range ticks {
			a.drawSingleTick(r, ctx, tickValue, size, a.tickLineWidth(), a.tickColor(), isXAxis)
		}
	}
}

// drawTicks draws tick marks at the specified positions.
func (a *Axis) drawTicks(r render.Renderer, ctx *DrawContext, ticks []float64, isXAxis bool) {
	for _, tickValue := range ticks {
		a.drawSingleTick(r, ctx, tickValue, a.TickSize, a.tickLineWidth(), a.tickColor(), isXAxis)
	}
}

// drawMinorTicks draws smaller tick marks at the specified positions.
func (a *Axis) drawMinorTicks(r render.Renderer, ctx *DrawContext, ticks []float64, isXAxis bool) {
	sz := a.minorTickSize()
	for _, tickValue := range ticks {
		a.drawSingleTick(r, ctx, tickValue, sz, a.minorTickLineWidth(), a.minorTickColor(), isXAxis)
	}
}

// drawSingleTick draws a single tick mark pointing outward from the plot area.
func (a *Axis) drawSingleTick(r render.Renderer, ctx *DrawContext, tickValue, tickSize, lineWidth float64, stroke render.Color, isXAxis bool) {
	var p1, p2 geom.Pt

	// Snap with the tick's own stroke-width parity so even-width marks centre on
	// the pixel boundary exactly like Matplotlib's PathSnapper (default 0.8pt
	// ticks round to an odd pixel width and are unchanged from snapDisplay*).
	lineWidthPx := pointsToPixels(ctx.RC, lineWidth)

	if isXAxis {
		spineY := getSpinePosition(a, ctx)
		spinePixel := axisTickDisplayPoint(a, ctx, tickValue, true, spineY)
		spinePixel.X = snapDisplayXForWidth(spinePixel.X, lineWidthPx)
		spinePixel.Y = snapDisplayYForWidth(spinePixel.Y, lineWidthPx, figureSnapHeight(ctx))

		p1, p2 = axisTickSegment(a, spinePixel, tickSize, true)
	} else {
		spineX := getSpinePosition(a, ctx)
		spinePixel := axisTickDisplayPoint(a, ctx, tickValue, false, spineX)
		spinePixel.X = snapDisplayXForWidth(spinePixel.X, lineWidthPx)
		spinePixel.Y = snapDisplayYForWidth(spinePixel.Y, lineWidthPx, figureSnapHeight(ctx))

		p1, p2 = axisTickSegment(a, spinePixel, tickSize, false)
	}

	// Create tick path
	path := geom.Path{}
	path.C = append(path.C, geom.MoveTo)
	path.V = append(path.V, p1)
	path.C = append(path.C, geom.LineTo)
	path.V = append(path.V, p2)

	// Draw the tick
	paint := render.Paint{
		LineWidth: pointsToPixels(ctx.RC, lineWidth),
		Stroke:    stroke,
		LineCap:   a.TickLineCap,
		LineJoin:  a.TickLineJoin,
		Dashes:    styleCloneDashes(a.Dashes),
	}
	r.Path(path, &paint)
}

func (a *Axis) majorTickTargetCount() int {
	if a == nil || a.MajorTickCount <= 0 {
		return 6
	}
	return a.MajorTickCount
}

func (a *Axis) minorTickTargetCount() int {
	if a == nil || a.MinorTickCount <= 0 {
		return 30
	}
	return a.MinorTickCount
}

func (a *Axis) majorTickTargetCountForContext(ctx *DrawContext, isXAxis bool) int {
	target := a.majorTickTargetCount()
	if ctx == nil || target <= 1 || a.majorTickCountFixed {
		return target
	}

	dpi := ctx.RC.DPI
	if dpi <= 0 {
		dpi = 100
	}
	clip := sharedRootTickClip(a, ctx, isXAxis)
	lengthPx := clip.H()
	labelSize := ctx.RC.TickLabelSize("y")
	labelAspect := 2.0
	if isXAxis {
		lengthPx = clip.W()
		labelSize = ctx.RC.TickLabelSize("x")
		labelAspect = 3.0
	}
	lengthPt := lengthPx * 72.0 / dpi
	minSpacingPt := labelSize * labelAspect
	if lengthPt <= 0 || minSpacingPt <= 0 {
		return target
	}

	capacity := int(math.Floor(lengthPt / minSpacingPt))
	if capacity < 1 {
		capacity = 1
	}
	if capacity < target {
		return capacity
	}
	return target
}

func sharedRootTickClip(axis *Axis, ctx *DrawContext, isXAxis bool) geom.Rect {
	if axis == nil || ctx == nil || ctx.Axes == nil {
		return ctx.Clip
	}
	root := (*Axes)(nil)
	if isXAxis && ctx.Axes.shareX != nil && ctx.Axes.XAxis == axis {
		root = ctx.Axes.xScaleRoot()
	}
	if !isXAxis && ctx.Axes.shareY != nil && ctx.Axes.YAxis == axis {
		root = ctx.Axes.yScaleRoot()
	}
	if root == nil && ctx.Axes.figure != nil {
		bestClip := ctx.Clip
		bestLength := bestClip.H()
		if isXAxis {
			bestLength = bestClip.W()
		}
		for _, candidate := range ctx.Axes.figure.Children {
			if candidate == nil {
				continue
			}
			if isXAxis && candidate.XAxis != axis {
				continue
			}
			if !isXAxis && candidate.YAxis != axis {
				continue
			}
			candidateClip := candidate.adjustedLayout(candidate.figure)
			candidateLength := candidateClip.H()
			if isXAxis {
				candidateLength = candidateClip.W()
			}
			if candidateLength > bestLength {
				bestLength = candidateLength
				root = candidate
				bestClip = candidateClip
			}
		}
		if root == nil {
			return bestClip
		}
	}
	if root == nil || root == ctx.Axes || root.figure == nil {
		return ctx.Clip
	}
	return root.adjustedLayout(root.figure)
}

func (a *Axis) minorTickTargetCountForContext(ctx *DrawContext, isXAxis bool) int {
	target := a.minorTickTargetCount()
	major := a.majorTickTargetCountForContext(ctx, isXAxis)
	limit := major * 5
	if limit < 1 {
		limit = 1
	}
	if target > limit {
		return limit
	}
	return target
}

func visibleTicks(ticks []float64, minVal, maxVal float64) []float64 {
	if len(ticks) == 0 {
		return nil
	}
	if minVal > maxVal {
		minVal, maxVal = maxVal, minVal
	}

	span := maxVal - minVal
	tol := 1e-9 * math.Max(1, span)

	out := make([]float64, 0, len(ticks))
	for _, tick := range ticks {
		if tick < minVal-tol || tick > maxVal+tol {
			continue
		}
		if approx(tick, minVal, tol) {
			tick = minVal
		} else if approx(tick, maxVal, tol) {
			tick = maxVal
		}
		out = append(out, tick)
	}
	return out
}

func axisTickDisplayPoint(a *Axis, ctx *DrawContext, tickValue float64, isXAxis bool, spineValue float64) geom.Pt {
	var point geom.Pt
	if !isXAxis {
		if pt, ok := skewYAxisDisplayPoint(a, ctx, tickValue); ok {
			point = pt
		} else {
			point = ctx.DataToPixel.Apply(geom.Pt{X: spineValue, Y: tickValue})
		}
	} else {
		point = ctx.DataToPixel.Apply(geom.Pt{X: tickValue, Y: spineValue})
	}

	if a != nil &&
		(a.SpinePositionMode == AxisSpinePositionAxes ||
			a.SpinePositionMode == AxisSpinePositionOutward) {
		if position, ok := axisSpineDisplayCoordinate(a, ctx); ok {
			if isXAxis {
				point.Y = position
			} else {
				point.X = position
			}
		}
	}
	return point
}

func axisTickSegment(axis *Axis, spine geom.Pt, tickSize float64, isXAxis bool) (geom.Pt, geom.Pt) {
	if axis == nil {
		return spine, spine
	}
	if tickSize > 0 {
		tickSize = math.Round(tickSize)
	}

	outward := tickSize
	if (axis.Side == AxisTop) || (axis.Side == AxisLeft) {
		outward = -tickSize
	}

	var inward float64
	if outward < 0 {
		inward = tickSize
	} else {
		inward = -tickSize
	}

	switch axis.TickDirection {
	case TickDirectionIn:
		if isXAxis {
			return spine, geom.Pt{X: spine.X, Y: spine.Y - inward}
		}
		return spine, geom.Pt{X: spine.X + inward, Y: spine.Y}
	case TickDirectionInOut:
		if isXAxis {
			return geom.Pt{X: spine.X, Y: spine.Y + outward/2}, geom.Pt{X: spine.X, Y: spine.Y - outward/2}
		}
		return geom.Pt{X: spine.X - outward/2, Y: spine.Y}, geom.Pt{X: spine.X + outward/2, Y: spine.Y}
	default:
		if isXAxis {
			return spine, geom.Pt{X: spine.X, Y: spine.Y - outward}
		}
		return spine, geom.Pt{X: spine.X + outward, Y: spine.Y}
	}
}

// SetTickDirection configures whether ticks point outward, inward, or in both directions.
func (a *Axis) SetTickDirection(direction string) error {
	if a == nil {
		return nil
	}
	switch strings.ToLower(strings.TrimSpace(direction)) {
	case "", "out", "outward":
		a.TickDirection = TickDirectionOut
	case "in", "inward":
		a.TickDirection = TickDirectionIn
	case "inout", "both":
		a.TickDirection = TickDirectionInOut
	default:
		return fmt.Errorf("unsupported tick direction %q", direction)
	}
	return nil
}

func axisStrokePaint(a *Axis, ctx *DrawContext, forTicks bool) render.Paint {
	if a == nil {
		return render.Paint{}
	}
	lineCap := a.LineCap
	join := a.LineJoin
	stroke := a.Color
	if forTicks {
		lineCap = a.TickLineCap
		join = a.TickLineJoin
		stroke = a.tickColor()
	}
	return render.Paint{
		LineWidth: pointsToPixels(ctx.RC, a.LineWidth),
		Stroke:    stroke,
		LineCap:   lineCap,
		LineJoin:  join,
		Dashes:    styleCloneDashes(a.Dashes),
	}
}

func (a *Axis) tickColor() render.Color {
	if a == nil {
		return render.Color{}
	}
	if a.TickColor != nil {
		return *a.TickColor
	}
	return a.Color
}

func (a *Axis) tickLabelColor() render.Color {
	if a == nil {
		return render.Color{}
	}
	if a.TickLabelColor != nil {
		return *a.TickLabelColor
	}
	return a.tickColor()
}

// minorTickColor resolves the minor tick mark color. A nil override falls back
// to the major tick color so that existing single-color configurations keep
// their behavior; an explicit minor color (via tick_params which="minor") is
// independent of the major color, matching matplotlib.
func (a *Axis) minorTickColor() render.Color {
	if a == nil {
		return render.Color{}
	}
	if a.MinorTickColor != nil {
		return *a.MinorTickColor
	}
	return a.tickColor()
}

func (a *Axis) minorTickLabelColor() render.Color {
	if a == nil {
		return render.Color{}
	}
	if a.MinorTickLabelColor != nil {
		return *a.MinorTickLabelColor
	}
	return a.minorTickColor()
}

func (a *Axis) tickLineWidth() float64 {
	if a == nil {
		return 0
	}
	if a.tickLineWidthSet {
		return a.TickLineWidth
	}
	if a.TickLineWidth > 0 {
		return a.TickLineWidth
	}
	return a.LineWidth
}

func (a *Axis) minorTickLineWidth() float64 {
	if a == nil {
		return 0
	}
	if a.minorTickLineWidthSet {
		return a.MinorTickLineWidth
	}
	if a.MinorTickLineWidth > 0 {
		return a.MinorTickLineWidth
	}
	return a.tickLineWidth()
}

func tickLevelSize(level TickLevel, fallback float64) float64 {
	if level.Size > 0 {
		return level.Size
	}
	return fallback
}

// minorTickSize falls back to matplotlib's xtick.minor.size / ytick.minor.size
// default of 2.0 pt when no explicit minor tick size is set.
func (a *Axis) minorTickSize() float64 {
	if a != nil && a.minorTickSizeSet {
		return a.MinorTickSize
	}
	if a.MinorTickSize > 0 {
		return a.MinorTickSize
	}
	return defaultMinorTickSizePx
}

func (a *Axis) AddTickLevel(level TickLevel) {
	if a == nil {
		return
	}
	level.LabelStyle = normalizeTickLabelStyle(level.LabelStyle)
	a.ExtraTickLevels = append(a.ExtraTickLevels, level)
}

func (a *Axis) ClearTickLevels() {
	if a == nil {
		return
	}
	a.ExtraTickLevels = nil
}
