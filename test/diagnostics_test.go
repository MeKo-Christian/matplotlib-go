package test

// Diagnostic / dev-mode parity probes. These tests don't fit the golden
// or matplotlib-ref subtest loops; they each render a specific case and
// inspect bespoke metrics (alpha residuals, histogram bar heights, RNG
// stream ordering, text-placement deltas vs Matplotlib, etc.). Most are
// gated by an environment variable so they only run when explicitly
// requested.

import (
	"encoding/json"
	"image"
	"image/color"
	"math"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/test/imagecmp"
	"github.com/cwbudde/matplotlib-go/test/parity"
	"github.com/cwbudde/matplotlib-go/transform"
)

// === from alpha_residual_diagnostics_test.go ==========================

type alphaResidualCase struct {
	name             string
	axes             image.Rectangle
	threshold        uint8
	maxHighDiffRatio float64
}

func TestAlphaResidualDiagnostics(t *testing.T) {
	cases := []alphaResidualCase{
		{
			name:             "fill_stacked",
			axes:             fixtureAxesRect(640, 360, 0.1, 0.1, 0.9, 0.9),
			threshold:        32,
			maxHighDiffRatio: 0.030,
		},
		{
			name:             "hist_strategies",
			axes:             fixtureAxesRect(640, 360, 0.12, 0.12, 0.95, 0.90),
			threshold:        32,
			maxHighDiffRatio: 0.015,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, _, err := parity.Render(tc.name)
			if err != nil {
				t.Fatalf("render parity example %s: %v", tc.name, err)
			}
			want, err := imagecmp.LoadPNG(filepath.Join("..", "testdata", "matplotlib_ref", tc.name+".png"))
			if err != nil {
				t.Fatalf("load matplotlib reference: %v", err)
			}
			diag := alphaResidualDiagnostics(got, want, tc.axes, tc.threshold)
			t.Logf(
				"axes=%v threshold=%d highDiff=%d/%d (%.4f) meanAbs=%.3f rmse=%.3f bbox=%v",
				tc.axes,
				tc.threshold,
				diag.highDiff,
				diag.total,
				diag.highDiffRatio(),
				diag.meanAbs(),
				diag.rmse(),
				diag.bbox,
			)
			if diag.highDiffRatio() > tc.maxHighDiffRatio {
				t.Fatalf("high-diff ratio %.4f exceeds %.4f", diag.highDiffRatio(), tc.maxHighDiffRatio)
			}
		})
	}
}

type alphaResidualSummary struct {
	total      int
	highDiff   int
	sumAbs     float64
	sumSquared float64
	bbox       image.Rectangle
	haveBBox   bool
}

func (s alphaResidualSummary) highDiffRatio() float64 {
	if s.total == 0 {
		return 0
	}
	return float64(s.highDiff) / float64(s.total)
}

func (s alphaResidualSummary) meanAbs() float64 {
	if s.total == 0 {
		return 0
	}
	return s.sumAbs / float64(s.total)
}

func (s alphaResidualSummary) rmse() float64 {
	if s.total == 0 {
		return 0
	}
	return math.Sqrt(s.sumSquared / float64(s.total))
}

func alphaResidualDiagnostics(got, want image.Image, rect image.Rectangle, threshold uint8) alphaResidualSummary {
	bounds := got.Bounds().Intersect(want.Bounds()).Intersect(rect)
	summary := alphaResidualSummary{bbox: image.Rectangle{}}
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gotColor := color.RGBAModel.Convert(got.At(x, y)).(color.RGBA)
			wantColor := color.RGBAModel.Convert(want.At(x, y)).(color.RGBA)
			diffR := absByteDiff(gotColor.R, wantColor.R)
			diffG := absByteDiff(gotColor.G, wantColor.G)
			diffB := absByteDiff(gotColor.B, wantColor.B)
			diffA := absByteDiff(gotColor.A, wantColor.A)
			maxDiff := maxByte4(diffR, diffG, diffB, diffA)
			mean := float64(diffR+diffG+diffB+diffA) / 4.0
			summary.total++
			summary.sumAbs += mean
			summary.sumSquared += (float64(diffR)*float64(diffR) + float64(diffG)*float64(diffG) + float64(diffB)*float64(diffB) + float64(diffA)*float64(diffA)) / 4.0
			if maxDiff <= threshold {
				continue
			}
			summary.highDiff++
			pt := image.Pt(x, y)
			if !summary.haveBBox {
				summary.bbox = image.Rectangle{Min: pt, Max: pt.Add(image.Pt(1, 1))}
				summary.haveBBox = true
				continue
			}
			summary.bbox = summary.bbox.Union(image.Rectangle{Min: pt, Max: pt.Add(image.Pt(1, 1))})
		}
	}
	return summary
}

