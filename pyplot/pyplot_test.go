package pyplot

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cwbudde/matplotlib-go/canvas"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/transform"
)

func TestFigureRegistryTracksCurrentFigureAndAxes(t *testing.T) {
	resetForTests()

	fig1 := Figure()
	if got := GCF(); got != fig1 {
		t.Fatalf("GCF() = %p, want %p", got, fig1)
	}

	ax1 := GCA()
	if ax1 == nil {
		t.Fatal("GCA() returned nil")
	}
	if len(fig1.Children) != 1 {
		t.Fatalf("len(fig1.Children) = %d, want 1", len(fig1.Children))
	}

	fig2 := FigureSized(900, 700)
	if got := GCF(); got != fig2 {
		t.Fatalf("after FigureSized, GCF() = %p, want %p", got, fig2)
	}
	if fig2.SizePx.X != 900 || fig2.SizePx.Y != 700 {
		t.Fatalf("FigureSized dimensions = %.0fx%.0f, want 900x700", fig2.SizePx.X, fig2.SizePx.Y)
	}
}

func TestSubplotReusesAxesForSameSlot(t *testing.T) {
	resetForTests()

	fig := Figure()
	ax1 := Subplot(2, 2, 3)
	ax2 := Subplot(2, 2, 3)
	if ax1 == nil || ax2 == nil {
		t.Fatal("Subplot returned nil axes")
	}
	if ax1 != ax2 {
		t.Fatalf("Subplot did not reuse axes: %p != %p", ax1, ax2)
	}
	if got := len(fig.Children); got != 1 {
		t.Fatalf("len(fig.Children) = %d, want 1", got)
	}
	if got := GCA(); got != ax1 {
		t.Fatalf("GCA() = %p, want %p", got, ax1)
	}
}

func TestSubplotsCreatesNewFigureAndCurrentAxes(t *testing.T) {
	resetForTests()

	fig, grid := Subplots(2, 2, core.WithSubplotShareX())
	if fig == nil {
		t.Fatal("Subplots returned nil figure")
	}
	if len(grid) != 2 || len(grid[0]) != 2 {
		t.Fatalf("Subplots grid dimensions = %dx%d, want 2x2", len(grid), len(grid[0]))
	}
	if got := GCF(); got != fig {
		t.Fatalf("GCF() = %p, want %p", got, fig)
	}
	if got := GCA(); got != grid[0][0] {
		t.Fatalf("GCA() = %p, want %p", got, grid[0][0])
	}
}

func TestStatefulHelpersDelegateToCurrentAxes(t *testing.T) {
	resetForTests()

	Plot([]float64{0, 1, 2}, []float64{1, 2, 3}, core.PlotOptions{Label: "line"})
	Title("Demo")
	XLabel("time")
	YLabel("value")
	legend := Legend()
	if legend == nil {
		t.Fatal("Legend() returned nil")
	}

	ax := GCA()
	if ax.Title != "Demo" {
		t.Fatalf("ax.Title = %q, want %q", ax.Title, "Demo")
	}
	if ax.XLabel != "time" || ax.YLabel != "value" {
		t.Fatalf("axis labels = (%q, %q), want (%q, %q)", ax.XLabel, ax.YLabel, "time", "value")
	}
	if len(ax.Artists) != 2 {
		t.Fatalf("len(ax.Artists) = %d, want 2", len(ax.Artists))
	}
}

func TestTextAndAnnotateDelegateToCurrentAxes(t *testing.T) {
	resetForTests()

	text := Text(0.2, 0.8, "note", core.TextOptions{
		FontSize: 14,
		HAlign:   core.TextAlignCenter,
	})
	annotation := Annotate("peak", 0.7, 0.3, core.AnnotationOptions{
		OffsetX: 10,
		OffsetY: -12,
	})

	if text == nil {
		t.Fatal("Text() returned nil")
	}
	if annotation == nil {
		t.Fatal("Annotate() returned nil")
	}
	ax := GCA()
	if len(ax.Artists) != 2 {
		t.Fatalf("len(ax.Artists) = %d, want 2", len(ax.Artists))
	}
	if text.Content != "note" || text.Position != (geom.Pt{X: 0.2, Y: 0.8}) {
		t.Fatalf("Text() artist = %+v, want delegated core text", text)
	}
	if annotation.Content != "peak" || annotation.Point != (geom.Pt{X: 0.7, Y: 0.3}) {
		t.Fatalf("Annotate() artist = %+v, want delegated core annotation", annotation)
	}
}

