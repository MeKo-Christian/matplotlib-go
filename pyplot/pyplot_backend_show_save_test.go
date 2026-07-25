package pyplot

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cwbudde/matplotlib-go/backends"
	"github.com/cwbudde/matplotlib-go/canvas"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/widgets"
)

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

func TestSwitchBackendClearsCachedManagersAndUsesNamedBackend(t *testing.T) {
	resetForTests()

	fig := Figure()
	oldCloseCalls := 0
	SetManagerFactory(func(got *core.Figure) (canvas.FigureManager, error) {
		return &testFigureManager{
			canvas:  &testFigureCanvas{figure: got},
			onClose: func() { oldCloseCalls++ },
			tools:   canvas.NewToolManager(),
		}, nil
	})
	if _, err := CurrentFigManager(); err != nil {
		t.Fatalf("CurrentFigManager before switch: %v", err)
	}

	newManagerCalls := 0
	backends.Register(backends.Backend("pyplot-switch-test"), &backends.BackendInfo{
		Name:      "Pyplot Switch Test",
		Available: true,
		Capabilities: []backends.Capability{
			backends.TextShaping,
			backends.FontHinting,
		},
		ManagerFactory: func(_ backends.Config, got *core.Figure) (canvas.FigureManager, error) {
			newManagerCalls++
			if got != fig {
				t.Fatalf("backend manager figure = %p, want %p", got, fig)
			}
			return &testFigureManager{
				canvas: &testFigureCanvas{figure: got},
				tools:  canvas.NewToolManager(),
			}, nil
		},
	})

	if err := SwitchBackend("pyplot-switch-test"); err != nil {
		t.Fatalf("SwitchBackend() error = %v", err)
	}
	if oldCloseCalls != 1 {
		t.Fatalf("old manager close calls = %d, want 1", oldCloseCalls)
	}
	if _, err := CurrentFigManager(); err != nil {
		t.Fatalf("CurrentFigManager after switch: %v", err)
	}
	if newManagerCalls != 1 {
		t.Fatalf("new backend manager calls = %d, want 1", newManagerCalls)
	}
}

func TestManagerEventWrappersUseCurrentFigureManager(t *testing.T) {
	resetForTests()

	fig := Figure()
	testCanvas := &testFigureCanvas{figure: fig}
	drawCalls := 0
	testCanvas.onDraw = func() { drawCalls++ }
	manager := &testFigureManager{
		canvas: testCanvas,
		tools:  canvas.NewToolManager(),
	}
	SetManagerFactory(func(got *core.Figure) (canvas.FigureManager, error) {
		if got != fig {
			t.Fatalf("factory figure = %p, want %p", got, fig)
		}
		return manager, nil
	})

	gotManager, err := CurrentFigManager()
	if err != nil {
		t.Fatalf("CurrentFigManager() error = %v", err)
	}
	if gotManager != manager {
		t.Fatalf("CurrentFigManager() = %p, want %p", gotManager, manager)
	}

	received := 0
	id, err := Connect(canvas.EventDraw, func(canvas.Event) error {
		received++
		return nil
	})
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	if id == 0 {
		t.Fatal("Connect() returned zero id")
	}
	if err := testCanvas.dispatch(canvas.Event{Type: canvas.EventDraw}); err != nil {
		t.Fatalf("dispatch before disconnect: %v", err)
	}
	if received != 1 {
		t.Fatalf("received events before disconnect = %d, want 1", received)
	}
	if err := Disconnect(id); err != nil {
		t.Fatalf("Disconnect() error = %v", err)
	}
	if err := testCanvas.dispatch(canvas.Event{Type: canvas.EventDraw}); err != nil {
		t.Fatalf("dispatch after disconnect: %v", err)
	}
	if received != 1 {
		t.Fatalf("received events after disconnect = %d, want still 1", received)
	}

	if err := DrawIfInteractive(); err != nil {
		t.Fatalf("DrawIfInteractive() while off error = %v", err)
	}
	if drawCalls != 0 {
		t.Fatalf("draw calls while interactive off = %d, want 0", drawCalls)
	}
	restore := Ion()
	if err := DrawIfInteractive(); err != nil {
		t.Fatalf("DrawIfInteractive() while on error = %v", err)
	}
	restore()
	if drawCalls != 1 {
		t.Fatalf("draw calls while interactive on = %d, want 1", drawCalls)
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
	button := widgets.NewButton(ax, "Run")
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

func TestInteractiveModeTogglesAndRestores(t *testing.T) {
	resetForTests()

	if IsInteractive() {
		t.Fatal("interactive mode should be disabled by default")
	}

	restoreOff := Ion()
	if !IsInteractive() {
		t.Fatal("Ion should enable interactive mode")
	}

	restoreOn := Ioff()
	if IsInteractive() {
		t.Fatal("Ioff should disable interactive mode")
	}
	restoreOn()
	if !IsInteractive() {
		t.Fatal("Ioff restore should restore previous enabled state")
	}
	restoreOff()
	if IsInteractive() {
		t.Fatal("Ion restore should restore previous disabled state")
	}
}
