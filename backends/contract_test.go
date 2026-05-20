package backends

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

type contractRenderer struct {
	render.NullRenderer
	pngPath    string
	svgPath    string
	svgOptions render.SVGOptions
	pdfPath    string
	pdfOptions render.PDFOptions
	psPath     string
	psOptions  render.PSOptions
	pgfPath    string
	pgfOptions render.PGFOptions
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

func (r *contractRenderer) SaveSVGWithOptions(path string, opts render.SVGOptions) error {
	r.svgOptions = opts
	return r.SaveSVG(path)
}

func (r *contractRenderer) SetSVGOptions(opts render.SVGOptions) {
	r.svgOptions = opts
}

func (r *contractRenderer) SavePDF(path string) error {
	r.pdfPath = path
	return nil
}

func (r *contractRenderer) SavePDFWithOptions(path string, opts render.PDFOptions) error {
	r.pdfOptions = opts
	return r.SavePDF(path)
}

func (r *contractRenderer) SetPDFOptions(opts render.PDFOptions) {
	r.pdfOptions = opts
}

func (r *contractRenderer) SavePS(path string) error {
	r.psPath = path
	return nil
}

func (r *contractRenderer) SetPSOptions(opts render.PSOptions) {
	r.psOptions = opts
}

func (r *contractRenderer) SavePGF(path string) error {
	r.pgfPath = path
	return nil
}

func (r *contractRenderer) SavePGFWithOptions(path string, opts render.PGFOptions) error {
	r.pgfOptions = opts
	return r.SavePGF(path)
}

func (r *contractRenderer) SetPGFOptions(opts render.PGFOptions) {
	r.pgfOptions = opts
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

func TestRegistrySaveViaExtensionForwardsSharedOptions(t *testing.T) {
	registry := NewRegistry()
	backend := Backend("contract")
	registry.Register(backend, &BackendInfo{
		Name:      "Contract",
		Available: true,
		SaveFormats: map[string]SaveHandler{
			".pdf": SavePDF,
			".pgf": SavePGF,
		},
	})

	renderer := &contractRenderer{}
	if err := registry.SaveViaExtension(backend, renderer, "plot.pdf", render.WithPDFMetadata(map[string]string{"Title": "Registry"})); err != nil {
		t.Fatalf("SaveViaExtension PDF returned error: %v", err)
	}
	if got := renderer.pdfOptions.Metadata["Title"]; got != "Registry" {
		t.Fatalf("PDF metadata = %q, want Registry", got)
	}

	if err := registry.SaveViaExtension(backend, renderer, "plot.pgf", render.WithPGFPreamble("\\usepackage{amsmath}")); err != nil {
		t.Fatalf("SaveViaExtension PGF returned error: %v", err)
	}
	if len(renderer.pgfOptions.Preamble) != 1 || renderer.pgfOptions.Preamble[0] != "\\usepackage{amsmath}" {
		t.Fatalf("PGF preamble = %#v, want amsmath", renderer.pgfOptions.Preamble)
	}
}

func TestRegistrySaveViaExtensionRejectsMismatchedSharedOptions(t *testing.T) {
	registry := NewRegistry()
	backend := Backend("contract")
	registry.Register(backend, &BackendInfo{
		Name:      "Contract",
		Available: true,
		SaveFormats: map[string]SaveHandler{
			".svg": SaveSVG,
		},
	})

	renderer := &contractRenderer{}
	err := registry.SaveViaExtension(backend, renderer, "plot.svg", render.WithPDFMetadata(map[string]string{"Title": "Wrong"}))
	if err == nil {
		t.Fatal("SaveViaExtension accepted mismatched PDF options for SVG output")
	}
	if renderer.svgPath != "" || renderer.pdfPath != "" {
		t.Fatalf("renderer should not have saved after mismatched options: svg=%q pdf=%q", renderer.svgPath, renderer.pdfPath)
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
