//go:build js && wasm

package wasm

import (
	"errors"
	"fmt"
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/backends/gobasic"
	plotcanvas "github.com/cwbudde/matplotlib-go/canvas"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"syscall/js"
)

type rasterRenderer interface {
	render.Renderer
	GetImage() *image.RGBA
}

type rasterRendererFactory func(w, h int, bg render.Color) (rasterRenderer, error)

type listener struct {
	target js.Value
	event  string
	fn     js.Func
}

// Manager is the wasm figure manager. It hosts a renderer-backed canvas,
// dispatches DOM events, and wires the standard [canvas.Navigation] and
// [canvas.ToolbarController] so browser demos support pan, zoom-to-rect,
// scroll-zoom, and home reset.
//
// Manager satisfies [canvas.FigureManager]; callers that need pan/zoom
// programmatically can reach for [Manager.Navigation] or
// [Manager.Toolbar].
type Manager struct {
	canvas *figureCanvas
	tools  *plotcanvas.ToolManager
	nav    *plotcanvas.Navigation
	wi     *plotcanvas.WidgetInteraction
	tb     *plotcanvas.ToolbarController
	home   figureHomeState
}

type figureCanvas struct {
	figure     *core.Figure
	element    js.Value
	context    js.Value
	factory    rasterRendererFactory
	dispatcher plotcanvas.Dispatcher
	hover      *plotcanvas.AxesHoverTracker
	listeners  []listener
	closed     bool
	nav        *plotcanvas.Navigation
	wi         *plotcanvas.WidgetInteraction
}

type figureHomeState struct {
	width  int
	height int
	axes   []axesHomeState
}

type axesHomeState struct {
	axes       *core.Axes
	xMin, xMax float64
	yMin, yMax float64
}

// NewGoBasicManager builds a wasm manager backed by the pure-Go gobasic
// renderer.
func NewGoBasicManager(elementID string, fig *core.Figure) (*Manager, error) {
	return newManager(elementID, fig, newGoBasicRenderer)
}

// NewAggManager builds a wasm manager backed by the AGG renderer.
func NewAggManager(elementID string, fig *core.Figure) (*Manager, error) {
	return newManager(elementID, fig, newAggRenderer)
}

func newManager(elementID string, fig *core.Figure, factory rasterRendererFactory) (*Manager, error) {
	if fig == nil {
		return nil, errors.New("canvas/wasm: nil figure")
	}
	if factory == nil {
		return nil, errors.New("canvas/wasm: nil renderer factory")
	}

	document := js.Global().Get("document")
	if document.IsUndefined() || document.IsNull() {
		return nil, errors.New("canvas/wasm: document is unavailable")
	}
	element := document.Call("getElementById", elementID)
	if element.IsUndefined() || element.IsNull() {
		return nil, fmt.Errorf("canvas/wasm: canvas element %q not found", elementID)
	}
	context := element.Call("getContext", "2d")
	if context.IsUndefined() || context.IsNull() {
		return nil, errors.New("canvas/wasm: 2d context is unavailable")
	}

	c := &figureCanvas{
		figure:  fig,
		element: element,
		context: context,
		factory: factory,
	}
	c.hover = plotcanvas.NewAxesHoverTracker(fig, &c.dispatcher)
	if tabIndex := element.Get("tabIndex"); tabIndex.IsUndefined() || tabIndex.Int() < 0 {
		element.Set("tabIndex", 0)
	}

	m := &Manager{
		canvas: c,
		tools:  plotcanvas.NewToolManager(),
		home:   snapshotFigureHome(fig),
	}

	// Wire the standard pan / zoom-to-rect / scroll-zoom helper onto the
	// dispatcher. After a navigation mutation the helper calls Draw,
	// which both repaints the canvas and re-evaluates any zoom rubber
	// band overlay.
	m.nav = plotcanvas.NewNavigation(fig, c.Draw)
	c.nav = m.nav
	m.nav.Attach(&c.dispatcher)
	m.wi = plotcanvas.NewWidgetInteraction(fig, c.Draw)
	c.wi = m.wi
	m.wi.Attach(&c.dispatcher)

	// Toolbar wraps the helper with standard pan / zoom toggles and the
	// home action. Save is intentionally left handler-less: the browser
	// host downloads the canvas via toBlob, not through Go.
	m.tb = plotcanvas.NewToolbarController(m.nav)
	m.tb.SetHandler(plotcanvas.ToolbarHome, func() error {
		return restoreFigureHome(m.home, c)
	})

	// Install DOM listeners after Navigation is attached so the helper
	// observes the very first event.
	c.installListeners()

	m.tools.Register(plotcanvas.ToolFunc{
		Name: "home",
		Run: func(plotcanvas.ToolArgs) error {
			return m.tb.Trigger(plotcanvas.ToolbarHome)
		},
	})
	m.tools.Register(plotcanvas.ToolFunc{
		Name: "redraw",
		Run: func(plotcanvas.ToolArgs) error {
			return c.Draw()
		},
	})
	m.tools.Register(plotcanvas.ToolFunc{
		Name: "pan",
		Run: func(plotcanvas.ToolArgs) error {
			return m.tb.Trigger(plotcanvas.ToolbarPan)
		},
	})
	m.tools.Register(plotcanvas.ToolFunc{
		Name: "zoom",
		Run: func(plotcanvas.ToolArgs) error {
			return m.tb.Trigger(plotcanvas.ToolbarZoom)
		},
	})
	m.tools.Register(plotcanvas.ToolFunc{
		Name: "save",
		Run: func(plotcanvas.ToolArgs) error {
			return m.tb.Trigger(plotcanvas.ToolbarSave)
		},
	})

	return m, nil
}

