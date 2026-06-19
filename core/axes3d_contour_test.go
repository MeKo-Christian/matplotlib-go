package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestAxes3DContourAndContourfCreateCollections(t *testing.T) {
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
	contour := ax.Contour(x, y, z)
	if contour == nil {
		t.Fatal("Contour returned nil")
	}
	if contourf := ax.Contourf(x, y, z); contourf == nil {
		t.Fatal("Contourf returned nil")
	}
}

func TestAxes3DTriContourAndTriContourfCreateCollections(t *testing.T) {
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
	values := []float64{0, 1, 1}
	contour := ax.TriContour(tri, values, PlotOptions{Levels: []float64{0.5}})
	if contour == nil {
		t.Fatal("TriContour returned nil")
	}
	if contourf := ax.TriContourf(tri, values, PlotOptions{Levels: []float64{0, 0.5, 1}}); contourf == nil {
		t.Fatal("TriContourf returned nil")
	}
	if invalid := ax.TriContour(tri, values[:2], PlotOptions{Levels: []float64{0.5}}); invalid != nil {
		t.Fatal("TriContour accepted mismatched value length")
	}
}

func TestAxes3DContourfDefaultsToOpaqueFacesLikeMatplotlib(t *testing.T) {
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
	values := []float64{0, 1, 1}
	triFilled := ax.TriContourf(tri, values, PlotOptions{Levels: []float64{0, 0.5, 1}})
	if triFilled == nil {
		t.Fatal("TriContourf returned nil")
	}
	for i, c := range triFilled.FaceColors {
		if c.A != 1 {
			t.Fatalf("TriContourf face color %d alpha = %v, want 1 (matplotlib draws filled contours opaque by default)", i, c.A)
		}
	}

	x := []float64{0, 1, 2}
	y := []float64{0, 1, 2}
	z := [][]float64{{0, 0.5, 1}, {0.5, 1, 1.5}, {1, 1.5, 2}}
	filled := ax.Contourf(x, y, z, PlotOptions{Levels: []float64{0, 1, 2}})
	if filled == nil {
		t.Fatal("Contourf returned nil")
	}
	for i, c := range filled.FaceColors {
		if c.A != 1 {
			t.Fatalf("Contourf face color %d alpha = %v, want 1 (matplotlib draws filled contours opaque by default)", i, c.A)
		}
	}
}

func TestAxes3DTriContourProjectsLinesAtContourLevels(t *testing.T) {
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
	contour := ax.TriContour(tri, []float64{0, 1, 1}, PlotOptions{Levels: []float64{0.5}})
	if contour == nil {
		t.Fatal("TriContour returned nil")
	}
	if got, want := len(contour.Segments), 1; got != want {
		t.Fatalf("tricontour segments = %d, want %d", got, want)
	}
	want := []Pt{
		ax.ProjectPoint(0.5, 0, 0.5),
		ax.ProjectPoint(0, 0.5, 0.5),
	}
	if !pointsEqual(contour.Segments[0], want, 1e-12) {
		t.Fatalf("tricontour segment = %+v, want projected contour level %+v", contour.Segments[0], want)
	}
}

func TestAxes3DTriContourfAutoscaleUsesFilledLevelMidpointsLikeMatplotlib(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	tri := Triangulation{
		X:         []float64{0, 1, 0, 1},
		Y:         []float64{0, 0, 1, 1},
		Triangles: [][3]int{{0, 1, 2}, {1, 3, 2}},
	}
	fill := ax.TriContourf(
		tri,
		[]float64{0.1, 0.2, 0.8, 0.9},
		PlotOptions{Levels: []float64{0, 1, 2}},
	)
	if fill == nil {
		t.Fatal("TriContourf returned nil")
	}
	mins, maxs := ax.projectionLimits()
	if !approx(mins[2], 0.4791666666666667, 1e-12) || !approx(maxs[2], 1.5208333333333333, 1e-12) {
		t.Fatalf("TriContourf projection z limits = %.12g..%.12g, want Matplotlib autoscale from filled midpoints plus 3D view margin", mins[2], maxs[2])
	}
}

