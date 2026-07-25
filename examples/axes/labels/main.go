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
	// Focus on title and axis-label placement, including the rotated y-label.
	fig := core.NewFigure(800, 500)

	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.15},
		Max: geom.Pt{X: 0.95, Y: 0.88},
	})

	ax.XScale = transform.NewLinear(0, 2*math.Pi)
	ax.YScale = transform.NewLinear(-1.5, 1.5)

	ax.SetTitle("Text Labels: Title, X-Label, Rotated Y-Label")
	ax.SetXLabel("Angle (radians)")
	ax.SetYLabel("Amplitude")

	// Minor ticks mirror Matplotlib's minorticks_on() call.
	ax.XAxis.MinorLocator = ticker.MinorLinearLocator{N: 5}
	ax.YAxis.MinorLocator = ticker.MinorLinearLocator{N: 4}

	ax.AddXGrid()
	ax.AddYGrid()

	// Single sine curve keeps attention on the label and tick text rendering.
	n := 200
	x := make([]float64, n)
	y := make([]float64, n)
	for i := range n {
		x[i] = 2 * math.Pi * float64(i) / float64(n-1)
		y[i] = math.Sin(x[i])
	}

	_, _ = ax.Plot(x, y, core.PlotOptions{Label: "sin(x)"})

	if err := fig.Save("labels.png"); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Created labels.png — title, x-label, rotated y-label, and tick labels")
}
