package gobasic

import (
	"github.com/cwbudde/matplotlib-go/backends"
	"github.com/cwbudde/matplotlib-go/render"
)

func init() {
	// Register GoBasic backend with the global registry
	backends.Register(backends.GoBasic, &backends.BackendInfo{
		Name:        "GoBasic",
		Description: "Pure Go renderer using golang.org/x/image/vector",
		Capabilities: []backends.Capability{
			backends.AntiAliasing, // Basic AA via vector rasterizer
			backends.PathClip,
			backends.DPIAware,
			backends.TextShaping,
			backends.TextPathing,
			backends.RotatedText,
			backends.VerticalText,
			backends.ImageTransform,
			backends.PathEffects,
			backends.PNGExport,
		},
		FallbackCapabilities: []backends.Capability{
			backends.MarkerBatch,
			backends.PathCollectionBatch,
			backends.QuadMeshBatch,
			backends.NativeHatcher,
		},
		SaveFormats: map[string]backends.SaveHandler{
			".png": backends.SavePNG,
		},
		Factory: func(config backends.Config) (render.Renderer, error) {
			return New(config.Width, config.Height, config.Background), nil
		},
		Available: true, // Always available - pure Go
	})
}
