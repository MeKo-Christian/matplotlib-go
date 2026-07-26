package core

import (
	"math"
	"reflect"
	"testing"

	matcolor "github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
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
		EdgeWidth: optional.Of(edgeWidth),
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
		EdgeColor: optional.Of(edge),
		EdgeWidth: optional.Of(width),
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
		EdgeColor: optional.Of(edge),
		EdgeWidth: optional.Of(width),
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

func TestAxesPColorMeshSupportsExplicitAntialiasing(t *testing.T) {
	ax := &Axes{
		XScale: transform.NewLinear(0, 2),
		YScale: transform.NewLinear(0, 2),
		XAxis:  NewXAxis(),
		YAxis:  NewYAxis(),
	}
	width := 0.75
	edge := render.Color{A: 1}
	antialias := true
	mesh := ax.PColorMesh([][]float64{{0, 1}}, MeshOptions{
		XEdges:    []float64{0, 1, 2},
		YEdges:    []float64{0, 1},
		EdgeColor: optional.Of(edge),
		EdgeWidth: optional.Of(width),
		Antialias: optional.Of(antialias),
	})
	if mesh == nil {
		t.Fatal("expected pcolormesh mesh")
	}

	r := &batchRecordingRenderer{returnNative: true}
	mesh.Draw(r, createTestDrawContext())
	if len(r.quadMeshBatches) != 1 || len(r.quadMeshBatches[0].Cells) == 0 {
		t.Fatalf("expected one native quad mesh batch, got %v", r.quadMeshBatches)
	}
	if !r.quadMeshBatches[0].Cells[0].Antialiased {
		t.Fatal("explicit antialiasing should enable pcolormesh antialiasing")
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
		Colormap: optional.Of(cmap),
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
	}, MeshOptions{})
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
		Colormap: optional.Of(cmapName),
		VMin:     optional.Of(vmin),
		VMax:     optional.Of(vmax),
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
		Colormap: optional.Of(cmap),
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
		Colormap: optional.Of(cmapName),
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
