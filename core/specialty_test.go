package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestAxesEventplotBuildsSegments(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	events := ax.Eventplot([][]float64{
		{0.5, 1.5},
		{2.0},
	}, EventPlotOptions{
		LineOffsets: []float64{1, 3},
		LineLengths: []float64{0.4, 1.2},
		Label:       "events",
	})
	if events == nil {
		t.Fatal("expected event collection")
	}
	if len(events.Segments) != 3 {
		t.Fatalf("len(events.Segments) = %d, want 3", len(events.Segments))
	}
	if got := events.Segments[0][0]; got != (geom.Pt{X: 0.5, Y: 0.8}) {
		t.Fatalf("first segment start = %+v, want {0.5 0.8}", got)
	}
	if got := events.Segments[2][1]; got != (geom.Pt{X: 2.0, Y: 3.6}) {
		t.Fatalf("third segment end = %+v, want {2.0 3.6}", got)
	}
}

func TestAxesHexbinAggregatesValues(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	hex := ax.Hexbin(
		[]float64{0.1, 0.2, 0.85},
		[]float64{0.1, 0.2, 0.8},
		HexbinOptions{
			GridSizeX: 2,
			GridSizeY: 2,
			C:         []float64{2, 4, 9},
			Reduce:    "mean",
			Extent: &geom.Rect{
				Min: geom.Pt{X: 0, Y: 0},
				Max: geom.Pt{X: 1, Y: 1},
			},
			Label: "hex",
		},
	)
	if hex == nil {
		t.Fatal("expected hexbin collection")
	}
	if len(hex.Values) != 3 {
		t.Fatalf("len(hex.Values) = %d, want 3", len(hex.Values))
	}
	if hex.Values[0] != 2 || hex.Values[1] != 4 || hex.Values[2] != 9 {
		t.Fatalf("unexpected values %v", hex.Values)
	}
	if hex.Counts[0] != 1 || hex.Counts[1] != 1 || hex.Counts[2] != 1 {
		t.Fatalf("unexpected counts %v", hex.Counts)
	}
	if !floatApprox(hex.BinCenters[1].X, 0.25, 1e-6) || !floatApprox(hex.BinCenters[1].Y, 0.25, 1e-6) {
		t.Fatalf("second center = %+v, want near {0.25 0.25}", hex.BinCenters[1])
	}
	if len(hex.EdgeColors) != len(hex.FaceColors) {
		t.Fatalf("hex edge colors len = %d, want face-colored edges for %d faces", len(hex.EdgeColors), len(hex.FaceColors))
	}
	if got, want := hex.EdgeWidth, pointsToPixels(fig.RC, 1); !floatApprox(got, want, 1e-12) {
		t.Fatalf("hex default linewidth = %v px, want Matplotlib patch.linewidth %v px", got, want)
	}
	mapping := hex.ScalarMap()
	if mapping.Colormap != "viridis" || mapping.VMin != 2 || mapping.VMax != 9 {
		t.Fatalf("unexpected scalar map %+v", mapping)
	}
}

func TestAxesHexbinExposesConfiguredNorm(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	hex := ax.Hexbin(
		[]float64{0.1, 0.2, 0.85},
		[]float64{0.1, 0.2, 0.8},
		HexbinOptions{
			GridSizeX: 2,
			GridSizeY: 2,
			C:         []float64{1, 10, 100},
			Reduce:    "mean",
			Norm:      LogNorm{VMin: 1, VMax: 100},
			Extent: &geom.Rect{
				Min: geom.Pt{X: 0, Y: 0},
				Max: geom.Pt{X: 1, Y: 1},
			},
		},
	)
	if hex == nil {
		t.Fatal("expected hexbin collection")
	}
	mapping := hex.ScalarMap()
	if mapping.Norm == nil || mapping.Norm.NormName() != "log" {
		t.Fatalf("hexbin norm = %#v, want log norm", mapping.Norm)
	}
}