func TestReferenceLineAndSpanHelpersDelegateToCurrentAxes(t *testing.T) {
	resetForTests()

	hLine := AxHLine(0.25)
	vLine := AxVLine(0.75)
	line := AxLine(geom.Pt{X: 0, Y: 0}, geom.Pt{X: 1, Y: 1})
	slopeLine := AxLineSlope(geom.Pt{X: 0.5, Y: 0.5}, 2)
	hSpan := AxHSpan(0.1, 0.2)
	vSpan := AxVSpan(0.3, 0.4)

	if hLine == nil || vLine == nil || line == nil || slopeLine == nil || hSpan == nil || vSpan == nil {
		t.Fatalf("reference helpers returned nil: h=%v v=%v line=%v slope=%v hspan=%v vspan=%v", hLine, vLine, line, slopeLine, hSpan, vSpan)
	}
	ax := GCA()
	if len(ax.Artists) != 6 {
		t.Fatalf("len(ax.Artists) = %d, want 6", len(ax.Artists))
	}
	if hLine.Start.Y != 0.25 || vLine.Start.X != 0.75 {
		t.Fatalf("line endpoints not delegated through core helpers: h=%+v v=%+v", hLine, vLine)
	}
	if line.Direction != (geom.Pt{X: 1, Y: 1}) || slopeLine.Direction != (geom.Pt{X: 1, Y: 2}) {
		t.Fatalf("axline directions = %+v / %+v, want (1,1) / (1,2)", line.Direction, slopeLine.Direction)
	}
	if hSpan.Start.Y != 0.1 || hSpan.End.Y != 0.2 || vSpan.Start.X != 0.3 || vSpan.End.X != 0.4 {
		t.Fatalf("span endpoints not delegated through core helpers: h=%+v v=%+v", hSpan, vSpan)
	}
}

func TestAxisLimitAndScaleHelpersDelegateToCurrentAxes(t *testing.T) {
	resetForTests()

	XLim(1, 100)
	YLim(-9, 9)
	if err := XScale("log", transform.WithScaleBase(10)); err != nil {
		t.Fatalf("XScale(log): %v", err)
	}
	if err := YScale("symlog", transform.WithScaleLinThresh(2)); err != nil {
		t.Fatalf("YScale(symlog): %v", err)
	}

	ax := GCA()
	if xMin, xMax := ax.XScale.Domain(); xMin != 1 || xMax != 100 {
		t.Fatalf("x domain = (%v, %v), want (1, 100)", xMin, xMax)
	}
	if yMin, yMax := ax.YScale.Domain(); yMin != -9 || yMax != 9 {
		t.Fatalf("y domain = (%v, %v), want (-9, 9)", yMin, yMax)
	}
	if _, ok := ax.XScale.(transform.Log); !ok {
		t.Fatalf("x scale = %T, want transform.Log", ax.XScale)
	}
	if _, ok := ax.YScale.(transform.SymLog); !ok {
		t.Fatalf("y scale = %T, want transform.SymLog", ax.YScale)
	}
}

