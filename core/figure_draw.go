package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func DrawFigure(fig *Figure, r render.Renderer) {
	DrawFigureWithOptions(fig, r, DrawOptions{})
}

func DrawFigureWithOptions(fig *Figure, r render.Renderer, opts DrawOptions) {
	vp := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: fig.SizePx.X, Y: fig.SizePx.Y}}
	_ = r.Begin(vp)
	defer r.End()
	setRendererResolution(r, fig.RC.DPI)

	prepareFigureLayout(fig, r, vp)
	syncAxesLocators(fig, r)
	syncColorbarAxesMeasured(fig, r, vp)
	syncAxesLocators(fig, r)
	alignment := computeFigureTextAlignment(fig, r, vp)

	drawnAxes := make([]geom.Rect, 0, len(fig.Children))
	for _, ax := range fig.Children {
		if isSecondaryAxes(ax) {
			continue
		}
		px := ax.adjustedLayout(fig)
		xAxis := ax.effectiveXAxis()
		yAxis := ax.effectiveYAxis()
		topAxis := ax.effectiveTopAxis()
		rightAxis := ax.effectiveRightAxis()

		// Build DrawContext with composed transform
		ctx := newAxesDrawContext(ax, fig, vp, px)
		ctx.DrawOptions = opts
		setRendererResolution(r, ctx.RC.DPI)

		// In an animated-overlay pass we only redraw the animated artists on
		// top of a previously captured background, so skip backgrounds,
		// chrome, axis ticks, spines, and labels.
		if opts.AnimatedFilter == AnimatedFilterOnlyAnimated {
			r.Save()
			r.ClipRect(px)
			if framePath, ok := projectionFramePath(ctx.Projection, px); ok {
				r.ClipPath(framePath)
			}
			for _, art := range sortedArtistDrawOrder(ax.Artists) {
				if _, ok := art.(*Legend); ok {
					continue
				}
				drawArtist(r, ctx, art)
			}
			for _, art := range sortedArtistDrawOrder(ax.WidgetArtists) {
				drawArtist(r, ctx, art)
			}
			r.Restore()
			for _, art := range sortedArtistDrawOrder(ax.Artists) {
				if overlay, ok := art.(OverlayArtist); ok {
					drawOverlayArtist(r, ctx, art, overlay)
				}
			}
			for _, art := range sortedArtistDrawOrder(ax.WidgetArtists) {
				if overlay, ok := art.(OverlayArtist); ok {
					drawOverlayArtist(r, ctx, art, overlay)
				}
			}
			continue
		}

		if ax.PatchVisible && shouldDrawAxesBackground(ctx.RC.AxesBackground, fig.RC.FigureBackground(), px, drawnAxes) {
			backgroundPath := pixelRectPath(px)
			if framePath, ok := projectionFramePath(ctx.Projection, px); ok {
				backgroundPath = framePath
			}
			r.Path(backgroundPath, &render.Paint{
				Fill: ctx.RC.AxesBackground,
			})
		}
		drawnAxes = append(drawnAxes, px)

		// Draw only clipped content (data and grids) while the axes clip is active.
		r.Save()
		r.ClipRect(px)
		if framePath, ok := projectionFramePath(ctx.Projection, px); ok {
			r.ClipPath(framePath)
		}

		for _, art := range sortedArtistDrawOrder(ax.Artists) {
			if _, ok := art.(*Legend); ok {
				continue
			}
			drawArtist(r, ctx, art)
		}
		for _, art := range sortedArtistDrawOrder(ax.WidgetArtists) {
			drawArtist(r, ctx, art)
		}

		r.Restore()

		for _, art := range sortedArtistDrawOrder(ax.Artists) {
			if overlay, ok := art.(OverlayArtist); ok {
				drawOverlayArtist(r, ctx, art, overlay)
			}
		}
		for _, art := range sortedArtistDrawOrder(ax.WidgetArtists) {
			if overlay, ok := art.(OverlayArtist); ok {
				drawOverlayArtist(r, ctx, art, overlay)
			}
		}

		drawSecondaryChildAxes(ax, fig, r, vp, opts, alignment)

		// Matplotlib draws Axis objects (ticks and tick labels, zorder 1.5)
		// before Spine artists (zorder 2.5). This matters at endpoint ticks:
		// the spine overpaints the tick cap by a single coverage level.
		if xAxis != nil {
			xAxis.DrawTicks(r, ctx)
			xAxis.DrawTickLabels(r, ctx)
		}
		if yAxis != nil {
			yAxis.DrawTicks(r, ctx)
			yAxis.DrawTickLabels(r, ctx)
		}
		if topAxis != nil {
			topAxis.DrawTicks(r, ctx)
			topAxis.DrawTickLabels(r, ctx)
		}
		if rightAxis != nil {
			rightAxis.DrawTicks(r, ctx)
			rightAxis.DrawTickLabels(r, ctx)
		}
		for _, extraAxis := range ax.ExtraAxes {
			if extraAxis != nil {
				extraAxis.DrawTicks(r, ctx)
				extraAxis.DrawTickLabels(r, ctx)
			}
		}

		// Draw spines outside the clip so they can straddle the axes edge the
		// same way Matplotlib does.
		if xAxis != nil {
			xAxis.Draw(r, ctx)
		}
		if yAxis != nil {
			yAxis.Draw(r, ctx)
		}
		if topAxis != nil {
			topAxis.Draw(r, ctx)
		}
		if rightAxis != nil {
			rightAxis.Draw(r, ctx)
		}
		for _, extraAxis := range ax.ExtraAxes {
			if extraAxis != nil {
				extraAxis.Draw(r, ctx)
			}
		}
		if ax.ShowFrame {
			ref := xAxis
			if ref == nil {
				ref = topAxis
			}
			if ref == nil {
				ref = yAxis
			}
			if ref == nil {
				ref = rightAxis
			}
			DrawFrame(r, ctx, ref, topAxis == nil, rightAxis == nil)
		}

		// Draw axes text labels outside the clip rect.
		drawAxesLabels(ax, r, ctx, px, alignment)
	}

	setRendererResolution(r, fig.RC.DPI)
	drawFigureArtistsWithOptions(fig, r, vp, opts)
	if opts.AnimatedFilter != AnimatedFilterOnlyAnimated {
		drawFigureLabels(fig, r, vp)
	}
}