func TestAxesHexbinLogBinsReducersAndMarginals(t *testing.T) {
	ax := NewFigure(640, 480).AddAxes(geom.Rect{})

	hex := ax.Hexbin(
		[]float64{0.1, 0.3, 0.8},
		[]float64{0.1, 0.4, 0.9},
		HexbinOptions{
			GridSizeX: 2,
			GridSizeY: 2,
			C:         []float64{2, 9, 5},
			Reduce:    "max",
			Bins:      "log",
			Marginals: true,
			Extent: &geom.Rect{
				Min: geom.Pt{X: 0, Y: 0},
				Max: geom.Pt{X: 2, Y: 2},
			},
		},
	)
	if hex == nil {
		t.Fatal("expected hexbin collection")
	}
	if hex.ScalarMap().Norm == nil || hex.ScalarMap().Norm.NormName() != "log" {
		t.Fatalf("hexbin bins=log norm = %#v, want log", hex.ScalarMap().Norm)
	}
	if hex.HBar == nil || hex.VBar == nil {
		t.Fatal("expected marginal bar collections")
	}
	if !(hex.HBar.Z() > hex.Z() && hex.VBar.Z() > hex.Z()) {
		t.Fatalf("marginal z orders h=%v v=%v hex=%v, want marginals above main hexbin like Matplotlib draw order", hex.HBar.Z(), hex.VBar.Z(), hex.Z())
	}
	wantLineWidth := pointsToPixels(ax.resolvedRC(), 1)
	if !floatApprox(hex.HBar.EdgeWidth, wantLineWidth, 1e-12) || !floatApprox(hex.VBar.EdgeWidth, wantLineWidth, 1e-12) {
		t.Fatalf("marginal edge widths h=%v v=%v, want Matplotlib PolyCollection default %v", hex.HBar.EdgeWidth, hex.VBar.EdgeWidth, wantLineWidth)
	}
	if len(hex.HBar.EdgeColors) != len(hex.HBar.FaceColors) || len(hex.VBar.EdgeColors) != len(hex.VBar.FaceColors) {
		t.Fatalf("marginal edge colors should mirror face colors like edgecolors='face': h=%d/%d v=%d/%d",
			len(hex.HBar.EdgeColors), len(hex.HBar.FaceColors), len(hex.VBar.EdgeColors), len(hex.VBar.FaceColors))
	}
}

func TestAxesHexbinLogScaleBuildsHexagonsInLogSpace(t *testing.T) {
	ax := NewFigure(640, 480).AddAxes(geom.Rect{})

	hex := ax.Hexbin(
		[]float64{1.2, 1.8, 2.6, 4.0, 6.5, 9.0, 14, 22, 35, 58, 92},
		[]float64{1.1, 2.2, 3.0, 5.5, 7.0, 12, 20, 28, 48, 80, 105},
		HexbinOptions{
			GridSizeX: 6,
			C:         []float64{1, 3, 2, 5, 7, 6, 11, 14, 18, 23, 30},
			Reduce:    "max",
			Bins:      "log",
			XScale:    "log",
			YScale:    "log",
		},
	)
	if hex == nil || len(hex.Polygons) == 0 {
		t.Fatal("expected log hexbin polygons")
	}

	maxLogWidth := 0.0
	for _, poly := range hex.Polygons {
		minX, maxX := math.Inf(1), math.Inf(-1)
		for _, pt := range poly {
			if pt.X <= 0 {
				t.Fatalf("log hexbin polygon contains non-positive x coordinate %+v", pt)
			}
			lx := math.Log10(pt.X)
			minX = math.Min(minX, lx)
			maxX = math.Max(maxX, lx)
		}
		maxLogWidth = math.Max(maxLogWidth, maxX-minX)
	}
	if maxLogWidth < 0.05 {
		t.Fatalf("log hexbin polygons are collapsed in log space; max width = %.4f", maxLogWidth)
	}
}

