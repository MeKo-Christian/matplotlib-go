package main

import (
	"fmt"
	"math"

	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/transform"
)

func main() {
	// Match the Python reference: one linear plot plus one log-scaled plot.
	fig := core.NewFigure(800, 600)

	ax := fig.AddAxes(geom.Rect{
		// Matplotlib's add_axes([0.15, 0.15, 0.80, 0.70]) expressed as
		// normalized lower-left and upper-right corners.
		Min: geom.Pt{X: 0.15, Y: 0.15},
		Max: geom.Pt{X: 0.95, Y: 0.85},
	})

	ax.XScale = transform.NewLinear(0, 10)
	ax.YScale = transform.NewLinear(-2, 3)

	// Sample the same smooth signal used by the Python counterpart.
	n := 50
	x := make([]float64, n)
	y := make([]float64, n)
	for i := 0; i < n; i++ {
		t := 10.0 * float64(i) / float64(n-1)
		x[i] = t
		y[i] = math.Sin(t) + 0.5*math.Cos(2*t)
	}

	line := &core.Line2D{
		XY:  make([]geom.Pt, n),
		W:   2.5,
		Col: render.Color{R: 0.2, G: 0.4, B: 0.8, A: 1}, // blue
	}
	for i := 0; i < n; i++ {
		line.XY[i] = geom.Pt{X: x[i], Y: y[i]}
	}
	ax.Add(line)

	// Overlay a few exact samples to exercise the scatter path.
	scatterX := []float64{1, 3, 5, 7, 9}
	scatterY := make([]float64, len(scatterX))
	for i, xi := range scatterX {
		scatterY[i] = math.Sin(xi) + 0.5*math.Cos(2*xi)
	}

	scatter := &core.Scatter2D{
		XY:     make([]geom.Pt, len(scatterX)),
		Size:   8.0,
		Color:  render.Color{R: 0.8, G: 0.2, B: 0.2, A: 1}, // red
		Marker: core.MarkerCircle,
	}
	for i := 0; i < len(scatterX); i++ {
		scatter.XY[i] = geom.Pt{X: scatterX[i], Y: scatterY[i]}
	}
	ax.Add(scatter)

	if err := fig.Save("axes_basic.png"); err != nil {
		fmt.Printf("Error saving PNG: %v\n", err)
		return
	}

	// Second panel mirrors ax.set_xscale("log") / ax.set_yscale("log").
	fig2 := core.NewFigure(800, 600)
	ax2 := fig2.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.15, Y: 0.15},
		Max: geom.Pt{X: 0.95, Y: 0.85},
	})

	ax2.SetXLimLog(0.1, 100, 10)
	ax2.SetYLimLog(1, 1000, 10)

	// Keep spine/tick styling in step with Python's tick_params call.
	ax2.XAxis.Color = render.Color{R: 0.3, G: 0.3, B: 0.3, A: 1} // dark gray
	ax2.YAxis.Color = render.Color{R: 0.3, G: 0.3, B: 0.3, A: 1}
	ax2.XAxis.LineWidth = 1.5
	ax2.YAxis.LineWidth = 1.5
	ax2.XAxis.TickSize = 8.0
	ax2.YAxis.TickSize = 8.0

	nExp := 20
	xExp := make([]float64, nExp)
	yExp := make([]float64, nExp)
	for i := 0; i < nExp; i++ {
		t := float64(i) / float64(nExp-1)
		xExp[i] = 0.1 + (100-0.1)*t
		yExp[i] = 1 + math.Exp(5*t)
	}

	expLine := &core.Line2D{
		XY:  make([]geom.Pt, nExp),
		W:   3.0,
		Col: render.Color{R: 0.8, G: 0.5, B: 0.2, A: 1}, // orange
	}
	for i := 0; i < nExp; i++ {
		expLine.XY[i] = geom.Pt{X: xExp[i], Y: yExp[i]}
	}
	ax2.Add(expLine)

	if err := fig2.Save("axes_logarithmic.png"); err != nil {
		fmt.Printf("Error saving logarithmic PNG: %v\n", err)
		return
	}

	fmt.Println("Successfully created axis examples!")
	fmt.Println("- axes_basic.png: Line plot with automatic linear axes")
	fmt.Println("- axes_logarithmic.png: Exponential data with logarithmic tick placement")
}
