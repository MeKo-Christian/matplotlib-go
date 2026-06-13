package core

import (
	"math"
	"reflect"
	"testing"

	matcolor "github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/transform"
)

func TestAxesPColorMeshAndColorbar(t *testing.T) {
	fig := NewFigure(800, 500)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.8, Y: 0.9},
	})

	edgeWidth := 1.0
	mesh := ax.PColorMesh([][]float64{
		{0, 1},
		{2, 3},
	}, MeshOptions{
		XEdges:    []float64{2, 4, 8},
		YEdges:    []float64{-1, 1, 5},
		EdgeWidth: &edgeWidth,
		Label:     "mesh",
	})
	if mesh == nil {
		t.Fatal("expected quad mesh")
	}
	if len(mesh.FaceColors) != 4 {
		t.Fatalf("expected 4 face colors, got %d", len(mesh.FaceColors))
	}
	mapping := mesh.ScalarMap()
	if mapping.Colormap != "viridis" || mapping.VMin != 0 || mapping.VMax != 3 {
		t.Fatalf("unexpected scalar map %+v", mapping)
	}
	bounds := mesh.Bounds(nil)
	if bounds.Min != (geom.Pt{X: 2, Y: -1}) || bounds.Max != (geom.Pt{X: 8, Y: 5}) {
		t.Fatalf("unexpected bounds %+v", bounds)
	}

	cb := fig.AddColorbar(ax, mesh, ColorbarOptions{Label: "density"})
	if cb == nil {
		t.Fatal("expected colorbar axes for mesh")
	}
	yMin, yMax := cb.YScale.Domain()
	if yMin != 0 || yMax != 3 {
		t.Fatalf("unexpected colorbar limits %v..%v", yMin, yMax)
	}
}

func TestAxesPColorFastUsesQuadMeshPath(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	mesh := ax.PColorFast([][]float64{
		{0, 1},
		{2, 3},
	}, MeshOptions{
		XEdges: []float64{-1, 0, 2},
		YEdges: []float64{10, 12, 15},
		Label:  "fast",
	})
	if mesh == nil {
		t.Fatal("expected quad mesh")
	}
	if mesh.Label != "fast" {
		t.Fatalf("mesh label = %q, want fast", mesh.Label)
	}
	if !reflect.DeepEqual(mesh.XEdges, []float64{-1, 0, 2}) {
		t.Fatalf("XEdges = %v", mesh.XEdges)
	}
	if !reflect.DeepEqual(mesh.YEdges, []float64{10, 12, 15}) {
		t.Fatalf("YEdges = %v", mesh.YEdges)
	}
	if len(ax.Artists) != 1 || ax.Artists[0] != mesh {
		t.Fatalf("PColorFast did not add the returned mesh to the axes")
	}
}

func TestAxesPColorUsesUnsnappedMeshEdgesLikeMatplotlib(t *testing.T) {
	ax := &Axes{
		XScale: transform.NewLinear(0, 2),
		YScale: transform.NewLinear(0, 2),
		XAxis:  NewXAxis(),
		YAxis:  NewYAxis(),
	}
	width := 0.75
	edge := render.Color{A: 1}
	mesh := ax.PColor([][]float64{{0, 1}}, MeshOptions{
		XEdges:    []float64{0, 1, 2},
		YEdges:    []float64{0, 1},
		EdgeColor: &edge,
		EdgeWidth: &width,
	})
	if mesh == nil {
		t.Fatal("expected pcolor mesh")
	}

	r := &batchRecordingRenderer{returnNative: true}
	mesh.Draw(r, createTestDrawContext())
	if len(r.quadMeshBatches) != 1 || len(r.quadMeshBatches[0].Cells) == 0 {
		t.Fatalf("expected one native quad mesh batch, got %v", r.quadMeshBatches)
	}
	if got := r.quadMeshBatches[0].Cells[0].Snap; got != render.SnapOff {
		t.Fatalf("pcolor snap = %v, want SnapOff like Matplotlib pcolor", got)
	}
	if !r.quadMeshBatches[0].Cells[0].Antialiased {
		t.Fatal("pcolor antialiasing should follow patch collection defaults when edges are stroked")
	}
}

func TestAxesPColorMeshDisablesAntialiasingLikeMatplotlib(t *testing.T) {
	ax := &Axes{
		XScale: transform.NewLinear(0, 2),
		YScale: transform.NewLinear(0, 2),
		XAxis:  NewXAxis(),
		YAxis:  NewYAxis(),
	}
	width := 0.75
	edge := render.Color{A: 1}
	mesh := ax.PColorMesh([][]float64{{0, 1}}, MeshOptions{
		XEdges:    []float64{0, 1, 2},
		YEdges:    []float64{0, 1},
		EdgeColor: &edge,
		EdgeWidth: &width,
	})
	if mesh == nil {
		t.Fatal("expected pcolormesh mesh")
	}

	r := &batchRecordingRenderer{returnNative: true}
	mesh.Draw(r, createTestDrawContext())
	if len(r.quadMeshBatches) != 1 || len(r.quadMeshBatches[0].Cells) == 0 {
		t.Fatalf("expected one native quad mesh batch, got %v", r.quadMeshBatches)
	}
	if got := r.quadMeshBatches[0].Cells[0].Snap; got != render.SnapOn {
		t.Fatalf("pcolormesh snap = %v, want SnapOn", got)
	}
	if r.quadMeshBatches[0].Cells[0].Antialiased {
		t.Fatal("pcolormesh antialiasing should default off like Matplotlib")
	}
}

func TestPColorMeshShadingAutoUsesCenterCoordinates(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	mesh := ax.PColorMesh([][]float64{
		{0, 1, 2},
		{3, 4, 5},
	}, MeshOptions{
		XEdges: []float64{0, 2, 5},
		YEdges: []float64{10, 16},
	})
	if mesh == nil {
		t.Fatal("expected quad mesh")
	}
	if mesh.Shading != MeshShadingFlat {
		t.Fatalf("mesh shading = %q, want flat after nearest-center expansion", mesh.Shading)
	}
	if want := []float64{-1, 1, 3.5, 6.5}; !reflect.DeepEqual(mesh.XEdges, want) {
		t.Fatalf("XEdges = %v, want %v", mesh.XEdges, want)
	}
	if want := []float64{7, 13, 19}; !reflect.DeepEqual(mesh.YEdges, want) {
		t.Fatalf("YEdges = %v, want %v", mesh.YEdges, want)
	}
}

func TestPColorMeshFlatRejectsCenterCoordinateShape(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	mesh := ax.PColorMesh([][]float64{
		{0, 1},
		{2, 3},
	}, MeshOptions{
		XEdges:  []float64{0, 1},
		YEdges:  []float64{0, 1},
		Shading: MeshShadingFlat,
	})
	if mesh != nil {
		t.Fatalf("expected flat shading to reject center-shaped coordinates, got %+v", mesh)
	}
}

