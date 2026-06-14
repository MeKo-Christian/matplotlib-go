package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestHist2D_Draw_Basic(t *testing.T) {
	hist := &Hist2D{
		Data:  []float64{1, 2, 2, 3, 3, 3, 4, 4, 5},
		Bins:  5,
		Color: render.Color{R: 0.3, G: 0.6, B: 0.9, A: 1},
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	if err := renderer.Begin(geom.Rect{}); err != nil {
		t.Fatal(err)
	}
	hist.Draw(renderer, ctx)
	if err := renderer.End(); err != nil {
		t.Fatal(err)
	}
}

func TestHist2D_Draw_Empty(t *testing.T) {
	hist := &Hist2D{
		Data:  []float64{},
		Color: render.Color{R: 1, G: 0, B: 0, A: 1},
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	if err := renderer.Begin(geom.Rect{}); err != nil {
		t.Fatal(err)
	}
	// Should not panic
	hist.Draw(renderer, ctx)
	if err := renderer.End(); err != nil {
		t.Fatal(err)
	}
}

func TestHist2D_Draw_SingleValue(t *testing.T) {
	hist := &Hist2D{
		Data:  []float64{5, 5, 5},
		Color: render.Color{R: 0.5, G: 0.5, B: 0.5, A: 1},
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	if err := renderer.Begin(geom.Rect{}); err != nil {
		t.Fatal(err)
	}
	hist.Draw(renderer, ctx)
	if err := renderer.End(); err != nil {
		t.Fatal(err)
	}

	edges, counts := hist.BinCounts()
	if len(edges) < 2 {
		t.Errorf("expected at least 2 edges, got %d", len(edges))
	}
	if len(counts) == 0 {
		t.Errorf("expected non-empty counts for non-empty data")
	}
}

func TestHist2D_Draw_WithEdgeColors(t *testing.T) {
	hist := &Hist2D{
		Data:      []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		Bins:      4,
		Color:     render.Color{R: 0.8, G: 0.2, B: 0.2, A: 1},
		EdgeColor: render.Color{R: 0, G: 0, B: 0, A: 1},
		EdgeWidth: 1.5,
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	if err := renderer.Begin(geom.Rect{}); err != nil {
		t.Fatal(err)
	}
	hist.Draw(renderer, ctx)
	if err := renderer.End(); err != nil {
		t.Fatal(err)
	}
}

func TestHist2D_Draw_Alpha(t *testing.T) {
	hist := &Hist2D{
		Data:  []float64{1, 2, 2, 3, 3, 3, 4},
		Alpha: 0.5,
		Color: render.Color{R: 0.5, G: 0.5, B: 1, A: 1},
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	if err := renderer.Begin(geom.Rect{}); err != nil {
		t.Fatal(err)
	}
	hist.Draw(renderer, ctx)
	if err := renderer.End(); err != nil {
		t.Fatal(err)
	}
}

func TestHist2D_Draw_AlphaOverridesColorAlpha(t *testing.T) {
	hist := &Hist2D{
		Data:      []float64{1, 2, 2, 3},
		BinEdges:  []float64{1, 2, 3, 4},
		Alpha:     0.6,
		Color:     render.Color{R: 0.5, G: 0.5, B: 1, A: 0.25},
		EdgeColor: render.Color{R: 0.1, G: 0.2, B: 0.3, A: 0.4},
		EdgeWidth: 1,
	}

	r := &recordingRenderer{}
	hist.Draw(r, createTestDrawContext())
	if len(r.pathCalls) < 2 {
		t.Fatalf("path calls = %d, want fill and stroke calls", len(r.pathCalls))
	}
	if got := r.pathCalls[0].paint.Fill.A; got != hist.Alpha {
		t.Fatalf("drawn fill alpha = %v, want explicit alpha override %v", got, hist.Alpha)
	}
	if got := r.pathCalls[1].paint.Stroke.A; got != hist.Alpha {
		t.Fatalf("drawn stroke alpha = %v, want explicit alpha override %v", got, hist.Alpha)
	}
}

func TestHist2D_BinCounts_ExplicitEdges(t *testing.T) {
	hist := &Hist2D{
		Data:     []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		BinEdges: []float64{0, 5, 10},
	}

	edges, counts := hist.BinCounts()

	if len(edges) != 3 {
		t.Fatalf("expected 3 edges, got %d", len(edges))
	}
	if len(counts) != 2 {
		t.Fatalf("expected 2 bins, got %d", len(counts))
	}
	// values 1-4 in [0,5), value 5 in [0,5) too (<=5 but <10 means in bin 0 only if v<5)
	// Actually: bin 0 = [0,5) → 1,2,3,4 = 4 values
	// bin 1 = [5,10] → 5,6,7,8,9,10 = 6 values
	if counts[0] != 4 {
		t.Errorf("expected counts[0]=4, got %v", counts[0])
	}
	if counts[1] != 6 {
		t.Errorf("expected counts[1]=6, got %v", counts[1])
	}
}

func TestHist2D_Normalization_Probability(t *testing.T) {
	hist := &Hist2D{
		Data: []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		Bins: 2,
		Norm: HistNormProbability,
	}

	_, counts := hist.BinCounts()

	total := 0.0
	for _, c := range counts {
		total += c
	}
	if math.Abs(total-1.0) > 1e-9 {
		t.Errorf("probability normalization should sum to 1, got %v", total)
	}
}

func TestHist2D_Normalization_Density(t *testing.T) {
	hist := &Hist2D{
		Data: []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		Bins: 2,
		Norm: HistNormDensity,
	}

	edges, counts := hist.BinCounts()

	// area = sum(count * width) should be 1
	area := 0.0
	for i, c := range counts {
		width := edges[i+1] - edges[i]
		area += c * width
	}
	if math.Abs(area-1.0) > 1e-9 {
		t.Errorf("density normalization area should be 1, got %v", area)
	}
}

func TestHist2D_WeightsAndDensityUseInRangeTotal(t *testing.T) {
	hist := &Hist2D{
		Data:     []float64{-1, 0.2, 0.4, 1.2, 2.2, 9},
		Weights:  []float64{100, 2, 3, 5, 10, 100},
		BinEdges: []float64{0, 1, 2, 3},
		Norm:     HistNormDensity,
	}

	edges, counts := hist.BinCounts()
	assertFloatSlices(t, "edges", edges, []float64{0, 1, 2, 3})
	assertFloatSlices(t, "weighted density", counts, []float64{0.25, 0.25, 0.5})
}

func TestHist2D_RangeSetsBinsAndExcludesOutliers(t *testing.T) {
	hist := &Hist2D{
		Data:  []float64{-10, 0.1, 0.9, 1.1, 1.9, 10},
		Bins:  2,
		Range: &HistRange{Min: 0, Max: 2},
	}

	edges, counts := hist.BinCounts()
	assertFloatSlices(t, "edges", edges, []float64{0, 1, 2})
	assertFloatSlices(t, "counts", counts, []float64{2, 2})
}

func TestHist2D_ReverseCumulativeDensity(t *testing.T) {
	hist := &Hist2D{
		Data:              []float64{0.2, 0.4, 1.2, 2.2},
		BinEdges:          []float64{0, 1, 2, 3},
		Norm:              HistNormDensity,
		Cumulative:        true,
		ReverseCumulative: true,
	}

	_, counts := hist.BinCounts()
	assertFloatSlices(t, "reverse cumulative density", counts, []float64{1, 0.5, 0.25})
}

func TestHist2D_Bounds_Empty(t *testing.T) {
	hist := &Hist2D{}
	bounds := hist.Bounds(nil)
	if bounds != (geom.Rect{}) {
		t.Errorf("expected zero bounds for empty histogram, got %v", bounds)
	}
}

func TestHist2D_Bounds_WithData(t *testing.T) {
	hist := &Hist2D{
		Data: []float64{1, 2, 3, 4, 5},
		Bins: 5,
	}

	bounds := hist.Bounds(nil)
	if bounds.Min.Y != 0 {
		t.Errorf("expected min Y=0 for histogram, got %v", bounds.Min.Y)
	}
	if bounds.Max.Y <= 0 {
		t.Errorf("expected positive max Y, got %v", bounds.Max.Y)
	}
	if bounds.Min.X >= bounds.Max.X {
		t.Errorf("expected min X < max X, got [%v, %v]", bounds.Min.X, bounds.Max.X)
	}
}

func TestHist2D_ZOrder(t *testing.T) {
	hist := &Hist2D{z: 1.5}
	if got := hist.Z(); got != 1.5 {
		t.Errorf("expected Z()=1.5, got %v", got)
	}
}

func TestBinStrategySturges(t *testing.T) {
	tests := []struct {
		n    int
		want int
	}{
		{1, 1},
		{2, 2},
		{10, 5},
		{100, 8},
		{1000, 11},
	}
	for _, tt := range tests {
		got := sturgesBins(tt.n)
		if got != tt.want {
			t.Errorf("sturgesBins(%d) = %d, want %d", tt.n, got, tt.want)
		}
	}
}

func TestFindBin(t *testing.T) {
	edges := []float64{0, 2, 4, 6, 8, 10}

	tests := []struct {
		v    float64
		want int
	}{
		{0.0, 0},
		{1.0, 0},
		{2.0, 1},
		{5.0, 2},
		{10.0, 4}, // right edge of last bin
		{-1.0, -1},
		{11.0, -1},
	}
	for _, tt := range tests {
		got := findBin(tt.v, edges)
		if got != tt.want {
			t.Errorf("findBin(%v, edges) = %d, want %d", tt.v, got, tt.want)
		}
	}
}

func TestAxes_Hist(t *testing.T) {
	fig := NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	data := []float64{1, 2, 2, 3, 3, 3, 4, 5}
	hist := ax.Hist(data)
	if hist == nil {
		t.Fatal("Hist should return non-nil for non-empty data")
	}

	// Test with empty data
	got := ax.Hist([]float64{})
	if got != nil {
		t.Errorf("Hist with empty data should return nil")
	}
}

func TestAxes_Hist_Options(t *testing.T) {
	fig := NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	col := render.Color{R: 1, G: 0, B: 0, A: 1}
	edge := render.Color{R: 0, G: 0, B: 0, A: 1}
	ew := 1.5
	alpha := 0.7
	bins := 10

	hist := ax.Hist([]float64{1, 2, 3, 4, 5}, HistOptions{
		Bins:      bins,
		Norm:      HistNormProbability,
		Color:     &col,
		EdgeColor: &edge,
		EdgeWidth: &ew,
		Alpha:     &alpha,
		Label:     "test",
	})

	if hist == nil {
		t.Fatal("expected non-nil histogram")
	}
	if hist.Label != "test" {
		t.Errorf("expected label 'test', got %q", hist.Label)
	}
}

func TestAxesHistRejectsMismatchedWeights(t *testing.T) {
	fig := NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	got := ax.Hist([]float64{1, 2, 3}, HistOptions{
		Weights: []float64{1, 2},
	})
	if got != nil {
		t.Fatal("Hist with mismatched weights should return nil")
	}
}

func TestAxesHistPreservesColorAlphaWhenAlphaOmitted(t *testing.T) {
	fig := NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	col := render.Color{R: 0.2, G: 0.4, B: 0.8, A: 0.35}
	hist := ax.Hist([]float64{1, 2, 2, 3}, HistOptions{Color: &col})
	if hist == nil {
		t.Fatal("expected non-nil histogram")
	}
	if hist.Alpha != 0 {
		t.Fatalf("omitted alpha stored Alpha = %v, want 0 sentinel", hist.Alpha)
	}

	r := &recordingRenderer{}
	hist.Draw(r, createTestDrawContext())
	if len(r.pathCalls) == 0 {
		t.Fatal("expected histogram path calls")
	}
	if got := r.pathCalls[0].paint.Fill.A; got != col.A {
		t.Fatalf("drawn fill alpha = %v, want color alpha %v", got, col.A)
	}
}

func TestHist2D_AllBinStrategies(t *testing.T) {
	data := make([]float64, 200)
	for i := range data {
		data[i] = float64(i % 50)
	}

	strategies := []BinStrategy{
		BinStrategyAuto,
		BinStrategySturges,
		BinStrategyScott,
		BinStrategyFD,
		BinStrategySqrt,
	}

	for _, start := range strategies {
		hist := &Hist2D{
			Data:     data,
			BinStrat: start,
			Color:    render.Color{R: 0.5, G: 0.5, B: 0.5, A: 1},
		}
		_, counts := hist.BinCounts()
		if len(counts) == 0 {
			t.Errorf("strategy %d produced 0 bins", start)
		}
	}
}
