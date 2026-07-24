package core

import (
	"math"
	"sort"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/transform"
)

type axisAlignmentKey struct {
	side AxisSide
	line int
}

type figureTextAlignment struct {
	titleExtents  map[axisAlignmentKey]float64
	xLabelExtents map[axisAlignmentKey]float64
	yLabelExtents map[axisAlignmentKey]float64
}

const (
	// Matplotlib figure-label defaults from Figure.suptitle/supxlabel/supylabel.
	// See third_party/matplotlib/lib/matplotlib/figure.py:_suplabels.
	matplotlibSuptitleY  = 0.98
	matplotlibSupxlabelY = 0.01
	matplotlibSupylabelX = 0.02

	// Matplotlib rc default legend.borderaxespad, in legend font-size units.
	// Used when stacking figure-level artists around automatically positioned
	// figure labels.
	matplotlibLegendBorderAxesPad = 0.5
)

func newAxesDrawContext(ax *Axes, fig *Figure, figureRect, clip geom.Rect) *DrawContext {
	proj := cloneProjection(ax.projection)
	if proj == nil {
		proj, _ = lookupProjection("rectilinear")
	}
	rawDataToAxes := proj.DataToAxes(ax)

	// Resolve the effective data->axes leg the same way Transform2D.transData
	// does: a nil projection transform (e.g. 3D, which pre-projects in the
	// artist) falls back to the per-axis scale transform. The raw value
	// (possibly nil) is kept in the field so the separable-decomposition
	// fast-paths behave exactly as before; the resolved value feeds the cache.
	effectiveDataToAxes := rawDataToAxes
	if effectiveDataToAxes == nil {
		effectiveDataToAxes = transform.NewScaleTransform(ax.effectiveXScale(), ax.effectiveYScale())
	}

	// Point the persistent transform graph at the current geometry. Both calls
	// invalidate downstream caches only when their input actually changed, so an
	// unchanged axes (repeated draw / same size) reuses the cached affine.
	ax.updateAxesBbox(clip)
	ax.refreshDataTransform(effectiveDataToAxes)

	return &DrawContext{
		DataToPixel: Transform2D{
			XScale:      ax.effectiveXScale(),
			YScale:      ax.effectiveYScale(),
			DataToAxes:  rawDataToAxes,
			AxesToPixel: ax.transAxes,
			composed:    ax.transData,
		},
		Axes:       ax,
		Projection: proj,
		RC:         ax.effectiveRC(fig),
		Clip:       clip,
		FigureRect: figureRect,
	}
}

func newFigureDrawContext(fig *Figure, figureRect geom.Rect) *DrawContext {
	return &DrawContext{
		DataToPixel: Transform2D{
			XScale:      transform.NewLinear(0, 1),
			YScale:      transform.NewLinear(0, 1),
			DataToAxes:  transform.NewScaleTransform(transform.NewLinear(0, 1), transform.NewLinear(0, 1)),
			AxesToPixel: transform.NewDisplayRectTransform(figureRect),
		},
		Axes:       nil,
		Projection: cloneProjection(nil),
		RC:         fig.RC,
		Clip:       figureRect,
		FigureRect: figureRect,
	}
}

func computeFigureTextAlignment(fig *Figure, r render.Renderer, figureRect geom.Rect) figureTextAlignment {
	alignment := figureTextAlignment{
		titleExtents:  map[axisAlignmentKey]float64{},
		xLabelExtents: map[axisAlignmentKey]float64{},
		yLabelExtents: map[axisAlignmentKey]float64{},
	}
	if fig == nil || fig.layoutEngine == LayoutEngineNone {
		return alignment
	}

	for _, ax := range fig.Children {
		px := ax.adjustedLayout(fig)
		ctx := newAxesDrawContext(ax, fig, figureRect, px)

		if ax.XLabel != "" {
			side := ax.effectiveXLabelSide()
			key := alignmentKey(side, xLabelSpinePixelY(side, px))
			extent := xLabelExtent(ax, r, ctx, px, side)
			if side == AxisTop {
				if current, ok := alignment.xLabelExtents[key]; !ok || extent < current {
					alignment.xLabelExtents[key] = extent
				}
			} else if current, ok := alignment.xLabelExtents[key]; !ok || extent > current {
				alignment.xLabelExtents[key] = extent
			}
		}

		if ax.YLabel != "" {
			side := ax.effectiveYLabelSide()
			key := alignmentKey(side, spinePixelX(side, px))
			extent := yLabelExtent(ax, r, ctx, px, side)
			if side == AxisRight {
				if current, ok := alignment.yLabelExtents[key]; !ok || extent > current {
					alignment.yLabelExtents[key] = extent
				}
			} else if current, ok := alignment.yLabelExtents[key]; !ok || extent < current {
				alignment.yLabelExtents[key] = extent
			}
		}
	}

	return alignment
}