func TestPColorMeshNearestRejectsEdgeCoordinateShape(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	mesh := ax.PColorMesh([][]float64{
		{0, 1},
		{2, 3},
	}, MeshOptions{
		XEdges:  []float64{0, 1, 2},
		YEdges:  []float64{0, 1, 2},
		Shading: MeshShadingNearest,
	})
	if mesh != nil {
		t.Fatalf("expected nearest shading to reject edge-shaped coordinates, got %+v", mesh)
	}
}

func TestPColorMeshGouraudRejectsEdgeCoordinateShape(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	mesh := ax.PColorMesh([][]float64{
		{0, 1},
		{2, 3},
	}, MeshOptions{
		XEdges:  []float64{0, 1, 2},
		YEdges:  []float64{0, 1, 2},
		Shading: MeshShadingGouraud,
	})
	if mesh != nil {
		t.Fatalf("expected Gouraud shading to reject edge-shaped coordinates, got %+v", mesh)
	}
}

func TestPColorMeshGouraudDrawsNativeTriangles(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	cmap := "viridis"

	mesh := ax.PColorMesh([][]float64{
		{0, 1},
		{2, 3},
	}, MeshOptions{
		XEdges:   []float64{0, 2},
		YEdges:   []float64{0, 3},
		Shading:  MeshShadingGouraud,
		Colormap: &cmap,
	})
	if mesh == nil {
		t.Fatal("expected gouraud mesh")
	}
	r := &batchRecordingRenderer{returnNative: true}
	mesh.Draw(r, createTestDrawContext())
	if len(r.gouraudBatches) != 1 {
		t.Fatalf("gouraud batches = %d, want 1", len(r.gouraudBatches))
	}
	triangles := r.gouraudBatches[0].Triangles
	if got := len(triangles); got != 4 {
		t.Fatalf("gouraud triangles = %d, want 4 Matplotlib center-fan triangles", got)
	}
	center := createTestDrawContext().DataToPixel.Apply(geom.Pt{X: 1, Y: 1.5})
	cornerColors := meshValueColors(mesh.Values, mesh.ScalarMap().Resolved(), mesh.alphaValue())
	wantCenter := averageColor4(cornerColors[0][0], cornerColors[0][1], cornerColors[1][1], cornerColors[1][0])
	for i, tri := range triangles {
		if !sameContourPoint(tri.P[2], center) {
			t.Fatalf("gouraud triangle %d center = %+v, want %+v", i, tri.P[2], center)
		}
		if !colorsApproxEqual(tri.Color[2], wantCenter, 1e-12) {
			t.Fatalf("gouraud triangle %d center color = %+v, want averaged corner color %+v", i, tri.Color[2], wantCenter)
		}
	}
	if len(r.quadMeshBatches) != 0 || len(r.pathCalls) != 0 {
		t.Fatalf("expected native gouraud only, quad batches=%d path calls=%d", len(r.quadMeshBatches), len(r.pathCalls))
	}
}

func colorsApproxEqual(a, b render.Color, tol float64) bool {
	return math.Abs(a.R-b.R) <= tol &&
		math.Abs(a.G-b.G) <= tol &&
		math.Abs(a.B-b.B) <= tol &&
		math.Abs(a.A-b.A) <= tol
}

func TestPColorMeshBadCellsAreTransparent(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	mesh := ax.PColorMesh([][]float64{
		{0, math.NaN()},
		{math.Inf(1), 3},
	})
	if mesh == nil {
		t.Fatal("expected quad mesh")
	}
	if mesh.FaceColors[1].A != 0 || mesh.FaceColors[2].A != 0 {
		t.Fatalf("bad cells should be transparent, got %+v", mesh.FaceColors)
	}
	mapping := mesh.ScalarMap()
	if mapping.VMin != 0 || mapping.VMax != 3 {
		t.Fatalf("bad cells should not affect scalar range, got %+v", mapping)
	}
}

func TestPColorMeshUsesBadUnderAndOverColormapColors(t *testing.T) {
	bad := render.Color{R: 0.55, G: 0.55, B: 0.55, A: 0.8}
	under := render.Color{R: 0.10, G: 0.25, B: 0.90, A: 1}
	over := render.Color{R: 0.90, G: 0.25, B: 0.10, A: 1}
	cmapName := "bounded mesh fixture"
	matcolor.RegisterColormap(cmapName, matcolor.NewColormap(cmapName, []matcolor.ColorStop{
		{Pos: 0, Color: render.Color{R: 0, G: 0, B: 0, A: 1}},
		{Pos: 1, Color: render.Color{R: 1, G: 1, B: 1, A: 1}},
	}).WithBad(bad).WithUnder(under).WithOver(over))

	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	vmin, vmax := 0.0, 1.0
	mesh := ax.PColorMesh([][]float64{
		{math.NaN(), -0.25},
		{0.5, 1.25},
	}, MeshOptions{
		Colormap: &cmapName,
		VMin:     &vmin,
		VMax:     &vmax,
	})
	if mesh == nil {
		t.Fatal("expected quad mesh")
	}
	if got := mesh.FaceColors[0]; got != bad {
		t.Fatalf("bad cell color = %#v, want %#v", got, bad)
	}
	if got := mesh.FaceColors[1]; got != under {
		t.Fatalf("under cell color = %#v, want %#v", got, under)
	}
	if got := mesh.FaceColors[3]; got != over {
		t.Fatalf("over cell color = %#v, want %#v", got, over)
	}
}

func TestPColorMeshUsesConfiguredNorm(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	cmap := "gray"
	mesh := ax.PColorMesh([][]float64{
		{1, 10, 100},
	}, MeshOptions{
		Colormap: &cmap,
		Norm:     LogNorm{VMin: 1, VMax: 100},
	})
	if mesh == nil {
		t.Fatal("expected quad mesh")
	}
	if mesh.Norm == nil || mesh.Norm.NormName() != "log" {
		t.Fatalf("mesh norm = %#v, want log norm", mesh.Norm)
	}
	if got := mesh.FaceColors[1].R; got < 0.49 || got > 0.51 {
		t.Fatalf("middle log-normalized face red = %v, want about 0.5", got)
	}
}

func TestPColorMeshMaskUsesBadColorAndExcludesScalarRange(t *testing.T) {
	bad := render.Color{R: 0.55, G: 0.55, B: 0.55, A: 0.8}
	cmapName := "masked mesh fixture"
	matcolor.RegisterColormap(cmapName, matcolor.NewColormap(cmapName, []matcolor.ColorStop{
		{Pos: 0, Color: render.Color{R: 0, G: 0, B: 0, A: 1}},
		{Pos: 1, Color: render.Color{R: 1, G: 1, B: 1, A: 1}},
	}).WithBad(bad))

	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	mesh := ax.PColorMesh([][]float64{
		{100, 2},
		{3, 4},
	}, MeshOptions{
		Colormap: &cmapName,
		Mask: [][]bool{
			{true, false},
			{false, false},
		},
	})
	if mesh == nil {
		t.Fatal("expected quad mesh")
	}
	if got := mesh.FaceColors[0]; got != bad {
		t.Fatalf("masked cell color = %#v, want bad color %#v", got, bad)
	}
	mapping := mesh.ScalarMap()
	if mapping.VMin != 2 || mapping.VMax != 4 {
		t.Fatalf("masked value should not affect scalar range, got %+v", mapping)
	}
}

