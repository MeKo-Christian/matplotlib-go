//go:build skia && skiagpu

package skia

import (
	"image"
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/backends"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// TestGPUScaffoldRendersThroughCPUReadback verifies the skiagpu build tag wires a
// GPU-mode renderer that reports the GPU render mode while still rasterizing
// deterministically through the CPU readback bridge (no native GPU surface yet).
func TestGPUScaffoldRendersThroughCPUReadback(t *testing.T) {
	r, err := New(backends.Config{
		Width:      64,
		Height:     64,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		Options:    backends.SkiaConfig{UseGPU: true, SampleCount: 1, ColorType: "RGBA8888"},
	})
	if err != nil {
		t.Fatalf("New(UseGPU:true) under skiagpu = %v, want success", err)
	}
	if !r.GPU() {
		t.Fatal("GPU() = false, want true under skiagpu build tag")
	}

	info := r.BridgeInfo()
	if info.Mode != ModeGPU {
		t.Fatalf("BridgeInfo().Mode = %q, want %q", info.Mode, ModeGPU)
	}
	if info.NativeSurface {
		t.Fatal("BridgeInfo().NativeSurface = true, want false (CPU readback scaffold)")
	}

	if got := BackendStrategy().GPUStatus; got != StatusPlanned {
		t.Fatalf("BackendStrategy().GPUStatus = %q, want %q under skiagpu", got, StatusPlanned)
	}

	// Deterministic CPU readback: a filled path must produce non-background pixels.
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 64, Y: 64}}
	if err := r.Begin(viewport); err != nil {
		t.Fatalf("Begin = %v", err)
	}
	path := geom.Path{
		V: []geom.Pt{{X: 10, Y: 10}, {X: 50, Y: 10}, {X: 50, Y: 50}, {X: 10, Y: 50}},
		C: []geom.Cmd{geom.MoveTo, geom.LineTo, geom.LineTo, geom.LineTo, geom.ClosePath},
	}
	r.Path(path, &render.Paint{Fill: render.Color{R: 0, G: 0, B: 0, A: 1}})
	if err := r.End(); err != nil {
		t.Fatalf("End = %v", err)
	}

	img := r.GetImage()
	if img == nil {
		t.Fatal("GetImage() = nil, want a CPU readback buffer")
	}
	if !hasNonBackgroundPixel(img) {
		t.Fatal("GPU scaffold produced an empty image; CPU readback did not render the path")
	}
}

func TestGPUScaffoldBackendComparisonReportLabelsGPUMode(t *testing.T) {
	report := backends.BackendComparisonReport(backends.Config{
		Width:      32,
		Height:     24,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		Options:    backends.SkiaConfig{UseGPU: true, SampleCount: 1, ColorType: "RGBA8888"},
	})
	if !strings.Contains(report, "skia/gpu") {
		t.Fatalf("BackendComparisonReport did not label Skia GPU mode:\n%s", report)
	}
}

func hasNonBackgroundPixel(img *image.RGBA) bool {
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			c := img.RGBAAt(x, y)
			if c.R != 0xff || c.G != 0xff || c.B != 0xff {
				return true
			}
		}
	}
	return false
}
