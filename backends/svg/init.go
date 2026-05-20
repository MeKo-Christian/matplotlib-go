package svg

import (
	"github.com/cwbudde/matplotlib-go/backends"
	"github.com/cwbudde/matplotlib-go/render"
)

func init() {
	backends.Register(backends.SVG, &backends.BackendInfo{
		Name:        "SVG",
		Description: "Pure Go SVG backend with path recording and native text output",
		Capabilities: []backends.Capability{
			backends.AntiAliasing,
			backends.VectorOutput,
			backends.PathClip,
			backends.PathEffects,
			backends.ClipPathTransform,
			backends.TextShaping,
			backends.RotatedText,
			backends.VerticalText,
			backends.TextPathing,
			backends.ImageTransform,
			backends.MarkerBatch,
			backends.PathCollectionBatch,
			backends.NativeHatcher,
			backends.MixedRasterVector,
			backends.SVGExport,
			backends.SVGOptionExport,
			backends.TeXMetrics,
			backends.TeXText,
			backends.RotatedTeX,
			backends.DPIAware,
		},
		FallbackCapabilities: []backends.Capability{
			backends.QuadMeshBatch,
		},
		SaveFormats: map[string]backends.SaveHandler{
			".svg": backends.SaveSVG,
		},
		Factory: func(config backends.Config) (render.Renderer, error) {
			return New(config.Width, config.Height, config.Background)
		},
		Available: true,
	})
}
