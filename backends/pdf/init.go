package pdf

import (
	"github.com/cwbudde/matplotlib-go/backends"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/render"
)

func init() {
	core.RegisterFigureOutputRenderer(".pdf", func(width, height int, background render.Color) (render.Renderer, error) {
		return New(width, height, background)
	})
	backends.Register(backends.PDF, &backends.BackendInfo{
		Name:        "PDF",
		Description: "Pure Go PDF backend with deterministic serialization and embedded-font text output",
		Capabilities: []backends.Capability{
			backends.AntiAliasing,
			backends.VectorOutput,
			backends.PathClip,
			backends.PathEffects,
			backends.TextShaping,
			backends.TextPathing,
			backends.RotatedText,
			backends.VerticalText,
			backends.FontKeyText,
			backends.FontKeyRotatedText,
			backends.FontKeyVerticalText,
			backends.ImageTransform,
			backends.NativeHatcher,
			backends.PatternFill,
			backends.GradientFill,
			backends.MarkerBatch,
			backends.PathCollectionBatch,
			backends.MixedRasterVector,
			backends.PDFExport,
			backends.DPIAware,
		},
		SaveFormats: map[string]backends.SaveHandler{
			".pdf": backends.SavePDF,
		},
		Factory: func(config backends.Config) (render.Renderer, error) {
			return New(config.Width, config.Height, config.Background)
		},
		Available: true,
	})
}
