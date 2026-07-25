package core_test

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func newSaveTestFigure(t *testing.T) (*core.Figure, *agg.Renderer) {
	t.Helper()
	fig := core.NewFigure(200, 150)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	_, _ = ax.Plot([]float64{0, 1, 2, 3}, []float64{0, 1, 0, 1})
	r, err := agg.New(200, 150, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatalf("agg.New: %v", err)
	}
	return fig, r
}

func decodePNG(t *testing.T, path string) image.Image {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return img
}

func TestSavePNGFacecolorOverride(t *testing.T) {
	fig, r := newSaveTestFigure(t)
	path := filepath.Join(t.TempDir(), "face.png")
	if err := core.SavePNG(fig, r, path, render.WithSaveFacecolor(render.Color{R: 1, G: 0, B: 0, A: 1})); err != nil {
		t.Fatalf("SavePNG: %v", err)
	}
	img := decodePNG(t, path)
	// A corner pixel (outside the axes) should be the red figure face color.
	rr, gg, bb, aa := img.At(2, 2).RGBA()
	if rr>>8 < 200 || gg>>8 > 60 || bb>>8 > 60 || aa>>8 < 200 {
		t.Fatalf("corner pixel = (%d,%d,%d,%d), want red", rr>>8, gg>>8, bb>>8, aa>>8)
	}
}

func TestSavePNGTransparent(t *testing.T) {
	fig, r := newSaveTestFigure(t)
	path := filepath.Join(t.TempDir(), "transparent.png")
	if err := core.SavePNG(fig, r, path, render.WithSaveTransparent(true)); err != nil {
		t.Fatalf("SavePNG: %v", err)
	}
	img := decodePNG(t, path)
	_, _, _, aa := img.At(2, 2).RGBA()
	if aa>>8 != 0 {
		t.Fatalf("corner alpha = %d, want 0 (transparent)", aa>>8)
	}
}

func TestSavePNGTightBboxShrinks(t *testing.T) {
	fig, r := newSaveTestFigure(t)
	path := filepath.Join(t.TempDir(), "tight.png")
	if err := core.SavePNG(fig, r, path, render.WithSaveBboxInches("tight"), render.WithSavePadInches(0)); err != nil {
		t.Fatalf("SavePNG: %v", err)
	}
	img := decodePNG(t, path)
	b := img.Bounds()
	if b.Dx() >= 200 || b.Dy() >= 150 {
		t.Fatalf("tight bbox = %dx%d, want smaller than 200x150", b.Dx(), b.Dy())
	}
	if b.Dx() <= 0 || b.Dy() <= 0 {
		t.Fatalf("tight bbox degenerate: %dx%d", b.Dx(), b.Dy())
	}
}

func TestSavePNGDPIRescalesCanvas(t *testing.T) {
	fig, r := newSaveTestFigure(t)
	// Default figure DPI is 100; doubling to 200 should double the pixel canvas.
	path := filepath.Join(t.TempDir(), "dpi.png")
	if err := core.SavePNG(fig, r, path, render.WithSaveDPI(200)); err != nil {
		t.Fatalf("SavePNG: %v", err)
	}
	img := decodePNG(t, path)
	b := img.Bounds()
	if b.Dx() != 400 || b.Dy() != 300 {
		t.Fatalf("dpi-scaled canvas = %dx%d, want 400x300", b.Dx(), b.Dy())
	}
}

func TestSaveFigFormatOverride(t *testing.T) {
	fig, r := newSaveTestFigure(t)
	// A path without a .png extension still saves as PNG via savefig.format.
	path := filepath.Join(t.TempDir(), "plot.out")
	if err := core.SaveFig(fig, r, path, render.WithSaveFormat("png")); err != nil {
		t.Fatalf("SaveFig: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected output file: %v", err)
	}
	decodePNG(t, path) // confirms PNG content
}
