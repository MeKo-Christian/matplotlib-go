package agg

import (
	"image"
	"image/color"
	"math"
	"reflect"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestBatchInterfaces(t *testing.T) {
	var _ render.MarkerDrawer = (*Renderer)(nil)
	var _ render.PathCollectionDrawer = (*Renderer)(nil)
	var _ render.QuadMeshDrawer = (*Renderer)(nil)
	var _ render.GouraudTriangleDrawer = (*Renderer)(nil)
}

func TestDrawMarkersBatchDrawsVisibleMarkers(t *testing.T) {
	r := mustNew(t, 40, 40)
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 40, Y: 40}}); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	marker := geom.Path{}
	marker.MoveTo(geom.Pt{X: -2, Y: -2})
	marker.LineTo(geom.Pt{X: 2, Y: -2})
	marker.LineTo(geom.Pt{X: 2, Y: 2})
	marker.LineTo(geom.Pt{X: -2, Y: 2})
	marker.Close()

	ok := r.DrawMarkers(render.MarkerBatch{
		Marker: marker,
		Items: []render.MarkerItem{
			{
				Offset: geom.Pt{X: 10, Y: 10},
				Transform: geom.Affine{
					A: 1,
					D: 1,
				},
				Paint: render.Paint{Fill: render.Color{R: 1, A: 1}},
			},
			{
				Offset: geom.Pt{X: 25, Y: 25},
				Transform: geom.Affine{
					A: 1,
					D: 1,
				},
				Paint: render.Paint{Fill: render.Color{B: 1, A: 1}},
			},
		},
	})
	if !ok {
		t.Fatal("DrawMarkers returned false")
	}
	if err := r.End(); err != nil {
		t.Fatalf("End failed: %v", err)
	}

	img := r.Image()
	if got := img.RGBAAt(10, 10); got.R < 200 {
		t.Fatalf("first marker center = %+v, want red", got)
	}
	if got := img.RGBAAt(25, 25); got.B < 200 {
		t.Fatalf("second marker center = %+v, want blue", got)
	}
}

func TestTransformMarkerPathDeviceCombinesTransformAndFlip(t *testing.T) {
	r := mustNew(t, 40, 40)
	marker := geom.Path{}
	marker.MoveTo(geom.Pt{X: -1, Y: -2})
	marker.LineTo(geom.Pt{X: 1, Y: 2})

	got := r.transformMarkerPathDevice(marker, geom.Affine{A: 2, D: 3}, geom.Pt{X: 10, Y: 12})
	wantDisplay := transformMarkerPath(marker, geom.Affine{A: 2, D: 3}, geom.Pt{X: 10, Y: 12})
	want := r.devPath(wantDisplay)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("transformMarkerPathDevice = %+v, want %+v", got, want)
	}

	pathSink := geom.Path{}
	allocs := testing.AllocsPerRun(1000, func() {
		pathSink = r.transformMarkerPathDevice(marker, geom.Affine{A: 2, D: 3}, geom.Pt{X: 10, Y: 12})
	})
	if pathSink.Validate() && allocs > 0 {
		t.Fatalf("transformMarkerPathDevice allocations = %.2f, want 0 after scratch warmup", allocs)
	}
}

func TestPreparePathForPaintKeepsFinitePathStorage(t *testing.T) {
	r := mustNew(t, 40, 40)
	path := geom.Path{}
	path.MoveTo(geom.Pt{X: 1, Y: 1})
	path.LineTo(geom.Pt{X: 5, Y: 5})
	paint := render.Paint{Fill: render.Color{A: 1}}

	got, ok := r.preparePathForPaint(path, &paint)
	if !ok {
		t.Fatal("preparePathForPaint returned !ok for finite path")
	}
	if len(got.V) == 0 || &got.V[0] != &path.V[0] {
		t.Fatalf("preparePathForPaint rebuilt finite path vertices: got=%p want=%p", &got.V[0], &path.V[0])
	}

	pathSink := geom.Path{}
	allocs := testing.AllocsPerRun(1000, func() {
		pathSink, ok = r.preparePathForPaint(path, &paint)
	})
	if !ok || len(pathSink.C) == 0 {
		t.Fatal("preparePathForPaint failed during allocation check")
	}
	if allocs > 0 {
		t.Fatalf("preparePathForPaint finite-path allocations = %.2f, want 0", allocs)
	}
}

func TestDrawQuadMeshBatchDrawsCells(t *testing.T) {
	r := mustNew(t, 40, 40)
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 40, Y: 40}}); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	ok := r.DrawQuadMesh(render.QuadMeshBatch{Cells: []render.QuadMeshCell{
		{
			Quad: [4]geom.Pt{
				{X: 5, Y: 5},
				{X: 20, Y: 5},
				{X: 20, Y: 20},
				{X: 5, Y: 20},
			},
			Face: render.Color{G: 1, A: 1},
		},
	}})
	if !ok {
		t.Fatal("DrawQuadMesh returned false")
	}
	if err := r.End(); err != nil {
		t.Fatalf("End failed: %v", err)
	}

	if got := r.Image().RGBAAt(10, 10); got.G < 200 {
		t.Fatalf("quad mesh cell center = %+v, want green", got)
	}
}