func TestAxes3DContourfProjectsFilledContourBands(t *testing.T) {
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
	fill := ax.Contourf(x, y, z)
	if fill == nil {
		t.Fatal("Contourf returned nil")
	}
	if got, cellCount := len(fill.Paths), 1; got <= cellCount {
		t.Fatalf("Contourf compound path count = %d, want filled contour band paths rather than %d grid cell", got, cellCount)
	}
}

func TestAxes3DContourfUsesExplicitZOffset(t *testing.T) {
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
	offset := -3.0
	fill := ax.Contourf(x, y, z, PlotOptions{LevelCount: 3, Offset: &offset})
	if fill == nil || len(fill.Paths) == 0 || len(fill.Paths[0].V) == 0 {
		t.Fatalf("Contourf returned no polygons: %+v", fill)
	}

	values := flattenGridValues(z)
	levels := contourLevels(values, nil, 3, true)
	mapping := resolveScalarMapValues(values, "viridis", nil, nil)
	mapping.VMin = levels[0]
	mapping.VMax = levels[len(levels)-1]
	rawPolygons, _ := contourGridBandPolygons(x, y, z, levels, ContourOptions{}, mapping, 0.45)
	if len(rawPolygons) == 0 || len(rawPolygons[0]) == 0 {
		t.Fatal("expected raw contour band polygons")
	}
	want := ax.ProjectPoint(rawPolygons[0][0].X, rawPolygons[0][0].Y, offset)
	if got := fill.Paths[0].V[0]; !approx(got.X, want.X, 1e-12) || !approx(got.Y, want.Y, 1e-12) {
		t.Fatalf("Contourf first point = %+v, want projection at explicit offset %+v", got, want)
	}
}

func TestAxes3DContourfUsesProjectedCollectionZLikeMatplotlib(t *testing.T) {
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
	offset := -3.0
	levelCount := 3
	fill := ax.Contourf(x, y, z, PlotOptions{LevelCount: levelCount, Offset: &offset})
	if fill == nil {
		t.Fatal("Contourf returned nil")
	}

	values := flattenGridValues(z)
	levels := contourLevels(values, nil, levelCount, true)
	mapping := resolveScalarMapValues(values, "viridis", nil, nil)
	mapping.VMin = levels[0]
	mapping.VMax = levels[len(levels)-1]
	rawPolygons, _ := contourGridBandPolygons(x, y, z, levels, ContourOptions{}, mapping, 0.45)
	depth := 0.0
	first := true
	for _, polygon := range rawPolygons {
		for _, pt := range polygon {
			_, zDepth := ax.projectPointDepth(pt.X, pt.Y, offset)
			if first || zDepth < depth {
				depth = zDepth
				first = false
			}
		}
	}
	if first {
		t.Fatal("expected raw contour band polygons")
	}
	want := computed3DCollectionZ(depth)
	if got := fill.Z(); !approx(got, want, 1e-12) {
		t.Fatalf("Contourf zorder = %.12g, want computed projected zorder %.12g like Matplotlib Collection3D", got, want)
	}
}

func TestAxes3DContourfAutoscaleUsesFilledLevelMidpointsLikeMatplotlib(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	fill := ax.Contourf(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0.1, 0.2}, {0.8, 0.9}},
		PlotOptions{Levels: []float64{0, 1, 2}},
	)
	if fill == nil {
		t.Fatal("Contourf returned nil")
	}
	mins, maxs := ax.projectionLimits()
	if !approx(mins[2], 0.4791666666666667, 1e-12) || !approx(maxs[2], 1.5208333333333333, 1e-12) {
		t.Fatalf("Contourf projection z limits = %.12g..%.12g, want Matplotlib autoscale from filled midpoints plus 3D view margin", mins[2], maxs[2])
	}
}

