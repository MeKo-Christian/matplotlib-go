package skia

import (
	"github.com/cwbudde/matplotlib-go/backends"
	"github.com/cwbudde/matplotlib-go/render"
)

func init() {
	available := isAvailable()
	capabilities := []backends.Capability(nil)
	fallbackCapabilities := []backends.Capability(nil)
	saveFormats := map[string]backends.SaveHandler(nil)
	if available {
		capabilities = []backends.Capability{
			backends.AntiAliasing,
			backends.PathClip,
			backends.TextShaping,
			backends.DPIAware,
			backends.TextPathing,
			backends.RotatedText,
			backends.VerticalText,
			backends.PathEffects,
			backends.RGBABuffer,
			backends.PNGExport,
		}
		fallbackCapabilities = []backends.Capability{
			backends.MarkerBatch,
			backends.PathCollectionBatch,
			backends.QuadMeshBatch,
			backends.NativeHatcher,
		}
		saveFormats = map[string]backends.SaveHandler{
			".png": backends.SavePNG,
		}
	}

	// Register Skia backend with the global registry
	backends.Register(backends.Skia, &backends.BackendInfo{
		Name:                 "Skia",
		Description:          "Opt-in Skia-tagged CPU raster backend; GPU and Skia-native optional paths are deferred",
		Capabilities:         capabilities,
		FallbackCapabilities: fallbackCapabilities,
		SaveFormats:          saveFormats,
		Factory: func(config backends.Config) (render.Renderer, error) {
			return New(config)
		},
		Available: available,
	})
}

// isAvailable checks if Skia dependencies are available at runtime.
func isAvailable() bool {
	return buildTagAvailable()
}
