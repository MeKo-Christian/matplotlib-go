package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

func TestAxes3DWireframeGeneratesLineCollection(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	x := []float64{0, 1}
	y := []float64{0, 1}
	z := [][]float64{
		{0, 1},
		{1, 2},
	}
	collection := ax.Wireframe(x, y, z)
	if collection == nil {
		t.Fatal("Wireframe returned nil")
	}
	if got, want := len(collection.Segments), 4; got != want {
		t.Fatalf("segment count = %d, want %d", got, want)
	}
}

func TestAxes3DWireframeTreatsZRowsAsYAndColumnsAsX(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetDistance(0)
	ax.SetView(0, 0)

	x := []float64{10, 20, 30}
	y := []float64{1, 2}
	z := [][]float64{
		{0, 0, 0},
		{0, 0, 0},
	}
	collection := ax.Wireframe(x, y, z)
	if collection == nil {
		t.Fatal("Wireframe returned nil")
	}
	if got, want := collection.Segments[0][0], (Pt{X: 10, Y: 1}); got != want {
		t.Fatalf("first wireframe point = %+v, want %+v", got, want)
	}
	if got, want := collection.Segments[0][1], (Pt{X: 20, Y: 1}); got != want {
		t.Fatalf("first wireframe row segment end = %+v, want %+v", got, want)
	}
}

func TestAxes3DWireframeSupportsRowColumnStridesLikeMatplotlib(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetDistance(0)
	ax.SetView(0, 0)

	x, y, z := testGrid3D(5, 5)
	rstride := 2
	cstride := 0
	collection := ax.Wireframe(x, y, z, PlotOptions{RStride: &rstride, CStride: &cstride})
	if collection == nil {
		t.Fatal("Wireframe returned nil")
	}
	if got, want := len(collection.Segments), 3; got != want {
		t.Fatalf("wireframe stride line count = %d, want sampled rows 0,2,4 only (%d)", got, want)
	}
	if got, want := len(collection.Segments[0]), 5; got != want {
		t.Fatalf("wireframe row polyline length = %d, want full row length %d", got, want)
	}
	if got, want := collection.Segments[1][0], (Pt{X: 0, Y: 2}); got != want {
		t.Fatalf("second sampled row starts at %+v, want row 2 start %+v", got, want)
	}
}

func TestAxes3DWireframeSupportsRowColumnCountsLikeMatplotlib(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	x, y, z := testGrid3D(9, 10)
	rcount := 3
	ccount := 4
	collection := ax.Wireframe(x, y, z, PlotOptions{RCount: &rcount, CCount: &ccount})
	if collection == nil {
		t.Fatal("Wireframe returned nil")
	}
	if got, want := len(collection.Segments), 8; got != want {
		t.Fatalf("wireframe count line count = %d, want 4 sampled rows + 4 sampled columns", got)
	}
}

func TestAxes3DWireframeDefaultLineWidthMatchesMatplotlib(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	x, y, z := testGrid3D(2, 2)
	collection := ax.Wireframe(x, y, z)
	if collection == nil {
		t.Fatal("Wireframe returned nil")
	}
	if got, want := collection.LineWidth, 1.5; got != want {
		t.Fatalf("wireframe default line width = %v, want Matplotlib lines.linewidth default in points %v", got, want)
	}
}

func TestAxes3DWireframeColorsApplyAlphaAndStayNonMappable(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	color := render.Color{R: 0.2, G: 0.4, B: 0.6, A: 1}
	alpha := 0.35
	lineWidth := 1.25
	collection := ax.Wireframe(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 1}, {1, 2}},
		PlotOptions{Color: &color, Alpha: &alpha, LineWidth: &lineWidth},
	)
	if collection == nil {
		t.Fatal("Wireframe returned nil")
	}
	if got := collection.Color; got != color {
		t.Fatalf("wireframe color = %+v, want explicit Line3DCollection color %+v", got, color)
	}
	if got := collection.Alpha; got != alpha {
		t.Fatalf("wireframe alpha = %v, want collection alpha %v", got, alpha)
	}
	if got := collection.LineWidth; got != lineWidth {
		t.Fatalf("wireframe line width = %v, want %v", got, lineWidth)
	}
	if array := collection.GetArray(); len(array) != 0 {
		t.Fatalf("wireframe scalar array = %v, want non-scalar-mappable Line3DCollection", array)
	}
	mapping := collection.ScalarMap()
	if mapping.Colormap != "" || mapping.Norm != nil || mapping.VMin != 0 || mapping.VMax != 0 {
		t.Fatalf("wireframe scalar map = %+v, want no scalar-map metadata", mapping)
	}
}

func TestAxes3DWireframeAxLimClipDropsOutsideRuns(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetXLim(0, 1)

	collection := ax.Wireframe(
		[]float64{0, 2},
		[]float64{0, 1},
		[][]float64{{0, 0}, {0, 0}},
		PlotOptions{AxLimClip: true},
	)
	if collection == nil {
		t.Fatal("Wireframe returned nil")
	}
	if got, want := len(collection.Segments), 1; got != want {
		t.Fatalf("wireframe clipped segments = %d, want only the in-limit column (%d)", got, want)
	}
	want := []Pt{
		ax.ProjectPoint(0, 0, 0),
		ax.ProjectPoint(0, 1, 0),
	}
	if !pointsEqual(collection.Segments[0], want, 1e-12) {
		t.Fatalf("wireframe clipped segment = %+v, want %+v", collection.Segments[0], want)
	}
}

func TestAxes3DSurfaceCreatesProjectedPolygons(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	x := []float64{0, 1}
	y := []float64{0, 1}
	z := [][]float64{
		{0, 1},
		{1, 2},
	}
	collection := ax.Surface(x, y, z)
	if collection == nil {
		t.Fatal("Surface returned nil")
	}
	if got, want := len(collection.Polygons), 1; got != want {
		t.Fatalf("surface polygon count = %d, want %d", got, want)
	}
	if got, want := len(collection.FaceColors), 1; got != want {
		t.Fatalf("surface face color count = %d, want %d", got, want)
	}
}

func TestAxes3DSurfaceUsesMatplotlibDefaultSampleCounts(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	x := make([]float64, 90)
	for i := range x {
		x[i] = float64(i)
	}
	y := make([]float64, 70)
	for i := range y {
		y[i] = float64(i)
	}
	z := make([][]float64, len(y))
	for row := range z {
		z[row] = make([]float64, len(x))
		for col := range z[row] {
			z[row][col] = float64(row + col)
		}
	}

	collection := ax.Surface(x, y, z)
	if collection == nil {
		t.Fatal("Surface returned nil")
	}
	if got, want := len(collection.Polygons), 35*45; got != want {
		t.Fatalf("surface polygon count = %d, want Matplotlib default rcount/ccount sampled count %d", got, want)
	}
}

func TestAxes3DSurfaceSupportsRowColumnStridesLikeMatplotlib(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	x, y, z := testGrid3D(5, 5)
	rstride := 2
	cstride := 2
	collection := ax.Surface(x, y, z, PlotOptions{RStride: &rstride, CStride: &cstride})
	if collection == nil {
		t.Fatal("Surface returned nil")
	}
	if got, want := len(collection.Polygons), 4; got != want {
		t.Fatalf("surface stride polygon count = %d, want 2x2 sampled patches", got)
	}
}