func TestAxes3DContourfUsesStructuredGridBandPolygons(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	offset := -1.0
	fill := ax.Contourf(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 1}, {1, 2}},
		PlotOptions{Levels: []float64{0.5, 1.5}, Offset: &offset},
	)
	if fill == nil {
		t.Fatal("Contourf returned nil")
	}
	if got, want := len(fill.Paths), 1; got != want {
		t.Fatalf("Contourf paths = %d, want one structured quad band path", got)
	}
}

func TestAxes3DContourfGroupsBandsIntoCompoundPathsLikeMatplotlib(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	fill := ax.Contourf(
		[]float64{0, 1, 2},
		[]float64{0, 1, 2},
		[][]float64{
			{0, 1, 0},
			{1, 2, 1},
			{0, 1, 0},
		},
		PlotOptions{Levels: []float64{0.5, 1.5}},
	)
	if fill == nil {
		t.Fatal("Contourf returned nil")
	}
	if got, want := len(fill.Paths), 1; got != want {
		t.Fatalf("Contourf paths = %d, want one compound path per filled contour band like Matplotlib", got)
	}
	if len(fill.Paths[0].C) == 0 || len(fill.Paths[0].V) <= 4 {
		t.Fatalf("Contourf compound path = %+v, want multiple cell polygons grouped into one path", fill.Paths[0])
	}
}

func TestCompoundContourPathsDissolvesSharedBandEdgesLikeMatplotlib(t *testing.T) {
	color := render.Color{R: 0.2, G: 0.4, B: 0.6, A: 1}
	paths, colors := compoundContourPaths(
		[][]geom.Pt{
			{
				{X: 0, Y: 0},
				{X: 1, Y: 0},
				{X: 1, Y: 1},
				{X: 0, Y: 1},
			},
			{
				{X: 1, Y: 0},
				{X: 2, Y: 0},
				{X: 2, Y: 1},
				{X: 1, Y: 1},
			},
		},
		[]render.Color{color, color},
	)
	if got, want := len(paths), 1; got != want {
		t.Fatalf("compound paths = %d, want one path for the filled band", got)
	}
	if got, want := len(colors), 1; got != want || colors[0] != color {
		t.Fatalf("compound colors = %+v, want [%+v]", colors, color)
	}
	moveCount := 0
	closeCount := 0
	for _, cmd := range paths[0].C {
		switch cmd {
		case geom.MoveTo:
			moveCount++
		case geom.ClosePath:
			closeCount++
		}
	}
	if moveCount != 1 || closeCount != 1 {
		t.Fatalf("compound path commands = %+v, want one closed region without the shared cell edge", paths[0].C)
	}
}

func TestAxes3DContourProjectsLinesAtContourLevels(t *testing.T) {
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
	ax.observe3DGrid(x, y, z)

	levelCount := 3
	got := ax.projectedContourSegments(x, y, z, levelCount)
	values := flattenGridValues(z)
	levels := contourLevels(values, nil, levelCount, false)
	rawLines, rawLevels := contourGridPolylines(x, y, z, levels)
	want := make([][]Pt, len(rawLines))
	for i, line := range rawLines {
		want[i] = make([]Pt, len(line))
		for j, pt := range line {
			want[i][j] = ax.ProjectPoint(pt.X, pt.Y, rawLevels[i])
		}
	}
	if len(got) != len(want) {
		t.Fatalf("contour segment count = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if !pointsEqual(got[i], want[i], 1e-12) {
			t.Fatalf("contour segment %d = %+v, want x/y contour projected at level z %+v", i, got[i], want[i])
		}
	}
}

