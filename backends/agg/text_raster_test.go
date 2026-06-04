package agg

import (
	"bytes"
	"math"
	"testing"

	"codeberg.org/go-fonts/dejavu/dejavusans"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestRasterTextUsesEmbeddedFontFace(t *testing.T) {
	r, err := New(220, 120, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	face := render.FontFace{Family: "DejaVu Sans", Data: dejavusans.TTF}
	metrics, ok := r.measureRasterText("Ag", face, 24)
	if !ok {
		t.Fatal("measureRasterText with embedded face failed")
	}
	if metrics.W <= 0 || metrics.H <= 0 {
		t.Fatalf("embedded raster metrics = %+v, want positive dimensions", metrics)
	}
	if _, ok := rasterFontHeightMetrics(face, 24, 72); !ok {
		t.Fatal("rasterFontHeightMetrics with embedded face failed")
	}

	if !r.drawRasterText("Ag", face, geom.Pt{X: 20, Y: 60}, 24, render.Color{R: 0, G: 0, B: 0, A: 1}) {
		t.Fatal("drawRasterText with embedded face failed")
	}
}

func TestDrawRasterTextUsesSharedCombiningMarkShape(t *testing.T) {
	face := render.FontFace{Family: "DejaVu Sans", Data: dejavusans.TTF}
	decomposed := renderRasterTextPixels(t, "e\u0301", face)
	precomposed := renderRasterTextPixels(t, "\u00e9", face)

	if !bytes.Equal(decomposed, precomposed) {
		t.Fatal("decomposed e-acute raster output differs from precomposed e-acute")
	}
}

func TestMeasureRasterTextUsesMatplotlibPlainTextAdvance(t *testing.T) {
	r, err := New(220, 120, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	face := render.FontFace{Family: "DejaVu Sans", Data: dejavusans.TTF}
	const size = 72.0
	shaped, ok := render.ShapeText("fi", geom.Pt{}, r.fontPixelSize(size), matplotlibPlainTextShapingOptions(fontReference(face)))
	if !ok || len(shaped.Glyphs) != 2 {
		t.Fatalf("plain text ShapeText(fi) = %+v, %v; want separate f and i glyphs", shaped, ok)
	}

	metrics, ok := r.measureRasterText("fi", face, size)
	if !ok {
		t.Fatal("measureRasterText(fi) failed")
	}
	if metrics.W != quantize(shaped.Advance.X) {
		t.Fatalf("measureRasterText(fi).W = %v, want shaped advance %v", metrics.W, quantize(shaped.Advance.X))
	}
}

func TestMeasureTextBoundsUsesMatplotlibPlainTextBounds(t *testing.T) {
	r, err := New(220, 120, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	const size = 72.0
	font := r.configureTextFont(size, "")
	if font.backend != textBackendRaster {
		t.Fatal("expected raster text backend")
	}
	shaped, ok := render.ShapeText("fi", geom.Pt{}, r.fontPixelSize(font.size), matplotlibPlainTextShapingOptions(fontReference(font.face)))
	if !ok || len(shaped.Glyphs) != 2 {
		t.Fatalf("plain text ShapeText(fi) = %+v, %v; want separate f and i glyphs", shaped, ok)
	}

	bounds, ok := r.MeasureTextBounds("fi", size, "")
	if !ok {
		t.Fatal("MeasureTextBounds(fi) failed")
	}
	nativeBounds, ok := r.measureNativeFreetypeTextBounds("fi", font.face, size, matplotlibTextHintingFactor)
	if !ok {
		t.Fatal("native MeasureTextBounds(fi) failed")
	}
	if math.Abs(bounds.X-nativeBounds.X) > 1e-9 ||
		math.Abs(bounds.Y-nativeBounds.Y) > 1e-9 ||
		math.Abs(bounds.W-nativeBounds.W) > 1e-9 ||
		math.Abs(bounds.H-nativeBounds.H) > 1e-9 {
		t.Fatalf("MeasureTextBounds(fi) = %+v, want native bounds %+v", bounds, nativeBounds)
	}
}

func renderRasterTextPixels(t *testing.T, text string, face render.FontFace) []byte {
	t.Helper()

	r, err := New(180, 120, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if !r.drawRasterText(text, face, geom.Pt{X: 20, Y: 80}, 48, render.Color{R: 0, G: 0, B: 0, A: 1}) {
		t.Fatalf("drawRasterText(%q) failed", text)
	}
	return append([]byte(nil), r.GetImage().Pix...)
}
