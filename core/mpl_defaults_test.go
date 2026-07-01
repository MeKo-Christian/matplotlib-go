package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

// Phase 19 default-value fidelity: an unstyled plot must use matplotlib
// 3.10.9's rc defaults. Each literal below is the matplotlib default.

func TestDefaultLineWidthMatchesMatplotlib(t *testing.T) {
	// rcParams["lines.linewidth"] == 1.5
	if got := style.Default.LineWidth; got != 1.5 {
		t.Fatalf("style.Default.LineWidth = %v, want 1.5 (matplotlib lines.linewidth)", got)
	}

	fig := NewFigure(100, 100)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})
	line := ax.Plot([]float64{0, 1}, []float64{0, 1})
	if line == nil {
		t.Fatal("Plot returned nil")
	}
	if line.W != 1.5 {
		t.Fatalf("unstyled Plot line width = %v, want 1.5", line.W)
	}
}

func TestDefaultHistBinsMatchesMatplotlib(t *testing.T) {
	// rcParams["hist.bins"] == 10
	if got := style.Default.HistBins; got != 10 {
		t.Fatalf("style.Default.HistBins = %v, want 10 (matplotlib hist.bins)", got)
	}

	data := make([]float64, 200)
	for i := range data {
		data[i] = float64(i % 50)
	}

	fig := NewFigure(100, 100)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})
	hist := ax.Hist(data)
	if hist == nil {
		t.Fatal("Hist returned nil")
	}
	edges, counts := hist.BinCounts()
	if len(counts) != 10 || len(edges) != 11 {
		t.Fatalf("unstyled Hist bins = %d (edges %d), want 10", len(counts), len(edges))
	}

	// A bare Hist2D (zero-value BinStrat) uses the same matplotlib default.
	bare := &Hist2D{Data: data}
	_, counts = bare.BinCounts()
	if len(counts) != 10 {
		t.Fatalf("bare Hist2D bins = %d, want 10", len(counts))
	}
}

func TestDefaultScatterSizeMatchesMatplotlib(t *testing.T) {
	// matplotlib scatter default: s = lines.markersize^2 = 6^2 = 36 pt^2.
	s := &Scatter2D{XY: []geom.Pt{{X: 1, Y: 1}}}
	if got := s.effectiveSize(); got != 36 {
		t.Fatalf("unset Scatter2D size = %v, want 36", got)
	}
	ctx := createTestDrawContext()
	pc := s.toPathCollection(nil, ctx)
	if pc == nil {
		t.Fatal("toPathCollection returned nil")
	}
	// 36 pt^2 -> sqrt(36) pt = 6 pt -> 6 * DPI/72 device pixels.
	want := 6.0 * ctx.RC.DPI / 72.0
	if math.Abs(pc.Size-want) > 1e-12 {
		t.Fatalf("collection size = %v px, want %v (36 pt^2 at DPI %v)", pc.Size, want, ctx.RC.DPI)
	}

	// Explicit sizes stay untouched.
	s = &Scatter2D{XY: []geom.Pt{{X: 1, Y: 1}}, Size: 100}
	if got := s.effectiveSize(); got != 100 {
		t.Fatalf("explicit Scatter2D size = %v, want 100", got)
	}
}

func TestDefaultMinorTickSizeMatchesMatplotlib(t *testing.T) {
	// rcParams["xtick.minor.size"] == rcParams["ytick.minor.size"] == 2.0 pt.
	if defaultMinorTickSizePt != 2.0 {
		t.Fatalf("defaultMinorTickSizePt = %v, want 2.0", defaultMinorTickSizePt)
	}
	want := 2.0 * 100.0 / 72.0
	for _, ax := range []*Axis{NewXAxis(), NewYAxis()} {
		if got := ax.minorTickSize(); math.Abs(got-want) > 1e-12 {
			t.Fatalf("side %v minorTickSize = %v px, want %v (2.0 pt at 100 DPI)", ax.Side, got, want)
		}
	}
}