func TestAxesHist2DCounts(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	result := ax.Hist2D(
		[]float64{0.1, 0.2, 0.8, 0.9},
		[]float64{0.1, 0.2, 0.8, 0.9},
		Hist2DOptions{
			XBinEdges: []float64{0, 0.5, 1},
			YBinEdges: []float64{0, 0.5, 1},
			Label:     "hist2d",
		},
	)
	if result == nil || result.Mesh == nil {
		t.Fatal("expected hist2d result mesh")
	}
	if got := result.Counts[0][0]; got != 2 {
		t.Fatalf("lower-left count = %v, want 2", got)
	}
	if got := result.Counts[1][1]; got != 2 {
		t.Fatalf("upper-right count = %v, want 2", got)
	}
	if got := result.Counts[0][1] + result.Counts[1][0]; got != 0 {
		t.Fatalf("unexpected off-diagonal counts %v", got)
	}
}

func TestAxesHist2DWeightsAndDensity(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	result := ax.Hist2D(
		[]float64{0.25, 0.75, 1.5},
		[]float64{0.25, 0.75, 0.5},
		Hist2DOptions{
			XBinEdges: []float64{0, 1, 2},
			YBinEdges: []float64{0, 1},
			Weights:   []float64{2, 1, 3},
			Norm:      HistNormDensity,
		},
	)
	if result == nil || result.Mesh == nil {
		t.Fatal("expected hist2d result mesh")
	}
	if got, want := result.Counts[0][0], 0.5; got != want {
		t.Fatalf("density count[0][0] = %v, want %v", got, want)
	}
	if got, want := result.Counts[0][1], 0.5; got != want {
		t.Fatalf("density count[0][1] = %v, want %v", got, want)
	}
}

func TestAxesContourAndContourf(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	grid := [][]float64{
		{0, 1, 2},
		{1, 2, 3},
		{2, 3, 4},
	}

	contours := ax.Contour(grid, ContourOptions{
		Levels:     []float64{1.5, 2.5},
		LabelLines: true,
		Label:      "contour",
	})
	if contours == nil || contours.Lines == nil {
		t.Fatal("expected contour lines")
	}
	if len(contours.labels) != 2 {
		t.Fatalf("expected one label per contour level, got %d", len(contours.labels))
	}
	if got, want := contours.Z(), defaultLineZ; got != want {
		t.Fatalf("contour default z = %v, want line z %v so isolines draw above patch collections", got, want)
	}

	filled := ax.Contourf(grid, ContourOptions{
		Levels: []float64{0, 1, 2, 3, 4},
		Label:  "filled",
	})
	if filled == nil || filled.Fills == nil {
		t.Fatal("expected filled contours")
	}
	mapping := filled.ScalarMap()
	if mapping.VMin != 0 || mapping.VMax != 4 {
		t.Fatalf("unexpected filled contour scalar map %+v", mapping)
	}
}

func TestContourInlineLabelsMatchMatplotlibMeshFixturePositions(t *testing.T) {
	fig := NewFigure(980, 620)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.57, Y: 0.57}, Max: geom.Pt{X: 0.96, Y: 0.93}})
	ax.SetXLim(0, 4)
	ax.SetYLim(0, 4)
	data := [][]float64{
		{0.0, 0.4, 0.8, 0.4, 0.0},
		{0.2, 0.8, 1.3, 0.8, 0.2},
		{0.3, 1.0, 1.7, 1.0, 0.3},
		{0.2, 0.8, 1.3, 0.8, 0.2},
		{0.0, 0.4, 0.8, 0.4, 0.0},
	}

	contours := ax.Contour(data, ContourOptions{
		Levels:     []float64{0.4, 0.8, 1.2, 1.6},
		LabelLines: true,
	})
	if contours == nil || contours.Lines == nil {
		t.Fatal("expected contour lines")
	}
	ctx := AxesDrawContext(ax, fig)
	_, _, _, labels := contourInlineLabelSegmentsForLevels(
		contours.Lines,
		contours.lineLevels,
		nil,
		contours.LabelFormatter,
		10,
		5,
		&recordingRenderer{},
		ctx,
	)

	want := map[float64]geom.Pt{
		0.8: {X: 2, Y: 4},
		1.2: {X: 2, Y: 3.2},
		1.6: {X: 2, Y: 2.25},
	}
	got := map[float64]geom.Pt{}
	for _, label := range labels {
		if _, ok := want[label.Level]; ok {
			got[label.Level] = label.Position
		}
	}
	for level, wantPos := range want {
		gotPos, ok := got[level]
		if !ok {
			t.Fatalf("missing inline contour label for level %v in %+v", level, labels)
		}
		if !pointsApprox(gotPos, wantPos, 1e-9) {
			t.Fatalf("inline contour label level %v position = %+v, want Matplotlib %+v (all labels %+v)", level, gotPos, wantPos, labels)
		}
	}
}

func TestStructuredContourOpenBoundaryPathsMatchContourpyOrder(t *testing.T) {
	const (
		rows = 8
		cols = 10
	)
	data := make([][]float64, rows)
	for y := 0; y < rows; y++ {
		data[y] = make([]float64, cols)
		yy := float64(y) / float64(rows-1)
		for x := 0; x < cols; x++ {
			xx := float64(x) / float64(cols-1)
			data[y][x] = 0.55 + 0.25*math.Sin((xx*2.3+0.35)*math.Pi) + 0.20*math.Cos((yy*2.8-0.35*0.4)*math.Pi)
		}
	}

	x := []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	y := []float64{0, 1, 2, 3, 4, 5, 6, 7}
	polylines, levels := contourGridPolylines(x, y, data, []float64{0.6})

	got := [][]geom.Pt{}
	for i, level := range levels {
		if math.Abs(level-0.6) <= 1e-9 {
			got = append(got, polylines[i])
		}
	}
	want := [][]geom.Pt{
		{
			{X: 3.2944010458613353, Y: 0},
			{X: 3, Y: 0.939110067830601},
			{X: 2.9862606219581256, Y: 1},
			{X: 2, Y: 1.825533129695295},
			{X: 1.6886888892082657, Y: 2},
			{X: 1, Y: 2.899329368532861},
			{X: 0, Y: 2.763194516755538},
		},
		{
			{X: 9, Y: 2.763194516755539},
			{X: 8, Y: 2.89932936853286},
			{X: 7.311311110791736, Y: 2},
			{X: 7, Y: 1.8255331296952941},
			{X: 6.013739378041875, Y: 1},
			{X: 6, Y: 0.9391100678305987},
			{X: 5.705598954138664, Y: 0},
		},
		{
			{X: 0, Y: 3.1383144172489397},
			{X: 1, Y: 3.0588001575583688},
			{X: 2, Y: 3.8215311550581923},
			{X: 2.1566630410641783, Y: 4},
			{X: 3, Y: 4.799189389904025},
			{X: 3.2944010458613353, Y: 5},
			{X: 3, Y: 5.939110067830604},
			{X: 2.986260621958126, Y: 6},
			{X: 2, Y: 6.8255331296952955},
			{X: 1.6886888892082672, Y: 7},
		},
		{
			{X: 7.311311110791734, Y: 7},
			{X: 7, Y: 6.825533129695295},
			{X: 6.0137393780418735, Y: 6},
			{X: 6, Y: 5.939110067830603},
			{X: 5.705598954138665, Y: 5},
			{X: 6, Y: 4.799189389904026},
			{X: 6.843336958935823, Y: 4},
			{X: 7, Y: 3.8215311550581936},
			{X: 8, Y: 3.0588001575583696},
			{X: 9, Y: 3.138314417248939},
		},
	}
	if len(got) != len(want) {
		t.Fatalf("level 0.6 polylines = %d, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if len(got[i]) != len(want[i]) {
			t.Fatalf("polyline %d length = %d, want %d: %+v", i, len(got[i]), len(want[i]), got[i])
		}
		for j := range want[i] {
			if !pointsApprox(got[i][j], want[i][j], 1e-9) {
				t.Fatalf("polyline %d point %d = %+v, want contourpy %+v (polyline %+v)", i, j, got[i][j], want[i][j], got[i])
			}
		}
	}
}