func TestConveniencePlotHelpersDelegateToCurrentAxes(t *testing.T) {
	resetForTests()

	dates := []time.Time{
		time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC),
	}
	if line := PlotDate(dates, []float64{1, 2}); line == nil {
		t.Fatal("PlotDate() returned nil")
	}
	if _, ok := GCA().XAxis.Locator.(core.DateLocator); !ok {
		t.Fatalf("PlotDate x-axis locator = %T, want DateLocator", GCA().XAxis.Locator)
	}

	resetForTests()
	if line := SemilogX([]float64{1, 10}, []float64{1, 2}); line == nil {
		t.Fatal("SemilogX() returned nil")
	}
	if _, ok := GCA().XScale.(transform.Log); !ok {
		t.Fatalf("SemilogX x scale = %T, want transform.Log", GCA().XScale)
	}

	resetForTests()
	if line := SemilogY([]float64{1, 2}, []float64{1, 10}); line == nil {
		t.Fatal("SemilogY() returned nil")
	}
	if _, ok := GCA().YScale.(transform.Log); !ok {
		t.Fatalf("SemilogY y scale = %T, want transform.Log", GCA().YScale)
	}

	resetForTests()
	if line := LogLog([]float64{1, 10}, []float64{1, 10}); line == nil {
		t.Fatal("LogLog() returned nil")
	}
	if _, ok := GCA().XScale.(transform.Log); !ok {
		t.Fatalf("LogLog x scale = %T, want transform.Log", GCA().XScale)
	}
	if _, ok := GCA().YScale.(transform.Log); !ok {
		t.Fatalf("LogLog y scale = %T, want transform.Log", GCA().YScale)
	}

	resetForTests()
	if bar := BarH([]float64{0, 1}, []float64{3, 4}); bar == nil || bar.Orientation != core.BarHorizontal {
		t.Fatalf("BarH() = %#v, want horizontal bar", bar)
	}
	if fill := Fill([]float64{0, 1, 0}, []float64{0, 0, 1}); fill == nil {
		t.Fatal("Fill() returned nil")
	}
	pie := Pie([]float64{1, 2}, core.PieOptions{Labels: []string{"A", "B"}})
	if pie == nil {
		t.Fatal("Pie() returned nil")
	}
	if labels := PieLabel(pie, []string{"one", "two"}); len(labels) != 2 {
		t.Fatalf("PieLabel() returned %d labels, want 2", len(labels))
	}
}

func TestColorbarUsesCurrentAxesAndFigure(t *testing.T) {
	resetForTests()

	img := Image([][]float64{
		{0, 1},
		{2, 3},
	})
	cb := Colorbar(img, core.ColorbarOptions{Label: "Intensity"})
	if cb == nil {
		t.Fatal("Colorbar() returned nil")
	}

	fig := GCF()
	if len(fig.Children) != 2 {
		t.Fatalf("len(fig.Children) = %d, want 2", len(fig.Children))
	}
	if cb.YLabel != "Intensity" {
		t.Fatalf("colorbar label = %q, want %q", cb.YLabel, "Intensity")
	}
}

func TestVectorFieldHelpersDelegateToCurrentAxes(t *testing.T) {
	resetForTests()

	q := Quiver(
		[]float64{0, 1},
		[]float64{0, 1},
		[]float64{1, 0.5},
		[]float64{0.25, 0.75},
		core.QuiverOptions{Label: "q"},
	)
	if q == nil {
		t.Fatal("Quiver() returned nil")
	}
	key := QuiverKey(q, 0.8, 0.2, 1, "1 unit")
	if key == nil {
		t.Fatal("QuiverKey() returned nil")
	}
	barbs := Barbs(
		[]float64{0.5},
		[]float64{0.5},
		[]float64{12},
		[]float64{3},
	)
	if barbs == nil {
		t.Fatal("Barbs() returned nil")
	}
	stream := Streamplot(
		[]float64{0, 1, 2},
		[]float64{0, 1},
		[][]float64{{1, 1, 1}, {1, 1, 1}},
		[][]float64{{0, 0.2, 0.2}, {0, 0.2, 0.2}},
		core.StreamplotOptions{StartPoints: []geom.Pt{{X: 0.2, Y: 0.4}}},
	)
	if stream == nil {
		t.Fatal("Streamplot() returned nil")
	}

	ax := GCA()
	if len(ax.Artists) != 4 {
		t.Fatalf("len(ax.Artists) = %d, want 4", len(ax.Artists))
	}
}