func newGoBasicRenderer(w, h int, bg render.Color) (rasterRenderer, error) {
	r := gobasic.New(w, h, bg)
	if r == nil {
		return nil, errors.New("canvas/wasm: failed to create GoBasic renderer")
	}
	return r, nil
}

func newAggRenderer(w, h int, bg render.Color) (rasterRenderer, error) {
	r, err := agg.New(w, h, bg)
	if err != nil {
		return nil, fmt.Errorf("canvas/wasm: create AGG renderer: %w", err)
	}
	return r, nil
}

// Canvas returns the canvas view of the manager.
func (m *Manager) Canvas() plotcanvas.FigureCanvas { return m.canvas }

// Show paints the figure once into the host canvas.
func (m *Manager) Show() error { return m.canvas.Draw() }

// Close detaches DOM listeners, releases the navigation helper, and
// emits the close lifecycle event.
func (m *Manager) Close() error {
	if m.nav != nil {
		m.nav.Detach()
	}
	if m.wi != nil {
		m.wi.Detach()
	}
	return m.canvas.Close()
}

// SetTitle sets the browser document title.
func (m *Manager) SetTitle(title string) {
	document := js.Global().Get("document")
	if document.IsUndefined() || document.IsNull() {
		return
	}
	document.Set("title", title)
}

// ToolManager returns the named-tool registry. Pan, zoom, home, save
// tools are registered up front and delegate to the toolbar controller.
func (m *Manager) ToolManager() *plotcanvas.ToolManager { return m.tools }

// Navigation returns the navigation helper, so callers can change mode
// or query an in-progress drag rectangle.
func (m *Manager) Navigation() *plotcanvas.Navigation { return m.nav }

// Toolbar returns the toolbar controller. Hosts can register a save
// handler or override toolbar actions through it.
func (m *Manager) Toolbar() *plotcanvas.ToolbarController { return m.tb }

func (c *figureCanvas) Figure() *core.Figure { return c.figure }

