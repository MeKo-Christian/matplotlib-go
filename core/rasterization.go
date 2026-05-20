package core

import "github.com/cwbudde/matplotlib-go/render"

const (
	autoRasterizeScatterPointThreshold = 1000
	autoRasterizeContourPathThreshold  = 200
)

// ArtistRasterization stores artist-level mixed raster/vector output intent.
//
// It is embedded by concrete artists that expose Matplotlib-style
// SetRasterized behavior without making rasterization mandatory on every
// custom Artist implementation.
type ArtistRasterization struct {
	rasterization render.Rasterization
}

// SetRasterized toggles rasterized output for this artist when a vector
// renderer supports mixed raster/vector embedding.
func (a *ArtistRasterization) SetRasterized(enabled bool) {
	if a == nil {
		return
	}
	if enabled {
		a.rasterization.Mode = render.RasterizeAlways
		return
	}
	a.rasterization.Mode = render.RasterizeNever
}

// SetRasterization sets the complete rasterization policy for this artist.
func (a *ArtistRasterization) SetRasterization(options render.Rasterization) {
	if a == nil {
		return
	}
	a.rasterization = options
}

// Rasterization returns the configured mixed output policy.
func (a *ArtistRasterization) Rasterization() render.Rasterization {
	if a == nil {
		return render.Rasterization{}
	}
	return a.rasterization
}

type rasterizationProvider interface {
	Rasterization() render.Rasterization
}

type pathEffectRasterizationProvider interface {
	rasterizationPathEffects() []render.PathEffect
}

func drawArtist(r render.Renderer, ctx *DrawContext, art Artist) {
	if art == nil {
		return
	}
	options, ok := artistRasterizationOptions(r, art, ctx)
	if !ok {
		art.Draw(r, ctx)
		return
	}
	controller, ok := r.(render.RasterizationController)
	if !ok || !controller.StartRasterized(options) {
		art.Draw(r, ctx)
		return
	}
	defer controller.StopRasterized()
	art.Draw(r, ctx)
}

func drawOverlayArtist(r render.Renderer, ctx *DrawContext, art Artist, overlay OverlayArtist) {
	if overlay == nil {
		return
	}
	options, ok := explicitArtistRasterizationOptions(art, ctx)
	if !ok {
		overlay.DrawOverlay(r, ctx)
		return
	}
	controller, ok := r.(render.RasterizationController)
	if !ok || !controller.StartRasterized(options) {
		overlay.DrawOverlay(r, ctx)
		return
	}
	defer controller.StopRasterized()
	overlay.DrawOverlay(r, ctx)
}

func explicitArtistRasterizationOptions(art Artist, ctx *DrawContext) (render.Rasterization, bool) {
	if provider, ok := art.(rasterizationProvider); ok {
		options := provider.Rasterization()
		switch options.Mode {
		case render.RasterizeAlways, render.RasterizeAuto:
			return withRasterizationDPI(options, ctx), true
		}
	}
	return render.Rasterization{}, false
}

func artistRasterizationOptions(r render.Renderer, art Artist, ctx *DrawContext) (render.Rasterization, bool) {
	if options, ok := explicitArtistRasterizationOptions(art, ctx); ok {
		return options, true
	}
	if provider, ok := art.(rasterizationProvider); ok && provider.Rasterization().Mode == render.RasterizeNever {
		return render.Rasterization{}, false
	}

	if !shouldAutoRasterizeArtist(r, art) {
		return render.Rasterization{}, false
	}
	return withRasterizationDPI(render.Rasterization{Mode: render.RasterizeAuto}, ctx), true
}

func withRasterizationDPI(options render.Rasterization, ctx *DrawContext) render.Rasterization {
	if options.DPI <= 0 && ctx != nil && ctx.RC.DPI > 0 {
		options.DPI = ctx.RC.DPI
	}
	return options
}

func shouldAutoRasterizeArtist(r render.Renderer, art Artist) bool {
	if _, ok := r.(render.RasterizationController); !ok {
		return false
	}
	if denseScatterArtist(art) {
		return true
	}
	if denseContourArtist(art) {
		return true
	}
	if hasUnsupportedFilterPathEffect(r, art) {
		return true
	}
	return false
}

func denseScatterArtist(art Artist) bool {
	scatter, ok := art.(*Scatter2D)
	return ok && scatter != nil && len(scatter.XY) >= autoRasterizeScatterPointThreshold
}

func denseContourArtist(art Artist) bool {
	contour, ok := art.(*ContourSet)
	if !ok || contour == nil {
		return false
	}
	return contourPathCount(contour) >= autoRasterizeContourPathThreshold
}

func contourPathCount(contour *ContourSet) int {
	if contour == nil {
		return 0
	}
	count := 0
	if contour.Lines != nil {
		count += len(contour.Lines.Segments)
	}
	if contour.Fills != nil {
		count += len(contour.Fills.Polygons)
		count += len(contour.Fills.Paths)
	}
	return count
}

func hasUnsupportedFilterPathEffect(r render.Renderer, art Artist) bool {
	if _, ok := r.(render.FilterRenderer); ok {
		return false
	}
	provider, ok := art.(pathEffectRasterizationProvider)
	if !ok {
		return false
	}
	for _, effect := range provider.rasterizationPathEffects() {
		if effect.Kind == render.PathEffectFilter {
			return true
		}
	}
	return false
}

func (l *Line2D) rasterizationPathEffects() []render.PathEffect {
	if l == nil {
		return nil
	}
	return l.PathEffects
}

func (s *Scatter2D) rasterizationPathEffects() []render.PathEffect {
	if s == nil {
		return nil
	}
	return s.PathEffects
}

func (c *Collection) rasterizationPathEffects() []render.PathEffect {
	if c == nil {
		return nil
	}
	return c.PathEffects
}

func (p *Patch) rasterizationPathEffects() []render.PathEffect {
	if p == nil {
		return nil
	}
	return p.PathEffects
}