func fixtureAxesRect(width, height int, minX, minY, maxX, maxY float64) image.Rectangle {
	return image.Rect(
		int(math.Floor(float64(width)*minX)),
		int(math.Floor(float64(height)*(1-maxY))),
		int(math.Ceil(float64(width)*maxX)),
		int(math.Ceil(float64(height)*(1-minY))),
	)
}

func absByteDiff(a, b uint8) uint8 {
	if a > b {
		return a - b
	}
	return b - a
}

func maxByte4(a, b, c, d uint8) uint8 {
	out := a
	if b > out {
		out = b
	}
	if c > out {
		out = c
	}
	if d > out {
		out = d
	}
	return out
}

// === from histogram_height_profile_test.go ============================

const (
	histogramFigureWidth  = 640
	histogramFigureHeight = 360

	// Histogram axes rectangle in figure-fraction coordinates, shared by all histogram tests.
	histogramAxesMinX = 0.12
	histogramAxesMinY = 0.12
	histogramAxesMaxX = 0.95
	histogramAxesMaxY = 0.90
)

type histogramHeightSeries struct {
	edges   []float64
	heights []float64
}

type histogramHeightCase struct {
	name    string
	series  []histogramHeightSeries
	maxDiff int
}

func TestHistogramTopHeights_ParsedDataParity(t *testing.T) {
	payload := loadMatplotlibRNGDebugPayload(t)
	getSeries := func(name string) histogramHeightSeries {
		data, ok := payload.HistogramData[name]
		if !ok {
			t.Fatalf("missing payload series %s", name)
		}
		if len(data.Edges) < 2 || len(data.Heights) == 0 {
			t.Fatalf("invalid payload series %s: edges=%d heights=%d", name, len(data.Edges), len(data.Heights))
		}
		return histogramHeightSeries{
			edges:   data.Edges,
			heights: data.Heights,
		}
	}

	basic := getSeries("hist_basic")
	density := getSeries("hist_density")
	strat1 := getSeries("hist_strategies_data1")
	strat2 := getSeries("hist_strategies_data2")

	axes := geom.Rect{
		Min: geom.Pt{X: histogramAxesMinX, Y: histogramAxesMinY},
		Max: geom.Pt{X: histogramAxesMaxX, Y: histogramAxesMaxY},
	}

	tests := []histogramHeightCase{
		{
			name:    "hist_basic",
			series:  []histogramHeightSeries{basic},
			maxDiff: 1,
		},
		{
			name:    "hist_density",
			series:  []histogramHeightSeries{density},
			maxDiff: 1,
		},
		{
			name:    "hist_strategies",
			series:  []histogramHeightSeries{strat1, strat2},
			maxDiff: 1,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			expected := expectedHistogramTopProfile(t, axes, tc.series)
			skipColumn := histogramBoundaryColumns(axes, tc.series, 1)

			got, _, err := parity.Render(tc.name)
			if err != nil {
				t.Fatalf("render parity example %s: %v", tc.name, err)
			}
			assertHistogramTopProfile(t, "got", got, expected, tc.maxDiff, skipColumn)

			goldenPath := filepath.Join("..", "testdata", "golden", tc.name+".png")
			golden, err := imagecmp.LoadPNG(goldenPath)
			if err != nil {
				t.Fatalf("load golden image %s: %v", goldenPath, err)
			}
			assertHistogramTopProfile(t, "golden", golden, expected, tc.maxDiff, skipColumn)

			refPath := filepath.Join("..", "testdata", "matplotlib_ref", tc.name+".png")
			ref, err := imagecmp.LoadPNG(refPath)
			if err != nil {
				t.Fatalf("load matplotlib reference image %s: %v", refPath, err)
			}
			assertHistogramTopProfile(t, "matplotlib_ref", ref, expected, tc.maxDiff, skipColumn)
		})
	}
}