func TestAxesHexbinLogMarginalsIncludeMaxEndpointLikeMatplotlib(t *testing.T) {
	ax := NewFigure(640, 480).AddAxes(geom.Rect{})

	hex := ax.Hexbin(
		[]float64{1.2, 1.8, 2.6, 4.0, 6.5, 9.0, 14, 22, 35, 58, 92},
		[]float64{1.1, 2.2, 3.0, 5.5, 7.0, 12, 20, 28, 48, 80, 105},
		HexbinOptions{
			GridSizeX: 6,
			C:         []float64{1, 3, 2, 5, 7, 6, 11, 14, 18, 23, 30},
			Reduce:    "max",
			Bins:      "log",
			XScale:    "log",
			YScale:    "log",
			Marginals: true,
		},
	)
	if hex == nil || hex.HBar == nil || hex.VBar == nil {
		t.Fatal("expected log hexbin with marginal bars")
	}

	mapping := hex.ScalarMap()
	wantH := []float64{3, 5, 7, 11, 18, 30}
	wantV := []float64{3, 2, 7, 11, 18, 30}
	assertMarginalColors := func(name string, got []render.Color, wantValues []float64) {
		t.Helper()
		if len(got) != len(wantValues) {
			t.Fatalf("%s colors = %d, want %d", name, len(got), len(wantValues))
		}
		for i, value := range wantValues {
			want := mapping.Color(value, 1)
			if got[i] != want {
				t.Fatalf("%s color[%d] = %+v, want scalar-mapped value %.1f color %+v", name, i, got[i], value, want)
			}
		}
	}
	assertMarginalColors("hbar", hex.HBar.FaceColors, wantH)
	assertMarginalColors("vbar", hex.VBar.FaceColors, wantV)
}

func TestAxesPieCreatesWedgesAndLabels(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	pie := ax.Pie([]float64{2, 3}, PieOptions{
		Labels:  []string{"A", "B"},
		AutoPct: "%.0f%%",
	})
	if pie == nil {
		t.Fatal("expected pie container")
	}
	if len(pie.Wedges) != 2 || len(pie.Labels) != 2 || len(pie.AutoText) != 2 {
		t.Fatalf("unexpected pie counts wedges=%d labels=%d auto=%d", len(pie.Wedges), len(pie.Labels), len(pie.AutoText))
	}
	if pie.Labels[0].ClipOn || pie.AutoText[0].ClipOn {
		t.Fatal("expected pie label text to draw outside the axes clip")
	}
	if pie.AutoText[0].Color != (render.Color{}) {
		t.Fatalf("autopct color = %+v, want default text color", pie.AutoText[0].Color)
	}
	if pie.Wedges[0].Theta1 != 0 || pie.Wedges[0].Theta2 != 144 {
		t.Fatalf("unexpected first wedge angles %.1f -> %.1f", pie.Wedges[0].Theta1, pie.Wedges[0].Theta2)
	}
	if bounds := pie.Wedges[0].Bounds(nil); bounds == (geom.Rect{}) {
		t.Fatal("expected wedge bounds")
	}
}

func TestAxesPieAdvancedOptionsAndPieLabel(t *testing.T) {
	ax := NewFigure(640, 480).AddAxes(geom.Rect{})
	normalize := false
	pie := ax.Pie([]float64{0.25, 0.25}, PieOptions{
		Labels:       []string{"A", "B"},
		Normalize:    &normalize,
		RotateLabels: true,
		Hatches:      []string{"/", "x"},
		Shadow:       true,
	})
	if pie == nil {
		t.Fatal("expected partial pie")
	}
	if pie.Wedges[0].Theta2 != 90 {
		t.Fatalf("first unnormalized wedge ends at %v, want 90", pie.Wedges[0].Theta2)
	}
	if pie.Wedges[0].Hatch != "/" || pie.Wedges[1].Hatch != "x" {
		t.Fatalf("hatches = %q, %q", pie.Wedges[0].Hatch, pie.Wedges[1].Hatch)
	}
	if len(pie.Shadows) != 2 {
		t.Fatalf("shadows = %d, want 2", len(pie.Shadows))
	}
	if len(pie.LabelAngles) != 2 || pie.LabelAngles[0] == 0 {
		t.Fatalf("label rotations = %v, want populated", pie.LabelAngles)
	}
	if len(pie.Labels) != 2 || pie.Labels[0].Angle == 0 {
		t.Fatalf("pie label text angle = %v, want non-zero rotation", pie.Labels[0].Angle)
	}
	if pie.Labels[0].FontSize != ax.effectiveRC(nil).TickLabelSize("x") {
		t.Fatalf("pie label font size = %v, want xtick label size %v", pie.Labels[0].FontSize, ax.effectiveRC(nil).TickLabelSize("x"))
	}
	if pie.Labels[0].VAlign != TextVAlignBottom || pie.Labels[1].VAlign != TextVAlignBottom {
		t.Fatalf("upper rotated pie label vertical alignment = %v, %v; want bottom", pie.Labels[0].VAlign, pie.Labels[1].VAlign)
	}
	deepPie := ax.Pie([]float64{0.22, 0.18, 0.30}, PieOptions{
		Labels:       []string{"Alpha", "Beta", "Gamma"},
		Normalize:    &normalize,
		RotateLabels: true,
		StartAngle:   30,
	})
	if deepPie == nil || len(deepPie.Labels) != 3 {
		t.Fatal("expected partial rotated pie labels")
	}
	if deepPie.Labels[2].VAlign != TextVAlignTop {
		t.Fatalf("lower rotated pie label vertical alignment = %v, want top", deepPie.Labels[2].VAlign)
	}
	added := ax.PieLabel(pie, []string{"one", "two"}, PieLabelOptions{Distance: 0.8, Rotate: true})
	if len(added) != 2 {
		t.Fatalf("PieLabel added %d labels, want 2", len(added))
	}
	if added[0].Angle == 0 {
		t.Fatal("PieLabel Rotate should set text angle")
	}
}

