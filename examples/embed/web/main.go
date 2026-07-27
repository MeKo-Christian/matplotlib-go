// Command web is an embedding example for the WebAgg backend.
// It serves the basic_line example over HTTP on
// :8080 and broadcasts AGG diff frames to every connected browser
// client.
//
// Run:
//
//	go run ./examples/embed/web
//	# then open http://localhost:8080
//
// Controls (in the browser):
//
//	pan toggle  → toolbar Pan button, then left-drag the figure
//	zoom toggle → toolbar Zoom button, then drag to rubber-band
//	scroll      → zoom under cursor
//	home        → toolbar Home button (reset view)
//	save        → toolbar Save button (writes out.png server-side)
//
// The save handler writes a fresh PNG render to ./out.png. Hosts that
// want to stream the PNG to the client instead can attach a per-client
// handler via Manager.Toolbar().SetHandler.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/backends/webagg"
	"github.com/cwbudde/matplotlib-go/canvas"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/examples/basic_line"
	"github.com/cwbudde/matplotlib-go/render"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	flag.Parse()

	fig := basic_line.Plot()
	mgr, err := webagg.NewManager(webagg.Options{
		Figure: fig,
		Renderer: func(w, h int, bg render.Color) (webagg.RasterRenderer, error) {
			return agg.New(w, h, bg)
		},
		HasBackground: true,
		Background:    render.Color{R: 1, G: 1, B: 1, A: 1},
		SaveHandler: func() error {
			return savePNG(fig, "out.png")
		},
	})
	if err != nil {
		log.Fatalf("webagg: %v", err)
	}
	mgr.SetTitle("matplotlib-go — basic_line (WebAgg)")

	// Wire a hover→toolbar-status bridge so the JS client's status
	// strip shows the data-space coordinate under the cursor.
	mgr.Connect(canvas.EventMouseMove, func(ev canvas.Event) error {
		if ax, _, ok := canvas.ResolveEventTarget(fig, ev.Position); ok {
			mgr.Toolbar().SetMessage(ax.FormatCoord(ev.Position))
		} else {
			mgr.Toolbar().SetMessage(fmt.Sprintf("%.0f, %.0f", ev.Position.X, ev.Position.Y))
		}
		return nil
	})

	srv, err := webagg.NewServer(webagg.ServerOptions{
		Manager: mgr,
		Logger:  log.Default(),
	})
	if err != nil {
		log.Fatalf("webagg.NewServer: %v", err)
	}

	log.Printf("listening on %s — open http://localhost%s", *addr, *addr)
	mux := http.NewServeMux()
	mux.Handle("/", srv)
	if err := http.ListenAndServe(*addr, mux); err != nil { //nolint:gosec // example
		log.Fatal(err)
	}
}

// savePNG re-renders the figure at its current display size with a
// fresh AGG context and writes it to path. Triggered by the toolbar
// Save button.
func savePNG(fig *core.Figure, path string) error {
	w, h := fig.CanvasSize()
	if w <= 0 || h <= 0 {
		w, h = basic_line.Width, basic_line.Height
	}
	r, err := agg.New(w, h, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		return err
	}
	return core.SavePNG(fig, r, path)
}
