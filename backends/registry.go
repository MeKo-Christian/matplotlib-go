package backends

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cwbudde/matplotlib-go/canvas"
	"github.com/cwbudde/matplotlib-go/render"
)

// Backend identifies a specific renderer implementation.
type Backend string

const (
	GoBasic Backend = "gobasic"
	AGG     Backend = "agg"
	Skia    Backend = "skia"
	SVG     Backend = "svg"
	PDF     Backend = "pdf"
	PS      Backend = "ps"
	PGF     Backend = "pgf"
)

// Capability represents a backend feature capability.
type Capability string

const (
	// Rendering quality capabilities
	AntiAliasing Capability = "antialiasing"
	SubPixel     Capability = "subpixel"
	PatternFill  Capability = "patternfill"
	GradientFill Capability = "gradientfill"
	PathEffects  Capability = "patheffects"
	PathClip     Capability = "pathclip"

	// Performance capabilities
	GPUAccel  Capability = "gpuaccel"
	Threading Capability = "threading"

	// Output capabilities
	DPIAware          Capability = "dpiaware"
	VectorOutput      Capability = "vectoroutput"
	TextShaping       Capability = "textshaping"
	FontHinting       Capability = "fonthinting"
	TextBounds        Capability = "textbounds"
	TextPathing       Capability = "textpathing"
	RotatedText       Capability = "rotatedtext"
	VerticalText      Capability = "verticaltext"
	ImageTransform    Capability = "imagetransform"
	RGBABuffer        Capability = "rgbabuffer"
	BufferRegion      Capability = "bufferregion"
	OffscreenFilter   Capability = "offscreenfilter"
	NativeHatcher     Capability = "nativehatcher"
	MixedRasterVector Capability = "mixedrastervector"
	PNGExport         Capability = "pngexport"
	SVGExport         Capability = "svgexport"
	PDFExport         Capability = "pdfexport"
	PSExport          Capability = "psexport"
	PGFExport         Capability = "pgfexport"

	// Batch drawing capabilities
	MarkerBatch          Capability = "markerbatch"
	PathCollectionBatch  Capability = "pathcollectionbatch"
	QuadMeshBatch        Capability = "quadmeshbatch"
	GouraudTriangleBatch Capability = "gouraudtrianglebatch"

	// Extended text capabilities — explicit-font variants of DrawText/Rotated/Vertical.
	FontKeyText         Capability = "fontkeytext"
	FontKeyRotatedText  Capability = "fontkeyrotatedtext"
	FontKeyVerticalText Capability = "fontkeyverticaltext"

	// TeX / LaTeX text pipeline capabilities.
	TeXMetrics Capability = "texmetrics"
	TeXText    Capability = "textex"
	RotatedTeX Capability = "rotatedtex"

	// Advanced clipping / vector export capabilities.
	ClipPathTransform Capability = "clippathtransform"
	SVGOptionExport   Capability = "svgoptionexport"
)

// CapabilityStatus reports whether a backend capability is unavailable,
// available only through renderer-neutral fallback, or implemented natively.
type CapabilityStatus string

const (
	CapabilityUnsupported CapabilityStatus = "unsupported"
	CapabilityFallback    CapabilityStatus = "fallback"
	CapabilityNative      CapabilityStatus = "native"
)

