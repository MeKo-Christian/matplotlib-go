// Command desktop is an embedding example for the Gio-backed desktop
// interactive backend (PLAN.md §4.2). It opens a window showing the
// basic_line example, wires the canvas to AGG, prints coordinate
// hovers via the toolbar status line, and supports pan / zoom /
// scroll-zoom out of the box.
//
// Run:
//
//	go run ./examples/embed/desktop
//
// Controls:
//
//	left-drag (pan mode)  → pan the axes
//	left-drag (zoom mode) → rubber-band zoom
//	scroll                → zoom under cursor
//	p                     → toggle pan mode
//	z                     → toggle zoom mode
//	r                     → reset view (home)
//	s                     → save figure as out.png
//	q / Esc               → close window
//
// This file imports the Gio adapter via the blank import below; that
// is what registers the desktop.Constructor consumed by desktop.New.
package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/backends/desktop"
	dgio "github.com/cwbudde/matplotlib-go/backends/desktop/gio"
	"github.com/cwbudde/matplotlib-go/canvas"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/examples/basic_line"
	"github.com/cwbudde/matplotlib-go/render"
)

func main() {
	fig := basic_line.Plot()

	b, err := desktop.New(desktop.Options{
		Figure:     fig,
		Title:      "matplotlib-go — basic_line",
		Width:      basic_line.Width,
		Height:     basic_line.Height,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		Renderer: func(w, h int, bg render.Color) (render.Renderer, error) {
			return agg.New(w, h, bg)
		},
		SaveHandler: func() error {
			return savePNG(fig, "out.png")
		},
	})
	if err != nil {
		if errors.Is(err, desktop.ErrNotImplemented) {
			log.Fatal("no desktop adapter linked in — import _ \"github.com/cwbudde/matplotlib-go/backends/desktop/gio\"")
		}
		log.Fatal(err)
	}

	wireKeyboardShortcuts(b)
	wireHoverCoordinates(b)

	go func() {
		if err := b.Run(); err != nil {
			log.Printf("backend: %v", err)
		}
		os.Exit(0)
	}()

	dgio.Main()
}

// wireKeyboardShortcuts maps the documented single-key shortcuts onto
// the toolbar actions. Modifiers are ignored to keep the example small.
func wireKeyboardShortcuts(b desktop.Backend) {
	b.Canvas().Connect(canvas.EventKeyPress, func(ev canvas.Event) error {
		switch ev.Key {
		case "P", "p":
			return b.Toolbar().Trigger(canvas.ToolbarPan)
		case "Z", "z":
			return b.Toolbar().Trigger(canvas.ToolbarZoom)
		case "R", "r":
			return b.Toolbar().Trigger(canvas.ToolbarHome)
		case "S", "s":
			return b.Toolbar().Trigger(canvas.ToolbarSave)
		case "Q", "q", "⎋":
			return b.Close()
		}
		return nil
	})
}

// wireHoverCoordinates updates the toolbar status line with the
// data-space coordinates under the cursor — Matplotlib's default
// format_coord behavior.
func wireHoverCoordinates(b desktop.Backend) {
	fig := b.Canvas().Figure()
	tb := b.Toolbar()
	b.Canvas().Connect(canvas.EventMouseMove, func(ev canvas.Event) error {
		if ax, _, ok := canvas.ResolveEventTarget(fig, ev.Position); ok {
			tb.SetMessage(ax.FormatCoord(ev.Position))
		} else {
			tb.SetMessage(fmt.Sprintf("%.0f, %.0f", ev.Position.X, ev.Position.Y))
		}
		return nil
	})
}

// savePNG renders the figure into a fresh AGG context at its current
// display size and writes it to path. The toolbar Save button calls this
// via the SaveHandler option above.
func savePNG(fig *core.Figure, path string) error {
	w, h := int(fig.SizePx.X), int(fig.SizePx.Y)
	if w <= 0 || h <= 0 {
		w, h = basic_line.Width, basic_line.Height
	}
	r, err := agg.New(w, h, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		return err
	}
	return core.SavePNG(fig, r, path)
}
