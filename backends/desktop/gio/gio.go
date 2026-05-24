// Package gio is the Gio (gioui.org) adapter for the desktop interactive
// backend selected in docs/adr/0001-desktop-interactive-backend.md.
//
// Importing this package registers a desktop.Constructor with
// backends/desktop so callers of desktop.New automatically get the Gio
// implementation. Typical usage:
//
//	import (
//	    "github.com/cwbudde/matplotlib-go/backends/desktop"
//	    _ "github.com/cwbudde/matplotlib-go/backends/desktop/gio"
//	    "github.com/cwbudde/matplotlib-go/backends/agg"
//	)
//
//	func main() {
//	    fig := buildFigure()
//	    b, err := desktop.New(desktop.Options{
//	        Figure:   fig,
//	        Title:    "Figure 1",
//	        Renderer: func(w, h int, bg render.Color) (render.Renderer, error) {
//	            return agg.New(w, h, bg)
//	        },
//	    })
//	    if err != nil { log.Fatal(err) }
//	    go func() { _ = b.Run() }()
//	    gio.Main()
//	}
//
// The first time a Backend's Run method is called, a new OS window is
// created. Run blocks until the user closes the window (DestroyEvent).
// Main must be called from the program's main goroutine; see
// gioui.org/app.Main for details.
package gio

