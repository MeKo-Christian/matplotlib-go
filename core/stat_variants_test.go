package core

import (
	"testing"

	matcolor "github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestAxesStackPlot_CumulativeLayers(t *testing.T) {
	ax := NewFigure(640, 360).AddAxes(geom.Rect{})

	fills := ax.StackPlot(
		[]float64{0, 1, 2},
		[][]float64{
			{1, 2, 3},
			{4, 5, 6},
		},
		StackPlotOptions{Labels: []string{"a", "b"}},
	)

	if len(fills) != 2 {
		t.Fatalf("got %d fills, want 2", len(fills))
	}
	if fills[0].Label != "a" || fills[1].Label != "b" {
		t.Fatalf("labels = %q, %q", fills[0].Label, fills[1].Label)
	}
	assertFloatSlices(t, "first lower", fills[0].Y1, []float64{0, 0, 0})
	assertFloatSlices(t, "first upper", fills[0].Y2, []float64{1, 2, 3})
	assertFloatSlices(t, "second lower", fills[1].Y1, []float64{1, 2, 3})
	assertFloatSlices(t, "second upper", fills[1].Y2, []float64{5, 7, 9})
}

func TestAxesBoxPlots_CreatesMultipleBoxes(t *testing.T) {
	ax := NewFigure(640, 360).AddAxes(geom.Rect{})
	width := 0.55
	positions := []float64{1.5, 2.5}
	colors := []render.Color{
		{R: 0.25, G: 0.55, B: 0.82, A: 1},
		{R: 0.80, G: 0.45, B: 0.20, A: 1},
	}

	boxes := ax.BoxPlots(
		[][]float64{
			{1, 2, 3},
			{4, 5, 6},
		},
		BoxPlotsOptions{
			Positions: positions,
			Width:     &width,
			Colors:    colors,
		},
	)

	if len(boxes) != 2 {
		t.Fatalf("got %d boxes, want 2", len(boxes))
	}
	if len(ax.Artists) != 2 {
		t.Fatalf("got %d artists, want 2", len(ax.Artists))
	}
	for i, box := range boxes {
		if box.Position != positions[i] {
			t.Fatalf("box %d position = %v, want %v", i, box.Position, positions[i])
		}
		if box.Width != width {
			t.Fatalf("box %d width = %v, want %v", i, box.Width, width)
		}
		if box.Color != colors[i] {
			t.Fatalf("box %d color = %+v, want %+v", i, box.Color, colors[i])
		}
	}
}