func alignmentKey(side AxisSide, coord float64) axisAlignmentKey {
	return axisAlignmentKey{
		side: side,
		line: int(math.Round(coord)),
	}
}

func titleTopExtent(ax *Axes, r render.Renderer, ctx *DrawContext, px geom.Rect) float64 {
	extent := titleTopExtentForAxes(ax, r, ctx, px)
	if ax == nil || ax.figure == nil || ctx == nil {
		return extent
	}
	for _, child := range ax.childAxes {
		if child == nil {
			continue
		}
		childPx := child.adjustedLayout(ax.figure)
		childCtx := newAxesDrawContext(child, ax.figure, ctx.FigureRect, childPx)
		childExtent := titleTopExtentForAxes(child, r, childCtx, childPx)
		extent = math.Max(extent, childExtent)
	}
	return extent
}

func titleTopExtentForAxes(ax *Axes, r render.Renderer, ctx *DrawContext, px geom.Rect) float64 {
	extent := px.Max.Y
	for _, candidate := range []*Axis{ax.effectiveXAxis(), ax.effectiveTopAxis()} {
		if candidate == nil || !candidate.ShowLabels {
			continue
		}
		if ax.tickLabelsHiddenForLayout(candidate) {
			continue
		}
		if !isPolarProjection(ctx.Projection) && candidate.Side != AxisTop {
			continue
		}
		if tickBounds, ok := axisTickLabelBounds(candidate, r, ctx); ok {
			extent = math.Max(extent, tickBounds.Max.Y)
		}
	}
	if ax.XLabel != "" && !ax.hideXLabel && ax.effectiveXLabelSide() == AxisTop {
		layout := measureSingleLineTextLayout(r, ax.XLabel, axisLabelFontSize(ctx), xAxisLabelFontKey(ax, ctx), ctx.RC.UseTeX)
		anchor, vAlign := xLabelAnchorPoint(ax, r, ctx, px, AxisTop, figureTextAlignment{})
		origin := alignedSingleLineOrigin(anchor, layout, TextAlignCenter, vAlign)
		if layout.Ascent > 0 {
			extent = math.Max(extent, origin.Y+layout.Ascent)
		} else if bounds, ok := textInkRect(origin, layout); ok {
			extent = math.Max(extent, bounds.Max.Y)
		}
	}
	return extent
}

func xLabelExtent(ax *Axes, r render.Renderer, ctx *DrawContext, px geom.Rect, side AxisSide) float64 {
	extent := xLabelSpinePixelY(side, px)
	xAxis := ax.axisForXLabelSide(side)
	if xAxis == nil {
		return extent
	}
	if bounds, ok := axisTickLabelBounds(xAxis, r, ctx); ok && !ax.tickLabelsHiddenForLayout(xAxis) {
		if side == AxisTop {
			return math.Max(extent, bounds.Max.Y)
		}
		return math.Min(extent, bounds.Min.Y)
	} else if xAxis.ShowTicks {
		if side == AxisTop {
			return extent + xAxis.TickSize*tickOutsidePaddingFactor(xAxis)
		}
		return extent - xAxis.TickSize*tickOutsidePaddingFactor(xAxis)
	}
	return extent
}

func yLabelExtent(ax *Axes, r render.Renderer, ctx *DrawContext, px geom.Rect, side AxisSide) float64 {
	extent := spinePixelX(side, px)
	yAxis := ax.axisForYLabelSide(side)
	if yAxis == nil {
		return extent
	}
	if bounds, ok := axisTickLabelBounds(yAxis, r, ctx); ok && !ax.tickLabelsHiddenForLayout(yAxis) {
		if side == AxisRight {
			return math.Max(extent, bounds.Max.X)
		}
		return math.Min(extent, bounds.Min.X)
	}
	return extent
}

