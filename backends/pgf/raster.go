package pgf

import (
	"github.com/cwbudde/matplotlib-go/backends/internal/mixedraster"
	"github.com/cwbudde/matplotlib-go/render"
)

// StartRasterized begins a transparent offscreen raster group for mixed output.
func (r *Renderer) StartRasterized(options render.Rasterization) bool {
	if r == nil || !r.began || r.raster != nil {
		return false
	}
	session, ok := mixedraster.Start(r.width, r.height, r.viewport, options, r.resolution, r.clipRect, r.clipPaths)
	if !ok {
		return false
	}
	r.raster = session
	return true
}

// StopRasterized embeds the active raster group as self-contained PGF pixels.
func (r *Renderer) StopRasterized() bool {
	if r == nil || r.raster == nil {
		return false
	}
	session := r.raster
	r.raster = nil
	img, rect, ok := session.Stop()
	if !ok {
		return false
	}
	r.DrawImage(img, rect)
	return true
}

func (r *Renderer) activeRaster() render.Renderer {
	if r == nil || r.raster == nil {
		return nil
	}
	return r.raster.Renderer()
}
