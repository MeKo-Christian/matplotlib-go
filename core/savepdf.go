package core

import (
	"errors"

	"github.com/cwbudde/matplotlib-go/render"
)

// SavePDF saves a figure to a PDF file using the provided renderer.
func SavePDF(fig *Figure, r render.Renderer, path string, opts ...render.SaveOption) error {
	if fig == nil {
		return errors.New("savepdf: nil figure")
	}
	// Seed the font policy from pdf.* rcParams; explicit per-call options win.
	opts = append([]render.SaveOption{render.WithPDFFontPolicy(pdfFontPolicyFromRC(fig.RC.PDF))}, opts...)
	saveOptions := render.ResolveSaveOptions(opts...)
	if err := saveOptions.ValidateForExtension(".pdf"); err != nil {
		return err
	}
	eff, drawOpts, resolved := prepareSaveFigure(fig, r, &saveOptions.Figure)
	if err := rejectTightBboxForVector(resolved.bboxTight, "PDF"); err != nil {
		return err
	}
	if setter, ok := r.(render.PDFOptionSetter); ok {
		setter.SetPDFOptions(saveOptions.PDF)
	}
	DrawFigureWithOptions(eff, r, drawOpts)
	if exporter, ok := r.(render.PDFOptionExporter); ok {
		return exporter.SavePDFWithOptions(path, saveOptions.PDF)
	}
	exporter, ok := r.(render.PDFExporter)
	if !ok {
		return errors.New("PDF export not supported for this renderer type")
	}
	return exporter.SavePDF(path)
}
