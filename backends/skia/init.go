package skia

import (
	"matplotlib-go/backends"
	"matplotlib-go/render"
)

func init() {
	// Register Skia backend with the global registry
	backends.Register(backends.Skia, &backends.BackendInfo{
		Name:        "Skia",
		Description: "High-quality anti-aliased renderer with optional GPU acceleration",
		Capabilities: []backends.Capability{
			backends.AntiAliasing,
			backends.SubPixel,
			backends.GradientFill,
			backends.PathClip,
			backends.GPUAccel,
			backends.VectorOutput,
			backends.TextShaping,
			backends.FontHinting,
		},
		Factory: func(config backends.Config) (render.Renderer, error) {
			return New(config)
		},
		Available: isAvailable(),
	})
}

// isAvailable checks if Skia dependencies are available at runtime.
func isAvailable() bool {
	// TODO: Check for Skia shared library
	// TODO: Check for required graphics drivers
	// For now, return false since it's not implemented
	return false
}