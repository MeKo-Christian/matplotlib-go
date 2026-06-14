package gio

import (
	"testing"

	"gioui.org/f32"
	"gioui.org/io/pointer"
	"github.com/cwbudde/matplotlib-go/backends/desktop"
	"github.com/cwbudde/matplotlib-go/backends/gobasic"
	"github.com/cwbudde/matplotlib-go/canvas"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/examplecatalog"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/test/parity"
)

func TestInteractiveCoverageMatrixGio(t *testing.T) {
	for _, row := range examplecatalog.InteractiveCoverageMatrix() {
		row := row
		t.Run(row.Topic, func(t *testing.T) {
			fig, _, err := parity.Figure(row.RepresentativeID)
			if err != nil {
				t.Fatalf("Figure(%q): %v", row.RepresentativeID, err)
			}
			b, err := New(desktop.Options{
				Figure: fig,
				Width:  int(fig.SizePx.X + 0.5),
				Height: int(fig.SizePx.Y + 0.5),
				Renderer: func(w, h int, bg render.Color) (render.Renderer, error) {
					return gobasic.New(w, h, bg), nil
				},
			})
			if err != nil {
				t.Fatalf("New: %v", err)
			}

			draws := 0
			b.Canvas().Connect(canvas.EventDraw, func(canvas.Event) error {
				draws++
				return nil
			})
			if err := b.Canvas().Draw(); err != nil {
				t.Fatalf("Draw: %v", err)
			}
			if draws == 0 {
				t.Fatal("draw event did not fire")
			}

			center, ok := representativePoint(fig)
			if !ok {
				t.Fatalf("%s has no axes to drive", row.RepresentativeID)
			}
			if err := b.Toolbar().Trigger(canvas.ToolbarPan); err != nil {
				t.Fatalf("trigger pan: %v", err)
			}
			b.dispatchPointer(pointer.Event{
				Kind:     pointer.Press,
				Position: f32.Pt(float32(center.X), float32(center.Y)),
				Buttons:  pointer.ButtonPrimary,
			})
			b.dispatchPointer(pointer.Event{
				Kind:     pointer.Drag,
				Position: f32.Pt(float32(center.X+12), float32(center.Y)),
				Buttons:  pointer.ButtonPrimary,
			})
			b.dispatchPointer(pointer.Event{
				Kind:     pointer.Release,
				Position: f32.Pt(float32(center.X+12), float32(center.Y)),
				Buttons:  pointer.ButtonPrimary,
			})
			b.dispatchPointer(pointer.Event{
				Kind:     pointer.Scroll,
				Position: f32.Pt(float32(center.X), float32(center.Y)),
				Scroll:   f32.Pt(0, 1),
			})
			b.dispatchPointer(pointer.Event{
				Kind:     pointer.Press,
				Position: f32.Pt(float32(center.X), float32(center.Y)),
				Buttons:  pointer.ButtonPrimary,
			})
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
