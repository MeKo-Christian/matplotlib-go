package backends_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/agg" // register AGG backend
	_ "github.com/cwbudde/matplotlib-go/backends/svg" // register SVG backend
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestRegistry_SaveViaExtension_PNG(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.png")

	r, err := backends.Create(backends.AGG, backends.Config{Width: 100, Height: 80, DPI: 72})
	if err != nil {
		t.Fatalf("Create AGG: %v", err)
	}
	if err := backends.DefaultRegistry.SaveViaExtension(backends.AGG, r, path); err != nil {
		t.Fatalf("SaveViaExtension: %v", err)
	}
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if fi.Size() == 0 {
		t.Fatalf("expected non-empty PNG file at %s, got 0 bytes", path)
	}
}

func TestRegistry_SaveViaExtension_SVG(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.svg")

	r, err := backends.Create(backends.SVG, backends.Config{Width: 100, Height: 80, DPI: 72})
	if err != nil {
		t.Fatalf("Create SVG: %v", err)
	}
	if err := backends.DefaultRegistry.SaveViaExtension(backends.SVG, r, path); err != nil {
		t.Fatalf("SaveViaExtension: %v", err)
	}
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if fi.Size() == 0 {
		t.Fatalf("expected non-empty SVG file at %s, got 0 bytes", path)
	}
}

func TestRegistry_SaveViaExtension_ForwardsSVGOptions(t *testing.T) {
	r, err := backends.Create(backends.SVG, backends.Config{
		Width:      120,
		Height:     80,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		DPI:        72,
	})
	if err != nil {
		t.Fatalf("creating SVG renderer failed: %v", err)
	}
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 120, Y: 80}}); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	if err := r.End(); err != nil {
		t.Fatalf("End failed: %v", err)
	}

	path := filepath.Join(t.TempDir(), "registry-options.svg")
	if err := backends.DefaultRegistry.SaveViaExtension(
		backends.SVG,
		r,
		path,
		render.WithSVGMetadata(map[string]string{"Title": "Registry"}),
	); err != nil {
		t.Fatalf("SaveViaExtension: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(data), `<meta name="Title" content="Registry" />`) {
		t.Fatalf("registry save did not forward SVG metadata option, got %q", string(data))
	}
}