func TestAxesViolinplotAddsCollections(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	violins := ax.Violinplot([][]float64{
		{1, 2, 2.5, 3, 4},
		{2, 2.1, 2.2, 3.4, 3.6},
	}, ViolinOptions{
		ShowMeans: specialtyBoolPtr(true),
		Alpha:     0.45,
		Label:     "spread",
	})
	if violins == nil {
		t.Fatal("expected violin container")
	}
	if violins.Bodies == nil || len(violins.Bodies.Polygons) != 2 {
		t.Fatal("expected two violin bodies")
	}
	if violins.Medians == nil || len(violins.Medians.Segments) != 2 {
		t.Fatal("expected median segments")
	}
	if violins.Means == nil || len(violins.Means.Segments) != 2 {
		t.Fatal("expected mean segments")
	}
	if violins.Extrema == nil || len(violins.Extrema.Segments) != 6 {
		t.Fatalf("expected extrema segments, got %d", len(violins.Extrema.Segments))
	}
	if got := violins.Extrema.LineCap; got != render.CapButt {
		t.Fatalf("two-sided violin extrema line cap = %v, want Matplotlib LineCollection butt cap", got)
	}
	if got := violins.Bodies.Alpha; got != 0.45 {
		t.Fatalf("violin body collection alpha = %v, want Matplotlib-style artist alpha", got)
	}
	if got := violins.Bodies.FaceColors[0].A; got != 1 {
		t.Fatalf("violin face color alpha = %v, want unmodified color alpha before collection alpha", got)
	}
	if got := violins.Bodies.EdgeWidth; got != 1 {
		t.Fatalf("violin body edge width = %v, want Matplotlib-rendered 1 px collection stroke", got)
	}
}

func TestAxesViolinUsesPrecomputedStats(t *testing.T) {
	ax := NewFigure(640, 360).AddAxes(geom.Rect{})
	defaultColor := ax.PeekColor()

	violins := ax.Violin(
		[]ViolinStat{
			{
				Coords:    []float64{1, 2, 3},
				Vals:      []float64{0.2, 1.0, 0.2},
				Mean:      2.1,
				Median:    2,
				Min:       1,
				Max:       3,
				Quantiles: []float64{1.5, 2.5},
			},
			{
				Coords: []float64{2, 3, 4},
				Vals:   []float64{0.3, 1.2, 0.3},
				Mean:   3,
				Median: 3,
				Min:    2,
				Max:    4,
			},
		},
		ViolinStatsOptions{
			Positions:   []float64{1.5, 2.5},
			Widths:      []float64{0.4, 0.6},
			ShowMeans:   boolPtr(true),
			ShowMedians: boolPtr(true),
			ShowExtrema: boolPtr(true),
			Side:        "high",
		},
	)

	if violins == nil {
		t.Fatal("expected precomputed violin container")
	}
	if violins.Bodies == nil || len(violins.Bodies.Polygons) != 2 {
		t.Fatalf("violin bodies = %#v, want two body polygons", violins.Bodies)
	}
	if len(violins.Bodies.FaceColors) != 2 || violins.Bodies.FaceColors[0] != defaultColor || violins.Bodies.FaceColors[1] != defaultColor {
		t.Fatalf("violin default face colors = %+v, want one Matplotlib cycle color %+v reused for all bodies", violins.Bodies.FaceColors, defaultColor)
	}
	if got := violins.Bodies.EdgeColor.A; got != 0 {
		t.Fatalf("violin default body edge alpha = %v, want no edge like Matplotlib violin bodies", got)
	}
	if violins.Means == nil || len(violins.Means.Segments) != 2 {
		t.Fatalf("mean segments = %#v, want 2", violins.Means)
	}
	if got := violins.Means.Color; got != defaultColor {
		t.Fatalf("violin mean line color = %+v, want default body color %+v", got, defaultColor)
	}
	if violins.Medians == nil || len(violins.Medians.Segments) != 2 {
		t.Fatalf("median segments = %#v, want 2", violins.Medians)
	}
	if violins.Quantiles == nil || len(violins.Quantiles.Segments) != 2 {
		t.Fatalf("quantile segments = %#v, want 2", violins.Quantiles)
	}
	if violins.Extrema == nil || len(violins.Extrema.Segments) != 6 {
		t.Fatalf("extrema segments = %#v, want 6", violins.Extrema)
	}
	if got := violins.Medians.LineCap; got != render.CapSquare {
		t.Fatalf("one-sided violin summary line cap = %v, want Matplotlib projecting cap", got)
	}
	if got := violins.Bodies.Polygons[0][0]; got != (geom.Pt{X: 1.5, Y: 1}) {
		t.Fatalf("first one-sided body point = %+v, want anchored at position", got)
	}
}