func drawFigureArtistsWithOptions(fig *Figure, r render.Renderer, figureRect geom.Rect, opts DrawOptions) {
	if fig == nil || len(fig.Artists) == 0 {
		return
	}

	if !fig.zsorted {
		sortArtists(fig.Artists)
		fig.zsorted = true
	}

	ctx := newFigureDrawContext(fig, figureRect)
	ctx.DrawOptions = opts
	stackOffsets := initialFigureArtistStackOffsets(fig, r, ctx)
	for _, art := range fig.Artists {
		artCtx := *ctx
		if loc, ok := figureArtistLocation(art); ok && !figureArtistUsesFigureCoordinates(art) {
			offset := stackOffsets[loc]
			artCtx.Clip = insetFigureArtistClip(ctx.Clip, loc, offset)
			if box, hasBox := figureArtistBoxRect(art, r, &artCtx); hasBox {
				stackOffsets[loc] = offset + box.H() + figureArtistStackPadPx(ctx)
			}
		}
		drawArtist(r, &artCtx, art)
		if overlay, ok := art.(OverlayArtist); ok {
			drawOverlayArtist(r, &artCtx, art, overlay)
		}
	}
}

func figureArtistUsesFigureCoordinates(art Artist) bool {
	var locator AnchoredBoxLocator
	switch a := art.(type) {
	case *Legend:
		locator = a.Locator
	case *AnchoredTextBox:
		locator = a.Locator
	}
	if locator == nil {
		return false
	}
	_, ok := locator.(figureCoordinateBoxLocator)
	return ok
}

func initialFigureArtistStackOffsets(fig *Figure, r render.Renderer, ctx *DrawContext) map[LegendLocation]float64 {
	offsets := map[LegendLocation]float64{}
	if fig == nil || ctx == nil {
		return offsets
	}
	if fig.SupTitle != "" {
		offset := figureLabelTightHeight(r, fig.SupTitle, figureTitleFontSize(ctx), fontKeyWithWeight(ctx.RC.FontKey, ctx.RC.Figure.TitleWeight), ctx.RC.UseTeX) + figureLabelTopInsetPx(fig, ctx)
		offsets[LegendUpperLeft] = offset
		offsets[LegendUpperRight] = offset
	}
	if fig.SupXLabel != "" {
		offset := figureLabelTightHeight(r, fig.SupXLabel, figureLabelFontSize(ctx), fontKeyWithWeight(ctx.RC.FontKey, ctx.RC.Figure.LabelWeight), ctx.RC.UseTeX) + figureLabelBottomInsetPx(fig, ctx)
		offsets[LegendLowerLeft] = offset
		offsets[LegendLowerRight] = offset
	}
	return offsets
}

func figureArtistStackPadPx(ctx *DrawContext) float64 {
	if ctx == nil {
		rc := style.CurrentDefaults()
		return pointsToPixels(rc, matplotlibLegendBorderAxesPad*rc.LegendSize())
	}
	return pointsToPixels(ctx.RC, matplotlibLegendBorderAxesPad*ctx.RC.LegendSize())
}

func insetFigureArtistClip(clip geom.Rect, location LegendLocation, offset float64) geom.Rect {
	if offset <= 0 {
		return clip
	}
	switch location {
	case LegendLowerLeft, LegendLowerRight:
		clip.Min.Y += offset
	default:
		clip.Max.Y -= offset
	}
	return clip
}

func figureArtistLocation(art Artist) (LegendLocation, bool) {
	switch a := art.(type) {
	case *Legend:
		return a.Location, true
	case *AnchoredTextBox:
		return a.Location, true
	default:
		return LegendUpperRight, false
	}
}

func figureArtistBoxRect(art Artist, r render.Renderer, ctx *DrawContext) (geom.Rect, bool) {
	switch a := art.(type) {
	case *Legend:
		return a.boxRect(r, ctx)
	case *AnchoredTextBox:
		return a.boxRect(r, ctx)
	default:
		return geom.Rect{}, false
	}
}

