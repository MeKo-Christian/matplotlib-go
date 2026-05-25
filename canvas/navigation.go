package canvas

import (
	"errors"
	"math"
	"sync"

	"github.com/cwbudde/matplotlib-go/internal/geom"
)

// NavigationMode selects which interaction the [Navigation] helper performs in
// response to mouse drags.
type NavigationMode uint8

const (
	// NavNone keeps the navigation helper passive — clicks pass through.
	NavNone NavigationMode = iota
	// NavPan drags the view limits along with the cursor; mirrors
	// NavigationToolbar2.pan.
	NavPan
	// NavZoom rubber-bands a rectangle on drag and zooms the axes to its
	// extents on release; mirrors NavigationToolbar2.zoom.
	NavZoom
)

// Navigation wires the standard pan / zoom-to-rect / scroll-zoom tools onto a
// dispatcher. The helper is backend agnostic: it mutates axes view limits and
// then requests a redraw through its draw callback. Backends typically pass
// their [DrawIdleCanvas.DrawIdle] method as the draw callback.
//
// The zero value is unusable — call [NewNavigation].
type Navigation struct {
	mu       sync.Mutex
	mode     NavigationMode
	dragging bool
	dragAxes *Axes
	dragFrom geom.Pt
	dragHome navHomeLimits
	dragLast geom.Pt
	rect     geom.Rect
	hasRect  bool

	figure   *Figure
	draw     func() error
	scrollK  float64
	connects []ConnectionID
	dispatch *Dispatcher
}

type navHomeLimits struct {
	xMin, xMax float64
	yMin, yMax float64
}

// NewNavigation creates a Navigation helper bound to the given figure. The
// draw callback is invoked after view limits change; pass nil to skip
// redrawing.
func NewNavigation(fig *Figure, draw func() error) *Navigation {
	return &Navigation{
		figure:  fig,
		draw:    draw,
		scrollK: 1.2,
	}
}

// SetScrollFactor overrides the scroll-zoom step (default 1.2). A click of the
// wheel multiplies / divides axis extents by this value; values < 1 are
// rejected.
func (n *Navigation) SetScrollFactor(factor float64) error {
	if n == nil {
		return errors.New("canvas: nil navigation")
	}
	if factor <= 1.0 {
		return errors.New("canvas: scroll factor must be > 1.0")
	}
	n.mu.Lock()
	n.scrollK = factor
	n.mu.Unlock()
	return nil
}

// SetMode changes the active interaction mode. A NavZoom→NavPan switch (or
// either→NavNone) drops any in-progress drag without applying it.
func (n *Navigation) SetMode(mode NavigationMode) {
	if n == nil {
		return
	}
	n.mu.Lock()
	if n.mode != mode {
		n.dragging = false
		n.hasRect = false
		n.dragAxes = nil
	}
	n.mode = mode
	n.mu.Unlock()
}