func TestMatrixAndSignalHelpersDelegateToCurrentAxes(t *testing.T) {
	resetForTests()

	mat := [][]float64{
		{0, 1},
		{2, 3},
	}
	bilinear := "bilinear"
	if img := ImShow(mat, core.ImShowOptions{Interpolation: &bilinear}); img == nil || img.Interpolation != "bilinear" {
		t.Fatalf("ImShow() = %#v, want image with bilinear interpolation", img)
	}
	if img := MatShow(mat); img == nil {
		t.Fatal("MatShow() returned nil")
	}
	if spy := Spy(mat); spy == nil {
		t.Fatal("Spy() returned nil")
	}
	if spec := Specgram([]float64{0, 1, 0, -1, 0, 1, 0, -1}, core.SpecgramOptions{NFFT: 4}); spec == nil {
		t.Fatal("Specgram() returned nil")
	}
	if psd := PSD([]float64{0, 1, 0, -1, 0, 1, 0, -1}, core.SignalSpectrumOptions{NFFT: 4}); psd == nil {
		t.Fatal("PSD() returned nil")
	}
	if mag := MagnitudeSpectrum([]float64{0, 1, 0, -1, 0, 1, 0, -1}, core.SignalSpectrumOptions{NFFT: 4}); mag == nil {
		t.Fatal("MagnitudeSpectrum() returned nil")
	}
	if angle := AngleSpectrum([]float64{0, 1, 0, -1, 0, 1, 0, -1}, core.SignalSpectrumOptions{NFFT: 4}); angle == nil {
		t.Fatal("AngleSpectrum() returned nil")
	}
	if phase := PhaseSpectrum([]float64{0, 1, 0, -1, 0, 1, 0, -1}, core.SignalSpectrumOptions{NFFT: 4}); phase == nil {
		t.Fatal("PhaseSpectrum() returned nil")
	}
	if csd := CSD([]float64{0, 1, 0, -1}, []float64{0, 1, 0, -1}, core.SignalSpectrumOptions{NFFT: 4}); csd == nil {
		t.Fatal("CSD() returned nil")
	}
	if cohere := Cohere([]float64{0, 1, 0, -1}, []float64{0, 1, 0, -1}, core.SignalSpectrumOptions{NFFT: 4}); cohere == nil {
		t.Fatal("Cohere() returned nil")
	}
	if xcorr := XCorr([]float64{1, 2, 3}, []float64{1, 2, 3}); xcorr == nil {
		t.Fatal("XCorr() returned nil")
	}
	if acorr := ACorr([]float64{1, 2, 3}); acorr == nil {
		t.Fatal("ACorr() returned nil")
	}
	if heatmap := AnnotatedHeatmap(mat); heatmap == nil {
		t.Fatal("AnnotatedHeatmap() returned nil")
	}
}