import (
	"errors"
	"image"
	"image/color"
	"sync"
	"sync/atomic"

	gioapp "gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/cwbudde/matplotlib-go/backends/desktop"
	"github.com/cwbudde/matplotlib-go/canvas"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func init() {
	desktop.Register(func(opts desktop.Options) (desktop.Backend, error) {
		return New(opts)
	})
}

// Main is an alias for gioui.org/app.Main. It must be called from the
// program's main goroutine, after launching one or more Backend.Run
// goroutines.
func Main() { gioapp.Main() }

// Backend is the Gio implementation of desktop.Backend. One Backend owns
// one Gio app.Window and one event-pump.
type Backend struct {
	opts desktop.Options
	cnv  *gioCanvas
	nav  *canvas.Navigation
	wi   *canvas.WidgetInteraction
	tb   *canvas.ToolbarController

	window *gioapp.Window

	runOnce sync.Once
	closed  atomic.Bool
}

// New constructs a Gio Backend for the supplied options. The window is
// not opened until Run is called.
func New(opts desktop.Options) (*Backend, error) {
	opts = desktop.WithDefaults(opts)
	if opts.Figure == nil {
		return nil, desktop.ErrNoFigure
	}

	w := new(gioapp.Window)
	w.Option(
		gioapp.Title(opts.Title),
		gioapp.Size(unit.Dp(opts.Width), unit.Dp(opts.Height)),
	)

	cnv := newGioCanvas(opts, w)

	nav := opts.Navigation
	if nav == nil {
		nav = canvas.NewNavigation(opts.Figure, cnv.DrawIdle)
	}
	nav.Attach(cnv.dispatcher)
	wi := canvas.NewWidgetInteraction(opts.Figure, cnv.DrawIdle)
	wi.Attach(cnv.dispatcher)

	tb := opts.Toolbar
	if tb == nil {
		tb = canvas.NewToolbarController(nav)
		if opts.SaveHandler != nil {
			tb.SetHandler(canvas.ToolbarSave, opts.SaveHandler)
		}
	}

	return &Backend{
		opts:   opts,
		cnv:    cnv,
		nav:    nav,
		wi:     wi,
		tb:     tb,
		window: w,
	}, nil
}

// Canvas returns the figure canvas wired to the Gio window.
func (b *Backend) Canvas() canvas.DrawIdleCanvas { return b.cnv }

// Navigation returns the navigation helper attached to the canvas.
func (b *Backend) Navigation() *canvas.Navigation { return b.nav }

// Toolbar returns the toolbar controller attached to the canvas.
func (b *Backend) Toolbar() canvas.NavigationToolbar { return b.tb }

// Close requests window destruction. The Run loop returns shortly after.
// Close is idempotent.
func (b *Backend) Close() error {
	if b.closed.Swap(true) {
		return nil
	}
	if b.wi != nil {
		b.wi.Detach()
	}
	if err := b.cnv.emitClose(); err != nil {
		return err
	}
	// Triggering Invalidate plus setting closed.Load makes Run exit at
	// the next event boundary; the underlying window will receive a
	// DestroyEvent once the OS confirms.
	if b.window != nil {
		b.window.Invalidate()
	}
	return nil
}

// Run blocks pumping the window's event loop until the user closes the
// window. Run is safe to call only once per backend; subsequent calls
// return nil immediately. Run must be invoked from a goroutine other
// than the one that calls Main.
func (b *Backend) Run() error {
	var runErr error
	b.runOnce.Do(func() {
		runErr = b.runLoop()
	})
	return runErr
}

func (b *Backend) runLoop() error {
	var ops op.Ops
	// tag identifies our input handler. Using the backend pointer
	// guarantees uniqueness without a runtime allocation per frame.
	tag := event.Tag(b)

	for {
		e := b.window.Event()
		switch e := e.(type) {
		case gioapp.DestroyEvent:
			if !b.closed.Swap(true) {
				_ = b.cnv.emitClose()
			}
			return e.Err
		case gioapp.FrameEvent:
			gtx := gioapp.NewContext(&ops, e)

			// 1. Drain queued input events targeting our tag. These
			//    correspond to event.Op registrations from the previous
			//    frame; on the first frame the queue is empty.
			b.drainInput(gtx.Source, tag)

			// 2. Reflect any size change before drawing the figure.
			if w, h := e.Size.X, e.Size.Y; w != b.cnv.Width() || h != b.cnv.Height() {
				_ = b.cnv.resize(w, h)
			}

			// 3. Ensure we have a fresh rendered bitmap.
			img, err := b.cnv.ensureImage()

			// 4. Build this frame's ops. We register the event tag over
			//    the entire window so future pointer / key events come
			//    back to us, then blit the AGG bitmap as the window
			//    contents.
			area := clip.Rect{Max: e.Size}.Push(gtx.Ops)
			event.Op(gtx.Ops, tag)
			if err == nil && img != nil {
				paint.NewImageOp(img).Add(gtx.Ops)
				paint.PaintOp{}.Add(gtx.Ops)
			} else {
				// Clear to background while a draw error or pending
				// renderer keeps the bitmap unavailable.
				paint.Fill(gtx.Ops, colorNRGBA(b.opts.Background))
			}
			area.Pop()

			e.Frame(gtx.Ops)

			// 5. Notify dispatcher handlers that a draw completed.
			_ = b.cnv.emitDraw(e.Size.X, e.Size.Y)

			if b.closed.Load() {
				// External Close() during this frame; let the OS deliver
				// DestroyEvent on the next iteration.
				continue
			}
		}
	}
}

// drainInput consumes every queued pointer/key event addressed to tag
// and forwards it through the canvas dispatcher.
func (b *Backend) drainInput(src interface {
	Event(filters ...event.Filter) (event.Event, bool)
}, tag event.Tag,
) {
	const pointerKinds = pointer.Press | pointer.Release | pointer.Move |
		pointer.Drag | pointer.Enter | pointer.Leave | pointer.Scroll

	for {
		ev, ok := src.Event(
			pointer.Filter{
				Target:  tag,
				Kinds:   pointerKinds,
				ScrollY: pointer.ScrollRange{Min: -1e6, Max: 1e6},
				ScrollX: pointer.ScrollRange{Min: -1e6, Max: 1e6},
			},
			key.Filter{Focus: tag, Optional: key.ModCtrl | key.ModShift | key.ModAlt | key.ModSuper | key.ModCommand},
		)
		if !ok {
			return
		}
		switch ev := ev.(type) {
		case pointer.Event:
			b.dispatchPointer(ev)
		case key.Event:
			b.dispatchKey(ev)
		}
	}
}

// dispatchPointer maps a Gio pointer.Event to the corresponding canvas
// event and emits it through the dispatcher.
func (b *Backend) dispatchPointer(ev pointer.Event) {
	pt := geom.Pt{X: float64(ev.Position.X), Y: float64(ev.Position.Y)}
	mods := mapModifiers(ev.Modifiers)
	btn := mapButtons(ev.Buttons)

	switch ev.Kind {
	case pointer.Press:
		evt := canvas.NewMouseEvent(canvas.EventMousePress, b.opts.Figure, pt, btn)
		evt.Modifiers = mods
		b.cnv.emit(evt.Event)
		canvas.EmitPick(b.cnv.dispatcher, b.opts.Figure, evt)
	case pointer.Release:
		evt := canvas.NewMouseEvent(canvas.EventMouseRelease, b.opts.Figure, pt, btn)
		evt.Modifiers = mods
		b.cnv.emit(evt.Event)
	case pointer.Move, pointer.Drag:
		evt := canvas.NewMouseEvent(canvas.EventMouseMove, b.opts.Figure, pt, btn)
		evt.Modifiers = mods
		b.cnv.emit(evt.Event)
		b.cnv.hover.Update(evt.Event)
	case pointer.Enter:
		evt := canvas.NewMouseEvent(canvas.EventFigureEnter, b.opts.Figure, pt, btn)
		evt.Modifiers = mods
		b.cnv.emit(evt.Event)
		b.cnv.hover.Update(evt.Event)
	case pointer.Leave:
		evt := canvas.NewMouseEvent(canvas.EventFigureLeave, b.opts.Figure, pt, btn)
		evt.Modifiers = mods
		b.cnv.emit(evt.Event)
		b.cnv.hover.Update(evt.Event)
	case pointer.Scroll:
		evt := canvas.NewMouseEvent(canvas.EventScroll, b.opts.Figure, pt, btn)
		evt.Modifiers = mods
		evt.DeltaX = float64(ev.Scroll.X)
		evt.DeltaY = float64(ev.Scroll.Y)
		b.cnv.emit(evt.Event)
	}
}

// dispatchKey forwards a Gio key.Event as a canvas KeyEvent.
func (b *Backend) dispatchKey(ev key.Event) {
	mods := mapModifiers(ev.Modifiers)
	switch ev.State {
	case key.Press:
		out := canvas.NewKeyEvent(canvas.EventKeyPress, b.opts.Figure, string(ev.Name), mods)
		b.cnv.emit(out.Event)
	case key.Release:
		out := canvas.NewKeyEvent(canvas.EventKeyRelease, b.opts.Figure, string(ev.Name), mods)
		b.cnv.emit(out.Event)
	}
}

// mapButtons projects a Gio button bitset onto a single canvas
// MouseButton. Matplotlib's MouseEvent.button is a single-button slot;
// the primary button wins when several are held simultaneously.
func mapButtons(b pointer.Buttons) canvas.MouseButton {
	switch {
	case b&pointer.ButtonPrimary != 0:
		return canvas.MouseButtonLeft
	case b&pointer.ButtonSecondary != 0:
		return canvas.MouseButtonRight
	case b&pointer.ButtonTertiary != 0:
		return canvas.MouseButtonMiddle
	default:
		return canvas.MouseButtonNone
	}
}

// mapModifiers projects Gio key modifiers to the canvas bitset.
func mapModifiers(m key.Modifiers) canvas.Modifier {
	return canvas.ModifierSet(
		m&key.ModShift != 0,
		m&key.ModCtrl != 0,
		m&key.ModAlt != 0,
		m&(key.ModSuper|key.ModCommand) != 0,
	)
}

// colorNRGBA projects a render.Color into the color.NRGBA expected by
// paint.Fill.
func colorNRGBA(c render.Color) color.NRGBA {
	return color.NRGBA{
		R: uint8(clamp01(c.R) * 255),
		G: uint8(clamp01(c.G) * 255),
		B: uint8(clamp01(c.B) * 255),
		A: uint8(clamp01(c.A) * 255),
	}
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// gioCanvas adapts a *core.Figure to the canvas.DrawIdleCanvas contract
// while sharing dispatcher state with the Gio window.
type gioCanvas struct {
	opts desktop.Options
	w    *gioapp.Window

	dispatcher *canvas.Dispatcher
	hover      *canvas.AxesHoverTracker

	mu     sync.Mutex
	width  int
	height int
	img    *image.RGBA
	dirty  bool
	rerr   error
}

func newGioCanvas(opts desktop.Options, w *gioapp.Window) *gioCanvas {
	dispatcher := &canvas.Dispatcher{}
	c := &gioCanvas{
		opts:       opts,
		w:          w,
		dispatcher: dispatcher,
		width:      opts.Width,
		height:     opts.Height,
		dirty:      true,
	}
	c.hover = canvas.NewAxesHoverTracker(opts.Figure, dispatcher)
	return c
}

// Figure returns the rendered figure.
func (c *gioCanvas) Figure() *core.Figure { return c.opts.Figure }

// Draw rasterizes the figure synchronously and emits a DrawEvent.
func (c *gioCanvas) Draw() error {
	if _, err := c.ensureImage(); err != nil {
		return err
	}
	w, h := c.size()
	return c.emitDraw(w, h)
}

// DrawIdle marks the canvas dirty and requests a new frame from Gio.
// The actual rasterization happens during the next FrameEvent.
func (c *gioCanvas) DrawIdle() error {
	c.markDirty()
	if c.w != nil {
		c.w.Invalidate()
	}
	return nil
}

// Resize re-sizes the rendered figure. Called by the Gio loop when a
// FrameEvent reports new dimensions.
func (c *gioCanvas) Resize(w, h int) error { return c.resize(w, h) }

// Connect registers an event handler on the shared dispatcher.
func (c *gioCanvas) Connect(t canvas.EventType, h canvas.Handler) canvas.ConnectionID {
	return c.dispatcher.Connect(t, h)
}

// Disconnect removes a previously-registered handler.
func (c *gioCanvas) Disconnect(id canvas.ConnectionID) { c.dispatcher.Disconnect(id) }

// Close emits a CloseEvent. Window destruction is owned by the Backend.
func (c *gioCanvas) Close() error { return c.emitClose() }

// Width returns the cached pixel width.
func (c *gioCanvas) Width() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.width
}

// Height returns the cached pixel height.
func (c *gioCanvas) Height() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.height
}

