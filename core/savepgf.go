package core

import (
	"errors"

	"github.com/cwbudde/matplotlib-go/render"
)

// SavePGF saves a figure to a PGF/TikZ file using the provided renderer.
func SavePGF(fig *Figure, r render.Renderer, path string, _ ...render.SVGOption) error {
	if fig == nil {
		return errors.New("savepgf: nil figure")
	}
	DrawFigure(fig, r)
	exporter, ok := r.(render.PGFExporter)
	if !ok {
		return errors.New("PGF export not supported for this renderer type")
	}
	return exporter.SavePGF(path)
}
