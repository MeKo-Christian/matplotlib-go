package main

import (
	"fmt"
	"math"

	"matplotlib-go/backends/gobasic"
	"matplotlib-go/core"
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
	"matplotlib-go/transform"
)

func main() {
	// Create a figure with dimensions 800x600
	fig := core.NewFigure(800, 600)

	// Add axes that take up most of the figure space
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.15, Y: 0.15}, // Leave more space for axes
		Max: geom.Pt{X: 0.95, Y: 0.85},
	})

	// Set up coordinate scales
	ax.XScale = transform.NewLinear(0, 10)
	ax.YScale = transform.NewLinear(-2, 3)

	// Create some sample data
	n := 50
	x := make([]float64, n)
	y := make([]float64, n)
	for i := 0; i < n; i++ {
		t := 10.0 * float64(i) / float64(n-1)
		x[i] = t
		y[i] = math.Sin(t) + 0.5*math.Cos(2*t)
	}

	// Add a line plot
	line := &core.Line2D{
		XY:  make([]geom.Pt, n),
		W:   2.5,
		Col: render.Color{R: 0.2, G: 0.4, B: 0.8, A: 1}, // blue
	}
	for i := 0; i < n; i++ {
		line.XY[i] = geom.Pt{X: x[i], Y: y[i]}
	}
	ax.Add(line)

	// Add some scatter points
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

	// The axes should automatically be drawn with default settings

	// Create a GoBasic renderer with white background
	r := gobasic.New(800, 600, render.Color{R: 1, G: 1, B: 1, A: 1})

	// Save as PNG
	err := core.SavePNG(fig, r, "axes_basic.png")
	if err != nil {
		fmt.Printf("Error saving PNG: %v\n", err)
		return
	}

	// Create a second example with custom axis settings
	fig2 := core.NewFigure(800, 600)
	ax2 := fig2.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.15, Y: 0.15},
		Max: geom.Pt{X: 0.95, Y: 0.85},
	})

	// Set up scales for logarithmic data
	ax2.XScale = transform.NewLinear(0.1, 100)
	ax2.YScale = transform.NewLinear(1, 1000)

	// Use logarithmic locators
	ax2.XAxis.Locator = core.LogLocator{Base: 10, Minor: false}
	ax2.YAxis.Locator = core.LogLocator{Base: 10, Minor: false}

	// Custom axis styling
	ax2.XAxis.Color = render.Color{R: 0.3, G: 0.3, B: 0.3, A: 1} // dark gray
	ax2.YAxis.Color = render.Color{R: 0.3, G: 0.3, B: 0.3, A: 1}
	ax2.XAxis.LineWidth = 1.5
	ax2.YAxis.LineWidth = 1.5
	ax2.XAxis.TickSize = 8.0
	ax2.YAxis.TickSize = 8.0

	// Generate exponential data
	nExp := 20
	xExp := make([]float64, nExp)
	yExp := make([]float64, nExp)
	for i := 0; i < nExp; i++ {
		t := float64(i) / float64(nExp-1)
		xExp[i] = 0.1 + (100-0.1)*t
		yExp[i] = 1 + math.Exp(5*t)
	}

	// Add exponential line
	expLine := &core.Line2D{
		XY:  make([]geom.Pt, nExp),
		W:   3.0,
		Col: render.Color{R: 0.8, G: 0.5, B: 0.2, A: 1}, // orange
	}
	for i := 0; i < nExp; i++ {
		expLine.XY[i] = geom.Pt{X: xExp[i], Y: yExp[i]}
	}
	ax2.Add(expLine)

	// Save the logarithmic example
	r2 := gobasic.New(800, 600, render.Color{R: 1, G: 1, B: 1, A: 1})
	err = core.SavePNG(fig2, r2, "axes_logarithmic.png")
	if err != nil {
		fmt.Printf("Error saving logarithmic PNG: %v\n", err)
		return
	}

	fmt.Println("Successfully created axis examples!")
	fmt.Println("- axes_basic.png: Line plot with automatic linear axes")
	fmt.Println("- axes_logarithmic.png: Exponential data with logarithmic tick placement")
}