func TestAxes3DContourUsesLevelColorsByDefault(t *testing.T) {
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
	contour := ax.Contour(x, y, z, PlotOptions{LevelCount: 4})
	if contour == nil {
		t.Fatal("Contour returned nil")
	}
	if len(contour.Colors) != len(contour.Segments) {
		t.Fatalf("contour colors = %d, segments = %d; want per-level colormapped colors by default", len(contour.Colors), len(contour.Segments))
	}
	if len(contour.Colors) > 1 && contour.Colors[0] == contour.Colors[len(contour.Colors)-1] {
		t.Fatalf("first and last contour colors are both %+v, want level-dependent colors", contour.Colors[0])
	}
}

func TestAxes3DContourExposesConfiguredScalarMapForColorbars(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	cmap := "magma"
	vmin := 0.0
	vmax := 10.0
	levels := []float64{2, 4}
	contour := ax.Contour(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 2}, {4, 6}},
		PlotOptions{
			Colormap: &cmap,
			VMin:     &vmin,
			VMax:     &vmax,
			Levels:   levels,
		},
	)
	if contour == nil {
		t.Fatal("Contour returned nil")
	}
	mapping := contour.ScalarMap()
	if mapping.Colormap != cmap || mapping.VMin != vmin || mapping.VMax != vmax {
		t.Fatalf("contour scalar map = %+v, want cmap=%q range %.1f..%.1f", mapping, cmap, vmin, vmax)
	}
	array := contour.GetArray()
	if len(array) != len(levels) {
		t.Fatalf("contour scalar array = %v, want Matplotlib line contour cvalues %v", array, levels)
	}
	for i := range levels {
		if !approx(array[i], levels[i], 1e-12) {
			t.Fatalf("contour scalar array = %v, want Matplotlib line contour cvalues %v", array, levels)
		}
	}
}

func TestAxes3DContourExplicitColorDisablesScalarMapStateLikeMatplotlib(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	override := render.Color{R: 0.1, G: 0.2, B: 0.3, A: 1}
	contour := ax.Contour(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 2}, {4, 6}},
		PlotOptions{Color: &override, LevelCount: 3},
	)
	if contour == nil {
		t.Fatal("Contour returned nil")
	}
	if contour.Colormap != "" || contour.Norm != nil || contour.VMin != 0 || contour.VMax != 0 {
		t.Fatalf("contour scalar map state = cmap=%q norm=%T vmin=%g vmax=%g, want no scalar-map metadata", contour.Colormap, contour.Norm, contour.VMin, contour.VMax)
	}
	if array := contour.GetArray(); len(array) != 0 {
		t.Fatalf("explicit-color contour scalar array = %v, want no mappable array", array)
	}
}

func TestAxes3DTriContourExposesLevelArrayForColorbars(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	cmap := "magma"
	vmin := 0.0
	vmax := 1.0
	tri := Triangulation{
		X:         []float64{0, 1, 0},
		Y:         []float64{0, 0, 1},
		Triangles: [][3]int{{0, 1, 2}},
	}
	levels := []float64{0.25, 0.5, 0.75}
	contour := ax.TriContour(
		tri,
		[]float64{0, 1, 1},
		PlotOptions{
			Colormap: &cmap,
			VMin:     &vmin,
			VMax:     &vmax,
			Levels:   levels,
		},
	)
	if contour == nil {
		t.Fatal("TriContour returned nil")
	}
	mapping := contour.ScalarMap()
	if mapping.Colormap != cmap || mapping.VMin != vmin || mapping.VMax != vmax {
		t.Fatalf("tricontour scalar map = %+v, want cmap=%q range %.1f..%.1f", mapping, cmap, vmin, vmax)
	}
	array := contour.GetArray()
	if len(array) != len(levels) {
		t.Fatalf("tricontour scalar array = %v, want Matplotlib line contour cvalues %v", array, levels)
	}
	for i := range levels {
		if !approx(array[i], levels[i], 1e-12) {
			t.Fatalf("tricontour scalar array = %v, want Matplotlib line contour cvalues %v", array, levels)
		}
	}
}

