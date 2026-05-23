package webagg

import (
	"bytes"
	"errors"
	"image"
	"image/png"
	"sync"
	"sync/atomic"
	"time"

	plotcanvas "github.com/cwbudde/matplotlib-go/canvas"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// RasterRenderer is the renderer contract WebAgg needs: it must paint
// into a RGBA backing buffer that we can either send whole or compare
// pixel-by-pixel against the prior frame.
type RasterRenderer interface {
	render.Renderer
	GetImage() *image.RGBA
}

// RendererFactory builds a fresh RasterRenderer at the given pixel size
// and clear color. It is invoked once per resize (the renderer is
// re-used across draws at a fixed size).
type RendererFactory func(width, height int, background render.Color) (RasterRenderer, error)

// Manager wraps a figure with the state needed to drive an interactive
// web session: a renderer, a per-session PNG diff buffer, a navigation
// helper, a toolbar controller, an event dispatcher, and a fan-out hub
// for connected WebSocket clients.
//
// Manager implements [canvas.FigureManager] and [canvas.FigureCanvas].
// The same Manager can be shared by multiple WebSocket clients; each
// new client starts with a full frame, then receives diff frames as
// the figure mutates.
type Manager struct {
	figure   *core.Figure
	factory  RendererFactory
	bg       render.Color
	nav      *plotcanvas.Navigation
	tb       *plotcanvas.ToolbarController
	dispatch *plotcanvas.Dispatcher
	hover    *plotcanvas.AxesHoverTracker

	hub *hub

	mu          sync.Mutex
	renderer    RasterRenderer
	rendererWH  geom.Pt
	lastBuf     []uint32 // packed prior frame, row-major, len = w*h
	lastW       int
	lastH       int
	imageMode   imageMode
	forceFull   bool
	pngStale    bool
	devPxRatio  float64
	lastMouseX  float64
	lastMouseY  float64
	eventLoop   plotcanvas.EventLoop
	drawQueued  bool
	staleRegs   []staleRegistration
	closed      atomic.Bool
	windowTitle string

	// home view limits — captured at the first paint so the toolbar
	// home button can restore them.
	home figureHome
}

// figureHome is the snapshot the home toolbar action restores to.
type figureHome struct {
	width  int
	height int
	axes   []axesHome
}

type axesHome struct {
	axes                   *core.Axes
	xMin, xMax, yMin, yMax float64
}

// Options bundles the inputs NewManager needs.
type Options struct {
	// Figure is the figure this manager paints. Required.
	Figure *core.Figure
	// Renderer builds an AGG-compatible raster renderer. Required.
	Renderer RendererFactory
	// Background is the clear color passed to the renderer factory. If
	// the zero value, the figure's RC.FigureBackground() is used.
	Background render.Color
	// HasBackground distinguishes the zero render.Color (transparent
	// black) from "use the figure's background". Set to true when
	// Background is meaningful.
	HasBackground bool
	// SaveHandler, if non-nil, is invoked when the client clicks the
	// toolbar Save button. The handler is free to write the PNG
	// anywhere — to a file, an HTTP response, etc.
	SaveHandler plotcanvas.ToolbarHandler
	// EventLoop schedules draw_idle callbacks. If nil, WebAgg uses a
	// small goroutine-backed loop that coalesces same-turn requests.
	EventLoop plotcanvas.EventLoop
}

// NewManager builds a manager from the options. The figure's pixel
// size is treated as the initial canvas size.
func NewManager(opts Options) (*Manager, error) {
	if opts.Figure == nil {
		return nil, errors.New("webagg: nil figure")
	}
	if opts.Renderer == nil {
		return nil, errors.New("webagg: nil renderer factory")
	}

	bg := opts.Background
	if !opts.HasBackground {
		bg = opts.Figure.RC.FigureBackground()
	}

	m := &Manager{
		figure:      opts.Figure,
		factory:     opts.Renderer,
		bg:          bg,
		dispatch:    &plotcanvas.Dispatcher{},
		imageMode:   imageModeFull,
		forceFull:   true,
		pngStale:    true,
		devPxRatio:  1,
		eventLoop:   opts.EventLoop,
		windowTitle: "matplotlib-go",
	}
	if m.eventLoop == nil {
		m.eventLoop = webEventLoop{}
	}
	m.hub = newHub(m)
	m.nav = plotcanvas.NewNavigation(opts.Figure, m.DrawIdle)
	m.nav.Attach(m.dispatch)
	m.hover = plotcanvas.NewAxesHoverTracker(opts.Figure, m.dispatch)
	m.tb = plotcanvas.NewToolbarController(m.nav)
	if opts.SaveHandler != nil {
		m.tb.SetHandler(plotcanvas.ToolbarSave, opts.SaveHandler)
	}
	m.tb.SetHandler(plotcanvas.ToolbarHome, m.home_)
	m.home = snapshotHome(opts.Figure)
	m.attachStaleCallbacks()
	return m, nil
}

// Figure returns the wrapped figure.
func (m *Manager) Figure() *core.Figure { return m.figure }

// Navigation returns the navigation helper. Tests and embedders may
// drive pan/zoom programmatically through it.
func (m *Manager) Navigation() *plotcanvas.Navigation { return m.nav }

// Toolbar returns the toolbar controller; clients trigger actions on
// it via the `toolbar_button` JSON event.
func (m *Manager) Toolbar() *plotcanvas.ToolbarController { return m.tb }

// Dispatcher returns the event dispatcher. Tests typically Emit
// directly to simulate a client without a WebSocket.
func (m *Manager) Dispatcher() *plotcanvas.Dispatcher { return m.dispatch }

// Canvas returns the canvas view of the manager. The manager itself is
// the canvas — there is one figure per manager.
func (m *Manager) Canvas() plotcanvas.FigureCanvas { return m }

// Show is a no-op for web; the figure is "shown" when a client connects.
func (m *Manager) Show() error { return nil }

// SetTitle stores the title that newly-connecting clients should set on
// their window. The change is broadcast as a `figure_label` event.
func (m *Manager) SetTitle(title string) {
	m.mu.Lock()
	m.windowTitle = title
	m.mu.Unlock()
	m.hub.broadcastJSON(&outboundEvent{Type: MsgFigureLabel, Label: title})
}

// ToolManager satisfies canvas.FigureManager. WebAgg does not expose a
// separate tool manager; toolbar actions flow through the dispatcher
// instead.
func (m *Manager) ToolManager() *plotcanvas.ToolManager { return plotcanvas.NewToolManager() }

// Close shuts down all WebSocket clients and emits the close lifecycle
// event.
func (m *Manager) Close() error {
	if !m.closed.CompareAndSwap(false, true) {
		return nil
	}
	m.detachStaleCallbacks()
	m.hub.closeAll()
	return m.dispatch.Emit(plotcanvas.Event{
		Type:   plotcanvas.EventClose,
		Figure: m.figure,
		Width:  int(m.figure.SizePx.X + 0.5),
		Height: int(m.figure.SizePx.Y + 0.5),
	})
}

// Connect subscribes a handler to a canvas event type.
func (m *Manager) Connect(eventType plotcanvas.EventType, handler plotcanvas.Handler) plotcanvas.ConnectionID {
	return m.dispatch.Connect(eventType, handler)
}

// Disconnect removes a previously-installed handler.
func (m *Manager) Disconnect(id plotcanvas.ConnectionID) { m.dispatch.Disconnect(id) }

// Draw repaints the figure and broadcasts the resulting PNG diff (or
// full frame) to every connected client.
func (m *Manager) Draw() error {
	return m.drawAndBroadcast()
}

// DrawIdle schedules a redraw for the next idle turn. Repeated calls before the
// idle callback runs collapse into one draw, matching Matplotlib's draw_idle
// behavior for event storms.
func (m *Manager) DrawIdle() error {
	if m.closed.Load() {
		return nil
	}
	m.mu.Lock()
	if m.drawQueued {
		m.mu.Unlock()
		return nil
	}
	m.drawQueued = true
	loop := m.eventLoop
	m.mu.Unlock()
	if loop == nil {
		loop = webEventLoop{}
	}
	if err := loop.CallSoon(m.flushDrawIdle); err != nil {
		m.mu.Lock()
		m.drawQueued = false
		m.mu.Unlock()
		return err
	}
	return nil
}

// CopyFromBBox captures a renderer pixel region for blitting. It returns nil
// until the first draw has created a renderer, or when the renderer does not
// implement render.BufferRegioner.
func (m *Manager) CopyFromBBox(bbox geom.Rect) *render.BufferRegion {
	m.mu.Lock()
	defer m.mu.Unlock()
	regioner, ok := m.renderer.(render.BufferRegioner)
	if !ok {
		return nil
	}
	return regioner.CopyFromBBox(bbox)
}

// RestoreRegion restores a previously captured renderer pixel region. Call
// Blit after restoring/drawing animated content to push the damaged pixels to
// clients.
func (m *Manager) RestoreRegion(region *render.BufferRegion, bbox *geom.Rect, offset geom.Pt) error {
	if m.closed.Load() {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	regioner, ok := m.renderer.(render.BufferRegioner)
	if !ok {
		return errors.New("webagg: renderer does not support buffer regions")
	}
	regioner.RestoreRegion(region, bbox, offset)
	m.pngStale = true
	return nil
}

// Blit broadcasts the current renderer buffer without redrawing the full
// figure. The bbox bounds the damaged pixels for diff generation.
func (m *Manager) Blit(bbox geom.Rect) error {
	if m.closed.Load() {
		return nil
	}
	m.mu.Lock()
	m.pngStale = true
	m.mu.Unlock()
	return m.hub.broadcastBlitFrame(&bbox)
}

func (m *Manager) flushDrawIdle() error {
	m.mu.Lock()
	if !m.drawQueued {
		m.mu.Unlock()
		return nil
	}
	m.drawQueued = false
	m.mu.Unlock()
	return m.drawAndBroadcast()
}

// Resize updates the figure pixel size, forces the next paint to be a
// full frame, and re-broadcasts.
func (m *Manager) Resize(width, height int) error {
	if width <= 0 || height <= 0 {
		return errors.New("webagg: resize dimensions must be positive")
	}
	m.mu.Lock()
	m.figure.SizePx = geom.Pt{X: float64(width), Y: float64(height)}
	m.forceFull = true
	m.pngStale = true
	m.renderer = nil
	m.mu.Unlock()

	if err := m.dispatch.Emit(plotcanvas.Event{
		Type:   plotcanvas.EventResize,
		Figure: m.figure,
		Width:  width,
		Height: height,
	}); err != nil {
		return err
	}
	// Echo the new size back to clients so they can re-stretch their
	// canvas element.
	forwardFalse := false
	m.hub.broadcastJSON(&outboundEvent{
		Type:      MsgResize,
		Size:      []float64{float64(width), float64(height)},
		ResizeFwd: &forwardFalse,
	})
	return m.drawAndBroadcast()
}

// drawAndBroadcast repaints (computing a fresh diff if anything has
// changed) and pushes the resulting PNG to every client.
func (m *Manager) drawAndBroadcast() error {
	if m.closed.Load() {
		return nil
	}
	m.mu.Lock()
	m.pngStale = true
	m.mu.Unlock()
	if err := m.dispatch.Emit(plotcanvas.Event{
		Type:   plotcanvas.EventDraw,
		Figure: m.figure,
		Width:  int(m.figure.SizePx.X + 0.5),
		Height: int(m.figure.SizePx.Y + 0.5),
	}); err != nil {
		return err
	}
	return m.hub.broadcastFrame()
}

// renderFrame paints the figure into the manager's renderer and returns
// the current image-mode plus a PNG payload that the client should
// apply.
func (m *Manager) renderFrame() (imageMode, []byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	w := int(m.figure.SizePx.X + 0.5)
	h := int(m.figure.SizePx.Y + 0.5)
	if w <= 0 || h <= 0 {
		return imageModeFull, nil, errors.New("webagg: figure has zero pixel size")
	}

	if m.renderer == nil || m.rendererWH.X != float64(w) || m.rendererWH.Y != float64(h) {
		r, err := m.factory(w, h, m.bg)
		if err != nil {
			return imageModeFull, nil, err
		}
		m.renderer = r
		m.rendererWH = geom.Pt{X: float64(w), Y: float64(h)}
		m.forceFull = true
	}

	core.DrawFigure(m.figure, m.renderer)
	clearStaleArtists(m.figure)
	img := m.renderer.GetImage()
	if img == nil {
		return imageModeFull, nil, errors.New("webagg: renderer returned no image")
	}
	return m.encodeFrameLocked(img, w, h, nil)
}

func (m *Manager) renderCurrentFrame(damage *geom.Rect) (imageMode, []byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	w := int(m.figure.SizePx.X + 0.5)
	h := int(m.figure.SizePx.Y + 0.5)
	if w <= 0 || h <= 0 {
		return imageModeFull, nil, errors.New("webagg: figure has zero pixel size")
	}
	if m.renderer == nil {
		return imageModeFull, nil, errors.New("webagg: no renderer to blit")
	}
	img := m.renderer.GetImage()
	if img == nil {
		return imageModeFull, nil, errors.New("webagg: renderer returned no image")
	}
	return m.encodeFrameLocked(img, w, h, damage)
}

func (m *Manager) encodeFrameLocked(img *image.RGBA, w, h int, damage *geom.Rect) (imageMode, []byte, error) {
	// Pack RGBA pixels into uint32 row-major buffer for fast diff.
	packed := packRGBA(img, w, h)
	mode := imageModeDiff
	hasTransparency := imageHasTransparency(img)
	if m.forceFull || hasTransparency || m.lastW != w || m.lastH != h || len(m.lastBuf) != len(packed) {
		mode = imageModeFull
	}

	var payload *image.RGBA
	if mode == imageModeFull {
		payload = img
		damage = nil
	} else if damage != nil {
		payload = diffImageInRect(img, m.lastBuf, packed, w, h, *damage)
	} else {
		payload = diffImage(img, m.lastBuf, packed, w, h)
	}

	// Persist the current frame for the next diff.
	if cap(m.lastBuf) >= len(packed) {
		m.lastBuf = m.lastBuf[:len(packed)]
	} else {
		m.lastBuf = make([]uint32, len(packed))
	}
	if damage == nil {
		copy(m.lastBuf, packed)
	} else {
		copyDamage(m.lastBuf, packed, w, h, *damage)
	}
	m.lastW = w
	m.lastH = h
	m.forceFull = false
	m.pngStale = false
	m.imageMode = mode

	var buf bytes.Buffer
	if err := png.Encode(&buf, payload); err != nil {
		return imageModeFull, nil, err
	}
	return mode, buf.Bytes(), nil
}

// home_ restores axes view limits to the snapshot captured at attach
// time. The trailing underscore avoids the name collision with the
// `home` field.
func (m *Manager) home_() error {
	for _, ax := range m.home.axes {
		if ax.axes == nil {
			continue
		}
		ax.axes.SetXLim(ax.xMin, ax.xMax)
		ax.axes.SetYLim(ax.yMin, ax.yMax)
	}
	return m.DrawIdle()
}

type webEventLoop struct{}

func (webEventLoop) CallSoon(callback func() error) error {
	if callback == nil {
		return nil
	}
	time.AfterFunc(time.Millisecond, func() {
		_ = callback()
	})
	return nil
}

func (webEventLoop) NewTimer(interval time.Duration, callback func() error) plotcanvas.Timer {
	return plotcanvas.NewTimer(interval, callback)
}

// snapshotHome captures the figure's initial size and per-axes view
// limits so the toolbar home button can restore them.
func snapshotHome(fig *core.Figure) figureHome {
	if fig == nil {
		return figureHome{}
	}
	state := figureHome{
		width:  int(fig.SizePx.X + 0.5),
		height: int(fig.SizePx.Y + 0.5),
		axes:   make([]axesHome, 0, len(fig.Children)),
	}
	for _, ax := range fig.Children {
		if ax == nil {
			continue
		}
		x0, x1 := 0.0, 1.0
		if ax.XScale != nil {
			x0, x1 = ax.XScale.Domain()
		}
		y0, y1 := 0.0, 1.0
		if ax.YScale != nil {
			y0, y1 = ax.YScale.Domain()
		}
		state.axes = append(state.axes, axesHome{axes: ax, xMin: x0, xMax: x1, yMin: y0, yMax: y1})
	}
	return state
}

// packRGBA copies an image's RGBA pixels into a uint32 row-major buffer
// — one pixel per uint32 — so diff comparisons run on a single value.
func packRGBA(img *image.RGBA, w, h int) []uint32 {
	out := make([]uint32, w*h)
	stride := img.Stride
	base := img.Pix
	for y := range h {
		row := base[y*stride : y*stride+w*4]
		dst := out[y*w : y*w+w]
		for x := range w {
			i := x * 4
			dst[x] = uint32(row[i]) |
				uint32(row[i+1])<<8 |
				uint32(row[i+2])<<16 |
				uint32(row[i+3])<<24
		}
	}
	return out
}

// imageHasTransparency reports whether any pixel has alpha < 255. The
// diff path cannot represent transparent pixels (it uses A=0 as the
// "unchanged" marker), so transparent frames must be sent as full
// frames.
func imageHasTransparency(img *image.RGBA) bool {
	stride := img.Stride
	w := img.Rect.Dx()
	h := img.Rect.Dy()
	base := img.Pix
	for y := range h {
		row := base[y*stride : y*stride+w*4]
		for x := range w {
			if row[x*4+3] != 0xff {
				return true
			}
		}
	}
	return false
}

// diffImage builds a new RGBA image where pixels that changed from the
// prior frame carry the new value and unchanged pixels are encoded as
// fully transparent black. The client paints the diff over its prior
// frame, so unchanged pixels remain visible.
func diffImage(curr *image.RGBA, lastBuf, packed []uint32, w, h int) *image.RGBA {
	out := image.NewRGBA(image.Rect(0, 0, w, h))
	stride := curr.Stride
	src := curr.Pix
	dst := out.Pix
	for y := range h {
		srow := src[y*stride : y*stride+w*4]
		drow := dst[y*out.Stride : y*out.Stride+w*4]
		prow := packed[y*w : y*w+w]
		brow := lastBuf[y*w : y*w+w]
		for x := range w {
			if prow[x] == brow[x] {
				// Stays transparent black — i.e. "no change".
				continue
			}
			i := x * 4
			drow[i+0] = srow[i+0]
			drow[i+1] = srow[i+1]
			drow[i+2] = srow[i+2]
			drow[i+3] = srow[i+3]
		}
	}
	return out
}

func diffImageInRect(curr *image.RGBA, lastBuf, packed []uint32, w, h int, damage geom.Rect) *image.RGBA {
	out := image.NewRGBA(image.Rect(0, 0, w, h))
	minX, minY, maxX, maxY, ok := pixelBounds(damage, w, h)
	if !ok {
		return out
	}
	stride := curr.Stride
	src := curr.Pix
	dst := out.Pix
	for y := minY; y < maxY; y++ {
		srow := src[y*stride : y*stride+w*4]
		drow := dst[y*out.Stride : y*out.Stride+w*4]
		prow := packed[y*w : y*w+w]
		brow := lastBuf[y*w : y*w+w]
		for x := minX; x < maxX; x++ {
			if prow[x] == brow[x] {
				continue
			}
			i := x * 4
			drow[i+0] = srow[i+0]
			drow[i+1] = srow[i+1]
			drow[i+2] = srow[i+2]
			drow[i+3] = srow[i+3]
		}
	}
	return out
}

func copyDamage(dst, src []uint32, w, h int, damage geom.Rect) {
	minX, minY, maxX, maxY, ok := pixelBounds(damage, w, h)
	if !ok {
		return
	}
	for y := minY; y < maxY; y++ {
		copy(dst[y*w+minX:y*w+maxX], src[y*w+minX:y*w+maxX])
	}
}

func pixelBounds(rect geom.Rect, w, h int) (minX, minY, maxX, maxY int, ok bool) {
	minX = int(rect.Min.X)
	minY = int(rect.Min.Y)
	maxX = int(rect.Max.X + 0.999999)
	maxY = int(rect.Max.Y + 0.999999)
	if minX < 0 {
		minX = 0
	}
	if minY < 0 {
		minY = 0
	}
	if maxX > w {
		maxX = w
	}
	if maxY > h {
		maxY = h
	}
	return minX, minY, maxX, maxY, minX < maxX && minY < maxY
}

var (
	_ plotcanvas.DrawIdleCanvas = (*Manager)(nil)
	_ plotcanvas.BlitCanvas     = (*Manager)(nil)
)
