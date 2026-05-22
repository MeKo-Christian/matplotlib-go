package skia

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
func BackendStrategy() Strategy {
	return Strategy{
		BuildTag:    "skia",
		Binding:     BindingExternalCAPI,
		DefaultMode: ModeCPU,
		CPUStatus:   StatusImplemented,
		GPUStatus:   StatusDeferred,
		CIDefault:   CIDefaultStub,
		RequiredLibraries: []string{
			"none for the skia-tagged CPU compatibility renderer",
			"Skia shared library for future native paths",
			"C ABI wrapper library for future native paths",
			"CGO_ENABLED=1 for future native paths",
		},
	}
}
