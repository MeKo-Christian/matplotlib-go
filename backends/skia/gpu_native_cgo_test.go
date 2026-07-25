//go:build skia && skiagpu && skiacgo

package skia

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/backends"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestNativeGPUSurfaceGradientAndReadback(t *testing.T) {
	r, err := New(backends.Config{
		Width:      64,
		Height:     48,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		Options:    backends.SkiaConfig{UseGPU: true, SampleCount: 4, ColorType: "RGBA8888"},
	})
	if err != nil {
		t.Fatalf("New GPU renderer: %v", err)
	}
	if !r.GPU() {
		t.Skipf("native GPU context unavailable: %s", r.BridgeInfo().Description)
	}
	if got := r.BridgeInfo(); !got.NativeSurface || !got.Accelerated || got.Mode != ModeGPU {
		t.Fatalf("BridgeInfo() = %+v, want accelerated native GPU mode", got)
	}
	if got := r.SampleCount(); got != 4 {
		t.Fatalf("SampleCount() = %d, want 4", got)
	}

	viewport := geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 64, Y: 48}}
	if err := r.Begin(viewport); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	path := geom.Path{
		V: []geom.Pt{{X: 4, Y: 4}, {X: 60, Y: 4}, {X: 60, Y: 44}, {X: 4, Y: 44}},
		C: []geom.Cmd{geom.MoveTo, geom.LineTo, geom.LineTo, geom.LineTo, geom.ClosePath},
	}
	r.Path(path, &render.Paint{FillGradient: render.GradientFill{
		Kind:  render.LinearGradient,
		Start: geom.Pt{X: 4, Y: 24},
		End:   geom.Pt{X: 60, Y: 24},
		Stops: []render.GradientStop{
			{Offset: 0, Color: render.Color{R: 1, A: 1}},
			{Offset: 1, Color: render.Color{B: 1, A: 1}},
		},
	}})
	r.FlushGPU()
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	img := r.GetImage()
	if img == nil {
		t.Fatal("GetImage() = nil after GPU readback")
	}
	left := img.RGBAAt(12, 24)
	right := img.RGBAAt(52, 24)
	if left == right || left.R <= left.B || right.B <= right.R {
		t.Fatalf("GPU gradient readback left=%v right=%v, want red-to-blue variation", left, right)
	}
}