func TestStructuredContourClosedPathStartMatchesContourpyOrder(t *testing.T) {
	const (
		rows = 8
		cols = 10
	)
	data := make([][]float64, rows)
	for y := 0; y < rows; y++ {
		data[y] = make([]float64, cols)
		yy := float64(y) / float64(rows-1)
		for x := 0; x < cols; x++ {
			xx := float64(x) / float64(cols-1)
			data[y][x] = 0.55 + 0.25*math.Sin((xx*2.3+0.35)*math.Pi) + 0.20*math.Cos((yy*2.8-0.35*0.4)*math.Pi)
		}
	}

	x := []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	y := []float64{0, 1, 2, 3, 4, 5, 6, 7}
	tests := []struct {
		name  string
		level float64
		want  []geom.Pt
	}{
		{
			name:  "small loop",
			level: 0.15,
			want: []geom.Pt{
				{X: 4, Y: 2.734446720898368},
				{X: 3.81083871078257, Y: 3},
				{X: 4, Y: 3.1551055422834224},
				{X: 5, Y: 3.1551055422834224},
				{X: 5.18916128921743, Y: 3},
				{X: 5, Y: 2.7344467208983674},
				{X: 4, Y: 2.734446720898368},
			},
		},
		{
			name:  "mid loop",
			level: 0.45,
			want: []geom.Pt{
				{X: 3, Y: 1.6315577244440652},
				{X: 2.5598249700346554, Y: 2},
				{X: 2.047105137818292, Y: 3},
				{X: 2.9249223106066436, Y: 4},
				{X: 3, Y: 4.071147346184701},
				{X: 4, Y: 4.753246180135713},
				{X: 5, Y: 4.753246180135713},
				{X: 6, Y: 4.071147346184701},
				{X: 6.075077689393357, Y: 4},
				{X: 6.952894862181708, Y: 3},
				{X: 6.440175029965345, Y: 2},
				{X: 6, Y: 1.6315577244440647},
				{X: 5, Y: 1.0290800106575034},
				{X: 4, Y: 1.0290800106575034},
				{X: 3, Y: 1.6315577244440652},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			polylines, levels := contourGridPolylines(x, y, data, []float64{tc.level})
			var got []geom.Pt
			for i, level := range levels {
				if math.Abs(level-tc.level) <= 1e-9 && contourPolylineClosed(polylines[i]) {
					got = polylines[i]
					break
				}
			}
			if len(got) != len(tc.want) {
				t.Fatalf("level %v closed path length = %d, want %d: %+v", tc.level, len(got), len(tc.want), got)
			}
			for i := range tc.want {
				if !pointsApprox(got[i], tc.want[i], 1e-6) {
					t.Fatalf("level %v closed path point %d = %+v, want contourpy %+v (path %+v)", tc.level, i, got[i], tc.want[i], got)
				}
			}
		})
	}
}

func TestContourInlineLabelsMatchMatplotlibArraysShowcaseLevel06(t *testing.T) {
	const (
		rows = 8
		cols = 10
	)
	data := make([][]float64, rows)
	for y := 0; y < rows; y++ {
		data[y] = make([]float64, cols)
		yy := float64(y) / float64(rows-1)
		for x := 0; x < cols; x++ {
			xx := float64(x) / float64(cols-1)
			data[y][x] = 0.55 + 0.25*math.Sin((xx*2.3+0.35)*math.Pi) + 0.20*math.Cos((yy*2.8-0.35*0.4)*math.Pi)
		}
	}

	fig := NewFigure(1240, 620)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.37, Y: 0.14}, Max: geom.Pt{X: 0.63, Y: 0.88}})
	ax.SetXLim(0, cols)
	ax.SetYLim(0, rows)
	contours := ax.Contour(data, ContourOptions{
		Levels:     []float64{0.6},
		X:          []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		Y:          []float64{0, 1, 2, 3, 4, 5, 6, 7},
		LabelLines: true,
		LabelFormatter: FormatStrFormatter{
			Pattern: "%.3g",
		},
	})
	if contours == nil || contours.Lines == nil {
		t.Fatal("expected contour lines")
	}

	ctx := AxesDrawContext(ax, fig)
	_, _, _, labels := contourInlineLabelSegmentsForLevels(
		contours.Lines,
		contours.lineLevels,
		nil,
		contours.LabelFormatter,
		10,
		5,
		&recordingRenderer{},
		ctx,
	)

	want := []geom.Pt{
		{X: 2, Y: 1.825533129695295},
		{X: 7, Y: 1.8255331296952941},
		{X: 3.2944010458613353, Y: 5},
		{X: 6, Y: 4.799189389904026},
	}
	got := []geom.Pt{}
	for _, label := range labels {
		got = append(got, label.Position)
	}
	if len(got) != len(want) {
		t.Fatalf("level 0.6 label positions = %+v, want %+v", got, want)
	}
	for i, wantPos := range want {
		if !pointsApprox(got[i], wantPos, 1e-9) {
			t.Fatalf("level 0.6 label %d = %+v, want Matplotlib %+v (all labels %+v)", i, got[i], wantPos, labels)
		}
	}
}

