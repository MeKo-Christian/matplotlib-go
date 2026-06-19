package pyplot

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/canvas"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

// ShowHandler renders or presents a figure when Show or Pause is called.
type ShowHandler func(*core.Figure) error

// Savefig renders the current figure to a file selected by extension.
func Savefig(path string, opts ...render.SaveOption) error {
	return saveFigure(GCF(), path, opts...)
}

// SwitchBackend selects the backend used for future pyplot figure managers.
// Cached managers are closed so existing figures will be recreated through the
// selected backend when next shown or drawn.
func SwitchBackend(choice string) error {
	backend, err := backends.ResolveBackend(choice, backends.TextCapabilities)
	if err != nil {
		return fmt.Errorf("pyplot: switch_backend %q: %w", choice, err)
	}
	closeCachedManagersAndSetFactory(func(fig *core.Figure) (canvas.FigureManager, error) {
		manager, _, err := backends.NewManager(string(backend), rendererConfig(fig), fig, backends.TextCapabilities)
		return manager, err
	})
	return nil
}

// SetShowHandler overrides how Show and Pause present figures. Passing nil
// restores the default manager-backed behavior.
func SetShowHandler(handler ShowHandler) {
	if handler == nil {
		SetManagerFactory(nil)
		return
	}
	SetManagerFactory(func(fig *core.Figure) (canvas.FigureManager, error) {
		return newShowHandlerManager(fig, handler), nil
	})
}

// Show renders all registered figures through the configured show handler.
func Show() error {
	registry.mu.Lock()
	figures := append([]*core.Figure(nil), registry.figures...)
	registry.mu.Unlock()

	for _, fig := range figures {
		if fig == nil {
			continue
		}
		manager, err := ensureManager(fig)
		if err != nil {
			return err
		}
		if err := manager.Show(); err != nil {
			return err
		}
	}
	return nil
}

// Draw redraws the current figure through its manager canvas.
func Draw() error {
	manager, err := ensureManager(GCF())
	if err != nil {
		return err
	}
	c := manager.Canvas()
	if c == nil {
		return nil
	}
	if idle, ok := c.(canvas.DrawIdleCanvas); ok {
		return idle.DrawIdle()
	}
	return c.Draw()
}

// DrawIfInteractive redraws the current figure only when interactive mode is enabled.
func DrawIfInteractive() error {
	if !IsInteractive() {
		return nil
	}
	return Draw()
}

// Connect registers an event handler on the current figure canvas.
func Connect(eventType canvas.EventType, handler canvas.Handler) (canvas.ConnectionID, error) {
	manager, err := ensureManager(GCF())
	if err != nil {
		return 0, err
	}
	return canvas.Connect(manager.Canvas(), eventType, handler), nil
}

// Disconnect removes an event handler from the current figure canvas.
func Disconnect(id canvas.ConnectionID) error {
	manager, err := ensureManager(GCF())
	if err != nil {
		return err
	}
	canvas.Disconnect(manager.Canvas(), id)
	return nil
}

// Pause renders open figures and then blocks for the requested interval.
func Pause(interval time.Duration) error {
	if err := Show(); err != nil {
		return err
	}
	if interval > 0 {
		time.Sleep(interval)
	}
	return nil
}

func defaultManagerFactory(fig *core.Figure) (canvas.FigureManager, error) {
	manager, _, err := backends.NewManagerFromEnv(rendererConfig(fig), fig, backends.TextCapabilities)
	if err != nil {
		return nil, err
	}
	return manager, nil
}

