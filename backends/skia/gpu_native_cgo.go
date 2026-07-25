// The skiagpu+skiacgo tier compiles the optional Ganesh/EGL path in
// skia_cwrap.cpp. These flags are deliberately isolated from ordinary
// skiacgo builds so CPU-native users do not acquire OpenGL/EGL dependencies.

//go:build skia && skiagpu && skiacgo

package skia

// #cgo CXXFLAGS: -DMGSK_ENABLE_GPU
// #cgo LDFLAGS: -lEGL -lGL -ldl
import "C"

const gpuNativeBuildEnabled = true
