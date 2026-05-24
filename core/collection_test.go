package core

import (
	"math"
	"testing"

	matcolor "github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestPathCollectionDrawAndBounds(t *testing.T) {
	pc := &PathCollection{
		Collection: Collection{Label: "markers", Alpha: 0.75, z: 3},
		Path:       polygonPath([]geom.Pt{{X: 0, Y: -0.5}, {X: 0.5, Y: 0.5}, {X: -0.5, Y: 0.5}}, true),
		Offsets:    []geom.Pt{{X: 1, Y: 2}, {X: 4, Y: 5}},
		Sizes:      []float64{2, 3},
		FaceColor:  render.Color{R: 0.2, G: 0.4, B: 0.8, A: 1},
		EdgeColor:  render.Color{R: 0.1, G: 0.1, B: 0.1, A: 1},
		EdgeWidth:  1.5,
	}

	r := &recordingRenderer{}
	pc.Draw(r, createTestDrawContext())

	if len(r.pathCalls) != 2 {
		t.Fatalf("expected 2 path calls, got %d", len(r.pathCalls))
	}
	if r.pathCalls[0].paint.Fill.A <= 0 || r.pathCalls[0].paint.Stroke.A <= 0 {
		t.Fatalf("expected fill and stroke paint, got %+v", r.pathCalls[0].paint)
	}

	bounds := pc.Bounds(nil)
	if bounds.Min.X >= 1 || bounds.Min.Y >= 2 {
		t.Fatalf("expected bounds expansion around first offset, got %+v", bounds)
	}
	if bounds.Max.X <= 4 || bounds.Max.Y <= 5 {
		t.Fatalf("expected bounds expansion around second offset, got %+v", bounds)
	}
}

func TestPathCollectionArtistAlphaMultipliesCollectionAlpha(t *testing.T) {
	pc := &PathCollection{
		Collection: Collection{Alpha: 0.5},
		Path:       markerRectanglePath(-0.5, -0.5, 0.5, 0.5),
		Offsets:    []geom.Pt{{X: 1, Y: 2}},
		Size:       1,
		FaceColor:  render.Color{R: 1, A: 0.8},
		EdgeColor:  render.Color{B: 1, A: 0.6},
		EdgeWidth:  1,
	}
	pc.SetAlpha(0.5)

	r := &recordingRenderer{}
	pc.Draw(r, createTestDrawContext())

	if len(r.pathCalls) != 1 {
		t.Fatalf("path calls = %d, want 1", len(r.pathCalls))
	}
	if got, want := r.pathCalls[0].paint.Fill.A, 0.2; got != want {
		t.Fatalf("fill alpha = %v, want %v", got, want)
	}
	if got, want := r.pathCalls[0].paint.Stroke.A, 0.15; got != want {
		t.Fatalf("stroke alpha = %v, want %v", got, want)
	}

	pc.SetAlpha(0)
	r.pathCalls = nil
	pc.Draw(r, createTestDrawContext())
	if len(r.pathCalls) != 0 {
		t.Fatalf("transparent collection drew %d paths, want 0", len(r.pathCalls))
	}
}

type batchRecordingRenderer struct {
	recordingRenderer
	markerBatches         []render.MarkerBatch
	pathCollectionBatches []render.PathCollectionBatch
	quadMeshBatches       []render.QuadMeshBatch
	gouraudBatches        []render.GouraudTriangleBatch
	returnNative          bool
}

type nativeHatchBatchRecordingRenderer struct {
	batchRecordingRenderer
}

func (r *nativeHatchBatchRecordingRenderer) SupportsNativeHatch() bool {
	return true
}

func (r *batchRecordingRenderer) DrawMarkers(batch render.MarkerBatch) bool {
	r.markerBatches = append(r.markerBatches, batch)
	return r.returnNative
}

func (r *batchRecordingRenderer) DrawPathCollection(batch render.PathCollectionBatch) bool {
	r.pathCollectionBatches = append(r.pathCollectionBatches, batch)
	return r.returnNative
}

func (r *batchRecordingRenderer) DrawQuadMesh(batch render.QuadMeshBatch) bool {
	r.quadMeshBatches = append(r.quadMeshBatches, batch)
	return r.returnNative
}