func TestDefaultMinorTickPadMatchesMatplotlib(t *testing.T) {
	// rcParams["xtick.minor.pad"] == rcParams["ytick.minor.pad"] == 3.4 pt
	// vs the major 3.5 pt.
	ax := NewXAxis()
	ctx := createTestDrawContext()

	majorPad := tickLabelPadForAxisSize(ax, 0, ax.MajorLabelStyle, ctx)
	minorPad := tickLabelPadForAxisSize(ax, 0, ax.MinorLabelStyle, ctx)
	if want := 3.5 * ctx.RC.DPI / 72.0; math.Abs(majorPad-want) > 1e-12 {
		t.Fatalf("major pad = %v px, want %v (3.5 pt)", majorPad, want)
	}
	if want := 3.4 * ctx.RC.DPI / 72.0; math.Abs(minorPad-want) > 1e-12 {
		t.Fatalf("minor pad = %v px, want %v (3.4 pt)", minorPad, want)
	}
}

func TestTickLabelPadNoContextAssumes100DPI(t *testing.T) {
	// The no-context fallback uses matplotlib's default figure DPI of 100,
	// not 96.
	got := tickLabelPadForSize(0, TickLabelStyle{}, nil)
	if want := 3.5 * 100.0 / 72.0; math.Abs(got-want) > 1e-12 {
		t.Fatalf("no-context tick label pad = %v px, want %v", got, want)
	}
}

func TestMPLStyleHistBinsReachesHist(t *testing.T) {
	data := make([]float64, 200)
	for i := range data {
		data[i] = float64(i % 50)
	}

	cases := []struct {
		style string
		want  int // expected bin count; numpy values verified with numpy 1.26.4
	}{
		{"hist.bins: 20\n", 20},
		{"hist.bins: auto\n", 9}, // np.histogram_bin_edges(data, bins="auto")
	}
	for _, tc := range cases {
		theme, report, err := style.ParseMPLStyle("test", tc.style)
		if err != nil {
			t.Fatalf("ParseMPLStyle(%q): %v", tc.style, err)
		}
		if len(report.Unsupported) > 0 {
			t.Fatalf("ParseMPLStyle(%q) unsupported entries: %+v", tc.style, report.Unsupported)
		}
		fig := NewFigure(100, 100, style.WithTheme(theme))
		ax := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})
		hist := ax.Hist(data)
		if hist == nil {
			t.Fatal("Hist returned nil")
		}
		_, counts := hist.BinCounts()
		if len(counts) != tc.want {
			t.Fatalf("%q Hist bins = %d, want %d", tc.style, len(counts), tc.want)
		}
	}
}

func TestMPLStyleLinesLinewidthReachesPlot(t *testing.T) {
	theme, report, err := style.ParseMPLStyle("test", "lines.linewidth: 3.25\n")
	if err != nil {
		t.Fatalf("ParseMPLStyle: %v", err)
	}
	if len(report.Unsupported) > 0 {
		t.Fatalf("ParseMPLStyle unsupported entries: %+v", report.Unsupported)
	}

	fig := NewFigure(100, 100, style.WithTheme(theme))
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})
	line := ax.Plot([]float64{0, 1}, []float64{0, 1})
	if line == nil {
		t.Fatal("Plot returned nil")
	}
	if line.W != 3.25 {
		t.Fatalf("styled Plot line width = %v, want 3.25 from lines.linewidth", line.W)
	}

	// An explicit option still wins over the rc value.
	w := 0.7
	line = ax.Plot([]float64{0, 1}, []float64{1, 0}, PlotOptions{LineWidth: &w})
	if line.W != 0.7 {
		t.Fatalf("explicit LineWidth = %v, want 0.7", line.W)
	}
}