func TestAxesViolinValidatesPrecomputedStats(t *testing.T) {
	ax := NewFigure(640, 360).AddAxes(geom.Rect{})

	if got := ax.Violin([]ViolinStat{{Coords: []float64{1, 2}, Vals: []float64{1}}}); got != nil {
		t.Fatalf("Violin with mismatched coords/vals returned %#v, want nil", got)
	}
	if got := ax.Violin(
		[]ViolinStat{
			{Coords: []float64{1, 2}, Vals: []float64{1, 1}, Mean: 1.5, Median: 1.5, Min: 1, Max: 2},
			{Coords: []float64{1, 2}, Vals: []float64{1, 1}, Mean: 1.5, Median: 1.5, Min: 1, Max: 2},
		},
		ViolinStatsOptions{Positions: []float64{1}},
	); got != nil {
		t.Fatalf("Violin with mismatched positions returned %#v, want nil", got)
	}
}

func TestAxesViolinplotSideOrientationQuantilesAndBandwidthMethod(t *testing.T) {
	ax := NewFigure(640, 480).AddAxes(geom.Rect{})
	violins := ax.Violinplot([][]float64{
		{1, 2, 3, 4, 5},
	}, ViolinOptions{
		Orientation:     "horizontal",
		Side:            "high",
		Quantiles:       [][]float64{{0.25, 0.75}},
		BandwidthMethod: "scott",
	})
	if violins == nil || violins.Bodies == nil || len(violins.Bodies.Polygons) != 1 {
		t.Fatal("expected horizontal violin body")
	}
	for _, pt := range violins.Bodies.Polygons[0] {
		if pt.Y < 1 {
			t.Fatalf("side=high should not extend below position, got point %+v", pt)
		}
	}
	if violins.Quantiles == nil || len(violins.Quantiles.Segments) != 2 {
		t.Fatalf("quantile segments = %#v, want 2", violins.Quantiles)
	}
	if got := violins.Extrema.LineCap; got != render.CapSquare {
		t.Fatalf("one-sided violin extrema line cap = %v, want Matplotlib projecting cap", got)
	}
}

