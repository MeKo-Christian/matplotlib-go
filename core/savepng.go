package core

import (
	"errors"

	"matplotlib-go/backends/gobasic"
	"matplotlib-go/render"
)

// PNGExporter defines the interface for renderers that can export to PNG.
type PNGExporter interface {
	SavePNG(path string) error
}

// SavePNG saves a figure to a PNG file using the provided renderer.
// This function draws the figure using the renderer and then exports to PNG.
func SavePNG(fig *Figure, r render.Renderer, path string) error {
	// Draw the figure using the renderer
	DrawFigure(fig, r)

	// Check if this renderer supports PNG export
	if exporter, ok := r.(PNGExporter); ok {
		return exporter.SavePNG(path)
	}

	// Fallback: Check if this is a GoBasic renderer (for backwards compatibility)
	if gb, ok := r.(*gobasic.Renderer); ok {
		return gb.SavePNG(path)
	}

	// For other renderer types, we don't have PNG export support yet
	return errors.New("PNG export not supported for this renderer type")
}
