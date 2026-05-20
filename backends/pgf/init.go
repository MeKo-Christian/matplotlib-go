package pgf

import (
	"github.com/cwbudde/matplotlib-go/backends"
	"github.com/cwbudde/matplotlib-go/render"
)

func init() {
	backends.Register(backends.PGF, &backends.BackendInfo{
		Name:        "PGF",
		Description: "Generator-only PGF/TikZ backend for LaTeX inclusion",
		Capabilities: []backends.Capability{
			backends.VectorOutput,
			backends.PathClip,
			backends.TextShaping,
			backends.RotatedText,
			backends.FontKeyText,
			backends.FontKeyRotatedText,
			backends.ImageTransform,
			backends.NativeHatcher,
			backends.MarkerBatch,
			backends.PathCollectionBatch,
			backends.PathEffects,
			backends.MixedRasterVector,
			backends.PGFExport,
			backends.DPIAware,
		},
		SaveFormats: map[string]backends.SaveHandler{
			".pgf": backends.SavePGF,
		},
		Factory: func(config backends.Config) (render.Renderer, error) {
			return New(config.Width, config.Height, config.Background)
		},
		Available: true,
	})
}
