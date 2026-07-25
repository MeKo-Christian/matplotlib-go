package skia

import "github.com/cwbudde/matplotlib-go/backends"

// BindingStrategy names the selected integration boundary for the Skia backend.
type BindingStrategy string

const (
	// BindingExternalCAPI means matplotlib-go will call a small, explicit C ABI
	// wrapper around Skia instead of depending on unstable Go bindings directly.
	BindingExternalCAPI BindingStrategy = "external-c-api"
)

// RenderMode describes a planned Skia rendering mode.
type RenderMode string

const (
	ModeCPU RenderMode = "cpu"
	ModeGPU RenderMode = "gpu"
)

// ImplementationStatus describes how far a strategy item has advanced.
type ImplementationStatus string

const (
	StatusImplemented ImplementationStatus = "implemented"
	StatusPlanned     ImplementationStatus = "planned"
	StatusDeferred    ImplementationStatus = "deferred"
)

// CIBackendPolicy describes which Skia path runs in default CI.
type CIBackendPolicy string

const (
	CIDefaultStub CIBackendPolicy = "stub"
)

// Strategy records the binding/build policy for the Skia backend.
type Strategy struct {
	BuildTag          string
	Binding           BindingStrategy
	DefaultMode       RenderMode
	CPUStatus         ImplementationStatus
	GPUStatus         ImplementationStatus
	CIDefault         CIBackendPolicy
	RequiredLibraries []string
}

// NativePathRequirement records one tracked Skia-native integration point.
// It is strategy metadata, not a runtime capability declaration.
type NativePathRequirement struct {
	Primitive           string
	Modes               []RenderMode
	Capabilities        []backends.Capability
	ExternalEntrypoints []string
	Status              ImplementationStatus
	BlockedBy           string
}

// BridgeInfo describes the concrete bridge used by one renderer instance.
type BridgeInfo struct {
	Binding       BindingStrategy
	Mode          RenderMode
	NativeSurface bool
	// Accelerated reports whether rasterization runs on a real GPU render
	// target (SkSurfaces::RenderTarget). It gates Renderer.GPU so requested GPU
	// mode remains distinguishable from successful hardware acceleration. Mode
	// may be ModeGPU while this is false when EGL/OpenGL is unavailable and the
	// renderer has selected its deterministic CPU fallback.
	Accelerated     bool
	SupportsShaders bool
	Description     string
}

// BackendStrategy returns the documented Skia integration strategy. It is kept
// as code so tests and docs can agree on the same build/dependency contract.
//
// The reported render mode and GPU status depend on the selected build tier:
// skiagpu makes GPU mode selectable, while skiagpu+skiacgo compiles the native
// Ganesh/EGL implementation. Runtime context availability remains visible
// through Renderer.GPU and BackendComparisonReport.
func BackendStrategy() Strategy {
	defaultMode := ModeCPU
	gpuStatus := StatusDeferred
	if gpuBuildEnabled {
		defaultMode = ModeGPU
		gpuStatus = StatusPlanned
	}
	if gpuNativeBuildEnabled {
		gpuStatus = StatusImplemented
	}
	return Strategy{
		BuildTag:    "skia",
		Binding:     BindingExternalCAPI,
		DefaultMode: defaultMode,
		CPUStatus:   StatusImplemented,
		GPUStatus:   gpuStatus,
		CIDefault:   CIDefaultStub,
		RequiredLibraries: []string{
			"none for the skia-tagged CPU compatibility renderer",
			"Skia shared library for skiacgo native paths",
			"EGL and OpenGL for skiagpu+skiacgo Ganesh surfaces",
			"CGO_ENABLED=1 for native paths",
		},
	}
}

// NativePathRequirements returns the explicit external Skia primitives tracked
// for native backend coverage.
func NativePathRequirements() []NativePathRequirement {
	const externalABIBlocker = "external Skia C-ABI wrapper and linked Skia library"
	return []NativePathRequirement{
		{
			Primitive:           "SkCanvas::drawPath native batches",
			Modes:               []RenderMode{ModeCPU, ModeGPU},
			Capabilities:        []backends.Capability{backends.MarkerBatch, backends.PathCollectionBatch},
			ExternalEntrypoints: []string{"mgsk_draw_markers", "mgsk_draw_path"},
			Status:              StatusImplemented,
		},
		{
			Primitive:           "SkVertices Gouraud triangles",
			Modes:               []RenderMode{ModeCPU, ModeGPU},
			Capabilities:        []backends.Capability{backends.GouraudTriangleBatch},
			ExternalEntrypoints: []string{"SkVertices::MakeCopy", "SkCanvas::drawVertices"},
			Status:              StatusImplemented,
		},
		{
			Primitive:           "SkVertices quad mesh cells",
			Modes:               []RenderMode{ModeCPU, ModeGPU},
			Capabilities:        []backends.Capability{backends.QuadMeshBatch},
			ExternalEntrypoints: []string{"SkVertices::MakeCopy", "SkCanvas::drawVertices"},
			Status:              StatusImplemented,
		},
		{
			Primitive:           "SkImage transformed blits",
			Modes:               []RenderMode{ModeCPU, ModeGPU},
			Capabilities:        []backends.Capability{backends.ImageTransform},
			ExternalEntrypoints: []string{"mgsk_draw_image", "SkImages::RasterFromPixmapCopy", "SkCanvas::drawImageRect"},
			Status:              StatusImplemented,
		},
		{
			Primitive:           "tiled SkShader",
			Modes:               []RenderMode{ModeCPU, ModeGPU},
			Capabilities:        []backends.Capability{backends.NativeHatcher},
			ExternalEntrypoints: []string{"mgsk_draw_hatch_path", "SkImage::makeShader", "SkCanvas::drawPath"},
			Status:              StatusImplemented,
		},
		{
			Primitive:           "SkSurface::MakeRenderTarget",
			Modes:               []RenderMode{ModeGPU},
			Capabilities:        []backends.Capability{backends.GPUAccel},
			ExternalEntrypoints: []string{"mgsk_surface_new_gpu", "SkSurfaces::RenderTarget", "GrDirectContext::flushAndSubmit", "SkSurface::readPixels"},
			Status:              StatusImplemented,
		},
	}
}

// ModeCapabilities returns the optional renderer capabilities targeted by a Skia
// render mode. It encodes the CPU-vs-GPU capability split as data so the
// comparison report and tests can reason about per-mode support before a second
// (GPU) registry entry becomes meaningful.
//
// GPU mode adds GPUAccel. Its concrete runtime status is native when a Ganesh
// render target was created, fallback when GPU mode was requested but EGL/GL is
// unavailable, and unsupported in CPU mode.
func ModeCapabilities(mode RenderMode) []backends.Capability {
	base := []backends.Capability{
		backends.AntiAliasing,
		backends.PatternFill,
		backends.GradientFill,
		backends.PathClip,
		backends.TextShaping,
		backends.DPIAware,
		backends.TextPathing,
		backends.RotatedText,
		backends.VerticalText,
		backends.ImageTransform,
		backends.OffscreenFilter,
		backends.PathEffects,
		backends.RGBABuffer,
		backends.PNGExport,
		backends.NativeHatcher,
		backends.MarkerBatch,
		backends.PathCollectionBatch,
		backends.QuadMeshBatch,
		backends.GouraudTriangleBatch,
	}
	switch mode {
	case ModeGPU:
		gpu := make([]backends.Capability, len(base), len(base)+1)
		copy(gpu, base)
		gpu = append(gpu, backends.GPUAccel)
		return gpu
	default:
		return base
	}
}