func TestDrawQuadMeshSnapsFractionalRectilinearEdges(t *testing.T) {
	r := mustNew(t, 40, 40)
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 40, Y: 40}}); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	ok := r.DrawQuadMesh(render.QuadMeshBatch{Cells: []render.QuadMeshCell{
		{
			Quad: [4]geom.Pt{
				{X: 10.2, Y: 10.2},
				{X: 25.2, Y: 10.2},
				{X: 25.2, Y: 25.2},
				{X: 10.2, Y: 25.2},
			},
			Edge:      render.Color{A: 1},
			LineWidth: 1,
			Snap:      render.SnapOn,
		},
	}})
	if !ok {
		t.Fatal("DrawQuadMesh returned false")
	}
	if err := r.End(); err != nil {
		t.Fatalf("End failed: %v", err)
	}

	if got := r.Image().RGBAAt(10, 15); got.R > 16 {
		t.Fatalf("snapped quad mesh edge pixel = %+v, want nearly black", got)
	}
}

func TestDrawGouraudTrianglesInterpolatesVertexColors(t *testing.T) {
	r := mustNew(t, 60, 60)
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 60, Y: 60}}); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	ok := r.DrawGouraudTriangles(render.GouraudTriangleBatch{Triangles: []render.GouraudTriangle{
		{
			P: [3]geom.Pt{
				{X: 5, Y: 5},
				{X: 45, Y: 5},
				{X: 5, Y: 45},
			},
			Color: [3]render.Color{
				{R: 1, A: 1},
				{G: 1, A: 1},
				{B: 1, A: 1},
			},
		},
	}})
	if !ok {
		t.Fatal("DrawGouraudTriangles returned false")
	}
	if err := r.End(); err != nil {
		t.Fatalf("End failed: %v", err)
	}

	// Display space is y-up: the triangle's red vertex at display (5,5) maps to
	// device (5,55), so sample near it in device space (display (10,10)).
	got := r.Image().RGBAAt(10, 50)
	if got == (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
		t.Fatal("triangle sample remained background white")
	}
	if got.R <= got.G || got.R <= got.B {
		t.Fatalf("triangle sample near red vertex = %+v, want red-dominant interpolation", got)
	}
}

// TestDrawGouraudTrianglesAntialiasesEdges verifies that the antialiased batch
// path produces partial-coverage fringe pixels along a slanted triangle edge,
// where the binary (Antialiased=false) path leaves a hard background/triangle
// step. Rendering the same triangle both ways and finding a pixel that is pure
// background under the binary path but partially blended under the AA path
// proves edge antialiasing is active.
func renderGouraudTriangleFlag(t *testing.T, antialiased bool) *image.RGBA {
	t.Helper()
	r := mustNew(t, 60, 60)
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 60, Y: 60}}); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	solid := render.Color{R: 0.1, G: 0.2, B: 0.8, A: 1}
	ok := r.DrawGouraudTriangles(render.GouraudTriangleBatch{
		Antialiased: antialiased,
		Triangles: []render.GouraudTriangle{
			{
				P: [3]geom.Pt{
					{X: 5, Y: 5},
					{X: 45, Y: 5},
					{X: 5, Y: 45},
				},
				Color: [3]render.Color{solid, solid, solid},
			},
		},
	})
	if !ok {
		t.Fatal("DrawGouraudTriangles returned false")
	}
	if err := r.End(); err != nil {
		t.Fatalf("End failed: %v", err)
	}
	out := image.NewRGBA(image.Rect(0, 0, 60, 60))
	copy(out.Pix, r.Image().Pix)
	out.Stride = r.Image().Stride
	return out
}

func TestDrawGouraudTrianglesAntialiasesEdges(t *testing.T) {
	aa := renderGouraudTriangleFlag(t, true)
	binary := renderGouraudTriangleFlag(t, false)

	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	fringe := 0
	for y := 0; y < 60; y++ {
		for x := 0; x < 60; x++ {
			b := binary.RGBAAt(x, y)
			a := aa.RGBAAt(x, y)
			// A fringe pixel: pure background under the binary path, but a
			// partial (non-background, non-fully-saturated) blend under AA.
			if b == white && a != white {
				// Partial coverage: the AA pixel is a blend toward the triangle
				// color, so at least one channel sits strictly between the
				// triangle color and white.
				if a.R > 0 && a.R < 255 || a.B > 100 && a.B < 255 {
					fringe++
				}
			}
		}
	}
	if fringe == 0 {
		t.Fatal("antialiased Gouraud produced no edge-fringe pixels; AA path appears inactive")
	}
}

func TestQuantize(t *testing.T) {
	cases := []struct {
		in, want float64
	}{
		{0, 0},
		{1.0, 1.0},
		{0.1234567890123, 0.123457}, // rounded to grid
		{-3.14159265, -3.141593},    // negative
		{1e-7, 0},                   // below grid, rounds to 0
		{0.0000005, 0.000001},       // half grid rounds up
		{100.123456789, 100.123457}, // large value
	}
	for _, tc := range cases {
		got := quantize(tc.in)
		if math.Abs(got-tc.want) > quantizationGrid/2 {
			t.Errorf("quantize(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestQuantizePt(t *testing.T) {
	pt := geom.Pt{X: 1.23456789, Y: 9.87654321}
	q := quantizePt(pt)
	if math.Abs(q.X-1.234568) > quantizationGrid {
		t.Errorf("X not quantized: %v", q.X)
	}
	if math.Abs(q.Y-9.876543) > quantizationGrid {
		t.Errorf("Y not quantized: %v", q.Y)
	}
}

func TestQuantizeIdempotent(t *testing.T) {
	v := 3.141592653589793
	q1 := quantize(v)
	q2 := quantize(q1)
	if q1 != q2 {
		t.Errorf("quantize not idempotent: %v != %v", q1, q2)
	}
}
