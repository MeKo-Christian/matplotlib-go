package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/transform"
)

func main() {
	var (
		width  = flag.Int("width", 800, "figure width in pixels")
		height = flag.Int("height", 500, "figure height in pixels")
		output = flag.String("output", "export.svg", "output SVG file")
	)
	flag.Parse()

	fig := core.NewFigure(*width, *height)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.18},
		Max: geom.Pt{X: 0.95, Y: 0.88},
	})

	ax.XScale = transform.NewLinear(0, 10)
	ax.YScale = transform.NewLinear(-1, 1)
	ax.SetTitle("SVG Export Demo")
	ax.SetXLabel("x")
	ax.SetYLabel("y")

	// Add a line plot with a title and labels that should remain text in the output.
	n := 80
	x := make([]float64, n)
	y1 := make([]float64, n)
	y2 := make([]float64, n)
	for i := range x {
		x[i] = float64(i) * 10.0 / float64(n-1)
		y1[i] = 0.8 * (x[i] - 5)
		y2[i] = 0.5 * (1 - x[i]/5)
	}
	_, _ = ax.Plot(x, y1, core.PlotOptions{Label: "line 1"})
	_, _ = ax.Plot(x, y2, core.PlotOptions{Label: "line 2"})

	// Add a legend and a rotated annotation so all text paths remain native text.
	ax.Add(&core.Line2D{
		XY: []geom.Pt{
			{X: 0, Y: -0.8},
			{X: 2, Y: 0.2},
			{X: 8, Y: 0.2},
			{X: 10, Y: -0.8},
		},
		Col:   render.Color{R: 0.5, G: 0.1, B: 0.1, A: 1},
		W:     2,
		Label: "diag",
	})
	ax.Text(4.2, 0.4, "native SVG text")
	ax.AddLegend()
	ax.Annotate("marker", 2.5, 0.3, core.AnnotationOptions{
		HAlign: core.TextAlignCenter,
	})

	config := backends.Config{
		Width:      *width,
		Height:     *height,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		DPI:        96.0,
	}

	renderer, err := backends.Create(backends.SVG, config)
	if err != nil {
		fmt.Printf("Error creating SVG renderer: %v\n", err)
		os.Exit(1)
	}

	err = core.SaveSVG(fig, renderer, *output)
	if err != nil {
		fmt.Printf("Error saving SVG: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Saved %s\n", *output)
}
