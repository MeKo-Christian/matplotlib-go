package backends

import (
	"errors"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cwbudde/matplotlib-go/canvas"
	"github.com/cwbudde/matplotlib-go/render"
)

type headlessCanvas struct {
	mu          sync.Mutex
	figure      *canvas.Figure
	backend     Backend
	required    []Capability
	config      Config
	factory     Factory
	renderer    render.Renderer
	dispatcher  canvas.Dispatcher
	pendingDraw bool
	closed      bool
}

type defaultManager struct {
	canvas  canvas.FigureCanvas
	tools   *canvas.ToolManager
	title   string
	home    figureHomeState
	saveFn  func(string, ...render.SaveOption) error
	homeFn  func() error
	drawFn  func() error
	closeFn func() error
}

type figureHomeState struct {
	width  int
	height int
	axes   []axesHomeState
}

type axesHomeState struct {
	axes       *canvas.Axes
	xMin, xMax float64
	yMin, yMax float64
}

func NewCanvas(choice string, config Config, fig *canvas.Figure, required []Capability) (canvas.FigureCanvas, Backend, error) {
	backend, info, err := resolveBackendInfo(choice, required)
	if err != nil {
		return nil, "", err
	}
	if info.CanvasFactory != nil {
		out, err := info.CanvasFactory(config, fig)
		return out, backend, err
	}
	return newHeadlessCanvas(fig, backend, config, info.Factory, required), backend, nil
}

func NewCanvasFromEnv(config Config, fig *canvas.Figure, required []Capability) (canvas.FigureCanvas, Backend, error) {
	return NewCanvas(strings.TrimSpace(os.Getenv("MATPLOTLIB_BACKEND")), config, fig, required)
}

func NewManager(choice string, config Config, fig *canvas.Figure, required []Capability) (canvas.FigureManager, Backend, error) {
	backend, info, err := resolveBackendInfo(choice, required)
	if err != nil {
		return nil, "", err
	}
	if info.ManagerFactory != nil {
		out, err := info.ManagerFactory(config, fig)
		return out, backend, err
	}
	if info.CanvasFactory != nil {
		out, err := info.CanvasFactory(config, fig)
		if err != nil {
			return nil, "", err
		}
		return newDefaultManager(out, nil), backend, nil
	}
	headless := newHeadlessCanvas(fig, backend, config, info.Factory, required)
	return newDefaultManager(headless, headless.save), backend, nil
}

func NewManagerFromEnv(config Config, fig *canvas.Figure, required []Capability) (canvas.FigureManager, Backend, error) {
	return NewManager(strings.TrimSpace(os.Getenv("MATPLOTLIB_BACKEND")), config, fig, required)
}

func resolveBackendInfo(choice string, required []Capability) (Backend, *BackendInfo, error) {
	backend, err := ResolveBackend(choice, required)
	if err != nil {
		return "", nil, err
	}
	info, ok := DefaultRegistry.Get(backend)
	if !ok {
		return "", nil, fmt.Errorf("unknown backend: %s", backend)
	}
	return backend, info, nil
}

func newHeadlessCanvas(fig *canvas.Figure, backend Backend, config Config, factory Factory, required []Capability) *headlessCanvas {
	return &headlessCanvas{
		figure:   fig,
		backend:  backend,
		required: append([]Capability(nil), required...),
		config:   config,
		factory:  factory,
	}
}

func (c *headlessCanvas) Figure() *canvas.Figure {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.figure
}

func (c *headlessCanvas) Draw() error {
	event, err := c.draw(false)
	if err != nil {
		return err
	}
	return c.dispatcher.Emit(event)
}

func (c *headlessCanvas) DrawIdle() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return errors.New("backends: canvas is closed")
	}
	if c.pendingDraw {
		c.mu.Unlock()
		return nil
	}
	c.pendingDraw = true
	c.mu.Unlock()

	event, err := c.draw(false)
	if err != nil {
		return err
	}
	return c.dispatcher.Emit(event)
}

// FrameRGBA returns the pixels produced by the most recent Draw when the active
// renderer exposes an RGBA buffer (render.RGBAExporter). It implements
// canvas.RasterCanvas so movie writers can grab frames from the canvas, matching
// matplotlib reading its Agg canvas buffer_rgba after a draw.
func (c *headlessCanvas) FrameRGBA() *image.RGBA {
	c.mu.Lock()
	renderer := c.renderer
	c.mu.Unlock()
	exporter, ok := renderer.(render.RGBAExporter)
	if !ok {
		return nil
	}
	return exporter.GetImage()
}