func TestMPLStyleImageOriginReachesImShow(t *testing.T) {
	theme, report, err := style.ParseMPLStyle("test", "image.origin: lower\n")
	if err != nil {
		t.Fatalf("ParseMPLStyle: %v", err)
	}
	if len(report.Unsupported) > 0 {
		t.Fatalf("unsupported entries: %+v", report.Unsupported)
	}

	data := [][]float64{{0, 1}, {2, 3}}

	fig := NewFigure(100, 100, style.WithTheme(theme))
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})
	img := ax.ImShow(data)
	if img == nil {
		t.Fatal("ImShow returned nil")
	}
	if img.Origin != ImageOriginLower {
		t.Fatalf("ImShow origin = %v, want ImageOriginLower from image.origin", img.Origin)
	}
	if ax.YInverted() {
		t.Fatal("image.origin: lower must not invert the y-axis")
	}

	// MatShow pins origin=upper regardless of rc, mirroring matplotlib matshow.
	ax2 := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})
	mat := ax2.MatShow(data)
	if mat == nil {
		t.Fatal("MatShow returned nil")
	}
	if mat.Origin != ImageOriginUpper {
		t.Fatalf("MatShow origin = %v, want ImageOriginUpper regardless of rc", mat.Origin)
	}
}

func TestMPLStyleImageOriginReachesImShowRGB(t *testing.T) {
	theme, _, err := style.ParseMPLStyle("test", "image.origin: lower\n")
	if err != nil {
		t.Fatalf("ParseMPLStyle: %v", err)
	}

	rgb := [][][]float64{
		{{1, 0, 0}, {0, 1, 0}},
		{{0, 0, 1}, {1, 1, 1}},
	}

	fig := NewFigure(100, 100, style.WithTheme(theme))
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})
	img := ax.ImShowRGB(rgb)
	if img == nil {
		t.Fatal("ImShowRGB returned nil")
	}
	if img.Origin != ImageOriginLower {
		t.Fatalf("ImShowRGB origin = %v, want ImageOriginLower from image.origin", img.Origin)
	}
}

func TestMPLStyleImageAspectReachesImShow(t *testing.T) {
	data := [][]float64{{0, 1}, {2, 3}}

	cases := []struct {
		style     string
		wantMode  string
		wantValue float64
	}{
		{"image.aspect: auto\n", "auto", 1},
		{"image.aspect: 3.0\n", "ratio", 3},
	}
	for _, tc := range cases {
		theme, report, err := style.ParseMPLStyle("test", tc.style)
		if err != nil {
			t.Fatalf("ParseMPLStyle(%q): %v", tc.style, err)
		}
		if len(report.Unsupported) > 0 {
			t.Fatalf("ParseMPLStyle(%q) unsupported entries: %+v", tc.style, report.Unsupported)
		}

		fig := NewFigure(100, 100, style.WithTheme(theme))
		ax := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})
		if img := ax.ImShow(data); img == nil {
			t.Fatal("ImShow returned nil")
		}
		if ax.aspectMode != tc.wantMode || ax.aspectValue != tc.wantValue {
			t.Fatalf("%q ImShow aspect = %s/%v, want %s/%v", tc.style, ax.aspectMode, ax.aspectValue, tc.wantMode, tc.wantValue)
		}

		// An explicit option still wins over the rc value.
		ax2 := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})
		if img := ax2.ImShow(data, ImShowOptions{Aspect: "equal"}); img == nil {
			t.Fatal("ImShow returned nil")
		}
		if ax2.aspectMode != "equal" {
			t.Fatalf("explicit Aspect=equal got mode %q", ax2.aspectMode)
		}
	}
}

func TestMPLStyleAxisBelowReachesGrid(t *testing.T) {
	theme, report, err := style.ParseMPLStyle("test", "axes.axisbelow: True\naxes.grid: True\n")
	if err != nil {
		t.Fatalf("ParseMPLStyle: %v", err)
	}
	if len(report.Unsupported) > 0 {
		t.Fatalf("unsupported entries: %+v", report.Unsupported)
	}

	fig := NewFigure(100, 100, style.WithTheme(theme))
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})
	found := false
	for _, art := range ax.Artists {
		if grid, ok := art.(*Grid); ok {
			found = true
			if grid.z != 0.5 {
				t.Fatalf("grid z = %v, want 0.5 from axes.axisbelow: True", grid.z)
			}
		}
	}
	if !found {
		t.Fatal("no grid artist found")
	}
}

