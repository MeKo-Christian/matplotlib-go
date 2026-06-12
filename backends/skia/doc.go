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
//   - GPU mode: a skiagpu build-tag scaffold that selects GPU mode but still
//     renders through deterministic CPU readback until native SkSurface support
//     lands
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
// The skia-tagged CPU compatibility renderer has no external runtime dependency
// beyond normal Go builds. The native cgo backend (-tags "skia skiacgo")
// additionally requires CGO_ENABLED=1, a built Skia shared library, and the
// matplotlib-go C ABI wrapper (skia_cwrap.h / skia_cwrap.cpp, compiled in this
// package). GPU mode will additionally require platform-specific graphics
// drivers and context setup.
//
// # Native cgo backend (skiacgo)
//
// The skiacgo build tag links a real Skia library through the narrow C ABI in
// skia_cwrap.h. The Skia include/library locations are supplied at build time
// via CGO_CXXFLAGS / CGO_LDFLAGS; the `build-skia-native` / `test-skia-native`
// just recipes wire them from SKIA_ROOT. With this tag the surfaceBridge is
// backed by a native SkSurface: gradient path fills (SkShaders gradients),
// transformed RGBA images (SkImage), marker batches and path collections
// (SkCanvas/SkPath), quad meshes, and Gouraud triangles (SkVertices) render
// natively, so IsCapabilityBridged reports ImageTransform, MarkerBatch,
// PathCollectionBatch, and QuadMeshBatch as native. Hatching still routes
// through the CPU bridge for now.
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
// -tags "skia skiagpu" also accept SkiaConfig.UseGPU and report the GPU scaffold
// mode, but still render through deterministic CPU readback. The CPU bridge
// consumes renderer-neutral pattern and gradient fills, including linear and
// radial gradients, stop opacity, transformed fills, and tiled pattern fills.
// Builds with -tags "skia skiacgo" link a real Skia library and render gradient
// fills, transformed-image blits, marker batches, path collections, quad
// meshes, and Gouraud triangles natively (see the native cgo section above).
// Native GPU surfaces and tiled-shader hatching are still unavailable;
// NativePathRequirements records implemented and deferred external primitives.
//
// # Configuration
//
// Use SkiaConfig to configure color formats and quality settings. UseGPU returns
// an error under plain -tags skia and selects the CPU-readback GPU scaffold under
// -tags "skia skiagpu":
//
//	config := backends.Config{
//		Width: 800, Height: 600,
//		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
//		Options: backends.SkiaConfig{
//			SampleCount: 4,
//		},
//	}
package skia