func TestAxes3DTriContourExplicitColorDisablesScalarMapStateLikeMatplotlib(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	override := render.Color{R: 0.1, G: 0.2, B: 0.3, A: 1}
	tri := Triangulation{
		X:         []float64{0, 1, 0},
		Y:         []float64{0, 0, 1},
		Triangles: [][3]int{{0, 1, 2}},
	}
	contour := ax.TriContour(tri, []float64{0, 1, 1}, PlotOptions{
		Color:  &override,
		Levels: []float64{0.5},
	})
	if contour == nil {
		t.Fatal("TriContour returned nil")
	}
	if contour.Colormap != "" || contour.Norm != nil || contour.VMin != 0 || contour.VMax != 0 {
		t.Fatalf("explicit-color tricontour scalar map state = cmap=%q norm=%T vmin=%g vmax=%g, want no scalar-map metadata", contour.Colormap, contour.Norm, contour.VMin, contour.VMax)
	}
	if array := contour.GetArray(); len(array) != 0 {
		t.Fatalf("explicit-color tricontour scalar array = %v, want no mappable array", array)
	}
}

func TestAxes3DContourfExposesConfiguredScalarMapForColorbars(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	cmap := "plasma"
	vmin := 0.0
	vmax := 12.0
	levels := []float64{0, 2, 4, 6}
	contour := ax.Contourf(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 2}, {4, 6}},
		PlotOptions{
			Colormap: &cmap,
			VMin:     &vmin,
			VMax:     &vmax,
			Levels:   levels,
		},
	)
	if contour == nil {
		t.Fatal("Contourf returned nil")
	}
	mapping := contour.ScalarMap()
	if mapping.Colormap != cmap || mapping.VMin != vmin || mapping.VMax != vmax {
		t.Fatalf("contourf scalar map = %+v, want cmap=%q range %.1f..%.1f", mapping, cmap, vmin, vmax)
	}
	array := contour.GetArray()
	want := []float64{1, 3, 5}
	if len(array) != len(want) {
		t.Fatalf("contourf scalar array = %v, want Matplotlib filled contour layer values %v", array, want)
	}
	for i := range want {
		if !approx(array[i], want[i], 1e-12) {
			t.Fatalf("contourf scalar array = %v, want Matplotlib filled contour layer values %v", array, want)
		}
	}
}

func TestAxes3DContourfExplicitColorDisablesScalarMapStateLikeMatplotlib(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	override := render.Color{R: 0.1, G: 0.2, B: 0.3, A: 1}
	contour := ax.Contourf(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 2}, {4, 6}},
		PlotOptions{Color: &override, LevelCount: 3},
	)
	if contour == nil {
		t.Fatal("Contourf returned nil")
	}
	if contour.Colormap != "" || contour.Norm != nil || contour.VMin != 0 || contour.VMax != 0 {
		t.Fatalf("contourf scalar map state = cmap=%q norm=%T vmin=%g vmax=%g, want no scalar-map metadata", contour.Colormap, contour.Norm, contour.VMin, contour.VMax)
	}
	if array := contour.GetArray(); len(array) != 0 {
		t.Fatalf("explicit-color contourf scalar array = %v, want no mappable array", array)
	}
}

func TestAxes3DTriContourfExposesLayerArrayForColorbars(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	cmap := "plasma"
	vmin := 0.0
	vmax := 6.0
	tri := Triangulation{
		X:         []float64{0, 1, 0, 1},
		Y:         []float64{0, 0, 1, 1},
		Triangles: [][3]int{{0, 1, 2}, {1, 3, 2}},
	}
	levels := []float64{0, 2, 4, 6}
	contour := ax.TriContourf(
		tri,
		[]float64{0, 2, 4, 6},
		PlotOptions{
			Colormap: &cmap,
			VMin:     &vmin,
			VMax:     &vmax,
			Levels:   levels,
		},
	)
	if contour == nil {
		t.Fatal("TriContourf returned nil")
	}
	mapping := contour.ScalarMap()
	if mapping.Colormap != cmap || mapping.VMin != vmin || mapping.VMax != vmax {
		t.Fatalf("tricontourf scalar map = %+v, want cmap=%q range %.1f..%.1f", mapping, cmap, vmin, vmax)
	}
	array := contour.GetArray()
	want := []float64{1, 3, 5}
	if len(array) != len(want) {
		t.Fatalf("tricontourf scalar array = %v, want Matplotlib filled contour layer values %v", array, want)
	}
	for i := range want {
		if !approx(array[i], want[i], 1e-12) {
			t.Fatalf("tricontourf scalar array = %v, want Matplotlib filled contour layer values %v", array, want)
		}
	}
}