func (r *batchRecordingRenderer) DrawGouraudTriangles(batch render.GouraudTriangleBatch) bool {
	r.gouraudBatches = append(r.gouraudBatches, batch)
	return r.returnNative
}

func TestPathCollectionUsesMarkerBatchWhenAvailable(t *testing.T) {
	pc := &PathCollection{
		Collection:    Collection{Alpha: 0.8},
		Path:          polygonPath([]geom.Pt{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 1}}, true),
		Offsets:       []geom.Pt{{X: 1, Y: 2}, {X: 4, Y: 5}},
		Size:          2,
		PathInDisplay: true,
		FaceColor:     render.Color{R: 1, A: 1},
		EdgeColor:     render.Color{A: 1},
		EdgeWidth:     1,
	}

	r := &batchRecordingRenderer{returnNative: true}
	pc.Draw(r, createTestDrawContext())

	if len(r.markerBatches) != 1 {
		t.Fatalf("marker batches = %d, want 1", len(r.markerBatches))
	}
	if len(r.pathCalls) != 0 || len(r.pathCollectionBatches) != 0 {
		t.Fatalf("expected marker native path only, pathCalls=%d pathCollectionBatches=%d", len(r.pathCalls), len(r.pathCollectionBatches))
	}
	items := r.markerBatches[0].Items
	if len(items) != 2 {
		t.Fatalf("marker items = %d, want 2", len(items))
	}
	if items[0].Offset != (geom.Pt{X: 60, Y: 430}) {
		t.Fatalf("first marker display offset = %+v", items[0].Offset)
	}
	if items[1].Transform.A != 2 || items[1].Transform.D != 2 {
		t.Fatalf("second marker transform = %+v", items[1].Transform)
	}
}

func TestPathCollectionUsesPathCollectionBatchForVaryingPerItemStyle(t *testing.T) {
	pc := &PathCollection{
		Collection:    Collection{Alpha: 0.5},
		Path:          polygonPath([]geom.Pt{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 1}}, true),
		Offsets:       []geom.Pt{{X: 1, Y: 2}, {X: 4, Y: 5}},
		Sizes:         []float64{2, 3},
		PathInDisplay: true,
		FaceColors: []render.Color{
			{R: 1, G: 0.1, B: 0.2, A: 0.8},
			{R: 0.2, G: 1, B: 0.3, A: 0.6},
		},
		EdgeColors: []render.Color{
			{R: 0.3, G: 0.2, B: 0.1, A: 1},
			{R: 0.1, G: 0.2, B: 0.3, A: 0.7},
		},
		EdgeWidths: []float64{1.25, 2.5},
	}

	r := &batchRecordingRenderer{returnNative: true}
	pc.Draw(r, createTestDrawContext())

	if len(r.markerBatches) != 0 {
		t.Fatalf("marker batches = %d, want none for varying collection state", len(r.markerBatches))
	}
	if len(r.pathCollectionBatches) != 1 {
		t.Fatalf("path collection batches = %d, want 1", len(r.pathCollectionBatches))
	}
	if len(r.pathCollectionBatches[0].Items) != 2 {
		t.Fatalf("path collection items = %d, want 2", len(r.pathCollectionBatches[0].Items))
	}
	if len(r.pathCalls) != 0 {
		t.Fatalf("fallback path calls = %d, want 0", len(r.pathCalls))
	}

	items := r.pathCollectionBatches[0].Items
	if got, want := items[0].Paint.Fill.A, 0.4; math.Abs(got-want) > 1e-12 {
		t.Fatalf("first fill alpha = %v, want %v", got, want)
	}
	if got, want := items[1].Paint.Stroke.A, 0.35; math.Abs(got-want) > 1e-12 {
		t.Fatalf("second stroke alpha = %v, want %v", got, want)
	}
	if got, want := items[0].Paint.LineWidth, 1.25; got != want {
		t.Fatalf("first linewidth = %v, want %v", got, want)
	}
	if got, want := items[1].Paint.LineWidth, 2.5; got != want {
		t.Fatalf("second linewidth = %v, want %v", got, want)
	}
}