func saveFigure(fig *core.Figure, path string, opts ...render.SaveOption) error {
	if fig == nil {
		return errors.New("pyplot: nil figure")
	}

	ext := strings.ToLower(filepath.Ext(path))
	if ext == "" {
		return fmt.Errorf("pyplot: savefig path %q has no extension", path)
	}

	choice := strings.TrimSpace(os.Getenv("MATPLOTLIB_BACKEND"))
	cfg := rendererConfig(fig)

	backend, err := backends.SelectBackendForExtension(choice, ext, nil)
	if err != nil {
		// If the user pinned a backend that does not handle this extension,
		// retry with auto so the registry picks any backend that can.
		if choice != "" {
			backend, err = backends.SelectBackendForExtension("", ext, nil)
		}
		if err != nil {
			return fmt.Errorf("pyplot: %w", err)
		}
	}

	renderer, err := backends.Create(backend, cfg)
	if err != nil {
		return fmt.Errorf("pyplot: create %s renderer: %w", backend, err)
	}

	saveOptions := render.ResolveSaveOptions(opts...)
	if err := saveOptions.ValidateForExtension(ext); err != nil {
		return fmt.Errorf("pyplot: %w", err)
	}
	if setter, ok := renderer.(render.SVGOptionSetter); ok {
		setter.SetSVGOptions(saveOptions.SVG)
	}
	if setter, ok := renderer.(render.PDFOptionSetter); ok {
		setter.SetPDFOptions(saveOptions.PDF)
	}
	if setter, ok := renderer.(render.PSOptionSetter); ok {
		setter.SetPSOptions(saveOptions.PS)
	}
	if setter, ok := renderer.(render.PGFOptionSetter); ok {
		setter.SetPGFOptions(saveOptions.PGF)
	}
	core.DrawFigure(fig, renderer)

	if err := backends.DefaultRegistry.SaveViaExtension(backend, renderer, path, opts...); err != nil {
		return fmt.Errorf("pyplot: %w", err)
	}
	return nil
}

func rendererConfig(fig *core.Figure) backends.Config {
	defaults := style.CurrentDefaults()
	background := defaults.FigureBackground()
	dpi := defaults.DPI
	width, height := defaults.DefaultFigureSizePx()

	if fig != nil {
		if fig.SizePx.X > 0 {
			width = int(fig.SizePx.X + 0.5)
		}
		if fig.SizePx.Y > 0 {
			height = int(fig.SizePx.Y + 0.5)
		}
		background = fig.RC.FigureBackground()
		dpi = fig.RC.DPI
	}

	return backends.Config{
		Width:      width,
		Height:     height,
		Background: background,
		DPI:        dpi,
	}
}

type showHandlerManager struct {
	canvas *showHandlerCanvas
	tools  *canvas.ToolManager
}

type showHandlerCanvas struct {
	figure     *core.Figure
	handler    ShowHandler
	dispatcher canvas.Dispatcher
	closed     bool
}

func newShowHandlerManager(fig *core.Figure, handler ShowHandler) canvas.FigureManager {
	c := &showHandlerCanvas{figure: fig, handler: handler}
	manager := &showHandlerManager{
		canvas: c,
		tools:  canvas.NewToolManager(),
	}
	manager.tools.Register(canvas.ToolFunc{
		Name: "redraw",
		Run: func(canvas.ToolArgs) error {
			return c.Draw()
		},
	})
	return manager
}

func (m *showHandlerManager) Canvas() canvas.FigureCanvas { return m.canvas }

func (m *showHandlerManager) Show() error { return m.canvas.Draw() }

func (m *showHandlerManager) Close() error { return m.canvas.Close() }

func (m *showHandlerManager) SetTitle(string) {}

func (m *showHandlerManager) ToolManager() *canvas.ToolManager { return m.tools }

func (c *showHandlerCanvas) Figure() *core.Figure { return c.figure }

func (c *showHandlerCanvas) Draw() error {
	if c.closed {
		return nil
	}
	if c.handler == nil {
		return nil
	}
	if err := c.handler(c.figure); err != nil {
		return err
	}
	return c.dispatcher.Emit(canvas.Event{
		Type:   canvas.EventDraw,
		Figure: c.figure,
		Width:  int(c.figure.SizePx.X + 0.5),
		Height: int(c.figure.SizePx.Y + 0.5),
	})
}

func (c *showHandlerCanvas) Resize(width, height int) error {
	if c.closed {
		return nil
	}
	if c.figure != nil {
		c.figure.SizePx.X = float64(width)
		c.figure.SizePx.Y = float64(height)
	}
	if err := c.dispatcher.Emit(canvas.Event{
		Type:   canvas.EventResize,
		Figure: c.figure,
		Width:  width,
		Height: height,
	}); err != nil {
		return err
	}
	return c.Draw()
}

func (c *showHandlerCanvas) Connect(eventType canvas.EventType, handler canvas.Handler) canvas.ConnectionID {
	return c.dispatcher.Connect(eventType, handler)
}

func (c *showHandlerCanvas) Disconnect(id canvas.ConnectionID) {
	c.dispatcher.Disconnect(id)
}

func (c *showHandlerCanvas) Close() error {
	if c.closed {
		return nil
	}
	c.closed = true
	return c.dispatcher.Emit(canvas.Event{
		Type:   canvas.EventClose,
		Figure: c.figure,
	})
}