func histogramBoundaryColumns(axes geom.Rect, series []histogramHeightSeries, pxTolerance int) []bool {
	ax := histogramAxesPixelRect(axes)

	xMin := math.Inf(1)
	xMax := math.Inf(-1)
	yMax := 0.0
	for _, s := range series {
		if len(s.edges) < 2 {
			continue
		}
		if s.edges[0] < xMin {
			xMin = s.edges[0]
		}
		if s.edges[len(s.edges)-1] > xMax {
			xMax = s.edges[len(s.edges)-1]
		}
		for _, h := range s.heights {
			if h > yMax {
				yMax = h
			}
		}
	}
	if xMin >= xMax || yMax <= 0 {
		return nil
	}

	xSpan := xMax - xMin
	xMin -= xSpan * 0.05
	xMax += xSpan * 0.05
	yScaleMax := yMax * 1.05

	xScale := transform.NewLinear(xMin, xMax)
	yScale := transform.NewLinear(0, yScaleMax)
	transform2D := core.Transform2D{
		XScale:      xScale,
		YScale:      yScale,
		AxesToPixel: transform.NewAffine(histogramAxesToPixel(ax)),
	}

	xStart := int(math.Ceil(ax.Min.X))
	xEnd := int(math.Floor(ax.Max.X))
	if xStart < 0 {
		xStart = 0
	}
	if xEnd > histogramFigureWidth {
		xEnd = histogramFigureWidth
	}
	skip := make([]bool, xEnd-xStart)

	for _, s := range series {
		for _, edge := range s.edges {
			pt := transform2D.Apply(geom.Pt{X: edge, Y: 0})
			center := int(math.Round(pt.X))
			for delta := -pxTolerance; delta <= pxTolerance; delta++ {
				x := center + delta
				if x < xStart || x >= xEnd {
					continue
				}
				skip[x-xStart] = true
			}
		}
	}

	return skip
}

func expectedHistogramTopProfile(t *testing.T, axes geom.Rect, series []histogramHeightSeries) []int {
	t.Helper()

	if len(series) == 0 {
		t.Fatal("expected at least one series")
	}

	xMin := math.Inf(1)
	xMax := math.Inf(-1)
	yMax := 0.0
	for _, s := range series {
		if len(s.edges) < 2 {
			continue
		}
		if s.edges[0] < xMin {
			xMin = s.edges[0]
		}
		if s.edges[len(s.edges)-1] > xMax {
			xMax = s.edges[len(s.edges)-1]
		}
		for _, h := range s.heights {
			if h > yMax {
				yMax = h
			}
		}
	}
	if xMin >= xMax || yMax <= 0 {
		t.Fatal("invalid histogram geometry for profile expectations")
	}
	axMargin := 0.05

	ax := histogramAxesPixelRect(axes)
	xSpan := xMax - xMin
	xMin -= xSpan * axMargin
	xMax += xSpan * axMargin
	yScaleMax := yMax
	yScaleMax *= (1.0 + axMargin)

	xScale := transform.NewLinear(xMin, xMax)
	yScale := transform.NewLinear(0, yScaleMax)
	transform2D := core.Transform2D{
		XScale:      xScale,
		YScale:      yScale,
		AxesToPixel: transform.NewAffine(histogramAxesToPixel(ax)),
	}

	xStart := int(math.Ceil(ax.Min.X))
	if xStart < 0 {
		xStart = 0
	}
	xEnd := int(math.Floor(ax.Max.X))
	if xEnd > histogramFigureWidth {
		xEnd = histogramFigureWidth
	}
	height := make([]int, xEnd-xStart)
	for i := range height {
		height[i] = -1
	}

	for _, s := range series {
		if len(s.edges) < 2 || len(s.heights) == 0 {
			continue
		}

		for binIdx := 0; binIdx < len(s.heights) && binIdx+1 < len(s.edges); binIdx++ {
			h := s.heights[binIdx]
			if h <= 0 {
				continue
			}

			leftPt := transform2D.Apply(geom.Pt{X: s.edges[binIdx], Y: 0})
			rightPt := transform2D.Apply(geom.Pt{X: s.edges[binIdx+1], Y: 0})
			topPt := transform2D.Apply(geom.Pt{X: s.edges[binIdx], Y: h})

			left := math.Min(leftPt.X, rightPt.X)
			right := math.Max(leftPt.X, rightPt.X)
			top := int(math.Round(topPt.Y))

			for x := xStart; x < xEnd; x++ {
				colLeft := float64(x)
				colRight := float64(x + 1)
				if colRight <= left || colLeft >= right {
					continue
				}
				idx := x - xStart
				if height[idx] < 0 || top < height[idx] {
					height[idx] = top
				}
			}
		}
	}

	return height
}

