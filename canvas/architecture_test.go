package canvas

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
)

func TestFirstClassEventWrappersPreserveCanonicalTypes(t *testing.T) {
	fig := &Figure{}

	draw := NewDrawEvent(fig, 640, 480)
	if draw.Type != EventDraw || draw.Figure != fig || draw.Width != 640 || draw.Height != 480 {
		t.Fatalf("draw event = %+v", draw.Event)
	}

	mouse := NewMouseEvent(EventMousePress, fig, geom.Pt{X: 12, Y: 34}, MouseButtonLeft)
	if mouse.Type != EventMousePress || mouse.Position.X != 12 || mouse.Button != MouseButtonLeft {
		t.Fatalf("mouse event = %+v", mouse.Event)
	}

	enter := NewMouseEvent(EventFigureEnter, fig, geom.Pt{X: 12, Y: 34}, MouseButtonNone)
	if enter.Type != EventFigureEnter || enter.Figure != fig || enter.Position.X != 12 {
		t.Fatalf("figure enter event = %+v", enter.Event)
	}

	leave := NewMouseEvent(EventFigureLeave, fig, geom.Pt{X: 56, Y: 78}, MouseButtonNone)
	if leave.Type != EventFigureLeave || leave.Figure != fig || leave.Position.Y != 78 {
		t.Fatalf("figure leave event = %+v", leave.Event)
	}

	axesEnter := NewMouseEvent(EventAxesEnter, fig, geom.Pt{X: 90, Y: 12}, MouseButtonNone)
	if axesEnter.Type != EventAxesEnter || axesEnter.Figure != fig || axesEnter.Position.X != 90 {
		t.Fatalf("axes enter event = %+v", axesEnter.Event)
	}

	axesLeave := NewMouseEvent(EventAxesLeave, fig, geom.Pt{X: 91, Y: 13}, MouseButtonNone)
	if axesLeave.Type != EventAxesLeave || axesLeave.Figure != fig || axesLeave.Position.Y != 13 {
		t.Fatalf("axes leave event = %+v", axesLeave.Event)
	}

	key := NewKeyEvent(EventKeyPress, fig, "ctrl+s", ModifierControl)
	if key.Type != EventKeyPress || key.Key != "ctrl+s" || key.Modifiers != ModifierControl {
		t.Fatalf("key event = %+v", key.Event)
	}

	pick := NewPickEvent(fig, nil, mouse)
	if pick.Type != EventPick || pick.Figure != fig {
		t.Fatalf("pick event = %+v", pick.Event)
	}
}

func TestDispatcherConnectionLifecycle(t *testing.T) {
	var dispatcher Dispatcher
	called := 0
	id := dispatcher.Connect(EventDraw, func(Event) error {
		called++
		return nil
	})
	if id == 0 {
		t.Fatal("connection id = 0, want non-zero")
	}

	if err := dispatcher.Emit(Event{Type: EventDraw}); err != nil {
		t.Fatal(err)
	}
	dispatcher.Disconnect(id)
	if err := dispatcher.Emit(Event{Type: EventDraw}); err != nil {
		t.Fatal(err)
	}
	if called != 1 {
		t.Fatalf("called = %d, want 1", called)
	}
}