func (c *figureCanvas) Draw() error {
	if c.closed {
		return errors.New("canvas/wasm: canvas is closed")
	}
	cssWidth, cssHeight := c.currentSize()
	if cssWidth <= 0 || cssHeight <= 0 {
		return errors.New("canvas/wasm: invalid canvas size")
	}

	pixelRatio := devicePixelRatio()
	backingWidth := int(math.Round(float64(cssWidth) * pixelRatio))
	backingHeight := int(math.Round(float64(cssHeight) * pixelRatio))
	if backingWidth <= 0 || backingHeight <= 0 {
		return errors.New("canvas/wasm: invalid backing size")
	}

	c.element.Set("width", backingWidth)
	c.element.Set("height", backingHeight)
	c.figure.SizePx.X = float64(cssWidth)
	c.figure.SizePx.Y = float64(cssHeight)

	renderer, err := c.factory(backingWidth, backingHeight, c.figure.RC.FigureBackground())
	if err != nil {
		return err
	}
	core.DrawFigure(c.figure, newScaledRenderer(renderer, pixelRatio))

	img := renderer.GetImage()
	pixels := js.Global().Get("Uint8ClampedArray").New(len(img.Pix))
	js.CopyBytesToJS(pixels, img.Pix)
	imageData := js.Global().Get("ImageData").New(pixels, backingWidth, backingHeight)
	c.context.Call("putImageData", imageData, 0, 0)

	// Overlay the in-progress zoom rubber band, if any, so users see the
	// rectangle they are dragging. We stroke directly on the 2D context;
	// since Draw is invoked on every mouse-move during a zoom drag, the
	// rectangle tracks the cursor.
	c.drawRubberband(pixelRatio)

	return c.dispatcher.Emit(plotcanvas.Event{
		Type:   plotcanvas.EventDraw,
		Figure: c.figure,
		Width:  cssWidth,
		Height: cssHeight,
	})
}

func (c *figureCanvas) Resize(width, height int) error {
	if c.closed {
		return errors.New("canvas/wasm: canvas is closed")
	}
	if width <= 0 || height <= 0 {
		return errors.New("canvas/wasm: resize dimensions must be positive")
	}

	c.element.Set("width", width)
	c.element.Set("height", height)
	c.figure.SizePx.X = float64(width)
	c.figure.SizePx.Y = float64(height)

	if err := c.dispatcher.Emit(plotcanvas.Event{
		Type:   plotcanvas.EventResize,
		Figure: c.figure,
		Width:  width,
		Height: height,
	}); err != nil {
		return err
	}
	return c.Draw()
}

func (c *figureCanvas) Connect(eventType plotcanvas.EventType, handler plotcanvas.Handler) plotcanvas.ConnectionID {
	return c.dispatcher.Connect(eventType, handler)
}

func (c *figureCanvas) Disconnect(id plotcanvas.ConnectionID) {
	c.dispatcher.Disconnect(id)
}

func (c *figureCanvas) Close() error {
	if c.closed {
		return nil
	}
	c.closed = true
	for _, listener := range c.listeners {
		listener.target.Call("removeEventListener", listener.event, listener.fn)
		listener.fn.Release()
	}
	c.listeners = nil
	return c.dispatcher.Emit(plotcanvas.Event{
		Type:   plotcanvas.EventClose,
		Figure: c.figure,
		Width:  int(c.figure.SizePx.X + 0.5),
		Height: int(c.figure.SizePx.Y + 0.5),
	})
}