func TestAxesBxpCreatesComponentArtistsAndTicks(t *testing.T) {
	ax := NewFigure(640, 360).AddAxes(geom.Rect{})
	mean := 2.2

	container := ax.Bxp(
		[]BxpStat{
			{Med: 2, Q1: 1, Q3: 3, Whislo: 0.5, Whishi: 3.5, Mean: &mean, Fliers: []float64{4.2}, Label: "A"},
			{Med: 5, Q1: 4, Q3: 6, Whislo: 3.5, Whishi: 6.5, Label: "B"},
		},
		BxpOptions{ShowMeans: boolPtr(true), Label: "stats"},
	)

	if container == nil {
		t.Fatal("expected bxp container")
	}
	if len(container.Boxes) != 2 || len(container.Medians) != 2 || len(container.Whiskers) != 4 {
		t.Fatalf("unexpected box component counts boxes=%d medians=%d whiskers=%d", len(container.Boxes), len(container.Medians), len(container.Whiskers))
	}
	if len(container.Caps) != 4 || len(container.Fliers) != 1 || len(container.Means) != 1 {
		t.Fatalf("unexpected optional component counts caps=%d fliers=%d means=%d", len(container.Caps), len(container.Fliers), len(container.Means))
	}
	if got := container.Means[0]; got.Marker != MarkerTriangle || got.MarkerSize != 6 || got.MarkerFaceColor != matcolor.Tab10[2] || got.MarkerEdgeColor != matcolor.Tab10[2] {
		t.Fatalf("mean marker style = marker %v size %v face %+v edge %+v, want Matplotlib default C2 triangle",
			got.Marker, got.MarkerSize, got.MarkerFaceColor, got.MarkerEdgeColor)
	}
	if got := container.Fliers[0]; got.Marker != MarkerCircle || got.MarkerSize != 6 || got.MarkerFaceSpec.Mode != MarkerColorNone || got.MarkerEdgeColor != (render.Color{R: 0, G: 0, B: 0, A: 1}) {
		t.Fatalf("flier marker style = marker %v size %v face mode %v edge %+v, want Matplotlib default open black circle",
			got.Marker, got.MarkerSize, got.MarkerFaceSpec.Mode, got.MarkerEdgeColor)
	}
	if container.Medians[0].Label != "stats" || container.Medians[1].Label != "" {
		t.Fatalf("median labels = %q, %q; want first legend label only", container.Medians[0].Label, container.Medians[1].Label)
	}
	if len(ax.Artists) != 14 {
		t.Fatalf("registered artists = %d, want 14 component Line2D artists", len(ax.Artists))
	}

	loc, ok := ax.XAxis.Locator.(FixedLocator)
	if !ok {
		t.Fatalf("x-axis locator = %T, want FixedLocator", ax.XAxis.Locator)
	}
	assertFloatSlices(t, "bxp ticks", loc.TicksList, []float64{1, 2})
	formatter, ok := ax.XAxis.Formatter.(FixedFormatter)
	if !ok {
		t.Fatalf("x-axis formatter = %T, want FixedFormatter", ax.XAxis.Formatter)
	}
	if len(formatter.Labels) != 2 || formatter.Labels[0] != "A" || formatter.Labels[1] != "B" {
		t.Fatalf("tick labels = %v, want [A B]", formatter.Labels)
	}
}

func TestAxesBxpValidatesOptionLengthsAndHorizontalTicks(t *testing.T) {
	ax := NewFigure(640, 360).AddAxes(geom.Rect{})
	stats := []BxpStat{
		{Med: 2, Q1: 1, Q3: 3, Whislo: 0.5, Whishi: 3.5},
		{Med: 5, Q1: 4, Q3: 6, Whislo: 3.5, Whishi: 6.5},
	}
	if got := ax.Bxp(stats, BxpOptions{Positions: []float64{1}}); got != nil {
		t.Fatalf("Bxp with mismatched positions returned %#v, want nil", got)
	}

	container := ax.Bxp(stats, BxpOptions{
		Positions:   []float64{1.5, 2.5},
		Widths:      []float64{0.4, 0.5},
		Orientation: "horizontal",
	})
	if container == nil {
		t.Fatal("expected horizontal bxp container")
	}
	if got := container.Medians[0].XY; len(got) != 2 || got[0] != (geom.Pt{X: 2, Y: 1.3}) || got[1] != (geom.Pt{X: 2, Y: 1.7}) {
		t.Fatalf("horizontal median data = %+v, want x fixed at median and y spanning box width", got)
	}
	if _, ok := ax.YAxis.Locator.(FixedLocator); !ok {
		t.Fatalf("y-axis locator = %T, want FixedLocator for horizontal Bxp", ax.YAxis.Locator)
	}
}

func TestAxesBoxPlotsManageTicksFalsePreservesLocator(t *testing.T) {
	ax := NewFigure(640, 360).AddAxes(geom.Rect{})
	ax.XAxis.Locator = AutoLocator{}
	manageTicks := false

	ax.BoxPlots(
		[][]float64{
			{1, 2, 3},
			{4, 5, 6},
		},
		BoxPlotsOptions{ManageTicks: &manageTicks},
	)

	if _, ok := ax.XAxis.Locator.(AutoLocator); !ok {
		t.Fatalf("x-axis locator = %T, want preserved AutoLocator when manage_ticks=False", ax.XAxis.Locator)
	}
}