func assertHistogramTopProfile(
	t *testing.T,
	name string,
	img image.Image,
	expected []int,
	maxDiff int,
	skip []bool,
) {
	t.Helper()

	ax := histogramAxesPixelRect(geom.Rect{
		Min: geom.Pt{X: histogramAxesMinX, Y: histogramAxesMinY},
		Max: geom.Pt{X: histogramAxesMaxX, Y: histogramAxesMaxY},
	})
	observed := extractHistogramTopProfile(img, ax)

	if len(skip) != len(observed) || len(skip) != len(expected) {
		t.Fatal("histogram boundary skip mask length mismatch")
	}

	if len(observed) != len(expected) {
		t.Fatalf("%s histogram profile width mismatch: got %d, expected %d", name, len(observed), len(expected))
	}

	mismatches := 0
	for i := range expected {
		if skip[i] {
			continue
		}
		x := i + int(math.Ceil(histogramAxesMinX*float64(histogramFigureWidth)))
		e := expected[i]
		o := observed[i]
		if e < 0 {
			if o >= 0 {
				mismatches++
				if mismatches <= 8 {
					t.Logf("%s unexpected top at x=%d: got=%d expected=%d", name, x, o, e)
				}
			}
			continue
		}
		if o < 0 {
			mismatches++
			if mismatches <= 8 {
				t.Logf("%s profile missing top at x=%d (expected=%d)", name, x, e)
			}
			continue
		}
		diff := int(math.Abs(float64(o - e)))
		if diff > maxDiff {
			mismatches++
			if mismatches <= 8 {
				t.Logf("%s profile mismatch at x=%d: got=%d expected=%d diff=%d", name, x, o, e, diff)
			}
		}
	}

	if mismatches > 0 {
		t.Fatalf("%s histogram height profile mismatch count=%d", name, mismatches)
	}
}

func histogramAxesPixelRect(axes geom.Rect) geom.Rect {
	return geom.Rect{
		Min: geom.Pt{
			X: float64(histogramFigureWidth) * axes.Min.X,
			Y: float64(histogramFigureHeight) * (1 - axes.Max.Y),
		},
		Max: geom.Pt{
			X: float64(histogramFigureWidth) * axes.Max.X,
			Y: float64(histogramFigureHeight) * (1 - axes.Min.Y),
		},
	}
}

func extractHistogramTopProfile(img image.Image, axes geom.Rect) []int {
	bounds := img.Bounds()
	xStart := int(math.Ceil(axes.Min.X))
	yStart := int(math.Ceil(axes.Min.Y))
	xEnd := int(math.Floor(axes.Max.X))
	yEnd := int(math.Floor(axes.Max.Y)) - 1

	if xStart < bounds.Min.X {
		xStart = bounds.Min.X
	}
	if xEnd > bounds.Max.X {
		xEnd = bounds.Max.X
	}
	if yStart < bounds.Min.Y {
		yStart = bounds.Min.Y
	}
	if yEnd > bounds.Max.Y {
		yEnd = bounds.Max.Y
	}

	profile := make([]int, xEnd-xStart)
	for i := range profile {
		profile[i] = -1
	}

	for x := xStart; x < xEnd; x++ {
		for y := yStart; y < yEnd; y++ {
			rgba := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
			if isHistogramFillPixel(rgba) {
				profile[x-xStart] = y
				break
			}
		}
	}

	return profile
}

