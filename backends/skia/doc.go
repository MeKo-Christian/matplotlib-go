// Package skia is the opt-in Skia renderer backend for matplotlib-go.
//
// # Strategy
//
// Skia is developed under the backend-deepening and renderer-effects tracks in
// PLAN.md. The chosen integration strategy is recorded by BackendStrategy:
//   - build tag: skia
//   - binding: a small external C ABI wrapper around Skia
//   - CPU mode: native raster SkSurface primitives under skiacgo, with the
//     shared raster surface retained for unsupported fallback paths
//   - GPU mode: a Ganesh render target over headless EGL/OpenGL under
//     skiagpu+skiacgo, with deterministic RGBA readback
//   - default CI: non-Skia stub builds; skia-tagged tests are gated
//
// The package deliberately does not depend on an unstable Go Skia binding. The
// current skia-tagged CPU renderer delegates fallback drawing to the shared
// pure-Go raster surface for fallback paths, clipping, text, RGBA access, and
// PNG export. Native fills, transformed images, hatches, and batches call a
// narrow C ABI controlled by this repository, keeping build failures and
// platform policy localized to the skia package.
//
// # Dependencies
//
// The skia-tagged CPU compatibility renderer has no external runtime dependency
// beyond normal Go builds. The native cgo backend (-tags "skia skiacgo")
// additionally requires CGO_ENABLED=1, a built Skia shared library, and the
// matplotlib-go C ABI wrapper (skia_cwrap.h / skia_cwrap.cpp, compiled in this
// package). Native GPU mode additionally requires EGL, OpenGL, Ganesh-enabled
// Skia, and a usable platform graphics driver.
//
// # Native cgo backend (skiacgo)
//
// The skiacgo build tag links a real Skia library through the narrow C ABI in
// skia_cwrap.h. The Skia include/library locations are supplied at build time
// via CGO_CXXFLAGS / CGO_LDFLAGS; the `build-skia-native` / `test-skia-native`
// just recipes wire them from SKIA_ROOT. With this tag the surfaceBridge is
// backed by a native SkSurface: gradient path fills (SkShaders gradients),
// transformed RGBA images (SkImage), tiled-shader hatches (SkShader), marker
// batches and path collections (SkCanvas/SkPath), quad meshes, and Gouraud
// triangles (SkVertices) render natively, so IsCapabilityBridged reports
// ImageTransform, NativeHatcher, MarkerBatch, PathCollectionBatch, and
// QuadMeshBatch as native.
//
// # Native GPU backend (skiagpu + skiacgo)
//
// The combined build tier exposes mgsk_surface_new_gpu and creates a Ganesh
// GrDirectContext over a headless EGL/OpenGL context. Skia draws into
// SkSurfaces::RenderTarget, FlushGPU submits queued work, and readPixels returns
// deterministic straight-alpha RGBA for compositing and golden tests. If EGL or
// the GPU context is unavailable, renderer construction succeeds with a native
// CPU raster fallback; BackendComparisonReport marks gpuaccel as fallback
// instead of claiming acceleration.
//
// FreeType caveat: Skia statically bundles its own FreeType when built without
// system FreeType. The native Skia backend uses Skia only for geometry (no
// text), so it does not link FreeType itself — but combining the `skiacgo` tag
// with the `freetype` tag (agg's vendored FreeType 2.6.1) in one binary
// produces duplicate FT_* symbols and a runtime crash. To use native Skia and
// agg native-FreeType text together, rebuild Skia with skia_use_freetype=false.
//
// # Current Status
//
// Default builds compile the unavailable stub in skia_stub.go and advertise no
// capabilities or save formats. Builds with -tags skia register an available
// CPU renderer for static raster output and PNG save dispatch. Builds with
// -tags "skia skiagpu" also accept SkiaConfig.UseGPU but use the deterministic
// CPU fallback without skiacgo. The CPU bridge
// consumes renderer-neutral pattern and gradient fills, including linear and
// radial gradients, stop opacity, transformed fills, and tiled pattern fills.
// Builds with -tags "skia skiacgo" link a real Skia library and render gradient
// fills, transformed-image blits, tiled-shader hatches, marker batches, path
// collections, quad meshes, and Gouraud triangles natively (see the native cgo
// section above). Builds with -tags "skia skiagpu skiacgo" add real Ganesh GPU
// surfaces while retaining deterministic CPU readback.
//
// # Configuration
//
// Use SkiaConfig to configure color formats and quality settings. UseGPU returns
// an error under plain -tags skia, selects CPU fallback under
// -tags "skia skiagpu", and attempts native acceleration under
// -tags "skia skiagpu skiacgo":
//
//	config := backends.Config{
//		Width: 800, Height: 600,
//		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
//		Options: backends.SkiaConfig{
//			SampleCount: 4,
//		},
//	}
package skia