// Mode returns the active interaction mode.
func (n *Navigation) Mode() NavigationMode {
	if n == nil {
		return NavNone
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.mode
}

// DragRect returns the in-progress zoom-rectangle in figure pixels, if any.
// Backends can use this to draw a feedback overlay on the canvas.
func (n *Navigation) DragRect() (geom.Rect, bool) {
	if n == nil {
		return geom.Rect{}, false
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	if !n.hasRect {
		return geom.Rect{}, false
	}
	return n.rect, true
}

// Attach subscribes the helper to the dispatcher's mouse and scroll events.
// Subsequent calls replace the prior subscription.
func (n *Navigation) Attach(dispatcher *Dispatcher) {
	if n == nil || dispatcher == nil {
		return
	}
	n.Detach()
	n.mu.Lock()
	n.dispatch = dispatcher
	n.connects = append(
		n.connects,
		dispatcher.Connect(EventMousePress, func(ev Event) error {
			n.handlePress(MouseEvent{Event: ev})
			return nil
		}),
		dispatcher.Connect(EventMouseMove, func(ev Event) error {
			n.handleMove(MouseEvent{Event: ev})
			return nil
		}),
		dispatcher.Connect(EventMouseRelease, func(ev Event) error {
			return n.handleRelease(MouseEvent{Event: ev})
		}),
		dispatcher.Connect(EventScroll, func(ev Event) error {
			return n.handleScroll(MouseEvent{Event: ev})
		}),
	)
	n.mu.Unlock()
}

// Detach removes any previously installed handlers.
func (n *Navigation) Detach() {
	if n == nil {
		return
	}
	n.mu.Lock()
	dispatcher := n.dispatch
	ids := n.connects
	n.connects = nil
	n.dispatch = nil
	n.mu.Unlock()
	if dispatcher == nil {
		return
	}
	for _, id := range ids {
		dispatcher.Disconnect(id)
	}
}

func (n *Navigation) handlePress(mouse MouseEvent) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.mode == NavNone {
		return
	}
	ax := mouse.Axes
	if ax == nil {
		ax, _, _ = ResolveEventTarget(n.figure, mouse.Position)
	}
	if ax == nil {
		return
	}
	n.dragging = true
	n.dragAxes = ax
	n.dragFrom = mouse.Position
	n.dragLast = mouse.Position
	n.dragHome = snapshotLimits(ax)
	n.rect = geom.Rect{Min: mouse.Position, Max: mouse.Position}
	n.hasRect = n.mode == NavZoom
}

func (n *Navigation) handleMove(mouse MouseEvent) {
	n.mu.Lock()
	if !n.dragging || n.dragAxes == nil {
		n.mu.Unlock()
		return
	}
	switch n.mode {
	case NavPan:
		ax := n.dragAxes
		from := n.dragLast
		n.dragLast = mouse.Position
		n.mu.Unlock()
		PanAxesPixels(ax, mouse.Position.X-from.X, mouse.Position.Y-from.Y)
		n.callDraw()
		return
	case NavZoom:
		n.rect = normalizeDragRect(n.dragFrom, mouse.Position)
		n.hasRect = true
	}
	n.mu.Unlock()
}

func (n *Navigation) handleRelease(mouse MouseEvent) error {
	n.mu.Lock()
	if !n.dragging {
		n.mu.Unlock()
		return nil
	}
	ax := n.dragAxes
	mode := n.mode
	rect := normalizeDragRect(n.dragFrom, mouse.Position)
	n.dragging = false
	n.dragAxes = nil
	n.hasRect = false
	n.mu.Unlock()

	switch mode {
	case NavPan:
		return n.callDraw()
	case NavZoom:
		if rect.W() < 1 || rect.H() < 1 {
			return nil
		}
		if err := ZoomAxesToRect(ax, rect); err != nil {
			return err
		}
		return n.callDraw()
	}
	return nil
}

func (n *Navigation) handleScroll(mouse MouseEvent) error {
	n.mu.Lock()
	scrollK := n.scrollK
	n.mu.Unlock()
	ax := mouse.Axes
	if ax == nil {
		ax, _, _ = ResolveEventTarget(n.figure, mouse.Position)
	}
	if ax == nil {
		return nil
	}
	step := mouse.DeltaY
	if step == 0 {
		step = mouse.DeltaX
	}
	if step == 0 {
		return nil
	}
	factor := scrollK
	if step > 0 {
		// scroll up zooms in (shrinks data extents).
		factor = 1 / scrollK
	}
	if err := ZoomAxesByFactor(ax, mouse.Position, factor); err != nil {
		return err
	}
	return n.callDraw()
}

func (n *Navigation) callDraw() error {
	n.mu.Lock()
	draw := n.draw
	n.mu.Unlock()
	if draw == nil {
		return nil
	}
	return draw()
}

func snapshotLimits(ax *Axes) navHomeLimits {
	if ax == nil {
		return navHomeLimits{}
	}
	state := navHomeLimits{}
	if ax.XScale != nil {
		state.xMin, state.xMax = ax.XScale.Domain()
	}
	if ax.YScale != nil {
		state.yMin, state.yMax = ax.YScale.Domain()
	}
	return state
}

func normalizeDragRect(a, b geom.Pt) geom.Rect {
	minX, maxX := a.X, b.X
	if maxX < minX {
		minX, maxX = maxX, minX
	}
	minY, maxY := a.Y, b.Y
	if maxY < minY {
		minY, maxY = maxY, minY
	}
	return geom.Rect{Min: geom.Pt{X: minX, Y: minY}, Max: geom.Pt{X: maxX, Y: maxY}}
}

// PanAxesPixels translates the axes view limits by (dxPx, dyPx) figure pixels.
// Positive dx moves data rightward; positive dy moves data downward (figure
// coordinates).
func PanAxesPixels(ax *Axes, dxPx, dyPx float64) {
	if ax == nil {
		return
	}
	clip := ax.DisplayRect()
	if clip.W() == 0 || clip.H() == 0 {
		return
	}
	xMin, xMax, okX := unitAxisLimits(ax, true)
	yMin, yMax, okY := unitAxisLimits(ax, false)
	if !okX || !okY {
		return
	}
	dx := (xMax - xMin) * (-dxPx / clip.W())
	// In figure pixels Y grows downward, but data Y typically grows upward.
	dy := (yMax - yMin) * (dyPx / clip.H())
	ax.SetXLim(xMin+dx, xMax+dx)
	ax.SetYLim(yMin+dy, yMax+dy)
}

// ZoomAxesToRect sets the axes' view limits to the data range that maps to
// the given figure-pixel rectangle.
func ZoomAxesToRect(ax *Axes, rectPx geom.Rect) error {
	if ax == nil {
		return errors.New("canvas: nil axes")
	}
	d1, ok1 := ax.PixelToData(rectPx.Min)
	d2, ok2 := ax.PixelToData(rectPx.Max)
	if !ok1 || !ok2 {
		return errors.New("canvas: axes rejected pixel inversion")
	}
	xMin, xMax := minMax(d1.X, d2.X)
	yMin, yMax := minMax(d1.Y, d2.Y)
	if xMin == xMax || yMin == yMax {
		return errors.New("canvas: degenerate zoom rectangle")
	}
	ax.SetXLim(xMin, xMax)
	ax.SetYLim(yMin, yMax)
	return nil
}

// ZoomAxesByFactor scales the axes' extents by factor, holding the data point
// under anchorPx fixed.
func ZoomAxesByFactor(ax *Axes, anchorPx geom.Pt, factor float64) error {
	if ax == nil {
		return errors.New("canvas: nil axes")
	}
	if factor <= 0 || math.IsNaN(factor) || math.IsInf(factor, 0) {
		return errors.New("canvas: invalid zoom factor")
	}
	anchor, ok := ax.PixelToData(anchorPx)
	if !ok {
		return errors.New("canvas: axes rejected pixel inversion")
	}
	xMin, xMax, okX := unitAxisLimits(ax, true)
	yMin, yMax, okY := unitAxisLimits(ax, false)
	if !okX || !okY {
		return errors.New("canvas: axes has no usable limits")
	}
	newX0 := anchor.X - (anchor.X-xMin)*factor
	newX1 := anchor.X + (xMax-anchor.X)*factor
	newY0 := anchor.Y - (anchor.Y-yMin)*factor
	newY1 := anchor.Y + (yMax-anchor.Y)*factor
	ax.SetXLim(newX0, newX1)
	ax.SetYLim(newY0, newY1)
	return nil
}

func unitAxisLimits(ax *Axes, isX bool) (float64, float64, bool) {
	if ax == nil {
		return 0, 0, false
	}
	if isX {
		if ax.XScale == nil {
			return 0, 0, false
		}
		minVal, maxVal := ax.XScale.Domain()
		return minVal, maxVal, true
	}
	if ax.YScale == nil {
		return 0, 0, false
	}
	minVal, maxVal := ax.YScale.Domain()
	return minVal, maxVal, true
}

func minMax(a, b float64) (float64, float64) {
	if a < b {
		return a, b
	}
	return b, a
}
