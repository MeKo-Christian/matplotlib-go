//go:build skia && !js

package webdemo

import (
	"fmt"

	"github.com/cwbudde/matplotlib-go/backends"
	"github.com/cwbudde/matplotlib-go/backends/skia"
	"github.com/cwbudde/matplotlib-go/render"
)

func optionalBackendDescriptors() []BackendDescriptor {
	return []BackendDescriptor{
		{
			ID:          "skia",
			Name:        "Skia",
			Description: "Skia-tagged CPU compatibility raster backend for parity comparisons.",
		},
	}
}

func newOptionalRasterRenderer(backendID string, width, height int, bg render.Color) (rasterRenderer, bool, error) {
	if backendID != "skia" {
		return nil, false, nil
	}
	r, err := skia.New(backends.Config{
		Width:      width,
		Height:     height,
		Background: bg,
		Options:    backends.SkiaConfig{UseGPU: false},
	})
	if err != nil {
		return nil, true, fmt.Errorf("webdemo: create skia renderer: %w", err)
	}
	return r, true, nil
}