func TestMPLStyleMarginsReachAutoscale(t *testing.T) {
	theme, _, err := style.ParseMPLStyle("test", "axes.xmargin: 0.1\naxes.ymargin: 0\n")
	if err != nil {
		t.Fatalf("ParseMPLStyle: %v", err)
	}

	fig := NewFigure(100, 100, style.WithTheme(theme))
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})
	ax.Plot([]float64{0, 10}, []float64{0, 10})

	xMin, xMax := ax.XScale.Domain()
	yMin, yMax := ax.YScale.Domain()
	if !floatApprox(xMin, -1, 1e-12) || !floatApprox(xMax, 11, 1e-12) {
		t.Fatalf("x limits = [%v, %v], want [-1, 11] from axes.xmargin", xMin, xMax)
	}
	if !floatApprox(yMin, 0, 1e-12) || !floatApprox(yMax, 10, 1e-12) {
		t.Fatalf("y limits = [%v, %v], want [0, 10] from axes.ymargin", yMin, yMax)
	}
}

func TestMPLStyleAutolimitModeReachesAxes(t *testing.T) {
	theme, _, err := style.ParseMPLStyle("test", "axes.autolimit_mode: round_numbers\n")
	if err != nil {
		t.Fatalf("ParseMPLStyle: %v", err)
	}

	fig := NewFigure(100, 100, style.WithTheme(theme))
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})
	if ax.autolimitMode != "round_numbers" {
		t.Fatalf("autolimitMode = %q, want round_numbers", ax.autolimitMode)
	}
}

func TestMPLStyleUnicodeMinusReachesTickLabels(t *testing.T) {
	t.Cleanup(style.ResetDefaults)

	if got := formatScalarTickLabel(ScalarFormatter{}, -1, 1); got != "−1" {
		t.Fatalf("default negative tick label = %q, want unicode minus", got)
	}

	if _, err := style.UpdateParams(style.Params{"axes.unicode_minus": "False"}); err != nil {
		t.Fatalf("UpdateParams: %v", err)
	}
	if got := formatScalarTickLabel(ScalarFormatter{}, -1, 1); got != "-1" {
		t.Fatalf("negative tick label = %q, want ASCII hyphen with axes.unicode_minus: False", got)
	}
}

func TestMPLStyleLinesStyleAndMarkersReachPlot(t *testing.T) {
	src := "lines.linestyle: --\nlines.marker: o\nlines.markersize: 10\nlines.markeredgewidth: 2.5\n"
	theme, report, err := style.ParseMPLStyle("test", src)
	if err != nil {
		t.Fatalf("ParseMPLStyle: %v", err)
	}
	if len(report.Unsupported) > 0 {
		t.Fatalf("unsupported entries: %+v", report.Unsupported)
	}

	fig := NewFigure(100, 100, style.WithTheme(theme))
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})
	line := ax.Plot([]float64{0, 1}, []float64{0, 1})
	if line == nil {
		t.Fatal("Plot returned nil")
	}
	if len(line.Dashes) == 0 {
		t.Fatal("lines.linestyle: -- did not produce a dash pattern")
	}
	if !line.MarkerSet || line.Marker != MarkerCircle {
		t.Fatalf("marker = %v (set=%v), want MarkerCircle from lines.marker", line.Marker, line.MarkerSet)
	}
	if line.MarkerSize != 10 {
		t.Fatalf("marker size = %v, want 10 from lines.markersize", line.MarkerSize)
	}
	if line.MarkerEdgeWidth != 2.5 {
		t.Fatalf("marker edge width = %v, want 2.5 from lines.markeredgewidth", line.MarkerEdgeWidth)
	}

	// Explicit options still win over the rc values.
	solid := LineStyle("-")
	marker := MarkerTriangleUp
	size := 4.0
	line2 := ax.Plot([]float64{0, 1}, []float64{1, 0}, PlotOptions{
		LineStyle:  solid,
		Marker:     &marker,
		MarkerSize: &size,
	})
	if len(line2.Dashes) != 0 {
		t.Fatalf("explicit solid linestyle still produced dashes %v", line2.Dashes)
	}
	if line2.Marker != MarkerTriangleUp || line2.MarkerSize != 4 {
		t.Fatalf("explicit marker options lost: %v/%v", line2.Marker, line2.MarkerSize)
	}
}

