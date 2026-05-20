package core

import (
	"errors"

	"github.com/cwbudde/matplotlib-go/render"
)

// SaveSVG saves a figure to an SVG file using the provided renderer.
// This function draws the figure using the renderer and then exports to SVG.
func SaveSVG(fig *Figure, r render.Renderer, path string, opts ...render.SaveOption) error {
	saveOptions := render.ResolveSaveOptions(opts...)
	if err := saveOptions.ValidateForExtension(".svg"); err != nil {
		return err
	}
	svgOptions := saveOptions.SVG
	if setter, ok := r.(render.SVGOptionSetter); ok {
		setter.SetSVGOptions(svgOptions)
	}

	// Draw the figure using the renderer.
	DrawFigure(fig, r)

	// Check if this renderer supports SVG export.
	if exporter, ok := r.(render.SVGOptionExporter); ok {
		return exporter.SaveSVGWithOptions(path, svgOptions)
	}
	if exporter, ok := r.(render.SVGExporter); ok {
		return exporter.SaveSVG(path)
	}

	// For other renderer types, we don't have SVG export support yet.
	return errors.New("SVG export not supported for this renderer type")
}