func isHistogramFillPixel(p color.RGBA) bool {
	if p.A == 0 {
		return false
	}
	// Discard pure/near-white background.
	if p.R > 248 && p.G > 248 && p.B > 248 {
		return false
	}

	maxCh := p.R
	if p.G > maxCh {
		maxCh = p.G
	}
	if p.B > maxCh {
		maxCh = p.B
	}
	minCh := p.R
	if p.G < minCh {
		minCh = p.G
	}
	if p.B < minCh {
		minCh = p.B
	}
	if maxCh-minCh <= 2 {
		return false
	}

	return true
}

func histogramAxesToPixel(rect geom.Rect) geom.Affine {
	sx := rect.W()
	sy := -rect.H()
	tx := rect.Min.X
	ty := rect.Max.Y
	return geom.Affine{A: sx, D: sy, E: tx, F: ty}
}

// === from rng_parity_test.go ==========================================

type rngDebugPayload struct {
	NormalData    map[string][]float64 `json:"normal_data"`
	UniformData   map[string][]float64 `json:"uniform_data"`
	HistogramData map[string]struct {
		Edges   []float64 `json:"edges"`
		Heights []float64 `json:"heights"`
	} `json:"histogram_data"`
}

func TestHistogramBarHeightParity(t *testing.T) {
	payload := loadMatplotlibRNGDebugPayload(t)

	basic := normalData(42, 0, 500, 5.0, 1.5)
	density := basic
	strategies1 := normalData(42, 0, 300, 4.0, 1.0)
	strategies2 := normalData(7, 0, 300, 7.0, 1.2)

	wantBasic, ok := payload.HistogramData["hist_basic"]
	if !ok {
		t.Fatalf("missing payload hist_basic")
	}
	wantDensity, ok := payload.HistogramData["hist_density"]
	if !ok {
		t.Fatalf("missing payload hist_density")
	}
	wantStrategies1, ok := payload.HistogramData["hist_strategies_data1"]
	if !ok {
		t.Fatalf("missing payload hist_strategies_data1")
	}
	wantStrategies2, ok := payload.HistogramData["hist_strategies_data2"]
	if !ok {
		t.Fatalf("missing payload hist_strategies_data2")
	}

	gotEdges, gotHeights := goHistogramBinCounts(basic, 0, core.HistNormCount, core.BinStrategySturges)
	compareHistogramHeights(t, "hist_basic", gotEdges, gotHeights, wantBasic.Edges, wantBasic.Heights, 1e-12)

	gotEdges, gotHeights = goHistogramBinCounts(density, 20, core.HistNormDensity, core.BinStrategyAuto)
	compareHistogramHeights(t, "hist_density", gotEdges, gotHeights, wantDensity.Edges, wantDensity.Heights, 1e-12)

	gotEdges, gotHeights = goHistogramBinCounts(strategies1, 15, core.HistNormProbability, core.BinStrategyAuto)
	compareHistogramHeights(t, "hist_strategies_data1", gotEdges, gotHeights, wantStrategies1.Edges, wantStrategies1.Heights, 1e-12)

	gotEdges, gotHeights = goHistogramBinCounts(strategies2, 15, core.HistNormProbability, core.BinStrategyAuto)
	compareHistogramHeights(t, "hist_strategies_data2", gotEdges, gotHeights, wantStrategies2.Edges, wantStrategies2.Heights, 1e-12)
}

