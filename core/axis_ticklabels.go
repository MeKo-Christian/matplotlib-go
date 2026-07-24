package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

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

	scalarFormatter, hasScalar := formatter.(ScalarFormatter)
	var scalarCtx scalarTickContext
	if hasScalar {
		scalarCtx = newScalarTickContext(scalarFormatter, ticks)
	}

	var rotRen render.RotatedTextDrawer
	if drawer, ok := r.(render.RotatedTextDrawer); ok {
		rotRen = drawer
	}

	for i, tickValue := range ticks {
		label := formatTickLabel(formatter, tickValue, i, ticks)
		if hasScalar {
			label = formatScalarTickLabelCtx(scalarFormatter, tickValue, scalarCtx)
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
	if !ok || formatter == nil {
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
	layout := measureSingleLineTextLayout(r, label, fontSize, fontKey, ctx.RC.UseTeX)
	offsetPadPx := pointsToPixels(styleOrCurrentRC(ctx), offsetTextPadPt)
	labelBounds, haveLabelBounds := tickLabelBoundsForLevel(a, r, ctx, ticks, formatter, style, tickSize, isXAxis)

	var anchor geom.Pt
	hAlign := TextAlignRight
	vAlign := textLayoutVAlignTop
	switch a.Side {
	case AxisBottom:
		bottom := ctx.Clip.Min.Y
		if haveLabelBounds {
			bottom = labelBounds.Min.Y
		}
		anchor = geom.Pt{X: ctx.Clip.Max.X, Y: bottom - offsetPadPx}
	case AxisTop:
		top := ctx.Clip.Max.Y
		if haveLabelBounds {
			top = labelBounds.Max.Y
		}
		anchor = geom.Pt{X: ctx.Clip.Max.X, Y: top + offsetPadPx}
		vAlign = textLayoutVAlignBottom
	case AxisLeft:
		// Matplotlib places the y-axis offset text above the top of the axis,
		// left-aligned to the left spine.
		anchor = geom.Pt{X: ctx.Clip.Min.X, Y: ctx.Clip.Max.Y + offsetPadPx}
		hAlign = TextAlignLeft
		vAlign = textLayoutVAlignBottom
	case AxisRight:
		anchor = geom.Pt{X: ctx.Clip.Max.X, Y: ctx.Clip.Max.Y + offsetPadPx}
		hAlign = TextAlignRight
		vAlign = textLayoutVAlignBottom
	default:
		return
	}
	origin := geom.Pt{
		X: anchor.X - textHorizontalOriginOffset(layout, hAlign),
		Y: anchor.Y + textBaselineOffset(layout, vAlign),
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
	padPt := defaultTickPadPt
	if style.padPtSet || style.PadPt > 0 {
		padPt = style.PadPt
	}
	// No-context fallback assumes matplotlib's default figure DPI of 100.
	padPx := padPt * 100.0 / 72.0
	if ctx != nil && ctx.RC.DPI > 0 {
		padPx = padPt * ctx.RC.DPI / 72.0
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

func styleOrCurrentRC(ctx *DrawContext) style.RC {
	if ctx != nil {
		return ctx.RC
	}
	return style.CurrentDefaults()
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

	scalarFormatter, hasScalar := formatter.(ScalarFormatter)
	var scalarCtx scalarTickContext
	if hasScalar {
		scalarCtx = newScalarTickContext(scalarFormatter, ticks)
	}

	for i, tickValue := range ticks {
		label := formatTickLabel(formatter, tickValue, i, ticks)
		if hasScalar {
			label = formatScalarTickLabelCtx(scalarFormatter, tickValue, scalarCtx)
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

// defaultMinorTickLabelStyle carries matplotlib's minor tick pad
// (xtick.minor.pad / ytick.minor.pad = 3.4 pt vs the major 3.5 pt).
func defaultMinorTickLabelStyle() TickLabelStyle {
	return TickLabelStyle{AutoAlign: true, PadPt: defaultMinorTickPadPt}
}

func normalizeTickLabelStyle(style TickLabelStyle) TickLabelStyle {
	if !style.AutoAlign && style.HAlign == TextAlignLeft && style.VAlign == TextVAlignBaseline && style.Pad == 0 && style.Rotation == 0 {
		style.AutoAlign = true
	}
	return style
}
