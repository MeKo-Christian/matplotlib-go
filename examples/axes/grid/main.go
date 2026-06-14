package main

import (
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
	// Demonstrate major and minor grid rendering on the same pair of curves.
	fig := core.NewFigure(800, 500)

	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.15},
		Max: geom.Pt{X: 0.95, Y: 0.88},
	})

	ax.XScale = transform.NewLinear(0, 10)
	ax.YScale = transform.NewLinear(-1.5, 1.5)

	ax.SetTitle("Grid Lines: Major (solid) & Minor (dashed)")
	ax.SetXLabel("x")
	ax.SetYLabel("y")

	// Major grid lines are solid.
	xGrid := ax.AddXGrid()
	xGrid.Color = render.Color{R: 0.7, G: 0.7, B: 0.7, A: 0.8}
	xGrid.LineWidth = 0.8

	yGrid := ax.AddYGrid()
	yGrid.Color = render.Color{R: 0.7, G: 0.7, B: 0.7, A: 0.8}
	yGrid.LineWidth = 0.8

	// Minor grid lines share the axes locators but use a lighter dashed style.
	xGrid.Minor = true
	xGrid.MinorDashes = []float64{2, 3}
	xGrid.MinorColor = render.Color{R: 0.85, G: 0.85, B: 0.85, A: 0.6}

	yGrid.Minor = true
	yGrid.MinorDashes = []float64{2, 3}
	yGrid.MinorColor = render.Color{R: 0.85, G: 0.85, B: 0.85, A: 0.6}

	ax.XAxis.MinorLocator = core.MinorLinearLocator{N: 5}
	ax.YAxis.MinorLocator = core.MinorLinearLocator{N: 5}

	// Plot the two waves after the grid setup so the example reads like the
	// Python reference: configure axes, then draw data.
	n := 200
	x := make([]float64, n)
	y1 := make([]float64, n)
	y2 := make([]float64, n)
	for i := range n {
		x[i] = 10.0 * float64(i) / float64(n-1)
		y1[i] = math.Sin(x[i])
		y2[i] = 0.7 * math.Sin(2*x[i]+0.5)
	}

	ax.Plot(x, y1, core.PlotOptions{Label: "sin(x)"})
	ax.Plot(x, y2, core.PlotOptions{Label: "0.7·sin(2x+0.5)"})

	r, _, err := backends.NewRendererFromEnv(backends.Config{
		Width:      800,
		Height:     500,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		DPI:        72.0,
	}, backends.TextCapabilities)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if err := core.SavePNG(fig, r, "grid.png"); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Created grid.png — major and minor grid lines with dashes")
}