// capabilityRuntimeChecks ties selected capabilities to concrete renderer interfaces.
var capabilityRuntimeChecks = map[Capability]func(render.Renderer) bool{
	DPIAware: func(r render.Renderer) bool {
		_, ok := r.(render.DPIAware)
		return ok
	},
	TextShaping: func(r render.Renderer) bool {
		_, ok := r.(render.TextDrawer)
		return ok
	},
	FontHinting: func(r render.Renderer) bool {
		_, ok := r.(render.TextFontMetricer)
		return ok
	},
	VectorOutput: func(r render.Renderer) bool {
		if _, hasSVG := r.(render.SVGExporter); hasSVG {
			return true
		}
		if _, hasPDF := r.(render.PDFExporter); hasPDF {
			return true
		}
		if _, hasPS := r.(render.PSExporter); hasPS {
			return true
		}
		return false
	},
	PNGExport: func(r render.Renderer) bool {
		_, ok := r.(render.PNGExporter)
		return ok
	},
	SVGExport: func(r render.Renderer) bool {
		_, ok := r.(render.SVGExporter)
		return ok
	},
	PDFExport: func(r render.Renderer) bool {
		_, ok := r.(render.PDFExporter)
		return ok
	},
	PSExport: func(r render.Renderer) bool {
		_, ok := r.(render.PSExporter)
		return ok
	},
	PGFExport: func(r render.Renderer) bool {
		_, ok := r.(render.PGFExporter)
		return ok
	},
	TextBounds: func(r render.Renderer) bool {
		_, ok := r.(render.TextBounder)
		return ok
	},
	TextPathing: func(r render.Renderer) bool {
		_, ok := r.(render.TextPather)
		return ok
	},
	RotatedText: func(r render.Renderer) bool {
		_, ok := r.(render.RotatedTextDrawer)
		return ok
	},
	VerticalText: func(r render.Renderer) bool {
		_, ok := r.(render.VerticalTextDrawer)
		return ok
	},
	ImageTransform: func(r render.Renderer) bool {
		_, ok := r.(render.ImageTransformer)
		return ok
	},
	RGBABuffer: func(r render.Renderer) bool {
		_, ok := r.(render.RGBAExporter)
		return ok
	},
	BufferRegion: func(r render.Renderer) bool {
		_, ok := r.(render.BufferRegioner)
		return ok
	},
	OffscreenFilter: func(r render.Renderer) bool {
		_, ok := r.(render.FilterRenderer)
		return ok
	},
	PatternFill: func(r render.Renderer) bool {
		filler, ok := r.(render.PatternFiller)
		return ok && filler.SupportsPatternFill()
	},
	GradientFill: func(r render.Renderer) bool {
		filler, ok := r.(render.GradientFiller)
		return ok && filler.SupportsGradientFill()
	},
	PathEffects: func(r render.Renderer) bool {
		_, ok := r.(render.PathEffectDrawer)
		return ok
	},
	MixedRasterVector: func(r render.Renderer) bool {
		_, ok := r.(render.RasterizationController)
		return ok
	},
	MarkerBatch: func(r render.Renderer) bool {
		_, ok := r.(render.MarkerDrawer)
		return ok
	},
	PathCollectionBatch: func(r render.Renderer) bool {
		_, ok := r.(render.PathCollectionDrawer)
		return ok
	},
	NativeHatcher: func(r render.Renderer) bool {
		_, ok := r.(render.NativeHatcher)
		return ok
	},
	QuadMeshBatch: func(r render.Renderer) bool {
		_, ok := r.(render.QuadMeshDrawer)
		return ok
	},
	GouraudTriangleBatch: func(r render.Renderer) bool {
		_, ok := r.(render.GouraudTriangleDrawer)
		return ok
	},
	FontKeyText: func(r render.Renderer) bool {
		_, ok := r.(render.FontTextDrawer)
		return ok
	},
	FontKeyRotatedText: func(r render.Renderer) bool {
		_, ok := r.(render.FontRotatedTextDrawer)
		return ok
	},
	FontKeyVerticalText: func(r render.Renderer) bool {
		_, ok := r.(render.FontVerticalTextDrawer)
		return ok
	},
	TeXMetrics: func(r render.Renderer) bool {
		_, ok := r.(render.TeXMetricer)
		return ok
	},
	TeXText: func(r render.Renderer) bool {
		_, ok := r.(render.TeXDrawer)
		return ok
	},
	RotatedTeX: func(r render.Renderer) bool {
		_, ok := r.(render.RotatedTeXDrawer)
		return ok
	},
	ClipPathTransform: func(r render.Renderer) bool {
		_, ok := r.(render.ClipPathTransformer)
		return ok
	},
	SVGOptionExport: func(r render.Renderer) bool {
		_, ok := r.(render.SVGOptionExporter)
		return ok
	},
}

// Config holds backend-specific configuration options.
type Config struct {
	// Common options
	Width       int
	Height      int
	Background  render.Color
	DPI         float64
	Transparent bool

	// Backend-specific options (use type assertion)
	Options interface{}
}

func (c Config) withDefaults() Config {
	if c.DPI == 0 {
		c.DPI = 72
	}
	if !c.Transparent {
		c.Background = normalizeBackground(c.Background)
	}
	return c
}

func normalizeBackground(c render.Color) render.Color {
	if c == (render.Color{}) {
		return render.Color{R: 1, G: 1, B: 1, A: 1}
	}
	if c.A == 0 && (c.R != 0 || c.G != 0 || c.B != 0) {
		c.A = 1
	}
	return c
}

