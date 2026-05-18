package core

import (
	"errors"

	"github.com/cwbudde/matplotlib-go/render"
)

// SavePDF saves a figure to a PDF file using the provided renderer.
func SavePDF(fig *Figure, r render.Renderer, path string, _ ...render.SVGOption) error {
	if fig == nil {
		return errors.New("savepdf: nil figure")
	}
	DrawFigure(fig, r)
	exporter, ok := r.(render.PDFExporter)
	if !ok {
		return errors.New("PDF export not supported for this renderer type")
	}
	return exporter.SavePDF(path)
}