func TestThreeDHelpersDelegateToCurrent3DAxes(t *testing.T) {
	resetForTests()

	ax := AddAxes3D(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.14},
		Max: geom.Pt{X: 0.88, Y: 0.88},
	})
	if ax == nil {
		t.Fatal("AddAxes3D() returned nil")
	}

	if line := Plot3D([]float64{0, 1}, []float64{0, 1}, []float64{0, 1}); line == nil {
		t.Fatal("Plot3D() returned nil")
	}
	if scatter := Scatter3D([]float64{0, 1}, []float64{0, 1}, []float64{1, 2}); scatter == nil {
		t.Fatal("Scatter3D() returned nil")
	}
	if wire := Wireframe([]float64{0, 1}, []float64{0, 1}, [][]float64{{0, 1}, {1, 2}}); wire == nil {
		t.Fatal("Wireframe() returned nil")
	}
	if surf := Surface([]float64{0, 1}, []float64{0, 1}, [][]float64{{0, 1}, {1, 2}}); surf == nil {
		t.Fatal("Surface() returned nil")
	}
	if voxel := Voxel([]float64{0, 1}, []float64{0, 1}, []float64{0, 1}, []float64{1, 1}, []float64{1, 1}, []float64{1, 1}); voxel == nil {
		t.Fatal("Voxel() returned nil")
	}
	tri := core.Triangulation{
		X:         []float64{0, 1, 1, 0},
		Y:         []float64{0, 0, 1, 1},
		Triangles: [][3]int{{0, 1, 2}, {0, 2, 3}},
	}
	if tris := Trisurf(tri, []float64{0, 1, 2, 3}); tris == nil {
		t.Fatal("Trisurf() returned nil")
	}
	if contour := Contour3D([]float64{0, 1}, []float64{0, 1}, [][]float64{{0, 1}, {1, 2}}); contour == nil {
		t.Fatal("Contour3D() returned nil")
	}
	if contourf := Contourf3D([]float64{0, 1}, []float64{0, 1}, [][]float64{{0, 1}, {1, 2}}); contourf == nil {
		t.Fatal("Contourf3D() returned nil")
	}
	if text := Text3D(0.2, 0.5, 0.9, "pt"); text == nil {
		t.Fatal("Text3D() returned nil")
	}
}

func TestParasiteAxesWrapperCreatesOverlayAxesAndSetsCurrent(t *testing.T) {
	resetForTests()

	fig := GCF()
	host := GCA()
	if fig == nil || host == nil {
		t.Fatal("initial figure/axes not available")
	}

	parasite := ParasiteAxes()
	if parasite == nil || parasite.Axes == nil {
		t.Fatal("ParasiteAxes() returned nil")
	}
	if got := GCA(); got != parasite.Axes {
		t.Fatalf("GCA() = %p, want %p", got, parasite.Axes)
	}
	if got := len(fig.Children); got != 2 {
		t.Fatalf("len(fig.Children) = %d, want 2", got)
	}
	if parasite.Host != host {
		t.Fatalf("parasite host = %p, want %p", parasite.Host, host)
	}
}

func TestSavefigWritesPNGAndSVG(t *testing.T) {
	resetForTests()
	t.Setenv("MATPLOTLIB_BACKEND", "gobasic")

	Plot([]float64{0, 1, 2}, []float64{2, 1, 3})
	Title("Savefig")

	dir := t.TempDir()
	pngPath := filepath.Join(dir, "plot.png")
	if err := Savefig(pngPath); err != nil {
		t.Fatalf("Savefig(%q) failed: %v", pngPath, err)
	}
	if info, err := os.Stat(pngPath); err != nil || info.Size() == 0 {
		t.Fatalf("PNG output missing or empty: info=%v err=%v", info, err)
	}

	svgPath := filepath.Join(dir, "plot.svg")
	if err := Savefig(svgPath); err != nil {
		t.Fatalf("Savefig(%q) failed: %v", svgPath, err)
	}
	data, err := os.ReadFile(svgPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) failed: %v", svgPath, err)
	}
	if !strings.Contains(string(data), "<svg") {
		t.Fatalf("SVG output does not start with <svg: %q", string(data))
	}
}

func TestShowAndPauseUseConfiguredHandler(t *testing.T) {
	resetForTests()

	fig1 := Figure()
	fig2 := Figure()
	var shown []*core.Figure

	SetShowHandler(func(fig *core.Figure) error {
		shown = append(shown, fig)
		return nil
	})

	if err := Show(); err != nil {
		t.Fatalf("Show() failed: %v", err)
	}
	if len(shown) != 2 || shown[0] != fig1 || shown[1] != fig2 {
		t.Fatalf("Show() figures = %v, want [%p %p]", shown, fig1, fig2)
	}

	shown = shown[:0]
	if err := Pause(5 * time.Millisecond); err != nil {
		t.Fatalf("Pause() failed: %v", err)
	}
	if len(shown) != 2 {
		t.Fatalf("Pause() show count = %d, want 2", len(shown))
	}
}