func TestHistogramRNGParity(t *testing.T) {
	payload := loadMatplotlibRNGDebugPayload(t)

	compareFloatSlices(t, payload.UniformData["pcg_42_0_1000"], pcgFloat64Samples(42, 0, 1000), "pcg_42_0_1000", 1e-15)
	compareFloatSlices(t, payload.UniformData["pcg_7_0_600"], pcgFloat64Samples(7, 0, 600), "pcg_7_0_600", 1e-15)

	wantBasic := normalData(42, 0, 500, 5.0, 1.5)
	wantDensity := normalData(42, 0, 500, 5.0, 1.5)
	wantStrategies1 := normalData(42, 0, 300, 4.0, 1.0)
	wantStrategies2 := normalData(7, 0, 300, 7.0, 1.2)

	compareFloatSlices(t, payload.NormalData["hist_basic"], wantBasic, "hist_basic", 1e-12)
	compareFloatSlices(t, payload.NormalData["hist_density"], wantDensity, "hist_density", 1e-12)
	compareFloatSlices(t, payload.NormalData["hist_strategies_data1"], wantStrategies1, "hist_strategies_data1", 1e-12)
	compareFloatSlices(t, payload.NormalData["hist_strategies_data2"], wantStrategies2, "hist_strategies_data2", 1e-12)

	// Verify RNG values are consumed in pair order for normal transformation:
	// first sample uses u1/u2 at indices 0/1, second sample uses 2/3, etc.
	checkHistogramRNGOrder(t, payload.UniformData["pcg_42_0_1000"], wantBasic, 5.0, 1.5, 3)
	checkHistogramRNGOrder(t, payload.UniformData["pcg_7_0_600"], wantStrategies2, 7.0, 1.2, 3)
}

func loadMatplotlibRNGDebugPayload(t *testing.T) rngDebugPayload {
	t.Helper()

	script := filepath.Join("matplotlib_ref", "generate.py")
	cmd := selectPythonCommand(t, "uv", script, "--emit-rng-debug")
	if cmd == nil {
		cmd = selectPythonCommand(t, "python3", script, "--emit-rng-debug")
	}
	if cmd == nil {
		t.Skip("matplotlib RNG parity check skipped: uv/python3 not available")
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run matplotlib RNG debug generator: %v\n%s", err, strings.TrimSpace(string(out)))
	}
	outStr := strings.TrimSpace(string(out))
	jsonStart := strings.Index(outStr, "{")
	if jsonStart == -1 {
		t.Fatalf("failed to parse RNG debug JSON: no JSON object in output\n%s", outStr)
	}

	var payload rngDebugPayload
	if err := json.Unmarshal([]byte(outStr[jsonStart:]), &payload); err != nil {
		t.Fatalf("failed to parse RNG debug JSON: %v\n%s", err, outStr[jsonStart:])
	}

	return payload
}

func selectPythonCommand(t *testing.T, name, script string, args ...string) *exec.Cmd {
	t.Helper()
	path, err := exec.LookPath(name)
	if err != nil {
		return nil
	}

	var cmd *exec.Cmd
	if name == "uv" {
		allArgs := append([]string{"run", script}, args...)
		cmd = exec.Command(path, allArgs...)
	} else {
		allArgs := append([]string{script}, args...)
		cmd = exec.Command(path, allArgs...)
	}
	return cmd
}

func compareFloatSlices(t *testing.T, got, want []float64, label string, eps float64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s length mismatch: got %d, want %d", label, len(got), len(want))
	}
	for i := range got {
		if math.Abs(got[i]-want[i]) > eps {
			t.Fatalf("%s mismatch at index %d: got %.17g, want %.17g", label, i, got[i], want[i])
		}
	}
}

func compareHistogramHeights(
	t *testing.T,
	label string,
	gotEdges, gotHeights, wantEdges, wantHeights []float64,
	eps float64,
) {
	t.Helper()
	compareFloatSlices(t, gotEdges, wantEdges, label+" edges", eps)
	compareFloatSlices(t, gotHeights, wantHeights, label+" heights", eps)
}

func pcgFloat64Samples(seed1, seed2 uint64, n int) []float64 {
	rng := rand.New(rand.NewPCG(seed1, seed2))
	out := make([]float64, n)
	for i := range out {
		out[i] = rng.Float64()
	}
	return out
}

