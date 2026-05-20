package core

import "github.com/cwbudde/matplotlib-go/render"

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

func drawArtist(r render.Renderer, ctx *DrawContext, art Artist) {
	if art == nil {
		return
	}
	options, ok := artistRasterizationOptions(art, ctx)
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
	options, ok := artistRasterizationOptions(art, ctx)
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

func artistRasterizationOptions(art Artist, ctx *DrawContext) (render.Rasterization, bool) {
	provider, ok := art.(rasterizationProvider)
	if !ok {
		return render.Rasterization{}, false
	}
	options := provider.Rasterization()
	switch options.Mode {
	case render.RasterizeAlways, render.RasterizeAuto:
	default:
		return render.Rasterization{}, false
	}
	if options.DPI <= 0 && ctx != nil && ctx.RC.DPI > 0 {
		options.DPI = ctx.RC.DPI
	}
	return options, true
}
