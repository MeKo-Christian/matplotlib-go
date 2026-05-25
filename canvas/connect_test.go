package canvas

import (
	"errors"
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
)

// fakeCanvas implements FigureCanvas in memory; it lets the connect /
// disconnect lifecycle be exercised without pulling a backend.
type fakeCanvas struct {
	fig        *Figure
	dispatcher Dispatcher
	closed     bool
}

func (c *fakeCanvas) Figure() *Figure       { return c.fig }
func (c *fakeCanvas) Draw() error           { return c.dispatcher.Emit(Event{Type: EventDraw, Figure: c.fig}) }
func (c *fakeCanvas) Resize(_, _ int) error { return nil }
func (c *fakeCanvas) Connect(t EventType, h Handler) ConnectionID {
	if c.closed {
		return 0
	}
	return c.dispatcher.Connect(t, h)
}
func (c *fakeCanvas) Disconnect(id ConnectionID) { c.dispatcher.Disconnect(id) }
func (c *fakeCanvas) Close() error {
	c.closed = true
	return c.dispatcher.Emit(Event{Type: EventClose, Figure: c.fig})
}

func TestMplConnectLifecycleRoundTrip(t *testing.T) {
	fig := &Figure{}
	c := &fakeCanvas{fig: fig}

	var seen []EventType
	id := Connect(c, EventMousePress, func(ev Event) error {
		seen = append(seen, ev.Type)
		return nil
	})
	if id == 0 {
		t.Fatalf("expected non-zero connection id")
	}

	// Fire a press, the handler runs.
	_ = c.dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Position: geom.Pt{X: 1, Y: 2}})
	if len(seen) != 1 || seen[0] != EventMousePress {
		t.Fatalf("seen = %v, want [mouse_press]", seen)
	}

	// Disconnect, then fire another press — no extra calls.
	Disconnect(c, id)
	_ = c.dispatcher.Emit(Event{Type: EventMousePress, Figure: fig})
	if len(seen) != 1 {
		t.Fatalf("seen after disconnect = %v, want unchanged", seen)
	}

	// Disconnect with zero ID is a no-op.
	Disconnect(c, 0)
	// Connect on a nil canvas is a no-op.
	if id := Connect(nil, EventDraw, func(Event) error { return nil }); id != 0 {
		t.Fatalf("Connect(nil) = %d, want 0", id)
	}
}

func TestEventLifecycleDrawCloseFlow(t *testing.T) {
	fig := &Figure{}
	c := &fakeCanvas{fig: fig}

	var events []EventType
	Connect(c, EventDraw, func(ev Event) error { events = append(events, ev.Type); return nil })
	Connect(c, EventClose, func(ev Event) error { events = append(events, ev.Type); return nil })

	if err := c.Draw(); err != nil {
		t.Fatalf("Draw() = %v", err)
	}
	if err := c.Close(); err != nil {
		t.Fatalf("Close() = %v", err)
	}

	want := []EventType{EventDraw, EventClose}
	if len(events) != len(want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	for i, e := range events {
		if e != want[i] {
			t.Fatalf("events[%d] = %s, want %s", i, e, want[i])
		}
	}
}

func TestHandlerErrorPropagates(t *testing.T) {
	var dispatcher Dispatcher
	dispatcher.Connect(EventDraw, func(Event) error { return errors.New("boom") })
	if err := dispatcher.Emit(Event{Type: EventDraw}); err == nil {
		t.Fatalf("expected error from handler, got nil")
	}
}