func TestAxesBoxPlotsManageTicksDefaultMatchesMatplotlib(t *testing.T) {
	ax := NewFigure(640, 360).AddAxes(geom.Rect{})

	ax.BoxPlots(
		[][]float64{
			{1, 2, 3},
			{4, 5, 6},
		},
		BoxPlotsOptions{Positions: []float64{1.5, 2.5}},
	)

	loc, ok := ax.XAxis.Locator.(FixedLocator)
	if !ok {
		t.Fatalf("x-axis locator = %T, want FixedLocator by default", ax.XAxis.Locator)
	}
	want := []float64{1.5, 2.5}
	if len(loc.TicksList) != len(want) {
		t.Fatalf("fixed ticks = %v, want %v", loc.TicksList, want)
	}
	for i := range want {
		if loc.TicksList[i] != want[i] {
			t.Fatalf("fixed tick %d = %v, want %v", i, loc.TicksList[i], want[i])
		}
	}
}

func TestAxesBoxPlotsDefaultWidthMatchesMatplotlibPositions(t *testing.T) {
	ax := NewFigure(640, 360).AddAxes(geom.Rect{})

	boxes := ax.BoxPlots([][]float64{
		{1, 2, 3},
		{2, 3, 4},
	})

	if len(boxes) != 2 {
		t.Fatalf("got %d boxes, want 2", len(boxes))
	}
	for i, box := range boxes {
		if box.Width != 0.15 {
			t.Fatalf("box %d width = %v, want Matplotlib default 0.15", i, box.Width)
		}
	}
}

func TestAxesBoxPlotDefaultMedianStyleMatchesMatplotlib(t *testing.T) {
	ax := NewFigure(640, 360).AddAxes(geom.Rect{})

	box := ax.BoxPlot([]float64{1, 2, 3, 4, 5})
	if box == nil {
		t.Fatal("expected box plot")
	}
	wantColor := matcolor.Tab10[1]
	if box.MedianColor != wantColor {
		t.Fatalf("median color = %+v, want Matplotlib boxplot.medianprops.color C1 %+v", box.MedianColor, wantColor)
	}
	if got, want := box.MedianWidth, pointsToPixels(ax.resolvedRC(), 1); got != want {
		t.Fatalf("median width = %v, want Matplotlib boxplot.medianprops.linewidth %v px", got, want)
	}
}

func TestAxesBoxPlotSubArtistsUseMatplotlibPathSnapping(t *testing.T) {
	ax := NewFigure(640, 360).AddAxes(geom.Rect{})
	box := ax.BoxPlot([]float64{1, 2, 3, 4, 5})
	if box == nil {
		t.Fatal("expected box plot")
	}

	renderer := &recordingRenderer{}
	box.Draw(renderer, createTestDrawContext())

	if len(renderer.pathCalls) != 6 {
		t.Fatalf("got %d path calls, want box, whiskers, caps, and median", len(renderer.pathCalls))
	}
	for i, call := range renderer.pathCalls {
		if call.paint.Snap != render.SnapAuto {
			t.Fatalf("path call %d snap = %v, want SnapAuto like Matplotlib PathPatch/Line2D", i, call.paint.Snap)
		}
	}
}

func TestAxesBoxPlotAlphaAppliesOnlyToBoxPatchLikeMatplotlib(t *testing.T) {
	alpha := 0.4
	black := render.Color{A: 1}
	box := &BoxPlot2D{
		Data:           []float64{1, 2, 3, 4, 100},
		Position:       1,
		Width:          0.6,
		Color:          render.Color{R: 0.25, G: 0.55, B: 0.82, A: 1},
		EdgeColor:      black,
		WhiskerColor:   black,
		CapColor:       black,
		MedianColor:    black,
		FlierColor:     black,
		FlierEdgeColor: black,
		EdgeWidth:      1,
		WhiskerWidth:   1,
		MedianWidth:    1,
		FlierEdgeWidth: 1,
		FlierSize:      4,
		Alpha:          alpha,
		PatchArtist:    true,
		ShowBox:        true,
		ShowCaps:       true,
		ShowFliers:     true,
	}

	renderer := &recordingRenderer{}
	box.Draw(renderer, createTestDrawContext())

	if len(renderer.pathCalls) != 7 {
		t.Fatalf("got %d path calls, want box, whiskers, caps, median, and one flier", len(renderer.pathCalls))
	}
	if got := renderer.pathCalls[0].paint.Fill.A; got != alpha {
		t.Fatalf("box fill alpha = %v, want %v", got, alpha)
	}
	if got := renderer.pathCalls[0].paint.Stroke.A; got != alpha {
		t.Fatalf("box edge alpha = %v, want %v", got, alpha)
	}
	for i, call := range renderer.pathCalls[1:] {
		if call.paint.Stroke.A != 1 {
			t.Fatalf("non-box path %d stroke alpha = %v, want opaque like Matplotlib Line2D/fliers", i+1, call.paint.Stroke.A)
		}
	}
	if got := renderer.pathCalls[len(renderer.pathCalls)-1].paint.Fill.A; got != 1 {
		t.Fatalf("flier fill alpha = %v, want opaque like Matplotlib flierprops", got)
	}
}