func (c *headlessCanvas) Resize(width, height int) error {
	if width <= 0 || height <= 0 {
		return errors.New("backends: resize dimensions must be positive")
	}

	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return errors.New("backends: canvas is closed")
	}
	c.config.Width = width
	c.config.Height = height
	if c.figure != nil {
		c.figure.SizePx.X = float64(width)
		c.figure.SizePx.Y = float64(height)
	}
	event := c.normalizeEvent(canvas.Event{
		Type:   canvas.EventResize,
		Width:  width,
		Height: height,
	})
	c.mu.Unlock()

	if err := c.dispatcher.Emit(event); err != nil {
		return err
	}
	drawEvent, err := c.draw(false)
	if err != nil {
		return err
	}
	return c.dispatcher.Emit(drawEvent)
}

func (c *headlessCanvas) Connect(eventType canvas.EventType, handler canvas.Handler) canvas.ConnectionID {
	return c.dispatcher.Connect(eventType, handler)
}

func (c *headlessCanvas) Disconnect(id canvas.ConnectionID) {
	c.dispatcher.Disconnect(id)
}

func (c *headlessCanvas) Close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	c.renderer = nil
	event := c.normalizeEvent(canvas.Event{Type: canvas.EventClose})
	c.mu.Unlock()
	return c.dispatcher.Emit(event)
}

func (c *headlessCanvas) save(path string, opts ...render.SaveOption) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("backends: save path is required")
	}
	ext := normalizeSaveFormat(filepath.Ext(path))
	if ext == "" {
		return fmt.Errorf("backends: unsupported save extension %q", filepath.Ext(path))
	}

	backend, factory, err := c.saveBackend(ext)
	if err != nil {
		return err
	}
	saveOptions := render.ResolveSaveOptions(opts...)
	if err := saveOptions.ValidateForExtension(ext); err != nil {
		return err
	}
	renderer, err := c.drawWithFactory(factory, opts...)
	if err != nil {
		return err
	}

	if err := DefaultRegistry.SaveViaExtension(backend, renderer, path, opts...); err != nil {
		return fmt.Errorf("backends: cannot save figure: %w", err)
	}
	return nil
}

func (c *headlessCanvas) saveBackend(ext string) (Backend, Factory, error) {
	c.mu.Lock()
	backend := c.backend
	factory := c.factory
	required := append([]Capability(nil), c.required...)
	c.mu.Unlock()

	if _, err := DefaultRegistry.SelectBackendForExtension(string(backend), ext, required); err == nil {
		return backend, factory, nil
	}

	fallback, err := DefaultRegistry.SelectBackendForExtension("", ext, required)
	if err != nil {
		return "", nil, err
	}
	info, ok := DefaultRegistry.Get(fallback)
	if !ok || info.Factory == nil {
		return "", nil, fmt.Errorf("backends: selected backend %s has no renderer factory", fallback)
	}
	return fallback, info.Factory, nil
}

func (c *headlessCanvas) drawWithFactory(factory Factory, opts ...render.SaveOption) (render.Renderer, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil, errors.New("backends: canvas is closed")
	}
	if factory == nil {
		return nil, errors.New("backends: nil renderer factory")
	}
	renderer, err := factory(c.config)
	if err != nil {
		return nil, err
	}
	applySaveOptions(renderer, render.ResolveSaveOptions(opts...))
	canvas.DrawFigure(c.figure, renderer)
	c.renderer = renderer
	c.pendingDraw = false
	return renderer, nil
}

func (c *headlessCanvas) draw(skipEmit bool) (canvas.Event, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return canvas.Event{}, errors.New("backends: canvas is closed")
	}
	c.pendingDraw = false
	if c.factory == nil {
		return canvas.Event{}, errors.New("backends: nil renderer factory")
	}

	renderer, err := c.factory(c.config)
	if err != nil {
		return canvas.Event{}, err
	}
	canvas.DrawFigure(c.figure, renderer)
	c.renderer = renderer

	event := c.normalizeEvent(canvas.Event{
		Type:   canvas.EventDraw,
		Width:  c.config.Width,
		Height: c.config.Height,
	})
	return event, nil
}

func (c *headlessCanvas) normalizeEvent(event canvas.Event) canvas.Event {
	if event.Figure == nil {
		event.Figure = c.figure
	}
	if event.Width == 0 {
		event.Width = c.config.Width
	}
	if event.Height == 0 {
		event.Height = c.config.Height
	}
	if event.Axes == nil && event.Figure != nil {
		if ax, data, ok := canvas.ResolveEventTarget(event.Figure, event.Position); ax != nil {
			event.Axes = ax
			event.DataPosition = data
			event.HasDataPosition = ok
		}
	}
	return event
}

