package backends

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

type contractRenderer struct {
	render.NullRenderer
	pngPath string
	svgPath string
}

func (r *contractRenderer) DrawText(string, geom.Pt, float64, render.Color) {}

func (r *contractRenderer) SavePNG(path string) error {
	r.pngPath = path
	return nil
}

func (r *contractRenderer) SaveSVG(path string) error {
	r.svgPath = path
	return nil
}

func TestRegistryVerifiesDeclaredRendererCapabilities(t *testing.T) {
	registry := NewRegistry()
	backend := Backend("contract")
	registry.Register(backend, &BackendInfo{
		Name:         "Contract",
		Available:    true,
		Capabilities: []Capability{TextShaping, VectorOutput},
		Factory: func(Config) (render.Renderer, error) {
			return &contractRenderer{}, nil
		},
	})

	if registry.SupportsRendererCapability(backend, &render.NullRenderer{}, TextShaping) {
		t.Fatal("null renderer unexpectedly supports text shaping")
	}
	if err := registry.VerifyRendererCapabilities(backend, &render.NullRenderer{}); err == nil {
		t.Fatal("VerifyRendererCapabilities accepted renderer missing declared optional interfaces")
	}
	if err := registry.VerifyRendererCapabilities(backend, &contractRenderer{}); err != nil {
		t.Fatalf("VerifyRendererCapabilities rejected contract renderer: %v", err)
	}
}

func TestRegistrySaveViaExtensionUsesBackendFormatHandlers(t *testing.T) {
	registry := NewRegistry()
	backend := Backend("contract")
	registry.Register(backend, &BackendInfo{
		Name:      "Contract",
		Available: true,
		SaveFormats: map[string]SaveHandler{
			".png": SavePNG,
		},
	})

	renderer := &contractRenderer{}
	if err := registry.SaveViaExtension(backend, renderer, "plot.png"); err != nil {
		t.Fatalf("SaveViaExtension returned error: %v", err)
	}
	if renderer.pngPath != "plot.png" {
		t.Fatalf("pngPath = %q, want plot.png", renderer.pngPath)
	}
	if err := registry.SaveViaExtension(backend, renderer, "plot.pdf"); err == nil {
		t.Fatal("SaveViaExtension accepted unsupported format")
	}
}

func TestRegistrySaveViaExtensionRequiresRegisteredSaveFormat(t *testing.T) {
	registry := NewRegistry()
	backend := Backend("implicit")
	registry.Register(backend, &BackendInfo{
		Name:      "Implicit",
		Available: true,
	})

	renderer := &contractRenderer{}
	if err := registry.SaveViaExtension(backend, renderer, "plot.png"); err == nil {
		t.Fatal("SaveViaExtension should not fall back to hard-coded PNG handling")
	}
	if renderer.pngPath != "" {
		t.Fatalf("SaveViaExtension invoked PNG exporter despite missing SaveFormats entry: %q", renderer.pngPath)
	}
}