func TestAxes3DTriContourfExplicitColorDisablesScalarMapStateLikeMatplotlib(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	override := render.Color{R: 0.1, G: 0.2, B: 0.3, A: 1}
	tri := Triangulation{
		X:         []float64{0, 1, 0, 1},
		Y:         []float64{0, 0, 1, 1},
		Triangles: [][3]int{{0, 1, 2}, {1, 3, 2}},
	}
	contour := ax.TriContourf(tri, []float64{0, 2, 4, 6}, PlotOptions{
		Color:  &override,
		Levels: []float64{0, 2, 4, 6},
	})
	if contour == nil {
		t.Fatal("TriContourf returned nil")
	}
	if contour.Colormap != "" || contour.Norm != nil || contour.VMin != 0 || contour.VMax != 0 {
		t.Fatalf("explicit-color tricontourf scalar map state = cmap=%q norm=%T vmin=%g vmax=%g, want no scalar-map metadata", contour.Colormap, contour.Norm, contour.VMin, contour.VMax)
	}
	if array := contour.GetArray(); len(array) != 0 {
		t.Fatalf("explicit-color tricontourf scalar array = %v, want no mappable array", array)
	}
}

func TestAxes3DContourSupportsMatplotlibZDirJuggling(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetDistance(0)
	ax.SetView(0, 0)

	contour := ax.Contour(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 1}, {1, 2}},
		PlotOptions{Levels: []float64{0.5}, ZDir: "x"},
	)
	if contour == nil {
		t.Fatal("Contour returned nil")
	}
	if got, want := len(contour.Segments), 1; got != want {
		t.Fatalf("contour segment count = %d, want %d", got, want)
	}
	want := []Pt{
		ax.ProjectPoint(0.5, 0, 0.5),
		ax.ProjectPoint(0.5, 0.5, 1),
		ax.ProjectPoint(0.5, 1, 1.5),
	}
	if !pointsEqual(contour.Segments[0], want, 1e-12) {
		t.Fatalf("x-directed contour = %+v, want Matplotlib rotate_axes/juggle_axes contour %+v", contour.Segments[0], want)
	}
}

func TestAxes3DContourUsesExplicitOffsetPlane(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	offset := -2.0
	contour := ax.Contour(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 1}, {1, 2}},
		PlotOptions{Levels: []float64{1}, Offset: &offset},
	)
	if contour == nil || len(contour.Segments) == 0 || len(contour.Segments[0]) == 0 {
		t.Fatalf("Contour returned no segments: %+v", contour)
	}

	rawLines, _ := contourGridPolylines([]float64{0, 1}, []float64{0, 1}, [][]float64{{0, 1}, {1, 2}}, []float64{1})
	if len(rawLines) == 0 || len(rawLines[0]) == 0 {
		t.Fatal("expected raw contour line")
	}
	want := ax.ProjectPoint(rawLines[0][0].X, rawLines[0][0].Y, offset)
	if got := contour.Segments[0][0]; !approx(got.X, want.X, 1e-12) || !approx(got.Y, want.Y, 1e-12) {
		t.Fatalf("contour offset point = %+v, want explicit offset projection %+v", got, want)
	}
}

