package core

import (
	"errors"

	"github.com/cwbudde/matplotlib-go/render"
)

// SavePGF saves a figure to a PGF/TikZ file using the provided renderer.
func SavePGF(fig *Figure, r render.Renderer, path string, opts ...render.SaveOption) error {
	if fig == nil {
		return errors.New("savepgf: nil figure")
	}
	saveOptions := render.ResolveSaveOptions(opts...)
	if err := saveOptions.ValidateForExtension(".pgf"); err != nil {
		return err
	}
	eff, drawOpts, resolved, rendererDPI := prepareVectorSaveFigure(fig, r, &saveOptions.Figure)
	if err := rejectTightBboxForVector(resolved.bboxTight, "PGF"); err != nil {
		return err
	}
	if setter, ok := r.(render.PGFOptionSetter); ok {
		setter.SetPGFOptions(saveOptions.PGF)
	}
	drawFigureWithOptionsAtDPI(eff, r, drawOpts, rendererDPI)
	if exporter, ok := r.(render.PGFOptionExporter); ok {
		return exporter.SavePGFWithOptions(path, saveOptions.PGF)
	}
	exporter, ok := r.(render.PGFExporter)
	if !ok {
		return errors.New("PGF export not supported for this renderer type")
	}
	return exporter.SavePGF(path)
}
