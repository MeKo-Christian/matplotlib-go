package ps

import (
	"github.com/cwbudde/matplotlib-go/backends"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/render"
)

func init() {
	factory := func(width, height int, background render.Color) (render.Renderer, error) {
		return New(width, height, background)
	}
	core.RegisterFigureOutputRenderer(".ps", factory)
	core.RegisterFigureOutputRenderer(".eps", factory)
	backends.Register(backends.PS, &backends.BackendInfo{
		Name:        "PostScript",
		Description: "Pure Go Level-2 PostScript/EPS backend with deterministic serialization",
		Capabilities: []backends.Capability{
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
			backends.MarkerBatch,
			backends.NativeHatcher,
			backends.PathCollectionBatch,
			backends.MixedRasterVector,
			backends.PSExport,
			backends.DPIAware,
		},
		SaveFormats: map[string]backends.SaveHandler{
			".ps":  backends.SavePS,
			".eps": backends.SavePS,
		},
		Factory: func(config backends.Config) (render.Renderer, error) {
			return New(config.Width, config.Height, config.Background)
		},
		Available: true,
	})
}
