package canvas

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/widgets"
)

func TestWidgetInteractionCursorHover(t *testing.T) {
	fig := core.NewFigure(200, 100)
	left := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 0.5, Y: 1}})
	right := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.5, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	leftCursor := widgets.NewCursor(left)
	multiCursor := widgets.NewMultiCursor(left, right)

	var dispatcher Dispatcher
	wi := NewWidgetInteraction(fig, func() error { return nil })
	wi.Attach(&dispatcher)
	defer wi.Detach()

	insideLeft := geom.Pt{X: 40, Y: 50}
	if err := dispatcher.Emit(Event{Type: EventMouseMove, Figure: fig, Position: insideLeft}); err != nil {
		t.Fatalf("move left: %v", err)
	}
	if _, _, ok := leftCursor.Position(); !ok {
		t.Fatal("left cursor should be visible inside left axis")
	}
	if _, _, ok := multiCursor.Position(); !ok {
		t.Fatal("multi cursor should be visible inside left axis")
	}

	insideRight := geom.Pt{X: 140, Y: 50}
	if err := dispatcher.Emit(Event{Type: EventMouseMove, Figure: fig, Position: insideRight}); err != nil {
		t.Fatalf("move right: %v", err)
	}
	if _, _, ok := leftCursor.Position(); ok {
		t.Fatal("left cursor should hide outside left axis")
	}
	if _, _, ok := multiCursor.Position(); !ok {
		t.Fatal("multi cursor should remain visible in right axis")
	}

	outside := geom.Pt{X: 250, Y: 50}
	if err := dispatcher.Emit(Event{Type: EventFigureLeave, Figure: fig, Position: outside}); err != nil {
		t.Fatalf("figure leave: %v", err)
	}
	if _, _, ok := multiCursor.Position(); ok {
		t.Fatal("multi cursor should hide on figure leave")
	}
}