func TestMPLStyleLinesStyleNoneSuppressesLine(t *testing.T) {
	theme, _, err := style.ParseMPLStyle("test", "lines.linestyle: none\nlines.marker: o\n")
	if err != nil {
		t.Fatalf("ParseMPLStyle: %v", err)
	}

	fig := NewFigure(100, 100, style.WithTheme(theme))
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})
	line := ax.Plot([]float64{0, 1}, []float64{0, 1})
	if line == nil {
		t.Fatal("Plot returned nil")
	}
	if line.W != 0 {
		t.Fatalf("line width = %v, want 0 (markers-only) from lines.linestyle: none", line.W)
	}
	if !line.MarkerSet {
		t.Fatal("marker not set despite lines.marker: o")
	}
}

func TestMPLStyleScatterDefaultsReachScatter(t *testing.T) {
	src := "scatter.marker: s\nscatter.edgecolors: none\nlines.markersize: 10\n"
	theme, report, err := style.ParseMPLStyle("test", src)
	if err != nil {
		t.Fatalf("ParseMPLStyle: %v", err)
	}
	if len(report.Unsupported) > 0 {
		t.Fatalf("unsupported entries: %+v", report.Unsupported)
	}

	fig := NewFigure(100, 100, style.WithTheme(theme))
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})
	sc := ax.Scatter([]float64{0, 1}, []float64{0, 1})
	if sc == nil {
		t.Fatal("Scatter returned nil")
	}
	if sc.Marker != MarkerSquare {
		t.Fatalf("marker = %v, want MarkerSquare from scatter.marker", sc.Marker)
	}
	if sc.EdgeColor.A != 0 {
		t.Fatalf("edge color = %+v, want transparent from scatter.edgecolors: none", sc.EdgeColor)
	}
	// matplotlib's scatter default size is lines.markersize^2.
	if sc.Size != 100 {
		t.Fatalf("size = %v, want 100 (lines.markersize^2)", sc.Size)
	}

	// Explicit options still win.
	marker := MarkerCircle
	edge := render.Color{R: 1, A: 1}
	size := 25.0
	sc2 := ax.Scatter([]float64{0, 1}, []float64{1, 0}, ScatterOptions{
		Marker: &marker, EdgeColor: &edge, Size: &size,
	})
	if sc2.Marker != MarkerCircle || sc2.EdgeColor != edge || sc2.Size != 25 {
		t.Fatalf("explicit scatter options lost: %v/%+v/%v", sc2.Marker, sc2.EdgeColor, sc2.Size)
	}
}

func TestMPLStyleErrorbarCapsizeReachesErrorBar(t *testing.T) {
	theme, _, err := style.ParseMPLStyle("test", "errorbar.capsize: 3\n")
	if err != nil {
		t.Fatalf("ParseMPLStyle: %v", err)
	}

	fig := NewFigure(100, 100, style.WithTheme(theme))
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})
	eb := ax.ErrorBar([]float64{0, 1}, []float64{0, 1}, nil, []float64{0.1, 0.1})
	if eb == nil {
		t.Fatal("ErrorBar returned nil")
	}
	// 3 pt caps at DPI 100: 2*3 pt * (100/72) px/pt.
	want := 6.0 * 100 / 72
	if !floatApprox(eb.CapSize, want, 1e-9) {
		t.Fatalf("cap size = %v px, want %v from errorbar.capsize", eb.CapSize, want)
	}

	// Default rc keeps caps off.
	fig2 := NewFigure(100, 100)
	ax2 := fig2.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})
	eb2 := ax2.ErrorBar([]float64{0, 1}, []float64{0, 1}, nil, []float64{0.1, 0.1})
	if eb2.CapSize != 0 {
		t.Fatalf("default cap size = %v, want 0", eb2.CapSize)
	}
}