func drawFigureLabels(fig *Figure, r render.Renderer, figureRect geom.Rect) {
	if fig == nil {
		return
	}

	textRen, ok := r.(render.TextDrawer)
	if !ok {
		return
	}

	ctx := newFigureDrawContext(fig, figureRect)
	titleColor := fig.RC.DefaultAxesTitleColor()
	labelColor := fig.RC.DefaultAxesLabelColor()
	titleSize := figureTitleFontSize(ctx)
	labelSize := figureLabelFontSize(ctx)
	titleFontKey := fontKeyWithWeight(fig.RC.FontKey, fig.RC.Figure.TitleWeight)
	labelFontKey := fontKeyWithWeight(fig.RC.FontKey, fig.RC.Figure.LabelWeight)
	centerX := figureRect.Min.X + figureRect.W()/2
	centerY := figureRect.Min.Y + figureRect.H()/2

	if fig.SupTitle != "" {
		layout := measureSingleLineTextLayout(r, fig.SupTitle, titleSize, titleFontKey, fig.RC.UseTeX)
		y := figureRect.Max.Y - figureLabelTopInsetPx(fig, ctx)
		anchor := geom.Pt{
			X: centerX,
			Y: y,
		}
		drawDisplayText(
			textRen,
			fig.SupTitle,
			alignedSingleLineOrigin(anchor, layout, TextAlignCenter, textLayoutVAlignTop),
			titleSize,
			titleColor,
			titleFontKey,
			fig.RC.UseTeX,
		)
	}

	if fig.SupXLabel != "" {
		layout := measureSingleLineTextLayout(r, fig.SupXLabel, labelSize, labelFontKey, fig.RC.UseTeX)
		y := figureRect.Min.Y + figureLabelBottomInsetPx(fig, ctx)
		anchor := geom.Pt{
			X: centerX,
			Y: y,
		}
		drawDisplayText(
			textRen,
			fig.SupXLabel,
			alignedSingleLineOrigin(anchor, layout, TextAlignCenter, textLayoutVAlignBottom),
			labelSize,
			labelColor,
			labelFontKey,
			fig.RC.UseTeX,
		)
	}

	if fig.SupYLabel != "" {
		layout := measureSingleLineTextLayout(r, fig.SupYLabel, labelSize, labelFontKey, fig.RC.UseTeX)
		leftPad := figureLabelLeftInsetPx(fig, ctx)
		p := geom.Pt{
			X: figureRect.Min.X + leftPad,
			Y: centerY,
		}
		switch ren := r.(type) {
		case render.RotatedTextDrawer:
			anchor := rotatedTextBackendAnchorFromP(p, layout, TextAlignLeft, textLayoutVAlignCenter, math.Pi/2, false)
			drawDisplayTextRotated(ren, fig.SupYLabel, anchor, labelSize, math.Pi/2, labelColor, labelFontKey, fig.RC.UseTeX)
		case render.VerticalTextDrawer:
			drawDisplayTextVertical(ren, fig.SupYLabel, p, labelSize, labelColor, labelFontKey)
		default:
			drawDisplayText(
				textRen,
				fig.SupYLabel,
				alignedSingleLineOrigin(p, layout, TextAlignLeft, textLayoutVAlignCenter),
				labelSize,
				labelColor,
				labelFontKey,
				fig.RC.UseTeX,
			)
		}
	}
}

func figureLabelTopInsetPx(fig *Figure, ctx *DrawContext) float64 {
	if fig != nil && fig.layoutEngine == LayoutEngineConstrained {
		return constrainedLayoutPadPx(fig)
	}
	if ctx == nil {
		return 0
	}
	return (1 - matplotlibSuptitleY) * ctx.FigureRect.H()
}

func figureLabelBottomInsetPx(fig *Figure, ctx *DrawContext) float64 {
	if fig != nil && fig.layoutEngine == LayoutEngineConstrained {
		return constrainedLayoutPadPx(fig)
	}
	if ctx == nil {
		return 0
	}
	return matplotlibSupxlabelY * ctx.FigureRect.H()
}

func figureLabelLeftInsetPx(fig *Figure, ctx *DrawContext) float64 {
	if fig != nil && fig.layoutEngine == LayoutEngineConstrained {
		return constrainedLayoutPadPx(fig)
	}
	if ctx == nil {
		return 0
	}
	return matplotlibSupylabelX * ctx.FigureRect.W()
}

func pointsToPixels(rc style.RC, points float64) float64 {
	dpi := rc.DPI
	if dpi <= 0 {
		dpi = style.CurrentDefaults().DPI
		if dpi <= 0 {
			dpi = 96
		}
	}
	return points * dpi / 72.0
}

func sortArtists(artists []Artist) {
	if len(artists) < 2 {
		return
	}
	sort.SliceStable(artists, func(i, j int) bool {
		zi, zj := artists[i].Z(), artists[j].Z()
		if zi == zj {
			return i < j
		}
		return zi < zj
	})
}

func sortedArtistDrawOrder(artists []Artist) []Artist {
	if len(artists) < 2 {
		return artists
	}
	drawOrder := append([]Artist(nil), artists...)
	sortArtists(drawOrder)
	return drawOrder
}
