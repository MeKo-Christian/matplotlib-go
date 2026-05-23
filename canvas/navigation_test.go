package canvas

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
)

func newNavigationFigure(t *testing.T) (*Figure, *Axes) {
	t.Helper()
	fig := core.NewFigure(100, 100)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 10)
	return fig, ax
}

func axisLimits(t *testing.T, ax *Axes) (float64, float64, float64, float64) {
	t.Helper()
	xMin, xMax := ax.XScale.Domain()
	yMin, yMax := ax.YScale.Domain()
	return xMin, xMax, yMin, yMax
}

func almost(a, b float64) bool {
	return math.Abs(a-b) < 1e-6
}

func TestPanAxesPixels(t *testing.T) {
	_, ax := newNavigationFigure(t)
	// Move 10 pixels right and 0 vertically.
	PanAxesPixels(ax, 10, 0)
	xMin, xMax, yMin, yMax := axisLimits(t, ax)
	if !almost(xMin, -1) || !almost(xMax, 9) {
		t.Fatalf("xlim after pan = (%g, %g), want (-1, 9)", xMin, xMax)
	}
	if !almost(yMin, 0) || !almost(yMax, 10) {
		t.Fatalf("ylim after pan = (%g, %g), want (0, 10)", yMin, yMax)
	}
}

func TestZoomAxesToRect(t *testing.T) {
	fig, ax := newNavigationFigure(t)
	ctx := core.AxesDrawContext(ax, fig)
	a := (&ctx.DataToPixel).Apply(geom.Pt{X: 2, Y: 2})
	b := (&ctx.DataToPixel).Apply(geom.Pt{X: 8, Y: 8})
	rect := geom.Rect{
		Min: geom.Pt{X: math.Min(a.X, b.X), Y: math.Min(a.Y, b.Y)},
		Max: geom.Pt{X: math.Max(a.X, b.X), Y: math.Max(a.Y, b.Y)},
	}
	if err := ZoomAxesToRect(ax, rect); err != nil {
		t.Fatalf("ZoomAxesToRect: %v", err)
	}
	xMin, xMax, yMin, yMax := axisLimits(t, ax)
	if !almost(xMin, 2) || !almost(xMax, 8) || !almost(yMin, 2) || !almost(yMax, 8) {
		t.Fatalf("limits = (%g,%g,%g,%g), want (2,8,2,8)", xMin, xMax, yMin, yMax)
	}
}

func TestZoomAxesByFactor(t *testing.T) {
	fig, ax := newNavigationFigure(t)
	ctx := core.AxesDrawContext(ax, fig)
	anchor := (&ctx.DataToPixel).Apply(geom.Pt{X: 5, Y: 5})
	if err := ZoomAxesByFactor(ax, anchor, 0.5); err != nil {
		t.Fatalf("ZoomAxesByFactor: %v", err)
	}
	xMin, xMax, yMin, yMax := axisLimits(t, ax)
	if !almost(xMin, 2.5) || !almost(xMax, 7.5) {
		t.Fatalf("xlim = (%g, %g), want (2.5, 7.5)", xMin, xMax)
	}
	if !almost(yMin, 2.5) || !almost(yMax, 7.5) {
		t.Fatalf("ylim = (%g, %g), want (2.5, 7.5)", yMin, yMax)
	}
}

func TestNavigationPanDragRedraws(t *testing.T) {
	fig, ax := newNavigationFigure(t)
	drawCalls := 0
	nav := NewNavigation(fig, func() error { drawCalls++; return nil })
	nav.SetMode(NavPan)

	var dispatcher Dispatcher
	nav.Attach(&dispatcher)
	defer nav.Detach()

	start := geom.Pt{X: 50, Y: 50}
	_ = dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: ax, Position: start, Button: MouseButtonLeft})
	_ = dispatcher.Emit(Event{Type: EventMouseMove, Figure: fig, Axes: ax, Position: geom.Pt{X: 60, Y: 50}, Button: MouseButtonLeft})
	_ = dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: ax, Position: geom.Pt{X: 60, Y: 50}, Button: MouseButtonLeft})

	if drawCalls < 1 {
		t.Fatalf("draw was not called after pan")
	}
	xMin, xMax, _, _ := axisLimits(t, ax)
	if !(xMin < 0 && xMax < 10) {
		t.Fatalf("xlim after pan drag = (%g, %g), expected shifted left", xMin, xMax)
	}
}

func TestNavigationZoomReleaseAppliesRect(t *testing.T) {
	fig, ax := newNavigationFigure(t)
	ctx := core.AxesDrawContext(ax, fig)
	nav := NewNavigation(fig, nil)
	nav.SetMode(NavZoom)

	var dispatcher Dispatcher
	nav.Attach(&dispatcher)
	defer nav.Detach()

	a := (&ctx.DataToPixel).Apply(geom.Pt{X: 1, Y: 1})
	b := (&ctx.DataToPixel).Apply(geom.Pt{X: 4, Y: 4})

	_ = dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: ax, Position: a, Button: MouseButtonLeft})
	_ = dispatcher.Emit(Event{Type: EventMouseMove, Figure: fig, Axes: ax, Position: b, Button: MouseButtonLeft})
	_ = dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: ax, Position: b, Button: MouseButtonLeft})

	xMin, xMax, yMin, yMax := axisLimits(t, ax)
	if !almost(xMin, 1) || !almost(xMax, 4) || !almost(yMin, 1) || !almost(yMax, 4) {
		t.Fatalf("limits = (%g,%g,%g,%g), want (1,4,1,4)", xMin, xMax, yMin, yMax)
	}
}

func TestNavigationScrollZooms(t *testing.T) {
	fig, ax := newNavigationFigure(t)
	ctx := core.AxesDrawContext(ax, fig)
	nav := NewNavigation(fig, nil)
	if err := nav.SetScrollFactor(2); err != nil {
		t.Fatalf("SetScrollFactor: %v", err)
	}

	var dispatcher Dispatcher
	nav.Attach(&dispatcher)
	defer nav.Detach()

	anchor := (&ctx.DataToPixel).Apply(geom.Pt{X: 5, Y: 5})
	_ = dispatcher.Emit(Event{
		Type:     EventScroll,
		Figure:   fig,
		Axes:     ax,
		Position: anchor,
		DeltaY:   1,
	})

	xMin, xMax, _, _ := axisLimits(t, ax)
	if !(xMax-xMin < 10) {
		t.Fatalf("scroll-in did not shrink axes: (%g, %g)", xMin, xMax)
	}
}