// GoBasicConfig holds GoBasic-specific options.
type GoBasicConfig struct {
	// No specific options yet
}

// AGGConfig holds AGG-specific options.
type AGGConfig struct {
	// No specific options yet; AGG provides anti-aliasing by default.
}

// SkiaConfig holds Skia-specific options.
type SkiaConfig struct {
	UseGPU      bool
	SampleCount int    // MSAA sample count
	ColorType   string // RGBA8888, etc.
}

// Factory creates a new renderer instance.
type Factory func(config Config) (render.Renderer, error)

// CanvasFactory creates a backend-managed figure canvas.
type CanvasFactory func(config Config, fig *canvas.Figure) (canvas.FigureCanvas, error)

// ManagerFactory creates a backend-managed figure manager.
type ManagerFactory func(config Config, fig *canvas.Figure) (canvas.FigureManager, error)

// BackendInfo describes a backend's capabilities and factory.
type BackendInfo struct {
	Name                 string
	Description          string
	Capabilities         []Capability
	FallbackCapabilities []Capability
	SaveFormats          map[string]SaveHandler
	Factory              Factory
	CanvasFactory        CanvasFactory
	ManagerFactory       ManagerFactory
	Available            bool // Whether dependencies are available
}

// SaveHandler writes output in a specific file format using the renderer.
type SaveHandler func(render.Renderer, string, ...render.SVGOption) error

func normalizeSaveFormat(format string) string {
	ext := strings.ToLower(strings.TrimSpace(format))
	if ext == "" {
		return ""
	}
	if ext[0] != '.' {
		ext = "." + ext
	}
	return ext
}

// SavePNG saves using the renderer PNG export interface.
func SavePNG(renderer render.Renderer, path string, _ ...render.SVGOption) error {
	exporter, ok := renderer.(render.PNGExporter)
	if !ok {
		return fmt.Errorf("backends: renderer does not implement PNG export")
	}
	return exporter.SavePNG(path)
}

// SavePDF saves using the renderer PDF export interface.
func SavePDF(renderer render.Renderer, path string, _ ...render.SVGOption) error {
	exporter, ok := renderer.(render.PDFExporter)
	if !ok {
		return fmt.Errorf("backends: renderer does not implement PDF export")
	}
	return exporter.SavePDF(path)
}

// SavePS saves using the renderer PostScript export interface.
func SavePS(renderer render.Renderer, path string, _ ...render.SVGOption) error {
	exporter, ok := renderer.(render.PSExporter)
	if !ok {
		return fmt.Errorf("backends: renderer does not implement PostScript export")
	}
	return exporter.SavePS(path)
}

// SavePGF saves using the renderer PGF export interface.
func SavePGF(renderer render.Renderer, path string, _ ...render.SVGOption) error {
	exporter, ok := renderer.(render.PGFExporter)
	if !ok {
		return fmt.Errorf("backends: renderer does not implement PGF export")
	}
	return exporter.SavePGF(path)
}

// SaveSVG saves using the renderer SVG export interface.
func SaveSVG(renderer render.Renderer, path string, opts ...render.SVGOption) error {
	exporter, ok := renderer.(render.SVGExporter)
	if !ok {
		return fmt.Errorf("backends: renderer does not implement SVG export")
	}
	if len(opts) > 0 {
		if setter, ok := renderer.(render.SVGOptionSetter); ok {
			setter.SetSVGOptions(render.ResolveSVGOptions(opts...))
		}
	}
	if optionExporter, ok := renderer.(render.SVGOptionExporter); ok && len(opts) > 0 {
		return optionExporter.SaveSVGWithOptions(path, render.ResolveSVGOptions(opts...))
	}
	return exporter.SaveSVG(path)
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
	sort.Slice(available, func(i, j int) bool {
		return preferredBackendLess(available[i], available[j])
	})
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

	return info.Factory(config.withDefaults())
}

// HasCapability checks if a backend supports a capability.
func (r *Registry) HasCapability(backend Backend, capability Capability) bool {
	info, ok := r.backends[backend]
	if !ok {
		return false
	}

	return info.hasCapability(capability)
}

func (i *BackendInfo) hasCapability(capability Capability) bool {
	if i == nil {
		return false
	}
	for _, c := range i.Capabilities {
		if c == capability {
			return true
		}
	}
	return false
}

func (i *BackendInfo) hasFallbackCapability(capability Capability) bool {
	if i == nil {
		return false
	}
	for _, c := range i.FallbackCapabilities {
		if c == capability {
			return true
		}
	}
	return false
}

