package agg

import (
	"github.com/cwbudde/matplotlib-go/backends"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/render"
)

func init() {
	core.RegisterFigureOutputRenderer(".png", func(width, height int, background render.Color) (render.Renderer, error) {
		return New(width, height, background)
	})
	backends.Register(backends.AGG, &backends.BackendInfo{
		Name:        "AGG",
		Description: "Anti-Grain Geometry renderer with high-quality anti-aliasing",
		Capabilities: []backends.Capability{
			backends.AntiAliasing,
			backends.SubPixel,
			backends.PathClip,
			backends.DPIAware,
			backends.TextShaping,
			backends.FontHinting,
			backends.TextBounds,
			backends.TextPathing,
			backends.RotatedText,
			backends.VerticalText,
			backends.FontKeyText,
			backends.FontKeyRotatedText,
			backends.FontKeyVerticalText,
			backends.ImageTransform,
			backends.RGBABuffer,
			backends.BufferRegion,
			backends.OffscreenFilter,
			backends.PatternFill,
			backends.GradientFill,
			backends.PathEffects,
			backends.NativeHatcher,
			backends.PNGExport,
			backends.MarkerBatch,
			backends.PathCollectionBatch,
			backends.QuadMeshBatch,
			backends.GouraudTriangleBatch,
		},
		SaveFormats: map[string]backends.SaveHandler{
			".png": backends.SavePNG,
		},
		Factory: func(config backends.Config) (render.Renderer, error) {
			return New(config.Width, config.Height, config.Background)
		},
		Available: true,
	})
}
