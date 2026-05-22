// Package skia is the opt-in Skia renderer backend for matplotlib-go.
//
// # Strategy
//
// Skia is developed under the backend-deepening and renderer-effects tracks in
// PLAN.md. The chosen integration strategy is recorded by BackendStrategy:
//   - build tag: skia
//   - binding: a small external C ABI wrapper around Skia for future native work
//   - first implementation target: CPU raster output through a Skia-local
//     bridge boundary, with the shared raster surface still used for fallback
//     drawing paths
//   - GPU mode: deferred until the CPU path and capability reporting are stable
//   - default CI: non-Skia stub builds; skia-tagged tests are gated
//
// The package deliberately does not depend on an unstable Go Skia binding. The
// current skia-tagged CPU renderer delegates fallback drawing to the shared
// pure-Go raster surface for paths, images, clipping, text, RGBA access, and
// PNG export, while shader fills route through a Skia-local surface bridge. The
// future external Skia integration should call a narrow C ABI that this
// repository controls, keeping build failures and platform policy localized to
// the skia package.
//
// # Dependencies
//
// The current skia-tagged CPU compatibility renderer has no external runtime
// dependency beyond normal Go builds. Future native Skia integration will
// require CGO_ENABLED=1, a Skia shared library, and the matplotlib-go C ABI
// wrapper library. GPU mode will additionally require platform-specific
// graphics drivers and context setup.
//
// # Current Status
//
// Default builds compile the unavailable stub in skia_stub.go and advertise no
// capabilities or save formats. Builds with -tags skia register an available
// CPU renderer for static raster output and PNG save dispatch. The CPU bridge
// consumes renderer-neutral pattern and gradient fills, including linear and
// radial gradients, stop opacity, transformed fills, and tiled pattern fills.
// GPU mode and external Skia-native batch drawing remain unavailable.
//
// # Configuration
//
// Use SkiaConfig to configure color formats and quality settings. UseGPU is
// reserved and currently returns an error because GPU mode is deferred:
//
//	config := backends.Config{
//		Width: 800, Height: 600,
//		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
//		Options: backends.SkiaConfig{
//			SampleCount: 4,
//		},
//	}
package skia