// SupportsRendererCapability checks both registry declaration and renderer implementation
// for optional capabilities that require runtime feature checks.
func (r *Registry) SupportsRendererCapability(backend Backend, renderer render.Renderer, capability Capability) bool {
	if renderer == nil {
		return false
	}
	if !r.HasCapability(backend, capability) {
		return false
	}

	check, ok := capabilityRuntimeChecks[capability]
	if !ok {
		return true
	}
	return check(renderer)
}

// RendererCapabilityStatus reports unsupported/fallback/native support for a
// capability on a concrete renderer.
func (r *Registry) RendererCapabilityStatus(backend Backend, renderer render.Renderer, capability Capability) CapabilityStatus {
	info, ok := r.Get(backend)
	if !ok || renderer == nil {
		return CapabilityUnsupported
	}
	if r.SupportsRendererCapability(backend, renderer, capability) {
		return CapabilityNative
	}
	if info.hasFallbackCapability(capability) {
		return CapabilityFallback
	}
	return CapabilityUnsupported
}

// VerifyRendererCapabilities ensures a renderer satisfies all capability-backed contracts
// for the selected backend.
func (r *Registry) VerifyRendererCapabilities(backend Backend, renderer render.Renderer) error {
	if renderer == nil {
		return fmt.Errorf("backends: nil renderer for %s", backend)
	}

	info, ok := r.Get(backend)
	if !ok {
		return fmt.Errorf("backends: unknown backend %s", backend)
	}
	for _, capability := range info.Capabilities {
		if !r.SupportsRendererCapability(backend, renderer, capability) {
			return fmt.Errorf("backends: backend %s does not fully implement capability %s", backend, capability)
		}
	}
	return nil
}

// SaveViaExtension dispatches to the backend-specific save handler for extension.
func (r *Registry) SaveViaExtension(backend Backend, renderer render.Renderer, path string, opts ...render.SVGOption) error {
	info, ok := r.Get(backend)
	if !ok {
		return fmt.Errorf("unknown backend: %s", backend)
	}
	return info.saveViaExtension(renderer, path, opts...)
}

func (i *BackendInfo) saveViaExtension(renderer render.Renderer, path string, opts ...render.SVGOption) error {
	if i == nil {
		return errors.New("backends: nil backend info")
	}

	ext := normalizeSaveFormat(filepath.Ext(path))
	if ext == "" {
		return fmt.Errorf("backends: unsupported save extension %q", filepath.Ext(path))
	}

	if i.SaveFormats != nil {
		if handler, ok := i.SaveFormats[ext]; ok {
			return handler(renderer, path, opts...)
		}
		return fmt.Errorf("backends: backend does not support extension %q", ext)
	}

	return fmt.Errorf("backends: backend has no registered save formats")
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

// SupportsRendererCapability checks default-registry capability contracts for renderer
// support, including runtime feature checks for interface-based capabilities.
func SupportsRendererCapability(backend Backend, renderer render.Renderer, capability Capability) bool {
	return DefaultRegistry.SupportsRendererCapability(backend, renderer, capability)
}

// RendererCapabilityStatus reports unsupported/fallback/native support using
// the default registry.
func RendererCapabilityStatus(backend Backend, renderer render.Renderer, capability Capability) CapabilityStatus {
	return DefaultRegistry.RendererCapabilityStatus(backend, renderer, capability)
}

// VerifyRendererCapabilities checks that the renderer satisfies all capability-backed
// requirements of the selected backend in the default registry.
func VerifyRendererCapabilities(backend Backend, renderer render.Renderer) error {
	return DefaultRegistry.VerifyRendererCapabilities(backend, renderer)
}

// GetBestBackend selects the best available backend for given requirements.
func GetBestBackend(required []Capability) (Backend, error) {
	available := Available()
	if len(available) == 0 {
		return "", errors.New("no backends available")
	}

	// Score backends based on capabilities
	bestBackend := available[0]
	bestScore := -1

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

// DefaultBackend returns the backend used when callers do not specify one.
// GoBasic is the default in this project to keep the renderer pure-Go.
func DefaultBackend() Backend {
	return GoBasic
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

func preferredBackendLess(a, b Backend) bool {
	if a == GoBasic && b != GoBasic {
		return true
	}
	if b == GoBasic && a != GoBasic {
		return false
	}
	return a < b
}
