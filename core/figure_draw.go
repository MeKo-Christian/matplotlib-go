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
	setRendererSketch(r, fig.RC.PathSketch)
	drawFigureBackground(r, vp, opts, fig)

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
			drawSecondaryChildAxes(ax, fig, r, vp, opts, alignment)
			continue
		}

		if ax.PatchVisible && !opts.Transparent && shouldDrawAxesBackground(ctx.RC.AxesBackground, fig.RC.FigureBackground(), px, drawnAxes) {
			backgroundPath := pixelRectPath(px)
			if framePath, ok := projectionFramePath(ctx.Projection, px); ok {
				backgroundPath = framePath
			}
			r.Path(backgroundPath, &render.Paint{
				Fill: ctx.RC.AxesBackground,
			})
		}
		drawnAxes = append(drawnAxes, px)

		// Matplotlib sorts regular artists, Axis objects (zorder 1.5), and
		// Spine artists (zorder 2.5) together. Keep the axes clip active only
		// for regular artist bands; ticks and spines can protrude past it.
		drawClippedAxesArtistsInZRange(r, ctx, px, ax.Artists, math.Inf(-1), 1.5, true)

		if xAxis != nil {
			xAxis.DrawTicks(r, ctx)
			if !ax.hideXTickLabels {
				xAxis.DrawTickLabels(r, ctx)
			}
		}
		if yAxis != nil {
			yAxis.DrawTicks(r, ctx)
			if !ax.hideYTickLabels {
				yAxis.DrawTickLabels(r, ctx)
			}
		}
		if topAxis != nil {
			topAxis.DrawTicks(r, ctx)
			if !ax.hideTopTickLabels {
				topAxis.DrawTickLabels(r, ctx)
			}
		}
		if rightAxis != nil {
			rightAxis.DrawTicks(r, ctx)
			if !ax.hideRightTickLabels {
				rightAxis.DrawTickLabels(r, ctx)
			}
		}
		for _, extraAxis := range ax.ExtraAxes {
			if extraAxis != nil {
				extraAxis.DrawTicks(r, ctx)
				extraAxis.DrawTickLabels(r, ctx)
			}
		}

		drawClippedAxesArtistsInZRange(r, ctx, px, ax.Artists, 1.5, 2.5, true)

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
			DrawFrame(
				r,
				ctx,
				ref,
				topAxis == nil && ax.fallbackSpineVisible(AxisTop),
				rightAxis == nil && ax.fallbackSpineVisible(AxisRight),
			)
		}

		drawClippedAxesArtistsInZRange(r, ctx, px, ax.Artists, 2.5, math.Inf(1), true)
		drawClippedAxesArtistsInZRange(r, ctx, px, ax.WidgetArtists, math.Inf(-1), math.Inf(1), false)

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

		// Draw axes text labels outside the clip rect.
		drawAxesLabels(ax, r, ctx, px, alignment)
	}

	setRendererResolution(r, fig.RC.DPI)
	drawFigureArtistsWithOptions(fig, r, vp, opts)
	if opts.AnimatedFilter != AnimatedFilterOnlyAnimated {
		drawFigureLabels(fig, r, vp)
	}
}