func (c *figureCanvas) installListeners() {
	c.on(c.element, "mousedown", func(this js.Value, args []js.Value) any {
		event := c.mouseEvent(plotcanvas.EventMousePress, args[0])
		c.focus()
		c.emit(event)
		plotcanvas.EmitPick(&c.dispatcher, c.figure, plotcanvas.MouseEvent{Event: event})
		return nil
	})
	c.on(c.element, "mouseup", func(this js.Value, args []js.Value) any {
		return c.emit(c.mouseEvent(plotcanvas.EventMouseRelease, args[0]))
	})
	c.on(c.element, "dblclick", func(this js.Value, args []js.Value) any {
		event := c.mouseEvent(plotcanvas.EventMousePress, args[0])
		event.DoubleClick = true
		c.emit(event)
		plotcanvas.EmitPick(&c.dispatcher, c.figure, plotcanvas.MouseEvent{Event: event})
		return nil
	})
	c.on(c.element, "mousemove", func(this js.Value, args []js.Value) any {
		event := c.mouseEvent(plotcanvas.EventMouseMove, args[0])
		result := c.emit(event)
		// During a zoom rubber-band drag the navigation helper updates
		// its rectangle without calling Draw — repaint here so the
		// overlay tracks the cursor.
		if c.nav != nil && c.nav.Mode() == plotcanvas.NavZoom {
			if _, ok := c.nav.DragRect(); ok {
				if err := c.Draw(); err != nil {
					js.Global().Get("console").Call("error", err.Error())
				}
			}
		}
		return result
	})
	c.on(c.element, "mouseenter", func(this js.Value, args []js.Value) any {
		return c.emit(c.mouseEvent(plotcanvas.EventFigureEnter, args[0]))
	})
	c.on(c.element, "mouseleave", func(this js.Value, args []js.Value) any {
		return c.emit(c.mouseEvent(plotcanvas.EventFigureLeave, args[0]))
	})
	c.on(c.element, "wheel", func(this js.Value, args []js.Value) any {
		args[0].Call("preventDefault")
		return c.emit(c.scrollEvent(args[0]))
	})
	c.on(c.element, "contextmenu", func(this js.Value, args []js.Value) any {
		// Right-click on the canvas pops the OS menu by default. When
		// pan or zoom is active that would interrupt the gesture, so
		// suppress it.
		if c.nav != nil && c.nav.Mode() != plotcanvas.NavNone {
			args[0].Call("preventDefault")
		}
		return nil
	})

	window := js.Global().Get("window")
	c.on(window, "keydown", func(this js.Value, args []js.Value) any {
		return c.emit(c.keyEvent(plotcanvas.EventKeyPress, args[0]))
	})
	c.on(window, "keyup", func(this js.Value, args []js.Value) any {
		return c.emit(c.keyEvent(plotcanvas.EventKeyRelease, args[0]))
	})
	c.on(window, "resize", func(this js.Value, _ []js.Value) any {
		width, height := c.elementClientSize()
		if width <= 0 || height <= 0 {
			return nil
		}
		if err := c.Resize(width, height); err != nil {
			js.Global().Get("console").Call("error", err.Error())
		}
		return nil
	})
}

func (c *figureCanvas) on(target js.Value, event string, fn func(this js.Value, args []js.Value) any) {
	callback := js.FuncOf(fn)
	target.Call("addEventListener", event, callback)
	c.listeners = append(c.listeners, listener{target: target, event: event, fn: callback})
}

func (c *figureCanvas) emit(event plotcanvas.Event) any {
	if c.closed {
		return nil
	}
	event.Figure = c.figure
	if event.Axes == nil {
		if ax, data, ok := plotcanvas.ResolveEventTarget(c.figure, event.Position); ax != nil {
			event.Axes = ax
			event.DataPosition = data
			event.HasDataPosition = ok
		}
	}
	if err := c.dispatcher.Emit(event); err != nil {
		js.Global().Get("console").Call("error", err.Error())
	}
	switch event.Type {
	case plotcanvas.EventMouseMove, plotcanvas.EventFigureEnter, plotcanvas.EventFigureLeave:
		c.hover.Update(event)
	}
	return nil
}

func (c *figureCanvas) mouseEvent(eventType plotcanvas.EventType, domEvent js.Value) plotcanvas.Event {
	position := elementPosition(c.element, domEvent)
	return plotcanvas.Event{
		Type:      eventType,
		Position:  position,
		Button:    mouseButton(domEvent.Get("button").Int()),
		Modifiers: modifiers(domEvent),
		Native:    domEvent,
	}
}

func (c *figureCanvas) scrollEvent(domEvent js.Value) plotcanvas.Event {
	position := elementPosition(c.element, domEvent)
	// The Navigation helper inverts step direction internally; scroll
	// up (negative deltaY in DOM coords) zooms in.
	return plotcanvas.Event{
		Type:      plotcanvas.EventScroll,
		Position:  position,
		DeltaX:    domEvent.Get("deltaX").Float(),
		DeltaY:    -domEvent.Get("deltaY").Float(),
		Modifiers: modifiers(domEvent),
		Native:    domEvent,
	}
}

