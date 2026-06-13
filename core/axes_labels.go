package core

import (
	"fmt"
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

func (a *Axes) SetTitle(title string) { a.Title = title }

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

	// Title: centered above the plot
	if ax.Title != "" {
		layout := measureSingleLineTextLayout(r, ax.Title, titleSize, ctx.RC.FontKey, ctx.RC.UseTeX)
		drawDisplayText(
			textRen,
			ax.Title,
			alignedSingleLineOrigin(titleAnchorPoint(ax, r, ctx, px, alignment), layout, TextAlignCenter, textLayoutVAlignBaseline),
			titleSize,
			titleColor,
			ctx.RC.FontKey,
			ctx.RC.UseTeX,
		)
	}

	// XLabel: centered relative to the selected axis side and padded from the
	// union of the spine and visible tick-label bounds, matching Matplotlib's
	// default label placement model.
	if ax.XLabel != "" && ax.ProjectionName() != "3d" {
		side := ax.effectiveXLabelSide()
		layout := measureSingleLineTextLayout(r, ax.XLabel, labelSize, ctx.RC.FontKey, ctx.RC.UseTeX)
		anchor, vAlign := xLabelAnchorPoint(ax, r, ctx, px, side, alignment)
		drawDisplayText(
			textRen,
			ax.XLabel,
			alignedSingleLineOrigin(anchor, layout, TextAlignCenter, vAlign),
			labelSize,
			labelColor,
			ctx.RC.FontKey,
			ctx.RC.UseTeX,
		)
	}

	// YLabel: vertical text if supported, else horizontal fallback
	if ax.YLabel != "" && ax.ProjectionName() != "3d" {
		side := ax.effectiveYLabelSide()
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
			yLabelLayout := measureSingleLineTextLayout(r, ax.YLabel, labelSize, ctx.RC.FontKey, ctx.RC.UseTeX)
			yLabelVAlign := textLayoutVAlignBottom
			if side == AxisRight {
				yLabelVAlign = textLayoutVAlignTop
			}
			backendAnchor := rotatedTextBackendAnchorFromP(anchor, yLabelLayout, TextAlignCenter, yLabelVAlign, labelAngle, true)
			drawDisplayTextRotated(ren, ax.YLabel, backendAnchor, labelSize, labelAngle, labelColor, ctx.RC.FontKey, ctx.RC.UseTeX)
		case render.VerticalTextDrawer:
			if angle < 0 {
				layout := measureSingleLineTextLayout(r, ax.YLabel, labelSize, ctx.RC.FontKey, ctx.RC.UseTeX)
				drawDisplayText(
					textRen,
					ax.YLabel,
					alignedSingleLineOrigin(geom.Pt{X: anchor.X, Y: px.Min.Y + px.H()/2}, layout, TextAlignCenter, textLayoutVAlignCenter),
					labelSize,
					labelColor,
					ctx.RC.FontKey,
					ctx.RC.UseTeX,
				)
			} else {
				drawDisplayTextVertical(ren, ax.YLabel, geom.Pt{X: anchor.X, Y: anchor.Y}, labelSize, labelColor, ctx.RC.FontKey)
			}
		default:
			layout := measureSingleLineTextLayout(r, ax.YLabel, labelSize, ctx.RC.FontKey, ctx.RC.UseTeX)
			drawDisplayText(
				textRen,
				ax.YLabel,
				alignedSingleLineOrigin(geom.Pt{X: anchor.X, Y: px.Min.Y + px.H()/2}, layout, TextAlignCenter, textLayoutVAlignCenter),
				labelSize,
				labelColor,
				ctx.RC.FontKey,
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
	titlePadPx := pointsToPixels(ctx.RC, 6)
	topExtent := titleTopExtent(ax, r, ctx, px)
	if aligned, ok := alignment.titleExtents[alignmentKey(AxisTop, spinePixelY(AxisTop, px))]; ok {
		topExtent = aligned
	}
	return geom.Pt{
		X: ctx.TransAxes().Apply(geom.Pt{X: 0.5, Y: 1}).X,
		Y: topExtent + titlePadPx,
	}
}

func xLabelAnchorPoint(ax *Axes, r render.Renderer, ctx *DrawContext, px geom.Rect, side AxisSide, alignment figureTextAlignment) (geom.Pt, textLayoutVerticalAlign) {
	anchor := ctx.TransAxes().Apply(geom.Pt{X: 0.5, Y: 0})

	if isGeoProjection(ax.projection) {
		if side == AxisTop {
			anchor.Y = px.Max.Y + axisLabelPadPx(ctx)
			return anchor, textLayoutVAlignBaseline
		}
		anchor.Y = px.Min.Y - axisLabelPadPx(ctx)
		return anchor, textLayoutVAlignTop
	}

	xAxis := ax.axisForXLabelSide(side)
	if side == AxisTop {
		topExtent := xLabelSpinePixelY(AxisTop, px)
		if xAxis != nil {
			if tickBounds, ok := axisTickLabelBounds(xAxis, r, ctx); ok {
				topExtent = math.Max(topExtent, tickBounds.Max.Y)
			}
		}
		if aligned, ok := alignment.xLabelExtents[alignmentKey(side, xLabelSpinePixelY(side, px))]; ok {
			topExtent = aligned
		}
		anchor.Y = topExtent + axisLabelPadPx(ctx)
		return anchor, textLayoutVAlignBaseline
	}

	bottomExtent := xLabelSpinePixelY(AxisBottom, px)
	if xAxis != nil {
		if tickBounds, ok := axisTickLabelBounds(xAxis, r, ctx); ok {
			bottomExtent = math.Min(bottomExtent, tickBounds.Min.Y)
		}
	}
	if aligned, ok := alignment.xLabelExtents[alignmentKey(side, xLabelSpinePixelY(side, px))]; ok {
		bottomExtent = aligned
	}
	anchor.Y = bottomExtent - axisLabelPadPx(ctx)
	return anchor, textLayoutVAlignTop
}

func yLabelAnchorPoint(ax *Axes, r render.Renderer, ctx *DrawContext, px geom.Rect, side AxisSide, alignment figureTextAlignment) geom.Pt {
	anchor := ctx.TransAxes().Apply(geom.Pt{X: 0, Y: 0.5})
	anchor.X = spinePixelX(AxisLeft, px) - axisLabelPadPx(ctx)

	yAxis := ax.axisForYLabelSide(side)
	if side == AxisRight {
		spineX := spinePixelX(AxisRight, px)
		rightExtent := spineX
		if tickBounds, ok := axisTickLabelBounds(yAxis, r, ctx); ok {
			rightExtent = math.Max(rightExtent, tickBounds.Max.X)
		} else if yAxis != nil && yAxis.ShowTicks {
			rightExtent += tickLabelPadPx(yAxis, ctx)
		}
		if aligned, ok := alignment.yLabelExtents[alignmentKey(side, spinePixelX(side, px))]; ok {
			rightExtent = aligned
		}
		anchor.X = rightExtent + axisLabelPadPx(ctx)
		return anchor
	}

	spineX := spinePixelX(AxisLeft, px)
	leftExtent := spineX
	if tickBounds, ok := axisTickLabelBounds(yAxis, r, ctx); ok {
		leftExtent = math.Min(leftExtent, tickBounds.Min.X)
	} else if yAxis != nil && yAxis.ShowTicks {
		leftExtent -= tickLabelPadPx(yAxis, ctx)
	}
	if aligned, ok := alignment.yLabelExtents[alignmentKey(side, spinePixelX(side, px))]; ok {
		leftExtent = aligned
	}
	leftExtent = math.Ceil(leftExtent)
	anchor.X = leftExtent - axisLabelPadPx(ctx)
	return anchor
}

func axisLabelPadPx(ctx *DrawContext) float64 {
	if ctx == nil {
		return pointsToPixels(style.CurrentDefaults(), 4)
	}
	return pointsToPixels(ctx.RC, 4)
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