func drawSecondaryChildAxes(parent *Axes, fig *Figure, r render.Renderer, vp geom.Rect, opts DrawOptions, alignment figureTextAlignment) {
	if parent == nil || fig == nil {
		return
	}
	for _, child := range parent.childAxes {
		if !isSecondaryAxes(child) {
			continue
		}
		px := child.adjustedLayout(fig)
		ctx := newAxesDrawContext(child, fig, vp, px)
		ctx.DrawOptions = opts
		setRendererResolution(r, ctx.RC.DPI)

		r.Save()
		r.ClipRect(px)
		if framePath, ok := projectionFramePath(ctx.Projection, px); ok {
			r.ClipPath(framePath)
		}
		for _, art := range sortedArtistDrawOrder(child.Artists) {
			if _, ok := art.(*Legend); ok {
				continue
			}
			drawArtist(r, ctx, art)
		}
		for _, art := range sortedArtistDrawOrder(child.WidgetArtists) {
			drawArtist(r, ctx, art)
		}
		r.Restore()

		for _, art := range sortedArtistDrawOrder(child.Artists) {
			if overlay, ok := art.(OverlayArtist); ok {
				drawOverlayArtist(r, ctx, art, overlay)
			}
		}
		for _, art := range sortedArtistDrawOrder(child.WidgetArtists) {
			if overlay, ok := art.(OverlayArtist); ok {
				drawOverlayArtist(r, ctx, art, overlay)
			}
		}

		for _, axis := range []*Axis{
			child.effectiveXAxis(),
			child.effectiveYAxis(),
			child.effectiveTopAxis(),
			child.effectiveRightAxis(),
		} {
			if axis == nil {
				continue
			}
			axis.DrawTicks(r, ctx)
			axis.DrawTickLabels(r, ctx)
		}
		for _, extraAxis := range child.ExtraAxes {
			if extraAxis == nil {
				continue
			}
			extraAxis.DrawTicks(r, ctx)
			extraAxis.DrawTickLabels(r, ctx)
		}
		for _, axis := range []*Axis{
			child.effectiveXAxis(),
			child.effectiveYAxis(),
			child.effectiveTopAxis(),
			child.effectiveRightAxis(),
		} {
			if axis != nil {
				axis.Draw(r, ctx)
			}
		}
		for _, extraAxis := range child.ExtraAxes {
			if extraAxis != nil {
				extraAxis.Draw(r, ctx)
			}
		}
		drawAxesLabels(child, r, ctx, px, alignment)
	}
}

func isSecondaryAxes(ax *Axes) bool {
	if ax == nil {
		return false
	}
	if _, ok := ax.XScale.(linkedSecondaryScale); ok {
		return true
	}
	if _, ok := ax.YScale.(linkedSecondaryScale); ok {
		return true
	}
	return false
}

func shouldDrawAxesBackground(axesBackground, figureBackground render.Color, px geom.Rect, previous []geom.Rect) bool {
	if axesBackground.A <= 0 {
		return false
	}
	if axesBackground != figureBackground {
		return true
	}
	for _, other := range previous {
		if rectsOverlap(px, other) {
			return true
		}
	}
	return false
}

func setRendererResolution(r render.Renderer, dpi float64) {
	if dpi <= 0 {
		return
	}
	if setter, ok := r.(render.DPIAware); ok {
		setter.SetResolution(uint(math.Round(dpi)))
	}
}
