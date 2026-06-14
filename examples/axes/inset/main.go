package main

import (
	"log"
	"math"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func main() {
	// Main axes plus a zoomed inset covering the same data range as Python's
	// ax.inset_axes([0.58, 0.55, 0.36, 0.38]).
	fig := core.NewFigure(720, 420)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.14},
		Max: geom.Pt{X: 0.92, Y: 0.86},
	})
	ax.SetTitle("Inset Axes")
	ax.SetXLabel("x")
	ax.SetYLabel("sin(x)")
	ax.SetXLim(0, 10)
	ax.SetYLim(-1.2, 1.2)

	x, y := wave(400, 0, 10)
	lineColor := render.Color{R: 0.12, G: 0.36, B: 0.72, A: 1}
	lineWidth := 2.0
	ax.Plot(x, y, core.PlotOptions{Color: &lineColor, LineWidth: &lineWidth})
	ax.AddGrid(core.AxisBottom)
	ax.AddGrid(core.AxisLeft)

	inset, _ := ax.ZoomedInset(
		geom.Rect{
			Min: geom.Pt{X: 0.58, Y: 0.55},
			Max: geom.Pt{X: 0.94, Y: 0.93},
		},
		[2]float64{2, 4},
		[2]float64{-0.2, 1.05},
	)
	inset.SetTitle("detail")
	inset.Plot(x, y, core.PlotOptions{Color: &lineColor, LineWidth: &lineWidth})
	inset.AddGrid(core.AxisBottom)
	inset.AddGrid(core.AxisLeft)

	r, err := agg.New(720, 420, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		log.Fatal(err)
	}
	core.DrawFigure(fig, r)
	if err := r.SavePNG("inset.png"); err != nil {
		log.Fatal(err)
	}
}

func wave(n int, minX, maxX float64) ([]float64, []float64) {
	x := make([]float64, n)
	y := make([]float64, n)
	for i := range n {
		t := float64(i) / float64(n-1)
		x[i] = minX + (maxX-minX)*t
		y[i] = math.Sin(x[i]) * (0.75 + 0.2*math.Cos(2*x[i]))
	}
	return x, y
}
