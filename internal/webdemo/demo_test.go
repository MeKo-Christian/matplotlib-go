package webdemo

import (
	"bytes"
	"image/png"
	"strings"
	"testing"
)

func TestCatalogStable(t *testing.T) {
	got := Catalog()
	wantIDs := []string{
		"lines",
		"scatter",
		"bars",
		"fills",
		"errorbars",
		"histogram",
		"heatmap",
		"axes",
		"composition",
		"subplots",
		"annotations",
		"patches",
		"mesh",
		"variants",
		"statistics",
		"specialty",
		"units",
		"vectors",
		"polar",
		"projections",
		"matrix",
		"animation",
	}
	if len(got) != len(wantIDs) {
		t.Fatalf("Catalog() len = %d, want %d", len(got), len(wantIDs))
	}

	for i, want := range wantIDs {
		if got[i].ID != want {
			t.Fatalf("Catalog()[%d].ID = %q, want %q", i, got[i].ID, want)
		}
		if got[i].Title == "" {
			t.Fatalf("Catalog()[%d].Title is empty", i)
		}
		if got[i].Description == "" {
			t.Fatalf("Catalog()[%d].Description is empty", i)
		}
	}
}

func TestAnimationGalleryAvailableInWASMDemoCatalog(t *testing.T) {
	if !ValidDemoID("animation") {
		t.Fatal("animation demo should be available in the WASM demo catalog")
	}
	fig, descriptor, err := Build("animation", 0, 0)
	if err != nil {
		t.Fatalf("Build(animation): %v", err)
	}
	if descriptor.ID != "animation" {
		t.Fatalf("descriptor.ID = %q, want animation", descriptor.ID)
	}
	if fig == nil || len(fig.Children) == 0 {
		t.Fatal("Build(animation) returned an empty figure")
	}
}

func TestRenderProducesImage(t *testing.T) {
	img, descriptor, err := Render("axes", 320, 180)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if descriptor.ID != "axes" {
		t.Fatalf("Render() descriptor ID = %q, want %q", descriptor.ID, "axes")
	}
	bounds := img.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		t.Fatalf("Render() bounds = %v, want positive", bounds)
	}

	allWhite := true
	for i := 0; i < len(img.Pix); i += 4 {
		if img.Pix[i] != 255 || img.Pix[i+1] != 255 || img.Pix[i+2] != 255 || img.Pix[i+3] != 255 {
			allWhite = false
			break
		}
	}
	if allWhite {
		t.Fatal("Render() returned an all-white image")
	}
}

func TestSourceUsesCanonicalParityExample(t *testing.T) {
	source, descriptor, err := Source("matrix")
	if err != nil {
		t.Fatalf("Source(matrix) error = %v", err)
	}
	if descriptor.ID != "matrix" {
		t.Fatalf("descriptor.ID = %q, want matrix", descriptor.ID)
	}
	if !strings.Contains(source, "examples/arrays_showcase/example.go") {
		t.Fatalf("Source(matrix) did not identify canonical showcase source path:\n%s", source)
	}
}

func TestBackendsStable(t *testing.T) {
	got := Backends()
	wantIDs := []string{"gobasic", "agg"}
	if ValidBackendID("skia") {
		wantIDs = append(wantIDs, "skia")
	}
	if len(got) != len(wantIDs) {
		t.Fatalf("Backends() len = %d, want %d", len(got), len(wantIDs))
	}
	for i, want := range wantIDs {
		if got[i].ID != want {
			t.Fatalf("Backends()[%d].ID = %q, want %q", i, got[i].ID, want)
		}
		if got[i].Name == "" {
			t.Fatalf("Backends()[%d].Name is empty", i)
		}
		if got[i].Description == "" {
			t.Fatalf("Backends()[%d].Description is empty", i)
		}
	}
}

func TestRenderWithBackendProducesImage(t *testing.T) {
	for _, backendID := range []string{"agg", "gobasic"} {
		img, descriptor, err := RenderWithBackend("axes", backendID, 320, 180)
		if err != nil {
			t.Fatalf("RenderWithBackend(%q) error = %v", backendID, err)
		}
		if descriptor.ID != "axes" {
			t.Fatalf("RenderWithBackend(%q) descriptor ID = %q, want axes", backendID, descriptor.ID)
		}
		if img == nil || img.Bounds().Empty() {
			t.Fatalf("RenderWithBackend(%q) returned empty image", backendID)
		}
	}
}