func TestPathCollectionSkipsEmptyAndInvisibleCollections(t *testing.T) {
	for _, tc := range []struct {
		name string
		pc   *PathCollection
	}{
		{
			name: "empty",
			pc:   &PathCollection{},
		},
		{
			name: "all invisible",
			pc: &PathCollection{
				Path:          polygonPath([]geom.Pt{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 1}}, true),
				Offsets:       []geom.Pt{{X: 1, Y: 2}, {X: 4, Y: 5}},
				PathInDisplay: true,
				FaceColors: []render.Color{
					{R: 1, A: 0},
					{G: 1, A: 0},
				},
				EdgeColors: []render.Color{
					{A: 0},
					{A: 0},
				},
				EdgeWidth: 2,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := &batchRecordingRenderer{returnNative: true}
			tc.pc.Draw(r, createTestDrawContext())

			if len(r.markerBatches) != 0 || len(r.pathCollectionBatches) != 0 || len(r.pathCalls) != 0 {
				t.Fatalf("expected no drawing, marker=%d collection=%d paths=%d", len(r.markerBatches), len(r.pathCollectionBatches), len(r.pathCalls))
			}
		})
	}
}

func TestPathCollectionLineOnlyUsesFaceColorAsStrokeWhenEdgeUnset(t *testing.T) {
	pc := &PathCollection{
		Collection:    Collection{Alpha: 0.5},
		Path:          polygonPath([]geom.Pt{{X: -0.5, Y: 0}, {X: 0.5, Y: 0}}, false),
		Offsets:       []geom.Pt{{X: 1, Y: 2}},
		Size:          2,
		PathInDisplay: true,
		FaceColor:     render.Color{R: 0.3, G: 0.4, B: 0.5, A: 0.8},
		EdgeWidth:     1.5,
		LineOnly:      true,
	}

	r := &batchRecordingRenderer{returnNative: true}
	pc.Draw(r, createTestDrawContext())

	if len(r.markerBatches) != 1 {
		t.Fatalf("marker batches = %d, want 1", len(r.markerBatches))
	}
	items := r.markerBatches[0].Items
	if len(items) != 1 {
		t.Fatalf("marker items = %d, want 1", len(items))
	}
	paint := items[0].Paint
	if paint.Fill.A != 0 {
		t.Fatalf("line-only fill alpha = %v, want 0", paint.Fill.A)
	}
	if got, want := paint.Stroke, (render.Color{R: 0.3, G: 0.4, B: 0.5, A: 0.4}); got != want {
		t.Fatalf("line-only stroke = %+v, want %+v", got, want)
	}
	if got, want := paint.LineWidth, 1.5; got != want {
		t.Fatalf("line-only linewidth = %v, want %v", got, want)
	}
}

func TestPathCollectionEdgeColorsFaceStyleUsesFaceColorsForStroke(t *testing.T) {
	faces := []render.Color{
		{R: 0.8, G: 0.1, B: 0.2, A: 0.6},
		{R: 0.1, G: 0.7, B: 0.3, A: 0.8},
	}
	pc := &PathCollection{
		Collection:    Collection{Alpha: 0.5},
		Path:          polygonPath([]geom.Pt{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 1}}, true),
		Offsets:       []geom.Pt{{X: 1, Y: 2}, {X: 4, Y: 5}},
		PathInDisplay: true,
		FaceColors:    faces,
		EdgeColors:    append([]render.Color(nil), faces...),
		EdgeWidth:     1,
	}

	r := &batchRecordingRenderer{returnNative: true}
	pc.Draw(r, createTestDrawContext())

	if len(r.pathCollectionBatches) != 1 {
		t.Fatalf("path collection batches = %d, want 1", len(r.pathCollectionBatches))
	}
	items := r.pathCollectionBatches[0].Items
	if len(items) != 2 {
		t.Fatalf("path collection items = %d, want 2", len(items))
	}
	for i, item := range items {
		if item.Paint.Stroke != item.Paint.Fill {
			t.Fatalf("item %d stroke = %+v, want face-colored edge %+v", i, item.Paint.Stroke, item.Paint.Fill)
		}
	}
}

