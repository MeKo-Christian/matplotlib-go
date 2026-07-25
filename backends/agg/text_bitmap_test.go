package agg

import (
	"image"
	"image/color"
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestBlendAlphaMaskAppliesTextAlphaAndClips(t *testing.T) {
	r := mustNew(t, 3, 3)
	mask := image.NewAlpha(image.Rect(0, 0, 3, 3))
	mask.SetAlpha(0, 0, color.Alpha{A: 255})
	mask.SetAlpha(1, 1, color.Alpha{A: 128})
	mask.SetAlpha(2, 2, color.Alpha{A: 255})

	if !r.blendAlphaMask(mask, -1, -1, render.Color{R: 1, G: 0, B: 0, A: 0.5}) {
		t.Fatal("blendAlphaMask returned false")
	}

	img := r.Image()
	if got := img.RGBAAt(0, 0); got.R != 255 || got.G < 188 || got.G > 193 || got.B < 188 || got.B > 193 || got.A != 255 {
		t.Fatalf("half-covered clipped text pixel = %+v, want red over white with alpha-scaled coverage", got)
	}
	if got := img.RGBAAt(1, 1); got.R != 255 || got.G < 126 || got.G > 129 || got.B < 126 || got.B > 129 || got.A != 255 {
		t.Fatalf("full-covered clipped text pixel = %+v, want red over white with text alpha", got)
	}
	if got := img.RGBAAt(2, 2); got != (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
		t.Fatalf("out-of-mask destination pixel changed: %+v", got)
	}
}

func TestPythonRoundTreatsFloatingHalfTiesLikeMatplotlibTextPlacement(t *testing.T) {
	if got := pythonRound(228.50000000000006); got != 228 {
		t.Fatalf("pythonRound near even half tie = %v, want 228", got)
	}
	if got := pythonRound(229.50000000000006); got != 230 {
		t.Fatalf("pythonRound near odd half tie = %v, want 230", got)
	}
	if got := pythonRound(228.5001); got != 229 {
		t.Fatalf("pythonRound non-tie = %v, want 229", got)
	}
}

func TestMathTextAlphaImageOneMinusLabelMatchesMatplotlibRasterBBox(t *testing.T) {
	r := mustNew(t, 120, 80)
	r.SetResolution(100)
	if _, ok := r.MeasureMathGlyphRun("1", 10, "DejaVu Sans"); !ok {
		t.Skip("pixel-exact MathText glyph metrics unavailable")
	}

	const fontKey = "DejaVu Sans"
	layout, ok := core.LayoutMathText(r, `\mathdefault{1-10^{-1}}`, 10, fontKey)
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}
	glyphs := make([]render.MathGlyphPlacement, 0, len(layout.Runs))
	for _, run := range layout.Runs {
		runFontKey := run.FontKey
		if runFontKey == "" {
			runFontKey = fontKey
		}
		glyphs = append(glyphs, render.MathGlyphPlacement{
			Text:     run.Text,
			FontSize: run.FontSize,
			FontKey:  runFontKey,
			Ox:       run.Offset.X,
			Oy:       run.Offset.Y,
		})
	}

	img, ok := r.mathTextAlphaImage(glyphs, nil, layout.Ascent, layout.Descent)
	if !ok {
		t.Fatal("mathTextAlphaImage returned !ok")
	}

	bounds := img.mask.Bounds()
	if bounds.Dx() != 60 || bounds.Dy() != 19 {
		t.Fatalf("mathtext mask size = %dx%d, want 60x19", bounds.Dx(), bounds.Dy())
	}
	ink, pixels, ok := alphaInkBounds(img.mask)
	if !ok {
		t.Fatal("mathtext mask has no ink")
	}
	// Matplotlib 3.10.9 MathTextParser("agg") RasterParse for
	// $\mathdefault{1-10^{-1}}$ at 10 pt / 100 dpi has image shape 60x19 and
	// nonzero alpha bbox (2, 1, 58, 14).
	if ink.Min.X != 2 || ink.Min.Y != 1 || ink.Max.X != 58 || ink.Max.Y != 14 {
		t.Fatalf("mathtext mask ink bbox = %+v (%d pixels), want min=(2,1) max=(58,14)", ink, pixels)
	}
}

func alphaInkBounds(img *image.Alpha) (renderBox, int, bool) {
	bounds := img.Bounds()
	minX, minY := bounds.Max.X, bounds.Max.Y
	maxX, maxY := bounds.Min.X, bounds.Min.Y
	pixels := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if img.AlphaAt(x, y).A == 0 {
				continue
			}
			pixels++
			if x < minX {
				minX = x
			}
			if y < minY {
				minY = y
			}
			if x >= maxX {
				maxX = x + 1
			}
			if y >= maxY {
				maxY = y + 1
			}
		}
	}
	if pixels == 0 {
		return renderBox{}, 0, false
	}
	return renderBox{Min: renderPoint{X: minX, Y: minY}, Max: renderPoint{X: maxX, Y: maxY}}, pixels, true
}

type renderPoint struct {
	X, Y int
}

type renderBox struct {
	Min, Max renderPoint
}
