package backends_test

import (
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/all" // register every built-in backend
	"github.com/cwbudde/matplotlib-go/render"
)

// TestRegisteredBackendsAdvertiseSupportedCapabilities walks every backend in
// DefaultRegistry, instantiates it, and verifies that each capability the
// backend declares as native is actually backed by the corresponding optional
// renderer interface in render/extensions.go.
//
// This is the runtime half of the capability contract (14.5): the registry
// guarantees that "declared native" matches "interface implemented", so artist
// code can rely on capability checks rather than backend-name conditionals.
func TestRegisteredBackendsAdvertiseSupportedCapabilities(t *testing.T) {
	available := backends.Available()
	if len(available) == 0 {
		t.Skip("no backends registered")
	}

	cfg := backends.Config{
		Width:      64,
		Height:     64,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		DPI:        72,
	}

	for _, backend := range available {
		backend := backend
		t.Run(string(backend), func(t *testing.T) {
			info, ok := backends.DefaultRegistry.Get(backend)
			if !ok {
				t.Fatalf("backend %s not registered", backend)
			}
			renderer, err := backends.Create(backend, cfg)
			if err != nil {
				t.Fatalf("Create(%s): %v", backend, err)
			}
			if err := backends.VerifyRendererCapabilities(backend, renderer); err != nil {
				t.Fatalf("VerifyRendererCapabilities(%s): %v", backend, err)
			}

			// Spot-check that fallback capabilities are NOT advertised as native
			// (they would surface as ✓! in the comparison report and indicate
			// declaration drift).
			for _, cap := range info.FallbackCapabilities {
				status := backends.RendererCapabilityStatus(backend, renderer, cap)
				if status == backends.CapabilityNative {
					t.Errorf("backend %s declares %s as fallback but renderer implements it natively; promote to Capabilities", backend, cap)
				}
			}
		})
	}
}

// TestPhase24NativeEffectBackendsStayNative pins the Phase 2.4 contract: the
// primary native targets must advertise and satisfy the renderer-neutral
// pattern, gradient, and path-effect capabilities through runtime interfaces.
func TestPhase24NativeEffectBackendsStayNative(t *testing.T) {
	cfg := backends.TestDefaultConfig(64, 64)
	for _, backend := range []backends.Backend{
		backends.AGG,
		backends.SVG,
		backends.PDF,
		backends.Skia,
	} {
		backend := backend
		t.Run(string(backend), func(t *testing.T) {
			info, ok := backends.DefaultRegistry.Get(backend)
			if !ok {
				t.Fatalf("backend %s not registered", backend)
			}
			if !info.Available {
				t.Skipf("backend %s is not available in this build", backend)
			}
			renderer, err := backends.Create(backend, cfg)
			if err != nil {
				t.Fatalf("Create(%s): %v", backend, err)
			}

			for _, cap := range []backends.Capability{
				backends.PatternFill,
				backends.GradientFill,
				backends.PathEffects,
			} {
				if status := backends.RendererCapabilityStatus(backend, renderer, cap); status != backends.CapabilityNative {
					t.Fatalf("RendererCapabilityStatus(%s, %s) = %s, want %s", backend, cap, status, backends.CapabilityNative)
				}
			}
		})
	}
}

// TestBackendComparisonReportContainsEveryBackend ensures the comparison
// helper introduced in 14.5 produces a row for each available backend and a
// column for each capability advertised by AllCapabilities.
func TestBackendComparisonReportContainsEveryBackend(t *testing.T) {
	report := backends.BackendComparisonReport(backends.Config{Width: 64, Height: 64, DPI: 72})
	if report == "" {
		t.Fatal("BackendComparisonReport returned empty output")
	}
	for _, backend := range backends.Available() {
		if !strings.Contains(report, string(backend)) {
			t.Errorf("report missing row for backend %s\n%s", backend, report)
		}
	}
	for _, cap := range backends.AllCapabilities {
		if !strings.Contains(report, string(cap)) {
			t.Errorf("report missing column for capability %s\n%s", cap, report)
		}
	}
	if !strings.Contains(report, string(backends.PGFExport)) {
		t.Errorf("report missing PGF export capability status\n%s", report)
	}
	if !strings.Contains(report, "SaveFormats") {
		t.Errorf("report missing registered save format column\n%s", report)
	}
	for _, ext := range []string{".pdf", ".ps", ".eps", ".pgf"} {
		if !strings.Contains(report, ext) {
			t.Errorf("report missing registered save extension %s\n%s", ext, report)
		}
	}
}

// TestSelectBackendForExtensionRoutesByCapability verifies that the registry
// helper used by pyplot picks the right backend for a given file extension.
func TestSelectBackendForExtensionRoutesByCapability(t *testing.T) {
	t.Run("png picks a PNG-capable backend", func(t *testing.T) {
		backend, err := backends.SelectBackendForExtension("", ".png", nil)
		if err != nil {
			t.Fatalf("SelectBackendForExtension(.png): %v", err)
		}
		if !backends.HasCapability(backend, backends.PNGExport) {
			t.Errorf("selected backend %s does not declare PNGExport", backend)
		}
	})

	t.Run("svg picks an SVG-capable backend", func(t *testing.T) {
		backend, err := backends.SelectBackendForExtension("", ".svg", nil)
		if err != nil {
			t.Fatalf("SelectBackendForExtension(.svg): %v", err)
		}
		if backend != backends.SVG {
			t.Errorf("expected SVG backend for .svg, got %s", backend)
		}
	})

	t.Run("explicit choice that does not support extension errors", func(t *testing.T) {
		_, err := backends.SelectBackendForExtension(string(backends.GoBasic), ".svg", nil)
		if err == nil {
			t.Fatal("expected error when GoBasic asked for .svg")
		}
	})

	t.Run("pdf picks a PDF-capable backend", func(t *testing.T) {
		backend, err := backends.SelectBackendForExtension("", ".pdf", nil)
		if err != nil {
			t.Fatalf("SelectBackendForExtension(.pdf): %v", err)
		}
		if backend != backends.PDF {
			t.Errorf("expected PDF backend for .pdf, got %s", backend)
		}
	})

	t.Run("ps picks a PostScript-capable backend", func(t *testing.T) {
		backend, err := backends.SelectBackendForExtension("", ".ps", nil)
		if err != nil {
			t.Fatalf("SelectBackendForExtension(.ps): %v", err)
		}
		if backend != backends.PS {
			t.Errorf("expected PS backend for .ps, got %s", backend)
		}
	})

	t.Run("eps picks a PostScript-capable backend", func(t *testing.T) {
		backend, err := backends.SelectBackendForExtension("", ".eps", nil)
		if err != nil {
			t.Fatalf("SelectBackendForExtension(.eps): %v", err)
		}
		if backend != backends.PS {
			t.Errorf("expected PS backend for .eps, got %s", backend)
		}
	})

	t.Run("pgf picks a PGF-capable backend", func(t *testing.T) {
		backend, err := backends.SelectBackendForExtension("", ".pgf", nil)
		if err != nil {
			t.Fatalf("SelectBackendForExtension(.pgf): %v", err)
		}
		if backend != backends.PGF {
			t.Errorf("expected PGF backend for .pgf, got %s", backend)
		}
	})

	t.Run("unknown extension errors", func(t *testing.T) {
		_, err := backends.SelectBackendForExtension("", ".bogus", nil)
		if err == nil {
			t.Fatal("expected error for unsupported extension")
		}
	})
}
