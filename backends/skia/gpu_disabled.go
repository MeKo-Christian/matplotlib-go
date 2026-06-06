//go:build !skiagpu

package skia

// gpuBuildEnabled reports whether the skiagpu build tag is set. Without it the
// renderer rejects GPU-mode configuration and only the CPU render mode exists.
const gpuBuildEnabled = false