func TestContourInlineLabelsMatchMatplotlibArraysShowcaseAllLevels(t *testing.T) {
	const (
		rows = 8
		cols = 10
	)
	data := make([][]float64, rows)
	for y := 0; y < rows; y++ {
		data[y] = make([]float64, cols)
		yy := float64(y) / float64(rows-1)
		for x := 0; x < cols; x++ {
			xx := float64(x) / float64(cols-1)
			data[y][x] = 0.55 + 0.25*math.Sin((xx*2.3+0.35)*math.Pi) + 0.20*math.Cos((yy*2.8-0.35*0.4)*math.Pi)
		}
	}

	fig := NewFigure(1240, 620)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.37, Y: 0.14}, Max: geom.Pt{X: 0.63, Y: 0.88}})
	ax.SetXLim(0, cols)
	ax.SetYLim(0, rows)
	contours := ax.Contour(data, ContourOptions{
		LevelCount: 6,
		X:          []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		Y:          []float64{0, 1, 2, 3, 4, 5, 6, 7},
		LabelLines: true,
	})
	if contours == nil || contours.Lines == nil {
		t.Fatal("expected contour lines")
	}

	ctx := AxesDrawContext(ax, fig)
	_, _, _, labels := contourInlineLabelSegmentsForLevels(
		contours.Lines,
		contours.lineLevels,
		nil,
		contours.LabelFormatter,
		10,
		5,
		&arraysShowcaseContourTextMetricRenderer{},
		ctx,
	)

	want := []contourLabel{
		{Text: "0.15", Level: 0.15, Position: geom.Pt{X: 5, Y: 3.1551055422834224}, Angle: -0.587233229688272},
		{Text: "0.3", Level: 0.30, Position: geom.Pt{X: 4, Y: 6.6721388277667995}, Angle: -0.410000655682552},
		{Text: "0.3", Level: 0.30, Position: geom.Pt{X: 5, Y: 4.02520413641639}, Angle: -0.477202539043145},
		{Text: "0.45", Level: 0.45, Position: geom.Pt{X: 4, Y: 6.029080010657504}, Angle: -0.410000655682552},
		{Text: "0.45", Level: 0.45, Position: geom.Pt{X: 6, Y: 4.071147346184701}, Angle: -0.984978996134542},
		{Text: "0.6", Level: 0.60, Position: geom.Pt{X: 2, Y: 1.825533129695295}, Angle: -0.881614573170370},
		{Text: "0.6", Level: 0.60, Position: geom.Pt{X: 7, Y: 1.8255331296952941}, Angle: 0.881614573170361},
		{Text: "0.6", Level: 0.60, Position: geom.Pt{X: 3.2944010458613353, Y: 5}, Angle: 1.313365091732432},
		{Text: "0.6", Level: 0.60, Position: geom.Pt{X: 6, Y: 4.799189389904026}, Angle: -0.958436492122234},
		{Text: "0.75", Level: 0.75, Position: geom.Pt{X: 2, Y: 1.1824743125859998}, Angle: -0.920842971683705},
		{Text: "0.75", Level: 0.75, Position: geom.Pt{X: 7, Y: 1.1824743125859987}, Angle: 0.920842971683706},
		{Text: "0.75", Level: 0.75, Position: geom.Pt{X: 2.443644072976639, Y: 5}, Angle: 1.366175082469990},
		{Text: "0.75", Level: 0.75, Position: geom.Pt{X: 7, Y: 4.579580104654563}, Angle: -0.940572796675585},
		{Text: "0.9", Level: 0.90, Position: geom.Pt{X: 1, Y: 1.0998415909700405}, Angle: -0.407194023520089},
		{Text: "0.9", Level: 0.90, Position: geom.Pt{X: 7.821846998608887, Y: 1}, Angle: 0.989551006032252},
		{Text: "0.9", Level: 0.90, Position: geom.Pt{X: 1.1781530013911132, Y: 6}, Angle: -0.989551006032250},
		{Text: "0.9", Level: 0.90, Position: geom.Pt{X: 7.484834469529565, Y: 5}, Angle: -1.301899328704219},
	}
	if len(labels) != len(want) {
		t.Fatalf("arrays contour labels = %d, want %d: %+v", len(labels), len(want), labels)
	}
	for i, wantLabel := range want {
		if labels[i].Text != wantLabel.Text || math.Abs(labels[i].Level-wantLabel.Level) > 1e-9 || !pointsApprox(labels[i].Position, wantLabel.Position, 1e-5) || !approx(labels[i].Angle, wantLabel.Angle, 1e-12) {
			t.Fatalf("arrays contour label %d = %q level %.12g at %+v angle %.15g, want %q level %.12g at %+v angle %.15g (all labels %+v)", i, labels[i].Text, labels[i].Level, labels[i].Position, labels[i].Angle, wantLabel.Text, wantLabel.Level, wantLabel.Position, wantLabel.Angle, labels)
		}
	}
}

func TestContourfUsesStructuredGridBandPolygons(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(unitRect())
	data := [][]float64{
		{0.0, 0.4, 0.8, 0.4, 0.0},
		{0.2, 0.8, 1.3, 0.8, 0.2},
		{0.3, 1.0, 1.7, 1.0, 0.3},
		{0.2, 0.8, 1.3, 0.8, 0.2},
		{0.0, 0.4, 0.8, 0.4, 0.0},
	}
	levels := []float64{0.2, 0.6, 1.0, 1.4, 1.8}

	filled := ax.Contourf(data, ContourOptions{Levels: levels})
	if filled == nil || filled.Fills == nil {
		t.Fatal("expected structured contourf fills")
	}
	x, y, values, ok := contourGridCoordsValues(data, []ContourOptions{{Levels: levels}})
	if !ok {
		t.Fatal("expected valid grid contour coordinates")
	}
	mapping, err := ResolveScalarMapValues(values, ScalarMapConfig{})
	if err != nil {
		t.Fatalf("resolve scalar map: %v", err)
	}
	mapping.Norm = Normalize{VMin: levels[0], VMax: levels[len(levels)-1]}
	mapping.VMin = levels[0]
	mapping.VMax = levels[len(levels)-1]
	wantPolygons, _ := contourGridBandPolygons(x, y, data, levels, ContourOptions{Levels: levels}, mapping, 1)
	if got, want := len(filled.Fills.Polygons), len(wantPolygons); got != want {
		t.Fatalf("contourf polygon count = %d, want structured grid count %d", got, want)
	}
}

func TestContourLevelsUseNiceLocatorForImplicitCounts(t *testing.T) {
	levels := contourLevels([]float64{0.287, 1.0}, nil, 6, false)
	want := []float64{0.15, 0.3, 0.45, 0.6, 0.75, 0.9, 1.05}
	if len(levels) != len(want) {
		t.Fatalf("levels = %v, want %v", levels, want)
	}
	for i := range want {
		if !approx(levels[i], want[i], 1e-12) {
			t.Fatalf("levels = %v, want %v", levels, want)
		}
	}

	filled := contourLevels([]float64{0.287, 1.0}, nil, 6, true)
	if len(filled) < 2 || filled[0] > 0.287 || filled[len(filled)-1] < 1.0 {
		t.Fatalf("filled levels should cover data range, got %v", filled)
	}
}

func TestStructuredContourBandClipsSingleQuadLikeMatplotlib(t *testing.T) {
	grid := [][]float64{
		{0, 1},
		{1, 2},
	}
	levels := []float64{0.5, 1.5}
	mapping := ScalarMapInfo{Colormap: "viridis", VMin: 0.5, VMax: 1.5}

	polygons, _ := contourGridBandPolygons(
		[]float64{0, 1},
		[]float64{0, 1},
		grid,
		levels,
		ContourOptions{},
		mapping,
		1,
	)
	if got, want := len(polygons), 1; got != want {
		t.Fatalf("structured contour band polygons = %d, want one Matplotlib quad path", got)
	}
	want := []geom.Pt{
		{X: 1, Y: 0},
		{X: 1, Y: 0.5},
		{X: 0.5, Y: 1},
		{X: 0, Y: 1},
		{X: 0, Y: 0.5},
		{X: 0.5, Y: 0},
	}
	if !pointsEqual(polygons[0], want, 1e-12) {
		t.Fatalf("structured contour band polygon = %+v, want Matplotlib path %+v", polygons[0], want)
	}
}

