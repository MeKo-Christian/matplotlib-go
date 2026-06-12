//go:build skia && skiacgo

package skia

import (
	"image"
	"image/color"
	"math"
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
// native bridge and flips linked native batch capabilities from bridged to
// native.
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
	if r.IsCapabilityBridged("pathcollectionbatch") {
		t.Error("pathcollectionbatch should report native under the skiacgo build")
	}
	if r.IsCapabilityBridged("quadmeshbatch") {
		t.Error("quadmeshbatch should report native under the skiacgo build")
	}
	if r.IsCapabilityBridged("imagetransform") {
		t.Error("imagetransform should report native under the skiacgo build")
	}
	if r.IsCapabilityBridged("nativehatcher") {
		t.Error("nativehatcher should report native under the skiacgo build")
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

// TestNativePathCollection renders multiple solid path-collection items through
// one native surface and confirms both fill and stroke survive compositing.
func TestNativePathCollection(t *testing.T) {
	bridge := selectSurfaceBridge(32, 32, ModeCPU).(nativeBatchBridge)
	dst := image.NewRGBA(image.Rect(0, 0, 32, 32))

	var left geom.Path
	left.MoveTo(geom.Pt{X: 4, Y: 4})
	left.LineTo(geom.Pt{X: 14, Y: 4})
	left.LineTo(geom.Pt{X: 14, Y: 18})
	left.LineTo(geom.Pt{X: 4, Y: 18})
	left.Close()

	var right geom.Path
	right.MoveTo(geom.Pt{X: 18, Y: 14})
	right.LineTo(geom.Pt{X: 28, Y: 14})
	right.LineTo(geom.Pt{X: 28, Y: 28})
	right.LineTo(geom.Pt{X: 18, Y: 28})
	right.Close()

	batch := render.PathCollectionBatch{Items: []render.PathCollectionItem{
		{
			Path:        left,
			Paint:       render.Paint{Fill: render.Color{R: 1, A: 1}},
			Antialiased: true,
		},
		{
			Path: right,
			Paint: render.Paint{
				Fill:      render.Color{G: 1, A: 1},
				Stroke:    render.Color{B: 1, A: 1},
				LineWidth: 2,
			},
			Antialiased: true,
		},
	}}

	if !bridge.drawPathCollectionNative(dst, batch, bridgeDrawState{}) {
		t.Fatal("drawPathCollectionNative returned false for solid path collection")
	}
	if c := dst.RGBAAt(8, 20); c.A == 0 || c.R <= c.G || c.R <= c.B {
		t.Fatalf("expected red fill in y-flipped first path, got %v", c)
	}
	if c := dst.RGBAAt(8, 8); c.A != 0 {
		t.Fatalf("expected original y-up first path location to stay empty, got %v", c)
	}
	if c := dst.RGBAAt(22, 12); c.A == 0 || c.G <= c.R || c.G <= c.B {
		t.Fatalf("expected green fill in y-flipped second path, got %v", c)
	}
	if c := dst.RGBAAt(18, 12); c.A == 0 || c.B <= c.G {
		t.Fatalf("expected blue stroke on y-flipped second path edge, got %v", c)
	}
}

// TestNativeQuadMesh renders face-only quadrilateral cells as native SkVertices
// triangles and confirms display-space coordinates are flipped into Skia device
// space.
func TestNativeQuadMesh(t *testing.T) {
	bridge := selectSurfaceBridge(32, 32, ModeCPU).(nativeBatchBridge)
	dst := image.NewRGBA(image.Rect(0, 0, 32, 32))

	batch := render.QuadMeshBatch{Cells: []render.QuadMeshCell{
		{
			Quad: [4]geom.Pt{
				{X: 4, Y: 4},
				{X: 14, Y: 4},
				{X: 14, Y: 18},
				{X: 4, Y: 18},
			},
			Face:        render.Color{R: 1, A: 1},
			Antialiased: true,
		},
		{
			Quad: [4]geom.Pt{
				{X: 18, Y: 14},
				{X: 28, Y: 14},
				{X: 28, Y: 28},
				{X: 18, Y: 28},
			},
			Face:        render.Color{G: 1, A: 1},
			Antialiased: true,
		},
	}}

	if !bridge.drawQuadMeshNative(dst, batch, bridgeDrawState{}) {
		t.Fatal("drawQuadMeshNative returned false for face-only quad mesh")
	}
	if c := dst.RGBAAt(8, 20); c.A == 0 || c.R <= c.G || c.R <= c.B {
		t.Fatalf("expected red face in y-flipped first cell, got %v", c)
	}
	if c := dst.RGBAAt(8, 8); c.A != 0 {
		t.Fatalf("expected original y-up first cell location to stay empty, got %v", c)
	}
	if c := dst.RGBAAt(22, 12); c.A == 0 || c.G <= c.R || c.G <= c.B {
		t.Fatalf("expected green face in y-flipped second cell, got %v", c)
	}
}

// TestNativeTransformedImage renders an RGBA image through the native bridge,
// confirming image-space row orientation, display->device y flipping, alpha, and
// clip-path masking match the renderer contract.
func TestNativeTransformedImage(t *testing.T) {
	bridge, ok := selectSurfaceBridge(20, 20, ModeCPU).(interface {
		drawImageTransformedNative(*image.RGBA, render.Image, geom.Affine, bridgeDrawState) bool
	})
	if !ok {
		t.Fatal("native bridge does not implement transformed image drawing")
	}
	dst := image.NewRGBA(image.Rect(0, 0, 20, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			dst.SetRGBA(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}

	src := image.NewRGBA(image.Rect(0, 0, 2, 2))
	src.SetRGBA(0, 0, color.RGBA{R: 255, A: 255})
	src.SetRGBA(1, 0, color.RGBA{G: 255, A: 255})
	src.SetRGBA(0, 1, color.RGBA{B: 255, A: 255})
	src.SetRGBA(1, 1, color.RGBA{R: 255, G: 255, A: 255})
	data := render.NewImageData(src)
	data.SetAlpha(0.5)

	ok = bridge.drawImageTransformedNative(dst, data, geom.Affine{
		A: 10,
		D: -10,
		F: 20,
	}, bridgeDrawState{})
	if !ok {
		t.Fatal("drawImageTransformedNative returned false for RGBA image")
	}

	topLeft := dst.RGBAAt(2, 2)
	if topLeft.R != 255 || math.Abs(float64(topLeft.G)-128) > 2 || math.Abs(float64(topLeft.B)-128) > 2 || topLeft.A != 255 {
		t.Fatalf("top-left source row not drawn as half-alpha red over white: %+v", topLeft)
	}
	topRight := dst.RGBAAt(12, 2)
	if topRight.G != 255 || math.Abs(float64(topRight.R)-128) > 2 || math.Abs(float64(topRight.B)-128) > 2 || topRight.A != 255 {
		t.Fatalf("top-right source row not drawn as half-alpha green over white: %+v", topRight)
	}

	clippedDst := image.NewRGBA(image.Rect(0, 0, 20, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			clippedDst.SetRGBA(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}
	var clip geom.Path
	clip.MoveTo(geom.Pt{X: 0, Y: 0})
	clip.LineTo(geom.Pt{X: 20, Y: 0})
	clip.LineTo(geom.Pt{X: 0, Y: 20})
	clip.Close()
	ok = bridge.drawImageTransformedNative(clippedDst, data, geom.Affine{
		A: 10,
		D: -10,
		F: 20,
	}, bridgeDrawState{clipPaths: []geom.Path{clip}})
	if !ok {
		t.Fatal("drawImageTransformedNative returned false for clipped RGBA image")
	}
	outside := clippedDst.RGBAAt(15, 4)
	if outside != (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
		t.Fatalf("transformed image escaped y-flipped clip path: %+v", outside)
	}
	inside := clippedDst.RGBAAt(4, 15)
	if inside.A != 255 || inside == (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
		t.Fatalf("clipped transformed image did not draw inside clip: %+v", inside)
	}
}

// TestNativeHatchShader renders a hatch fill through the native bridge. The
// direct bridge assertion keeps this test red while hatches are only expanded
// through render.DrawHatchFallback.
func TestNativeHatchShader(t *testing.T) {
	bridge, ok := selectSurfaceBridge(32, 32, ModeCPU).(interface {
		drawHatchPathNative(*image.RGBA, geom.Path, render.Paint, bridgeDrawState) bool
	})
	if !ok {
		t.Fatal("native bridge does not implement hatch shader drawing")
	}
	dst := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			dst.SetRGBA(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}

	paint := render.Paint{
		Hatch:          "/",
		HatchColor:     render.Color{R: 1, A: 1},
		HatchLineWidth: 2,
		HatchSpacing:   8,
	}
	if !bridge.drawHatchPathNative(dst, rectPath(4, 4, 28, 28), paint, bridgeDrawState{}) {
		t.Fatal("drawHatchPathNative returned false for slash hatch")
	}

	redPixels := 0
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			c := dst.RGBAAt(x, y)
			if c.R > 180 && c.G < 120 && c.B < 120 {
				redPixels++
			}
		}
	}
	if redPixels == 0 {
		t.Fatal("native hatch shader produced no red hatch pixels")
	}
	if got := dst.RGBAAt(1, 1); got != (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
		t.Fatalf("native hatch shader drew outside the path clip: %+v", got)
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
