package core

import (
	"fmt"
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

func (a *Axes) SetTitle(title string) { a.Title = title }

// SetTitleLocation aligns the axes title to the left, center, or right axes
// edge. It overrides the axes.titlelocation value captured at axes creation.
func (a *Axes) SetTitleLocation(location string) error {
	if a == nil {
		return nil
	}
	a.ensureRCTextDefaults()
	switch strings.ToLower(strings.TrimSpace(location)) {
	case "left":
		a.titleLocation = "left"
	case "", "center":
		a.titleLocation = "center"
	case "right":
		a.titleLocation = "right"
	default:
		return fmt.Errorf("unsupported title location %q", location)
	}
	return nil
}

// SetTitleY fixes the title at an axes-relative vertical coordinate and
// disables automatic lifting above top-axis decorations.
func (a *Axes) SetTitleY(y float64) {
	if a == nil {
		return
	}
	a.ensureRCTextDefaults()
	a.titleY = y
	a.titleYSet = true
}

// SetTitleAutoY restores automatic title placement above top-axis decorations.
func (a *Axes) SetTitleAutoY() {
	if a == nil {
		return
	}
	a.ensureRCTextDefaults()
	a.titleY = 0
	a.titleYSet = false
}

// SetTitlePad sets the display-space title padding in points.
func (a *Axes) SetTitlePad(pad float64) {
	if a != nil {
		a.ensureRCTextDefaults()
		a.titlePadPt = pad
	}
}

// SetTitleWeight sets the numeric title font weight (400 is normal, 700 bold).
func (a *Axes) SetTitleWeight(weight int) {
	if a != nil {
		a.ensureRCTextDefaults()
		a.titleWeight = normalizedAxesFontWeight(weight)
	}
}

func (f *Figure) SetSuptitle(title string) {
	if f != nil {
		f.SupTitle = title
	}
}

func (f *Figure) SetSupTitle(title string) { f.SetSuptitle(title) }

func (f *Figure) SetSupxlabel(label string) {
	if f != nil {
		f.SupXLabel = label
	}
}

func (f *Figure) SetSupXLabel(label string) { f.SetSupxlabel(label) }

func (f *Figure) SetSupylabel(label string) {
	if f != nil {
		f.SupYLabel = label
	}
}

func (f *Figure) SetSupYLabel(label string) { f.SetSupylabel(label) }

func (a *Axes) SetXLabel(label string) { a.XLabel = label }

func (a *Axes) SetYLabel(label string) { a.YLabel = label }

// SetXLabelPad sets the x-axis label padding in points.
func (a *Axes) SetXLabelPad(pad float64) {
	if a != nil {
		a.ensureRCTextDefaults()
		a.xLabelPadPt = pad
	}
}

// SetYLabelPad sets the y-axis label padding in points.
func (a *Axes) SetYLabelPad(pad float64) {
	if a != nil {
		a.ensureRCTextDefaults()
		a.yLabelPadPt = pad
	}
}

// SetXLabelWeight sets the numeric x-axis label font weight.
func (a *Axes) SetXLabelWeight(weight int) {
	if a != nil {
		a.ensureRCTextDefaults()
		a.xLabelWeight = normalizedAxesFontWeight(weight)
	}
}

// SetYLabelWeight sets the numeric y-axis label font weight.
func (a *Axes) SetYLabelWeight(weight int) {
	if a != nil {
		a.ensureRCTextDefaults()
		a.yLabelWeight = normalizedAxesFontWeight(weight)
	}
}

func (a *Axes) SetXLabelPosition(position string) error {
	if a == nil {
		return nil
	}
	switch strings.ToLower(strings.TrimSpace(position)) {
	case "", "bottom":
		a.xLabelSide = AxisBottom
	case "top":
		a.xLabelSide = AxisTop
	default:
		return fmt.Errorf("unsupported x label position %q", position)
	}
	return nil
}

func (a *Axes) SetYLabelPosition(position string) error {
	if a == nil {
		return nil
	}
	switch strings.ToLower(strings.TrimSpace(position)) {
	case "", "left":
		a.yLabelSide = AxisLeft
	case "right":
		a.yLabelSide = AxisRight
	default:
		return fmt.Errorf("unsupported y label position %q", position)
	}
	return nil
}

func drawAxesLabels(ax *Axes, r render.Renderer, ctx *DrawContext, px geom.Rect, alignment figureTextAlignment) {
	textRen, ok := r.(render.TextDrawer)
	if !ok {
		return
	}

	titleColor := ctx.RC.DefaultAxesTitleColor()
	labelColor := ctx.RC.DefaultAxesLabelColor()
	titleSize := titleFontSize(ctx)
	labelSize := axisLabelFontSize(ctx)

	// Title: aligned at the configured axes edge and either automatically
	// lifted above top decorations or fixed at an axes-relative y coordinate.
	if ax.Title != "" {
		fontKey := axesTitleFontKey(ax, ctx)
		hAlign := axesTitleAlignment(ax)
		layout := measureSingleLineTextLayout(r, ax.Title, titleSize, fontKey, ctx.RC.UseTeX)
		drawDisplayText(
			textRen,
			ax.Title,
			alignedSingleLineOrigin(titleAnchorPoint(ax, r, ctx, px, alignment), layout, hAlign, textLayoutVAlignBaseline),
			titleSize,
			titleColor,
			fontKey,
			ctx.RC.UseTeX,
		)
	}

	// XLabel: centered relative to the selected axis side and padded from the
	// union of the spine and visible tick-label bounds, matching Matplotlib's
	// default label placement model.
	if ax.XLabel != "" && !ax.hideXLabel && ax.ProjectionName() != "3d" {
		side := ax.effectiveXLabelSide()
		fontKey := xAxisLabelFontKey(ax, ctx)
		layout := measureSingleLineTextLayout(r, ax.XLabel, labelSize, fontKey, ctx.RC.UseTeX)
		anchor, vAlign := xLabelAnchorPoint(ax, r, ctx, px, side, alignment)
		drawDisplayText(
			textRen,
			ax.XLabel,
			alignedSingleLineOrigin(anchor, layout, TextAlignCenter, vAlign),
			labelSize,
			labelColor,
			fontKey,
			ctx.RC.UseTeX,
		)
	}

	// YLabel: vertical text if supported, else horizontal fallback
	if ax.YLabel != "" && !ax.hideYLabel && ax.ProjectionName() != "3d" {
		side := ax.effectiveYLabelSide()
		fontKey := yAxisLabelFontKey(ax, ctx)
		anchor := yLabelAnchorPoint(ax, r, ctx, px, side, alignment)
		angle := -math.Pi / 2
		if side == AxisRight {
			angle = math.Pi / 2
		}
		switch ren := r.(type) {
		case render.RotatedTextDrawer:
			// matplotlib draws BOTH the left and right y-axis labels at
			// rotation=90 (reading bottom-to-top) with rotation_mode="anchor",
			// ha="center", and va="bottom" (left) / va="top" (right). The left
			// label is NOT mirrored — both sides share the +90° orientation.
			labelAngle := math.Pi / 2
			yLabelLayout := measureSingleLineTextLayout(r, ax.YLabel, labelSize, fontKey, ctx.RC.UseTeX)
			yLabelVAlign := textLayoutVAlignBottom
			if side == AxisRight {
				yLabelVAlign = textLayoutVAlignTop
			}
			backendAnchor := rotatedTextBackendAnchorFromP(anchor, yLabelLayout, TextAlignCenter, yLabelVAlign, labelAngle, true)
			drawDisplayTextRotated(ren, ax.YLabel, backendAnchor, labelSize, labelAngle, labelColor, fontKey, ctx.RC.UseTeX)
		case render.VerticalTextDrawer:
			if angle < 0 {
				layout := measureSingleLineTextLayout(r, ax.YLabel, labelSize, fontKey, ctx.RC.UseTeX)
				drawDisplayText(
					textRen,
					ax.YLabel,
					alignedSingleLineOrigin(geom.Pt{X: anchor.X, Y: px.Min.Y + px.H()/2}, layout, TextAlignCenter, textLayoutVAlignCenter),
					labelSize,
					labelColor,
					fontKey,
					ctx.RC.UseTeX,
				)
			} else {
				drawDisplayTextVertical(ren, ax.YLabel, geom.Pt{X: anchor.X, Y: anchor.Y}, labelSize, labelColor, fontKey)
			}
		default:
			layout := measureSingleLineTextLayout(r, ax.YLabel, labelSize, fontKey, ctx.RC.UseTeX)
			drawDisplayText(
				textRen,
				ax.YLabel,
				alignedSingleLineOrigin(geom.Pt{X: anchor.X, Y: px.Min.Y + px.H()/2}, layout, TextAlignCenter, textLayoutVAlignCenter),
				labelSize,
				labelColor,
				fontKey,
				ctx.RC.UseTeX,
			)
		}
	}
}

func titleFontSize(ctx *DrawContext) float64 {
	if ctx == nil {
		return 12
	}
	return ctx.RC.TitleSize()
}

func axisLabelFontSize(ctx *DrawContext) float64 {
	if ctx == nil {
		return 8
	}
	return ctx.RC.AxisLabelSize()
}

func figureLabelFontSize(ctx *DrawContext) float64 {
	if ctx == nil {
		return 12
	}
	return ctx.RC.TitleSize()
}

func titleAnchorPoint(ax *Axes, r render.Renderer, ctx *DrawContext, px geom.Rect, alignment figureTextAlignment) geom.Pt {
	titlePadPx := pointsToPixels(ctx.RC, axesTitlePadPt(ax, ctx))
	topExtent := 0.0
	if ax != nil && ax.titleYSet {
		topExtent = ctx.TransAxes().Apply(geom.Pt{Y: ax.titleY}).Y
	} else {
		topExtent = titleTopExtent(ax, r, ctx, px)
		if aligned, ok := alignment.titleExtents[alignmentKey(AxisTop, spinePixelY(AxisTop, px))]; ok {
			topExtent = aligned
		}
	}
	y := topExtent + titlePadPx
	if isPolarProjection(ctx.Projection) && !isRadarProjection(ctx.Projection) {
		y = math.Ceil(y)
	}
	return geom.Pt{
		X: ctx.TransAxes().Apply(geom.Pt{X: axesTitleX(ax), Y: 1}).X,
		Y: y,
	}
}

func xLabelAnchorPoint(ax *Axes, r render.Renderer, ctx *DrawContext, px geom.Rect, side AxisSide, alignment figureTextAlignment) (geom.Pt, textLayoutVerticalAlign) {
	anchor := ctx.TransAxes().Apply(geom.Pt{X: 0.5, Y: 0})

	if isGeoProjection(ax.projection) {
		if side == AxisTop {
			anchor.Y = px.Max.Y + xAxisLabelPadPx(ax, ctx)
			return anchor, textLayoutVAlignBaseline
		}
		anchor.Y = px.Min.Y - xAxisLabelPadPx(ax, ctx)
		return anchor, textLayoutVAlignTop
	}

	xAxis := ax.axisForXLabelSide(side)
	if side == AxisTop {
		topExtent := xAxisSpinePixelY(xAxis, ctx, AxisTop, px)
		if xAxis != nil {
			if tickBounds, ok := xLabelTickLabelBounds(ax, xAxis, r, ctx); ok {
				topExtent = math.Max(topExtent, tickBounds.Max.Y)
			} else if xAxis.ShowTicks {
				topExtent += xAxis.TickSize * tickOutsidePaddingFactor(xAxis)
			}
		}
		if aligned, ok := alignment.xLabelExtents[alignmentKey(side, xLabelSpinePixelY(side, px))]; ok {
			topExtent = aligned
		}
		anchor.Y = topExtent + xAxisLabelPadPx(ax, ctx)
		return anchor, textLayoutVAlignBaseline
	}

	bottomExtent := xAxisSpinePixelY(xAxis, ctx, AxisBottom, px)
	if xAxis != nil {
		if tickBounds, ok := xLabelTickLabelBounds(ax, xAxis, r, ctx); ok {
			bottomExtent = math.Min(bottomExtent, tickBounds.Min.Y)
		} else if xAxis.ShowTicks {
			bottomExtent -= xAxis.TickSize * tickOutsidePaddingFactor(xAxis)
		}
	}
	if aligned, ok := alignment.xLabelExtents[alignmentKey(side, xLabelSpinePixelY(side, px))]; ok {
		bottomExtent = aligned
	}
	anchor.Y = bottomExtent - xAxisLabelPadPx(ax, ctx)
	return anchor, textLayoutVAlignTop
}

func xLabelTickLabelBounds(ax *Axes, xAxis *Axis, r render.Renderer, ctx *DrawContext) (geom.Rect, bool) {
	if ax != nil && isPolarProjection(ax.projection) {
		return xAxis.polarThetaTickLabelWindowBounds(r, ctx)
	}
	return axisTickLabelBounds(xAxis, r, ctx)
}

func yLabelAnchorPoint(ax *Axes, r render.Renderer, ctx *DrawContext, px geom.Rect, side AxisSide, alignment figureTextAlignment) geom.Pt {
	anchor := ctx.TransAxes().Apply(geom.Pt{X: 0, Y: 0.5})
	anchor.X = spinePixelX(AxisLeft, px) - yAxisLabelPadPx(ax, ctx)

	yAxis := ax.axisForYLabelSide(side)
	if side == AxisRight {
		spineX := yAxisSpinePixelX(yAxis, ctx, AxisRight, px)
		rightExtent := spineX
		if tickBounds, ok := axisTickLabelBounds(yAxis, r, ctx); ok {
			rightExtent = math.Max(rightExtent, tickBounds.Max.X)
		} else if yAxis != nil && yAxis.ShowTicks {
			rightExtent += tickLabelPadPx(yAxis, ctx)
		}
		if aligned, ok := alignment.yLabelExtents[alignmentKey(side, spinePixelX(side, px))]; ok {
			rightExtent = aligned
		}
		anchor.X = rightExtent + yAxisLabelPadPx(ax, ctx)
		return anchor
	}

	spineX := yAxisSpinePixelX(yAxis, ctx, AxisLeft, px)
	leftExtent := spineX
	if tickBounds, ok := axisTickLabelBounds(yAxis, r, ctx); ok {
		leftExtent = math.Min(leftExtent, tickBounds.Min.X)
	} else if yAxis != nil && yAxis.ShowTicks {
		leftExtent -= tickLabelPadPx(yAxis, ctx)
	}
	if aligned, ok := alignment.yLabelExtents[alignmentKey(side, spinePixelX(side, px))]; ok {
		leftExtent = aligned
	}
	anchor.X = leftExtent - yAxisLabelPadPx(ax, ctx)
	return anchor
}

func xAxisSpinePixelY(axis *Axis, ctx *DrawContext, side AxisSide, px geom.Rect) float64 {
	if axis != nil && axis.SpinePositionMode != AxisSpinePositionBoundary {
		if position, ok := axisSpineDisplayCoordinate(axis, ctx); ok {
			return position
		}
	}
	return xLabelSpinePixelY(side, px)
}

func yAxisSpinePixelX(axis *Axis, ctx *DrawContext, side AxisSide, px geom.Rect) float64 {
	if axis != nil && axis.SpinePositionMode != AxisSpinePositionBoundary {
		if position, ok := axisSpineDisplayCoordinate(axis, ctx); ok {
			return position
		}
	}
	return spinePixelX(side, px)
}

func axisLabelPadPx(ctx *DrawContext) float64 {
	if ctx == nil {
		rc := style.CurrentDefaults()
		return pointsToPixels(rc, rc.Axes.LabelPad)
	}
	return pointsToPixels(ctx.RC, ctx.RC.Axes.LabelPad)
}

func xAxisLabelPadPx(ax *Axes, ctx *DrawContext) float64 {
	if ax == nil || !ax.textDefaultsSet {
		return axisLabelPadPx(ctx)
	}
	return pointsToPixels(ctx.RC, ax.xLabelPadPt)
}

func yAxisLabelPadPx(ax *Axes, ctx *DrawContext) float64 {
	if ax == nil || !ax.textDefaultsSet {
		return axisLabelPadPx(ctx)
	}
	return pointsToPixels(ctx.RC, ax.yLabelPadPt)
}

func (a *Axes) applyRCTextDefaults(rc *style.RC) {
	if a == nil || rc == nil {
		return
	}
	a.titleLocation = normalizedAxesTitleLocation(rc.Axes.TitleLocation)
	a.titleY = rc.Axes.TitleY
	a.titleYSet = rc.Axes.TitleYSet
	a.titlePadPt = rc.Axes.TitlePad
	a.titleWeight = normalizedAxesFontWeight(rc.Axes.TitleWeight)
	a.xLabelPadPt = rc.Axes.LabelPad
	a.yLabelPadPt = rc.Axes.LabelPad
	a.xLabelWeight = normalizedAxesFontWeight(rc.Axes.LabelWeight)
	a.yLabelWeight = normalizedAxesFontWeight(rc.Axes.LabelWeight)
	a.textDefaultsSet = true
}

func (a *Axes) ensureRCTextDefaults() {
	if a == nil || a.textDefaultsSet {
		return
	}
	rc := a.resolvedRC()
	a.applyRCTextDefaults(&rc)
}

func normalizedAxesTitleLocation(location string) string {
	switch strings.ToLower(strings.TrimSpace(location)) {
	case "left", "right":
		return strings.ToLower(strings.TrimSpace(location))
	default:
		return "center"
	}
}

func normalizedAxesFontWeight(weight int) int {
	if weight <= 0 {
		return 400
	}
	return weight
}

func axesTitleX(ax *Axes) float64 {
	if ax == nil {
		return 0.5
	}
	switch ax.titleLocation {
	case "left":
		return 0
	case "right":
		return 1
	default:
		return 0.5
	}
}

func axesTitleAlignment(ax *Axes) TextAlign {
	if ax == nil {
		return TextAlignCenter
	}
	switch ax.titleLocation {
	case "left":
		return TextAlignLeft
	case "right":
		return TextAlignRight
	default:
		return TextAlignCenter
	}
}

func axesTitlePadPt(ax *Axes, ctx *DrawContext) float64 {
	if ax != nil && ax.textDefaultsSet {
		return ax.titlePadPt
	}
	if ctx != nil {
		return ctx.RC.Axes.TitlePad
	}
	return style.CurrentDefaults().Axes.TitlePad
}

func axesTitleFontKey(ax *Axes, ctx *DrawContext) string {
	if ctx == nil {
		return ""
	}
	weight := ctx.RC.Axes.TitleWeight
	if ax != nil && ax.textDefaultsSet {
		weight = ax.titleWeight
	}
	return fontKeyWithWeight(ctx.RC.FontKey, weight)
}

func xAxisLabelFontKey(ax *Axes, ctx *DrawContext) string {
	if ctx == nil {
		return ""
	}
	weight := ctx.RC.Axes.LabelWeight
	if ax != nil && ax.textDefaultsSet {
		weight = ax.xLabelWeight
	}
	return fontKeyWithWeight(ctx.RC.FontKey, weight)
}

func yAxisLabelFontKey(ax *Axes, ctx *DrawContext) string {
	if ctx == nil {
		return ""
	}
	weight := ctx.RC.Axes.LabelWeight
	if ax != nil && ax.textDefaultsSet {
		weight = ax.yLabelWeight
	}
	return fontKeyWithWeight(ctx.RC.FontKey, weight)
}

func fontKeyWithWeight(fontKey string, weight int) string {
	weight = normalizedAxesFontWeight(weight)
	props := render.ParseFontProperties(fontKey)
	if props.Weight == weight {
		return fontKey
	}
	props.Weight = weight
	return render.FontPropertiesKey(props)
}

func (a *Axes) effectiveXLabelSide() AxisSide {
	if a == nil {
		return AxisBottom
	}
	if a.xLabelSide == AxisTop {
		return AxisTop
	}
	return AxisBottom
}

func (a *Axes) effectiveYLabelSide() AxisSide {
	if a == nil {
		return AxisLeft
	}
	if a.yLabelSide == AxisRight {
		return AxisRight
	}
	return AxisLeft
}

func (a *Axes) axisForXLabelSide(side AxisSide) *Axis {
	if a == nil {
		return nil
	}
	if side == AxisTop {
		return a.XAxisTop
	}
	if a.XAxis != nil {
		return a.XAxis
	}
	return a.XAxisTop
}

func (a *Axes) axisForYLabelSide(side AxisSide) *Axis {
	if a == nil {
		return nil
	}
	if side == AxisRight {
		if a.YAxisRight != nil {
			return a.YAxisRight
		}
		return nil
	}
	if a.YAxis != nil {
		return a.YAxis
	}
	return a.YAxisRight
}

func spinePixelX(side AxisSide, px geom.Rect) float64 {
	p1, _ := spinePixelEndpoints(side, px)
	return p1.X
}

func spinePixelY(side AxisSide, px geom.Rect) float64 {
	p1, _ := spinePixelEndpoints(side, px)
	return p1.Y
}

func xLabelSpinePixelY(side AxisSide, px geom.Rect) float64 {
	switch side {
	case AxisBottom:
		return math.Round(px.Min.Y) - 0.5
	case AxisTop:
		return math.Round(px.Max.Y) - 0.5
	default:
		return spinePixelY(side, px)
	}
}