func TestStructuredContourLineClipsSingleSaddleQuadLikeMatplotlib(t *testing.T) {
	grid := [][]float64{
		{0, 1},
		{1, 0},
	}
	polylines, levels := contourGridPolylines([]float64{0, 1}, []float64{0, 1}, grid, []float64{0.5})
	if got, want := len(polylines), 1; got != want {
		t.Fatalf("structured contour polylines = %d, want one Matplotlib saddle path", got)
	}
	if got, want := len(levels), 1; got != want || levels[0] != 0.5 {
		t.Fatalf("structured contour levels = %v, want [0.5]", levels)
	}
	want := []geom.Pt{
		{X: 0, Y: 0.5},
		{X: 0.5, Y: 1},
		{X: 1, Y: 0.5},
		{X: 0.5, Y: 0},
	}
	if !pointsEqual(polylines[0], want, 1e-12) {
		t.Fatalf("structured saddle contour = %+v, want Matplotlib path %+v", polylines[0], want)
	}
}

func TestAxesContourUsesStructuredGridLinesLikeMatplotlib(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(unitRect())
	grid := [][]float64{
		{0, 1},
		{1, 0},
	}

	contours := ax.Contour(grid, ContourOptions{Levels: []float64{0.5}})
	if contours == nil || contours.Lines == nil {
		t.Fatal("expected contour lines")
	}
	if got, want := len(contours.Lines.Segments), 1; got != want {
		t.Fatalf("public contour segments = %d, want one structured saddle path: %+v", got, contours.Lines.Segments)
	}
	want := []geom.Pt{
		{X: 0, Y: 0.5},
		{X: 0.5, Y: 1},
		{X: 1, Y: 0.5},
		{X: 0.5, Y: 0},
	}
	if !pointsEqual(contours.Lines.Segments[0], want, 1e-12) {
		t.Fatalf("public structured contour = %+v, want Matplotlib path %+v", contours.Lines.Segments[0], want)
	}
}

func TestStructuredContourBandSplitsSaddleQuadLikeMatplotlib(t *testing.T) {
	grid := [][]float64{
		{0, 1},
		{1, 0},
	}
	levels := []float64{0.5, 1.5}
	mapping := ScalarMapInfo{Colormap: "viridis", VMin: 0.5, VMax: 1.5}

	polygons, _ := contourGridBandPolygons(
		[]float64{0, 1},
		[]float64{0, 1},
		grid,
		levels,
		ContourOptions{},
		mapping,
		1,
	)
	if got, want := len(polygons), 2; got != want {
		t.Fatalf("structured saddle band polygons = %d, want two Matplotlib triangles: %+v", got, polygons)
	}
	want := [][]geom.Pt{
		{
			{X: 1, Y: 0},
			{X: 1, Y: 0.5},
			{X: 0.5, Y: 0},
			{X: 1, Y: 0},
		},
		{
			{X: 0.5, Y: 1},
			{X: 0, Y: 1},
			{X: 0, Y: 0.5},
			{X: 0.5, Y: 1},
		},
	}
	for i := range want {
		if !pointsEqual(polygons[i], want[i], 1e-12) {
			t.Fatalf("structured saddle polygon %d = %+v, want %+v", i, polygons[i], want[i])
		}
	}
}

func TestStructuredContourBandTouchesBoundaryLikeMatplotlib(t *testing.T) {
	grid := [][]float64{
		{0, 1},
		{1, 2},
	}
	levels := []float64{0, 1}
	mapping := ScalarMapInfo{Colormap: "viridis", VMin: 0, VMax: 1}

	polygons, _ := contourGridBandPolygons(
		[]float64{0, 1},
		[]float64{0, 1},
		grid,
		levels,
		ContourOptions{},
		mapping,
		1,
	)
	if got, want := len(polygons), 1; got != want {
		t.Fatalf("boundary-touching band polygons = %d, want one Matplotlib boundary path: %+v", got, polygons)
	}
	want := []geom.Pt{
		{X: 1, Y: 0},
		{X: 1, Y: 0},
		{X: 0, Y: 1},
		{X: 0, Y: 1},
		{X: 0, Y: 0},
		{X: 1, Y: 0},
	}
	if !pointsEqual(polygons[0], want, 1e-12) {
		t.Fatalf("boundary-touching band polygon = %+v, want Matplotlib path %+v", polygons[0], want)
	}
}

func TestStructuredContourBandLeavesInteriorHoleLikeMatplotlib(t *testing.T) {
	grid := [][]float64{
		{1, 1, 1},
		{1, 0, 1},
		{1, 1, 1},
	}
	levels := []float64{0.5, 1.5}
	mapping := ScalarMapInfo{Colormap: "viridis", VMin: 0.5, VMax: 1.5}

	polygons, _ := contourGridBandPolygons(
		[]float64{0, 1, 2},
		[]float64{0, 1, 2},
		grid,
		levels,
		ContourOptions{},
		mapping,
		1,
	)
	if got, want := len(polygons), 4; got != want {
		t.Fatalf("hole band polygons = %d, want four clipped cell polygons: %+v", got, polygons)
	}
	center := geom.Pt{X: 1, Y: 1}
	for _, polygon := range polygons {
		if pointInPolygon(center, polygon) {
			t.Fatalf("interior hole center was filled by polygon: %+v", polygon)
		}
	}
	holeBoundary := []geom.Pt{
		{X: 1, Y: 0.5},
		{X: 1.5, Y: 1},
		{X: 1, Y: 1.5},
		{X: 0.5, Y: 1},
	}
	for _, want := range holeBoundary {
		if !contourPolygonsContainPoint(polygons, want) {
			t.Fatalf("hole boundary point %+v missing from polygons: %+v", want, polygons)
		}
	}
}

func TestTriContourfSkipsMaskedTriangles(t *testing.T) {
	tri := Triangulation{
		X:         []float64{0, 1, 0, 1},
		Y:         []float64{0, 0, 1, 1},
		Triangles: [][3]int{{0, 1, 2}, {1, 3, 2}},
		Mask:      []bool{true, false},
	}
	values := []float64{0, 1, 1, 2}
	polygons, _ := contourBandPolygons(
		tri,
		values,
		[]float64{0.5, 1.5},
		ContourOptions{},
		ScalarMapInfo{Colormap: "viridis", VMin: 0.5, VMax: 1.5},
		1,
	)
	if got, want := len(polygons), 1; got != want {
		t.Fatalf("masked tricontourf polygons = %d, want only unmasked triangle contribution: %+v", got, polygons)
	}
	for _, pt := range polygons[0] {
		if pt == (geom.Pt{X: 0, Y: 0}) {
			t.Fatalf("masked triangle-only point leaked into polygon: %+v", polygons[0])
		}
	}
}

