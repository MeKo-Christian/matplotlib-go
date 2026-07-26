package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func pathPixelExtents(p geom.Path) (dx, dy float64) {
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)
	for _, v := range p.V {
		minX = math.Min(minX, v.X)
		maxX = math.Max(maxX, v.X)
		minY = math.Min(minY, v.Y)
		maxY = math.Max(maxY, v.Y)
	}
	return maxX - minX, maxY - minY
}

func TestBoxPlotComputesMean(t *testing.T) {
	box := &BoxPlot2D{Data: []float64{1, 2, 3, 4, 10}}
	box.ensureComputed()
	if got, want := box.stats.mean, 4.0; got != want {
		t.Fatalf("mean = %v, want %v", got, want)
	}
}

func TestBoxPlotScalarWhisControlsOutliers(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5, 6, 7, 8, 20}

	tight := 1.5
	box := &BoxPlot2D{Data: data, Whis: &tight}
	box.ensureComputed()
	if len(box.stats.outliers) != 1 || box.stats.outliers[0] != 20 {
		t.Fatalf("whis=1.5 outliers = %v, want [20]", box.stats.outliers)
	}
	if box.stats.upperWhisker != 8 {
		t.Fatalf("whis=1.5 upper whisker = %v, want 8", box.stats.upperWhisker)
	}

	wide := 4.0
	box2 := &BoxPlot2D{Data: data, Whis: &wide}
	box2.ensureComputed()
	if len(box2.stats.outliers) != 0 {
		t.Fatalf("whis=4 outliers = %v, want none", box2.stats.outliers)
	}
	if box2.stats.upperWhisker != 20 {
		t.Fatalf("whis=4 upper whisker = %v, want 20", box2.stats.upperWhisker)
	}
}

func TestBoxPlotPercentileWhiskerClampsToQuartile(t *testing.T) {
	// whis=(40, 60) lands the fences inside the box, so Matplotlib clamps the
	// whiskers to Q1/Q3 rather than pulling them in past the quartiles.
	whis := [2]float64{40, 60}
	box := &BoxPlot2D{Data: []float64{1, 2, 3, 4, 5, 6, 7, 8, 9}, WhiskerPercentiles: &whis}
	box.ensureComputed()
	if box.stats.lowerWhisker != box.stats.q1 {
		t.Fatalf("lower whisker = %v, want clamped to q1 %v", box.stats.lowerWhisker, box.stats.q1)
	}
	if box.stats.upperWhisker != box.stats.q3 {
		t.Fatalf("upper whisker = %v, want clamped to q3 %v", box.stats.upperWhisker, box.stats.q3)
	}
}

func TestBootstrapMedianCIIsDeterministicAndDistinct(t *testing.T) {
	sorted := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	lo1, hi1 := bootstrapMedianCI(sorted, 2000)
	lo2, hi2 := bootstrapMedianCI(sorted, 2000)
	if lo1 != lo2 || hi1 != hi2 {
		t.Fatalf("bootstrap CI not deterministic: (%v,%v) vs (%v,%v)", lo1, hi1, lo2, hi2)
	}
	if lo1 >= hi1 {
		t.Fatalf("bootstrap CI not ordered: (%v,%v)", lo1, hi1)
	}
	// The bootstrap interval should differ from the analytic 1.57*IQR/sqrt(N) one.
	stats := computeBoxPlotStats(sorted, 1.5)
	aLo, aHi := boxPlotMedianCI(sorted, stats)
	if lo1 == aLo && hi1 == aHi {
		t.Fatalf("bootstrap CI matches analytic fallback exactly: (%v,%v)", lo1, hi1)
	}
}

func TestBoxPlotHorizontalOrientationSwapsAxes(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5}
	ctx := createTestDrawContext()

	vert := &BoxPlot2D{Data: data, Position: 1, Width: 0.6, EdgeColor: render.Color{A: 1}, EdgeWidth: 1, ShowBox: true}
	vr := &recordingRenderer{}
	vert.Draw(vr, ctx)
	vdx, vdy := pathPixelExtents(vr.pathCalls[0].path)
	if vdx >= vdy {
		t.Fatalf("vertical box extents dx=%v dy=%v, want narrow X, tall Y", vdx, vdy)
	}

	horiz := &BoxPlot2D{Data: data, Position: 1, Width: 0.6, Orientation: "horizontal", EdgeColor: render.Color{A: 1}, EdgeWidth: 1, ShowBox: true}
	hr := &recordingRenderer{}
	horiz.Draw(hr, ctx)
	hdx, hdy := pathPixelExtents(hr.pathCalls[0].path)
	if hdx <= hdy {
		t.Fatalf("horizontal box extents dx=%v dy=%v, want wide X, short Y", hdx, hdy)
	}
}

