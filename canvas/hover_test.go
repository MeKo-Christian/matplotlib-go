package canvas

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
)

func TestAxesHoverTrackerEmitsEnterLeaveOnAxesChanges(t *testing.T) {
	fig := core.NewFigure(200, 100)
	left := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 0.5, Y: 1}})
	right := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.5, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})

	var dispatcher Dispatcher
	var got []Event
	dispatcher.Connect(EventAxesEnter, func(ev Event) error {
		got = append(got, ev)
		return nil
	})
	dispatcher.Connect(EventAxesLeave, func(ev Event) error {
		got = append(got, ev)
		return nil
	})

	tracker := NewAxesHoverTracker(fig, &dispatcher)
	tracker.Update(Event{Type: EventMouseMove, Figure: fig, Position: geom.Pt{X: 25, Y: 50}})
	tracker.Update(Event{Type: EventMouseMove, Figure: fig, Position: geom.Pt{X: 75, Y: 50}})
	tracker.Update(Event{Type: EventMouseMove, Figure: fig, Position: geom.Pt{X: 125, Y: 50}})
	tracker.Update(Event{Type: EventMouseMove, Figure: fig, Position: geom.Pt{X: 199, Y: 50}})
	tracker.Update(Event{Type: EventMouseMove, Figure: fig, Position: geom.Pt{X: 250, Y: 50}})

	wantTypes := []EventType{EventAxesEnter, EventAxesLeave, EventAxesEnter, EventAxesLeave}
	wantAxes := []*Axes{left, left, right, right}
	if len(got) != len(wantTypes) {
		t.Fatalf("events = %d, want %d", len(got), len(wantTypes))
	}
	for i := range wantTypes {
		if got[i].Type != wantTypes[i] {
			t.Fatalf("event %d type = %s, want %s", i, got[i].Type, wantTypes[i])
		}
		if got[i].Axes != wantAxes[i] {
			t.Fatalf("event %d axes mismatch", i)
		}
	}
}

func TestAxesHoverTrackerLeavesOnFigureLeave(t *testing.T) {
	fig := core.NewFigure(100, 100)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})

	var dispatcher Dispatcher
	var got []EventType
	dispatcher.Connect(EventAxesEnter, func(ev Event) error {
		got = append(got, ev.Type)
		return nil
	})
	dispatcher.Connect(EventAxesLeave, func(ev Event) error {
		got = append(got, ev.Type)
		if ev.Axes != ax {
			t.Fatalf("leave axes mismatch")
		}
		return nil
	})

	tracker := NewAxesHoverTracker(fig, &dispatcher)
	tracker.Update(Event{Type: EventMouseMove, Figure: fig, Position: geom.Pt{X: 50, Y: 50}})
	tracker.Update(Event{Type: EventFigureLeave, Figure: fig, Position: geom.Pt{X: 110, Y: 50}})

	want := []EventType{EventAxesEnter, EventAxesLeave}
	if len(got) != len(want) {
		t.Fatalf("events = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("event %d = %s, want %s", i, got[i], want[i])
		}
	}
}
