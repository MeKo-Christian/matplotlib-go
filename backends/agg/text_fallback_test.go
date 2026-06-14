package agg

import (
	"image/color"
	"math"
	"testing"

	agglib "github.com/cwbudde/agg_go"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestGSVTextFallbackIsDisabledByDefault(t *testing.T) {
	r := mustNew(t, 120, 60)
	r.defaultFontFace = render.FontFace{Family: "Broken Emergency Test", Data: []byte("not a font")}
	r.fontPath = ""

	if err := r.Begin(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 120, Y: 60}}); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	r.DrawText("Fallback", geom.Pt{X: 8, Y: 32}, 18, render.Color{A: 1})
	if err := r.End(); err != nil {
		t.Fatalf("End failed: %v", err)
	}

	_, pixels, ok := inkBounds(r.GetImage(), color.RGBA{R: 255, G: 255, B: 255, A: 255})
	if ok || pixels != 0 {
		t.Fatalf("default text path should not draw GSV fallback pixels, got %d", pixels)
	}
	if r.fallback {
		t.Fatal("default text path should not mark GSV fallback as used")
	}
}

func TestGSVTextFallbackRequiresExplicitEmergencyOptIn(t *testing.T) {
	r := mustNew(t, 120, 60)
	r.defaultFontFace = render.FontFace{Family: "Broken Emergency Test", Data: []byte("not a font")}
	r.fontPath = ""
	r.SetEmergencyTextFallback(true)

	if err := r.Begin(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 120, Y: 60}}); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	r.DrawText("Fallback", geom.Pt{X: 8, Y: 32}, 18, render.Color{A: 1})
	if err := r.End(); err != nil {
		t.Fatalf("End failed: %v", err)
	}

	_, pixels, ok := inkBounds(r.GetImage(), color.RGBA{R: 255, G: 255, B: 255, A: 255})
	if !ok || pixels == 0 {
		t.Fatal("explicit emergency text fallback should draw GSV pixels")
	}
	if !r.fallback || !r.EmergencyTextFallbackUsed() {
		t.Fatal("expected emergency text fallback diagnostics to be marked")
	}
}

func TestGSVTextFallbackMetricsRequireExplicitEmergencyOptIn(t *testing.T) {
	r := mustNew(t, 120, 60)
	r.defaultFontFace = render.FontFace{Family: "Broken Emergency Test", Data: []byte("not a font")}
	r.fontPath = ""

	if got := r.MeasureText("Fallback", 18, ""); got != (render.TextMetrics{}) {
		t.Fatalf("default metrics should not use GSV fallback, got %+v", got)
	}
	if r.fallback {
		t.Fatal("default metrics should not mark GSV fallback as used")
	}

	r.SetEmergencyTextFallback(true)
	metrics := r.MeasureText("Fallback", 18, "")
	if metrics.W <= 0 || metrics.H <= 0 {
		t.Fatalf("emergency fallback metrics = %+v, want positive dimensions", metrics)
	}
	if !r.EmergencyTextFallbackUsed() {
		t.Fatal("expected metric fallback to mark emergency fallback diagnostics")
	}
}

func TestEmergencyGSVTextFallbackRestoresStrokeState(t *testing.T) {
	r := mustNew(t, 120, 60)
	r.defaultFontFace = render.FontFace{Family: "Broken Emergency Test", Data: []byte("not a font")}
	r.fontPath = ""
	r.SetEmergencyTextFallback(true)

	r.ctx.SetStrokeWidth(7)
	r.ctx.SetLineCap(agglib.CapSquare)
	r.ctx.SetLineJoin(agglib.JoinBevel)

	if err := r.Begin(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 120, Y: 60}}); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	r.DrawText("Fallback", geom.Pt{X: 8, Y: 32}, 18, render.Color{A: 1})
	if err := r.End(); err != nil {
		t.Fatalf("End failed: %v", err)
	}

	if got := r.ctx.painter.GetLineWidth(); math.Abs(got-7) > 1e-9 {
		t.Fatalf("line width leaked from emergency fallback: got %v, want 7", got)
	}
	if got := r.ctx.painter.GetLineCap(); got != agglib.CapSquare {
		t.Fatalf("line cap leaked from emergency fallback: got %v, want %v", got, agglib.CapSquare)
	}
	if got := r.ctx.painter.GetLineJoin(); got != agglib.JoinBevel {
		t.Fatalf("line join leaked from emergency fallback: got %v, want %v", got, agglib.JoinBevel)
	}
}
