package backends

import (
	"errors"
	"fmt"

	"matplotlib-go/render"
)

// Backend identifies a specific renderer implementation.
type Backend string

const (
	GoBasic Backend = "gobasic"
	Skia    Backend = "skia"
	// Future backends: AGG, PDF, SVG, etc.
)

// Capability represents a backend feature capability.
type Capability string

const (
	// Rendering quality capabilities
	AntiAliasing Capability = "antialiasing"
	SubPixel     Capability = "subpixel"
	GradientFill Capability = "gradientfill"
	PathClip     Capability = "pathclip"
	
	// Performance capabilities
	GPUAccel     Capability = "gpuaccel"
	Threading    Capability = "threading"
	
	// Output capabilities
	VectorOutput Capability = "vectoroutput"
	TextShaping  Capability = "textshaping"
	FontHinting  Capability = "fonthinting"
)

// Config holds backend-specific configuration options.
type Config struct {
	// Common options
	Width      int
	Height     int
	Background render.Color
	DPI        float64
	
	// Backend-specific options (use type assertion)
	Options interface{}
}

// GoBasicConfig holds GoBasic-specific options.
type GoBasicConfig struct {
	// No specific options yet
}

// SkiaConfig holds Skia-specific options.
type SkiaConfig struct {
	UseGPU      bool
	SampleCount int // MSAA sample count
	ColorType   string // RGBA8888, etc.
}

// Factory creates a new renderer instance.
type Factory func(config Config) (render.Renderer, error)

// BackendInfo describes a backend's capabilities and factory.
type BackendInfo struct {
	Name         string
	Description  string
	Capabilities []Capability
	Factory      Factory
	Available    bool // Whether dependencies are available
}

// Registry manages available rendering backends.
type Registry struct {
	backends map[Backend]*BackendInfo
}

// NewRegistry creates a new backend registry.
func NewRegistry() *Registry {
	return &Registry{
		backends: make(map[Backend]*BackendInfo),
	}
}

// Register adds a backend to the registry.
func (r *Registry) Register(backend Backend, info *BackendInfo) {
	r.backends[backend] = info
}

// Get retrieves backend info.
func (r *Registry) Get(backend Backend) (*BackendInfo, bool) {
	info, ok := r.backends[backend]
	return info, ok
}

// Available returns all available backends.
func (r *Registry) Available() []Backend {
	var available []Backend
	for backend, info := range r.backends {
		if info.Available {
			available = append(available, backend)
		}
	}
	return available
}

// Create instantiates a renderer using the specified backend.
func (r *Registry) Create(backend Backend, config Config) (render.Renderer, error) {
	info, ok := r.backends[backend]
	if !ok {
		return nil, fmt.Errorf("unknown backend: %s", backend)
	}
	
	if !info.Available {
		return nil, fmt.Errorf("backend %s is not available (missing dependencies?)", backend)
	}
	
	return info.Factory(config)
}

// HasCapability checks if a backend supports a capability.
func (r *Registry) HasCapability(backend Backend, capability Capability) bool {
	info, ok := r.backends[backend]
	if !ok {
		return false
	}
	
	for _, c := range info.Capabilities {
		if c == capability {
			return true
		}
	}
	return false
}

// DefaultRegistry is the global backend registry.
var DefaultRegistry = NewRegistry()

// Convenience functions using the default registry

// Register registers a backend in the default registry.
func Register(backend Backend, info *BackendInfo) {
	DefaultRegistry.Register(backend, info)
}

// Create creates a renderer using the default registry.
func Create(backend Backend, config Config) (render.Renderer, error) {
	return DefaultRegistry.Create(backend, config)
}

// Available returns available backends from the default registry.
func Available() []Backend {
	return DefaultRegistry.Available()
}

// HasCapability checks capability in the default registry.
func HasCapability(backend Backend, capability Capability) bool {
	return DefaultRegistry.HasCapability(backend, capability)
}

// GetBestBackend selects the best available backend for given requirements.
func GetBestBackend(required []Capability) (Backend, error) {
	available := Available()
	if len(available) == 0 {
		return "", errors.New("no backends available")
	}
	
	// Score backends based on capabilities
	bestBackend := available[0]
	bestScore := 0
	
	for _, backend := range available {
		score := 0
		allRequired := true
		
		for _, capability := range required {
			if HasCapability(backend, capability) {
				score++
			} else {
				allRequired = false
			}
		}
		
		// Prefer backends that have all required capabilities
		if allRequired && score > bestScore {
			bestBackend = backend
			bestScore = score
		}
	}
	
	return bestBackend, nil
}

// SimpleConfig creates a basic config for testing/simple use.
func SimpleConfig(width, height int, bg render.Color) Config {
	return Config{
		Width:      width,
		Height:     height,
		Background: bg,
		DPI:        72.0,
	}
}