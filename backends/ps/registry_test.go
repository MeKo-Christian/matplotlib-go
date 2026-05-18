package ps_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/ps"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestPSBackendIsRegistered(t *testing.T) {
	info, ok := backends.DefaultRegistry.Get(backends.PS)
	if !ok {
		t.Fatalf("PS backend not registered")
	}
	if !info.Available {
		t.Errorf("PS backend should be available")
	}
	if _, ok := info.SaveFormats[".ps"]; !ok {
		t.Errorf("expected .ps in SaveFormats")
	}
	if _, ok := info.SaveFormats[".eps"]; !ok {
		t.Errorf("expected .eps in SaveFormats")
	}
}

func TestPSBackend_AdvertisedCapabilitiesAreImplemented(t *testing.T) {
	r, err := backends.Create(backends.PS, backends.Config{
		Width: 200, Height: 100,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := backends.DefaultRegistry.VerifyRendererCapabilities(backends.PS, r); err != nil {
		t.Errorf("VerifyRendererCapabilities: %v", err)
	}
}

func TestPSBackendAdvertisesImageTransform(t *testing.T) {
	r, err := backends.Create(backends.PS, backends.Config{
		Width: 200, Height: 100,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	status := backends.DefaultRegistry.RendererCapabilityStatus(backends.PS, r, backends.ImageTransform)
	if status != backends.CapabilityNative {
		t.Fatalf("ImageTransform status = %s, want %s", status, backends.CapabilityNative)
	}
}

func TestPSBackendAdvertisesNativeHatcher(t *testing.T) {
	r, err := backends.Create(backends.PS, backends.Config{
		Width: 200, Height: 100,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	status := backends.DefaultRegistry.RendererCapabilityStatus(backends.PS, r, backends.NativeHatcher)
	if status != backends.CapabilityNative {
		t.Fatalf("NativeHatcher status = %s, want %s", status, backends.CapabilityNative)
	}
}

func TestPSBackendAdvertisesMarkerBatch(t *testing.T) {
	r, err := backends.Create(backends.PS, backends.Config{
		Width: 200, Height: 100,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	status := backends.DefaultRegistry.RendererCapabilityStatus(backends.PS, r, backends.MarkerBatch)
	if status != backends.CapabilityNative {
		t.Fatalf("MarkerBatch status = %s, want %s", status, backends.CapabilityNative)
	}
}

func TestPSBackendAdvertisesPathCollectionBatch(t *testing.T) {
	r, err := backends.Create(backends.PS, backends.Config{
		Width: 200, Height: 100,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	status := backends.DefaultRegistry.RendererCapabilityStatus(backends.PS, r, backends.PathCollectionBatch)
	if status != backends.CapabilityNative {
		t.Fatalf("PathCollectionBatch status = %s, want %s", status, backends.CapabilityNative)
	}
}

func TestSavePSViaRegistry(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "registry.ps")

	r, err := backends.Create(backends.PS, backends.Config{
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
	if err := backends.DefaultRegistry.SaveViaExtension(backends.PS, r, out); err != nil {
		t.Fatalf("SaveViaExtension: %v", err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%!PS-Adobe-3.0")) {
		t.Errorf("missing PostScript header")
	}
	if !bytes.Contains(data, []byte("%%BoundingBox: 0 0 200 100")) {
		t.Errorf("missing bounding box")
	}
	if !bytes.HasSuffix(data, []byte("%%EOF\n")) {
		t.Errorf("missing %%%%EOF trailer")
	}
}
