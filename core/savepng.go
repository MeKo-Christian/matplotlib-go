package core

import (
	"errors"

	"github.com/cwbudde/matplotlib-go/render"
)

// SavePNG saves a figure to a PNG file using the provided renderer.
// This function draws the figure using the renderer and then exports to PNG.
//
// The renderer must implement render.PNGExporter; the same interface contract
// used by the backends-registry SaveViaExtension dispatch.
func SavePNG(fig *Figure, r render.Renderer, path string, opts ...render.SaveOption) error {
	saveOptions := render.ResolveSaveOptions(opts...)
	if err := saveOptions.ValidateForExtension(".png"); err != nil {
		return err
	}
	DrawFigure(fig, r)

	exporter, ok := r.(render.PNGExporter)
	if !ok {
		return errors.New("PNG export not supported for this renderer type")
	}
	return exporter.SavePNG(path)
}
