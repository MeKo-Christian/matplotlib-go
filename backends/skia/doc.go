// Package skia provides a Skia-based renderer backend for matplotlib-go.
//
// This backend uses Skia graphics library for high-quality anti-aliased
// rendering with optional GPU acceleration.
//
// # Requirements
//
// This backend requires CGO and Skia C++ library bindings. Currently supports:
//   - go-skia (github.com/go-gl/skia) - Pure Go bindings
//   - skia-safe (via FFI) - Rust-based bindings with Go wrapper
//
// # Build Tags
//
// Use build tags to enable Skia support:
//   go build -tags skia ./...
//
// # Dependencies
//
// The Skia backend requires:
//   - Skia shared library (.so/.dylib/.dll)
//   - CGO_ENABLED=1
//   - Platform-specific graphics drivers for GPU acceleration
//
// # Capabilities
//
// The Skia backend provides:
//   - High-quality anti-aliasing
//   - GPU acceleration (optional)
//   - Advanced text shaping
//   - Path clipping support
//   - Gradient fills
//   - Multiple output formats (PNG, PDF, SVG)
//
// # Configuration
//
// Use SkiaConfig to configure GPU usage, color formats, and quality settings:
//
//	config := backends.Config{
//		Width: 800, Height: 600,
//		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
//		Options: backends.SkiaConfig{
//			UseGPU: true,
//			SampleCount: 4, // 4x MSAA
//		},
//	}
//
// # Status
//
// PHASE B: Stub implementation. Skia integration planned for later phases
// when performance or quality demands exceed GoBasic capabilities.
package skia