func (c *gioCanvas) size() (int, int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.width, c.height
}

func (c *gioCanvas) markDirty() {
	c.mu.Lock()
	c.dirty = true
	c.mu.Unlock()
}

func (c *gioCanvas) resize(w, h int) error {
	if w <= 0 || h <= 0 {
		return errors.New("gio: invalid resize dimensions")
	}
	c.mu.Lock()
	c.width, c.height = w, h
	c.img = nil
	c.dirty = true
	c.mu.Unlock()
	return c.dispatcher.Emit(canvas.NewResizeEvent(c.opts.Figure, w, h).Event)
}

// ensureImage renders the figure into c.img if dirty, then returns a
// pointer to the bitmap. The returned image is shared with the caller;
// the next call to ensureImage may overwrite it, but Gio uploads happen
// synchronously within the frame, so this is safe in practice.
func (c *gioCanvas) ensureImage() (*image.RGBA, error) {
	c.mu.Lock()
	if !c.dirty && c.img != nil {
		img := c.img
		c.mu.Unlock()
		return img, nil
	}
	w, h := c.width, c.height
	c.mu.Unlock()

	if w <= 0 || h <= 0 {
		return nil, nil
	}

	r, err := c.opts.Renderer(w, h, c.opts.Background)
	if err != nil {
		c.mu.Lock()
		c.rerr = err
		c.mu.Unlock()
		return nil, err
	}

	// Update figure pixel size so the figure tree lays out into the
	// current window before the draw call.
	c.opts.Figure.SizePx = geom.Pt{X: float64(w), Y: float64(h)}
	core.DrawFigure(c.opts.Figure, r)

	img, err := extractRGBA(r, w, h)
	if err != nil {
		c.mu.Lock()
		c.rerr = err
		c.mu.Unlock()
		return nil, err
	}

	c.mu.Lock()
	c.img = img
	c.dirty = false
	c.rerr = nil
	c.mu.Unlock()
	return img, nil
}

