package agg

import (
	"image/color"
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestDrawTextWithFontDoesNotMutateLegacyFontState(t *testing.T) {
	r := mustNew(t, 120, 80)
	r.lastFontKey = "legacy-font"

	if err := r.Begin(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 120, Y: 80}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	r.DrawTextWithFont("explicit", geom.Pt{X: 10, Y: 35}, 12, render.Color{A: 1}, "DejaVu Sans")
	r.DrawTextRotatedWithFont("rotated", geom.Pt{X: 60, Y: 50}, 12, math.Pi/8, render.Color{A: 1}, "DejaVu Sans")
	r.DrawTextVerticalWithFont("vertical", geom.Pt{X: 90, Y: 50}, 12, render.Color{A: 1}, "DejaVu Sans")
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	if r.lastFontKey != "legacy-font" {
		t.Fatalf("lastFontKey = %q, want legacy-font", r.lastFontKey)
	}
}

func TestMeasureText(t *testing.T) {
	r := mustNew(t, 100, 100)
	m := r.MeasureText("Hello", 12.0, "")
	if m.W <= 0 || m.H <= 0 {
		t.Errorf("text metrics should be positive: W=%f H=%f", m.W, m.H)
	}

	empty := r.MeasureText("", 12.0, "")
	if empty.W != 0 || empty.H != 0 {
		t.Errorf("empty text should have zero metrics")
	}
}

func TestGlyphRunRendersShapedGlyphs(t *testing.T) {
	r := mustNew(t, 220, 120)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 220, Y: 120}}
	if err := r.Begin(viewport); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	layout, ok := render.LayoutTextGlyphs("Ag", geom.Pt{}, 24, "DejaVu Sans")
	if !ok {
		t.Fatal("LayoutTextGlyphs(\"Ag\") failed")
	}
	run := render.GlyphRun{
		Size:    24,
		Origin:  geom.Pt{X: 40, Y: 80},
		FontKey: "DejaVu Sans",
		Glyphs:  make([]render.Glyph, 0, len(layout.Glyphs)),
	}
	for _, glyph := range layout.Glyphs {
		run.Glyphs = append(run.Glyphs, render.Glyph{
			ID:     uint32(glyph.GlyphIndex),
			Offset: glyph.Origin,
		})
	}

	r.GlyphRun(run, render.Color{A: 1})
	if err := r.End(); err != nil {
		t.Fatalf("End failed: %v", err)
	}

	if _, _, ok := inkBounds(r.GetImage(), color.RGBA{R: 255, G: 255, B: 255, A: 255}); !ok {
		t.Fatal("GlyphRun should draw visible text from shaped glyph IDs")
	}
}

func TestMeasureTextScalesWithResolution(t *testing.T) {
	r := mustNew(t, 100, 100)

	r.SetResolution(72)
	width72 := r.MeasureText("Hello", 12, "").W

	r.SetResolution(100)
	width100 := r.MeasureText("Hello", 12, "").W

	if width72 <= 0 || width100 <= 0 {
		t.Fatalf("expected positive widths, got 72dpi=%v 100dpi=%v", width72, width100)
	}
	if width100 <= width72 {
		t.Fatalf("expected width to increase with DPI, got 72dpi=%v 100dpi=%v", width72, width100)
	}
}

func TestDrawTextRotatedMaintainsReadableFootprint(t *testing.T) {
	r := mustNew(t, 220, 220)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 220, Y: 220}}
	if err := r.Begin(viewport); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	const size = 24.0
	metrics := r.MeasureText("Amplitude", size, "")
	if metrics.W <= 0 || metrics.H <= 0 {
		t.Fatalf("expected positive text metrics, got %+v", metrics)
	}

	r.DrawTextRotated("Amplitude", geom.Pt{X: 72, Y: 160}, size, math.Pi/2, render.Color{R: 0, G: 0, B: 0, A: 1})
	if err := r.End(); err != nil {
		t.Fatalf("End failed: %v", err)
	}

	bounds, pixels, ok := inkBounds(r.GetImage(), color.RGBA{R: 255, G: 255, B: 255, A: 255})
	if !ok {
		t.Fatal("expected rotated text to draw visible ink")
	}
	if bounds.W() < metrics.H*0.75 {
		t.Fatalf("rotated text width too small: got=%v want_at_least=%v bounds=%+v", bounds.W(), metrics.H*0.75, bounds)
	}
	if bounds.H() < metrics.W*0.65 {
		t.Fatalf("rotated text height too small: got=%v want_at_least=%v bounds=%+v", bounds.H(), metrics.W*0.65, bounds)
	}
	if pixels < 250 {
		t.Fatalf("rotated text ink coverage unexpectedly sparse: pixels=%d bounds=%+v", pixels, bounds)
	}
}

func TestDrawTextRotatedMatchesMatplotlibRightYLabelInkBounds(t *testing.T) {
	r := mustNew(t, 640, 360)
	r.SetResolution(100)
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 640, Y: 360}}); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	r.DrawTextRotatedWithFont("log value", geom.Pt{X: 564.0331732855902, Y: 178.2}, 10, math.Pi/2, render.Color{A: 1}, "DejaVu Sans")
	if err := r.End(); err != nil {
		t.Fatalf("End failed: %v", err)
	}

	bounds, pixels, ok := inkBounds(r.GetImage(), color.RGBA{R: 255, G: 255, B: 255, A: 255})
	if !ok || pixels == 0 {
		t.Fatal("expected rotated label ink")
	}
	want := geom.Rect{Min: geom.Pt{X: 555, Y: 152}, Max: geom.Pt{X: 569, Y: 214}}
	if bounds != want {
		t.Fatalf("rotated right-y label ink bounds = %v, want matplotlib %v (pixels=%d)", bounds, want, pixels)
	}
}

func TestDrawTextUsesMatplotlibRoundHalfEvenForBitmapOrigin(t *testing.T) {
	r := mustNew(t, 120, 80)
	r.SetResolution(100)
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 120, Y: 80}}); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	r.DrawTextWithFont("0", geom.Pt{X: 59.625, Y: 40}, 10, render.Color{A: 1}, "DejaVu Sans")
	if err := r.End(); err != nil {
		t.Fatalf("End failed: %v", err)
	}

	bounds, _, ok := inkBounds(r.GetImage(), color.RGBA{R: 255, G: 255, B: 255, A: 255})
	if !ok {
		t.Fatal("expected visible text ink")
	}
	if got, want := bounds.Min.X, 60.0; got != want {
		t.Fatalf("text ink min x = %v, want matplotlib round-half-even placement %v", got, want)
	}
}
