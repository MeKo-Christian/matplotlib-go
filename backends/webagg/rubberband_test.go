package webagg

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/backends/gobasic"
	"github.com/cwbudde/matplotlib-go/canvas"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// TestRubberbandEmittedDuringZoomDrag verifies the server emits rubberband
// overlay events while dragging a zoom rectangle and clears it on release,
// closing the gap where the JS handler had nothing to consume.
func TestRubberbandEmittedDuringZoomDrag(t *testing.T) {
	fig := core.NewFigure(100, 80)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 10)
	mgr, err := NewManager(Options{
		Figure: fig,
		Renderer: func(w, h int, bg render.Color) (RasterRenderer, error) {
			return gobasic.New(w, h, bg), nil
		},
		HasBackground: true,
		Background:    render.Color{R: 1, G: 1, B: 1, A: 1},
	})
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	// Capture rubberband broadcasts via the handler the manager installed.
	var got []struct {
		rect   geom.Rect
		active bool
	}
	mgr.Navigation().SetRubberbandHandler(func(rect geom.Rect, active bool) {
		got = append(got, struct {
			rect   geom.Rect
			active bool
		}{rect, active})
	})

	mgr.Navigation().SetMode(canvas.NavZoom)
	d := mgr.Dispatcher()
	press := geom.Pt{X: 20, Y: 20}
	drag := geom.Pt{X: 60, Y: 50}
	_ = d.Emit(canvas.Event{Type: canvas.EventMousePress, Figure: fig, Position: press})
	_ = d.Emit(canvas.Event{Type: canvas.EventMouseMove, Figure: fig, Position: drag})
	_ = d.Emit(canvas.Event{Type: canvas.EventMouseRelease, Figure: fig, Position: drag})

	if len(got) < 2 {
		t.Fatalf("expected at least one active and one clearing rubberband event, got %d", len(got))
	}
	if !got[0].active {
		t.Fatalf("first rubberband event should be active (drag in progress)")
	}
	if got[len(got)-1].active {
		t.Fatalf("final rubberband event should clear the overlay (active=false)")
	}
}