func contourPolygonsContainPoint(polygons [][]geom.Pt, want geom.Pt) bool {
	for _, polygon := range polygons {
		for _, got := range polygon {
			if sameContourPoint(got, want) {
				return true
			}
		}
	}
	return false
}

func TestTriangulationArtists(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	tri := Triangulation{
		X:         []float64{0, 1, 0, 1},
		Y:         []float64{0, 0, 1, 1},
		Triangles: [][3]int{{0, 1, 3}, {0, 3, 2}},
	}

	lineWidth := 2.0
	plot := ax.TriPlot(tri, TriPlotOptions{LineWidth: &lineWidth, Label: "mesh"})
	if plot == nil {
		t.Fatal("expected triplot collection")
	}
	if len(plot.Segments) != 5 {
		t.Fatalf("expected 5 unique edges, got %d", len(plot.Segments))
	}

	colorMesh := ax.TriColor(tri, []float64{0, 1, 2, 3}, TriColorOptions{Label: "tripcolor"})
	if colorMesh == nil {
		t.Fatal("expected tripcolor collection")
	}
	if len(colorMesh.Polygons) != 2 || len(colorMesh.FaceColors) != 2 {
		t.Fatalf("unexpected tripcolor polygon/color counts: %d / %d", len(colorMesh.Polygons), len(colorMesh.FaceColors))
	}

	contours := ax.TriContour(tri, []float64{0, 1, 2, 3}, ContourOptions{Levels: []float64{1.5}})
	if contours == nil || contours.Lines == nil {
		t.Fatal("expected tricontour lines")
	}
	if got, want := contours.Z(), defaultLineZ; got != want {
		t.Fatalf("tricontour default z = %v, want line z %v so isolines draw above filled collections", got, want)
	}

	filled := ax.TriContourf(tri, []float64{0, 1, 2, 3}, ContourOptions{Levels: []float64{0, 1, 2, 3}})
	if filled == nil || filled.Fills == nil {
		t.Fatal("expected tricontourf fills")
	}
	if got, want := filled.Z(), defaultPatchZ; got != want {
		t.Fatalf("tricontourf default z = %v, want patch z %v", got, want)
	}
}

func TestTriContourLevelCountKeepsMatplotlibLocatorBounds(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	tri := Triangulation{
		X: []float64{0, 0.85, 1.75, 2.85, 0.2, 1.1, 2.1, 0.55, 1.55, 2.55},
		Y: []float64{0, 0.2, 0.05, 0.3, 1, 1.15, 1.25, 2.15, 2.3, 2.05},
		Triangles: [][3]int{
			{0, 1, 4},
			{1, 5, 4},
			{1, 2, 5},
			{2, 6, 5},
			{2, 3, 6},
			{4, 5, 7},
			{5, 8, 7},
			{5, 6, 8},
			{6, 9, 8},
		},
	}
	values := []float64{
		0.6655574652398372,
		1.4476504946307964,
		1.2769269603531197,
		-0.34020831291716114,
		-0.24685398108361084,
		0.3579864612613407,
		-0.48559426177070847,
		0.7782732859305255,
		1.1192548874057175,
		-0.4800029281709043,
	}

	contours := ax.TriContour(tri, values, ContourOptions{LevelCount: 6})
	if contours == nil {
		t.Fatal("expected tricontour set")
	}
	want := []float64{-0.6, -0.3, 0, 0.3, 0.6, 0.9, 1.2, 1.5}
	if len(contours.Levels) != len(want) {
		t.Fatalf("tricontour levels = %v, want %v", contours.Levels, want)
	}
	for i := range want {
		if !approx(contours.Levels[i], want[i], 1e-12) {
			t.Fatalf("tricontour levels[%d] = %v, want %v (all %v)", i, contours.Levels[i], want[i], contours.Levels)
		}
	}
}

func TestTriColorExposesConfiguredNorm(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	tri := Triangulation{
		X:         []float64{0, 1, 0},
		Y:         []float64{0, 0, 1},
		Triangles: [][3]int{{0, 1, 2}},
	}
	mesh := ax.TriColor(tri, []float64{10}, TriColorOptions{
		Norm: PowerNorm{Gamma: 2, VMin: 0, VMax: 20},
	})
	if mesh == nil {
		t.Fatal("expected tripcolor collection")
	}
	if mesh.Norm == nil || mesh.Norm.NormName() != "power" {
		t.Fatalf("tripcolor norm = %#v, want power norm", mesh.Norm)
	}
}

func TestContourSetExposesConfiguredNorm(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	contours := ax.Contourf([][]float64{
		{1, 10},
		{10, 100},
	}, ContourOptions{
		Levels: []float64{1, 10, 100},
		Norm:   LogNorm{VMin: 1, VMax: 100},
	})
	if contours == nil {
		t.Fatal("expected contour set")
	}
	mapping := contours.ScalarMap()
	if mapping.Norm == nil || mapping.Norm.NormName() != "log" {
		t.Fatalf("contour norm = %#v, want log norm", mapping.Norm)
	}
}

func TestContourLabelsDrawOverlay(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	contours := ax.Contour([][]float64{
		{0, 1, 2},
		{1, 2, 3},
		{2, 3, 4},
	}, ContourOptions{
		Levels:     []float64{1.5},
		LabelLines: true,
	})
	if contours == nil {
		t.Fatal("expected contour set")
	}

	var renderer contourTextRenderer
	DrawFigure(fig, &renderer)
	if len(renderer.texts) == 0 {
		t.Fatal("expected contour labels to be rendered")
	}
}

func TestAxesClabelDelegatesToContourSetAndFiltersLevels(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	contours := ax.Contour([][]float64{
		{0, 1, 2},
		{1, 2, 3},
		{2, 3, 4},
	}, ContourOptions{
		Levels: []float64{1, 2, 3},
	})
	if contours == nil {
		t.Fatal("expected contour set")
	}

	fontSize := 12.0
	color := render.Color{R: 0.8, G: 0.1, B: 0.2, A: 1}
	labels := ax.Clabel(contours, ClabelOptions{
		Levels:    []float64{2},
		Formatter: FuncFormatter(func(float64) string { return "L2" }),
		FontSize:  &fontSize,
		Color:     &color,
	})

	if len(labels) != 1 {
		t.Fatalf("labels = %d, want one filtered contour label", len(labels))
	}
	if labels[0].Text != "L2" || labels[0].Level != 2 || labels[0].Color != color {
		t.Fatalf("label = %+v, want level 2 text/color", labels[0])
	}
	if !contours.inlineLabels {
		t.Fatal("Axes.Clabel should default to inline labels like Matplotlib")
	}
	if contours.LabelFontSize != fontSize {
		t.Fatalf("label font size = %v, want %v", contours.LabelFontSize, fontSize)
	}

	var renderer contourTextRenderer
	DrawFigure(fig, &renderer)
	if len(renderer.texts) == 0 {
		t.Fatal("expected clabel text to be rendered")
	}
}

