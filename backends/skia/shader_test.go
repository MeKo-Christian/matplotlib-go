//go:build skia

package skia

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/backends"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestSkiaTaggedRendererExposesCPUSurfaceBridge(t *testing.T) {
	r, err := New(backends.TestDefaultConfig(16, 12))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	info := r.BridgeInfo()
	if info.Mode != ModeCPU {
		t.Fatalf("BridgeInfo().Mode = %q, want %q", info.Mode, ModeCPU)
	}
	if info.Binding != BindingExternalCAPI {
		t.Fatalf("BridgeInfo().Binding = %q, want %q", info.Binding, BindingExternalCAPI)
	}
	if !info.SupportsShaders {
		t.Fatal("CPU bridge should report shader support")
	}
}

func TestSkiaTaggedRendererAdvertisesShaderFillCapabilities(t *testing.T) {
	r := mustNewShaderRenderer(t, 10, 10)
	gradient, ok := any(r).(render.GradientFiller)
	if !ok {
		t.Fatal("Skia renderer should implement render.GradientFiller")
	}
	if !gradient.SupportsGradientFill() {
		t.Fatal("Skia renderer should support gradient fills")
	}
	pattern, ok := any(r).(render.PatternFiller)
	if !ok {
		t.Fatal("Skia renderer should implement render.PatternFiller")
	}
	if !pattern.SupportsPatternFill() {
		t.Fatal("Skia renderer should support pattern fills")
	}
}

func TestSkiaLinearGradientFillProducesGradientAcrossX(t *testing.T) {
	r := mustNewShaderRenderer(t, 40, 10)
	mustBeginShaderRenderer(t, r, 40, 10)

	r.Path(rectPath(0, 0, 40, 10), &render.Paint{
		FillGradient: render.GradientFill{
			Kind:  render.LinearGradient,
			Start: geom.Pt{X: 0, Y: 5},
			End:   geom.Pt{X: 40, Y: 5},
			Stops: []render.GradientStop{
				{Offset: 0, Color: render.Color{R: 1, A: 1}},
				{Offset: 1, Color: render.Color{B: 1, A: 1}},
			},
		},
	})

	mustEndShaderRenderer(t, r)
	leftR, _, leftB, _ := skiaPixelAt(t, r, 2, 5)
	rightR, _, rightB, _ := skiaPixelAt(t, r, 37, 5)
	if leftR < 200 || leftB > 80 {
		t.Fatalf("left edge should be mostly red, got R=%d B=%d", leftR, leftB)
	}
	if rightB < 200 || rightR > 80 {
		t.Fatalf("right edge should be mostly blue, got R=%d B=%d", rightR, rightB)
	}
}

func TestSkiaRadialGradientFillProducesCenterToEdgeFalloff(t *testing.T) {
	r := mustNewShaderRenderer(t, 40, 40)
	mustBeginShaderRenderer(t, r, 40, 40)

	r.Path(rectPath(0, 0, 40, 40), &render.Paint{
		FillGradient: render.GradientFill{
			Kind:   render.RadialGradient,
			Center: geom.Pt{X: 20, Y: 20},
			Radius: 20,
			Stops: []render.GradientStop{
				{Offset: 0, Color: render.Color{R: 1, G: 1, B: 1, A: 1}},
				{Offset: 1, Color: render.Color{A: 1}},
			},
		},
	})

	mustEndShaderRenderer(t, r)
	cR, cG, cB, _ := skiaPixelAt(t, r, 20, 20)
	eR, eG, eB, _ := skiaPixelAt(t, r, 0, 0)
	if cR < 200 || cG < 200 || cB < 200 {
		t.Fatalf("expected near-white center, got (%d,%d,%d)", cR, cG, cB)
	}
	if eR > 80 || eG > 80 || eB > 80 {
		t.Fatalf("expected dark edge, got (%d,%d,%d)", eR, eG, eB)
	}
}

func TestSkiaGradientFillHonorsStopOpacity(t *testing.T) {
	r := mustNewShaderRenderer(t, 20, 10)
	mustBeginShaderRenderer(t, r, 20, 10)

	r.Path(rectPath(0, 0, 20, 10), &render.Paint{
		FillGradient: render.GradientFill{
			Kind:  render.LinearGradient,
			Start: geom.Pt{X: 0, Y: 5},
			End:   geom.Pt{X: 20, Y: 5},
			Stops: []render.GradientStop{
				{Offset: 0, Color: render.Color{R: 1, A: 0.5}},
				{Offset: 1, Color: render.Color{R: 1, A: 0.5}},
			},
		},
	})

	mustEndShaderRenderer(t, r)
	rv, gv, bv, av := skiaPixelAt(t, r, 10, 5)
	if rv < 240 || gv < 100 || gv > 180 || bv < 100 || bv > 180 || av != 255 {
		t.Fatalf("half-alpha red over white should produce pink, got RGBA=(%d,%d,%d,%d)", rv, gv, bv, av)
	}
}