func TestAxesTableDrawsCellsAndText(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	table := ax.Table(TableOptions{
		CellText:  [][]string{{"1.0", "2.0"}, {"3.0", "4.0"}},
		RowLabels: []string{"R1", "R2"},
		ColLabels: []string{"C1", "C2"},
		BBox:      geom.Rect{Min: geom.Pt{X: 0.15, Y: 0.15}, Max: geom.Pt{X: 0.85, Y: 0.55}},
	})
	if table == nil {
		t.Fatal("expected table artist")
	}
	if len(table.Cells) != 3 || len(table.Cells[0]) != 3 {
		t.Fatalf("unexpected table grid %dx%d", len(table.Cells), len(table.Cells[0]))
	}
	if table.Cells[0][0].Rect != (geom.Rect{}) {
		t.Fatalf("top-left row-label/header intersection rect = %+v, want empty", table.Cells[0][0].Rect)
	}
	if got, want := table.Cells[1][0].Rect.Max.X, table.BBox.Min.X; !floatApprox(got, want, 1e-12) {
		t.Fatalf("row label right edge = %v, want bbox left %v", got, want)
	}
	if table.HeaderTextColor != (render.Color{A: 1}) || table.EdgeColor != (render.Color{A: 1}) {
		t.Fatalf("table defaults headerText=%+v edge=%+v, want opaque black", table.HeaderTextColor, table.EdgeColor)
	}

	var renderer specialtyRecordingRenderer
	DrawFigure(fig, &renderer)
	if renderer.pathCount < 8 {
		t.Fatalf("expected at least 8 cell paths, got %d", renderer.pathCount)
	}
	if len(renderer.texts) < 8 {
		t.Fatalf("expected cell/header text draws, got %v", renderer.texts)
	}
}

func TestAxesTableDrawsAsUnclippedOverlayByDefault(t *testing.T) {
	table := (&Axes{}).Table(TableOptions{
		CellText: [][]string{{"1"}},
		BBox:     geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}},
	})
	if table == nil {
		t.Fatal("expected table artist")
	}

	ctx := &DrawContext{
		Clip: geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 100}},
	}
	var clipped specialtyRecordingRenderer
	table.Draw(&clipped, ctx)
	if clipped.pathCount != 0 {
		t.Fatalf("Table.Draw path count = %d, want default table to defer to unclipped overlay draw", clipped.pathCount)
	}

	var overlay specialtyRecordingRenderer
	table.DrawOverlay(&overlay, ctx)
	if overlay.pathCount == 0 {
		t.Fatal("Table.DrawOverlay drew no cell paths")
	}
}

func TestAxesTableHonorsAlignmentPadding(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	table := ax.Table(TableOptions{
		CellText:  [][]string{{"L", "R"}},
		RowLabels: []string{"row"},
		ColLabels: []string{"C1", "C2"},
		BBox:      geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}},
		FontSize:  10,
		CellLoc:   "left",
		RowLoc:    "right",
		ColLoc:    "center",
	})
	if table == nil {
		t.Fatal("expected table artist")
	}
	if got := table.Cells[1][1].HAlign; got != TextAlignLeft {
		t.Fatalf("data cell align = %v, want left", got)
	}
	if got := table.Cells[1][0].HAlign; got != TextAlignRight {
		t.Fatalf("row label align = %v, want right", got)
	}
	if got := table.Cells[0][1].HAlign; got != TextAlignCenter {
		t.Fatalf("column label align = %v, want center", got)
	}

	var renderer specialtyRecordingRenderer
	table.DrawOverlay(&renderer, &DrawContext{
		Clip: geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 100}},
	})
	origin, ok := renderer.textOrigins["L"]
	if !ok {
		t.Fatalf("expected data text draw, got %v", renderer.texts)
	}
	if !floatApprox(origin.X, 5, 1e-12) {
		t.Fatalf("left-aligned data text origin x = %v, want 10%% cell padding at 5px", origin.X)
	}
}

func TestAxesTableCentersTextUsingTextAdvance(t *testing.T) {
	table := (&Axes{}).Table(TableOptions{
		CellText: [][]string{{"ink"}},
		BBox:     geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}},
		FontSize: 10,
		CellLoc:  "center",
	})
	if table == nil {
		t.Fatal("expected table artist")
	}

	renderer := specialtyRecordingRenderer{
		textBounds: map[string]render.TextBounds{
			"ink": {X: 1, Y: -8, W: 40, H: 10},
		},
		textMetrics: map[string]render.TextMetrics{
			"ink": {W: 38, H: 10, Ascent: 8, Descent: 2},
		},
	}
	table.DrawOverlay(&renderer, &DrawContext{
		Clip: geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 100}},
	})

	origin, ok := renderer.textOrigins["ink"]
	if !ok {
		t.Fatalf("expected table text draw, got %v", renderer.texts)
	}
	if got, want := origin.X, 31.0; !floatApprox(got, want, 1e-12) {
		t.Fatalf("centered table text origin x = %v, want text advance centered at cell anchor: %v", got, want)
	}
}