func TestAxes3DSurfaceSupportsRowColumnCountsLikeMatplotlib(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	x, y, z := testGrid3D(9, 10)
	rcount := 3
	ccount := 4
	collection := ax.Surface(x, y, z, PlotOptions{RCount: &rcount, CCount: &ccount})
	if collection == nil {
		t.Fatal("Surface returned nil")
	}
	if got, want := len(collection.Polygons), 9; got != want {
		t.Fatalf("surface count polygon count = %d, want %d sampled patches for 9x10 grid with rcount=3, ccount=4", got, want)
	}
}

func TestAxes3DPlotSurfaceGridHonorsSurfaceOptions(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	antialiased := false
	collection := ax.PlotSurfaceGrid(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 1}, {2, 3}},
		PlotOptions{Antialiased: &antialiased},
	)
	if collection == nil {
		t.Fatal("PlotSurfaceGrid returned nil")
	}
	if got, want := collection.Antialias, render.AntialiasOff; got != want {
		t.Fatalf("plot surface grid antialias = %v, want %v", got, want)
	}
}

func TestAxes3DSurfaceDefaultHasNoEdgeColorsLikeMatplotlibCmapSurface(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	collection := ax.Surface(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 1}, {1, 2}},
	)
	if collection == nil {
		t.Fatal("Surface returned nil")
	}
	if got := collection.EdgeColor.A; got != 0 {
		t.Fatalf("surface default edge alpha = %v, want 0 like Matplotlib cmap plot_surface edgecolors", got)
	}
	if got, want := collection.EdgeWidth, 1.0; got != want {
		t.Fatalf("surface default linewidth = %v, want %v like Matplotlib plot_surface", got, want)
	}
	if len(collection.FaceColors) == 0 || collection.FaceColors[0].A != 1 {
		t.Fatalf("surface default face alpha = %v, want opaque Matplotlib default", collection.FaceColors)
	}
}

func TestAxes3DSurfaceExposesScalarMapForColorbars(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	cmap := "plasma"
	vmin := 0.0
	vmax := 10.0
	surface := ax.Surface(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 2}, {4, 6}},
		PlotOptions{Colormap: &cmap, VMin: &vmin, VMax: &vmax},
	)
	if surface == nil {
		t.Fatal("Surface returned nil")
	}
	mapping := surface.ScalarMap()
	if mapping.Colormap != cmap || mapping.VMin != vmin || mapping.VMax != vmax {
		t.Fatalf("surface scalar map = %+v, want cmap=%q range %.1f..%.1f", mapping, cmap, vmin, vmax)
	}
	array := surface.GetArray()
	if got, want := len(array), 1; got != want {
		t.Fatalf("surface scalar array len = %d, want %d Matplotlib average-z value", got, want)
	}
	if len(array) == 1 && !approx(array[0], 3, 1e-12) {
		t.Fatalf("surface scalar array = %v, want Matplotlib face average z [3]", array)
	}
}

func TestAxes3DSurfaceUsesFaceColorsForEdgesLikeMatplotlib(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	face := render.Color{R: 0.2, G: 0.4, B: 0.6, A: 1}
	surface := ax.Surface(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 2}, {4, 6}},
		PlotOptions{FaceColors: []render.Color{face}},
	)
	if surface == nil {
		t.Fatal("Surface returned nil")
	}
	if got, want := len(surface.EdgeColors), len(surface.FaceColors); got != want {
		t.Fatalf("surface edge colors len = %d, want face-colored edges for %d faces", got, want)
	}
	if got := surface.EdgeColors[0]; got != surface.FaceColors[0] {
		t.Fatalf("surface edge color = %+v, want same sampled face color %+v", got, surface.FaceColors[0])
	}
}

func TestAxes3DSurfaceHonorsAntialiasSetting(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	antialiased := false
	surface := ax.Surface(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 2}, {4, 6}},
		PlotOptions{Antialiased: &antialiased},
	)
	if surface == nil {
		t.Fatal("Surface returned nil")
	}
	if got, want := surface.Antialias, render.AntialiasOff; got != want {
		t.Fatalf("surface antialias = %v, want %v", got, want)
	}
}

func TestAxes3DSurfaceAxLimClipDropsFacesOutsideExplicit3DLimits(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetXLim(0, 1.5)

	surface := ax.Surface(
		[]float64{0, 1, 2},
		[]float64{0, 1},
		[][]float64{{0, 1, 2}, {0, 1, 2}},
		PlotOptions{AxLimClip: true},
	)
	if surface == nil {
		t.Fatal("Surface returned nil")
	}
	if got, want := len(surface.Polygons), 1; got != want {
		t.Fatalf("surface clipped polygon count = %d, want one fully in-limit face", got)
	}
}

func TestAxes3DTrisurfExposesConfiguredNorm(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	tri := Triangulation{
		X:         []float64{0, 1, 0},
		Y:         []float64{0, 0, 1},
		Triangles: [][3]int{{0, 1, 2}},
	}
	cmap := "inferno"
	surface := ax.Trisurf(tri, []float64{1, 10, 100}, PlotOptions{
		Colormap: &cmap,
		Norm:     LogNorm{VMin: 1, VMax: 100},
	})
	if surface == nil {
		t.Fatal("Trisurf returned nil")
	}
	mapping := surface.ScalarMap()
	if mapping.Colormap != cmap || mapping.Norm == nil || mapping.Norm.NormName() != "log" {
		t.Fatalf("trisurf scalar map = %+v, want inferno/log norm", mapping)
	}
	array := surface.GetArray()
	if got, want := len(array), 1; got != want {
		t.Fatalf("trisurf scalar array len = %d, want %d Matplotlib average-z value", got, want)
	}
	if len(array) == 1 && !approx(array[0], 37, 1e-12) {
		t.Fatalf("trisurf scalar array = %v, want Matplotlib triangle average z [37]", array)
	}
}

func TestAxes3DTrisurfHonorsAntialiasSetting(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	antialiased := false
	tri := Triangulation{
		X:         []float64{0, 1, 0},
		Y:         []float64{0, 0, 1},
		Triangles: [][3]int{{0, 1, 2}},
	}
	surface := ax.Trisurf(tri, []float64{1, 10, 100}, PlotOptions{Antialiased: &antialiased})
	if surface == nil {
		t.Fatal("Trisurf returned nil")
	}
	if got, want := surface.Antialias, render.AntialiasOff; got != want {
		t.Fatalf("trisurf antialias = %v, want %v", got, want)
	}
}