func TestPathCollectionSetArrayRefreshesMappedFacesAndFaceEdges(t *testing.T) {
	cmapName := "path-collection-scalar-array"
	low := render.Color{R: 1, A: 1}
	high := render.Color{B: 1, A: 1}
	matcolor.RegisterColormap(cmapName, matcolor.NewColormap(cmapName, []matcolor.ColorStop{
		{Pos: 0, Color: low},
		{Pos: 1, Color: high},
	}))

	pc := &PathCollection{
		Collection: Collection{
			Colormap: cmapName,
		},
		Path:          polygonPath([]geom.Pt{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 1}}, true),
		Offsets:       []geom.Pt{{X: 1, Y: 2}, {X: 4, Y: 5}},
		PathInDisplay: true,
		FaceColor:     render.Color{G: 1, A: 1},
		EdgeColor:     render.Color{A: 1},
		EdgeWidth:     1,
	}
	pc.SetEdgeColorFace()
	if err := pc.SetArray([]float64{0, 10}); err != nil {
		t.Fatalf("SetArray: %v", err)
	}

	if got := pc.GetArray(); len(got) != 2 || got[0] != 0 || got[1] != 10 {
		t.Fatalf("GetArray = %v, want copied scalar values", got)
	}
	if got, want := pc.FaceColors[0], low; got != want {
		t.Fatalf("first mapped face = %+v, want %+v", got, want)
	}
	if got, want := pc.FaceColors[1], high; got != want {
		t.Fatalf("second mapped face = %+v, want %+v", got, want)
	}
	if got, want := pc.EdgeColors[1], pc.FaceColors[1]; got != want {
		t.Fatalf("face-style edge = %+v, want mapped face %+v", got, want)
	}

	if err := pc.SetCLim(0, 20); err != nil {
		t.Fatalf("SetCLim: %v", err)
	}
	if pc.FaceColors[1] == high {
		t.Fatalf("SetCLim did not refresh mapped face colors: %+v", pc.FaceColors)
	}
	if got, want := pc.EdgeColors[1], pc.FaceColors[1]; got != want {
		t.Fatalf("face-style edge after clim = %+v, want mapped face %+v", got, want)
	}
}

func TestPathCollectionFallsBackWhenMarkerBatchDeclines(t *testing.T) {
	pc := &PathCollection{
		Path:          polygonPath([]geom.Pt{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 1}}, true),
		Offsets:       []geom.Pt{{X: 1, Y: 2}, {X: 4, Y: 5}},
		PathInDisplay: true,
		FaceColor:     render.Color{R: 1, A: 1},
	}

	r := &batchRecordingRenderer{returnNative: false}
	pc.Draw(r, createTestDrawContext())

	if len(r.markerBatches) != 1 {
		t.Fatalf("marker batches = %d, want attempted native marker batch", len(r.markerBatches))
	}
	if len(r.pathCollectionBatches) != 1 {
		t.Fatalf("path collection batches = %d, want attempted native path collection", len(r.pathCollectionBatches))
	}
	if len(r.pathCalls) != 2 {
		t.Fatalf("fallback path calls = %d, want 2", len(r.pathCalls))
	}
}

func TestPathCollectionFallsBackWhenPathCollectionBatchDeclines(t *testing.T) {
	pc := &PathCollection{
		Path:          polygonPath([]geom.Pt{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 1}}, true),
		Offsets:       []geom.Pt{{X: 1, Y: 2}, {X: 4, Y: 5}},
		Sizes:         []float64{2, 3},
		PathInDisplay: true,
		FaceColors: []render.Color{
			{R: 1, A: 1},
			{G: 1, A: 1},
		},
	}

	r := &batchRecordingRenderer{returnNative: false}
	pc.Draw(r, createTestDrawContext())

	if len(r.markerBatches) != 0 {
		t.Fatalf("marker batches = %d, want none for varying path collection", len(r.markerBatches))
	}
	if len(r.pathCollectionBatches) != 1 {
		t.Fatalf("path collection batches = %d, want attempted native path collection", len(r.pathCollectionBatches))
	}
	if len(r.pathCalls) != 2 {
		t.Fatalf("fallback path calls = %d, want 2", len(r.pathCalls))
	}
}

