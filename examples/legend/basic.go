package main

import (
	"fmt"
	"math"

	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
)

func main() {
	fig := core.NewFigure(1000, 700)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.11, Y: 0.12},
		Max: geom.Pt{X: 0.94, Y: 0.88},
	})

	const n = 120
	x := make([]float64, n)
	sine := make([]float64, n)
	cosine := make([]float64, n)
	for i := 0; i < n; i++ {
		xv := 2 * math.Pi * float64(i) / float64(n-1)
		x[i] = xv
		sine[i] = math.Sin(xv)
		cosine[i] = 0.7 * math.Cos(2*xv)
	}

	// Legend entries are driven by artist labels, matching Matplotlib's legend().
	ax.Plot(x, sine, core.PlotOptions{Label: "sin(x)"})
	ax.Plot(x, cosine, core.PlotOptions{
		Label:  "0.7 cos(2x)",
		Dashes: []float64{8, 5},
	})
	ax.Scatter(
		[]float64{0.6, 1.9, 3.4, 5.1},
		[]float64{0.56, 0.95, -0.26, -0.93},
		core.ScatterOptions{Label: "samples"},
	)

	ax.SetTitle("Legend Entries")
	ax.SetXLabel("x")
	ax.SetYLabel("y")
	ax.SetXLim(0, 2*math.Pi)
	ax.SetYLim(-1.3, 1.3)
	ax.AddXGrid()
	ax.AddYGrid()

	// Place the legend in the same corner as loc="upper left".
	legend := ax.AddLegend()
	legend.Location = core.LegendUpperLeft

	if err := fig.Save("legend_basic.png"); err != nil {
		fmt.Printf("error saving PNG: %v\n", err)
		return
	}

	fmt.Println("saved legend_basic.png")
}
