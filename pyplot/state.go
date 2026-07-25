package pyplot

import (
	"errors"
	"sync"

	"github.com/cwbudde/matplotlib-go/canvas"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/transform"
)

// ManagerFactory creates a figure manager for pyplot Show/Pause lifecycle calls.
type ManagerFactory func(*core.Figure) (canvas.FigureManager, error)

type registryState struct {
	mu             sync.Mutex
	current        *core.Figure
	figures        []*core.Figure
	currentAxes    map[*core.Figure]*core.Axes
	subplotAxes    map[*core.Figure]map[string]*core.Axes
	managers       map[*core.Figure]canvas.FigureManager
	managerFactory ManagerFactory
	interactive    bool
}

var registry = registryState{
	currentAxes:    make(map[*core.Figure]*core.Axes),
	subplotAxes:    make(map[*core.Figure]map[string]*core.Axes),
	managers:       make(map[*core.Figure]canvas.FigureManager),
	managerFactory: defaultManagerFactory,
}

// Figure creates a new current figure using Matplotlib-like default dimensions.
func Figure(opts ...style.Option) *core.Figure {
	width, height := style.CurrentDefaults().DefaultFigureSizePx()
	return FigureSized(width, height, opts...)
}

// FigureSized creates a new current figure with explicit pixel dimensions.
func FigureSized(width, height int, opts ...style.Option) *core.Figure {
	if width <= 0 {
		width, _ = style.CurrentDefaults().DefaultFigureSizePx()
	}
	if height <= 0 {
		_, height = style.CurrentDefaults().DefaultFigureSizePx()
	}

	fig := core.NewFigure(width, height, opts...)
	registry.mu.Lock()
	registry.figures = append(registry.figures, fig)
	registry.current = fig
	registry.mu.Unlock()
	return fig
}

// GCF returns the current figure, creating one if necessary.
func GCF() *core.Figure {
	registry.mu.Lock()
	fig := registry.current
	registry.mu.Unlock()
	if fig != nil {
		return fig
	}
	return Figure()
}

// GCA returns the current axes for the current figure, creating a default
// 1x1 subplot if necessary.
func GCA() *core.Axes {
	fig := GCF()

	registry.mu.Lock()
	ax := registry.currentAxes[fig]
	registry.mu.Unlock()
	if ax != nil {
		return ax
	}
	return ensureDefaultAxes(fig)
}

// SCA sets the provided registered axes as current.
func SCA(ax *core.Axes) error {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	fig := registry.figureForAxesLocked(ax)
	if fig == nil {
		return errors.New("pyplot: axes is not registered")
	}
	registry.current = fig
	registry.currentAxes[fig] = ax
	return nil
}

// DelAxes removes the provided axes from its registered figure.
func DelAxes(ax *core.Axes) error {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	fig := registry.figureForAxesLocked(ax)
	if fig == nil {
		return errors.New("pyplot: axes is not registered")
	}

	filtered := fig.Children[:0]
	for _, child := range fig.Children {
		if child != ax {
			filtered = append(filtered, child)
		}
	}
	fig.Children = filtered
	for key, subplot := range registry.subplotAxes[fig] {
		if subplot == ax {
			delete(registry.subplotAxes[fig], key)
		}
	}
	if registry.currentAxes[fig] == ax {
		delete(registry.currentAxes, fig)
		if len(fig.Children) > 0 {
			registry.currentAxes[fig] = fig.Children[len(fig.Children)-1]
		}
	}
	registry.current = fig
	return nil
}

// IsInteractive reports whether pyplot interactive mode is enabled.
func IsInteractive() bool {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	return registry.interactive
}

// Ion enables pyplot interactive mode and returns a restore function for the
// previous state.
func Ion() func() {
	return setInteractive(true)
}

// Ioff disables pyplot interactive mode and returns a restore function for the
// previous state.
func Ioff() func() {
	return setInteractive(false)
}

// SetManagerFactory overrides how pyplot creates managers for Show and Pause.
// Passing nil restores the default backend-driven manager selection.
func SetManagerFactory(factory ManagerFactory) {
	if factory == nil {
		factory = defaultManagerFactory
	}
	closeCachedManagersAndSetFactory(factory)
}

func closeCachedManagersAndSetFactory(factory ManagerFactory) {
	registry.mu.Lock()
	existing := registry.managers
	registry.managers = make(map[*core.Figure]canvas.FigureManager)
	registry.managerFactory = factory
	registry.mu.Unlock()

	for _, manager := range existing {
		if manager != nil {
			_ = manager.Close()
		}
	}
}

// CurrentFigManager returns the cached or newly-created manager for the current figure.
func CurrentFigManager() (canvas.FigureManager, error) {
	return ensureManager(GCF())
}

