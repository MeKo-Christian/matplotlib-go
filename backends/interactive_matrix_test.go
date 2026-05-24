package backends_test

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/backends"
	"github.com/cwbudde/matplotlib-go/backends/desktop"
	"github.com/cwbudde/matplotlib-go/backends/desktop/gio"
	"github.com/cwbudde/matplotlib-go/backends/gobasic"
	"github.com/cwbudde/matplotlib-go/backends/webagg"
	"github.com/cwbudde/matplotlib-go/canvas"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/examplecatalog"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/test/parity"
)

func TestInteractiveMatrixHasFigureFactoryForEveryCatalogTopic(t *testing.T) {
	for _, row := range examplecatalog.InteractiveCoverageMatrix() {
		if row.RepresentativeID == "" {
			t.Fatalf("topic %q has no representative case", row.Topic)
		}
		if _, _, err := parity.Figure(row.RepresentativeID); err != nil {
			t.Fatalf("topic %q representative %q is not figure-backed: %v", row.Topic, row.RepresentativeID, err)
		}
	}
}

func TestInteractiveMatrixCoversEveryCatalogTopic(t *testing.T) {
	topics := map[string]bool{}
	for _, c := range examplecatalog.Cases() {
		topics[c.Topic] = true
	}
	rows := examplecatalog.InteractiveCoverageMatrix()
	if len(rows) != len(topics) {
		t.Fatalf("matrix rows = %d, catalog topics = %d", len(rows), len(topics))
	}
	for _, row := range rows {
		if !topics[row.Topic] {
			t.Fatalf("matrix topic %q is not in catalog", row.Topic)
		}
		if !row.WebAgg || !row.Gio {
			t.Fatalf("topic %q must cover WebAgg and Gio, got webagg=%v gio=%v", row.Topic, row.WebAgg, row.Gio)
		}
		delete(topics, row.Topic)
	}
	for topic := range topics {
		t.Fatalf("missing interactive coverage row for topic %q", topic)
	}
}

func TestInteractiveBackendsExposeCommonSurface(t *testing.T) {
	fig := core.NewFigure(200, 100)
	reg := backends.NewRegistry()
	reg.Register(backends.Backend("surface"), &backends.BackendInfo{
		Name:      "Surface",
		Available: true,
		Factory: func(backends.Config) (render.Renderer, error) {
			return &render.NullRenderer{}, nil
		},
	})
	previous := backends.DefaultRegistry
	backends.DefaultRegistry = reg
	t.Cleanup(func() {
		backends.DefaultRegistry = previous
	})

	headless, _, err := backends.NewManager("surface", backends.SimpleConfig(200, 100, render.Color{A: 1}), fig, nil)
	if err != nil {
		t.Fatalf("headless NewManager: %v", err)
	}
	assertCommonSurface(t, "headless", headless.Canvas(), headless.ToolManager(), false)

	web, err := webagg.NewManager(webagg.Options{
		Figure: core.NewFigure(200, 100),
		Renderer: func(w, h int, bg render.Color) (webagg.RasterRenderer, error) {
			return gobasic.New(w, h, bg), nil
		},
		HasBackground: true,
		Background:    render.Color{R: 1, G: 1, B: 1, A: 1},
	})
	if err != nil {
		t.Fatalf("webagg NewManager: %v", err)
	}
	assertCommonSurface(t, "webagg", web.Canvas(), web.ToolManager(), true)

	desktopBackend, err := gio.New(desktop.Options{
		Figure: core.NewFigure(200, 100),
		Width:  200,
		Height: 100,
		Renderer: func(w, h int, bg render.Color) (render.Renderer, error) {
			return gobasic.New(w, h, bg), nil
		},
	})
	if err != nil {
		t.Fatalf("gio New: %v", err)
	}
	assertCommonSurface(t, "gio", desktopBackend.Canvas(), canvas.NewToolManager(), false)
}

func assertCommonSurface(t *testing.T, name string, c canvas.FigureCanvas, tools *canvas.ToolManager, wantBlit bool) {
	t.Helper()
	if c == nil {
		t.Fatalf("%s canvas is nil", name)
	}
	if c.Figure() == nil {
		t.Fatalf("%s Figure() is nil", name)
	}
	if _, ok := c.(canvas.DrawIdleCanvas); !ok {
		t.Fatalf("%s does not implement DrawIdleCanvas", name)
	}
	if _, ok := c.(canvas.BlitCanvas); ok != wantBlit {
		t.Fatalf("%s BlitCanvas = %v, want %v", name, ok, wantBlit)
	}
	id := c.Connect(canvas.EventDraw, func(canvas.Event) error { return nil })
	if id == 0 {
		t.Fatalf("%s Connect returned zero id", name)
	}
	c.Disconnect(id)
	if tools == nil {
		t.Fatalf("%s ToolManager is nil", name)
	}
}
