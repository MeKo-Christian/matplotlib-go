package core_test

import (
	"os"
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestSaveSVGWithSupportedRenderer(t *testing.T) {
	fig := core.NewFigure(120, 80)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.SetTitle("SaveSVG test")
	_, _ = ax.Plot([]float64{0, 10}, []float64{0, 1}, core.PlotOptions{})

	renderer, err := backends.Create(backends.SVG, backends.Config{
		Width:      120,
		Height:     80,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		DPI:        72,
	})
	if err != nil {
		t.Fatalf("creating SVG renderer failed: %v", err)
	}

	out, err := os.CreateTemp("", "matplotlib-go-save-svg-*.svg")
	if err != nil {
		t.Fatalf("CreateTemp failed: %v", err)
	}
	path := out.Name()
	out.Close()
	t.Cleanup(func() { _ = os.Remove(path) })

	err = core.SaveSVG(fig, renderer, path)
	if err != nil {
		t.Fatalf("SaveSVG failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "<svg") {
		t.Fatal("SVG output missing <svg> root")
	}
	if !strings.Contains(content, "<text") {
		t.Fatal("expected text output in SVG export")
	}
}

func TestSaveSVGPassesSVGOptionsToRenderer(t *testing.T) {
	fig := core.NewFigure(120, 80)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.SetTitle("Path text")
	_, _ = ax.Plot([]float64{0, 10}, []float64{0, 1}, core.PlotOptions{})

	renderer, err := backends.Create(backends.SVG, backends.Config{
		Width:      120,
		Height:     80,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		DPI:        72,
	})
	if err != nil {
		t.Fatalf("creating SVG renderer failed: %v", err)
	}

	out, err := os.CreateTemp("", "matplotlib-go-save-svg-options-*.svg")
	if err != nil {
		t.Fatalf("CreateTemp failed: %v", err)
	}
	path := out.Name()
	out.Close()
	t.Cleanup(func() { _ = os.Remove(path) })

	err = core.SaveSVG(
		fig,
		renderer,
		path,
		render.WithSVGFontPolicy(render.SVGFontPolicyPath),
		render.WithSVGMetadata(map[string]string{"Title": "Configured"}),
		render.WithSVGHashSalt("core"),
	)
	if err != nil {
		t.Fatalf("SaveSVG failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, `<meta name="Title" content="Configured" />`) {
		t.Fatalf("SaveSVG should forward metadata options, got %q", content)
	}
	if strings.Contains(content, "<text") {
		t.Fatalf("SaveSVG should apply path font policy before drawing, got %q", content)
	}
	if strings.Contains(content, `id="clip1"`) {
		t.Fatalf("SaveSVG should apply hash salt before clip IDs are registered, got %q", content)
	}
}

func TestSaveSVGUnsupportedRenderer(t *testing.T) {
	fig := core.NewFigure(20, 20)
	err := core.SaveSVG(fig, &render.NullRenderer{}, "unsupported.svg")
	if err == nil || !strings.Contains(err.Error(), "not supported") {
		t.Fatalf("expected unsupported renderer error, got %v", err)
	}
}
