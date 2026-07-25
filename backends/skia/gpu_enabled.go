//go:build skiagpu

package skia

// gpuBuildEnabled reports whether the skiagpu build tag is set. When true the
// renderer accepts GPU-mode configuration and reports the GPU render mode.
// Combined with skiacgo it creates a native Ganesh surface over EGL/OpenGL;
// without skiacgo it retains the deterministic CPU fallback.
//
// The flag is keyed on skiagpu alone (not skia) so the build-tag-free strategy.go
// can report GPU status in every build configuration; the GPU renderer itself
// lives in skia.go and still requires the skia tag.
const gpuBuildEnabled = true
