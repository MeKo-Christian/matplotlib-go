package backends

import (
	"fmt"
	"sort"
	"strings"

	"github.com/cwbudde/matplotlib-go/render"
)

// AllCapabilities lists every capability advertised by the registry, in a stable
// display order grouped by area (rendering, performance, text, output, batches).
//
// Keep this slice in sync with the constants declared in registry.go; the
// CapabilityMatrix, comparison reports, and verification tests iterate it.
var AllCapabilities = []Capability{
	// Rendering quality.
	AntiAliasing, SubPixel, PatternFill, GradientFill, PathEffects,
	PathClip, ClipPathTransform,
	// Performance.
	GPUAccel, Threading,
	// Text.
	DPIAware, TextShaping, FontHinting, TextBounds, TextPathing,
	RotatedText, VerticalText,
	FontKeyText, FontKeyRotatedText, FontKeyVerticalText,
	TeXMetrics, TeXText, RotatedTeX,
	// Imaging / buffers.
	ImageTransform, RGBABuffer, BufferRegion, OffscreenFilter,
	NativeHatcher, MixedRasterVector,
	// Export.
	VectorOutput, PNGExport, SVGExport, SVGOptionExport, PDFExport, PSExport, PGFExport,
	// Batches.
	MarkerBatch, PathCollectionBatch, QuadMeshBatch, GouraudTriangleBatch,
}

// CapabilityMatrix returns a formatted table of backend capabilities.
//
// Each cell reports CapabilityStatus markers: ✓ native, ~ fallback, · unsupported.
// Capabilities that require a live renderer (everything in the runtime-checks map)
// are reported as declared by the registry; native verification happens through
// VerifyRendererCapabilities and BackendComparisonReport.
func CapabilityMatrix() string {
	available := Available()
	if len(available) == 0 {
		return "No backends available"
	}

	nameWidth := len("Backend")
	for _, backend := range available {
		if n := len(backend); n > nameWidth {
			nameWidth = n
		}
	}
	colWidth := 0
	for _, capability := range AllCapabilities {
		if n := len(capability); n > colWidth {
			colWidth = n
		}
	}
	colWidth += 2

	var b strings.Builder
	fmt.Fprintf(&b, "%-*s", nameWidth+2, "Backend")
	for _, capability := range AllCapabilities {
		fmt.Fprintf(&b, "%-*s", colWidth, string(capability))
	}
	b.WriteByte('\n')

	fmt.Fprintf(&b, "%-*s", nameWidth+2, strings.Repeat("-", nameWidth))
	for range AllCapabilities {
		fmt.Fprintf(&b, "%-*s", colWidth, strings.Repeat("-", colWidth-2))
	}
	b.WriteByte('\n')

	for _, backend := range available {
		fmt.Fprintf(&b, "%-*s", nameWidth+2, string(backend))
		info, _ := DefaultRegistry.Get(backend)
		for _, capability := range AllCapabilities {
			marker := declaredCapabilityMarker(info, capability)
			fmt.Fprintf(&b, "%-*s", colWidth, marker)
		}
		b.WriteByte('\n')
	}

	return b.String()
}

// BackendComparisonReport instantiates each available backend with a small
// throwaway configuration, then reports declared-vs-runtime support for every
// capability in AllCapabilities as native / fallback / unsupported.
//
// Backends that fail to instantiate are reported with their error so the
// caller can still see the matrix at a glance during CI or local audits.
func BackendComparisonReport(config Config) string {
	return DefaultRegistry.BackendComparisonReport(config)
}

// BackendComparisonReport is the registry-scoped variant of the top-level
// helper. It enables tests that swap DefaultRegistry to assert against the
// active registry rather than the global one.
func (r *Registry) BackendComparisonReport(config Config) string {
	available := r.Available()
	if len(available) == 0 {
		return "No backends available\n"
	}

	if config.Width <= 0 {
		config.Width = 64
	}
	if config.Height <= 0 {
		config.Height = 64
	}

	type row struct {
		backend  Backend
		renderer render.Renderer
		instErr  error
	}
	rows := make([]row, 0, len(available))
	for _, backend := range available {
		renderer, err := r.Create(backend, config)
		rows = append(rows, row{backend: backend, renderer: renderer, instErr: err})
	}

	nameWidth := len("Backend")
	for _, rw := range rows {
		if n := len(rw.backend); n > nameWidth {
			nameWidth = n
		}
	}
	colWidth := 0
	for _, capability := range AllCapabilities {
		if n := len(capability); n > colWidth {
			colWidth = n
		}
	}
	colWidth += 2

	var b strings.Builder
	fmt.Fprintf(&b, "%-*s", nameWidth+2, "Backend")
	for _, capability := range AllCapabilities {
		fmt.Fprintf(&b, "%-*s", colWidth, string(capability))
	}
	b.WriteByte('\n')

	fmt.Fprintf(&b, "%-*s", nameWidth+2, strings.Repeat("-", nameWidth))
	for range AllCapabilities {
		fmt.Fprintf(&b, "%-*s", colWidth, strings.Repeat("-", colWidth-2))
	}
	b.WriteByte('\n')

	for _, rw := range rows {
		fmt.Fprintf(&b, "%-*s", nameWidth+2, string(rw.backend))
		if rw.instErr != nil {
			fmt.Fprintf(&b, "instantiation error: %v\n", rw.instErr)
			continue
		}
		info, _ := r.Get(rw.backend)
		for _, capability := range AllCapabilities {
			status := r.RendererCapabilityStatus(rw.backend, rw.renderer, capability)
			marker := capabilityStatusMarker(status, info, capability)
			fmt.Fprintf(&b, "%-*s", colWidth, marker)
		}
		b.WriteByte('\n')
	}

	return b.String()
}

