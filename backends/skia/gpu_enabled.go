//go:build skia && skiagpu

package skia

// gpuBuildEnabled reports whether the skiagpu build tag is set. When true the
// renderer accepts GPU-mode configuration and reports the GPU render mode, but
// the underlying surface is still the deterministic CPU readback bridge: this is
// scaffolding for a future native SkSurface::MakeRenderTarget path, not real GPU
// acceleration. See backends/skia/strategy.go.
const gpuBuildEnabled = true