func TestAxes3DStemProjectsBaselineStemsAndMarkers(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetDistance(0)
	ax.SetView(0, 0)

	bottom := -1.0
	container := ax.Stem3D(
		[]float64{0, 1},
		[]float64{2, 3},
		[]float64{4, 5},
		Stem3DOptions{Bottom: &bottom},
	)
	if container == nil {
		t.Fatal("Stem3D returned nil")
	}
	if got, want := len(container.StemLines.Segments), 2; got != want {
		t.Fatalf("stem segment count = %d, want %d", got, want)
	}
	wantStem := []Pt{
		ax.ProjectPoint(0, 2, bottom),
		ax.ProjectPoint(0, 2, 4),
	}
	if !pointsEqual(container.StemLines.Segments[0], wantStem, 1e-12) {
		t.Fatalf("first stem = %+v, want projected z-oriented stem %+v", container.StemLines.Segments[0], wantStem)
	}
	wantBaseline := []Pt{
		ax.ProjectPoint(0, 2, bottom),
		ax.ProjectPoint(1, 3, bottom),
	}
	if !pointsEqual(container.Baseline.XY, wantBaseline, 1e-12) {
		t.Fatalf("stem baseline = %+v, want projected baseline %+v", container.Baseline.XY, wantBaseline)
	}
	if got, want := len(container.MarkerCollection.Offsets), 2; got != want {
		t.Fatalf("stem marker count = %d, want %d", got, want)
	}
	palette := style.Default.Palette()
	if got, want := container.StemLines.Color, palette[0]; got != want {
		t.Fatalf("stem line color = %+v, want Matplotlib linefmt C0 %+v", got, want)
	}
	if got, want := container.StemLines.LineCap, render.CapButt; got != want {
		t.Fatalf("stem line cap = %v, want Matplotlib LineCollection default cap %v", got, want)
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

func TestAxes3DStemColorsApplyAlphaAndStayNonMappable(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	color := render.Color{R: 0.2, G: 0.4, B: 0.6, A: 1}
	baseline := render.Color{R: 0.8, G: 0.1, B: 0.3, A: 1}
	markerEdge := render.Color{R: 0.9, G: 0.7, B: 0.2, A: 1}
	alpha := 0.4
	container := ax.Stem3D(
		[]float64{0, 1},
		[]float64{0, 1},
		[]float64{1, 2},
		Stem3DOptions{
			Color:           &color,
			BaselineColor:   &baseline,
			MarkerEdgeColor: &markerEdge,
			Alpha:           &alpha,
		},
	)
	if container == nil {
		t.Fatal("Stem3D returned nil")
	}
	wantStem := color
	wantStem.A *= alpha
	if got := container.StemLines.Color; got != wantStem {
		t.Fatalf("stem line color = %+v, want %+v", got, wantStem)
	}
	if got := container.MarkerCollection.FaceColor; got != wantStem {
		t.Fatalf("stem marker face color = %+v, want %+v", got, wantStem)
	}
	wantEdge := markerEdge
	wantEdge.A *= alpha
	if got := container.MarkerCollection.EdgeColor; got != wantEdge {
		t.Fatalf("stem marker edge color = %+v, want %+v", got, wantEdge)
	}
	wantBaseline := baseline
	wantBaseline.A *= alpha
	if got := container.Baseline.Col; got != wantBaseline {
		t.Fatalf("stem baseline color = %+v, want %+v", got, wantBaseline)
	}
	if array := container.StemLines.GetArray(); len(array) != 0 {
		t.Fatalf("stem line scalar array = %v, want non-scalar-mappable LineCollection", array)
	}
	if array := container.MarkerCollection.GetArray(); len(array) != 0 {
		t.Fatalf("stem marker scalar array = %v, want non-scalar-mappable PathCollection", array)
	}
	mapping := container.StemLines.ScalarMap()
	if mapping.Colormap != "" || mapping.Norm != nil || mapping.VMin != 0 || mapping.VMax != 0 {
		t.Fatalf("stem scalar map = %+v, want no scalar-map metadata", mapping)
	}
}

func TestAxes3DStemSupportsMatplotlibOrientationJuggling(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetDistance(0)
	ax.SetView(0, 0)

	bottom := -2.0
	container := ax.Stem3D(
		[]float64{1},
		[]float64{2},
		[]float64{3},
		Stem3DOptions{Bottom: &bottom, Orientation: "x"},
	)
	if container == nil {
		t.Fatal("Stem3D returned nil")
	}
	want := []Pt{
		ax.ProjectPoint(bottom, 2, 3),
		ax.ProjectPoint(1, 2, 3),
	}
	if !pointsEqual(container.StemLines.Segments[0], want, 1e-12) {
		t.Fatalf("x-oriented stem = %+v, want Matplotlib orientation='x' stem %+v", container.StemLines.Segments[0], want)
	}
}

func TestAxes3DStemAxLimClipDropsOutsideStemsAndMarkers(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetXLim(0, 1)

	container := ax.Stem3D(
		[]float64{0.25, 2},
		[]float64{0, 0},
		[]float64{1, 1},
		Stem3DOptions{AxLimClip: true},
	)
	if container == nil {
		t.Fatal("Stem3D returned nil")
	}
	if got, want := len(container.StemLines.Segments), 1; got != want {
		t.Fatalf("clipped stem segments = %d, want %d", got, want)
	}
	if got, want := len(container.MarkerCollection.Offsets), 1; got != want {
		t.Fatalf("clipped stem markers = %d, want %d", got, want)
	}
	if got, want := len(container.Baseline.XY), 1; got != want {
		t.Fatalf("clipped stem baseline points = %d, want %d", got, want)
	}
	want := []Pt{
		ax.ProjectPoint(0.25, 0, 0),
		ax.ProjectPoint(0.25, 0, 1),
	}
	if !pointsEqual(container.StemLines.Segments[0], want, 1e-12) {
		t.Fatalf("clipped stem segment = %+v, want %+v", container.StemLines.Segments[0], want)
	}
}

func TestAxes3DFillBetweenCreatesProjectedQuadBands(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetDistance(0)
	ax.SetView(0, 0)

	fill := ax.FillBetween3D(
		[]float64{0, 1, 2},
		[]float64{0, 0, 0},
		[]float64{1, 1, 1},
		[]float64{0, 1, 2},
		[]float64{1, 1, 1},
		[]float64{0, 0, 0},
		FillBetween3DOptions{Mode: FillBetween3DModeQuad},
	)
	if fill == nil {
		t.Fatal("FillBetween3D returned nil")
	}
	if got, want := len(fill.Polygons), 2; got != want {
		t.Fatalf("FillBetween3D polygon count = %d, want one quad per adjacent pair (%d)", got, want)
	}
	wantFirst := []Pt{
		ax.ProjectPoint(0, 0, 1),
		ax.ProjectPoint(1, 0, 1),
		ax.ProjectPoint(1, 1, 0),
		ax.ProjectPoint(0, 1, 0),
	}
	if !pointsEqual(fill.Polygons[0], wantFirst, 1e-12) {
		t.Fatalf("first fill polygon = %+v, want projected quad %+v", fill.Polygons[0], wantFirst)
	}
}

func TestAxes3DFillBetweenColorsApplyAlphaAndStayNonMappable(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	color := render.Color{R: 0.2, G: 0.4, B: 0.6, A: 1}
	edge := render.Color{R: 0.8, G: 0.1, B: 0.3, A: 1}
	alpha := 0.35
	edgeWidth := 1.25
	// shade=false isolates the alpha propagation under test; quad mode would
	// otherwise shade each face like matplotlib's fill_between(shade=None).
	noShade := false
	fill := ax.FillBetween3D(
		[]float64{0, 1, 2},
		[]float64{0, 0, 0},
		[]float64{1, 1, 1},
		[]float64{0, 1, 2},
		[]float64{1, 1, 1},
		[]float64{0, 0, 0},
		FillBetween3DOptions{
			Color:     &color,
			EdgeColor: &edge,
			EdgeWidth: &edgeWidth,
			Alpha:     &alpha,
			Mode:      FillBetween3DModeQuad,
			Shade:     &noShade,
		},
	)
	if fill == nil {
		t.Fatal("FillBetween3D returned nil")
	}
	wantFace := color
	wantFace.A *= alpha
	if got, want := len(fill.FaceColors), len(fill.Polygons); got != want {
		t.Fatalf("fill face colors = %d, polygons = %d; want one color per quad", got, want)
	}
	for i, got := range fill.FaceColors {
		if got != wantFace {
			t.Fatalf("fill face color %d = %+v, want %+v", i, got, wantFace)
		}
	}
	wantEdge := edge
	wantEdge.A *= alpha
	if got := fill.EdgeColor; got != wantEdge {
		t.Fatalf("fill edge color = %+v, want %+v", got, wantEdge)
	}
	if got := fill.EdgeWidth; got != edgeWidth {
		t.Fatalf("fill edge width = %v, want %v", got, edgeWidth)
	}
	if array := fill.GetArray(); len(array) != 0 {
		t.Fatalf("fill scalar array = %v, want non-scalar-mappable PolyCollection", array)
	}
	mapping := fill.ScalarMap()
	if mapping.Colormap != "" || mapping.Norm != nil || mapping.VMin != 0 || mapping.VMax != 0 {
		t.Fatalf("fill scalar map = %+v, want no scalar-map metadata", mapping)
	}

	polygonFill := ax.FillBetween3D(
		[]float64{0, 1, 2},
		[]float64{0, 0, 0},
		[]float64{1, 1, 1},
		[]float64{0, 1, 2},
		[]float64{1, 1, 1},
		[]float64{0, 0, 0},
		FillBetween3DOptions{Color: &color, Alpha: &alpha, Mode: FillBetween3DModePolygon},
	)
	if polygonFill == nil {
		t.Fatal("polygon FillBetween3D returned nil")
	}
	if got, want := len(polygonFill.Polygons), 1; got != want {
		t.Fatalf("polygon fill polygons = %d, want %d", got, want)
	}
	if got, want := len(polygonFill.FaceColors), 1; got != want {
		t.Fatalf("polygon fill face colors = %d, want %d", got, want)
	}
	if polygonFill.FaceColors[0] != wantFace {
		t.Fatalf("polygon fill face color = %+v, want %+v", polygonFill.FaceColors[0], wantFace)
	}
}

func TestAxes3DFillBetweenAxLimClipDropsOutsidePolygons(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetXLim(0, 1.25)

	fill := ax.FillBetween3D(
		[]float64{0, 1, 2},
		[]float64{0, 0, 0},
		[]float64{1, 1, 1},
		[]float64{0, 1, 2},
		[]float64{1, 1, 1},
		[]float64{0, 0, 0},
		FillBetween3DOptions{Mode: FillBetween3DModeQuad, AxLimClip: true},
	)
	if fill == nil {
		t.Fatal("FillBetween3D returned nil")
	}
	if got, want := len(fill.Polygons), 1; got != want {
		t.Fatalf("clipped FillBetween3D polygons = %d, want only the in-limit quad (%d)", got, want)
	}
	want := []Pt{
		ax.ProjectPoint(0, 0, 1),
		ax.ProjectPoint(1, 0, 1),
		ax.ProjectPoint(1, 1, 0),
		ax.ProjectPoint(0, 1, 0),
	}
	if !pointsEqual(fill.Polygons[0], want, 1e-12) {
		t.Fatalf("clipped FillBetween3D polygon = %+v, want %+v", fill.Polygons[0], want)
	}
}

func TestAxes3DBarProjects2DBarsIntoSelectedZDirection(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetDistance(0)
	ax.SetView(0, 0)

	width := 0.4
	zs := []float64{2}
	bars := ax.Bar(
		[]float64{1},
		[]float64{3},
		Bar3DPlaneOptions{Width: &width, Zs: zs, ZDir: "y"},
	)
	if bars == nil {
		t.Fatal("Axes3D.Bar returned nil")
	}
	if got, want := len(bars.Polygons), 1; got != want {
		t.Fatalf("projected bar polygon count = %d, want %d", got, want)
	}
	want := []Pt{
		ax.ProjectPoint(0.8, 2, 0),
		ax.ProjectPoint(0.8, 2, 3),
		ax.ProjectPoint(1.2, 2, 3),
		ax.ProjectPoint(1.2, 2, 0),
	}
	if !pointsEqual(bars.Polygons[0], want, 1e-12) {
		t.Fatalf("projected y-dir bar = %+v, want Matplotlib juggle_axes projection %+v", bars.Polygons[0], want)
	}
}

func TestAxes3DQuiverUsesMatplotlibTailPivotGeometry(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetDistance(0)
	ax.SetView(0, 0)

	length := 2.0
	q := ax.Quiver(
		[]float64{0},
		[]float64{0},
		[]float64{0},
		[]float64{1},
		[]float64{0},
		[]float64{0},
		Quiver3DOptions{Length: &length, Pivot: "tail"},
	)
	if q == nil {
		t.Fatal("Quiver returned nil")
	}
	if got, want := len(q.Segments), 3; got != want {
		t.Fatalf("quiver segment count = %d, want shaft plus two arrowhead segments (%d)", got, want)
	}
	wantShaft := []Pt{
		ax.ProjectPoint(2, 0, 0),
		ax.ProjectPoint(0, 0, 0),
	}
	if !pointsEqual(q.Segments[0], wantShaft, 1e-12) {
		t.Fatalf("quiver shaft = %+v, want Matplotlib tail-pivot shaft %+v", q.Segments[0], wantShaft)
	}
	if got, want := q.LineWidth, 1.5; !approx(got, want, 1e-12) {
		t.Fatalf("quiver line width = %v, want Matplotlib Line3DCollection default 1.5 pt = %v px", got, want)
	}
}

func TestAxes3DQuiverColorsApplyAlphaAndStayNonMappable(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	color := render.Color{R: 0.2, G: 0.4, B: 0.6, A: 1}
	alpha := 0.45
	q := ax.Quiver(
		[]float64{0},
		[]float64{0},
		[]float64{0},
		[]float64{1},
		[]float64{0},
		[]float64{0},
		Quiver3DOptions{Color: &color, Alpha: &alpha},
	)
	if q == nil {
		t.Fatal("Quiver returned nil")
	}
	want := color
	want.A *= alpha
	if got := q.Color; got != want {
		t.Fatalf("quiver color = %+v, want %+v", got, want)
	}
	if array := q.GetArray(); len(array) != 0 {
		t.Fatalf("quiver scalar array = %v, want non-scalar-mappable Line3DCollection", array)
	}
	mapping := q.ScalarMap()
	if mapping.Colormap != "" || mapping.Norm != nil || mapping.VMin != 0 || mapping.VMax != 0 {
		t.Fatalf("quiver scalar map = %+v, want no scalar-map metadata", mapping)
	}
}

func TestAxes3DQuiverNormalizesVectorsAndSupportsMiddlePivot(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetDistance(0)
	ax.SetView(0, 0)

	length := 4.0
	q := ax.Quiver(
		[]float64{0},
		[]float64{0},
		[]float64{0},
		[]float64{2},
		[]float64{0},
		[]float64{0},
		Quiver3DOptions{Length: &length, Normalize: true, Pivot: "middle"},
	)
	if q == nil {
		t.Fatal("Quiver returned nil")
	}
	wantShaft := []Pt{
		ax.ProjectPoint(2, 0, 0),
		ax.ProjectPoint(-2, 0, 0),
	}
	if !pointsEqual(q.Segments[0], wantShaft, 1e-12) {
		t.Fatalf("normalized middle-pivot shaft = %+v, want %+v", q.Segments[0], wantShaft)
	}
}

func TestAxes3DQuiverAxLimClipDropsOutsideArrows(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetXLim(0, 1)

	length := 0.2
	q := ax.Quiver(
		[]float64{0.25, 2},
		[]float64{0, 0},
		[]float64{0, 0},
		[]float64{1, 1},
		[]float64{0, 0},
		[]float64{0, 0},
		Quiver3DOptions{Length: &length, Pivot: "tail", AxLimClip: true},
	)
	if q == nil {
		t.Fatal("Quiver returned nil")
	}
	if got, want := len(q.Segments), 3; got != want {
		t.Fatalf("clipped quiver segments = %d, want one arrow with shaft plus two heads (%d)", got, want)
	}
	wantShaft := []Pt{
		ax.ProjectPoint(0.45, 0, 0),
		ax.ProjectPoint(0.25, 0, 0),
	}
	if !pointsEqual(q.Segments[0], wantShaft, 1e-12) {
		t.Fatalf("clipped quiver shaft = %+v, want %+v", q.Segments[0], wantShaft)
	}
}

func TestAxes3DErrorBarProjectsXYZRangesAndCaps(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetDistance(0)
	ax.SetView(0, 0)

	capSize := 0.2
	errs := ax.ErrorBar3D(
		[]float64{1},
		[]float64{2},
		[]float64{3},
		[]float64{0.5},
		[]float64{0.25},
		[]float64{1},
		ErrorBar3DOptions{CapSize: &capSize},
	)
	if errs == nil {
		t.Fatal("ErrorBar3D returned nil")
	}
	if got, want := len(errs.Segments), 15; got != want {
		t.Fatalf("3D errorbar segment count = %d, want 3 bars plus 12 cap segments", got)
	}
	wantXRange := []Pt{
		ax.ProjectPoint(0.5, 2, 3),
		ax.ProjectPoint(1.5, 2, 3),
	}
	if !pointsEqual(errs.Segments[0], wantXRange, 1e-12) {
		t.Fatalf("x error range = %+v, want projected x range %+v", errs.Segments[0], wantXRange)
	}
	wantZRange := []Pt{
		ax.ProjectPoint(1, 2, 2),
		ax.ProjectPoint(1, 2, 4),
	}
	if !pointsEqual(errs.Segments[2], wantZRange, 1e-12) {
		t.Fatalf("z error range = %+v, want projected z range %+v", errs.Segments[2], wantZRange)
	}
}

func TestAxes3DErrorBarColorsApplyAlphaAndStayNonMappable(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	color := render.Color{R: 0.2, G: 0.4, B: 0.6, A: 1}
	alpha := 0.25
	errs := ax.ErrorBar3D(
		[]float64{0},
		[]float64{0},
		[]float64{1},
		nil,
		nil,
		[]float64{0.1},
		ErrorBar3DOptions{Color: &color, Alpha: &alpha},
	)
	if errs == nil {
		t.Fatal("ErrorBar3D returned nil")
	}
	want := color
	want.A *= alpha
	if got := errs.Color; got != want {
		t.Fatalf("errorbar color = %+v, want %+v", got, want)
	}
	if array := errs.GetArray(); len(array) != 0 {
		t.Fatalf("errorbar scalar array = %v, want non-scalar-mappable Line3DCollection", array)
	}
	mapping := errs.ScalarMap()
	if mapping.Colormap != "" || mapping.Norm != nil || mapping.VMin != 0 || mapping.VMax != 0 {
		t.Fatalf("errorbar scalar map = %+v, want no scalar-map metadata", mapping)
	}
}

func TestAxes3DErrorBarUsesComputedDepthZOrder(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	low := ax.ErrorBar3D([]float64{0}, []float64{0}, []float64{0}, nil, nil, []float64{0.1})
	high := ax.ErrorBar3D([]float64{0}, []float64{0}, []float64{1}, nil, nil, []float64{0.1})
	if low == nil || high == nil {
		t.Fatalf("expected errorbar collections, got low=%v high=%v", low, high)
	}
	if !(high.Z() > low.Z()) {
		t.Fatalf("3D errorbar zorders = low %.6g high %.6g, want projected depth ordering", low.Z(), high.Z())
	}
}

func TestAxes3DErrorBarAxLimClipDropsOutsideRanges(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetXLim(0, 1)

	errs := ax.ErrorBar3D(
		[]float64{0.5, 2},
		[]float64{0, 0},
		[]float64{0, 0},
		nil,
		nil,
		[]float64{0.1, 0.1},
		ErrorBar3DOptions{AxLimClip: true},
	)
	if errs == nil {
		t.Fatal("ErrorBar3D returned nil")
	}
	if got, want := len(errs.Segments), 1; got != want {
		t.Fatalf("clipped errorbar segments = %d, want only the in-limit z range (%d)", got, want)
	}
	want := []Pt{
		ax.ProjectPoint(0.5, 0, -0.1),
		ax.ProjectPoint(0.5, 0, 0.1),
	}
	if !pointsEqual(errs.Segments[0], want, 1e-12) {
		t.Fatalf("clipped errorbar segment = %+v, want %+v", errs.Segments[0], want)
	}
}

func TestAxes3DFillBetweenQuadModeShadesFacesLikeMatplotlib(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	// A non-planar ribbon: auto mode resolves to 'quad', whose matplotlib
	// shade default is true, so each quad gets a lightsource-shaded copy of
	// the base color (art3d._shade_colors via Poly3DCollection(shade=True)).
	n := 12
	x1 := make([]float64, n)
	y1 := make([]float64, n)
	z1 := make([]float64, n)
	x2 := make([]float64, n)
	y2 := make([]float64, n)
	z2 := make([]float64, n)
	for i := range x1 {
		t1 := 4 * math.Pi * float64(i) / float64(n-1)
		x1[i] = math.Cos(t1)
		y1[i] = math.Sin(t1)
		z1[i] = float64(i) / float64(n-1)
		x2[i] = 0.4 * math.Cos(t1)
		y2[i] = 0.4 * math.Sin(t1)
		z2[i] = z1[i]
	}
	base := render.Color{R: 0.2, G: 0.4, B: 0.6, A: 1}
	c := ax.FillBetween(x1, y1, z1, x2, y2, z2, FillBetween3DOptions{Color: &base})
	if c == nil {
		t.Fatal("FillBetween returned nil")
	}
	if len(c.FaceColors) != n-1 {
		t.Fatalf("face colors = %d, want %d quads", len(c.FaceColors), n-1)
	}
	allEqual := true
	for _, fc := range c.FaceColors[1:] {
		if fc != c.FaceColors[0] {
			allEqual = false
			break
		}
	}
	if allEqual {
		t.Fatalf("quad-mode fill_between face colors are uniform %+v; want per-quad lightsource shading", c.FaceColors[0])
	}
	for i, fc := range c.FaceColors {
		if fc.R > base.R || fc.G > base.G || fc.B > base.B {
			t.Fatalf("face color %d = %+v brighter than base %+v; shading only darkens", i, fc, base)
		}
		if fc.A != base.A {
			t.Fatalf("face color %d alpha = %v, want %v (shading preserves alpha)", i, fc.A, base.A)
		}
	}

	// Planar input: auto mode resolves to 'polygon', whose shade default is
	// false, so the single polygon keeps the unshaded base color.
	xs := []float64{0, 1, 2, 3}
	ones := []float64{1, 1, 1, 1}
	zeros := []float64{0, 0, 0, 0}
	flat := ax.FillBetween(xs, zeros, zeros, xs, ones, zeros, FillBetween3DOptions{Color: &base})
	if flat == nil {
		t.Fatal("planar FillBetween returned nil")
	}
	for i, fc := range flat.FaceColors {
		if fc != base {
			t.Fatalf("polygon-mode face color %d = %+v, want unshaded base %+v", i, fc, base)
		}
	}
}

func TestAxes3DBar3DCreatesSegments(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	collection := ax.Bar3D(
		[]float64{0, 1},
		[]float64{0, 1},
		[]float64{0, 1},
		[]float64{1, 1},
		[]float64{1, 1},
		[]float64{1, 1},
	)
	if collection == nil {
		t.Fatal("Bar3D returned nil")
	}
	if got, want := len(collection.Segments), 16; got != want {
		t.Fatalf("segment count = %d, want %d", got, want)
	}
	if got, want := collection.Alpha, 0.0; got != want {
		t.Fatalf("default Bar3D edge alpha = %v, want Matplotlib default edgecolors-none alpha %v", got, want)
	}
	if got := collection.Color.A; got != 0 {
		t.Fatalf("default Bar3D edge color alpha = %v, want Matplotlib default edgecolors none", got)
	}
	edgeRecorder := &batchRecordingRenderer{}
	collection.Draw(edgeRecorder, newAxesDrawContext(ax.Axes, fig, fig.DisplayRect(), ax.adjustedLayout(fig)))
	if len(edgeRecorder.pathCalls) != 0 {
		t.Fatalf("default Bar3D drew %d edge paths, want none like Matplotlib edgecolors=[]", len(edgeRecorder.pathCalls))
	}
	foundFaces := false
	for _, artist := range ax.Artists {
		polys, ok := artist.(*PolyCollection)
		if ok && len(polys.Polygons) == 6*2 {
			foundFaces = true
			break
		}
	}
	if !foundFaces {
		t.Fatal("Bar3D did not add filled projected cuboid faces")
	}
}

func TestAxes3DBar3DSingleColorAppliesAlphaShadingAndStaysNonMappable(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	color := render.Color{R: 0.2, G: 0.4, B: 0.6, A: 1}
	alpha := 0.4
	lineWidth := 1.5
	edges := ax.Bar3D(
		[]float64{0},
		[]float64{0},
		[]float64{0},
		[]float64{1},
		[]float64{1},
		[]float64{1},
		Bar3DOptions{Color: &color, Alpha: &alpha, LineWidth: &lineWidth},
	)
	if edges == nil {
		t.Fatal("Bar3D returned nil")
	}
	if got := edges.Color; got != color {
		t.Fatalf("Bar3D edge color = %+v, want unshaded explicit line color %+v", got, color)
	}
	if got := edges.Alpha; got != alpha {
		t.Fatalf("Bar3D edge collection alpha = %v, want %v", got, alpha)
	}
	if array := edges.GetArray(); len(array) != 0 {
		t.Fatalf("Bar3D edge scalar array = %v, want non-scalar-mappable LineCollection", array)
	}

	var faces *PolyCollection
	for _, artist := range ax.Artists {
		polys, ok := artist.(*PolyCollection)
		if ok && len(polys.Polygons) == 6 {
			faces = polys
			break
		}
	}
	if faces == nil {
		t.Fatal("Bar3D did not add filled projected cuboid faces")
	}
	if got, want := len(faces.FaceColors), 6; got != want {
		t.Fatalf("Bar3D face colors = %d, want %d", got, want)
	}
	wantFaceColors := []render.Color{
		{R: 0.083333, G: 0.166667, B: 0.250000, A: 0.4},
		{R: 0.176667, G: 0.353333, B: 0.530000, A: 0.4},
		{R: 0.106667, G: 0.213333, B: 0.320000, A: 0.4},
		{R: 0.153333, G: 0.306667, B: 0.460000, A: 0.4},
		{R: 0.083333, G: 0.166667, B: 0.250000, A: 0.4},
		{R: 0.176667, G: 0.353333, B: 0.530000, A: 0.4},
	}
	for i, got := range faces.FaceColors {
		want := wantFaceColors[i]
		if !approx(got.R, want.R, 1e-6) || !approx(got.G, want.G, 1e-6) || !approx(got.B, want.B, 1e-6) || !approx(got.A, want.A, 1e-12) {
			t.Fatalf("Bar3D face color %d = %+v, want Matplotlib shaded color %+v", i, got, want)
		}
	}
	if array := faces.GetArray(); len(array) != 0 {
		t.Fatalf("Bar3D face scalar array = %v, want non-scalar-mappable PolyCollection", array)
	}
	mapping := faces.ScalarMap()
	if mapping.Colormap != "" || mapping.Norm != nil || mapping.VMin != 0 || mapping.VMax != 0 {
		t.Fatalf("Bar3D face scalar map = %+v, want no scalar-map metadata", mapping)
	}
}

func TestAxes3DBar3DSupportsPerBarAndPerFaceColorsLikeMatplotlib(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	perBar := []render.Color{
		{R: 1, G: 0, B: 0, A: 0.21},
		{R: 0, G: 1, B: 0, A: 0.77},
	}
	ax.Bar3D(
		[]float64{0, 2},
		[]float64{0, 0},
		[]float64{0, 0},
		[]float64{1, 1},
		[]float64{1, 1},
		[]float64{1, 1},
		Bar3DOptions{Colors: perBar},
	)
	perBarFaces := latestBar3DFaceCollection(t, ax, 12)
	for _, want := range []float64{0.21, 0.77} {
		if got, wantCount := countFaceColorAlpha(perBarFaces.FaceColors, want), 6; got != wantCount {
			t.Fatalf("per-bar face alpha %.2f count = %d, want %d in %+v", want, got, wantCount, perBarFaces.FaceColors)
		}
	}

	facePattern := []render.Color{
		{R: 1, G: 0, B: 0, A: 0.11},
		{R: 0, G: 1, B: 0, A: 0.22},
		{R: 0, G: 0, B: 1, A: 0.33},
		{R: 1, G: 1, B: 0, A: 0.44},
		{R: 1, G: 0, B: 1, A: 0.55},
		{R: 0, G: 1, B: 1, A: 0.66},
	}
	ax.Bar3D(
		[]float64{0},
		[]float64{0},
		[]float64{0},
		[]float64{1},
		[]float64{1},
		[]float64{1},
		Bar3DOptions{Colors: facePattern},
	)
	facePatternFaces := latestBar3DFaceCollection(t, ax, 6)
	for _, color := range facePattern {
		if got, wantCount := countFaceColorAlpha(facePatternFaces.FaceColors, color.A), 1; got != wantCount {
			t.Fatalf("six-face alpha %.2f count = %d, want %d in %+v", color.A, got, wantCount, facePatternFaces.FaceColors)
		}
	}

	allFaces := make([]render.Color, 12)
	for i := range allFaces {
		allFaces[i] = render.Color{R: float64(i) / 12, G: 0.25, B: 0.75, A: 0.05 + float64(i)*0.03}
	}
	ax.Bar3D(
		[]float64{0, 2},
		[]float64{0, 0},
		[]float64{0, 0},
		[]float64{1, 1},
		[]float64{1, 1},
		[]float64{1, 1},
		Bar3DOptions{Colors: allFaces},
	)
	allFaceCollection := latestBar3DFaceCollection(t, ax, 12)
	for _, color := range allFaces {
		if got, wantCount := countFaceColorAlpha(allFaceCollection.FaceColors, color.A), 1; got != wantCount {
			t.Fatalf("6*N face alpha %.2f count = %d, want %d in %+v", color.A, got, wantCount, allFaceCollection.FaceColors)
		}
	}
}

func TestAxes3DBar3DAxLimClipDropsOutsideCuboids(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetXLim(0, 1)

	collection := ax.Bar3D(
		[]float64{0, 2},
		[]float64{0, 0},
		[]float64{0, 0},
		[]float64{0.5, 0.5},
		[]float64{0.5, 0.5},
		[]float64{0.5, 0.5},
		Bar3DOptions{AxLimClip: true},
	)
	if collection == nil {
		t.Fatal("Bar3D returned nil")
	}
	if got, want := len(collection.Segments), 8; got != want {
		t.Fatalf("clipped Bar3D edge segments = %d, want one in-limit cuboid (%d)", got, want)
	}
	foundClippedFaces := false
	for _, artist := range ax.Artists {
		polys, ok := artist.(*PolyCollection)
		if ok && len(polys.Polygons) == 6 {
			foundClippedFaces = true
			break
		}
	}
	if !foundClippedFaces {
		t.Fatal("Bar3D did not add exactly one in-limit cuboid's filled faces")
	}
}

func TestAxes3DTrisurfCreatesSinglePolyCollection(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	tri := Triangulation{
		X:         []float64{0, 1, 1, 0},
		Y:         []float64{0, 0, 1, 1},
		Triangles: [][3]int{{0, 1, 2}, {0, 2, 3}},
	}
	collection := ax.Trisurf(tri, []float64{0, 1, 2, 3})
	if collection == nil {
		t.Fatal("Trisurf returned nil")
	}
	polyCount := 0
	lineCount := 0
	for _, artist := range ax.Artists {
		switch art := artist.(type) {
		case *PolyCollection:
			if len(art.Polygons) == 2 {
				polyCount++
			}
		case *LineCollection:
			lineCount++
		}
	}
	if polyCount != 1 || lineCount != 0 {
		t.Fatalf("Trisurf artists = %d matching PolyCollection, %d LineCollection; want one Poly3DCollection-equivalent and no separate edge collection", polyCount, lineCount)
	}
}

func TestAxes3DTrisurfAutoTriangulatesWhenTrianglesAreOmitted(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	tri := Triangulation{
		X: []float64{0, 1, 1, 0},
		Y: []float64{0, 0, 1, 1},
	}
	collection := ax.Trisurf(tri, []float64{0, 1, 2, 3})
	if collection == nil {
		t.Fatal("Trisurf returned nil")
	}
	if len(collection.Polygons) != 2 {
		t.Fatalf("auto-triangulated polygon count = %d, want 2", len(collection.Polygons))
	}
}

func TestAxes3DTrisurfSkipsMaskedTriangles(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	tri := Triangulation{
		X:         []float64{0, 1, 1, 0},
		Y:         []float64{0, 0, 1, 1},
		Triangles: [][3]int{{0, 1, 2}, {0, 2, 3}},
		Mask:      []bool{false, true},
	}
	collection := ax.Trisurf(tri, []float64{0, 1, 2, 3})
	if collection == nil {
		t.Fatal("Trisurf returned nil")
	}
	if got, want := len(collection.Polygons), 1; got != want {
		t.Fatalf("masked trisurf polygon count = %d, want %d visible triangle", got, want)
	}
}

func TestAxes3DTrisurfUsesConfiguredEdgeColor(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	edge := render.Color{R: 0.2, G: 0.3, B: 0.4, A: 1}
	width := 0.25
	tri := Triangulation{
		X:         []float64{0, 1, 0},
		Y:         []float64{0, 0, 1},
		Triangles: [][3]int{{0, 1, 2}},
	}
	collection := ax.Trisurf(tri, []float64{0, 1, 2}, PlotOptions{
		EdgeColor: &edge,
		EdgeWidth: &width,
	})
	if collection == nil {
		t.Fatal("Trisurf returned nil")
	}
	if got := collection.EdgeColor; got != edge {
		t.Fatalf("trisurf edge color = %+v, want %+v", got, edge)
	}
	if got := collection.EdgeWidth; got != width {
		t.Fatalf("trisurf edge width = %v, want %v", got, width)
	}
}

func TestAxes3DTrisurfShadesFaceColorsLikeMatplotlib(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	base := render.Color{R: 1, G: 0.5, B: 0.05, A: 1}
	tri := Triangulation{
		X:         []float64{0, 1, 0},
		Y:         []float64{0, 0, 1},
		Triangles: [][3]int{{0, 1, 2}},
	}
	collection := ax.Trisurf(tri, []float64{0, 0, 1}, PlotOptions{Color: &base})
	if collection == nil {
		t.Fatal("Trisurf returned nil")
	}
	if got, want := len(collection.FaceColors), 1; got != want {
		t.Fatalf("trisurf face colors = %d, want %d shaded color per face", got, want)
	}
	if collection.FaceColors[0] == base {
		t.Fatalf("trisurf face color = %+v, want Matplotlib-style shaded variant of %+v", collection.FaceColors[0], base)
	}
}

func TestAxes3DVoxelsCullInternalFacesLikeMatplotlib(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	voxels := ax.Voxels([][][]bool{
		{{true}},
		{{true}},
	})
	if got, want := len(voxels), 2; got != want {
		t.Fatalf("voxel collection count = %d, want %d filled voxels", got, want)
	}
	totalFaces := 0
	for _, voxel := range voxels {
		totalFaces += len(voxel.Polygons)
	}
	if got, want := totalFaces, 10; got != want {
		t.Fatalf("visible voxel face count = %d, want %d after internal-face culling", got, want)
	}
}

func TestAxes3DVoxelsApplyFaceEdgeAlphaAndStayNonMappable(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	defaultFace := render.Color{R: 0.2, G: 0.4, B: 0.6, A: 1}
	defaultEdge := render.Color{R: 0.8, G: 0.1, B: 0.3, A: 1}
	overrideFace := render.Color{R: 0.9, G: 0.7, B: 0.2, A: 1}
	overrideEdge := render.Color{R: 0.1, G: 0.8, B: 0.4, A: 1}
	alpha := 0.5
	shade := false

	voxels := ax.Voxels([][][]bool{
		{{true}},
		{{true}},
	}, VoxelOptions{
		FaceColor:  &defaultFace,
		FaceColors: map[[3]int]render.Color{{1, 0, 0}: overrideFace},
		EdgeColor:  &defaultEdge,
		EdgeColors: map[[3]int]render.Color{{1, 0, 0}: overrideEdge},
		Alpha:      &alpha,
		Shade:      &shade,
	})
	if got, want := len(voxels), 2; got != want {
		t.Fatalf("voxel collection count = %d, want %d", got, want)
	}

	wantDefaultFace := defaultFace
	wantDefaultFace.A *= alpha
	wantDefaultEdge := defaultEdge
	wantDefaultEdge.A *= alpha
	assertVoxelCollectionColors(t, voxels[[3]int{0, 0, 0}], wantDefaultFace, wantDefaultEdge)

	wantOverrideFace := overrideFace
	wantOverrideFace.A *= alpha
	wantOverrideEdge := overrideEdge
	wantOverrideEdge.A *= alpha
	assertVoxelCollectionColors(t, voxels[[3]int{1, 0, 0}], wantOverrideFace, wantOverrideEdge)
}

func TestAxes3DVoxelsPerVoxelEdgeColorsEnableEdges(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	edge := render.Color{R: 0.8, G: 0.1, B: 0.3, A: 1}
	shade := false
	voxels := ax.Voxels([][][]bool{{{true}}}, VoxelOptions{
		EdgeColors: map[[3]int]render.Color{{0, 0, 0}: edge},
		Shade:      &shade,
	})
	voxel := voxels[[3]int{0, 0, 0}]
	if voxel == nil {
		t.Fatal("missing voxel collection")
	}
	if got, want := voxel.EdgeColor, edge; got != want {
		t.Fatalf("voxel edge color = %+v, want %+v", got, want)
	}
	if got, want := voxel.EdgeWidth, 1.0; got != want {
		t.Fatalf("voxel edge width = %v, want visible Matplotlib-style per-voxel edges width %v", got, want)
	}
}

func TestAxes3DVoxelsShadeFaceColorsByDefault(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	face := render.Color{R: 0.9, G: 0.7, B: 0.2, A: 1}
	voxels := ax.Voxels([][][]bool{{{true}}}, VoxelOptions{FaceColor: &face})
	voxel := voxels[[3]int{0, 0, 0}]
	if voxel == nil {
		t.Fatal("missing voxel collection")
	}
	shaded := false
	for _, got := range voxel.FaceColors {
		if got.A != face.A {
			t.Fatalf("shaded voxel face alpha = %v, want preserved alpha %v", got.A, face.A)
		}
		if got.R != face.R || got.G != face.G || got.B != face.B {
			shaded = true
		}
	}
	if !shaded {
		t.Fatalf("voxel face colors = %+v, want Matplotlib-style shaded variants of %+v", voxel.FaceColors, face)
	}
}

func TestAxes3DVoxelsAxLimClipDropsOutsideVoxelFaces(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetXLim(0, 1)

	voxels := ax.Voxels([][][]bool{
		{{true}},
		{{true}},
	}, VoxelOptions{AxLimClip: true})
	if got, want := len(voxels), 1; got != want {
		t.Fatalf("clipped voxel collection count = %d, want only the in-limit voxel (%d)", got, want)
	}
	voxel, ok := voxels[[3]int{0, 0, 0}]
	if !ok {
		t.Fatal("missing in-limit voxel collection")
	}
	if got, want := len(voxel.Polygons), 5; got != want {
		t.Fatalf("in-limit voxel visible faces = %d, want 5 after adjacent-face culling", got)
	}
}

func TestAxes3DVoxelsAxLimClipClearsStaleFacesAfterViewLimitChange(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetXLim(0, 2)

	voxels := ax.Voxels([][][]bool{
		{{true}},
		{{true}},
	}, VoxelOptions{AxLimClip: true})
	if got, want := len(voxels), 2; got != want {
		t.Fatalf("initial voxel collection count = %d, want %d", got, want)
	}
	outside := voxels[[3]int{1, 0, 0}]
	if outside == nil || len(outside.Polygons) == 0 {
		t.Fatalf("expected second voxel to start visible, got %+v", outside)
	}

	ax.SetXLim(0, 1)
	if got := len(outside.Polygons); got != 0 {
		t.Fatalf("stale clipped voxel polygons = %d, want cleared after view-limit reprojection", got)
	}
	if got := len(outside.FaceColors); got != 0 {
		t.Fatalf("stale clipped voxel face colors = %d, want cleared after view-limit reprojection", got)
	}
}

func TestAxes3DVoxelsResortFacesAfterViewChange(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	filled := [][][]bool{{{true}}}
	voxels := ax.Voxels(filled)
	voxel := voxels[[3]int{0, 0, 0}]
	if voxel == nil {
		t.Fatal("missing voxel collection")
	}

	ax.SetView(0, 0)
	want := ax.projectVoxelCollections(filled, VoxelOptions{}, 1)[[3]int{0, 0, 0}]
	if len(want.polygons) == 0 {
		t.Fatal("expected projected voxel faces")
	}
	if !pointsEqual(voxel.Polygons[0], want.polygons[0], 1e-12) {
		t.Fatalf("voxel first face after view change = %+v, want depth-sorted face %+v", voxel.Polygons[0], want.polygons[0])
	}
}

func TestAxes3DVoxelCallsBarLikeSegments(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	collection := ax.Voxel(
		[]float64{0, 1},
		[]float64{0, 1},
		[]float64{0, 1},
		[]float64{1, 1},
		[]float64{1, 1},
		[]float64{1, 1},
	)
	if collection == nil {
		t.Fatal("Voxel returned nil")
	}
	if got, want := len(collection.Segments), 16; got != want {
		t.Fatalf("segment count = %d, want %d", got, want)
	}
}
