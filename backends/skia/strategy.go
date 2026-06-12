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

// NativePathRequirement records one deferred Skia-native integration point.
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
	Binding         BindingStrategy
	Mode            RenderMode
	NativeSurface   bool
	SupportsShaders bool
	Description     string
}

// BackendStrategy returns the documented Skia integration strategy. It is kept
// as code so tests and docs can agree on the same build/dependency contract.
//
// The reported render mode and GPU status depend on the skiagpu build tag: with
// the tag the GPU render mode is selectable and reported as StatusPlanned (the
// CPU-readback scaffold is in place but native GPU rendering is not); without it
// GPU stays StatusDeferred and the default mode is CPU.
func BackendStrategy() Strategy {
	defaultMode := ModeCPU
	gpuStatus := StatusDeferred
	if gpuBuildEnabled {
		defaultMode = ModeGPU
		gpuStatus = StatusPlanned
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
			"Skia shared library for future native paths",
			"C ABI wrapper library for future native paths",
			"CGO_ENABLED=1 for future native paths",
		},
	}
}

// NativePathRequirements returns the explicit external Skia primitives tracked
// for promotion from CPU compatibility paths to truly native backend paths.
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
			Primitive:           "tiled SkShader",
			Modes:               []RenderMode{ModeCPU, ModeGPU},
			Capabilities:        []backends.Capability{backends.NativeHatcher},
			ExternalEntrypoints: []string{"SkShader", "SkCanvas::drawPath"},
			Status:              StatusDeferred,
			BlockedBy:           externalABIBlocker,
		},
		{
			Primitive:           "SkSurface::MakeRenderTarget",
			Modes:               []RenderMode{ModeGPU},
			Capabilities:        []backends.Capability{backends.GPUAccel},
			ExternalEntrypoints: []string{"SkSurface::MakeRenderTarget", "SkSurface::readPixels"},
			Status:              StatusDeferred,
			BlockedBy:           "external Skia GPU library, platform GPU context, and deterministic CPU readback",
		},
	}
}

// ModeCapabilities returns the optional renderer capabilities targeted by a Skia
// render mode. It encodes the CPU-vs-GPU capability split as data so the
// comparison report and tests can reason about per-mode support before a second
// (GPU) registry entry becomes meaningful.
//
// Today the CPU and GPU sets are identical because the GPU mode is a
// deterministic CPU-readback scaffold; the helper exists so that as native GPU
// paths land (e.g. drawAtlas/SkVertices promoted from bridged to native) the
// divergence is expressed in one place rather than scattered through reporting.
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
		gpu := make([]backends.Capability, len(base))
		copy(gpu, base)
		return gpu
	default:
		return base
	}
}