func TestAxesTableAutoSizesRowLabelsFromRendererTextBounds(t *testing.T) {
	table := (&Axes{}).Table(TableOptions{
		CellText:  [][]string{{"1"}},
		RowLabels: []string{"R"},
		BBox:      geom.Rect{Min: geom.Pt{X: 0.4, Y: 0}, Max: geom.Pt{X: 0.9, Y: 1}},
		FontSize:  10,
	})
	if table == nil {
		t.Fatal("expected table artist")
	}

	renderer := specialtyRecordingRenderer{
		textBounds: map[string]render.TextBounds{
			"R": {X: 1, Y: -8, W: 6, H: 10},
		},
		textMetrics: map[string]render.TextMetrics{
			"R": {W: 10, H: 10, Ascent: 8, Descent: 2},
		},
	}
	table.DrawOverlay(&renderer, &DrawContext{
		Clip: geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 100}},
	})
	if len(renderer.paths) == 0 {
		t.Fatal("expected row-label cell path")
	}
	bounds, ok := pathBounds(renderer.paths[0])
	if !ok {
		t.Fatal("expected row-label path bounds")
	}
	if got, want := bounds.Min.X, 28.0; !floatApprox(got, want, 1e-12) {
		t.Fatalf("row-label left edge = %v, want bbox-scaled measured text width left of shifted grid", got)
	}
	if got, want := bounds.Max.X, 34.0; !floatApprox(got, want, 1e-12) {
		t.Fatalf("row-label right edge = %v, want Matplotlib shifted data-grid left edge", got)
	}
	if len(renderer.paths) < 2 {
		t.Fatal("expected data cell path")
	}
	dataBounds, ok := pathBounds(renderer.paths[1])
	if !ok {
		t.Fatal("expected data cell path bounds")
	}
	if got, want := dataBounds.Min.X, 34.0; !floatApprox(got, want, 1e-12) {
		t.Fatalf("data-cell left edge = %v, want Matplotlib shifted data-grid left edge", got)
	}
}

func TestAxesTableUsesMatplotlibPatchLineWidthDefault(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{})

	table := ax.Table(TableOptions{
		CellText: [][]string{{"value"}},
	})
	if table == nil {
		t.Fatal("expected table artist")
	}

	if got, want := table.LineWidth, fig.RC.DPI/72.0; !floatApprox(got, want, 1e-12) {
		t.Fatalf("default table line width = %v, want matplotlib patch.linewidth 1pt at %v DPI = %v px", got, fig.RC.DPI, want)
	}
}

func TestAxesTableCellsUseMatplotlibPatchSnapAuto(t *testing.T) {
	table := (&Axes{}).Table(TableOptions{
		CellText: [][]string{{"value"}},
		BBox:     geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.2}, Max: geom.Pt{X: 0.9, Y: 0.8}},
	})
	if table == nil {
		t.Fatal("expected table artist")
	}

	var renderer specialtyRecordingRenderer
	table.DrawOverlay(&renderer, &DrawContext{
		Clip: geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 100}},
	})
	if len(renderer.paints) == 0 {
		t.Fatal("expected table cell paint")
	}
	if got := renderer.paints[0].Snap; got != render.SnapAuto {
		t.Fatalf("table cell snap = %v, want Matplotlib patch snap auto", got)
	}
}

func TestSankeyBuilderCreatesDiagram(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	builder := NewSankey(ax, SankeyOptions{
		Center: geom.Pt{X: 0.2, Y: 0.5},
		Scale:  0.1,
	})
	if builder == nil {
		t.Fatal("expected sankey builder")
	}
	diagram := builder.Add([]float64{-2, 3}, SankeyAddOptions{
		Labels:       []string{"Loss", "Gain"},
		Orientations: []int{-1, 1},
	})
	if diagram == nil {
		t.Fatal("expected sankey diagram")
	}
	if diagram.Trunk == nil || len(diagram.Ribbons) != 1 || len(diagram.Labels) != 2 || len(diagram.Values) != 2 {
		t.Fatalf("unexpected sankey content %+v", diagram)
	}
	if got, want := diagram.Trunk.Height, 0.3; !floatApprox(got, want, 1e-12) {
		t.Fatalf("trunk height = %v, want max flow sum scaled to %v", got, want)
	}
	if diagram.Values[0].Content != "2" || diagram.Values[1].Content != "3" {
		t.Fatalf("unexpected flow value labels %q %q", diagram.Values[0].Content, diagram.Values[1].Content)
	}
	if finished := builder.Finish(); len(finished) != 1 {
		t.Fatalf("Finish len = %d, want 1", len(finished))
	}
	if len(ax.Artists) < 5 {
		t.Fatalf("expected artists to be added to axes, got %d", len(ax.Artists))
	}
}

