package pgf_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/pgf"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestPGFBackendIsRegistered(t *testing.T) {
	info, ok := backends.DefaultRegistry.Get(backends.PGF)
	if !ok {
		t.Fatalf("PGF backend not registered")
	}
	if !info.Available {
		t.Errorf("PGF backend should be available")
	}
	if _, ok := info.SaveFormats[".pgf"]; !ok {
		t.Errorf("expected .pgf in SaveFormats")
	}
}

func TestPGFBackend_AdvertisedCapabilitiesAreImplemented(t *testing.T) {
	r, err := backends.Create(backends.PGF, backends.Config{
		Width: 200, Height: 100,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := backends.DefaultRegistry.VerifyRendererCapabilities(backends.PGF, r); err != nil {
		t.Errorf("VerifyRendererCapabilities: %v", err)
	}
}

func TestSavePGFViaRegistry(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "registry.pgf")

	r, err := backends.Create(backends.PGF, backends.Config{
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
	if err := backends.DefaultRegistry.SaveViaExtension(backends.PGF, r, out); err != nil {
		t.Fatalf("SaveViaExtension: %v", err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !bytes.Contains(data, []byte(`\begin{pgfpicture}`)) {
		t.Errorf("missing pgfpicture start")
	}
	if !bytes.Contains(data, []byte(`\pgfpathlineto`)) {
		t.Errorf("missing line path command")
	}
}

func TestSavePGFViaRegistryForwardsOptions(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "registry-options.pgf")

	r, err := backends.Create(backends.PGF, backends.Config{
		Width: 200, Height: 100,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	if err := backends.DefaultRegistry.SaveViaExtension(
		backends.PGF,
		r,
		out,
		render.WithPGFMetadata(map[string]string{"Title": "Registry"}),
		render.WithPGFPreamble("\\usepackage{amsmath}"),
		render.WithPGFVerificationMode(render.PGFVerificationModeStrict),
	); err != nil {
		t.Fatalf("SaveViaExtension: %v", err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	for _, want := range [][]byte{
		[]byte(`% metadata Title: Registry`),
		[]byte(`% preamble: \usepackage{amsmath}`),
		[]byte(`% verification: strict`),
	} {
		if !bytes.Contains(data, want) {
			t.Fatalf("missing %q in\n%s", want, data)
		}
	}
}