// declaredCapabilityMarker renders a registry-only status (no live renderer):
// native if Capabilities lists it, fallback (~) if FallbackCapabilities lists
// it, unsupported (·) otherwise.
func declaredCapabilityMarker(info *BackendInfo, capability Capability) string {
	if info == nil {
		return "·"
	}
	if info.hasCapability(capability) {
		return "✓"
	}
	if info.hasFallbackCapability(capability) {
		return "~"
	}
	return "·"
}

// capabilityStatusMarker renders a runtime CapabilityStatus, decorating native
// support that was declared only as fallback in the registry (which would be a
// declaration drift worth noticing).
func capabilityStatusMarker(status CapabilityStatus, info *BackendInfo, capability Capability) string {
	switch status {
	case CapabilityNative:
		if info != nil && info.hasFallbackCapability(capability) && !info.hasCapability(capability) {
			return "✓!"
		}
		return "✓"
	case CapabilityFallback:
		return "~"
	default:
		return "·"
	}
}

// RequiredCapabilities defines capability sets for common use cases.
var RequiredCapabilities = map[string][]Capability{
	"basic": {
		AntiAliasing,
	},
	"publication": {
		AntiAliasing,
		VectorOutput,
		TextShaping,
	},
	"interactive": {
		AntiAliasing,
		GPUAccel,
		Threading,
	},
	"scientific": {
		AntiAliasing,
		SubPixel,
		VectorOutput,
		TextShaping,
		FontHinting,
	},
}

// GetRecommendedBackend returns the best backend for a specific use case.
func GetRecommendedBackend(useCase string) (Backend, error) {
	required, ok := RequiredCapabilities[useCase]
	if !ok {
		return "", fmt.Errorf("unknown use case: %s", useCase)
	}

	return GetBestBackend(required)
}

// BackendsForExtension lists every available backend whose registered
// SaveFormats include the requested file extension (with or without leading dot).
//
// The result is sorted by the same preference order used by Available().
func BackendsForExtension(ext string) []Backend {
	return DefaultRegistry.BackendsForExtension(ext)
}

// BackendsForExtension is the registry-scoped variant.
func (r *Registry) BackendsForExtension(ext string) []Backend {
	normalized := normalizeSaveFormat(ext)
	if normalized == "" {
		return nil
	}
	var out []Backend
	for _, backend := range r.Available() {
		info, ok := r.Get(backend)
		if !ok || info.SaveFormats == nil {
			continue
		}
		if _, ok := info.SaveFormats[normalized]; ok {
			out = append(out, backend)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		return preferredBackendLess(out[i], out[j])
	})
	return out
}

// SelectBackendForExtension picks a backend that supports the requested
// extension. choice may be the empty string, "auto", or a specific backend
// name; if non-empty, the choice is validated against the extension as well.
//
// When choice is empty, the first available backend whose SaveFormats contain
// ext is returned. The returned backend is guaranteed to satisfy the requested
// capabilities.
func SelectBackendForExtension(choice, ext string, required []Capability) (Backend, error) {
	return DefaultRegistry.SelectBackendForExtension(choice, ext, required)
}

// SelectBackendForExtension is the registry-scoped variant.
func (r *Registry) SelectBackendForExtension(choice, ext string, required []Capability) (Backend, error) {
	normalized := normalizeSaveFormat(ext)
	if normalized == "" {
		return "", fmt.Errorf("backends: empty file extension")
	}
	candidates := r.BackendsForExtension(normalized)
	if len(candidates) == 0 {
		return "", fmt.Errorf("backends: no available backend supports extension %q", normalized)
	}

	normalizedChoice := strings.ToLower(strings.TrimSpace(choice))
	if normalizedChoice != "" && normalizedChoice != string(AutoBackend) && normalizedChoice != "default" {
		want := Backend(normalizedChoice)
		for _, candidate := range candidates {
			if candidate == want {
				if len(required) > 0 && !hasAllCapabilitiesIn(r, candidate, required) {
					return "", fmt.Errorf("backend %s does not satisfy required capabilities", candidate)
				}
				return candidate, nil
			}
		}
		return "", fmt.Errorf("backend %s does not support extension %q", want, normalized)
	}

	for _, candidate := range candidates {
		if len(required) == 0 || hasAllCapabilitiesIn(r, candidate, required) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("backends: no backend supporting %q satisfies required capabilities", normalized)
}

func hasAllCapabilitiesIn(r *Registry, backend Backend, required []Capability) bool {
	for _, capability := range required {
		if !r.HasCapability(backend, capability) {
			return false
		}
	}
	return true
}