func TestPathCollectionNativeBatchCarriesHatchAntialiasAndSnap(t *testing.T) {
	pc := &PathCollection{
		Collection:    Collection{Alpha: 0.5, Antialias: render.AntialiasOff},
		Path:          polygonPath([]geom.Pt{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 1}}, true),
		Offsets:       []geom.Pt{{X: 1, Y: 2}},
		PathInDisplay: true,
		FaceColor:     render.Color{R: 0.2, G: 0.3, B: 0.4, A: 1},
		Hatch:         "/",
		HatchColor:    render.Color{R: 1, A: 1},
		HatchWidth:    1.25,
	}

	r := &nativeHatchBatchRecordingRenderer{
		batchRecordingRenderer: batchRecordingRenderer{returnNative: true},
	}
	pc.Draw(r, createTestDrawContext())

	if len(r.markerBatches) != 0 {
		t.Fatalf("marker batches = %d, want none for hatched path collection", len(r.markerBatches))
	}
	if len(r.pathCollectionBatches) != 1 {
		t.Fatalf("path collection batches = %d, want 1", len(r.pathCollectionBatches))
	}
	items := r.pathCollectionBatches[0].Items
	if len(items) != 1 {
		t.Fatalf("path collection items = %d, want 1", len(items))
	}
	item := items[0]
	if item.Hatch != "/" {
		t.Fatalf("hatch = %q, want /", item.Hatch)
	}
	if got, want := item.HatchColor, (render.Color{R: 1, A: 0.5}); got != want {
		t.Fatalf("hatch color = %+v, want %+v", got, want)
	}
	if item.HatchWidth != 1.25 {
		t.Fatalf("hatch width = %v, want 1.25", item.HatchWidth)
	}
	if item.Antialiased {
		t.Fatal("batch item antialias = true, want false")
	}
	if item.Paint.Snap != render.SnapAuto {
		t.Fatalf("paint snap = %v, want SnapAuto", item.Paint.Snap)
	}
}

func TestLineCollectionLegendEntry(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	ax.AddCollection(&LineCollection{
		Collection: Collection{Label: "segments"},
		Segments: [][]geom.Pt{
			{{X: 0, Y: 0}, {X: 1, Y: 1}},
			{{X: 1, Y: 0}, {X: 2, Y: 1}},
		},
		Color:     render.Color{R: 0.2, G: 0.2, B: 0.8, A: 1},
		LineWidth: 2,
	})

	entries := ax.AddLegend().collectEntries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 legend entry, got %d", len(entries))
	}
	if entries[0].kind != legendEntryLine || entries[0].Label != "segments" {
		t.Fatalf("unexpected legend entry: %+v", entries[0])
	}
}

func TestQuadMeshDrawsEachCell(t *testing.T) {
	mesh := &QuadMesh{
		PatchCollection: PatchCollection{
			Collection: Collection{Label: "mesh"},
			FaceColors: []render.Color{
				{R: 1, G: 0, B: 0, A: 1},
				{R: 0, G: 1, B: 0, A: 1},
				{R: 0, G: 0, B: 1, A: 1},
				{R: 1, G: 1, B: 0, A: 1},
			},
			EdgeColor: render.Color{R: 0, G: 0, B: 0, A: 1},
			EdgeWidth: 1,
		},
		XEdges: []float64{0, 1, 2},
		YEdges: []float64{0, 1, 2},
	}

	r := &recordingRenderer{}
	mesh.Draw(r, createTestDrawContext())

	if len(r.pathCalls) != 4 {
		t.Fatalf("expected 4 quad cells, got %d", len(r.pathCalls))
	}

	bounds := mesh.Bounds(nil)
	want := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 2, Y: 2}}
	if bounds != want {
		t.Fatalf("bounds = %+v, want %+v", bounds, want)
	}
}