func drawClippedAxesArtistsInZRange(r render.Renderer, ctx *DrawContext, px geom.Rect, artists []Artist, minZ, maxZ float64, skipLegends bool) {
	if len(artists) == 0 {
		return
	}
	drawOrder := sortedArtistDrawOrder(artists)
	shouldDraw := false
	for _, art := range drawOrder {
		if art == nil {
			continue
		}
		if skipLegends {
			if _, ok := art.(*Legend); ok {
				continue
			}
		}
		z := art.Z()
		if z >= minZ && z < maxZ {
			shouldDraw = true
			break
		}
	}
	if !shouldDraw {
		return
	}
	r.Save()
	r.ClipRect(px)
	if framePath, ok := projectionFramePath(ctx.Projection, px); ok {
		r.ClipPath(framePath)
	}
	for _, art := range drawOrder {
		if art == nil {
			continue
		}
		if skipLegends {
			if _, ok := art.(*Legend); ok {
				continue
			}
		}
		z := art.Z()
		if z < minZ || z >= maxZ {
			continue
		}
		drawArtist(r, ctx, art)
	}
	r.Restore()
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
		if opts.AnimatedFilter == AnimatedFilterOnlyAnimated {
			continue
		}

		for _, ai := range []struct {
			axis   *Axis
			hidden bool
		}{
			{child.effectiveXAxis(), child.hideXTickLabels},
			{child.effectiveYAxis(), child.hideYTickLabels},
			{child.effectiveTopAxis(), child.hideTopTickLabels},
			{child.effectiveRightAxis(), child.hideRightTickLabels},
		} {
			if ai.axis == nil {
				continue
			}
			ai.axis.DrawTicks(r, ctx)
			if !ai.hidden {
				ai.axis.DrawTickLabels(r, ctx)
			}
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

// drawFigureBackground paints the save-time figure face fill and edge stroke
// requested via DrawOptions (savefig.facecolor/edgecolor). It is a no-op for the
// common case where neither override is set.
//
// When the global sketch/xkcd perturbation is active, the figure background is
// instead reproduced the way Matplotlib draws it: as a real, non-antialiased
// figure patch (figure.patch, antialiased=False) composited over a transparent
// canvas. The xkcd wiggle then perforates the canvas border with the same
// fully-transparent notch pixels the reference renderer produces, which an
// opaque clear can never reproduce.
func drawFigureBackground(r render.Renderer, vp geom.Rect, opts DrawOptions, fig *Figure) {
	if opts.AnimatedFilter == AnimatedFilterOnlyAnimated {
		return
	}
	if drawSketchedFigurePatch(r, vp, opts, fig) {
		return
	}
	if opts.FigureBackground != nil && opts.FigureBackground.A > 0 {
		r.Path(pixelRectPath(vp), &render.Paint{Fill: *opts.FigureBackground})
	}
	if opts.FigureEdge != nil && opts.FigureEdge.A > 0 && opts.FigureEdgeWidth > 0 {
		r.Path(pixelRectPath(vp), &render.Paint{
			Stroke:    *opts.FigureEdge,
			LineWidth: opts.FigureEdgeWidth,
		})
	}
}

// drawSketchedFigurePatch reproduces Matplotlib's transparent-canvas figure
// patch when the global sketch is active. It clears the buffer transparent and
// fills the full-canvas rectangle with the figure facecolor using non-
// antialiased (binary-coverage) fill — exactly as Matplotlib renders
// figure.patch (antialiased=False). The renderer's default sketch (set from
// fig.RC.PathSketch) wiggles the rectangle edges, so the border picks up the
// same fully-transparent notch pixels as the reference. Returns false (leaving
// the caller's opaque-clear path intact) unless the backend supports a
// transparent clear and a sketch is genuinely active.
func drawSketchedFigurePatch(r render.Renderer, vp geom.Rect, opts DrawOptions, fig *Figure) bool {
	if fig == nil || !render.SketchActive(fig.RC.PathSketch) {
		return false
	}
	clearer, ok := r.(render.TransparentClearer)
	if !ok {
		return false
	}
	face := fig.RC.FigureBackground()
	if opts.FigureBackground != nil {
		face = *opts.FigureBackground
	}
	clearer.ClearTransparent()
	// A fully transparent figure (savefig transparent=True) draws no patch.
	if !opts.Transparent && face.A > 0 {
		r.Path(pixelRectPath(vp), &render.Paint{
			Fill:      face,
			Antialias: render.AntialiasOff,
		})
	}
	if opts.FigureEdge != nil && opts.FigureEdge.A > 0 && opts.FigureEdgeWidth > 0 {
		r.Path(pixelRectPath(vp), &render.Paint{
			Stroke:    *opts.FigureEdge,
			LineWidth: opts.FigureEdgeWidth,
		})
	}
	return true
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

// setRendererSketch pushes the global sketch/xkcd default onto renderers that
// support it, so every path is perturbed unless an artist overrides per-paint.
func setRendererSketch(r render.Renderer, params render.SketchParams) {
	if setter, ok := r.(render.SketchAware); ok {
		setter.SetDefaultSketch(params)
	}
}
