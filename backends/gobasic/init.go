package gobasic

import (
	"matplotlib-go/backends"
	"matplotlib-go/render"
)

func init() {
	// Register GoBasic backend with the global registry
	backends.Register(backends.GoBasic, &backends.BackendInfo{
		Name:        "GoBasic",
		Description: "Pure Go renderer using golang.org/x/image/vector",
		Capabilities: []backends.Capability{
			backends.AntiAliasing, // Basic AA via vector rasterizer
			backends.PathClip,     // Rectangular clipping implemented
			backends.VectorOutput, // Can generate vector-like output
		},
		Factory: func(config backends.Config) (render.Renderer, error) {
			return New(config.Width, config.Height, config.Background), nil
		},
		Available: true, // Always available - pure Go
	})
}