func checkHistogramRNGOrder(t *testing.T, uniforms, normals []float64, mean, std float64, maxSamples int) {
	t.Helper()
	if maxSamples <= 0 {
		return
	}
	if len(normals) < maxSamples {
		t.Fatalf("normal sample count %d is less than maxSamples %d", len(normals), maxSamples)
	}
	if len(uniforms) < maxSamples*2 {
		t.Fatalf("uniform sample count %d is less than needed %d", len(uniforms), maxSamples*2)
	}

	for i := 0; i < maxSamples; i++ {
		want := normalFromUniformPair(uniforms[i*2], uniforms[i*2+1], mean, std)
		if math.Abs(normals[i]-want) > 1e-15 {
			t.Fatalf("pair-order mismatch at sample %d: got %.17g, want %.17g", i, normals[i], want)
		}
	}
}

func normalFromUniformPair(u1, u2, mean, std float64) float64 {
	return math.Sqrt(-2*math.Log(u1))*math.Cos(2*math.Pi*u2)*std + mean
}

func goHistogramBinCounts(data []float64, bins int, norm core.HistNorm, strategy core.BinStrategy) ([]float64, []float64) {
	h := &core.Hist2D{
		Data:     data,
		Bins:     bins,
		Norm:     norm,
		BinStrat: strategy,
	}
	return h.BinCounts()
}

// === from bar_text_diagnostic_test.go =========================
//
// Records DrawText calls from the AGG renderer and compares the resulting
// metrics against Matplotlib's text placement for the same figure. Skipped
// unless MPL_GO_TEXT_DIAG is set; also skipped under purego builds where
// the AGG renderer's FreeType pipeline isn't compiled in (without it the
// metrics would be apples-to-oranges).

type textDrawRecord struct {
	Text    string                   `json:"text"`
	Size    float64                  `json:"size"`
	Origin  geom.Pt                  `json:"origin"`
	Metrics render.TextMetrics       `json:"metrics"`
	Heights render.FontHeightMetrics `json:"heights,omitempty"`
	Bounds  render.TextBounds        `json:"bounds"`
	Ink     geom.Rect                `json:"ink"`
}

type textRecordingRenderer struct {
	*agg.Renderer
	records []textDrawRecord
}

func (r *textRecordingRenderer) DrawText(text string, origin geom.Pt, size float64, color render.Color) {
	bounds, _ := r.MeasureTextBounds(text, size, "")
	metrics := r.MeasureText(text, size, "")
	heights, _ := r.MeasureFontHeights(size, "")
	r.records = append(r.records, textDrawRecord{
		Text:    text,
		Size:    size,
		Origin:  origin,
		Metrics: metrics,
		Heights: heights,
		Bounds:  bounds,
		Ink: geom.Rect{
			Min: geom.Pt{X: origin.X + bounds.X, Y: origin.Y + bounds.Y},
			Max: geom.Pt{X: origin.X + bounds.X + bounds.W, Y: origin.Y + bounds.Y + bounds.H},
		},
	})
	r.Renderer.DrawText(text, origin, size, color)
}

func (r *textRecordingRenderer) DrawTextWithFont(text string, origin geom.Pt, size float64, color render.Color, fontKey string) {
	bounds, _ := r.MeasureTextBounds(text, size, fontKey)
	metrics := r.MeasureText(text, size, fontKey)
	heights, _ := r.MeasureFontHeights(size, fontKey)
	r.records = append(r.records, textDrawRecord{
		Text:    text,
		Size:    size,
		Origin:  origin,
		Metrics: metrics,
		Heights: heights,
		Bounds:  bounds,
		Ink: geom.Rect{
			Min: geom.Pt{X: origin.X + bounds.X, Y: origin.Y + bounds.Y},
			Max: geom.Pt{X: origin.X + bounds.X + bounds.W, Y: origin.Y + bounds.Y + bounds.H},
		},
	})
	r.Renderer.DrawTextWithFont(text, origin, size, color, fontKey)
}

