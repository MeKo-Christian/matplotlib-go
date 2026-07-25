package main

import (
	"fmt"
	"math"

	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/ticker"
	"github.com/cwbudde/matplotlib-go/transform"
)

func main() {
	// Exercise visible spines with both major and minor ticks enabled.
	fig := core.NewFigure(800, 500)

	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.15},
		Max: geom.Pt{X: 0.95, Y: 0.88},
	})

	ax.XScale = transform.NewLinear(0, 2*math.Pi)
	ax.YScale = transform.NewLinear(-1.5, 1.5)

	ax.SetTitle("Axis Spines, Major & Minor Ticks")
	ax.SetXLabel("x")
	ax.SetYLabel("y")

	// Minor locators mirror Matplotlib's minorticks_on() behavior.
	ax.XAxis.MinorLocator = ticker.MinorLinearLocator{N: 5}
	ax.YAxis.MinorLocator = ticker.MinorLinearLocator{N: 4}

	// Add grid for major ticks only.
	ax.AddXGrid()
	ax.AddYGrid()

	// Plot sine and cosine over the same sampled domain.
	n := 200
	x := make([]float64, n)
	sinY := make([]float64, n)
	cosY := make([]float64, n)
	for i := range n {
		x[i] = 2 * math.Pi * float64(i) / float64(n-1)
		sinY[i] = math.Sin(x[i])
		cosY[i] = math.Cos(x[i])
	}

	ax.Plot(x, sinY, core.PlotOptions{Label: "sin(x)"})
	ax.Plot(x, cosY, core.PlotOptions{Label: "cos(x)"})

	if err := fig.Save("spines.png"); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Created spines.png — axis spines with major and minor ticks")
}
