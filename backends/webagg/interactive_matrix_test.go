package webagg

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/backends/gobasic"
	plotcanvas "github.com/cwbudde/matplotlib-go/canvas"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/examplecatalog"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/test/parity"
)

func TestInteractiveCoverageMatrixWebAgg(t *testing.T) {
	for _, row := range examplecatalog.InteractiveCoverageMatrix() {
		row := row
		t.Run(row.Topic, func(t *testing.T) {
			fig, _, err := parity.Figure(row.RepresentativeID)
			if err != nil {
				t.Fatalf("Figure(%q): %v", row.RepresentativeID, err)
			}
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

			draws := 0
			mgr.Connect(plotcanvas.EventDraw, func(plotcanvas.Event) error {
				draws++
				return nil
			})
			if err := mgr.Draw(); err != nil {
				t.Fatalf("Draw: %v", err)
			}
			if draws == 0 {
				t.Fatal("draw event did not fire")
			}

			center, ok := representativePoint(fig)
			if !ok {
				t.Fatalf("%s has no axes to drive", row.RepresentativeID)
			}
			for _, msg := range []map[string]any{
				{"type": "toolbar_button", "name": "pan"},
				{"type": "button_press", "x": center.X, "y": center.Y, "button": 0},
				{"type": "motion_notify", "x": center.X + 12, "y": center.Y, "buttons": 1},
				{"type": "button_release", "x": center.X + 12, "y": center.Y, "button": 0},
				{"type": "scroll", "x": center.X, "y": center.Y, "step": 1.0},
				{"type": "button_press", "x": center.X, "y": center.Y, "button": 0},
			} {
				if err := mgr.HandleClientMessage(mustJSON(t, msg)); err != nil {
					t.Fatalf("HandleClientMessage %v: %v", msg, err)
				}
			}
		})
	}
}

func representativePoint(fig *core.Figure) (geom.Pt, bool) {
	if fig == nil || len(fig.Children) == 0 {
		return geom.Pt{}, false
	}
	ax := fig.Children[0]
	if ax == nil {
		return geom.Pt{}, false
	}
	rect := ax.RectFraction
	return geom.Pt{
		X: fig.SizePx.X * (rect.Min.X + rect.Max.X) / 2,
		Y: fig.SizePx.Y * (1 - (rect.Min.Y+rect.Max.Y)/2),
	}, true
}