func TestSetManagerFactoryCachesManagerPerFigure(t *testing.T) {
	resetForTests()

	fig := Figure()
	factoryCalls := 0
	showCalls := 0

	SetManagerFactory(func(got *core.Figure) (canvas.FigureManager, error) {
		factoryCalls++
		if got != fig {
			t.Fatalf("factory figure = %p, want %p", got, fig)
		}
		return &testFigureManager{
			canvas: &testFigureCanvas{figure: got},
			onShow: func() { showCalls++ },
			tools:  canvas.NewToolManager(),
		}, nil
	})

	if err := Show(); err != nil {
		t.Fatalf("Show() error = %v", err)
	}
	if err := Show(); err != nil {
		t.Fatalf("Show() second call error = %v", err)
	}

	if factoryCalls != 1 {
		t.Fatalf("factory calls = %d, want 1", factoryCalls)
	}
	if showCalls != 2 {
		t.Fatalf("show calls = %d, want 2", showCalls)
	}
}

func TestCloseRemovesFiguresAndClosesManagers(t *testing.T) {
	resetForTests()

	fig1 := Figure()
	fig2 := Figure()
	showCalls := map[*core.Figure]int{}
	closeCalls := map[*core.Figure]int{}

	SetManagerFactory(func(fig *core.Figure) (canvas.FigureManager, error) {
		return &testFigureManager{
			canvas: &testFigureCanvas{figure: fig},
			onShow: func() { showCalls[fig]++ },
			onClose: func() {
				closeCalls[fig]++
			},
			tools: canvas.NewToolManager(),
		}, nil
	})

	if err := Show(); err != nil {
		t.Fatalf("Show() error = %v", err)
	}
	if err := Close(fig1); err != nil {
		t.Fatalf("Close(fig1) error = %v", err)
	}
	if closeCalls[fig1] != 1 {
		t.Fatalf("fig1 close calls = %d, want 1", closeCalls[fig1])
	}
	if GCF() != fig2 {
		t.Fatal("closing fig1 should leave fig2 current")
	}

	showCalls = map[*core.Figure]int{}
	if err := Show(); err != nil {
		t.Fatalf("Show() after Close(fig1) error = %v", err)
	}
	if showCalls[fig1] != 0 || showCalls[fig2] != 1 {
		t.Fatalf("show calls after Close(fig1) = fig1:%d fig2:%d, want 0/1", showCalls[fig1], showCalls[fig2])
	}

	if err := Close(); err != nil {
		t.Fatalf("Close() current error = %v", err)
	}
	if closeCalls[fig2] != 1 {
		t.Fatalf("fig2 close calls = %d, want 1", closeCalls[fig2])
	}

	registry.mu.Lock()
	remaining := len(registry.figures)
	current := registry.current
	registry.mu.Unlock()
	if remaining != 0 || current != nil {
		t.Fatalf("registry after Close() = remaining %d current %p, want empty/nil", remaining, current)
	}
}

func TestCloseAllRemovesEveryFigure(t *testing.T) {
	resetForTests()

	fig1 := Figure()
	fig2 := Figure()
	closeCalls := map[*core.Figure]int{}
	SetManagerFactory(func(fig *core.Figure) (canvas.FigureManager, error) {
		return &testFigureManager{
			canvas: &testFigureCanvas{figure: fig},
			onClose: func() {
				closeCalls[fig]++
			},
			tools: canvas.NewToolManager(),
		}, nil
	})
	if err := Show(); err != nil {
		t.Fatalf("Show() error = %v", err)
	}

	if err := CloseAll(); err != nil {
		t.Fatalf("CloseAll() error = %v", err)
	}
	if closeCalls[fig1] != 1 || closeCalls[fig2] != 1 {
		t.Fatalf("close calls = fig1:%d fig2:%d, want 1/1", closeCalls[fig1], closeCalls[fig2])
	}

	registry.mu.Lock()
	remaining := len(registry.figures)
	managers := len(registry.managers)
	current := registry.current
	registry.mu.Unlock()
	if remaining != 0 || managers != 0 || current != nil {
		t.Fatalf("registry after CloseAll = figures:%d managers:%d current:%p, want empty", remaining, managers, current)
	}
}