func TestPatchCollectionUsesPathCollectionBatchWhenAvailable(t *testing.T) {
	pc := &PatchCollection{
		Collection: Collection{Alpha: 0.5},
		Paths: []geom.Path{
			patchRectPath(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}}),
			patchRectPath(geom.Rect{Min: geom.Pt{X: 1, Y: 1}, Max: geom.Pt{X: 2, Y: 2}}),
		},
		FaceColor: render.Color{R: 0.2, A: 1},
		EdgeColor: render.Color{A: 1},
		EdgeWidth: 1,
	}

	r := &batchRecordingRenderer{returnNative: true}
	pc.Draw(r, createTestDrawContext())

	if len(r.pathCollectionBatches) != 1 {
		t.Fatalf("path collection batches = %d, want 1", len(r.pathCollectionBatches))
	}
	if len(r.pathCollectionBatches[0].Items) != 2 {
		t.Fatalf("batch items = %d, want 2", len(r.pathCollectionBatches[0].Items))
	}
	if len(r.pathCalls) != 0 {
		t.Fatalf("fallback path calls = %d, want 0", len(r.pathCalls))
	}
}

func TestPatchCollectionWithHatchKeepsFallbackPath(t *testing.T) {
	pc := &PatchCollection{
		Paths: []geom.Path{
			patchRectPath(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}}),
		},
		FaceColor:  render.Color{R: 0.2, A: 1},
		Hatch:      "/",
		HatchColor: render.Color{A: 1},
		HatchWidth: 1,
	}

	r := &batchRecordingRenderer{returnNative: true}
	pc.Draw(r, createTestDrawContext())

	if len(r.pathCollectionBatches) != 0 {
		t.Fatal("hatched patch collection should not use path collection batch yet")
	}
	if len(r.pathCalls) == 0 {
		t.Fatal("hatched patch collection should draw via fallback path calls")
	}
}

func TestPatchCollectionNativeBatchCarriesHatchAntialiasAndSnap(t *testing.T) {
	pc := &PatchCollection{
		Collection: Collection{Alpha: 0.5, Antialias: render.AntialiasOff},
		Paths: []geom.Path{
			patchRectPath(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}}),
		},
		FaceColor:  render.Color{R: 0.2, A: 1},
		Hatch:      "x",
		HatchColor: render.Color{G: 1, A: 1},
		HatchWidth: 2,
	}

	r := &nativeHatchBatchRecordingRenderer{
		batchRecordingRenderer: batchRecordingRenderer{returnNative: true},
	}
	pc.Draw(r, createTestDrawContext())

	if len(r.pathCollectionBatches) != 1 {
		t.Fatalf("path collection batches = %d, want 1", len(r.pathCollectionBatches))
	}
	items := r.pathCollectionBatches[0].Items
	if len(items) != 1 {
		t.Fatalf("batch items = %d, want 1", len(items))
	}
	item := items[0]
	if item.Hatch != "x" {
		t.Fatalf("hatch = %q, want x", item.Hatch)
	}
	if got, want := item.HatchColor, (render.Color{G: 1, A: 0.5}); got != want {
		t.Fatalf("hatch color = %+v, want %+v", got, want)
	}
	if item.HatchWidth != 2 {
		t.Fatalf("hatch width = %v, want 2", item.HatchWidth)
	}
	if item.Antialiased {
		t.Fatal("batch item antialias = true, want false")
	}
	if item.Paint.Snap != render.SnapAuto {
		t.Fatalf("paint snap = %v, want SnapAuto", item.Paint.Snap)
	}
}

func TestQuadMeshUsesNativeBatchWhenAvailable(t *testing.T) {
	mesh := &QuadMesh{
		PatchCollection: PatchCollection{
			FaceColor: render.Color{R: 1, A: 1},
		},
		XEdges: []float64{0, 1, 2},
		YEdges: []float64{0, 1, 2},
	}

	r := &batchRecordingRenderer{returnNative: true}
	mesh.Draw(r, createTestDrawContext())

	if len(r.quadMeshBatches) != 1 {
		t.Fatalf("quad mesh batches = %d, want 1", len(r.quadMeshBatches))
	}
	if len(r.quadMeshBatches[0].Cells) != 4 {
		t.Fatalf("quad mesh cells = %d, want 4", len(r.quadMeshBatches[0].Cells))
	}
	if len(r.pathCalls) != 0 {
		t.Fatalf("fallback path calls = %d, want 0", len(r.pathCalls))
	}
}