func TestAxesBoxPlot_AdvancedStatOptions(t *testing.T) {
	ax := NewFigure(640, 360).AddAxes(geom.Rect{})
	notch := true
	ci := [2]float64{2.8, 3.2}
	median := 3.0
	whis := [2]float64{0, 100}
	marker := MarkerDiamond
	edge := render.Color{R: 0.2, G: 0.1, B: 0.8, A: 1}
	edgeWidth := 1.75

	box := ax.BoxPlot([]float64{1, 2, 3, 4, 100}, BoxPlotOptions{
		Notch:              &notch,
		ConfidenceInterval: &ci,
		CustomMedian:       &median,
		WhiskerPercentiles: &whis,
		FlierMarker:        &marker,
		FlierEdgeColor:     &edge,
		FlierEdgeWidth:     &edgeWidth,
	})
	if box == nil {
		t.Fatal("expected box plot")
	}
	box.ensureComputed()
	if !box.Notch || box.stats.median != median || box.stats.ciLow != ci[0] || box.stats.ciHigh != ci[1] {
		t.Fatalf("advanced stats not applied: notch=%v stats=%+v", box.Notch, box.stats)
	}
	if box.stats.lowerWhisker != 1 || box.stats.upperWhisker != 100 || len(box.stats.outliers) != 0 {
		t.Fatalf("percentile whiskers not applied: %+v", box.stats)
	}
	if box.FlierMarker != marker || box.FlierEdgeColor != edge || box.FlierEdgeWidth != edgeWidth {
		t.Fatalf("flier customization not applied: marker=%v edge=%+v width=%v", box.FlierMarker, box.FlierEdgeColor, box.FlierEdgeWidth)
	}
}

func TestAxesBoxPlot_PercentileWhiskersUseNearestInlier(t *testing.T) {
	ax := NewFigure(640, 360).AddAxes(geom.Rect{})
	whis := [2]float64{5, 95}

	box := ax.BoxPlot([]float64{1.1, 1.8, 2.2, 2.6, 2.9, 3.1, 3.7, 6.8}, BoxPlotOptions{
		WhiskerPercentiles: &whis,
	})
	if box == nil {
		t.Fatal("expected box plot")
	}
	box.ensureComputed()
	if box.stats.lowerWhisker != 1.8 || box.stats.upperWhisker != 3.7 {
		t.Fatalf("whiskers = %.3f, %.3f; want nearest inliers 1.8, 3.7", box.stats.lowerWhisker, box.stats.upperWhisker)
	}
	if len(box.stats.outliers) != 2 || box.stats.outliers[0] != 1.1 || box.stats.outliers[1] != 6.8 {
		t.Fatalf("outliers = %v; want [1.1 6.8]", box.stats.outliers)
	}
}

func TestAxesECDF_ComputesSortedStepData(t *testing.T) {
	ax := NewFigure(640, 360).AddAxes(geom.Rect{})

	line := ax.ECDF([]float64{3, 1, 2, 2})
	if line == nil {
		t.Fatal("expected ECDF line")
	}
	if line.DrawStyle != LineDrawStyleStepsPost {
		t.Fatalf("draw style = %v, want steps-post", line.DrawStyle)
	}
	want := []geom.Pt{
		{X: 1, Y: 0},
		{X: 1, Y: 0.25},
		{X: 2, Y: 0.5},
		{X: 2, Y: 0.75},
		{X: 3, Y: 1},
	}
	if len(line.XY) != len(want) {
		t.Fatalf("got %d points, want %d", len(line.XY), len(want))
	}
	for i := range want {
		if line.XY[i] != want[i] {
			t.Fatalf("point %d = %+v, want %+v", i, line.XY[i], want[i])
		}
	}
}

