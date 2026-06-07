//go:build skia && skiacgo

package skia

import (
	"image"
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/backends"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// TestNativeSkiaVersionLinked proves the wrapper is linked against a real Skia
// library by reading its milestone string back across the C ABI.
func TestNativeSkiaVersionLinked(t *testing.T) {
	v := nativeSkiaVersion()
	if !strings.Contains(v, "milestone") || strings.TrimSpace(v) == "Skia milestone 0" {
		t.Fatalf("unexpected Skia version string %q (is libskia linked?)", v)
	}
	t.Logf("linked Skia: %s", v)
}

// TestNativeBridgeReportsNativeSurface confirms the skiacgo build selects the
// native bridge and flips the marker-batch capability from bridged to native.
func TestNativeBridgeReportsNativeSurface(t *testing.T) {
	bridge := selectSurfaceBridge(16, 16, ModeCPU)
	if !bridge.Info().NativeSurface {
		t.Fatal("native build did not select a native-surface bridge")
	}

	r, err := New(backends.Config{
		Width:      16,
		Height:     16,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		Options:    backends.SkiaConfig{SampleCount: 1, ColorType: "RGBA8888"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if r.IsCapabilityBridged("markerbatch") {
		t.Error("markerbatch should report native under the skiacgo build")
	}
	if !r.IsCapabilityBridged("pathcollectionbatch") {
		t.Error("pathcollectionbatch should still report bridged (no native batch yet)")
	}
}

// TestNativeGradientFill renders a horizontal red->blue linear gradient through
// the native surface and checks it both filled and varies left-to-right.
func TestNativeGradientFill(t *testing.T) {
	bridge := selectSurfaceBridge(32, 32, ModeCPU)
	dst := image.NewRGBA(image.Rect(0, 0, 32, 32))

	var path geom.Path
	path.MoveTo(geom.Pt{X: 2, Y: 2})
	path.LineTo(geom.Pt{X: 30, Y: 2})
	path.LineTo(geom.Pt{X: 30, Y: 30})
	path.LineTo(geom.Pt{X: 2, Y: 30})
	path.Close()

	paint := render.Paint{
		FillGradient: render.GradientFill{
			Kind:  render.LinearGradient,
			Start: geom.Pt{X: 0, Y: 16},
			End:   geom.Pt{X: 32, Y: 16},
			Stops: []render.GradientStop{
				{Offset: 0, Color: render.Color{R: 1, A: 1}},
				{Offset: 1, Color: render.Color{B: 1, A: 1}},
			},
		},
	}

	if !bridge.DrawPathFill(dst, path, paint, bridgeDrawState{}) {
		t.Fatal("DrawPathFill returned false for a plain linear gradient")
	}

	left := dst.RGBAAt(5, 16)
	right := dst.RGBAAt(27, 16)
	if left.A == 0 || right.A == 0 {
		t.Fatalf("gradient did not fill: left=%v right=%v", left, right)
	}
	if !(left.R > left.B) {
		t.Errorf("left edge should be reddish, got %v", left)
	}
	if !(right.B > right.R) {
		t.Errorf("right edge should be bluish, got %v", right)
	}
}

// TestNativeMarkers renders a small square marker at two offsets through the
// native batch path and confirms it produced opaque green ink.
func TestNativeMarkers(t *testing.T) {
	bridge := selectSurfaceBridge(32, 32, ModeCPU).(nativeBatchBridge)
	dst := image.NewRGBA(image.Rect(0, 0, 32, 32))

	var marker geom.Path
	marker.MoveTo(geom.Pt{X: -4, Y: -4})
	marker.LineTo(geom.Pt{X: 4, Y: -4})
	marker.LineTo(geom.Pt{X: 4, Y: 4})
	marker.LineTo(geom.Pt{X: -4, Y: 4})
	marker.Close()

	green := render.Color{R: 0, G: 0.6, B: 0, A: 1}
	batch := render.MarkerBatch{
		Marker: marker,
		Items: []render.MarkerItem{
			{Offset: geom.Pt{X: 8, Y: 16}, Transform: geom.Identity(), Paint: render.Paint{Fill: green}, Antialiased: true},
			{Offset: geom.Pt{X: 24, Y: 16}, Transform: geom.Identity(), Paint: render.Paint{Fill: green}, Antialiased: true},
		},
	}

	if !bridge.drawMarkersNative(dst, batch, bridgeDrawState{}, 32) {
		t.Fatal("drawMarkersNative returned false for solid-fill markers")
	}

	opaque := 0
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			if dst.RGBAAt(x, y).A > 0 {
				opaque++
			}
		}
	}
	if opaque == 0 {
		t.Fatal("native markers produced no ink")
	}
	// Each marker center maps (offset.X, height-offset.Y) -> (8,16) and (24,16).
	if c := dst.RGBAAt(8, 16); c.A == 0 || c.G <= c.R {
		t.Errorf("expected green ink near first marker center, got %v", c)
	}
}

// TestNativeGouraud renders one interpolated-color triangle via SkVertices and
// confirms its interior is shaded.
func TestNativeGouraud(t *testing.T) {
	bridge := selectSurfaceBridge(32, 32, ModeCPU).(nativeBatchBridge)
	dst := image.NewRGBA(image.Rect(0, 0, 32, 32))

	batch := render.GouraudTriangleBatch{
		Triangles: []render.GouraudTriangle{{
			P: [3]geom.Pt{{X: 6, Y: 6}, {X: 26, Y: 6}, {X: 16, Y: 26}},
			Color: [3]render.Color{
				{R: 1, A: 1},
				{G: 1, A: 1},
				{B: 1, A: 1},
			},
		}},
	}

	if !bridge.drawGouraudNative(dst, batch, bridgeDrawState{}) {
		t.Fatal("drawGouraudNative returned false")
	}
	if c := dst.RGBAAt(16, 12); c.A == 0 {
		t.Fatalf("triangle interior not shaded at (16,12): %v", c)
	}
}