func TestCLFClearsCurrentFigureAndAxesRegistry(t *testing.T) {
	resetForTests()

	fig := Figure()
	ax := GCA()
	Plot([]float64{0, 1}, []float64{0, 1})
	fig.Add(&core.Text{Position: geom.Pt{X: 0.5, Y: 0.5}, Content: "figure note"})

	CLF()

	if GCF() != fig {
		t.Fatal("CLF should keep the same current figure")
	}
	if len(fig.Children) != 0 || len(fig.Artists) != 0 {
		t.Fatalf("figure after CLF = %d axes, %d artists; want empty", len(fig.Children), len(fig.Artists))
	}
	registry.mu.Lock()
	currentAxes := registry.currentAxes[fig]
	subplots := len(registry.subplotAxes[fig])
	registry.mu.Unlock()
	if currentAxes != nil || subplots != 0 {
		t.Fatalf("registry after CLF = current axes %p subplots %d, want nil/0", currentAxes, subplots)
	}

	next := GCA()
	if next == nil || next == ax {
		t.Fatalf("GCA after CLF = %p, want new axes", next)
	}
	if len(fig.Children) != 1 || fig.Children[0] != next {
		t.Fatalf("figure axes after new GCA = %d, want one new current axes", len(fig.Children))
	}
}

func TestCLAClearsCurrentAxesButKeepsItCurrent(t *testing.T) {
	resetForTests()

	ax := GCA()
	Plot([]float64{0, 1}, []float64{0, 1})
	button := ax.Button("Run")
	if button == nil {
		t.Fatal("button constructor returned nil")
	}
	Title("old title")
	XLabel("old x")
	YLabel("old y")
	XLim(2, 3)
	YLim(4, 5)

	CLA()

	if GCA() != ax {
		t.Fatal("CLA should keep the same current axes")
	}
	if len(ax.Artists) != 0 || len(ax.WidgetArtists) != 0 {
		t.Fatalf("axes after CLA = %d artists, %d widgets; want empty", len(ax.Artists), len(ax.WidgetArtists))
	}
	if ax.Title != "" || ax.XLabel != "" || ax.YLabel != "" {
		t.Fatalf("labels after CLA = %q/%q/%q, want empty", ax.Title, ax.XLabel, ax.YLabel)
	}
	x0, x1 := ax.XScale.Domain()
	y0, y1 := ax.YScale.Domain()
	if x0 != 0 || x1 != 1 || y0 != 0 || y1 != 1 {
		t.Fatalf("limits after CLA = x[%v,%v] y[%v,%v], want [0,1]/[0,1]", x0, x1, y0, y1)
	}
}

func TestDrawUsesCurrentFigureManagerCanvas(t *testing.T) {
	resetForTests()

	fig := Figure()
	drawCalls := 0
	SetManagerFactory(func(got *core.Figure) (canvas.FigureManager, error) {
		if got != fig {
			t.Fatalf("manager figure = %p, want %p", got, fig)
		}
		return &testFigureManager{
			canvas: &testFigureCanvas{
				figure: got,
				onDraw: func() {
					drawCalls++
				},
			},
			tools: canvas.NewToolManager(),
		}, nil
	})

	if err := Draw(); err != nil {
		t.Fatalf("Draw() error = %v", err)
	}
	if drawCalls != 1 {
		t.Fatalf("draw calls = %d, want 1", drawCalls)
	}
}

