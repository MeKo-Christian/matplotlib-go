//go:build skia && !skiacgo

package skia

// selectSurfaceBridge returns the pure-Go CPU surface bridge. Without the
// skiacgo build tag there is no linked Skia library, so the deterministic CPU
// bridge owns all rasterization. The native variant lives in native_cgo.go.
func selectSurfaceBridge(width, height int, mode RenderMode) surfaceBridge {
	return selectSurfaceBridgeWithSamples(width, height, mode, 1)
}

func selectSurfaceBridgeWithSamples(width, height int, mode RenderMode, _ int) surfaceBridge {
	return newCPUSurfaceBridge(width, height, mode)
}