func TestSankeyMatchesMatplotlibSingleDiagramGeometry(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	builder := NewSankey(ax, SankeyOptions{Scale: 0.16, Offset: 0.2})
	diagram := builder.Add([]float64{-2, 3, 1.5}, SankeyAddOptions{
		Labels:       []string{"Waste", "CPU", "Cache"},
		Orientations: []int{-1, 1, 1},
	})
	if diagram == nil || diagram.Patch == nil {
		t.Fatal("expected sankey diagram patch")
	}
	builder.Finish()

	wantTips := []geom.Pt{
		{X: 0.66, Y: -0.5694289299236832},
		{X: -0.74, Y: 0.4086160885174527},
		{X: -1.35, Y: 0.5093080442587265},
	}
	for i, want := range wantTips {
		if !floatApprox(diagram.Tips[i].X, want.X, 1e-12) || !floatApprox(diagram.Tips[i].Y, want.Y, 1e-12) {
			t.Fatalf("tip %d = %+v, want %+v", i, diagram.Tips[i], want)
		}
		if diagram.Angles[i] != sankeyDown {
			t.Fatalf("angle %d = %d, want DOWN", i, diagram.Angles[i])
		}
	}

	bounds, ok := pathBounds(diagram.Patch.Path)
	if !ok {
		t.Fatal("expected path bounds")
	}
	if !floatApprox(bounds.Min.X, -1.47, 1e-12) ||
		!floatApprox(bounds.Max.X, 0.85, 1e-12) ||
		!floatApprox(bounds.Min.Y, -0.5694289299236832, 1e-12) ||
		!floatApprox(bounds.Max.Y, 0.61, 1e-12) {
		t.Fatalf("path bounds = %+v", bounds)
	}

	xMin, xMax := ax.XScale.Domain()
	yMin, yMax := ax.YScale.Domain()
	if !floatApprox(xMin, -1.87, 1e-12) || !floatApprox(xMax, 1.25, 1e-12) ||
		!floatApprox(yMin, -1.1694289299236833, 1e-12) || !floatApprox(yMax, 1.1093080442587264, 1e-12) {
		t.Fatalf("finished limits = x(%v, %v) y(%v, %v)", xMin, xMax, yMin, yMax)
	}
}

type specialtyRecordingRenderer struct {
	render.NullRenderer
	pathCount   int
	paths       []geom.Path
	paints      []render.Paint
	texts       []string
	textOrigins map[string]geom.Pt
	textBounds  map[string]render.TextBounds
	textMetrics map[string]render.TextMetrics
}

func (r *specialtyRecordingRenderer) Path(path geom.Path, paint *render.Paint) {
	r.pathCount++
	r.paths = append(r.paths, path)
	if paint != nil {
		r.paints = append(r.paints, *paint)
	}
}

func (r *specialtyRecordingRenderer) DrawText(text string, pt geom.Pt, _ float64, _ render.Color) {
	r.texts = append(r.texts, text)
	if r.textOrigins == nil {
		r.textOrigins = map[string]geom.Pt{}
	}
	r.textOrigins[text] = pt
}

func (r *specialtyRecordingRenderer) MeasureText(text string, size float64, _ string) render.TextMetrics {
	if r.textMetrics != nil {
		if metrics, ok := r.textMetrics[text]; ok {
			return metrics
		}
	}
	return render.TextMetrics{
		W:       float64(len(text)) * size * 0.55,
		H:       size,
		Ascent:  size * 0.8,
		Descent: size * 0.2,
	}
}

func (r *specialtyRecordingRenderer) MeasureTextBounds(text string, _ float64, _ string) (render.TextBounds, bool) {
	if r.textBounds == nil {
		return render.TextBounds{}, false
	}
	bounds, ok := r.textBounds[text]
	return bounds, ok
}

func specialtyBoolPtr(v bool) *bool {
	return &v
}