func newDefaultManager(figCanvas canvas.FigureCanvas, saveFn func(string, ...render.SaveOption) error) canvas.FigureManager {
	manager := &defaultManager{
		canvas: figCanvas,
		tools:  canvas.NewToolManager(),
		saveFn: saveFn,
	}
	manager.drawFn = figCanvas.Draw
	manager.closeFn = figCanvas.Close
	manager.home = snapshotFigureHome(figCanvas.Figure())
	manager.homeFn = func() error {
		return restoreFigureHome(manager.home, figCanvas)
	}
	manager.tools.Register(canvas.ToolFunc{
		Name: "home",
		Run: func(canvas.ToolArgs) error {
			return manager.homeFn()
		},
	})
	manager.tools.Register(canvas.ToolFunc{
		Name: "redraw",
		Run: func(canvas.ToolArgs) error {
			return manager.drawFn()
		},
	})
	manager.tools.Register(canvas.ToolFunc{
		Name: "save",
		Run: func(args canvas.ToolArgs) error {
			if manager.saveFn == nil {
				return errors.New("backends: active canvas does not support save")
			}
			opts, err := saveOptionsFromPayload(args.Payload)
			if err != nil {
				return err
			}
			return manager.saveFn(args.Path, opts...)
		},
	})
	return manager
}

func applySaveOptions(renderer render.Renderer, opts render.SaveOptions) {
	if setter, ok := renderer.(render.SVGOptionSetter); ok {
		setter.SetSVGOptions(opts.SVG)
	}
	if setter, ok := renderer.(render.PDFOptionSetter); ok {
		setter.SetPDFOptions(opts.PDF)
	}
	if setter, ok := renderer.(render.PSOptionSetter); ok {
		setter.SetPSOptions(opts.PS)
	}
	if setter, ok := renderer.(render.PGFOptionSetter); ok {
		setter.SetPGFOptions(opts.PGF)
	}
}

func saveOptionsFromPayload(payload any) ([]render.SaveOption, error) {
	switch v := payload.(type) {
	case nil:
		return nil, nil
	case render.SaveOption:
		return []render.SaveOption{v}, nil
	case []render.SaveOption:
		return v, nil
	case render.SaveOptions:
		return []render.SaveOption{v}, nil
	default:
		return nil, fmt.Errorf("backends: unsupported save tool payload %T", payload)
	}
}

func (m *defaultManager) Canvas() canvas.FigureCanvas { return m.canvas }

func (m *defaultManager) Show() error {
	if m == nil || m.drawFn == nil {
		return nil
	}
	return m.drawFn()
}

func (m *defaultManager) Close() error {
	if m == nil || m.closeFn == nil {
		return nil
	}
	return m.closeFn()
}

func (m *defaultManager) SetTitle(title string) { m.title = title }

func (m *defaultManager) ToolManager() *canvas.ToolManager { return m.tools }

func snapshotFigureHome(fig *canvas.Figure) figureHomeState {
	state := figureHomeState{}
	if fig == nil {
		return state
	}
	state.width = int(fig.SizePx.X + 0.5)
	state.height = int(fig.SizePx.Y + 0.5)
	state.axes = make([]axesHomeState, 0, len(fig.Children))
	for _, ax := range fig.Children {
		if ax == nil {
			continue
		}
		xMin, xMax := 0.0, 1.0
		if ax.XScale != nil {
			xMin, xMax = ax.XScale.Domain()
		}
		yMin, yMax := 0.0, 1.0
		if ax.YScale != nil {
			yMin, yMax = ax.YScale.Domain()
		}
		state.axes = append(state.axes, axesHomeState{
			axes: ax,
			xMin: xMin,
			xMax: xMax,
			yMin: yMin,
			yMax: yMax,
		})
	}
	return state
}

var (
	_ canvas.DrawIdleCanvas = (*headlessCanvas)(nil)
	_ canvas.RasterCanvas   = (*headlessCanvas)(nil)
	_ canvas.FigureManager  = (*defaultManager)(nil)
)

func restoreFigureHome(state figureHomeState, figCanvas canvas.FigureCanvas) error {
	fig := figCanvas.Figure()
	if fig == nil {
		return nil
	}
	for _, axState := range state.axes {
		if axState.axes == nil {
			continue
		}
		axState.axes.SetXLim(axState.xMin, axState.xMax)
		axState.axes.SetYLim(axState.yMin, axState.yMax)
	}
	width := int(fig.SizePx.X + 0.5)
	height := int(fig.SizePx.Y + 0.5)
	if state.width > 0 && state.height > 0 && (width != state.width || height != state.height) {
		return figCanvas.Resize(state.width, state.height)
	}
	return figCanvas.Draw()
}
