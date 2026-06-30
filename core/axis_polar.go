package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func (a *Axis) drawPolarSpine(r render.Renderer, ctx *DrawContext) {
	if a == nil || ctx == nil || !a.ShowSpine {
		return
	}

	center, radius := polarCenterAndRadius(ctx.Clip)
	paint := axisStrokePaint(a, ctx, false)

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
	paint := axisStrokePaint(a, ctx, true)
	paint.LineWidth = pointsToPixels(ctx.RC, lineWidth)

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
	paint := axisStrokePaint(a, ctx, true)
	paint.LineWidth = pointsToPixels(ctx.RC, lineWidth)
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

func (a *Axis) drawPolarThetaTickLabels(textRen render.TextDrawer, r render.Renderer, ctx *DrawContext, ticks []float64, formatter Formatter, style TickLabelStyle, tickSize float64) {
	if formatter == nil || len(ticks) == 0 {
		return
	}

	center, radius := polarCenterAndRadius(ctx.Clip)
	style = normalizeTickLabelStyle(style)
	fontSize := tickLabelFontSizeForStyle(a, style, ctx)
	fontKey := tickLabelFontKey(style, ctx)
	labelPadPx := polarThetaTickLabelPadPx(a, tickSize, style, ctx)

	format := tickLabelFormatter(formatter, ticks)
	for i, tick := range ticks {
		label := format(tick, i)
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

	format := tickLabelFormatter(formatter, ticks)
	for i, tick := range ticks {
		label := format(tick, i)
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

	format := tickLabelFormatter(formatter, ticks)
	for i, tick := range ticks {
		label := format(tick, i)
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

func (a *Axis) polarThetaTickLabelWindowBounds(r render.Renderer, ctx *DrawContext) (geom.Rect, bool) {
	if a == nil || ctx == nil || ctx.DataToPixel.XScale == nil {
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
			ticks:    polarThetaTicks(a, ctx.DataToPixel.XScale, false),
			format:   a.Formatter,
			style:    a.MajorLabelStyle,
			tickSize: a.TickSize,
		})
	}
	if a.ShowMinorLabels {
		levels = append(levels, polarLabel{
			ticks:    polarThetaTicks(a, ctx.DataToPixel.XScale, true),
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
		bounds, ok := a.polarThetaTickLabelWindowBoundsForLevel(r, ctx, level.ticks, level.format, level.style, level.tickSize)
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

func (a *Axis) polarThetaTickLabelWindowBoundsForLevel(r render.Renderer, ctx *DrawContext, ticks []float64, formatter Formatter, style TickLabelStyle, tickSize float64) (geom.Rect, bool) {
	if formatter == nil || len(ticks) == 0 {
		return geom.Rect{}, false
	}

	center, outerRadius := polarCenterAndRadius(ctx.Clip)
	style = normalizeTickLabelStyle(style)
	fontSize := tickLabelFontSizeForStyle(a, style, ctx)
	fontKey := tickLabelFontKey(style, ctx)
	lineHeightMin := pointsToPixels(ctx.RC, fontSize)

	var (
		union geom.Rect
		have  bool
	)
	format := tickLabelFormatter(formatter, ticks)
	for i, tick := range ticks {
		label := format(tick, i)
		if label == "" {
			continue
		}
		layout := measureSingleLineTextLayout(r, label, fontSize, fontKey, ctx.RC.UseTeX)
		angle := polarAngleForTheta(ctx.Projection, ctx.DataToPixel.XScale, tick)
		anchor := polarPixelPoint(center, outerRadius+polarThetaTickLabelPadPx(a, tickSize, style, ctx), angle)
		lineHeight := math.Max(layout.Height, lineHeightMin)
		bounds, ok := alignedTextLayoutRect(anchor, layout, TextAlignCenter, textLayoutVAlignCenter, lineHeight)
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