// Close removes the given figures from pyplot state and closes their cached
// managers. With no arguments, Close closes the current figure.
func Close(figs ...*core.Figure) error {
	registry.mu.Lock()
	if len(figs) == 0 {
		if registry.current == nil {
			registry.mu.Unlock()
			return nil
		}
		figs = []*core.Figure{registry.current}
	}

	targets := make(map[*core.Figure]struct{}, len(figs))
	managers := make([]canvas.FigureManager, 0, len(figs))
	for _, fig := range figs {
		if fig == nil {
			continue
		}
		if _, seen := targets[fig]; seen {
			continue
		}
		targets[fig] = struct{}{}
		if manager := registry.managers[fig]; manager != nil {
			managers = append(managers, manager)
		}
		delete(registry.managers, fig)
		delete(registry.currentAxes, fig)
		delete(registry.subplotAxes, fig)
	}
	if len(targets) > 0 {
		filtered := registry.figures[:0]
		for _, fig := range registry.figures {
			if _, closeFig := targets[fig]; closeFig {
				continue
			}
			filtered = append(filtered, fig)
		}
		registry.figures = filtered
		if _, closedCurrent := targets[registry.current]; closedCurrent {
			registry.current = nil
			if len(registry.figures) > 0 {
				registry.current = registry.figures[len(registry.figures)-1]
			}
		}
	}
	registry.mu.Unlock()

	var err error
	for _, manager := range managers {
		if manager == nil {
			continue
		}
		err = errors.Join(err, manager.Close())
	}
	return err
}

// CloseAll removes every registered pyplot figure and closes cached managers.
func CloseAll() error {
	registry.mu.Lock()
	figures := append([]*core.Figure(nil), registry.figures...)
	registry.mu.Unlock()
	return Close(figures...)
}

// CLF clears the current figure while keeping it registered and current.
func CLF() {
	fig := GCF()
	if fig == nil {
		return
	}

	fig.Children = nil
	fig.Artists = nil
	fig.SupTitle = ""
	fig.SupXLabel = ""
	fig.SupYLabel = ""

	registry.mu.Lock()
	delete(registry.currentAxes, fig)
	delete(registry.subplotAxes, fig)
	registry.current = fig
	registry.mu.Unlock()
}

// CLA clears the current axes while keeping it current.
func CLA() {
	clearAxes(GCA())
}

func clearAxes(ax *core.Axes) {
	if ax == nil {
		return
	}
	ax.Artists = nil
	ax.WidgetArtists = nil
	ax.Title = ""
	ax.XLabel = ""
	ax.YLabel = ""
	ax.XScale = transform.NewLinear(0, 1)
	ax.YScale = transform.NewLinear(0, 1)
	ax.XAxis = core.NewXAxis()
	ax.YAxis = core.NewYAxis()
	ax.XAxisTop = nil
	ax.YAxisRight = nil
	ax.ExtraAxes = nil
	ax.ShowFrame = true
}

func setCurrentAxes(ax *core.Axes) *core.Axes {
	if ax == nil {
		return nil
	}
	registry.mu.Lock()
	if fig := registry.figureForAxesLocked(ax); fig != nil {
		registry.rememberAxesLocked(fig, ax, "")
	}
	registry.mu.Unlock()
	return ax
}

func ensureDefaultAxes(fig *core.Figure) *core.Axes {
	registry.mu.Lock()
	if ax := registry.currentAxes[fig]; ax != nil {
		registry.mu.Unlock()
		return ax
	}
	registry.mu.Unlock()
	return Subplot(1, 1, 1)
}

func (r *registryState) rememberAxesLocked(fig *core.Figure, ax *core.Axes, key string) {
	if fig == nil || ax == nil {
		return
	}
	if r.subplotAxes[fig] == nil {
		r.subplotAxes[fig] = make(map[string]*core.Axes)
	}
	if key != "" {
		r.subplotAxes[fig][key] = ax
	}
	r.current = fig
	r.currentAxes[fig] = ax
}

func (r *registryState) figureForAxesLocked(ax *core.Axes) *core.Figure {
	if ax == nil {
		return nil
	}
	for _, fig := range r.figures {
		for _, child := range fig.Children {
			if child == ax {
				return fig
			}
		}
	}
	return nil
}

func resetForTests() {
	registry.mu.Lock()
	registry.current = nil
	registry.figures = nil
	registry.currentAxes = make(map[*core.Figure]*core.Axes)
	registry.subplotAxes = make(map[*core.Figure]map[string]*core.Axes)
	registry.managers = make(map[*core.Figure]canvas.FigureManager)
	registry.managerFactory = defaultManagerFactory
	registry.interactive = false
	registry.mu.Unlock()
	style.ResetDefaults()
}

func setInteractive(enabled bool) func() {
	registry.mu.Lock()
	previous := registry.interactive
	registry.interactive = enabled
	registry.mu.Unlock()
	return func() {
		registry.mu.Lock()
		registry.interactive = previous
		registry.mu.Unlock()
	}
}

func ensureManager(fig *core.Figure) (canvas.FigureManager, error) {
	registry.mu.Lock()
	if manager := registry.managers[fig]; manager != nil {
		registry.mu.Unlock()
		return manager, nil
	}
	factory := registry.managerFactory
	registry.mu.Unlock()

	manager, err := factory(fig)
	if err != nil {
		return nil, err
	}

	registry.mu.Lock()
	if existing := registry.managers[fig]; existing != nil {
		registry.mu.Unlock()
		_ = manager.Close()
		return existing, nil
	}
	registry.managers[fig] = manager
	registry.mu.Unlock()
	return manager, nil
}
