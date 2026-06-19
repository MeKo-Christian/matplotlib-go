package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
)

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