func TestBoxPlotShowMeansDrawsMarkerOrLine(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5}
	black := render.Color{A: 1}
	ctx := createTestDrawContext()

	marker := &BoxPlot2D{
		Data: data, Position: 1, Width: 0.6,
		EdgeColor: black, MedianColor: black, MeanColor: black, WhiskerColor: black, CapColor: black,
		EdgeWidth: 1, WhiskerWidth: 1, MedianWidth: 1, FlierSize: 6,
		ShowBox: true, ShowCaps: true, ShowMeans: true,
	}
	mr := &recordingRenderer{}
	marker.Draw(mr, ctx)
	// box + 2 whiskers + 2 caps + median + mean marker = 7 path calls.
	if len(mr.pathCalls) != 7 {
		t.Fatalf("showmeans marker: %d path calls, want 7", len(mr.pathCalls))
	}

	line := &BoxPlot2D{
		Data: data, Position: 1, Width: 0.6,
		EdgeColor: black, MedianColor: black, MeanColor: black, WhiskerColor: black, CapColor: black,
		EdgeWidth: 1, WhiskerWidth: 1, MedianWidth: 1,
		ShowBox: true, ShowCaps: true, ShowMeans: true, MeanLine: true,
	}
	lr := &recordingRenderer{}
	line.Draw(lr, ctx)
	if len(lr.pathCalls) != 7 {
		t.Fatalf("meanline: %d path calls, want 7", len(lr.pathCalls))
	}
	// The mean line spans the box width; assert it is horizontal (constant Y).
	meanLine := lr.pathCalls[6].path
	if len(meanLine.V) != 2 || meanLine.V[0].Y != meanLine.V[1].Y {
		t.Fatalf("mean line is not a horizontal segment: %+v", meanLine.V)
	}
}

func TestBoxPlotShowBoxAndShowCapsGateDrawing(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5}
	black := render.Color{A: 1}
	ctx := createTestDrawContext()

	box := &BoxPlot2D{
		Data: data, Position: 1, Width: 0.6,
		EdgeColor: black, MedianColor: black, WhiskerColor: black, CapColor: black,
		EdgeWidth: 1, WhiskerWidth: 1, MedianWidth: 1,
		ShowBox: false, ShowCaps: false,
	}
	r := &recordingRenderer{}
	box.Draw(r, ctx)
	// No box, no caps: only 2 whiskers + median = 3 path calls.
	if len(r.pathCalls) != 3 {
		t.Fatalf("showbox=false showcaps=false: %d path calls, want 3 (whiskers + median)", len(r.pathCalls))
	}
}

func TestBoxPlotDefaultIsUnfilledAndDoesNotCycleColor(t *testing.T) {
	ax := NewFigure(640, 360).AddAxes(geom.Rect{})
	before := ax.NextColor()
	ax2 := NewFigure(640, 360).AddAxes(geom.Rect{})

	box := ax2.BoxPlot([]float64{1, 2, 3, 4, 5}, BoxPlotOptions{})
	if box == nil {
		t.Fatal("expected box plot")
	}
	if box.PatchArtist {
		t.Fatal("default boxplot should be unfilled (patch_artist=False)")
	}
	// The default box must not consume the color cycle: a fresh axes' first
	// cycle color is unchanged after creating a default boxplot.
	if got := ax2.NextColor(); got != before {
		t.Fatalf("default boxplot consumed the color cycle: first cycle color = %+v, want %+v", got, before)
	}

	// Drawing the default box must not emit any fill.
	r := &recordingRenderer{}
	box.Draw(r, createTestDrawContext())
	for i, c := range r.pathCalls {
		if c.paint.Fill.A != 0 {
			t.Fatalf("default boxplot path %d has fill alpha %v, want unfilled", i, c.paint.Fill.A)
		}
	}
}
