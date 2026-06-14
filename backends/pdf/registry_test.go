package pdf_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/pdf"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestPDFBackendIsRegistered(t *testing.T) {
	info, ok := backends.DefaultRegistry.Get(backends.PDF)
	if !ok {
		t.Fatalf("PDF backend not registered")
	}
	if !info.Available {
		t.Errorf("PDF backend should be available")
	}
	if _, ok := info.SaveFormats[".pdf"]; !ok {
		t.Errorf("expected .pdf in SaveFormats")
	}
}

func TestPDFBackend_AdvertisedCapabilitiesAreImplemented(t *testing.T) {
	r, err := backends.Create(backends.PDF, backends.Config{
		Width: 200, Height: 100,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := backends.DefaultRegistry.VerifyRendererCapabilities(backends.PDF, r); err != nil {
		t.Errorf("VerifyRendererCapabilities: %v", err)
	}
}

func TestSavePDFViaRegistry(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "registry.pdf")

	r, err := backends.Create(backends.PDF, backends.Config{
		Width: 200, Height: 100,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	var p geom.Path
	p.MoveTo(geom.Pt{X: 10, Y: 10})
	p.LineTo(geom.Pt{X: 50, Y: 50})
	r.Path(p, &render.Paint{
		Stroke:    render.Color{R: 0, G: 0, B: 0, A: 1},
		LineWidth: 1,
	})
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	if err := backends.DefaultRegistry.SaveViaExtension(backends.PDF, r, out); err != nil {
		t.Fatalf("SaveViaExtension: %v", err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF-1.7\n")) {
		t.Errorf("missing PDF header")
	}
	if !bytes.HasSuffix(data, []byte("%%EOF\n")) {
		t.Errorf("missing %%%%EOF trailer")
	}
}
