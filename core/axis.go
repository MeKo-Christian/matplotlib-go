package core

import (
	"fmt"
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

// AxisSide specifies which side of the plot area an axis is on.
type AxisSide uint8

const (
	AxisBottom AxisSide = iota // x-axis at bottom
	AxisTop                    // x-axis at top
	AxisLeft                   // y-axis at left
	AxisRight                  // y-axis at right
)

// Matplotlib defaults axes.linewidth to 0.8 pt and major ticks to 3.5 pt.
// The default figure DPI is 100, so store the pixel equivalent for axes
// constructed without a figure-backed RC.
const (
	defaultAxisLineWidth = 0.8 * 100.0 / 72.0
	defaultTickSizePt    = 3.5
	defaultTickSizePx    = defaultTickSizePt * 100.0 / 72.0
	defaultTickPadPt     = 3.5
)

// TickLabelStyle captures axis-owned label placement and orientation.
type TickLabelStyle struct {
	Rotation  float64
	Pad       float64
	HAlign    TextAlign
	VAlign    TextVerticalAlign
	FontSize  float64
	FontKey   string
	AutoAlign bool
}

// TickLevel adds an optional additional tick/label row to an axis.
type TickLevel struct {
	Locator    Locator
	Formatter  Formatter
	Size       float64
	ShowTicks  bool
	ShowLabels bool
	LabelStyle TickLabelStyle
}

// TickDirection controls whether ticks point outward, inward, or straddle the spine.
type TickDirection uint8

const (
	TickDirectionOut TickDirection = iota
	TickDirectionIn
	TickDirectionInOut
)

// AxisSpinePositionMode controls where the axis spine is drawn.
type AxisSpinePositionMode uint8

const (
	AxisSpinePositionBoundary AxisSpinePositionMode = iota
	AxisSpinePositionData
)

// Axis renders axis spines, ticks, and labels for a single dimension.
type Axis struct {
	Side                AxisSide     // which side of the plot
	Locator             Locator      // major tick position calculator
	MinorLocator        Locator      // minor tick position calculator (nil = no minor ticks)
	Formatter           Formatter    // major tick label formatter
	MinorFormatter      Formatter    // optional minor tick label formatter
	Color               render.Color // axis spine color, and tick/label color unless overridden
	TickColor           *render.Color
	TickLabelColor      *render.Color
	MinorTickColor      *render.Color // minor tick mark color (nil falls back to TickColor)
	MinorTickLabelColor *render.Color // minor tick label color (nil falls back to MinorTickColor)
	LineWidth           float64       // width of axis spine
	LineCap             render.LineCap
	LineJoin            render.LineJoin
	TickLineCap         render.LineCap
	TickLineJoin        render.LineJoin
	TickLineWidth       float64
	MinorTickLineWidth  float64
	Dashes              []float64
	TickSize            float64 // length of major tick marks (in pixels)
	MinorTickSize       float64 // length of minor tick marks (in pixels); 0 uses TickSize*0.6
	MajorTickCount      int     // target major tick count for automatic locators
	MinorTickCount      int     // target minor tick count for automatic locators
	majorTickCountFixed bool
	TickDirection       TickDirection
	SpinePositionMode   AxisSpinePositionMode
	SpinePosition       float64
	ShowSpine           bool // whether to draw the axis line
	ShowTicks           bool // whether to draw major/minor tick marks
	ShowLabels          bool // whether to draw major tick labels
	ShowMinorLabels     bool // whether to draw minor tick labels
	MajorLabelStyle     TickLabelStyle
	MinorLabelStyle     TickLabelStyle
	ExtraTickLevels     []TickLevel
	z                   float64 // z-order
}

// NewXAxis creates an axis for the bottom (x-axis).
func NewXAxis() *Axis {
	return &Axis{
		Side:              AxisBottom,
		Locator:           AutoLocator{},
		Formatter:         ScalarFormatter{Prec: 3},
		Color:             render.Color{R: 0, G: 0, B: 0, A: 1}, // black
		LineWidth:         defaultAxisLineWidth,
		LineCap:           render.CapSquare,
		LineJoin:          render.JoinMiter,
		TickLineCap:       render.CapButt,
		TickLineJoin:      render.JoinMiter,
		TickSize:          defaultTickSizePx,
		MajorTickCount:    9,
		MinorTickCount:    30,
		TickDirection:     TickDirectionOut,
		SpinePositionMode: AxisSpinePositionBoundary,
		ShowSpine:         true,
		ShowTicks:         true,
		ShowLabels:        true,
		MajorLabelStyle:   defaultTickLabelStyle(),
		MinorLabelStyle:   defaultTickLabelStyle(),
	}
}

// NewYAxis creates an axis for the left (y-axis).
func NewYAxis() *Axis {
	return &Axis{
		Side:              AxisLeft,
		Locator:           AutoLocator{},
		Formatter:         ScalarFormatter{Prec: 3},
		Color:             render.Color{R: 0, G: 0, B: 0, A: 1}, // black
		LineWidth:         defaultAxisLineWidth,
		LineCap:           render.CapSquare,
		LineJoin:          render.JoinMiter,
		TickLineCap:       render.CapButt,
		TickLineJoin:      render.JoinMiter,
		TickSize:          defaultTickSizePx,
		MajorTickCount:    9,
		MinorTickCount:    30,
		TickDirection:     TickDirectionOut,
		SpinePositionMode: AxisSpinePositionBoundary,
		ShowSpine:         true,
		ShowTicks:         true,
		ShowLabels:        true,
		MajorLabelStyle:   defaultTickLabelStyle(),
		MinorLabelStyle:   defaultTickLabelStyle(),
	}
}

// Draw renders the axis spine on the axes edge.
func (a *Axis) Draw(r render.Renderer, ctx *DrawContext) {
	if isPolarProjection(ctx.Projection) {
		a.drawPolarSpine(r, ctx)
		return
	}
	if framePath, ok := projectionFramePath(ctx.Projection, ctx.Clip); ok {
		if a.ShowSpine && (a.Side == AxisBottom || a.Side == AxisTop) {
			paint := axisStrokePaint(a, false)
			paint.LineCap = render.CapButt
			r.Path(framePath, &paint)
		}
		return
	}
	if a.ShowSpine {
		a.drawSpine(r, ctx)
	}
}

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

	if a.ShowTicks {
		// Minor ticks first
		if a.MinorLocator != nil {
			minorLoc := locatorWithMajorContext(a.MinorLocator, a.Locator)
			minorTicks := visibleTicks(minorLoc.Ticks(domainMin, domainMax, a.minorTickTargetCountForContext(ctx, isXAxis)), domainMin, domainMax)
			if len(minorTicks) > 0 {
				a.drawMinorTicks(r, ctx, minorTicks, isXAxis)
			}
		}

		// Major ticks
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

// drawSpine draws the main axis line directly in pixel space, centered on the
// axes edge so the stroke can extend on both sides like Matplotlib's spines.
func (a *Axis) drawSpine(r render.Renderer, ctx *DrawContext) {
	px := ctx.Clip
	lw := a.LineWidth
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
// given side of px. This is the most honest translation of Matplotlib 3.10.9's
// AGG PathSnapper for axis-aligned, 1 px-ish spines: SNAP_AUTO rounds display
// coordinates with floor(coord+0.5)+0.5, so RMSE regressions after this point
// should be chased in layout/geometry instead of by weakening the snap.
func spinePixelEndpoints(side AxisSide, px geom.Rect) (geom.Pt, geom.Pt) {
	x1 := snapDisplayX(px.Min.X)
	y1 := snapDisplayY(px.Min.Y)
	x2 := snapDisplayX(px.Max.X)
	y2 := snapDisplayY(px.Max.Y)

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

func snapDisplayY(y float64) float64 {
	return math.Floor(y+0.5) - 0.5
}

// drawTicks draws tick marks at the specified positions.
func (a *Axis) drawTicks(r render.Renderer, ctx *DrawContext, ticks []float64, isXAxis bool) {
	for _, tickValue := range ticks {
		a.drawSingleTick(r, ctx, tickValue, a.TickSize, a.tickLineWidth(), a.tickColor(), isXAxis)
	}
}

// drawMinorTicks draws smaller tick marks at the specified positions.
func (a *Axis) drawMinorTicks(r render.Renderer, ctx *DrawContext, ticks []float64, isXAxis bool) {
	sz := a.MinorTickSize
	if sz <= 0 {
		sz = a.TickSize * 0.6
	}
	for _, tickValue := range ticks {
		a.drawSingleTick(r, ctx, tickValue, sz, a.minorTickLineWidth(), a.minorTickColor(), isXAxis)
	}
}

// drawSingleTick draws a single tick mark pointing outward from the plot area.
func (a *Axis) drawSingleTick(r render.Renderer, ctx *DrawContext, tickValue, tickSize, lineWidth float64, stroke render.Color, isXAxis bool) {
	var p1, p2 geom.Pt

	if isXAxis {
		spineY := getSpinePosition(a, ctx)
		spinePixel := axisTickDisplayPoint(a, ctx, tickValue, true, spineY)
		spinePixel.X = math.Round(spinePixel.X) + 0.5
		spinePixel.Y = math.Round(spinePixel.Y) - 0.5

		p1, p2 = axisTickSegment(a, spinePixel, tickSize, true)
	} else {
		spineX := getSpinePosition(a, ctx)
		spinePixel := axisTickDisplayPoint(a, ctx, tickValue, false, spineX)
		spinePixel.X = math.Round(spinePixel.X) + 0.5
		spinePixel.Y = math.Round(spinePixel.Y) - 0.5

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
		LineWidth: lineWidth,
		Stroke:    stroke,
		LineCap:   a.TickLineCap,
		LineJoin:  a.TickLineJoin,
		Dashes:    styleCloneDashes(a.Dashes),
	}
	r.Path(path, &paint)
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
		LineWidth: ref.LineWidth,
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
		p1, p2 := spinePixelEndpoints(AxisTop, ctx.Clip)
		drawLine(p1, p2)
	}
	if drawRight {
		p1, p2 := spinePixelEndpoints(AxisRight, ctx.Clip)
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

// Z returns the z-order for sorting.
func (a *Axis) Z() float64 {
	return a.z
}

// Bounds returns an empty rect for now.
func (a *Axis) Bounds(*DrawContext) geom.Rect {
	return geom.Rect{}
}

// DrawTickLabels draws tick labels outside the clip region (call after r.Restore()).
func (a *Axis) DrawTickLabels(r render.Renderer, ctx *DrawContext) {
	if isPolarProjection(ctx.Projection) {
		a.drawPolarTickLabels(r, ctx)
		return
	}
	if isGeoProjection(ctx.Projection) {
		a.drawGeoTickLabels(r, ctx)
		return
	}
	if !a.ShowLabels && !a.ShowMinorLabels && len(a.ExtraTickLevels) == 0 {
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
		isXAxis = false
	}
	if a.ShowLabels && a.Locator != nil && a.Formatter != nil {
		ticks := visibleTicks(a.Locator.Ticks(domainMin, domainMax, a.majorTickTargetCountForContext(ctx, isXAxis)), domainMin, domainMax)
		a.drawTickLabels(r, ctx, ticks, a.Formatter, a.MajorLabelStyle, a.TickSize, a.tickLabelColor(), isXAxis)
		a.drawTickOffsetText(r, ctx, ticks, a.Formatter, a.MajorLabelStyle, a.TickSize, a.tickLabelColor(), isXAxis)
	}
	if a.ShowMinorLabels && a.MinorLocator != nil && a.MinorFormatter != nil {
		minorLoc := locatorWithMajorContext(a.MinorLocator, a.Locator)
		ticks := visibleTicks(minorLoc.Ticks(domainMin, domainMax, a.minorTickTargetCountForContext(ctx, isXAxis)), domainMin, domainMax)
		a.drawTickLabels(r, ctx, ticks, a.MinorFormatter, a.MinorLabelStyle, a.minorTickSize(), a.minorTickLabelColor(), isXAxis)
	}
	for _, level := range a.ExtraTickLevels {
		if !level.ShowLabels || level.Locator == nil || level.Formatter == nil {
			continue
		}
		ticks := visibleTicks(level.Locator.Ticks(domainMin, domainMax, a.majorTickTargetCountForContext(ctx, isXAxis)), domainMin, domainMax)
		a.drawTickLabels(r, ctx, ticks, level.Formatter, normalizeTickLabelStyle(level.LabelStyle), tickLevelSize(level, a.TickSize), a.tickLabelColor(), isXAxis)
	}
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
	lengthPx := ctx.Clip.H()
	labelSize := ctx.RC.TickLabelSize("y")
	labelAspect := 2.0
	if isXAxis {
		lengthPx = ctx.Clip.W()
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

// drawTickLabels draws text labels for a single tick level if the renderer supports text.
func (a *Axis) drawTickLabels(r render.Renderer, ctx *DrawContext, ticks []float64, formatter Formatter, style TickLabelStyle, tickSize float64, labelColor render.Color, isXAxis bool) {
	textRen, ok := r.(render.TextDrawer)
	if !ok || formatter == nil {
		return
	}

	style = normalizeTickLabelStyle(style)
	fontSize := tickLabelFontSizeForStyle(a, style, ctx)
	fontKey := tickLabelFontKey(style, ctx)
	labelPadPx := tickLabelPadForAxisSize(a, tickSize, style, ctx)

	step := 0.0
	if len(ticks) >= 2 {
		step = ticks[1] - ticks[0]
	}

	var rotRen render.RotatedTextDrawer
	if drawer, ok := r.(render.RotatedTextDrawer); ok {
		rotRen = drawer
	}

	for i, tickValue := range ticks {
		label := formatTickLabel(formatter, tickValue, i, ticks)
		if scalarFormatter, ok := formatter.(ScalarFormatter); ok && step > 0 {
			label = formatScalarTickLabel(scalarFormatter, tickValue, step)
		}
		if label == "" {
			continue
		}

		layout := measureSingleLineTextLayout(r, label, fontSize, fontKey, ctx.RC.UseTeX)

		labelPos, ok := tickLabelOrigin(a, ctx, tickValue, layout, labelPadPx, style, isXAxis)
		if !ok {
			continue
		}

		if style.Rotation != 0 && rotRen != nil {
			hAlign, vAlign := resolvedTickLabelLayoutAlignments(a.Side, style, isXAxis)
			angle := style.Rotation * math.Pi / 180.0
			drawDisplayTextRotated(rotRen, label, tickLabelRotationAnchor(labelPos, layout, hAlign, vAlign, angle), fontSize, angle, labelColor, fontKey, ctx.RC.UseTeX)
			continue
		}

		drawDisplayText(textRen, label, labelPos, fontSize, labelColor, fontKey, ctx.RC.UseTeX)
	}
}

func (a *Axis) drawTickOffsetText(r render.Renderer, ctx *DrawContext, ticks []float64, formatter Formatter, style TickLabelStyle, tickSize float64, labelColor render.Color, isXAxis bool) {
	textRen, ok := r.(render.TextDrawer)
	if !ok || formatter == nil || !isXAxis {
		return
	}
	offsetter, ok := formatter.(OffsetFormatter)
	if !ok {
		return
	}
	label := offsetter.OffsetText(ticks)
	if label == "" {
		return
	}

	style = normalizeTickLabelStyle(style)
	fontSize := tickLabelFontSizeForStyle(a, style, ctx)
	fontKey := tickLabelFontKey(style, ctx)
	labelPadPx := tickLabelPadForAxisSize(a, tickSize, style, ctx)
	layout := measureSingleLineTextLayout(r, label, fontSize, fontKey, ctx.RC.UseTeX)
	gap := 0.3 * fontSize

	var anchor geom.Pt
	switch a.Side {
	case AxisBottom:
		anchor = geom.Pt{X: ctx.Clip.Max.X, Y: ctx.Clip.Min.Y - labelPadPx - layout.Height - gap}
	case AxisTop:
		anchor = geom.Pt{X: ctx.Clip.Max.X, Y: ctx.Clip.Max.Y + labelPadPx + layout.Height + gap}
	default:
		return
	}
	origin := geom.Pt{
		X: anchor.X - textHorizontalOriginOffset(layout, TextAlignRight),
		Y: anchor.Y + textBaselineOffset(layout, textLayoutVAlignTop),
	}
	drawDisplayText(textRen, label, origin, fontSize, labelColor, fontKey, ctx.RC.UseTeX)
}

func tickLabelFontSize(a *Axis, ctx *DrawContext) float64 {
	return tickLabelFontSizeForStyle(a, TickLabelStyle{}, ctx)
}

func tickLabelFontSizeForStyle(a *Axis, style TickLabelStyle, ctx *DrawContext) float64 {
	if style.FontSize > 0 {
		return style.FontSize
	}
	if ctx == nil {
		return 8
	}

	switch {
	case a != nil && (a.Side == AxisLeft || a.Side == AxisRight):
		return ctx.RC.TickLabelSize("y")
	default:
		return ctx.RC.TickLabelSize("x")
	}
}

func tickLabelFontKey(style TickLabelStyle, ctx *DrawContext) string {
	if style.FontKey != "" {
		return style.FontKey
	}
	if ctx == nil {
		return ""
	}
	return ctx.RC.FontKey
}

func tickLabelPadPx(a *Axis, ctx *DrawContext) float64 {
	if a == nil {
		return tickLabelPadForSize(0, TickLabelStyle{}, ctx)
	}
	return tickLabelPadForAxisSize(a, a.TickSize, a.MajorLabelStyle, ctx)
}

func tickLabelPadForSize(tickSize float64, style TickLabelStyle, ctx *DrawContext) float64 {
	return tickLabelPadForAxisSize(nil, tickSize, style, ctx)
}

func tickLabelPadForAxisSize(a *Axis, tickSize float64, style TickLabelStyle, ctx *DrawContext) float64 {
	padPx := defaultTickPadPt * 96.0 / 72.0
	if ctx != nil && ctx.RC.DPI > 0 {
		padPx = defaultTickPadPt * ctx.RC.DPI / 72.0
	}
	if style.Pad > 0 {
		padPx = style.Pad
	}
	return tickSize*tickOutsidePaddingFactor(a) + padPx
}

func tickOutsidePaddingFactor(a *Axis) float64 {
	if a == nil {
		return 1
	}
	switch a.TickDirection {
	case TickDirectionIn:
		return 0
	case TickDirectionInOut:
		return 0.5
	default:
		return 1
	}
}

func tickLabelOrigin(a *Axis, ctx *DrawContext, tickValue float64, layout singleLineTextLayout, labelPadPx float64, style TickLabelStyle, isXAxis bool) (geom.Pt, bool) {
	if a == nil || ctx == nil {
		return geom.Pt{}, false
	}
	style = normalizeTickLabelStyle(style)

	if isXAxis {
		spineY := getSpinePosition(a, ctx)
		tickPos := axisTickDisplayPoint(a, ctx, tickValue, true, spineY)
		hAlign, vAlign := resolvedTickLabelLayoutAlignments(a.Side, style, true)

		switch a.Side {
		case AxisBottom:
			anchor := geom.Pt{X: tickPos.X, Y: tickPos.Y - labelPadPx}
			return geom.Pt{
				X: anchor.X - tickLabelLeftOffset(hAlign, layout),
				Y: anchor.Y + textBaselineOffset(layout, vAlign),
			}, true
		case AxisTop:
			anchor := geom.Pt{X: tickPos.X, Y: tickPos.Y + labelPadPx}
			return geom.Pt{
				X: anchor.X - tickLabelLeftOffset(hAlign, layout),
				Y: anchor.Y + textBaselineOffset(layout, vAlign),
			}, true
		default:
			return geom.Pt{}, false
		}
	}

	spineX := getSpinePosition(a, ctx)
	tickPos := axisTickDisplayPoint(a, ctx, tickValue, false, spineX)
	hAlign, vAlign := resolvedTickLabelLayoutAlignments(a.Side, style, false)

	switch a.Side {
	case AxisLeft:
		anchor := geom.Pt{X: tickPos.X - labelPadPx, Y: tickPos.Y}
		return geom.Pt{
			X: anchor.X - tickLabelLeftOffsetForLeftAxis(hAlign, layout),
			Y: anchor.Y + textBaselineOffset(layout, vAlign),
		}, true
	case AxisRight:
		anchor := geom.Pt{X: tickPos.X + labelPadPx, Y: tickPos.Y}
		return geom.Pt{
			X: anchor.X - tickLabelLeftOffsetForRightAxis(hAlign, layout),
			Y: anchor.Y + textBaselineOffset(layout, vAlign),
		}, true
	default:
		return geom.Pt{}, false
	}
}

func axisTickDisplayPoint(a *Axis, ctx *DrawContext, tickValue float64, isXAxis bool, spineValue float64) geom.Pt {
	if !isXAxis {
		if pt, ok := skewYAxisDisplayPoint(a, ctx, tickValue); ok {
			return pt
		}
		return ctx.DataToPixel.Apply(geom.Pt{X: spineValue, Y: tickValue})
	}
	return ctx.DataToPixel.Apply(geom.Pt{X: tickValue, Y: spineValue})
}

func axisSpinePixelEndpoints(axis *Axis, ctx *DrawContext, px geom.Rect) (geom.Pt, geom.Pt) {
	if axis == nil {
		return geom.Pt{}, geom.Pt{}
	}
	if ctx == nil || axis.SpinePositionMode == AxisSpinePositionBoundary {
		return spinePixelEndpoints(axis.Side, px)
	}

	switch axis.Side {
	case AxisBottom, AxisTop:
		xMin, xMax := ctx.DataToPixel.XScale.Domain()
		y := getSpinePosition(axis, ctx)
		p1 := ctx.DataToPixel.Apply(geom.Pt{X: xMin, Y: y})
		p2 := ctx.DataToPixel.Apply(geom.Pt{X: xMax, Y: y})
		p1.X = math.Round(p1.X) + 0.5
		p2.X = math.Round(p2.X) + 0.5
		p1.Y = math.Round(p1.Y) - 0.5
		p2.Y = math.Round(p2.Y) - 0.5
		return p1, p2
	case AxisLeft, AxisRight:
		yMin, yMax := ctx.DataToPixel.YScale.Domain()
		x := getSpinePosition(axis, ctx)
		p1 := ctx.DataToPixel.Apply(geom.Pt{X: x, Y: yMin})
		p2 := ctx.DataToPixel.Apply(geom.Pt{X: x, Y: yMax})
		p1.X = math.Round(p1.X) + 0.5
		p2.X = math.Round(p2.X) + 0.5
		p1.Y = math.Round(p1.Y) - 0.5
		p2.Y = math.Round(p2.Y) - 0.5
		return p1, p2
	default:
		return geom.Pt{}, geom.Pt{}
	}
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

func textInkRect(origin geom.Pt, layout singleLineTextLayout) (geom.Rect, bool) {
	if layout.HaveInkBounds && layout.InkBounds.W > 0 && layout.InkBounds.H > 0 {
		return geom.Rect{
			Min: geom.Pt{X: origin.X + layout.InkBounds.X, Y: origin.Y - (layout.InkBounds.Y + layout.InkBounds.H)},
			Max: geom.Pt{X: origin.X + layout.InkBounds.X + layout.InkBounds.W, Y: origin.Y - layout.InkBounds.Y},
		}, true
	}
	if layout.Width <= 0 || layout.Height <= 0 {
		return geom.Rect{}, false
	}
	return geom.Rect{
		Min: geom.Pt{X: origin.X, Y: origin.Y - layout.Descent},
		Max: geom.Pt{X: origin.X + layout.Width, Y: origin.Y + layout.Ascent},
	}, true
}

func axisTickLabelBounds(a *Axis, r render.Renderer, ctx *DrawContext) (geom.Rect, bool) {
	if a == nil || ctx == nil {
		return geom.Rect{}, false
	}
	if isPolarProjection(ctx.Projection) {
		return a.polarTickLabelBounds(r, ctx)
	}
	if isGeoProjection(ctx.Projection) {
		return a.geoTickLabelBounds(r, ctx)
	}

	var (
		domainMin float64
		domainMax float64
		isXAxis   bool
	)

	switch a.Side {
	case AxisBottom, AxisTop:
		domainMin, domainMax = ctx.DataToPixel.XScale.Domain()
		isXAxis = true
	case AxisLeft, AxisRight:
		domainMin, domainMax = ctx.DataToPixel.YScale.Domain()
	default:
		return geom.Rect{}, false
	}

	var (
		union geom.Rect
		have  bool
	)

	type labelLevel struct {
		ticks     []float64
		formatter Formatter
		style     TickLabelStyle
		tickSize  float64
	}

	levels := make([]labelLevel, 0, 2+len(a.ExtraTickLevels))
	if a.ShowLabels && a.Locator != nil && a.Formatter != nil {
		levels = append(levels, labelLevel{
			ticks:     visibleTicks(a.Locator.Ticks(domainMin, domainMax, a.majorTickTargetCountForContext(ctx, isXAxis)), domainMin, domainMax),
			formatter: a.Formatter,
			style:     a.MajorLabelStyle,
			tickSize:  a.TickSize,
		})
	}
	if a.ShowMinorLabels && a.MinorLocator != nil && a.MinorFormatter != nil {
		minorLoc := locatorWithMajorContext(a.MinorLocator, a.Locator)
		levels = append(levels, labelLevel{
			ticks:     visibleTicks(minorLoc.Ticks(domainMin, domainMax, a.minorTickTargetCountForContext(ctx, isXAxis)), domainMin, domainMax),
			formatter: a.MinorFormatter,
			style:     a.MinorLabelStyle,
			tickSize:  a.minorTickSize(),
		})
	}
	for _, level := range a.ExtraTickLevels {
		if !level.ShowLabels || level.Locator == nil || level.Formatter == nil {
			continue
		}
		levels = append(levels, labelLevel{
			ticks:     visibleTicks(level.Locator.Ticks(domainMin, domainMax, a.majorTickTargetCountForContext(ctx, isXAxis)), domainMin, domainMax),
			formatter: level.Formatter,
			style:     normalizeTickLabelStyle(level.LabelStyle),
			tickSize:  tickLevelSize(level, a.TickSize),
		})
	}

	for _, level := range levels {
		bounds, ok := tickLabelBoundsForLevel(a, r, ctx, level.ticks, level.formatter, level.style, level.tickSize, isXAxis)
		if !ok {
			continue
		}

		if !have {
			union = bounds
			have = true
			continue
		}
		union = geom.Rect{
			Min: geom.Pt{
				X: math.Min(union.Min.X, bounds.Min.X),
				Y: math.Min(union.Min.Y, bounds.Min.Y),
			},
			Max: geom.Pt{
				X: math.Max(union.Max.X, bounds.Max.X),
				Y: math.Max(union.Max.Y, bounds.Max.Y),
			},
		}
	}

	return union, have
}

func (a *Axis) drawPolarSpine(r render.Renderer, ctx *DrawContext) {
	if a == nil || ctx == nil || !a.ShowSpine {
		return
	}

	center, radius := polarCenterAndRadius(ctx.Clip)
	paint := axisStrokePaint(a, false)

	switch a.Side {
	case AxisBottom, AxisTop:
		path := polarProjectionFramePath(ctx.Projection, ctx.Clip)
		r.Path(path, &paint)
	case AxisLeft, AxisRight:
		labelAngle := polarRadialLabelAngleForProjection(ctx.Projection)
		path := geom.Path{}
		path.MoveTo(center)
		path.LineTo(polarPixelPoint(center, radius, labelAngle))
		r.Path(path, &paint)
	}
}

func (a *Axis) drawPolarTicks(r render.Renderer, ctx *DrawContext) {
	if a == nil || ctx == nil || !a.ShowTicks {
		return
	}

	switch a.Side {
	case AxisBottom, AxisTop:
		a.drawPolarThetaTicks(r, ctx, polarThetaTicks(a, ctx.DataToPixel.XScale, true), a.minorTickSize(), a.minorTickLineWidth())
		a.drawPolarThetaTicks(r, ctx, polarThetaTicks(a, ctx.DataToPixel.XScale, false), a.TickSize, a.tickLineWidth())
	case AxisLeft, AxisRight:
		a.drawPolarRadialTicks(r, ctx, polarRadialTicks(a, ctx.DataToPixel.YScale, true), a.minorTickSize(), a.minorTickLineWidth())
		a.drawPolarRadialTicks(r, ctx, polarRadialTicks(a, ctx.DataToPixel.YScale, false), a.TickSize, a.tickLineWidth())
	}
}

func (a *Axis) drawPolarThetaTicks(r render.Renderer, ctx *DrawContext, ticks []float64, tickSize, lineWidth float64) {
	if len(ticks) == 0 || tickSize <= 0 {
		return
	}
	center, radius := polarCenterAndRadius(ctx.Clip)
	paint := axisStrokePaint(a, true)
	paint.LineWidth = lineWidth

	for _, tick := range ticks {
		angle := polarAngleForTheta(ctx.Projection, ctx.DataToPixel.XScale, tick)
		path := geom.Path{}
		path.MoveTo(polarPixelPoint(center, radius, angle))
		path.LineTo(polarPixelPoint(center, radius+tickSize, angle))
		r.Path(path, &paint)
	}
}

func (a *Axis) drawPolarRadialTicks(r render.Renderer, ctx *DrawContext, ticks []float64, tickSize, lineWidth float64) {
	if len(ticks) == 0 || tickSize <= 0 {
		return
	}
	center, outerRadius := polarCenterAndRadius(ctx.Clip)
	paint := axisStrokePaint(a, true)
	paint.LineWidth = lineWidth
	labelAngle := polarRadialLabelAngleForProjection(ctx.Projection)

	for _, tick := range ticks {
		u := ctx.DataToPixel.YScale.Fwd(tick)
		radius := outerRadius * u
		if radius <= 0 {
			continue
		}
		span := tickSize / math.Max(radius, 1)
		segments := int(math.Max(float64(polarArcMinSegments), math.Ceil(span*24)))
		path := polarArcPath(center, radius, labelAngle-span/2, labelAngle+span/2, segments, false)
		r.Path(path, &paint)
	}
}

func (a *Axis) drawPolarTickLabels(r render.Renderer, ctx *DrawContext) {
	if a == nil || ctx == nil {
		return
	}
	textRen, ok := r.(render.TextDrawer)
	if !ok {
		return
	}

	switch a.Side {
	case AxisBottom, AxisTop:
		if a.ShowLabels {
			a.drawPolarThetaTickLabels(textRen, r, ctx, polarThetaTicks(a, ctx.DataToPixel.XScale, false), a.Formatter, a.MajorLabelStyle, a.TickSize)
		}
		if a.ShowMinorLabels {
			a.drawPolarThetaTickLabels(textRen, r, ctx, polarThetaTicks(a, ctx.DataToPixel.XScale, true), a.MinorFormatter, a.MinorLabelStyle, a.minorTickSize())
		}
	case AxisLeft, AxisRight:
		if a.ShowLabels {
			a.drawPolarRadialTickLabels(textRen, r, ctx, polarRadialTicks(a, ctx.DataToPixel.YScale, false), a.Formatter, a.MajorLabelStyle, a.TickSize)
		}
		if a.ShowMinorLabels {
			a.drawPolarRadialTickLabels(textRen, r, ctx, polarRadialTicks(a, ctx.DataToPixel.YScale, true), a.MinorFormatter, a.MinorLabelStyle, a.minorTickSize())
		}
	}
}

func axisStrokePaint(a *Axis, forTicks bool) render.Paint {
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
		LineWidth: a.LineWidth,
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
	if a.TickLineWidth > 0 {
		return a.TickLineWidth
	}
	return a.LineWidth
}

func (a *Axis) minorTickLineWidth() float64 {
	if a == nil {
		return 0
	}
	if a.MinorTickLineWidth > 0 {
		return a.MinorTickLineWidth
	}
	return a.tickLineWidth()
}

func (a *Axis) drawPolarThetaTickLabels(textRen render.TextDrawer, r render.Renderer, ctx *DrawContext, ticks []float64, formatter Formatter, style TickLabelStyle, tickSize float64) {
	if formatter == nil || len(ticks) == 0 {
		return
	}

	center, radius := polarCenterAndRadius(ctx.Clip)
	style = normalizeTickLabelStyle(style)
	fontSize := tickLabelFontSizeForStyle(a, style, ctx)
	fontKey := tickLabelFontKey(style, ctx)
	labelPadPx := polarThetaTickLabelPadPx(a, tickSize, style, ctx)

	for i, tick := range ticks {
		label := formatTickLabelForTicks(formatter, tick, i, ticks)
		if label == "" {
			continue
		}
		layout := measureSingleLineTextLayout(r, label, fontSize, fontKey, ctx.RC.UseTeX)
		angle := polarAngleForTheta(ctx.Projection, ctx.DataToPixel.XScale, tick)
		anchor := polarPixelPoint(center, radius+labelPadPx, angle)
		drawDisplayText(textRen, label, alignedSingleLineOrigin(anchor, layout, TextAlignCenter, textLayoutVAlignCenter), fontSize, a.tickLabelColor(), fontKey, ctx.RC.UseTeX)
	}
}

func polarThetaTickLabelPadPx(a *Axis, tickSize float64, style TickLabelStyle, ctx *DrawContext) float64 {
	rc := styleOrCurrentRC(ctx)
	padPx := tickSize*tickOutsidePaddingFactor(a) + pointsToPixels(rc, defaultTickPadPt)
	if style.Pad > 0 {
		padPx = tickSize*tickOutsidePaddingFactor(a) + style.Pad
	}
	return padPx + pointsToPixels(rc, 7)
}

func styleOrCurrentRC(ctx *DrawContext) style.RC {
	if ctx != nil {
		return ctx.RC
	}
	return style.CurrentDefaults()
}

func (a *Axis) drawPolarRadialTickLabels(textRen render.TextDrawer, r render.Renderer, ctx *DrawContext, ticks []float64, formatter Formatter, style TickLabelStyle, tickSize float64) {
	if formatter == nil || len(ticks) == 0 {
		return
	}

	center, outerRadius := polarCenterAndRadius(ctx.Clip)
	style = normalizeTickLabelStyle(style)
	fontSize := tickLabelFontSizeForStyle(a, style, ctx)
	fontKey := tickLabelFontKey(style, ctx)
	labelPadPx := tickLabelPadForAxisSize(a, tickSize, style, ctx)
	fullCircle := polarIsFullCircle(ctx.DataToPixel.XScale)
	if fullCircle {
		labelPadPx = 0
	}
	labelAngle := polarRadialLabelAngleForProjection(ctx.Projection)

	for i, tick := range ticks {
		label := formatTickLabelForTicks(formatter, tick, i, ticks)
		if label == "" {
			continue
		}
		layout := measureSingleLineTextLayout(r, label, fontSize, fontKey, ctx.RC.UseTeX)
		radius := outerRadius * ctx.DataToPixel.YScale.Fwd(tick)
		anchor := polarPixelPoint(center, radius+labelPadPx, labelAngle)
		hAlign, vAlign := polarRadialTickLabelAlignments(fullCircle, labelAngle)
		drawDisplayText(textRen, label, alignedSingleLineOrigin(anchor, layout, hAlign, vAlign), fontSize, a.tickLabelColor(), fontKey, ctx.RC.UseTeX)
	}
}

func (a *Axis) polarTickLabelBounds(r render.Renderer, ctx *DrawContext) (geom.Rect, bool) {
	if a == nil || ctx == nil {
		return geom.Rect{}, false
	}

	type polarLabel struct {
		ticks    []float64
		format   Formatter
		style    TickLabelStyle
		tickSize float64
	}

	levels := []polarLabel{}
	if a.ShowLabels {
		levels = append(levels, polarLabel{
			ticks: func() []float64 {
				if a.Side == AxisBottom || a.Side == AxisTop {
					return polarThetaTicks(a, ctx.DataToPixel.XScale, false)
				}
				return polarRadialTicks(a, ctx.DataToPixel.YScale, false)
			}(),
			format:   a.Formatter,
			style:    a.MajorLabelStyle,
			tickSize: a.TickSize,
		})
	}
	if a.ShowMinorLabels {
		levels = append(levels, polarLabel{
			ticks: func() []float64 {
				if a.Side == AxisBottom || a.Side == AxisTop {
					return polarThetaTicks(a, ctx.DataToPixel.XScale, true)
				}
				return polarRadialTicks(a, ctx.DataToPixel.YScale, true)
			}(),
			format:   a.MinorFormatter,
			style:    a.MinorLabelStyle,
			tickSize: a.minorTickSize(),
		})
	}

	var (
		union geom.Rect
		have  bool
	)

	for _, level := range levels {
		if level.format == nil || len(level.ticks) == 0 {
			continue
		}
		bounds, ok := a.polarTickLabelBoundsForLevel(r, ctx, level.ticks, level.format, level.style, level.tickSize)
		if !ok {
			continue
		}
		if !have {
			union = bounds
			have = true
			continue
		}
		union = geom.Rect{
			Min: geom.Pt{X: math.Min(union.Min.X, bounds.Min.X), Y: math.Min(union.Min.Y, bounds.Min.Y)},
			Max: geom.Pt{X: math.Max(union.Max.X, bounds.Max.X), Y: math.Max(union.Max.Y, bounds.Max.Y)},
		}
	}

	return union, have
}

func (a *Axis) polarTickLabelBoundsForLevel(r render.Renderer, ctx *DrawContext, ticks []float64, formatter Formatter, style TickLabelStyle, tickSize float64) (geom.Rect, bool) {
	center, outerRadius := polarCenterAndRadius(ctx.Clip)
	style = normalizeTickLabelStyle(style)
	fontSize := tickLabelFontSizeForStyle(a, style, ctx)
	fontKey := tickLabelFontKey(style, ctx)
	labelPadPx := tickLabelPadForAxisSize(a, tickSize, style, ctx)
	labelAngle := polarRadialLabelAngleForProjection(ctx.Projection)

	var (
		union geom.Rect
		have  bool
	)

	for i, tick := range ticks {
		label := formatTickLabelForTicks(formatter, tick, i, ticks)
		if label == "" {
			continue
		}
		layout := measureSingleLineTextLayout(r, label, fontSize, fontKey, ctx.RC.UseTeX)

		var (
			anchor geom.Pt
			hAlign TextAlign
			vAlign textLayoutVerticalAlign
		)

		if a.Side == AxisBottom || a.Side == AxisTop {
			angle := polarAngleForTheta(ctx.Projection, ctx.DataToPixel.XScale, tick)
			anchor = polarPixelPoint(center, outerRadius+polarThetaTickLabelPadPx(a, tickSize, style, ctx), angle)
			hAlign, vAlign = TextAlignCenter, textLayoutVAlignCenter
		} else {
			radialLabelPadPx := labelPadPx
			fullCircle := polarIsFullCircle(ctx.DataToPixel.XScale)
			if fullCircle {
				radialLabelPadPx = 0
			}
			radius := outerRadius * ctx.DataToPixel.YScale.Fwd(tick)
			anchor = polarPixelPoint(center, radius+radialLabelPadPx, labelAngle)
			hAlign, vAlign = polarRadialTickLabelAlignments(fullCircle, labelAngle)
		}

		origin := alignedSingleLineOrigin(anchor, layout, hAlign, vAlign)
		inkRect, ok := textInkRect(origin, layout)
		if !ok {
			continue
		}
		if !have {
			union = inkRect
			have = true
			continue
		}
		union = geom.Rect{
			Min: geom.Pt{X: math.Min(union.Min.X, inkRect.Min.X), Y: math.Min(union.Min.Y, inkRect.Min.Y)},
			Max: geom.Pt{X: math.Max(union.Max.X, inkRect.Max.X), Y: math.Max(union.Max.Y, inkRect.Max.Y)},
		}
	}

	return union, have
}

func tickLabelBoundsForLevel(a *Axis, r render.Renderer, ctx *DrawContext, ticks []float64, formatter Formatter, style TickLabelStyle, tickSize float64, isXAxis bool) (geom.Rect, bool) {
	if len(ticks) == 0 || formatter == nil {
		return geom.Rect{}, false
	}

	style = normalizeTickLabelStyle(style)
	fontSize := tickLabelFontSizeForStyle(a, style, ctx)
	fontKey := tickLabelFontKey(style, ctx)
	labelPadPx := tickLabelPadForAxisSize(a, tickSize, style, ctx)

	var (
		union geom.Rect
		have  bool
	)

	step := 0.0
	if len(ticks) >= 2 {
		step = ticks[1] - ticks[0]
	}

	for i, tickValue := range ticks {
		label := formatTickLabel(formatter, tickValue, i, ticks)
		if scalarFormatter, ok := formatter.(ScalarFormatter); ok && step > 0 {
			label = formatScalarTickLabel(scalarFormatter, tickValue, step)
		}
		if label == "" {
			continue
		}

		layout := measureSingleLineTextLayout(r, label, fontSize, fontKey, ctx.RC.UseTeX)
		origin, ok := tickLabelOrigin(a, ctx, tickValue, layout, labelPadPx, style, isXAxis)
		if !ok {
			continue
		}
		lineHeight := math.Max(layout.Height, pointsToPixels(ctx.RC, fontSize))
		inkRect, ok := tickLabelDisplayRect(a.Side, style, isXAxis, origin, layout, lineHeight)
		if !ok {
			continue
		}

		if !have {
			union = inkRect
			have = true
			continue
		}
		union = geom.Rect{
			Min: geom.Pt{
				X: math.Min(union.Min.X, inkRect.Min.X),
				Y: math.Min(union.Min.Y, inkRect.Min.Y),
			},
			Max: geom.Pt{
				X: math.Max(union.Max.X, inkRect.Max.X),
				Y: math.Max(union.Max.Y, inkRect.Max.Y),
			},
		}
	}

	return union, have
}

func tickLabelDisplayRect(side AxisSide, style TickLabelStyle, isXAxis bool, origin geom.Pt, layout singleLineTextLayout, lineHeight float64) (geom.Rect, bool) {
	if style.Rotation == 0 {
		hAlign, vAlign := resolvedTickLabelLayoutAlignments(side, style, isXAxis)
		anchor := geom.Pt{
			X: origin.X + textHorizontalOriginOffset(layout, hAlign),
			Y: origin.Y - textBaselineOffset(layout, vAlign),
		}
		return alignedTextLayoutRect(anchor, layout, hAlign, vAlign, lineHeight)
	}

	hAlign, vAlign := resolvedTickLabelLayoutAlignments(side, style, isXAxis)
	angle := style.Rotation * math.Pi / 180.0
	// Faithful rotated extent: the metric box (baseline-left at the matplotlib draw
	// origin, x∈[0,W], y∈[-Descent,Ascent] in y-up) rotated by +angle about the
	// origin — matching what the backend renders.
	o := tickLabelDrawOrigin(origin, layout, hAlign, vAlign, angle, false)
	w := layout.Width
	if w <= 0 && layout.HaveInkBounds {
		w = layout.InkBounds.W
	}
	if w <= 0 {
		return geom.Rect{}, false
	}
	cosT := math.Cos(angle)
	sinT := math.Sin(angle)
	corners := [4][2]float64{{0, -layout.Descent}, {0, layout.Ascent}, {w, layout.Ascent}, {w, -layout.Descent}}
	var out geom.Rect
	for i, c := range corners {
		px := o.X + c[0]*cosT - c[1]*sinT
		py := o.Y + c[0]*sinT + c[1]*cosT
		if i == 0 {
			out = geom.Rect{Min: geom.Pt{X: px, Y: py}, Max: geom.Pt{X: px, Y: py}}
			continue
		}
		out.Min.X = math.Min(out.Min.X, px)
		out.Min.Y = math.Min(out.Min.Y, py)
		out.Max.X = math.Max(out.Max.X, px)
		out.Max.Y = math.Max(out.Max.Y, py)
	}
	return out, true
}

func alignedTextLayoutRect(anchor geom.Pt, layout singleLineTextLayout, hAlign TextAlign, vAlign textLayoutVerticalAlign, lineHeight float64) (geom.Rect, bool) {
	width := layout.Width
	if width <= 0 && layout.HaveInkBounds {
		width = layout.InkBounds.W
	}
	height := lineHeight
	if height <= 0 {
		height = layout.Height
	}
	if height <= 0 && layout.HaveInkBounds {
		height = layout.InkBounds.H
	}
	if width <= 0 || height <= 0 {
		return geom.Rect{}, false
	}

	var minX float64
	switch hAlign {
	case TextAlignLeft:
		minX = anchor.X
	case TextAlignRight:
		minX = anchor.X - width
	default:
		minX = anchor.X - width/2
	}

	var minY float64
	switch vAlign {
	case textLayoutVAlignTop:
		minY = anchor.Y - height
	case textLayoutVAlignBottom:
		minY = anchor.Y
	case textLayoutVAlignCenter, textLayoutVAlignCenterBaseline:
		minY = anchor.Y - height/2
	default:
		minY = anchor.Y - layout.Descent
	}

	return geom.Rect{
		Min: geom.Pt{X: minX, Y: minY},
		Max: geom.Pt{X: minX + width, Y: minY + height},
	}, true
}

func tickLabelCenterOffsetX(layout singleLineTextLayout) float64 {
	return layout.Width / 2
}

func tickLabelLeftOffset(hAlign TextAlign, layout singleLineTextLayout) float64 {
	switch hAlign {
	case TextAlignLeft:
		return 0
	case TextAlignRight:
		return layout.Width
	default:
		return tickLabelCenterOffsetX(layout)
	}
}

func tickLabelLeftOffsetForLeftAxis(hAlign TextAlign, layout singleLineTextLayout) float64 {
	switch hAlign {
	case TextAlignLeft:
		return 0
	case TextAlignCenter:
		return layout.Width / 2
	default:
		return layout.Width
	}
}

func tickLabelLeftOffsetForRightAxis(hAlign TextAlign, layout singleLineTextLayout) float64 {
	switch hAlign {
	case TextAlignRight:
		return layout.Width
	case TextAlignCenter:
		return layout.Width / 2
	default:
		return 0
	}
}

// tickLabelDrawOriginFromP ports matplotlib's Text._get_layout (matplotlib/text.py)
// for a single line and returns the glyph baseline-left draw origin in y-up display
// space, given the text anchor point P, the line metrics, the horizontal/vertical
// alignment, the rotation angle (radians, CCW), and the rotation mode.
//
// matplotlib lays the unrotated metric box at x∈[0,W], y∈[-h,0] (y-up) with the
// baseline-left at (0, -(h-d)). It rotates by M, aligns the (hAlign,vAlign)
// reference to P, and the per-line draw position is M·(0,-(h-d)) minus the
// alignment offset. In rotation_mode="default" the offset comes from the *rotated*
// bounding box; in "anchor" it comes from the *unrotated* box, then rotated by M.
func tickLabelDrawOriginFromP(p geom.Pt, layout singleLineTextLayout, hAlign TextAlign, vAlign textLayoutVerticalAlign, angle float64, anchorMode bool) geom.Pt {
	w := layout.Width
	d := layout.Descent
	h := layout.Ascent + d
	baseline := layout.Ascent // matplotlib's baseline = h - descent = ascent

	cosT := math.Cos(angle)
	sinT := math.Sin(angle)
	rot := func(x, y float64) (float64, float64) {
		return x*cosT - y*sinT, x*sinT + y*cosT
	}

	var offsetX, offsetY float64
	if anchorMode {
		// rotation_mode="anchor": offsets from the UNROTATED box, then rotated.
		switch hAlign {
		case TextAlignRight:
			offsetX = w
		case TextAlignCenter:
			offsetX = w / 2
		default: // left
			offsetX = 0
		}
		switch vAlign {
		case textLayoutVAlignTop:
			offsetY = 0
		case textLayoutVAlignCenter:
			offsetY = -h / 2
		case textLayoutVAlignBaseline:
			offsetY = -baseline
		case textLayoutVAlignCenterBaseline:
			offsetY = -baseline / 2
		default: // bottom
			offsetY = -h
		}
		offsetX, offsetY = rot(offsetX, offsetY)
	} else {
		// rotation_mode="default": offsets from the ROTATED bounding box.
		cornersX := [4]float64{0, 0, w, w}
		cornersY := [4]float64{-h, 0, 0, -h}
		var rxMin, rxMax, ryMin, ryMax float64
		for i := range 4 {
			rx, ry := rot(cornersX[i], cornersY[i])
			if i == 0 || rx < rxMin {
				rxMin = rx
			}
			if i == 0 || rx > rxMax {
				rxMax = rx
			}
			if i == 0 || ry < ryMin {
				ryMin = ry
			}
			if i == 0 || ry > ryMax {
				ryMax = ry
			}
		}
		switch hAlign {
		case TextAlignRight:
			offsetX = rxMax
		case TextAlignCenter:
			offsetX = (rxMin + rxMax) / 2
		default: // left
			offsetX = rxMin
		}
		switch vAlign {
		case textLayoutVAlignTop:
			offsetY = ryMax
		case textLayoutVAlignCenter:
			offsetY = (ryMin + ryMax) / 2
		case textLayoutVAlignBaseline:
			offsetY = ryMin + d
		case textLayoutVAlignCenterBaseline:
			offsetY = ryMin + (ryMax - ryMin) - baseline/2
		default: // bottom
			offsetY = ryMin
		}
	}

	// matplotlib: draw_origin = P + M·(0, -(h-d)) - (offsetX, offsetY)
	blX, blY := rot(0, -(h - d))
	return geom.Pt{
		X: p.X + blX - offsetX,
		Y: p.Y + blY - offsetY,
	}
}

// tickLabelDrawOrigin recovers matplotlib's text anchor point P from the unrotated
// draw origin (undoing the alignment) and returns the matplotlib baseline-left draw
// origin via tickLabelDrawOriginFromP.
func tickLabelDrawOrigin(origin geom.Pt, layout singleLineTextLayout, hAlign TextAlign, vAlign textLayoutVerticalAlign, angle float64, anchorMode bool) geom.Pt {
	p := geom.Pt{
		X: origin.X + textHorizontalOriginOffset(layout, hAlign),
		Y: origin.Y - textBaselineOffset(layout, vAlign),
	}
	return tickLabelDrawOriginFromP(p, layout, hAlign, vAlign, angle, anchorMode)
}

// rotatedTextBackendAnchorFromP maps matplotlib's baseline-left draw origin O to the
// bottom-center anchor the AGG backend rotates about. The backend renders the
// baseline-left at anchor - R(angle)·(W/2, Descent) (proven from drawTextRotatedDirect
// + the y-down device flip), and its metrics.W/Descent equal layout.Width/Descent, so
// anchor = O + R(angle)·(W/2, Descent) makes the rendered glyphs land exactly on O.
func rotatedTextBackendAnchorFromP(p geom.Pt, layout singleLineTextLayout, hAlign TextAlign, vAlign textLayoutVerticalAlign, angle float64, anchorMode bool) geom.Pt {
	o := tickLabelDrawOriginFromP(p, layout, hAlign, vAlign, angle, anchorMode)
	cosT := math.Cos(angle)
	sinT := math.Sin(angle)
	w := layout.Width
	d := layout.Descent
	return geom.Pt{
		X: o.X + (w/2*cosT - d*sinT),
		Y: o.Y + (w/2*sinT + d*cosT),
	}
}

// tickLabelRotationAnchor returns the AGG backend rotation pivot for a tick label
// drawn with matplotlib's rotation_mode="default".
func tickLabelRotationAnchor(origin geom.Pt, layout singleLineTextLayout, hAlign TextAlign, vAlign textLayoutVerticalAlign, angle float64) geom.Pt {
	p := geom.Pt{
		X: origin.X + textHorizontalOriginOffset(layout, hAlign),
		Y: origin.Y - textBaselineOffset(layout, vAlign),
	}
	return rotatedTextBackendAnchorFromP(p, layout, hAlign, vAlign, angle, false)
}

func resolvedTickLabelAlignments(side AxisSide, style TickLabelStyle, isXAxis bool) (TextAlign, TextVerticalAlign) {
	if !style.AutoAlign {
		return style.HAlign, style.VAlign
	}
	if isXAxis {
		switch side {
		case AxisBottom:
			return TextAlignCenter, TextVAlignTop
		case AxisTop:
			return TextAlignCenter, TextVAlignBottom
		}
	}
	switch side {
	case AxisLeft:
		return TextAlignRight, TextVAlignMiddle
	case AxisRight:
		return TextAlignLeft, TextVAlignMiddle
	default:
		return TextAlignCenter, TextVAlignMiddle
	}
}

func resolvedTickLabelLayoutAlignments(side AxisSide, style TickLabelStyle, isXAxis bool) (TextAlign, textLayoutVerticalAlign) {
	hAlign, vAlign := resolvedTickLabelAlignments(side, style, isXAxis)
	if isXAxis {
		return hAlign, layoutVerticalAlign(vAlign, false)
	}
	return hAlign, layoutVerticalAlign(vAlign, true)
}

func defaultTickLabelStyle() TickLabelStyle {
	return TickLabelStyle{AutoAlign: true}
}

func normalizeTickLabelStyle(style TickLabelStyle) TickLabelStyle {
	if !style.AutoAlign && style.HAlign == TextAlignLeft && style.VAlign == TextVAlignBaseline && style.Pad == 0 && style.Rotation == 0 {
		style.AutoAlign = true
	}
	return style
}

func tickLevelSize(level TickLevel, fallback float64) float64 {
	if level.Size > 0 {
		return level.Size
	}
	return fallback
}

func (a *Axis) minorTickSize() float64 {
	if a.MinorTickSize > 0 {
		return a.MinorTickSize
	}
	return a.TickSize * 0.6
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
