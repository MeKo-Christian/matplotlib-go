package core

import (
	"math"
	"reflect"
	"testing"

	matcolor "github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
)

var pathCollectionDisplayPathSink geom.Path

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

func TestPathCollectionDisplayPathInDisplayCombinesScaleAndTranslate(t *testing.T) {
	pc := &PathCollection{
		Path:          markerRectanglePath(-0.5, -0.5, 0.5, 0.5),
		Offsets:       []geom.Pt{{X: 10, Y: 20}},
		Size:          3,
		PathInDisplay: true,
	}

	got := pc.displayPathAt(nil, 0, pc.Path)
	want := scaleAndTranslatePath(pc.Path, 3, geom.Pt{X: 10, Y: 20})
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("displayPathAt = %+v, want %+v", got, want)
	}

	allocs := testing.AllocsPerRun(1000, func() {
		pathCollectionDisplayPathSink = pc.displayPathAt(nil, 0, pc.Path)
	})
	if allocs > 2 {
		t.Fatalf("displayPathAt PathInDisplay allocations = %.2f, want <= 2", allocs)
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

func TestPathCollectionOffsetCoordsCanDifferFromPathTransform(t *testing.T) {
	pc := &PathCollection{
		Collection:    Collection{},
		Path:          markerRectanglePath(-1, -1, 1, 1),
		Offsets:       []geom.Pt{{X: 2, Y: 8}},
		PathInDisplay: true,
		FaceColor:     render.Color{A: 1},
	}
	pc.SetTransformCoords(Coords(CoordFigure))
	pc.SetOffsetCoords(Coords(CoordData))

	r := &recordingRenderer{}
	pc.Draw(r, createTestDrawContext())

	if len(r.pathCalls) != 1 {
		t.Fatalf("path call count = %d, want 1", len(r.pathCalls))
	}
	if got := r.pathCalls[0].path.V[0]; got != (geom.Pt{X: 69, Y: 369}) {
		t.Fatalf("first display-path point = %+v, want data offset with display marker delta", got)
	}
}

func TestPathCollectionOffsetCoordsControlBoundsForDisplayPaths(t *testing.T) {
	pc := &PathCollection{
		Collection:    Collection{},
		Path:          markerRectanglePath(-1, -1, 1, 1),
		Offsets:       []geom.Pt{{X: 2, Y: 8}},
		PathInDisplay: true,
		FaceColor:     render.Color{A: 1},
	}
	pc.SetTransformCoords(Coords(CoordFigure))
	if got := pc.Bounds(createTestDrawContext()); got != (geom.Rect{}) {
		t.Fatalf("figure-coordinate display-path bounds = %+v, want empty data bounds", got)
	}

	pc.SetOffsetCoords(Coords(CoordData))
	want := geom.Rect{Min: geom.Pt{X: 2, Y: 8}, Max: geom.Pt{X: 2, Y: 8}}
	if got := pc.Bounds(createTestDrawContext()); got != want {
		t.Fatalf("data-offset display-path bounds = %+v, want %+v", got, want)
	}
}

func TestPathCollectionUsesPathCollectionForVaryingPerItemStyle(t *testing.T) {
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
		t.Fatalf("marker batches = %d, want none for varying per-item PathCollection style", len(r.markerBatches))
	}
	if len(r.pathCollectionBatches) != 1 {
		t.Fatalf("path collection batches = %d, want 1", len(r.pathCollectionBatches))
	}
	if len(r.pathCalls) != 0 {
		t.Fatalf("fallback path calls = %d, want 0", len(r.pathCalls))
	}

	items := r.pathCollectionBatches[0].Items
	if len(items) != 2 {
		t.Fatalf("path collection items = %d, want 2", len(items))
	}
	if got, want := items[0].Paint.Fill.A, 0.4; math.Abs(got-want) > 1e-12 {
		t.Fatalf("first fill alpha = %v, want %v", got, want)
	}
	if got, want := items[1].Paint.Stroke.A, 0.35; math.Abs(got-want) > 1e-12 {
		t.Fatalf("second stroke alpha = %v, want %v", got, want)
	}
	if got, want := items[0].Paint.LineWidth, (1.25 * 100.0 / 72.0); got != want {
		t.Fatalf("first linewidth = %v, want %v", got, want)
	}
	if got, want := items[1].Paint.LineWidth, (2.5 * 100.0 / 72.0); got != want {
		t.Fatalf("second linewidth = %v, want %v", got, want)
	}
	firstBounds, ok := pathBounds(items[0].Path)
	if !ok {
		t.Fatal("first path collection item has no bounds")
	}
	secondBounds, ok := pathBounds(items[1].Path)
	if !ok {
		t.Fatal("second path collection item has no bounds")
	}
	if got, want := firstBounds.W(), 2.0; got != want {
		t.Fatalf("first marker path width = %v, want %v", got, want)
	}
	if got, want := secondBounds.W(), 3.0; got != want {
		t.Fatalf("second marker path width = %v, want %v", got, want)
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
	if got, want := paint.LineWidth, (1.5 * 100.0 / 72.0); got != want {
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

	if len(r.markerBatches) != 0 {
		t.Fatalf("marker batches = %d, want none for varying face colors", len(r.markerBatches))
	}
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

	if got := pc.Array(); len(got) != 2 || got[0] != 0 || got[1] != 10 {
		t.Fatalf("Array = %v, want copied scalar values", got)
	}
	if got, want := pc.FaceColors[0], low; got != want {
		t.Fatalf("first mapped face = %+v, want %+v", got, want)
	}
	if got, want := pc.FaceColors[1], high; got != want {
		t.Fatalf("second mapped face = %+v, want %+v", got, want)
	}
	if got, want := pc.edgeColorAt(1), pc.FaceColors[1]; got != want {
		t.Fatalf("face-style edge = %+v, want mapped face %+v", got, want)
	}

	if err := pc.SetCLim(0, 20); err != nil {
		t.Fatalf("SetCLim: %v", err)
	}
	if pc.FaceColors[1] == high {
		t.Fatalf("SetCLim did not refresh mapped face colors: %+v", pc.FaceColors)
	}
	if got, want := pc.edgeColorAt(1), pc.FaceColors[1]; got != want {
		t.Fatalf("face-style edge after clim = %+v, want mapped face %+v", got, want)
	}
}

func TestPathCollectionMutableSettersCloneAndMarkStale(t *testing.T) {
	pc := &PathCollection{}
	pc.SetStale(false)

	offsets := []geom.Pt{{X: 1, Y: 2}}
	pc.SetOffsets(offsets)
	offsets[0] = geom.Pt{X: 9, Y: 9}
	if got, want := pc.Offsets[0], (geom.Pt{X: 1, Y: 2}); got != want {
		t.Fatalf("offset clone = %+v, want %+v", got, want)
	}
	if !pc.Stale() {
		t.Fatal("SetOffsets did not mark collection stale")
	}

	pc.SetStale(false)
	sizes := []float64{4, 9}
	pc.SetSizes(sizes)
	sizes[0] = 99
	if got, want := pc.Sizes[0], 4.0; got != want {
		t.Fatalf("size clone = %v, want %v", got, want)
	}
	if !pc.Stale() {
		t.Fatal("SetSizes did not mark collection stale")
	}

	faces := []render.Color{{R: 1, A: 1}}
	pc.SetFaceColors(faces)
	faces[0] = render.Color{B: 1, A: 1}
	if got, want := pc.FaceColors[0], (render.Color{R: 1, A: 1}); got != want {
		t.Fatalf("face color clone = %+v, want %+v", got, want)
	}

	edges := []render.Color{{G: 1, A: 1}}
	pc.SetEdgeColors(edges)
	edges[0] = render.Color{R: 1, A: 1}
	if got, want := pc.EdgeColors[0], (render.Color{G: 1, A: 1}); got != want {
		t.Fatalf("edge color clone = %+v, want %+v", got, want)
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

func TestPathCollectionFallsBackWhenMarkerAndPathCollectionBatchesDecline(t *testing.T) {
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
		t.Fatalf("marker batches = %d, want none for varying per-item collection style", len(r.markerBatches))
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
	if want := (1.25 * 100.0 / 72.0); item.HatchWidth != want {
		t.Fatalf("hatch width = %v, want %v", item.HatchWidth, want)
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

func TestLineCollectionDrawUsesMatplotlibAutoSnap(t *testing.T) {
	lines := &LineCollection{
		Collection: Collection{Alpha: 1},
		Segments:   [][]geom.Pt{{{X: 0, Y: 1}, {X: 4, Y: 1}}},
		Color:      render.Color{A: 1},
		LineWidth:  1.4,
	}

	r := &recordingRenderer{}
	lines.Draw(r, createTestDrawContext())

	if len(r.pathCalls) != 1 {
		t.Fatalf("path calls = %d, want 1", len(r.pathCalls))
	}
	if got := r.pathCalls[0].paint.Snap; got != render.SnapAuto {
		t.Fatalf("line collection snap = %v, want Matplotlib snap=None/SnapAuto", got)
	}
}

func TestAxesHLinesBroadcastsEndpointsAndRegistersCollection(t *testing.T) {
	ax := NewFigure(640, 360).AddAxes(geom.Rect{})
	color := render.Color{R: 0.1, G: 0.2, B: 0.3, A: 1}

	lines := ax.HLines(
		[]float64{1, 2, 3},
		[]float64{0},
		[]float64{4},
		LineCollectionOptions{
			Label:     "thresholds",
			Alpha:     optional.Of(0.75),
			Color:     optional.Of(color),
			LineWidth: optional.Of(2.5),
			LineCap:   render.CapSquare,
		},
	)

	if lines == nil {
		t.Fatal("expected HLines collection")
	}
	if len(lines.Segments) != 3 {
		t.Fatalf("segments = %d, want 3", len(lines.Segments))
	}
	if got := lines.Segments[1]; got[0] != (geom.Pt{X: 0, Y: 2}) || got[1] != (geom.Pt{X: 4, Y: 2}) {
		t.Fatalf("second horizontal segment = %+v", got)
	}
	if len(ax.Artists) != 1 || ax.Artists[0] != lines {
		t.Fatalf("registered artists = %d, want returned collection", len(ax.Artists))
	}
	if lines.Coords != Coords(CoordData) || lines.Label != "thresholds" || lines.Alpha != 0.75 {
		t.Fatalf("collection metadata = coords=%+v label=%q alpha=%v", lines.Coords, lines.Label, lines.Alpha)
	}
	if lines.Color != color || lines.LineWidth != 2.5 || lines.LineCap != render.CapSquare {
		t.Fatalf("line style = color=%+v width=%v cap=%v", lines.Color, lines.LineWidth, lines.LineCap)
	}
}

func TestAxesVLinesBroadcastsExtentsAndRejectsMismatchedLengths(t *testing.T) {
	ax := NewFigure(640, 360).AddAxes(geom.Rect{})

	lines := ax.VLines([]float64{1, 2}, []float64{-1}, []float64{3}, LineCollectionOptions{})
	if lines == nil {
		t.Fatal("expected VLines collection")
	}
	if len(lines.Segments) != 2 {
		t.Fatalf("segments = %d, want 2", len(lines.Segments))
	}
	if got := lines.Segments[0]; got[0] != (geom.Pt{X: 1, Y: -1}) || got[1] != (geom.Pt{X: 1, Y: 3}) {
		t.Fatalf("first vertical segment = %+v", got)
	}
	if got := ax.VLines([]float64{1, 2}, []float64{0, 1, 2}, []float64{3}, LineCollectionOptions{}); got != nil {
		t.Fatalf("VLines with mismatched lengths returned %#v, want nil", got)
	}
	if got := ax.HLines([]float64{1, 2}, []float64{0}, []float64{3, 4, 5}, LineCollectionOptions{}); got != nil {
		t.Fatalf("HLines with mismatched lengths returned %#v, want nil", got)
	}
}

func TestLineCollectionSetArrayRefreshesStrokeColors(t *testing.T) {
	cmapName := "linecollection-scalar-array"
	low := render.Color{R: 1, A: 1}
	high := render.Color{B: 1, A: 1}
	matcolor.RegisterColormap(cmapName, matcolor.NewColormap(cmapName, []matcolor.ColorStop{
		{Pos: 0, Color: low},
		{Pos: 1, Color: high},
	}))

	lines := &LineCollection{
		Collection: Collection{
			Colormap: cmapName,
		},
		Segments: [][]geom.Pt{
			{{X: 0, Y: 0}, {X: 1, Y: 1}},
			{{X: 1, Y: 0}, {X: 2, Y: 1}},
		},
		Color:     render.Color{G: 1, A: 1},
		LineWidth: 1,
	}

	if err := lines.SetArray([]float64{0, 10}); err != nil {
		t.Fatalf("SetArray: %v", err)
	}
	if got := lines.Array(); len(got) != 2 || got[0] != 0 || got[1] != 10 {
		t.Fatalf("Array = %v, want copied scalar values", got)
	}
	if got, want := lines.Colors[0], low; got != want {
		t.Fatalf("first line color = %+v, want %+v", got, want)
	}
	if got, want := lines.Colors[1], high; got != want {
		t.Fatalf("last line color = %+v, want %+v", got, want)
	}

	lines.SetColormap("plasma")
	if got := lines.ScalarMap().Colormap; got != "plasma" {
		t.Fatalf("line collection colormap = %q, want plasma", got)
	}
	if lines.Colors[0] == low && lines.Colors[1] == high {
		t.Fatal("line colors did not refresh after colormap update")
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
	if want := (2 * 100.0 / 72.0); item.HatchWidth != want {
		t.Fatalf("hatch width = %v, want %v", item.HatchWidth, want)
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

func TestQuadMeshSetArrayRefreshesFlatColorsAndFaceEdges(t *testing.T) {
	cmapName := "quadmesh-flat-scalar-array"
	low := render.Color{R: 1, A: 1}
	high := render.Color{B: 1, A: 1}
	matcolor.RegisterColormap(cmapName, matcolor.NewColormap(cmapName, []matcolor.ColorStop{
		{Pos: 0, Color: low},
		{Pos: 1, Color: high},
	}))

	mesh := &QuadMesh{
		PatchCollection: PatchCollection{
			Collection: Collection{
				Colormap: cmapName,
			},
			FaceColor: render.Color{G: 1, A: 1},
			EdgeColor: render.Color{A: 1},
			EdgeWidth: 1,
		},
		XEdges:  []float64{0, 1, 2},
		YEdges:  []float64{0, 1, 2},
		Shading: MeshShadingFlat,
	}
	mesh.SetEdgeColorFace()

	if err := mesh.SetArray([]float64{0, 10, 5, 15}); err != nil {
		t.Fatalf("SetArray: %v", err)
	}
	if got := mesh.Array(); len(got) != 4 || got[0] != 0 || got[3] != 15 {
		t.Fatalf("Array = %v, want copied flattened mesh values", got)
	}
	if got, want := mesh.FaceColors[0], low; got != want {
		t.Fatalf("first flat face = %+v, want %+v", got, want)
	}
	if got, want := mesh.FaceColors[3], high; got != want {
		t.Fatalf("last flat face = %+v, want %+v", got, want)
	}
	if got, want := colorAt(mesh.EdgeColor, mesh.resolvedEdgeColors(), 3), mesh.FaceColors[3]; got != want {
		t.Fatalf("face-style flat edge = %+v, want mapped face %+v", got, want)
	}

	mesh.SetColormap("plasma")
	if got := mesh.ScalarMap().Colormap; got != "plasma" {
		t.Fatalf("mesh colormap = %q, want plasma", got)
	}
	if got, want := colorAt(mesh.EdgeColor, mesh.resolvedEdgeColors(), 3), mesh.FaceColors[3]; got != want {
		t.Fatalf("face-style edge after cmap = %+v, want mapped face %+v", got, want)
	}
}

func TestQuadMeshSetEdgesClonesValidatesAndMarksStale(t *testing.T) {
	mesh := &QuadMesh{
		XEdges:  []float64{0, 1, 2},
		YEdges:  []float64{0, 1},
		Shading: MeshShadingFlat,
	}
	if err := mesh.SetArray([]float64{1, 2}); err != nil {
		t.Fatalf("SetArray: %v", err)
	}
	mesh.SetStale(false)

	xEdges := []float64{-1, 0, 1}
	yEdges := []float64{2, 4}
	if err := mesh.SetEdges(xEdges, yEdges); err != nil {
		t.Fatalf("SetEdges: %v", err)
	}
	xEdges[0] = 99
	yEdges[0] = 99
	if got, want := mesh.XEdges[0], -1.0; got != want {
		t.Fatalf("x edge clone = %v, want %v", got, want)
	}
	if got, want := mesh.YEdges[0], 2.0; got != want {
		t.Fatalf("y edge clone = %v, want %v", got, want)
	}
	if !mesh.Stale() {
		t.Fatal("SetEdges did not mark mesh stale")
	}
	if got, want := mesh.Bounds(nil), (geom.Rect{Min: geom.Pt{X: -1, Y: 2}, Max: geom.Pt{X: 1, Y: 4}}); got != want {
		t.Fatalf("bounds after SetEdges = %+v, want %+v", got, want)
	}
	if err := mesh.SetEdges([]float64{0, 1, 2, 3}, []float64{0, 1}); err == nil {
		t.Fatal("SetEdges accepted coordinates incompatible with existing scalar array")
	}
	if err := mesh.SetEdges([]float64{0, math.NaN(), 2}, []float64{0, 1}); err == nil {
		t.Fatal("SetEdges accepted non-finite coordinates")
	}
}

func TestQuadMeshSetArrayKeepsBadUnderOverColorsAfterMappingChanges(t *testing.T) {
	cmapName := "quadmesh-flat-scalar-array-bounds"
	bad := render.Color{R: 1, G: 0.2, A: 1}
	under := render.Color{R: 0.2, G: 0.4, A: 1}
	over := render.Color{R: 0.4, B: 1, A: 1}
	low := render.Color{G: 1, A: 1}
	high := render.Color{B: 1, A: 1}
	matcolor.RegisterColormap(cmapName, matcolor.NewColormap(cmapName, []matcolor.ColorStop{
		{Pos: 0, Color: low},
		{Pos: 1, Color: high},
	}).WithBad(bad).WithUnder(under).WithOver(over))

	mesh := &QuadMesh{
		PatchCollection: PatchCollection{
			Collection: Collection{
				Colormap: cmapName,
			},
		},
		XEdges:  []float64{0, 1, 2},
		YEdges:  []float64{0, 1, 2},
		Shading: MeshShadingFlat,
	}

	if err := mesh.SetNorm(Normalize{VMin: 0, VMax: 1}); err != nil {
		t.Fatalf("SetNorm: %v", err)
	}
	if err := mesh.SetArray([]float64{-1, math.NaN(), 0.5, 2}); err != nil {
		t.Fatalf("SetArray: %v", err)
	}
	if got := mesh.FaceColors[0]; got != under {
		t.Fatalf("under face color = %+v, want under color %+v", got, under)
	}
	if got := mesh.FaceColors[1]; got != bad {
		t.Fatalf("NaN face color = %+v, want bad color %+v", got, bad)
	}
	if got := mesh.FaceColors[3]; got != over {
		t.Fatalf("over face color = %+v, want over color %+v", got, over)
	}

	if err := mesh.SetCLim(0, 1); err != nil {
		t.Fatalf("SetCLim: %v", err)
	}
	if got := mesh.FaceColors[0]; got != under {
		t.Fatalf("under face after clim = %+v, want under color %+v", got, under)
	}
	if got := mesh.FaceColors[1]; got != bad {
		t.Fatalf("NaN face after clim = %+v, want bad color %+v", got, bad)
	}
	if got := mesh.FaceColors[3]; got != over {
		t.Fatalf("over face after clim = %+v, want over color %+v", got, over)
	}
}

func TestQuadMeshSetArrayRefreshesGouraudValuesAndFallbackFaces(t *testing.T) {
	cmapName := "quadmesh-gouraud-scalar-array"
	low := render.Color{R: 1, A: 1}
	high := render.Color{B: 1, A: 1}
	matcolor.RegisterColormap(cmapName, matcolor.NewColormap(cmapName, []matcolor.ColorStop{
		{Pos: 0, Color: low},
		{Pos: 1, Color: high},
	}))

	mesh := &QuadMesh{
		PatchCollection: PatchCollection{
			Collection: Collection{
				Colormap: cmapName,
			},
		},
		XEdges:  []float64{0, 1},
		YEdges:  []float64{0, 1},
		Shading: MeshShadingGouraud,
	}

	if err := mesh.SetArray([]float64{0, 10, 10, 20}); err != nil {
		t.Fatalf("SetArray: %v", err)
	}
	if len(mesh.Values) != 2 || len(mesh.Values[0]) != 2 || mesh.Values[1][1] != 20 {
		t.Fatalf("gouraud values = %v, want updated 2x2 grid", mesh.Values)
	}
	if len(mesh.FaceColors) != 1 {
		t.Fatalf("gouraud fallback faces = %d, want 1", len(mesh.FaceColors))
	}
	if got, want := mesh.FaceColors[0], mesh.ScalarMap().Color(10, 1); got != want {
		t.Fatalf("gouraud fallback face = %+v, want average mapped face %+v", got, want)
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

	errs, err := ax.ErrorBarContainer([]float64{1, 2}, []float64{3, 4}, []float64{0.1}, []float64{0.2}, ErrorBarOptions{Label: "errs"})
	if err != nil {
		t.Fatalf("ErrorBarContainer() returned error: %v", err)
	}
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
	if got, want := container.StemLines.LineWidth, 1.5; !approx(got, want, 1e-12) {
		t.Fatalf("stem line width = %v, want Matplotlib default 1.5 pt = %v px", got, want)
	}
	if got, want := container.MarkerCollection.EdgeWidth, 1.0; !approx(got, want, 1e-12) {
		t.Fatalf("stem marker edge width = %v, want Matplotlib default 1 pt = %v px", got, want)
	}
	if got, want := container.Baseline.Col, palette[3]; got != want {
		t.Fatalf("stem baseline color = %+v, want Matplotlib basefmt C3 %+v", got, want)
	}
	if got, want := container.Baseline.W, 1.5; !approx(got, want, 1e-12) {
		t.Fatalf("stem baseline width = %v, want Matplotlib default 1.5 pt = %v px", got, want)
	}
}
