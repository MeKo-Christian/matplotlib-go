package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
)

func TestTriContourfSkipsMaskedTriangles(t *testing.T) {
	tri := Triangulation{
		X:         []float64{0, 1, 0, 1},
		Y:         []float64{0, 0, 1, 1},
		Triangles: [][3]int{{0, 1, 2}, {1, 3, 2}},
		Mask:      []bool{true, false},
	}
	values := []float64{0, 1, 1, 2}
	polygons, _, _ := contourBandPolygons(
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

func TestNewTriangulationBuildsDelaunayMesh(t *testing.T) {
	tri, err := NewTriangulation(
		[]float64{0, 1, 0, 1},
		[]float64{0, 0, 1, 1},
	)
	if err != nil {
		t.Fatalf("NewTriangulation: %v", err)
	}
	if len(tri.Triangles) != 2 {
		t.Fatalf("triangles = %v, want two triangles covering the square", tri.Triangles)
	}
	if len(tri.X) != 4 || len(tri.Y) != 4 {
		t.Fatalf("coordinate lengths = %d/%d, want copied input coordinates", len(tri.X), len(tri.Y))
	}

	tri.X[0] = 99
	originalX := []float64{0, 1, 0, 1}
	tri, err = NewTriangulation(originalX, []float64{0, 0, 1, 1})
	if err != nil {
		t.Fatalf("NewTriangulation after mutation: %v", err)
	}
	originalX[0] = 42
	if tri.X[0] != 0 {
		t.Fatalf("NewTriangulation retained caller slice: got x[0]=%v, want independent copy", tri.X[0])
	}
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
	plot := ax.TriPlot(tri, TriPlotOptions{LineWidth: optional.Of(lineWidth), Label: "mesh"})
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