func TestSkiaGradientFillAppliesTransform(t *testing.T) {
	r := mustNewShaderRenderer(t, 40, 10)
	mustBeginShaderRenderer(t, r, 40, 10)

	r.Path(rectPath(0, 0, 40, 10), &render.Paint{
		FillGradient: render.GradientFill{
			Kind:         render.LinearGradient,
			Start:        geom.Pt{X: 0, Y: 5},
			End:          geom.Pt{X: 20, Y: 5},
			Transform:    geom.Affine{A: 1, D: 1, E: 20},
			HasTransform: true,
			Stops: []render.GradientStop{
				{Offset: 0, Color: render.Color{R: 1, A: 1}},
				{Offset: 1, Color: render.Color{B: 1, A: 1}},
			},
		},
	})

	mustEndShaderRenderer(t, r)
	leftR, _, leftB, _ := skiaPixelAt(t, r, 2, 5)
	shiftR, _, shiftB, _ := skiaPixelAt(t, r, 22, 5)
	if leftR < 200 || leftB > 80 {
		t.Fatalf("before transformed gradient start should clamp red, got R=%d B=%d", leftR, leftB)
	}
	if shiftR < 200 || shiftB > 80 {
		t.Fatalf("transformed gradient start should be red, got R=%d B=%d", shiftR, shiftB)
	}
}

func TestSkiaPatternFillRepeatsTileAcrossPath(t *testing.T) {
	r := mustNewShaderRenderer(t, 24, 12)
	mustBeginShaderRenderer(t, r, 24, 12)

	r.Path(rectPath(0, 0, 24, 12), &render.Paint{
		FillPattern: render.PatternFill{
			ID:         "stripe",
			Cell:       geom.Rect{Max: geom.Pt{X: 8, Y: 12}},
			Path:       rectPath(0, 0, 4, 12),
			Foreground: render.Color{R: 1, A: 1},
			Background: render.Color{B: 1, A: 1},
		},
	})

	mustEndShaderRenderer(t, r)
	leftR, _, leftB, _ := skiaPixelAt(t, r, 2, 6)
	gapR, _, gapB, _ := skiaPixelAt(t, r, 6, 6)
	repeatR, _, repeatB, _ := skiaPixelAt(t, r, 10, 6)
	if leftR < 200 || leftB > 80 {
		t.Fatalf("first stripe should be red, got R=%d B=%d", leftR, leftB)
	}
	if gapB < 200 || gapR > 80 {
		t.Fatalf("gap should be blue, got R=%d B=%d", gapR, gapB)
	}
	if repeatR < 200 || repeatB > 80 {
		t.Fatalf("repeated stripe should be red, got R=%d B=%d", repeatR, repeatB)
	}
}

func TestSkiaPatternFillAppliesTileTransform(t *testing.T) {
	r := mustNewShaderRenderer(t, 24, 12)
	mustBeginShaderRenderer(t, r, 24, 12)

	r.Path(rectPath(0, 0, 24, 12), &render.Paint{
		FillPattern: render.PatternFill{
			ID:           "shifted-stripe",
			Cell:         geom.Rect{Max: geom.Pt{X: 8, Y: 12}},
			Path:         rectPath(0, 0, 4, 12),
			Foreground:   render.Color{R: 1, A: 1},
			Background:   render.Color{B: 1, A: 1},
			Transform:    geom.Affine{A: 1, D: 1, E: 4},
			HasTransform: true,
		},
	})

	mustEndShaderRenderer(t, r)
	gapR, _, gapB, _ := skiaPixelAt(t, r, 2, 6)
	shiftR, _, shiftB, _ := skiaPixelAt(t, r, 6, 6)
	if gapB < 200 || gapR > 80 {
		t.Fatalf("unshifted stripe location should be blue, got R=%d B=%d", gapR, gapB)
	}
	if shiftR < 200 || shiftB > 80 {
		t.Fatalf("transformed stripe location should be red, got R=%d B=%d", shiftR, shiftB)
	}
}

