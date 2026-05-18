package core

import (
	"errors"

	"github.com/cwbudde/matplotlib-go/render"
)

// SavePS saves a figure to a PostScript or EPS file using the provided
// renderer.
func SavePS(fig *Figure, r render.Renderer, path string, _ ...render.SVGOption) error {
	if fig == nil {
		return errors.New("saveps: nil figure")
	}
	DrawFigure(fig, r)
	exporter, ok := r.(render.PSExporter)
	if !ok {
		return errors.New("PostScript export not supported for this renderer type")
	}
	return exporter.SavePS(path)
}
