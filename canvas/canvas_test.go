package canvas

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
)

func TestResolveEventTarget(t *testing.T) {
	fig := core.NewFigure(200, 100)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.25, Y: 0.25},
		Max: geom.Pt{X: 0.75, Y: 0.75},
	})
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 20)

	target, data, ok := ResolveEventTarget(fig, geom.Pt{X: 100, Y: 50})
	if target != ax {
		t.Fatalf("ResolveEventTarget() axes = %p, want %p", target, ax)
	}
	if !ok {
		t.Fatal("ResolveEventTarget() did not resolve data coordinates")
	}
	if data.X != 5 {
		t.Fatalf("data.X = %v, want 5", data.X)
	}
	if data.Y != 10 {
		t.Fatalf("data.Y = %v, want 10", data.Y)
	}

	target, _, ok = ResolveEventTarget(fig, geom.Pt{X: 5, Y: 5})
	if target != nil || ok {
		t.Fatalf("ResolveEventTarget(outside) = (%v, %v), want (nil, false)", target, ok)
	}
}

func TestDispatcherPreservesOrderAndDisconnects(t *testing.T) {
	var dispatcher Dispatcher
	var calls []int

	first := dispatcher.Connect(EventDraw, func(Event) error {
		calls = append(calls, 1)
		return nil
	})
	dispatcher.Connect(EventDraw, func(Event) error {
		calls = append(calls, 2)
		return nil
	})
	dispatcher.Disconnect(first)

	if err := dispatcher.Emit(Event{Type: EventDraw}); err != nil {
		t.Fatalf("Emit() error = %v", err)
	}
	if len(calls) != 1 || calls[0] != 2 {
		t.Fatalf("call order = %v, want [2]", calls)
	}
}
