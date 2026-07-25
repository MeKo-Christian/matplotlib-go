package main

import (
	"flag"
	"fmt"
	"math"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/transform"
)

func main() {
	output := flag.String("out", "agg_demo.png", "output PNG file")
	flag.Parse()

	// Match the Python counterpart's 8x5 inch, 100 dpi canvas as pixels.
	fig := core.NewFigure(800, 500)

	// The normalized rectangle mirrors fig.add_axes([0.12, 0.18, 0.83, 0.70]).
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.18},
		Max: geom.Pt{X: 0.95, Y: 0.88},
	})

	n := 200
	x := make([]float64, n)
	sinY := make([]float64, n)
	cosY := make([]float64, n)
	for i := 0; i < n; i++ {
		x[i] = float64(i) / float64(n-1) * 10.0
		sinY[i] = math.Sin(x[i])
		cosY[i] = math.Cos(x[i])
	}

	// Leave colors on the default cycle so Go and Python demonstrate the same behavior.
	_, _ = ax.Plot(x, sinY, core.PlotOptions{Label: "sin(x)"})
	_, _ = ax.Plot(x, cosY, core.PlotOptions{Label: "cos(x)"})

	ax.SetTitle("Sine and Cosine Waves")
	ax.SetXLabel("x (radians)")
	ax.SetYLabel("y")
	ax.XScale = transform.NewLinear(0, 10)
	ax.YScale = transform.NewLinear(-1.2, 1.2)
	ax.AddXGrid()
	ax.AddYGrid()

	// Request a text-capable raster backend; the environment can still override it.
	r, _, err := backends.NewRendererFromEnv(backends.Config{
		Width:      800,
		Height:     500,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		DPI:        100.0,
	}, backends.TextCapabilities)
	if err != nil {
		fmt.Printf("Error creating renderer: %v\n", err)
		return
	}

	err = core.SavePNG(fig, r, *output)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Created %s - sine/cosine plot with AGG anti-aliased rendering,\n", *output)
	fmt.Println("axis ticks, grid lines, and labels!")
}