// extractRGBA pulls the rasterized image out of a renderer. Renderers
// that implement render.RGBAExporter (the AGG backend does) provide the
// most efficient path. We do not depend on the AGG package directly so
// this adapter stays toolkit-clean.
func extractRGBA(r render.Renderer, w, h int) (*image.RGBA, error) {
	type rgbaExporter interface {
		GetImage() *image.RGBA
	}
	if ex, ok := r.(rgbaExporter); ok {
		img := ex.GetImage()
		if img != nil {
			return img, nil
		}
	}
	// Fallback for renderers without GetImage: return an empty bitmap
	// sized to the window. Callers see a blank canvas instead of a crash.
	return image.NewRGBA(image.Rect(0, 0, w, h)), nil
}

// emit dispatches an arbitrary event through the canvas dispatcher.
func (c *gioCanvas) emit(ev canvas.Event) {
	_ = c.dispatcher.Emit(ev)
}

func (c *gioCanvas) emitDraw(w, h int) error {
	return c.dispatcher.Emit(canvas.NewDrawEvent(c.opts.Figure, w, h).Event)
}

func (c *gioCanvas) emitClose() error {
	return c.dispatcher.Emit(canvas.NewCloseEvent(c.opts.Figure).Event)
}

// Compile-time checks.
var (
	_ canvas.DrawIdleCanvas = (*gioCanvas)(nil)
	_ desktop.Backend       = (*Backend)(nil)
)
