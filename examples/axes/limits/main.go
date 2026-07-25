package main

import (
	"fmt"
	"math"

	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
)

func main() {
	// Leave limits unset until both lines are present, then autoscale with a
	// Matplotlib-style fractional margin.
	fig := core.NewFigure(800, 500)

	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.15},
		Max: geom.Pt{X: 0.95, Y: 0.88},
	})

	ax.SetTitle("Auto-Scaled Axes with 5% Margin")
	ax.SetXLabel("x")
	ax.SetYLabel("y")

	n := 100
	x := make([]float64, n)
	y1 := make([]float64, n)
	y2 := make([]float64, n)
	for i := range n {
		x[i] = 3.0 + 4.0*float64(i)/float64(n-1) // x in [3, 7]
		y1[i] = math.Sin(x[i]) * 2.5             // y roughly in [-2.5, 2.5]
		y2[i] = math.Cos(x[i])*1.5 + 0.5         // shifted cosine
	}

	_, _ = ax.Plot(x, y1, core.PlotOptions{Label: "2.5·sin(x)"})
	_, _ = ax.Plot(x, y2, core.PlotOptions{Label: "1.5·cos(x)+0.5"})

	// Auto-scale with 5% margin; no need to specify XScale/YScale manually.
	ax.AutoScale(0.05)

	ax.AddXGrid()
	ax.AddYGrid()

	if err := fig.Save("limits.png"); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Created limits.png — auto-scaled axes with margin")
}