func TestRCUpdatesActiveDefaultsForNewFigures(t *testing.T) {
	resetForTests()

	if err := RC("figure", style.Params{"dpi": "144"}); err != nil {
		t.Fatalf("RC() error = %v", err)
	}
	if err := RC("axes", style.Params{"facecolor": "#ddeeff"}); err != nil {
		t.Fatalf("RC() error = %v", err)
	}

	fig := Figure()
	if fig.RC.DPI != 144 {
		t.Fatalf("figure DPI = %v, want 144", fig.RC.DPI)
	}
	if got := fig.RC.AxesBackground; got.R != 0xdd/255.0 || got.G != 0xee/255.0 || got.B != 0xff/255.0 {
		t.Fatalf("axes facecolor = %+v", got)
	}
}

func TestRCContextTemporarilyOverridesDefaults(t *testing.T) {
	resetForTests()

	if err := RC("figure", style.Params{"dpi": "120"}); err != nil {
		t.Fatalf("RC() error = %v", err)
	}

	restore, err := RCContext(style.Params{"figure.dpi": "220"})
	if err != nil {
		t.Fatalf("RCContext() error = %v", err)
	}

	if got := Figure().RC.DPI; got != 220 {
		t.Fatalf("context figure DPI = %v, want 220", got)
	}

	restore()
	if got := Figure().RC.DPI; got != 120 {
		t.Fatalf("restored figure DPI = %v, want 120", got)
	}
}

func TestRCDefaultsResetsActiveDefaults(t *testing.T) {
	resetForTests()

	if err := RC("figure", style.Params{"dpi": "144"}); err != nil {
		t.Fatalf("RC() error = %v", err)
	}
	RCDefaults()

	if got := Figure().RC.DPI; got != style.Default.DPI {
		t.Fatalf("figure DPI = %v, want default %v", got, style.Default.DPI)
	}
}

func TestLoadRCFileUpdatesDefaults(t *testing.T) {
	resetForTests()

	path := filepath.Join(t.TempDir(), "matplotlibrc")
	if err := os.WriteFile(path, []byte("figure.dpi: 175\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := LoadRCFile(path); err != nil {
		t.Fatalf("LoadRCFile() error = %v", err)
	}
	if got := Figure().RC.DPI; got != 175 {
		t.Fatalf("figure DPI = %v, want 175", got)
	}
}

func TestFigureUsesConfiguredFigureSize(t *testing.T) {
	resetForTests()

	if err := RC("figure", style.Params{
		"dpi":     "120",
		"figsize": "7.5, 5",
	}); err != nil {
		t.Fatalf("RC() error = %v", err)
	}

	fig := Figure()
	if fig.SizePx.X != 900 || fig.SizePx.Y != 600 {
		t.Fatalf("figure size = %.0fx%.0f, want 900x600", fig.SizePx.X, fig.SizePx.Y)
	}
}

type testFigureManager struct {
	canvas  canvas.FigureCanvas
	onShow  func()
	onClose func()
	tools   *canvas.ToolManager
}

func (m *testFigureManager) Canvas() canvas.FigureCanvas { return m.canvas }

func (m *testFigureManager) Show() error {
	if m.onShow != nil {
		m.onShow()
	}
	return nil
}

func (m *testFigureManager) Close() error {
	if m.onClose != nil {
		m.onClose()
	}
	return nil
}

func (m *testFigureManager) SetTitle(string) {}

func (m *testFigureManager) ToolManager() *canvas.ToolManager { return m.tools }

type testFigureCanvas struct {
	figure *core.Figure
	onDraw func()
}

func (c *testFigureCanvas) Figure() *core.Figure { return c.figure }

func (c *testFigureCanvas) Draw() error {
	if c.onDraw != nil {
		c.onDraw()
	}
	return nil
}

func (c *testFigureCanvas) Resize(width, height int) error {
	if c.figure != nil {
		c.figure.SizePx.X = float64(width)
		c.figure.SizePx.Y = float64(height)
	}
	return nil
}

func (c *testFigureCanvas) Connect(canvas.EventType, canvas.Handler) canvas.ConnectionID { return 0 }

func (c *testFigureCanvas) Disconnect(canvas.ConnectionID) {}

func (c *testFigureCanvas) Close() error { return nil }
