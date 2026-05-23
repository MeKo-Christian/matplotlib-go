package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/cwbudde/matplotlib-go/backends"
	"github.com/cwbudde/matplotlib-go/backends/desktop"
	"github.com/cwbudde/matplotlib-go/backends/desktop/gio"
	"github.com/cwbudde/matplotlib-go/backends/gobasic"
	"github.com/cwbudde/matplotlib-go/backends/webagg"
	"github.com/cwbudde/matplotlib-go/canvas"
	"github.com/cwbudde/matplotlib-go/examples/basic_line"
	"github.com/cwbudde/matplotlib-go/render"
)

func main() {
	backend := flag.String("backend", "headless", "headless, gio, or webagg")
	addr := flag.String("addr", ":8080", "address for webagg")
	flag.Parse()

	fig := basic_line.Plot()
	switch *backend {
	case "headless":
		manager, _, err := backends.NewManager("gobasic", backends.SimpleConfig(
			int(fig.SizePx.X+0.5),
			int(fig.SizePx.Y+0.5),
			render.Color{R: 1, G: 1, B: 1, A: 1},
		), fig, nil)
		if err != nil {
			log.Fatal(err)
		}
		wireEvents(manager.Canvas())
		if err := manager.Show(); err != nil {
			log.Fatal(err)
		}
	case "gio":
		manager, err := gio.New(desktop.Options{
			Figure: fig,
			Width:  int(fig.SizePx.X + 0.5),
			Height: int(fig.SizePx.Y + 0.5),
			Renderer: func(w, h int, bg render.Color) (render.Renderer, error) {
				return gobasic.New(w, h, bg), nil
			},
		})
		if err != nil {
			log.Fatal(err)
		}
		wireEvents(manager.Canvas())
		go func() {
			if err := manager.Run(); err != nil {
				log.Fatal(err)
			}
		}()
		gio.Main()
	case "webagg":
		manager, err := webagg.NewManager(webagg.Options{
			Figure: fig,
			Renderer: func(w, h int, bg render.Color) (webagg.RasterRenderer, error) {
				return gobasic.New(w, h, bg), nil
			},
			HasBackground: true,
			Background:    render.Color{R: 1, G: 1, B: 1, A: 1},
		})
		if err != nil {
			log.Fatal(err)
		}
		wireEvents(manager.Canvas())
		server, err := webagg.NewServer(webagg.ServerOptions{Manager: manager})
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("serving WebAgg on http://localhost%s", *addr)
		log.Fatal(http.ListenAndServe(*addr, server))
	default:
		log.Fatalf("unknown backend %q", *backend)
	}
}

func wireEvents(c canvas.FigureCanvas) {
	c.Connect(canvas.EventMouseMove, func(ev canvas.Event) error {
		if ev.Axes != nil {
			fmt.Println(ev.Axes.FormatCoord(ev.Position))
		}
		return nil
	})
	c.Connect(canvas.EventPick, func(ev canvas.Event) error {
		fmt.Printf("picked at %.1f, %.1f\n", ev.Position.X, ev.Position.Y)
		return nil
	})
}