func TestBarBasicTextPlacementDiagnostic(t *testing.T) {
	if os.Getenv("MPL_GO_TEXT_DIAG") == "" {
		t.Skip("set MPL_GO_TEXT_DIAG=1 to log Go vs Matplotlib text placement")
	}
	if agg.NativeFreetypeVersion() == "" {
		t.Skip("bar text diagnostic requires the cgo FreeType pipeline; build without -tags purego")
	}

	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.SetXLim(0, 6)
	ax.SetYLim(0, 10)
	ax.SetTitle("Basic Bars")

	base, err := agg.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatalf("create AGG renderer: %v", err)
	}
	rec := &textRecordingRenderer{Renderer: base}
	core.DrawFigure(fig, rec)

	goPayload, err := json.MarshalIndent(rec.records, "", "  ")
	if err != nil {
		t.Fatalf("marshal Go records: %v", err)
	}
	t.Logf("go text records:\n%s", goPayload)
	for _, size := range []float64{10, 12} {
		metrics := rec.MeasureText("lp", size, "")
		bounds, _ := rec.MeasureTextBounds("lp", size, "")
		heights, _ := rec.MeasureFontHeights(size, "")
		t.Logf("go lp size %.0f: metrics=%+v bounds=%+v heights=%+v", size, metrics, bounds, heights)
	}

	python, err := matplotlibPythonPathForDiag(t)
	if err != nil {
		t.Skipf("Matplotlib Python unavailable: %v", err)
	}
	mplRecords := runMatplotlibBarTextDiagnostic(t, python)
	mplPayload, err := json.MarshalIndent(mplRecords, "", "  ")
	if err != nil {
		t.Fatalf("marshal Matplotlib records: %v", err)
	}
	t.Logf("matplotlib text records:\n%s", mplPayload)
}

func matplotlibPythonPathForDiag(t *testing.T) (string, error) {
	t.Helper()
	candidates := []string{}
	if env := os.Getenv("MATPLOTLIB_GO_PYTHON"); env != "" {
		candidates = append(candidates, env)
	}
	candidates = append(candidates, "/usr/bin/python3", "python3")
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		cmd := exec.Command(candidate, "-c", "import matplotlib")
		cmd.Env = append(os.Environ(), "MPLCONFIGDIR="+t.TempDir())
		if err := cmd.Run(); err == nil {
			return candidate, nil
		}
	}
	return "", exec.ErrNotFound
}

func runMatplotlibBarTextDiagnostic(t *testing.T, python string) []map[string]any {
	t.Helper()
	repoRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	script := `
import json
import matplotlib
matplotlib.use("Agg")
from matplotlib.backends.backend_agg import FigureCanvasAgg
from test.matplotlib_ref.common import _bar_basic_scaffold

fig, ax = _bar_basic_scaffold(show_ticks=True, show_tick_labels=True, show_title=True)
canvas = FigureCanvasAgg(fig)
canvas.draw()
renderer = canvas.get_renderer()
height = fig.bbox.height

records = []
for group, labels in [
    ("x", ax.get_xticklabels()),
    ("y", ax.get_yticklabels()),
    ("title", [ax.title]),
]:
    for label in labels:
        if not label.get_visible() or label.get_text() == "":
            continue
        bbox = label.get_window_extent(renderer=renderer)
        anchor = label.get_transform().transform(label.get_position())
        records.append({
            "group": group,
            "text": label.get_text(),
            "fontsize": label.get_fontsize(),
            "anchor": {"x": float(anchor[0]), "y": float(height - anchor[1])},
            "bbox": {
                "min": {"x": float(bbox.x0), "y": float(height - bbox.y1)},
                "max": {"x": float(bbox.x1), "y": float(height - bbox.y0)},
            },
            "ha": label.get_ha(),
            "va": label.get_va(),
        })
print(json.dumps(records))
`
	cmd := exec.Command(python, "-c", script)
	cmd.Env = append(os.Environ(), "MPLCONFIGDIR="+t.TempDir(), "PYTHONPATH="+repoRoot)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			t.Fatalf("run Matplotlib text diagnostic: %v\n%s", err, exitErr.Stderr)
		}
		t.Fatalf("run Matplotlib text diagnostic: %v", err)
	}
	var records []map[string]any
	if err := json.Unmarshal(out, &records); err != nil {
		t.Fatalf("decode Matplotlib text diagnostic: %v\n%s", err, out)
	}
	return records
}
