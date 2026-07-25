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
		// These are the capabilities the skia-tagged tier genuinely provides.
		// In the pure-Go build (no skiacgo) the batch/transform/hatch entries are
		// satisfied through the embedded gobasic CPU surface, so they report as
		// CapabilityBridged (not native Skia) via RendererCapabilityStatus; only
		// the skiacgo build flips them to native. GPU acceleration is
		// mode- and runtime-specific, so RuntimeCapabilityStatus reports it
		// instead of placing it on this shared registry declaration.
		capabilities = []backends.Capability{
			backends.AntiAliasing,
			backends.PatternFill,
			backends.GradientFill,
			backends.PathClip,
			backends.TextShaping,
			backends.DPIAware,
			backends.TextPathing,
			backends.RotatedText,
			backends.VerticalText,
			backends.ImageTransform,
			backends.OffscreenFilter,
			backends.PathEffects,
			backends.RGBABuffer,
			backends.PNGExport,
			backends.NativeHatcher,
			backends.MarkerBatch,
			backends.PathCollectionBatch,
			backends.QuadMeshBatch,
			backends.GouraudTriangleBatch,
		}
		saveFormats = map[string]backends.SaveHandler{
			".png": backends.SavePNG,
		}
	}

	// Register Skia backend with the global registry
	backends.Register(backends.Skia, &backends.BackendInfo{
		Name:                 "Skia",
		Description:          "Opt-in Skia raster backend with native skiacgo CPU primitives and optional Ganesh GPU surfaces",
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