func TestRenderPNGProducesPNGBytes(t *testing.T) {
	data, descriptor, err := RenderPNG("polar", 320, 180)
	if err != nil {
		t.Fatalf("RenderPNG() error = %v", err)
	}
	if descriptor.ID != "polar" {
		t.Fatalf("descriptor.ID = %q, want polar", descriptor.ID)
	}
	if !bytes.HasPrefix(data, []byte("\x89PNG\r\n\x1a\n")) {
		t.Fatalf("RenderPNG() did not return PNG signature: %x", data[:8])
	}
	if _, err := png.Decode(bytes.NewReader(data)); err != nil {
		t.Fatalf("RenderPNG() returned undecodable PNG: %v", err)
	}
}

func TestBuildRejectsUnknownDemo(t *testing.T) {
	if _, _, err := Build("nope", 100, 100); err == nil {
		t.Fatal("Build() for unknown demo returned nil error")
	}
}

func TestDefaultDemoIDAndValidDemoID(t *testing.T) {
	id := DefaultDemoID()
	if id == "" {
		t.Fatal("DefaultDemoID() is empty")
	}
	if !ValidDemoID(id) {
		t.Fatalf("ValidDemoID(%q) = false, want true", id)
	}
	if ValidDemoID("not-a-demo") {
		t.Fatal("ValidDemoID(\"not-a-demo\") = true, want false")
	}
}

func TestDefaultBackendIDAndValidBackendID(t *testing.T) {
	id := DefaultBackendID()
	if id == "" {
		t.Fatal("DefaultBackendID() is empty")
	}
	if !ValidBackendID(id) {
		t.Fatalf("ValidBackendID(%q) = false, want true", id)
	}
	if ValidBackendID("not-a-backend") {
		t.Fatal("ValidBackendID(\"not-a-backend\") = true, want false")
	}
}

func TestSourceReturnsCanonicalParitySource(t *testing.T) {
	for _, descriptor := range Catalog() {
		source, got, err := Source(descriptor.ID)
		if err != nil {
			t.Fatalf("Source(%q) error = %v", descriptor.ID, err)
		}
		if got.ID != descriptor.ID {
			t.Fatalf("Source(%q) descriptor ID = %q, want %q", descriptor.ID, got.ID, descriptor.ID)
		}
		if !strings.HasPrefix(source, "// examples/") {
			t.Fatalf("Source(%q) does not start with a canonical showcase source path: %.32q", descriptor.ID, source)
		}
		if !strings.Contains(source, "func Plot() *core.Figure") {
			t.Fatalf("Source(%q) does not include the canonical example Plot function", descriptor.ID)
		}
	}
}

func TestSourceRejectsUnknownDemo(t *testing.T) {
	if _, _, err := Source("nope"); err == nil {
		t.Fatal("Source() for unknown demo returned nil error")
	}
}

// TestBuildEachShowcaseProducesFigure smoke-tests every cataloged web demo:
// dispatching to the underlying showcase Plot() must succeed, return a figure
// with positive intrinsic size, and contain at least one axes.
func TestBuildEachShowcaseProducesFigure(t *testing.T) {
	for _, descriptor := range Catalog() {
		t.Run(descriptor.ID, func(t *testing.T) {
			fig, got, err := Build(descriptor.ID, 320, 180)
			if err != nil {
				t.Fatalf("Build(%q) error = %v", descriptor.ID, err)
			}
			if got.ID != descriptor.ID {
				t.Fatalf("descriptor.ID = %q, want %q", got.ID, descriptor.ID)
			}
			if fig == nil {
				t.Fatal("Build() returned nil figure")
			}
			if fig.SizePx.X <= 0 || fig.SizePx.Y <= 0 {
				t.Fatalf("fig.SizePx = (%v, %v), want positive", fig.SizePx.X, fig.SizePx.Y)
			}
			if len(fig.Children) == 0 {
				t.Fatal("fig has no axes")
			}
		})
	}
}

// TestRenderEachShowcaseProducesImage verifies every cataloged web demo
// renders to a non-empty, non-blank image via the default backend.
func TestRenderEachShowcaseProducesImage(t *testing.T) {
	for _, descriptor := range Catalog() {
		t.Run(descriptor.ID, func(t *testing.T) {
			img, _, err := Render(descriptor.ID, 0, 0)
			if err != nil {
				t.Fatalf("Render(%q) error = %v", descriptor.ID, err)
			}
			if img == nil || img.Bounds().Empty() {
				t.Fatalf("Render(%q) returned empty image", descriptor.ID)
			}
			allWhite := true
			for i := 0; i < len(img.Pix); i += 4 {
				if img.Pix[i] != 255 || img.Pix[i+1] != 255 || img.Pix[i+2] != 255 || img.Pix[i+3] != 255 {
					allWhite = false
					break
				}
			}
			if allWhite {
				t.Fatalf("Render(%q) returned an all-white image", descriptor.ID)
			}
		})
	}
}