func TestAxes3DContourAxLimClipUsesExplicit3DLimits(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetDistance(0)
	ax.SetView(0, 0)
	ax.SetXLim(0, 0.75)
	ax.SetYLim(0, 1)

	x := []float64{0, 0.5, 1}
	y := []float64{0, 0.5, 1}
	z := [][]float64{
		{0, 0.5, 1},
		{0.5, 1, 1.5},
		{1, 1.5, 2},
	}
	contour := ax.Contour(x, y, z, PlotOptions{Levels: []float64{1}, AxLimClip: true})
	if contour == nil {
		t.Fatal("Contour returned nil")
	}
	if got, want := len(contour.Segments), 1; got != want {
		t.Fatalf("contour clipped segment count = %d, want %d", got, want)
	}

	rawLines, _ := contourGridPolylines(x, y, z, []float64{1})
	if len(rawLines) != 1 {
		t.Fatalf("raw contour lines = %d, want 1", len(rawLines))
	}
	wantRuns := ax.clip3DPolylineRuns(contourPolyline3D(rawLines[0], 1, "z"))
	if len(wantRuns) != 1 {
		t.Fatalf("clipped contour runs = %d, want 1", len(wantRuns))
	}
	want := make([]Pt, len(wantRuns[0]))
	for i, point := range wantRuns[0] {
		want[i] = ax.ProjectPoint(point[0], point[1], point[2])
	}
	if !pointsEqual(contour.Segments[0], want, 1e-12) {
		t.Fatalf("contour clipped segment = %+v, want explicit-view-limit run %+v", contour.Segments[0], want)
	}
}

func TestAxes3DContourZOrderUsesContourGeometry(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	x := []float64{-1, 0, 1}
	y := []float64{-1, 0, 1}
	z := [][]float64{
		{0.2, 0.6, 0.2},
		{0.6, 1.0, 0.6},
		{0.2, 0.6, 0.2},
	}
	surface := ax.Surface(x, y, z)
	contour := ax.Contour(x, y, z, PlotOptions{LevelCount: 4})
	if surface == nil || contour == nil {
		t.Fatalf("expected surface and contour collections, got surface=%v contour=%v", surface, contour)
	}
	if !(surface.Z() > contour.Z()) {
		t.Fatalf("surface zorder %.6g, contour zorder %.6g; want surface drawn over 3D contour lines like Matplotlib computed_zorder", surface.Z(), contour.Z())
	}
}

func TestAxes3DContourfUsesFilledLevelMidpointsByDefault(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	fill := ax.Contourf(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 1}, {1, 2}},
		PlotOptions{Levels: []float64{0, 1, 2}},
	)
	if fill == nil || len(fill.Paths) == 0 || len(fill.Paths[0].V) == 0 {
		t.Fatalf("Contourf returned no paths: %+v", fill)
	}

	rawPolygons := contourGridBandPolygonsForLevel([]float64{0, 1}, []float64{0, 1}, [][]float64{{0, 1}, {1, 2}}, 0, 1)
	if len(rawPolygons) == 0 || len(rawPolygons[0]) == 0 {
		t.Fatal("expected raw first-band contour polygon")
	}
	want := ax.ProjectPoint(rawPolygons[0][0].X, rawPolygons[0][0].Y, 0.5)
	if got := fill.Paths[0].V[0]; !approx(got.X, want.X, 1e-12) || !approx(got.Y, want.Y, 1e-12) {
		t.Fatalf("contourf midpoint point = %+v, want projection at first-band midpoint %+v", got, want)
	}
}

func TestAxes3DContourfAxLimClipDropsOffsetBandsOutsideExplicitZLimits(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetZLim(0, 0.75)

	offset := 1.0
	fill := ax.Contourf(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 1}, {1, 2}},
		PlotOptions{Levels: []float64{0, 1, 2}, Offset: &offset, AxLimClip: true},
	)
	if fill != nil {
		t.Fatalf("Contourf with offset plane outside z limits returned %+v, want nil", fill)
	}
}