func TestAxesClabelManualPositionsPlaceNearestContourLabels(t *testing.T) {
	ax := NewFigure(640, 480).AddAxes(geom.Rect{})
	contours := ax.Contour([][]float64{
		{0, 1, 2},
		{1, 2, 3},
		{2, 3, 4},
	}, ContourOptions{
		Levels: []float64{1, 2, 3},
	})
	if contours == nil {
		t.Fatal("expected contour set")
	}

	inline := false
	labels := ax.Clabel(contours, ClabelOptions{
		Levels:          []float64{2},
		ManualPositions: []geom.Pt{{X: 1, Y: 1}},
		Inline:          &inline,
	})

	if len(labels) != 1 {
		t.Fatalf("manual labels = %d, want 1", len(labels))
	}
	if labels[0].Level != 2 {
		t.Fatalf("manual label level = %v, want nearest requested level 2", labels[0].Level)
	}
	if contours.inlineLabels {
		t.Fatal("manual non-inline clabel should not erase contour lines")
	}
	if len(contours.labels) != 1 || contours.labels[0].Position == (geom.Pt{}) {
		t.Fatalf("stored contour labels = %+v", contours.labels)
	}
}

func TestContourInlineLabelAngleUsesMatplotlibDisplayConvention(t *testing.T) {
	screen := []geom.Pt{
		{X: 0, Y: 10},
		{X: 5, Y: 5},
		{X: 10, Y: 0},
	}
	data := []geom.Pt{
		{X: 0, Y: 0},
		{X: 5, Y: 5},
		{X: 10, Y: 10},
	}

	angle, parts := splitContourPolylineForLabel(data, screen, 1, 4, 0)
	if len(parts) == 0 {
		t.Fatal("expected split contour parts")
	}
	if !approx(angle, -math.Pi/4, 1e-12) {
		t.Fatalf("angle = %v, want %v", angle, -math.Pi/4)
	}
}

func TestContourInlineLabelErasesAcrossClosedPathBoundary(t *testing.T) {
	screen := []geom.Pt{
		{X: 0, Y: 0},
		{X: 10, Y: 0},
		{X: 10, Y: 10},
		{X: 0, Y: 10},
		{X: 0, Y: 0},
	}
	data := append([]geom.Pt(nil), screen...)

	angle, parts := splitContourPolylineForLabel(data, screen, 0, 4, 1)
	if got, want := len(parts), 1; got != want {
		t.Fatalf("closed contour split parts = %d, want %d: %+v", got, want, parts)
	}
	want := []geom.Pt{
		{X: 3, Y: 0},
		{X: 10, Y: 0},
		{X: 10, Y: 10},
		{X: 0, Y: 10},
		{X: 0, Y: 3},
	}
	if !pointsEqual(parts[0], want, 1e-12) {
		t.Fatalf("closed contour split = %+v, want %+v", parts[0], want)
	}
	if !approx(angle, -math.Pi/4, 1e-12) {
		t.Fatalf("angle = %v, want %v", angle, -math.Pi/4)
	}
}

func TestContourInlineLabelsCoverDenseSparseAndShortContours(t *testing.T) {
	ctx := createTestDrawContext()
	renderer := &contourTextRenderer{}
	fontSize := 10.0

	dense := make([]geom.Pt, 0, 41)
	for i := 0; i <= 40; i++ {
		x := float64(i) / 10
		dense = append(dense, geom.Pt{X: x, Y: 2 + 0.05*math.Sin(x)})
	}
	sparse := []geom.Pt{
		{X: 0, Y: 5},
		{X: 4, Y: 5},
	}
	short := []geom.Pt{
		{X: 0, Y: 8},
		{X: 0.05, Y: 8},
	}
	lines := &LineCollection{
		Segments: [][]geom.Pt{dense, sparse, short},
		Colors: []render.Color{
			{R: 1, A: 1},
			{G: 1, A: 1},
			{B: 1, A: 1},
		},
		LineWidths: []float64{1, 2, 3},
	}

	segments, colors, widths, labels := contourInlineLabelSegments(
		lines,
		[]float64{1, 2, 3},
		ScalarFormatter{Prec: 0},
		fontSize,
		renderer,
		ctx,
	)
	if got, want := len(labels), 2; got != want {
		t.Fatalf("inline labels = %d, want dense and sparse labels only: %+v", got, labels)
	}
	if labels[0].Text != "1" || labels[1].Text != "2" {
		t.Fatalf("inline label texts = %q, %q; want dense/sparse levels 1 and 2", labels[0].Text, labels[1].Text)
	}
	if len(segments) <= len(lines.Segments) {
		t.Fatalf("segments were not split for inline erasure: got %d, original %d", len(segments), len(lines.Segments))
	}
	if len(colors) != len(segments) || len(widths) != len(segments) {
		t.Fatalf("split line style arrays do not match segments: segments=%d colors=%d widths=%d", len(segments), len(colors), len(widths))
	}
	if !approx(labels[1].Angle, 0, 1e-12) {
		t.Fatalf("sparse horizontal label angle = %v, want 0", labels[1].Angle)
	}
	for _, label := range labels {
		screenPos := ctx.DataToPixel.Apply(label.Position)
		if screenPos.X < 40 || screenPos.X > 460 || screenPos.Y < -60 || screenPos.Y > 460 {
			t.Fatalf("label position %+v maps outside expected display area: %+v", label.Position, screenPos)
		}
	}
}

func TestContourRotatedTextAnchorKeepsCenterFixed(t *testing.T) {
	center := geom.Pt{X: 100, Y: 200}
	angle := math.Pi / 6
	layout := singleLineTextLayout{
		TextLineLayout: render.TextLineLayout{
			Width:   24,
			Height:  12,
			Ascent:  9,
			Descent: 3,
		},
	}

	anchor := contourRotatedTextAnchor(center, layout, angle)
	want := rotatedTextBackendAnchorFromP(center, layout, TextAlignCenter, textLayoutVAlignCenter, angle, false)
	if !pointsApprox(anchor, want, 1e-12) {
		t.Fatalf("contour rotated text anchor = %+v, want center-aligned backend anchor %+v", anchor, want)
	}
}

type contourTextRenderer struct {
	render.NullRenderer
	texts []string
}

func (r *contourTextRenderer) DrawText(text string, _ geom.Pt, _ float64, _ render.Color) {
	r.texts = append(r.texts, text)
}

func (r *contourTextRenderer) MeasureText(text string, size float64, _ string) render.TextMetrics {
	return render.TextMetrics{
		W:       float64(len(text)) * size * 0.5,
		H:       size,
		Ascent:  size * 0.8,
		Descent: size * 0.2,
	}
}

type arraysShowcaseContourTextMetricRenderer struct {
	recordingRenderer
}

func (r *arraysShowcaseContourTextMetricRenderer) MeasureText(text string, size float64, _ string) render.TextMetrics {
	widths := map[string]float64{
		"0.15": 30.75,
		"0.3":  21.875,
		"0.30": 30.625,
		"0.45": 30.625,
		"0.6":  22,
		"0.60": 30.75,
		"0.75": 30.625,
		"0.9":  22,
		"0.90": 30.75,
	}
	width, ok := widths[text]
	if !ok {
		width = float64(len(text)) * size * 0.5
	}
	return render.TextMetrics{
		W:       width,
		H:       size * 1.4,
		Ascent:  size,
		Descent: size * 0.4,
	}
}
