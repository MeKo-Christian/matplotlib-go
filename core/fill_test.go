package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestFill2D_Draw_FillBetween(t *testing.T) {
	// Create a fill between two curves
	fill := &Fill2D{
		X:     []float64{0, 1, 2, 3, 4},
		Y1:    []float64{1, 3, 2, 4, 1},
		Y2:    []float64{0, 1, 0.5, 2, 0},
		Color: render.Color{R: 0.5, G: 0.7, B: 0.9, A: 0.6},
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic
	fill.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestFill2D_Draw_FillToBaseline(t *testing.T) {
	// Create a fill to baseline
	fill := &Fill2D{
		X:        []float64{0, 1, 2, 3, 4},
		Y1:       []float64{2, 4, 3, 5, 2},
		Y2:       nil, // use baseline
		Baseline: 1.0,
		Color:    render.Color{R: 1, G: 0.5, B: 0.2, A: 0.8},
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic
	fill.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestFill2D_EmptyData(t *testing.T) {
	// Test with empty data
	fill := &Fill2D{
		X:     []float64{},
		Y1:    []float64{},
		Color: render.Color{R: 1, G: 0, B: 0, A: 1},
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic with empty data
	fill.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestFill2D_SinglePoint(t *testing.T) {
	// Test with single point (should not draw)
	fill := &Fill2D{
		X:     []float64{1},
		Y1:    []float64{2},
		Color: render.Color{R: 1, G: 0, B: 0, A: 1},
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic but also should not draw
	fill.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestFill2D_MismatchedLengths(t *testing.T) {
	// Test with mismatched array lengths
	fill := &Fill2D{
		X:     []float64{0, 1, 2, 3, 4}, // 5 elements
		Y1:    []float64{1, 3, 2},       // 3 elements
		Y2:    []float64{0, 1},          // 2 elements
		Color: render.Color{R: 0.5, G: 0.5, B: 1, A: 0.5},
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should only use min length (2 in this case)
	fill.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestFill2D_AlphaOverride(t *testing.T) {
	// Test alpha override
	fill := &Fill2D{
		X:     []float64{0, 1, 2},
		Y1:    []float64{1, 2, 1},
		Color: render.Color{R: 1, G: 0, B: 0, A: 1}, // original alpha = 1
		Alpha: 0.3,                                  // override to 0.3
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should use alpha override
	fill.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestFill2D_ZOrder(t *testing.T) {
	fill := &Fill2D{z: 1.5}

	if got := fill.Z(); got != 1.5 {
		t.Errorf("Expected Z() = 1.5, got %v", got)
	}
}

func TestFill2D_Bounds(t *testing.T) {
	// Test empty fill
	fill := &Fill2D{}
	bounds := fill.Bounds(nil)
	expected := geom.Rect{}
	if bounds != expected {
		t.Errorf("Expected Bounds() = %v, got %v", expected, bounds)
	}

	// Test fill with data
	fill = &Fill2D{
		X:  []float64{0, 2, 1},
		Y1: []float64{1, 3, 2},
		Y2: []float64{0, 1, 0.5},
	}
	bounds = fill.Bounds(nil)

	// Should include all X and Y values
	if bounds.Min.X != 0 || bounds.Max.X != 2 {
		t.Errorf("Expected X bounds [0, 2], got [%v, %v]", bounds.Min.X, bounds.Max.X)
	}
	if bounds.Min.Y != 0 || bounds.Max.Y != 3 {
		t.Errorf("Expected Y bounds [0, 3], got [%v, %v]", bounds.Min.Y, bounds.Max.Y)
	}

	// Test with baseline
	fill = &Fill2D{
		X:        []float64{0, 1, 2},
		Y1:       []float64{2, 3, 1},
		Y2:       nil,
		Baseline: -1.0,
	}
	bounds = fill.Bounds(nil)

	// Should include baseline in Y bounds
	if bounds.Min.Y != -1.0 || bounds.Max.Y != 3 {
		t.Errorf("Expected Y bounds [-1, 3], got [%v, %v]", bounds.Min.Y, bounds.Max.Y)
	}
}

func TestFillBetween(t *testing.T) {
	x := []float64{0, 1, 2}
	y1 := []float64{1, 2, 1}
	y2 := []float64{0, 1, 0}
	color := render.Color{R: 1, G: 0, B: 0, A: 0.5}

	fill := FillBetween(x, y1, y2, color)

	if len(fill.X) != len(x) {
		t.Errorf("Expected X length %d, got %d", len(x), len(fill.X))
	}
	if len(fill.Y1) != len(y1) {
		t.Errorf("Expected Y1 length %d, got %d", len(y1), len(fill.Y1))
	}
	if len(fill.Y2) != len(y2) {
		t.Errorf("Expected Y2 length %d, got %d", len(y2), len(fill.Y2))
	}
	if fill.Color != color {
		t.Errorf("Expected color %v, got %v", color, fill.Color)
	}
}

func TestFillToBaseline(t *testing.T) {
	x := []float64{0, 1, 2}
	y := []float64{1, 2, 1}
	baseline := 0.5
	color := render.Color{R: 0, G: 1, B: 0, A: 0.7}

	fill := FillToBaseline(x, y, baseline, color)

	if len(fill.X) != len(x) {
		t.Errorf("Expected X length %d, got %d", len(x), len(fill.X))
	}
	if len(fill.Y1) != len(y) {
		t.Errorf("Expected Y1 length %d, got %d", len(y), len(fill.Y1))
	}
	if fill.Y2 != nil {
		t.Errorf("Expected Y2 to be nil, got %v", fill.Y2)
	}
	if fill.Baseline != baseline {
		t.Errorf("Expected baseline %v, got %v", baseline, fill.Baseline)
	}
	if fill.Color != color {
		t.Errorf("Expected color %v, got %v", color, fill.Color)
	}
}

func TestAxesFillCreatesClosedPolygonCollection(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	color := render.Color{R: 0.2, G: 0.4, B: 0.6, A: 0.5}
	edge := render.Color{R: 0.1, G: 0.1, B: 0.1, A: 1}
	edgeWidth := 1.5

	fill := ax.Fill(
		[]float64{0, 2, 1},
		[]float64{0, 0, 3},
		FillOptions{Color: &color, EdgeColor: &edge, EdgeWidth: &edgeWidth, Label: "triangle"},
	)
	if fill == nil {
		t.Fatal("Fill() returned nil")
	}
	if len(fill.Polygons) != 1 || len(fill.Polygons[0]) != 3 {
		t.Fatalf("Fill polygon size = %v, want one triangle", fill.Polygons)
	}
	if fill.FaceColors[0] != color {
		t.Fatalf("Fill face color = %v, want %v", fill.FaceColors[0], color)
	}
	if fill.EdgeColor != edge || fill.EdgeWidth != edgeWidth {
		t.Fatalf("Fill edge style = (%v, %v), want (%v, %v)", fill.EdgeColor, fill.EdgeWidth, edge, edgeWidth)
	}
	if fill.Label != "triangle" {
		t.Fatalf("Fill label = %q, want triangle", fill.Label)
	}
}

func TestFillPlotPreservesColorAlphaWhenAlphaOmitted(t *testing.T) {
	fig := NewFigure(320, 240)
	ax := fig.AddAxes(unitRect())

	fillColor := render.Color{R: 0.36, G: 0.56, B: 0.92, A: 0.2}
	fill := ax.FillToBaseline(
		[]float64{0, 1, 2},
		[]float64{1, 2, 1},
		FillOptions{Color: &fillColor},
	)
	if fill == nil {
		t.Fatal("expected fill artist")
	}
	if fill.Alpha != 0 {
		t.Fatalf("omitted alpha stored Alpha = %v, want 0 sentinel", fill.Alpha)
	}

	r := &recordingRenderer{}
	fill.Draw(r, createTestDrawContext())
	if len(r.pathCalls) != 1 {
		t.Fatalf("path calls = %d, want 1", len(r.pathCalls))
	}
	if got := r.pathCalls[0].paint.Fill.A; got != fillColor.A {
		t.Fatalf("drawn fill alpha = %v, want color alpha %v", got, fillColor.A)
	}
}

func TestFillPlotExplicitAlphaOverridesColorAlpha(t *testing.T) {
	fig := NewFigure(320, 240)
	ax := fig.AddAxes(unitRect())

	fillColor := render.Color{R: 0.3, G: 0.7, B: 0.9, A: 0.7}
	alpha := 0.4
	fill, err := ax.FillBetween(
		[]float64{0, 1, 2},
		[]float64{1, 2, 1},
		[]float64{0, 1, 0},
		FillOptions{Color: &fillColor, Alpha: &alpha},
	)
	if err != nil {
		t.Fatalf("FillBetween() returned error: %v", err)
	}
	if fill == nil {
		t.Fatal("expected fill artist")
	}
	// An explicit alpha is baked into the resolved color (so alpha=0 is
	// honored); the legacy float64 Alpha field stays at its 0 "unset" sentinel.
	if fill.Alpha != 0 {
		t.Fatalf("stored Alpha = %v, want 0 sentinel (alpha baked into color)", fill.Alpha)
	}
	if fill.Color.A != alpha {
		t.Fatalf("baked color alpha = %v, want explicit alpha %v", fill.Color.A, alpha)
	}

	r := &recordingRenderer{}
	fill.Draw(r, createTestDrawContext())
	if len(r.pathCalls) != 1 {
		t.Fatalf("path calls = %d, want 1", len(r.pathCalls))
	}
	if got := r.pathCalls[0].paint.Fill.A; got != alpha {
		t.Fatalf("drawn fill alpha = %v, want explicit alpha %v", got, alpha)
	}
}

func TestFillPlotExplicitAlphaOverridesEdgeColorAlpha(t *testing.T) {
	fill := &Fill2D{
		X:         []float64{0, 1, 2},
		Y1:        []float64{1, 2, 1},
		Color:     render.Color{R: 0.3, G: 0.7, B: 0.9, A: 0.7},
		EdgeColor: render.Color{R: 0.1, G: 0.2, B: 0.3, A: 0.25},
		EdgeWidth: 1,
		Alpha:     0.4,
	}

	r := &recordingRenderer{}
	fill.Draw(r, createTestDrawContext())
	if len(r.pathCalls) != 1 {
		t.Fatalf("path calls = %d, want 1", len(r.pathCalls))
	}
	if got := r.pathCalls[0].paint.Stroke.A; got != fill.Alpha {
		t.Fatalf("drawn edge alpha = %v, want explicit alpha override %v", got, fill.Alpha)
	}
}

func TestFillBetweenXPreservesColorAlphaWhenAlphaOmitted(t *testing.T) {
	fig := NewFigure(320, 240)
	ax := fig.AddAxes(unitRect())

	fillColor := render.Color{R: 0.24, G: 0.68, B: 0.54, A: 0.72}
	fill := ax.FillBetweenX(
		[]float64{0, 1, 2},
		[]float64{1, 2, 1},
		[]float64{0, 1, 0},
		FillOptions{Color: &fillColor},
	)
	if fill == nil {
		t.Fatal("expected fill artist")
	}
	if fill.Alpha != 0 {
		t.Fatalf("omitted alpha stored Alpha = %v, want 0 sentinel", fill.Alpha)
	}

	r := &recordingRenderer{}
	fill.Draw(r, createTestDrawContext())
	if len(r.pathCalls) != 1 {
		t.Fatalf("path calls = %d, want 1", len(r.pathCalls))
	}
	if got := r.pathCalls[0].paint.Fill.A; got != fillColor.A {
		t.Fatalf("drawn fill alpha = %v, want color alpha %v", got, fillColor.A)
	}
}

func TestFillBetweenAutoScalesYAndPreservesManualX(t *testing.T) {
	ax := NewFigure(640, 360).AddAxes(geom.Rect{})
	ax.SetXLim(0, 10)

	if _, err := ax.FillBetween(
		[]float64{-100, 100},
		[]float64{-50, 50},
		[]float64{0, 0},
	); err != nil {
		t.Fatalf("FillBetween() returned error: %v", err)
	}

	xMin, xMax := ax.XScale.Domain()
	yMin, yMax := ax.YScale.Domain()
	if xMin != 0 || xMax != 10 {
		t.Fatalf("x limits = [%v, %v], want [0, 10]", xMin, xMax)
	}
	if !floatApprox(yMin, -55, 1e-12) || !floatApprox(yMax, 55, 1e-12) {
		t.Fatalf("y limits = [%v, %v], want [-55, 55]", yMin, yMax)
	}
}

func TestFillBetweenWhereSplitsContiguousRegions(t *testing.T) {
	ax := NewFigure(640, 360).AddAxes(unitRect())
	fill, err := ax.FillBetween(
		[]float64{0, 1, 2, 3, 4},
		[]float64{1, 2, 3, 4, 5},
		[]float64{0, 0, 0, 0, 0},
		FillOptions{Where: []bool{true, true, false, true, true}},
	)
	if err != nil {
		t.Fatalf("FillBetween() returned error: %v", err)
	}
	if fill == nil {
		t.Fatal("FillBetween returned nil")
	}

	r := &recordingRenderer{}
	fill.Draw(r, createTestDrawContext())
	if len(r.pathCalls) != 2 {
		t.Fatalf("path calls = %d, want one polygon per contiguous true region", len(r.pathCalls))
	}
	if got, want := fill.Bounds(nil), (geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 4, Y: 5}}); got != want {
		t.Fatalf("bounds = %+v, want %+v", got, want)
	}
}

func TestFill2DSingleRegionUsesMatplotlibCollectionPath(t *testing.T) {
	fill := &Fill2D{
		X:     []float64{0, 1},
		Y1:    []float64{1, 1},
		Y2:    []float64{0, 0},
		Color: render.Color{A: 1},
	}
	ctx := createTestDrawContext()
	ctx.FigureRect = geom.Rect{Max: geom.Pt{X: 500, Y: 500}}
	r := &batchRecordingRenderer{returnNative: true}

	fill.Draw(r, ctx)

	if len(r.pathCalls) != 0 {
		t.Fatalf("fallback path calls = %d, want native collection path", len(r.pathCalls))
	}
	if len(r.pathCollectionBatches) != 1 {
		t.Fatalf("path collection batches = %d, want Matplotlib FillBetweenPolyCollection draw path", len(r.pathCollectionBatches))
	}
	items := r.pathCollectionBatches[0].Items
	if len(items) != 1 {
		t.Fatalf("path collection items = %d, want one fill_between polygon", len(items))
	}
	want := ctx.DataToPixel.Apply(geom.Pt{X: 0, Y: 0})
	if got := items[0].Path.V[0]; got != want {
		t.Fatalf("single-region first vertex = %+v, want Matplotlib transformed vertex %+v before backend placement", got, want)
	}
}

func TestFill2DMultiRegionUsesMatplotlibGenericCollectionPlacement(t *testing.T) {
	fill := &Fill2D{
		X:     []float64{0, 1, 2, 3, 4},
		Y1:    []float64{1, 1, 1, 1, 1},
		Y2:    []float64{0, 0, 0, 0, 0},
		Where: []bool{true, true, false, true, true},
		Color: render.Color{A: 1},
	}
	ctx := createTestDrawContext()
	ctx.FigureRect = geom.Rect{Max: geom.Pt{X: 500, Y: 500}}
	r := &batchRecordingRenderer{returnNative: true}

	fill.Draw(r, ctx)

	if len(r.pathCalls) != 0 {
		t.Fatalf("fallback path calls = %d, want native collection path", len(r.pathCalls))
	}
	if len(r.pathCollectionBatches) != 1 {
		t.Fatalf("path collection batches = %d, want one FillBetweenPolyCollection-style batch", len(r.pathCollectionBatches))
	}
	items := r.pathCollectionBatches[0].Items
	if len(items) != 2 {
		t.Fatalf("path collection items = %d, want two drawable contiguous regions", len(items))
	}
	want := ctx.DataToPixel.Apply(geom.Pt{X: 0, Y: 0})
	if got := items[0].Path.V[0]; got != want {
		t.Fatalf("multi-region first vertex = %+v, want Matplotlib generic collection placement %+v", got, want)
	}
}

func TestFillBetweenWhereInterpolatesCrossingBoundary(t *testing.T) {
	ax := NewFigure(640, 360).AddAxes(unitRect())
	fill, err := ax.FillBetween(
		[]float64{0, 1, 2},
		[]float64{-1, 1, 1},
		[]float64{0, 0, 0},
		FillOptions{
			Where:       []bool{false, true, true},
			Interpolate: true,
		},
	)
	if err != nil {
		t.Fatalf("FillBetween() returned error: %v", err)
	}
	if fill == nil {
		t.Fatal("FillBetween returned nil")
	}

	if got, want := fill.Bounds(nil), (geom.Rect{Min: geom.Pt{X: 0.5, Y: 0}, Max: geom.Pt{X: 2, Y: 1}}); got != want {
		t.Fatalf("interpolated bounds = %+v, want %+v", got, want)
	}
}

func TestFillBetweenStepPostExpandsRegionSamples(t *testing.T) {
	fill := &Fill2D{
		X:     []float64{0, 1, 2},
		Y1:    []float64{1, 3, 2},
		Y2:    []float64{0, 0, 0},
		Step:  FillStepPost,
		Color: render.Color{A: 1},
	}

	path := fill.createFillPath(3, createTestDrawContext())
	if len(path.V) != 12 {
		t.Fatalf("step-post path vertices = %d, want expanded Matplotlib step polygon vertices", len(path.V))
	}
}

func TestFill2D_EdgeColors(t *testing.T) {
	// Test with edge colors and width
	fill := &Fill2D{
		X:         []float64{0, 1, 2},
		Y1:        []float64{1, 2, 1},
		Color:     render.Color{R: 0.5, G: 0.5, B: 0.5, A: 1},
		EdgeColor: render.Color{R: 1, G: 0, B: 0, A: 1},
		EdgeWidth: 2.0,
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic with edge colors
	fill.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestFill2DDrawUsesMatplotlibFillCollectionSnap(t *testing.T) {
	fill := &Fill2D{
		X:         []float64{0, 1, 2},
		Y1:        []float64{1, 2, 1},
		Y2:        []float64{0, 0, 0},
		Color:     render.Color{A: 1},
		EdgeColor: render.Color{A: 1},
		EdgeWidth: 1,
	}
	r := &recordingRenderer{}

	fill.Draw(r, createTestDrawContext())

	if len(r.pathCalls) != 1 {
		t.Fatalf("path calls = %d, want 1", len(r.pathCalls))
	}
	if got := r.pathCalls[0].paint.Snap; got != render.SnapAuto {
		t.Fatalf("fill paint snap = %v, want Matplotlib fill collection auto snapping", got)
	}
}

func TestFill2DDrawUsesPathCollectionBatchWhenAvailable(t *testing.T) {
	fill := &Fill2D{
		X:         []float64{0, 1, 2, 3, 4},
		Y1:        []float64{1, 2, 1, 2, 1},
		Y2:        []float64{0, 0, 0, 0, 0},
		Where:     []bool{true, true, false, true, true},
		Color:     render.Color{A: 1},
		EdgeColor: render.Color{A: 1},
		EdgeWidth: 2,
	}
	r := &fillPathCollectionRecorder{}

	fill.Draw(r, createTestDrawContext())

	if len(r.batches) != 1 {
		t.Fatalf("path collection batches = %d, want 1", len(r.batches))
	}
	if len(r.pathCalls) != 0 {
		t.Fatalf("fallback path calls = %d, want native path collection", len(r.pathCalls))
	}
	if got := len(r.batches[0].Items); got != 2 {
		t.Fatalf("batch items = %d, want two fill polygons", got)
	}
	if got, want := r.batches[0].Items[0].Paint.LineWidth, (2 * 100.0 / 72.0); got != want {
		t.Fatalf("batch linewidth = %v, want %v", got, want)
	}
}

type fillPathCollectionRecorder struct {
	recordingRenderer
	batches []render.PathCollectionBatch
}

func (r *fillPathCollectionRecorder) DrawPathCollection(batch render.PathCollectionBatch) bool {
	r.batches = append(r.batches, batch)
	return true
}

func TestFill2DDrawLeavesThickFillEdgesUnsnapped(t *testing.T) {
	fill := &Fill2D{
		X:         []float64{0, 1, 2},
		Y1:        []float64{1, 2, 1},
		Y2:        []float64{0, 0, 0},
		Color:     render.Color{A: 1},
		EdgeColor: render.Color{A: 1},
		EdgeWidth: 2,
	}
	r := &recordingRenderer{}

	fill.Draw(r, createTestDrawContext())

	if len(r.pathCalls) != 1 {
		t.Fatalf("path calls = %d, want 1", len(r.pathCalls))
	}
	if got := r.pathCalls[0].paint.Snap; got != render.SnapAuto {
		t.Fatalf("thick fill edge snap = %v, want Matplotlib unsnapped thick fill path", got)
	}
}

func TestFill2DDrawUsesMatplotlibCollectionJoinStyle(t *testing.T) {
	fill := &Fill2D{
		X:         []float64{0, 1, 2},
		Y1:        []float64{1, 2, 1},
		Y2:        []float64{0, 0, 0},
		Color:     render.Color{A: 1},
		EdgeColor: render.Color{A: 1},
		EdgeWidth: 2,
	}
	r := &recordingRenderer{}

	fill.Draw(r, createTestDrawContext())

	if len(r.pathCalls) != 1 {
		t.Fatalf("path calls = %d, want 1", len(r.pathCalls))
	}
	if got := r.pathCalls[0].paint.LineJoin; got != render.JoinRound {
		t.Fatalf("fill edge join = %v, want Matplotlib collection default %v", got, render.JoinRound)
	}
	if got := r.pathCalls[0].paint.LineCap; got != render.CapButt {
		t.Fatalf("fill edge cap = %v, want Matplotlib collection default %v", got, render.CapButt)
	}
}

func TestFill2D_AlphaEdgeCases(t *testing.T) {
	testCases := []struct {
		name  string
		alpha float64
	}{
		{"Zero alpha", 0.0},
		{"Negative alpha", -0.5},
		{"Greater than 1 alpha", 1.5},
		{"Valid alpha", 0.5},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fill := &Fill2D{
				X:         []float64{0, 1, 2},
				Y1:        []float64{1, 2, 1},
				Color:     render.Color{R: 1, G: 0, B: 0, A: 1},
				EdgeColor: render.Color{R: 0, G: 1, B: 0, A: 1},
				EdgeWidth: 1.0,
				Alpha:     tc.alpha,
			}

			renderer := &render.NullRenderer{}
			ctx := createTestDrawContext()

			err := renderer.Begin(geom.Rect{})
			if err != nil {
				t.Fatalf("Failed to begin rendering: %v", err)
			}

			// Should not panic with edge case alpha values
			fill.Draw(renderer, ctx)

			err = renderer.End()
			if err != nil {
				t.Fatalf("Failed to end rendering: %v", err)
			}
		})
	}
}

func TestFill2D_LargeDataset(t *testing.T) {
	// Test performance with many points
	const numPoints = 10000
	x := make([]float64, numPoints)
	y1 := make([]float64, numPoints)
	y2 := make([]float64, numPoints)

	for i := 0; i < numPoints; i++ {
		x[i] = float64(i)
		y1[i] = float64(i % 100)
		y2[i] = float64((i % 50))
	}

	fill := &Fill2D{
		X:         x,
		Y1:        y1,
		Y2:        y2,
		Color:     render.Color{R: 0.3, G: 0.7, B: 0.9, A: 0.6},
		EdgeColor: render.Color{R: 0.1, G: 0.3, B: 0.5, A: 1},
		EdgeWidth: 1.0,
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should handle large dataset without issues
	fill.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}

	// Test bounds calculation with large dataset
	bounds := fill.Bounds(nil)
	if bounds == (geom.Rect{}) {
		t.Error("Expected non-empty bounds for large dataset")
	}

	// Check that bounds are reasonable
	if bounds.Min.X != 0 || bounds.Max.X != float64(numPoints-1) {
		t.Errorf("Expected X bounds [0, %d], got [%v, %v]", numPoints-1, bounds.Min.X, bounds.Max.X)
	}
}

func TestFill2D_NegativeValues(t *testing.T) {
	// Test with negative values
	fill := &Fill2D{
		X:         []float64{-2, -1, 0, 1, 2},
		Y1:        []float64{-1, 2, -0.5, 3, -2},
		Y2:        []float64{-3, -1, -2, 0, -4},
		Color:     render.Color{R: 1, G: 0.5, B: 0.2, A: 0.8},
		EdgeColor: render.Color{R: 0.7, G: 0.3, B: 0.1, A: 1},
		EdgeWidth: 1.5,
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should handle negative values correctly
	fill.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}

	// Test bounds with negative values
	bounds := fill.Bounds(nil)
	if bounds.Min.X != -2 || bounds.Max.X != 2 {
		t.Errorf("Expected X bounds [-2, 2], got [%v, %v]", bounds.Min.X, bounds.Max.X)
	}
	if bounds.Min.Y != -4 || bounds.Max.Y != 3 {
		t.Errorf("Expected Y bounds [-4, 3], got [%v, %v]", bounds.Min.Y, bounds.Max.Y)
	}
}
