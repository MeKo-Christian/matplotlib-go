package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
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