func TestFillBetweenPolyCollectionBounds(t *testing.T) {
	fill := &FillBetweenPolyCollection{
		PatchCollection: PatchCollection{
			Collection: Collection{Label: "band"},
			FaceColor:  render.Color{R: 0.2, G: 0.6, B: 0.8, A: 0.5},
		},
		X:        []float64{0, 1, 2},
		Y1:       []float64{1, 2, 1.5},
		Baseline: 0,
	}

	bounds := fill.Bounds(nil)
	want := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 2, Y: 2}}
	if bounds != want {
		t.Fatalf("bounds = %+v, want %+v", bounds, want)
	}
}

func TestScatterCustomMarkerPathDrawsViaPathCollection(t *testing.T) {
	custom := polygonPath([]geom.Pt{
		{X: 0, Y: -0.5},
		{X: 0.5, Y: 0.5},
		{X: -0.5, Y: 0.5},
	}, true)

	scatter := &Scatter2D{
		XY:         []geom.Pt{{X: 1, Y: 1}, {X: 2, Y: 2}},
		Sizes:      []float64{4, 6},
		MarkerPath: custom,
		Color:      render.Color{R: 0.9, G: 0.2, B: 0.2, A: 1},
		EdgeColor:  render.Color{R: 0.1, G: 0.1, B: 0.1, A: 1},
		EdgeWidth:  1,
		Label:      "custom",
	}

	r := &recordingRenderer{}
	scatter.Draw(r, createTestDrawContext())

	if len(r.pathCalls) != 2 {
		t.Fatalf("expected 2 marker paths, got %d", len(r.pathCalls))
	}
	entry, ok := scatter.legendEntry()
	if !ok || len(entry.markerPath.C) == 0 {
		t.Fatalf("expected custom marker path in legend entry, got %+v", entry)
	}
}

func TestBarAndErrorbarContainers(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	bars := ax.BarContainer([]float64{1, 2}, []float64{3, -1}, BarOptions{Label: "bars"})
	if bars == nil || bars.Len() != 2 {
		t.Fatalf("unexpected bar container: %+v", bars)
	}
	if got := bars.Patches[0].Bounds(nil); got.Min.X >= got.Max.X || got.Min.Y >= got.Max.Y {
		t.Fatalf("expected concrete rectangle bounds, got %+v", got)
	}

	errs := ax.ErrorBarContainer([]float64{1, 2}, []float64{3, 4}, []float64{0.1}, []float64{0.2}, ErrorBarOptions{Label: "errs"})
	if errs == nil || errs.Len() != 2 {
		t.Fatalf("unexpected errorbar container: %+v", errs)
	}
	if len(errs.Artists()) != 1 {
		t.Fatalf("expected one errorbar artist, got %d", len(errs.Artists()))
	}
}

func TestStemContainerAddsArtists(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	container := ax.Stem([]float64{0, 1, 2}, []float64{1, 3, 2}, StemOptions{Label: "stem"})
	if container == nil {
		t.Fatal("expected stem container")
	}
	if container.Len() != 3 {
		t.Fatalf("stem len = %d, want 3", container.Len())
	}
	if len(container.Artists()) != 3 {
		t.Fatalf("expected 3 child artists, got %d", len(container.Artists()))
	}
	if len(ax.Artists) != 3 {
		t.Fatalf("expected stem artists to be added to axes, got %d", len(ax.Artists))
	}
	palette := ax.resolvedRC().Palette()
	if got, want := container.StemLines.Color, palette[0]; got != want {
		t.Fatalf("stem line color = %+v, want Matplotlib linefmt C0 %+v", got, want)
	}
	if got, want := container.MarkerCollection.FaceColor, palette[0]; got != want {
		t.Fatalf("stem marker color = %+v, want Matplotlib markerfmt C0 %+v", got, want)
	}
	if got, want := container.MarkerCollection.Size, pointsToPixels(ax.resolvedRC(), 6); !approx(got, want, 1e-12) {
		t.Fatalf("stem marker size = %v, want Matplotlib 6 point Line2D marker diameter %v", got, want)
	}
	if got, want := container.Baseline.Col, palette[3]; got != want {
		t.Fatalf("stem baseline color = %+v, want Matplotlib basefmt C3 %+v", got, want)
	}
}
