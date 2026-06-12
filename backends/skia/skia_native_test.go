//go:build skia

package skia

import (
	"image/color"
	"testing"

	"github.com/cwbudde/matplotlib-go/backends"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestSkiaTaggedRendererImplementsNativeBatchInterfaces(t *testing.T) {
	r := mustNewShaderRenderer(t, 20, 20)

	var _ render.MarkerDrawer = r
	var _ render.PathCollectionDrawer = r
	var _ render.QuadMeshDrawer = r
	var _ render.GouraudTriangleDrawer = r
	var _ render.NativeHatcher = r

	for _, cap := range []backends.Capability{
		backends.ImageTransform,
		backends.GouraudTriangleBatch,
	} {
		if status := backends.RendererCapabilityStatus(backends.Skia, r, cap); status != backends.CapabilityNative {
			t.Fatalf("RendererCapabilityStatus(skia, %s) = %s, want native", cap, status)
		}
	}
	for _, cap := range []backends.Capability{
		backends.QuadMeshBatch,
		backends.NativeHatcher,
	} {
		if status := backends.RendererCapabilityStatus(backends.Skia, r, cap); status != backends.CapabilityBridged {
			t.Fatalf("RendererCapabilityStatus(skia, %s) = %s, want bridged", cap, status)
		}
	}

	// MarkerBatch and PathCollectionBatch are native when a real Skia library is
	// linked (skiacgo build), otherwise satisfied through the CPU bridge.
	wantNativeBatch := backends.CapabilityBridged
	if r.BridgeInfo().NativeSurface {
		wantNativeBatch = backends.CapabilityNative
	}
	for _, cap := range []backends.Capability{
		backends.MarkerBatch,
		backends.PathCollectionBatch,
	} {
		if status := backends.RendererCapabilityStatus(backends.Skia, r, cap); status != wantNativeBatch {
			t.Fatalf("RendererCapabilityStatus(skia, %s) = %s, want %s", cap, status, wantNativeBatch)
		}
	}
}

func TestSkiaDrawMarkersBatchDrawsVisibleMarkers(t *testing.T) {
	r := mustNewShaderRenderer(t, 30, 15)
	mustBeginShaderRenderer(t, r, 30, 15)

	marker := rectPath(-2, -2, 2, 2)
	if ok := r.DrawMarkers(render.MarkerBatch{
		Marker: marker,
		Items: []render.MarkerItem{
			{Offset: geom.Pt{X: 8, Y: 7}, Paint: render.Paint{Fill: render.Color{R: 1, A: 1}}},
			{Offset: geom.Pt{X: 22, Y: 7}, Paint: render.Paint{Fill: render.Color{B: 1, A: 1}}},
		},
	}); !ok {
		t.Fatal("DrawMarkers returned false")
	}

	mustEndShaderRenderer(t, r)
	assertPixelMostly(t, r, 8, 7, color.RGBA{R: 255, A: 255})
	assertPixelMostly(t, r, 22, 7, color.RGBA{B: 255, A: 255})
}

func TestSkiaDrawPathCollectionAndQuadMeshDrawVisibleCells(t *testing.T) {
	r := mustNewShaderRenderer(t, 40, 20)
	mustBeginShaderRenderer(t, r, 40, 20)

	if ok := r.DrawPathCollection(render.PathCollectionBatch{Items: []render.PathCollectionItem{
		{
			Path:  rectPath(1, 1, 12, 12),
			Paint: render.Paint{Fill: render.Color{G: 1, A: 1}},
		},
	}}); !ok {
		t.Fatal("DrawPathCollection returned false")
	}
	if ok := r.DrawQuadMesh(render.QuadMeshBatch{Cells: []render.QuadMeshCell{
		{
			Quad: [4]geom.Pt{
				{X: 20, Y: 2},
				{X: 36, Y: 2},
				{X: 36, Y: 18},
				{X: 20, Y: 18},
			},
			Face: render.Color{R: 1, B: 1, A: 1},
		},
	}}); !ok {
		t.Fatal("DrawQuadMesh returned false")
	}

	mustEndShaderRenderer(t, r)
	assertPixelMostly(t, r, 6, 6, color.RGBA{G: 255, A: 255})
	assertPixelMostly(t, r, 28, 10, color.RGBA{R: 255, B: 255, A: 255})
}

func TestSkiaDrawGouraudTrianglesInterpolatesVertexColors(t *testing.T) {
	r := mustNewShaderRenderer(t, 24, 24)
	mustBeginShaderRenderer(t, r, 24, 24)

	if ok := r.DrawGouraudTriangles(render.GouraudTriangleBatch{Triangles: []render.GouraudTriangle{
		{
			P: [3]geom.Pt{
				{X: 2, Y: 2},
				{X: 22, Y: 2},
				{X: 12, Y: 22},
			},
			Color: [3]render.Color{
				{R: 1, A: 1},
				{G: 1, A: 1},
				{B: 1, A: 1},
			},
		},
	}}); !ok {
		t.Fatal("DrawGouraudTriangles returned false")
	}

	mustEndShaderRenderer(t, r)
	rv, gv, bv, _ := skiaPixelAt(t, r, 12, 8)
	if rv < 40 || gv < 40 || bv < 40 {
		t.Fatalf("expected interpolated interior color, got RGB=(%d,%d,%d)", rv, gv, bv)
	}
}

func TestSkiaNativeHatchDrawsWithinPathClip(t *testing.T) {
	r := mustNewShaderRenderer(t, 24, 24)
	if !r.SupportsNativeHatch() {
		t.Fatal("SupportsNativeHatch returned false")
	}
	mustBeginShaderRenderer(t, r, 24, 24)

	r.Path(rectPath(2, 2, 22, 22), &render.Paint{
		Hatch:          "|",
		HatchColor:     render.Color{G: 1, A: 1},
		HatchLineWidth: 2,
		HatchSpacing:   6,
	})

	mustEndShaderRenderer(t, r)
	if got := countMostlyGreenPixels(r); got == 0 {
		t.Fatal("native hatch did not draw visible green pixels")
	}
	if got := r.GetImage().RGBAAt(0, 0); got != (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
		t.Fatalf("hatch drew outside path clip at (0,0): %#v", got)
	}
}

func assertPixelMostly(t *testing.T, r *Renderer, x, y int, want color.RGBA) {
	t.Helper()
	got := r.GetImage().RGBAAt(x, y)
	if want.R > 0 && got.R < 200 {
		t.Fatalf("pixel (%d,%d) red = %d, want mostly red", x, y, got.R)
	}
	if want.G > 0 && got.G < 200 {
		t.Fatalf("pixel (%d,%d) green = %d, want mostly green", x, y, got.G)
	}
	if want.B > 0 && got.B < 200 {
		t.Fatalf("pixel (%d,%d) blue = %d, want mostly blue", x, y, got.B)
	}
	if want.A > 0 && got.A < 200 {
		t.Fatalf("pixel (%d,%d) alpha = %d, want mostly opaque", x, y, got.A)
	}
}

func countMostlyGreenPixels(r *Renderer) int {
	img := r.GetImage()
	if img == nil {
		return 0
	}
	count := 0
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			c := img.RGBAAt(x, y)
			if c.G > 180 && c.R < 120 && c.B < 120 {
				count++
			}
		}
	}
	return count
}