func (c *figureCanvas) keyEvent(eventType plotcanvas.EventType, domEvent js.Value) plotcanvas.Event {
	return plotcanvas.Event{
		Type:      eventType,
		Figure:    c.figure,
		Key:       plotcanvas.NormalizeKey(domEvent.Get("key").String()),
		Modifiers: modifiers(domEvent),
		Native:    domEvent,
	}
}

// drawRubberband strokes a dashed rectangle on top of the most recent
// frame whenever the navigation helper reports an in-progress zoom
// drag. The rectangle is in CSS-pixel space, so we scale by the device
// pixel ratio to match the canvas backing buffer.
func (c *figureCanvas) drawRubberband(pixelRatio float64) {
	if c.nav == nil {
		return
	}
	rect, ok := c.nav.DragRect()
	if !ok {
		return
	}
	if rect.W() < 1 || rect.H() < 1 {
		return
	}
	ctx := c.context
	if ctx.IsUndefined() || ctx.IsNull() {
		return
	}
	x := rect.Min.X * pixelRatio
	y := rect.Min.Y * pixelRatio
	w := rect.W() * pixelRatio
	h := rect.H() * pixelRatio
	ctx.Call("save")
	ctx.Set("strokeStyle", "#000")
	ctx.Set("lineWidth", math.Max(1, pixelRatio))
	if setLineDash := ctx.Get("setLineDash"); setLineDash.Type() == js.TypeFunction {
		dashes := js.Global().Get("Array").New()
		dashes.Call("push", 4*pixelRatio)
		dashes.Call("push", 3*pixelRatio)
		ctx.Call("setLineDash", dashes)
	}
	ctx.Call("strokeRect", x, y, w, h)
	ctx.Call("restore")
}

func (c *figureCanvas) currentSize() (int, int) {
	width, height := c.elementClientSize()
	if width > 0 && height > 0 {
		return width, height
	}
	return int(c.figure.SizePx.X + 0.5), int(c.figure.SizePx.Y + 0.5)
}

func (c *figureCanvas) elementClientSize() (int, int) {
	width := c.element.Get("clientWidth").Int()
	height := c.element.Get("clientHeight").Int()
	return width, height
}

func (c *figureCanvas) focus() {
	if c.element.IsUndefined() || c.element.IsNull() {
		return
	}
	if focus := c.element.Get("focus"); focus.Type() == js.TypeFunction {
		c.element.Call("focus")
	}
}

func elementPosition(element, event js.Value) geom.Pt {
	rect := element.Call("getBoundingClientRect")
	x := event.Get("clientX").Float() - rect.Get("left").Float()
	y := event.Get("clientY").Float() - rect.Get("top").Float()
	return geom.Pt{X: x, Y: y}
}

func devicePixelRatio() float64 {
	ratio := js.Global().Get("devicePixelRatio")
	if ratio.Type() != js.TypeNumber {
		return 1
	}
	value := ratio.Float()
	if value <= 0 || math.IsNaN(value) || math.IsInf(value, 0) {
		return 1
	}
	return value
}

func mouseButton(button int) plotcanvas.MouseButton {
	return plotcanvas.MouseButtonFromJSIndex(button)
}

func modifiers(event js.Value) plotcanvas.Modifier {
	return plotcanvas.ModifierSet(
		event.Get("shiftKey").Bool(),
		event.Get("ctrlKey").Bool(),
		event.Get("altKey").Bool(),
		event.Get("metaKey").Bool(),
	)
}

func snapshotFigureHome(fig *core.Figure) figureHomeState {
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

func restoreFigureHome(state figureHomeState, c *figureCanvas) error {
	for _, axState := range state.axes {
		if axState.axes == nil {
			continue
		}
		axState.axes.SetXLim(axState.xMin, axState.xMax)
		axState.axes.SetYLim(axState.yMin, axState.yMax)
	}
	width := int(c.figure.SizePx.X + 0.5)
	height := int(c.figure.SizePx.Y + 0.5)
	if width != state.width || height != state.height {
		return c.Resize(state.width, state.height)
	}
	return c.Draw()
}