func TestSkiaHatchTakesPrecedenceOverPatternFill(t *testing.T) {
	r := mustNewShaderRenderer(t, 24, 12)
	mustBeginShaderRenderer(t, r, 24, 12)

	r.Path(rectPath(0, 0, 24, 12), &render.Paint{
		Hatch:          "|",
		HatchColor:     render.Color{G: 1, A: 1},
		HatchLineWidth: 2,
		HatchSpacing:   6,
		FillPattern: render.PatternFill{
			ID:         "red-tile",
			Cell:       geom.Rect{Max: geom.Pt{X: 8, Y: 12}},
			Path:       rectPath(0, 0, 8, 12),
			Foreground: render.Color{R: 1, A: 1},
		},
	})

	mustEndShaderRenderer(t, r)
	redFound := false
	greenFound := false
	for y := 1; y < 11; y++ {
		for x := 1; x < 23; x++ {
			rv, gv, bv, _ := skiaPixelAt(t, r, x, y)
			if rv > 180 && gv < 80 && bv < 80 {
				redFound = true
			}
			if gv > 180 && rv < 120 && bv < 120 {
				greenFound = true
			}
		}
	}
	if redFound {
		t.Fatal("pattern fill painted red pixels even though hatch should take precedence")
	}
	if !greenFound {
		t.Fatal("expected hatch to paint green pixels")
	}
}

func TestSkiaSolidFillStillWorksAfterGradient(t *testing.T) {
	r := mustNewShaderRenderer(t, 40, 20)
	mustBeginShaderRenderer(t, r, 40, 20)

	r.Path(rectPath(0, 0, 20, 20), &render.Paint{
		FillGradient: render.GradientFill{
			Kind:  render.LinearGradient,
			Start: geom.Pt{X: 0, Y: 10},
			End:   geom.Pt{X: 20, Y: 10},
			Stops: []render.GradientStop{
				{Offset: 0, Color: render.Color{R: 1, A: 1}},
				{Offset: 1, Color: render.Color{G: 1, A: 1}},
			},
		},
	})
	r.Path(rectPath(20, 0, 40, 20), &render.Paint{Fill: render.Color{B: 1, A: 1}})

	mustEndShaderRenderer(t, r)
	rv, _, bv, _ := skiaPixelAt(t, r, 30, 10)
	if bv < 200 || rv > 60 {
		t.Fatalf("right half should be solid blue, got R=%d B=%d", rv, bv)
	}
}

func TestSkiaTaggedRegistryAdvertisesShaderCapabilities(t *testing.T) {
	renderer, err := backends.Create(backends.Skia, backends.TestDefaultConfig(32, 24))
	if err != nil {
		t.Fatalf("Create(skia) error = %v", err)
	}
	for _, cap := range []backends.Capability{backends.PatternFill, backends.GradientFill} {
		if status := backends.RendererCapabilityStatus(backends.Skia, renderer, cap); status != backends.CapabilityNative {
			t.Fatalf("RendererCapabilityStatus(skia, %s) = %s, want native", cap, status)
		}
	}
}

func mustNewShaderRenderer(t *testing.T, w, h int) *Renderer {
	t.Helper()
	r, err := New(backends.TestDefaultConfig(w, h))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return r
}

func mustBeginShaderRenderer(t *testing.T, r *Renderer, w, h int) {
	t.Helper()
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: float64(w), Y: float64(h)}}); err != nil {
		t.Fatalf("Begin() error = %v", err)
	}
}

func mustEndShaderRenderer(t *testing.T, r *Renderer) {
	t.Helper()
	if err := r.End(); err != nil {
		t.Fatalf("End() error = %v", err)
	}
}

func rectPath(x0, y0, x1, y1 float64) geom.Path {
	var p geom.Path
	p.MoveTo(geom.Pt{X: x0, Y: y0})
	p.LineTo(geom.Pt{X: x1, Y: y0})
	p.LineTo(geom.Pt{X: x1, Y: y1})
	p.LineTo(geom.Pt{X: x0, Y: y1})
	p.Close()
	return p
}

func skiaPixelAt(t *testing.T, r *Renderer, x, y int) (uint8, uint8, uint8, uint8) {
	t.Helper()
	img := r.Image()
	if img == nil {
		t.Fatal("nil image")
	}
	c := img.RGBAAt(x, y)
	return c.R, c.G, c.B, c.A
}