func TestAxesECDF_CompressMatchesMatplotlibFirstDuplicate(t *testing.T) {
	ax := NewFigure(640, 360).AddAxes(geom.Rect{})

	line := ax.ECDF([]float64{3, 1, 2, 2}, ECDFOptions{Compress: true})
	if line == nil {
		t.Fatal("expected ECDF line")
	}
	want := []geom.Pt{
		{X: 1, Y: 0},
		{X: 1, Y: 0.25},
		{X: 2, Y: 0.5},
		{X: 3, Y: 1},
	}
	if len(line.XY) != len(want) {
		t.Fatalf("got %d points, want %d", len(line.XY), len(want))
	}
	for i := range want {
		if line.XY[i] != want[i] {
			t.Fatalf("point %d = %+v, want %+v", i, line.XY[i], want[i])
		}
	}
}

func TestHist2D_CumulativeProbability(t *testing.T) {
	hist := &Hist2D{
		Data:       []float64{0.2, 0.4, 1.2, 2.2},
		BinEdges:   []float64{0, 1, 2, 3},
		Norm:       HistNormProbability,
		Cumulative: true,
	}

	_, counts := hist.BinCounts()
	assertFloatSlices(t, "counts", counts, []float64{0.5, 0.75, 1})
}

func TestHist2D_StepFilledDrawsClosedPath(t *testing.T) {
	hist := &Hist2D{
		Data:      []float64{0.2, 0.4, 1.2},
		BinEdges:  []float64{0, 1, 2},
		HistType:  HistTypeStepFilled,
		Color:     render.Color{R: 0.2, G: 0.4, B: 0.8, A: 1},
		EdgeColor: render.Color{R: 0, G: 0, B: 0, A: 1},
		EdgeWidth: 1,
	}

	r := &recordingRenderer{}
	hist.Draw(r, createTestDrawContext())

	if len(r.pathCalls) != 1 {
		t.Fatalf("expected one path call, got %d", len(r.pathCalls))
	}
	call := r.pathCalls[0]
	if call.paint.Fill.A == 0 || call.paint.Stroke.A == 0 {
		t.Fatalf("unexpected paint = %+v", call.paint)
	}
	if call.paint.Snap != render.SnapAuto {
		t.Fatalf("stepfilled snap = %v, want Matplotlib patch SnapAuto", call.paint.Snap)
	}
	if len(call.path.C) == 0 || call.path.C[len(call.path.C)-1] != geom.ClosePath {
		t.Fatalf("expected closed path, got %+v", call.path.C)
	}
}

func TestAxesHistMulti_UsesSharedEdgesAndStackedBaselines(t *testing.T) {
	ax := NewFigure(640, 360).AddAxes(geom.Rect{})

	hists := ax.HistMulti(
		[][]float64{
			{0.2, 0.4, 1.2},
			{0.7, 1.4, 1.6},
		},
		MultiHistOptions{
			BinEdges: []float64{0, 1, 2},
			Stacked:  true,
		},
	)

	if len(hists) != 2 {
		t.Fatalf("got %d hists, want 2", len(hists))
	}
	edges0, counts0 := hists[0].BinCounts()
	edges1, _ := hists[1].BinCounts()
	assertFloatSlices(t, "edges0", edges0, []float64{0, 1, 2})
	assertFloatSlices(t, "edges1", edges1, []float64{0, 1, 2})
	assertFloatSlices(t, "baselines", hists[1].Baselines, counts0)
}

func assertFloatSlices(t *testing.T, name string, got, want []float64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s length = %d, want %d (%v)", name, len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s[%d] = %v, want %v (all %v)", name, i, got[i], want[i], got)
		}
	}
}
