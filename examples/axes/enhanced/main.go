package main

import (
	"fmt"
	"math"

	"matplotlib-go/backends/gobasic"
	"matplotlib-go/core"
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

func main() {
	// Create a figure with dimensions 1000x800
	fig := core.NewFigure(1000, 800)

	// Add axes with more space for labels
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.12},
		Max: geom.Pt{X: 0.95, Y: 0.88},
	})

	// Set up coordinate scales using the new convenience methods
	ax.SetXLim(-5, 5)
	ax.SetYLim(-3, 4)

	// Add grid lines (these will be drawn behind the data)
	xGrid := ax.AddXGrid()
	yGrid := ax.AddYGrid()

	// Customize grid appearance
	xGrid.Color = render.Color{R: 0.7, G: 0.7, B: 0.7, A: 1}
	yGrid.Color = render.Color{R: 0.7, G: 0.7, B: 0.7, A: 1}
	xGrid.LineWidth = 0.5
	yGrid.LineWidth = 0.5

	// Generate multiple mathematical functions
	n := 200
	x := make([]float64, n)
	for i := 0; i < n; i++ {
		x[i] = -5.0 + 10.0*float64(i)/float64(n-1)
	}

	// Function 1: Sine wave
	y1 := make([]float64, n)
	for i := 0; i < n; i++ {
		y1[i] = 2 * math.Sin(x[i])
	}

	line1 := &core.Line2D{
		XY:  make([]geom.Pt, n),
		W:   2.5,
		Col: render.Color{R: 0.8, G: 0.2, B: 0.2, A: 1}, // red
	}
	for i := 0; i < n; i++ {
		line1.XY[i] = geom.Pt{X: x[i], Y: y1[i]}
	}
	ax.Add(line1)

	// Function 2: Damped cosine
	y2 := make([]float64, n)
	for i := 0; i < n; i++ {
		y2[i] = math.Exp(-x[i]*x[i]/10) * math.Cos(3*x[i])
	}

	line2 := &core.Line2D{
		XY:     make([]geom.Pt, n),
		W:      2.0,
		Col:    render.Color{R: 0.2, G: 0.6, B: 0.2, A: 1}, // green
		Dashes: []float64{8, 4},                            // dashed line
	}
	for i := 0; i < n; i++ {
		line2.XY[i] = geom.Pt{X: x[i], Y: y2[i]}
	}
	ax.Add(line2)

	// Function 3: Polynomial
	y3 := make([]float64, n)
	for i := 0; i < n; i++ {
		y3[i] = 0.1*x[i]*x[i]*x[i] - 0.5*x[i] + 1
	}

	line3 := &core.Line2D{
		XY:  make([]geom.Pt, n),
		W:   2.0,
		Col: render.Color{R: 0.2, G: 0.2, B: 0.8, A: 1}, // blue
	}
	for i := 0; i < n; i++ {
		line3.XY[i] = geom.Pt{X: x[i], Y: y3[i]}
	}
	ax.Add(line3)

	// Add scatter points at critical points
	criticalX := []float64{-3, -1, 0, 1, 3}
	criticalY := make([]float64, len(criticalX))
	colors := []render.Color{
		{R: 1, G: 0.5, B: 0, A: 1}, // orange
		{R: 0.5, G: 0, B: 1, A: 1}, // purple
		{R: 1, G: 0, B: 0.5, A: 1}, // pink
		{R: 0, G: 1, B: 0.5, A: 1}, // cyan
		{R: 1, G: 1, B: 0, A: 1},   // yellow
	}

	for i, xi := range criticalX {
		criticalY[i] = 2 * math.Sin(xi)
	}

	scatter := &core.Scatter2D{
		XY:     make([]geom.Pt, len(criticalX)),
		Sizes:  []float64{10, 12, 15, 12, 10},
		Colors: colors,
		Marker: core.MarkerDiamond,
	}
	for i := 0; i < len(criticalX); i++ {
		scatter.XY[i] = geom.Pt{X: criticalX[i], Y: criticalY[i]}
	}
	ax.Add(scatter)

	// Customize axis appearance
	ax.XAxis.Color = render.Color{R: 0.2, G: 0.2, B: 0.2, A: 1} // dark gray
	ax.YAxis.Color = render.Color{R: 0.2, G: 0.2, B: 0.2, A: 1}
	ax.XAxis.LineWidth = 1.5
	ax.YAxis.LineWidth = 1.5
	ax.XAxis.TickSize = 6.0
	ax.YAxis.TickSize = 6.0

	// Create a GoBasic renderer with light background
	r := gobasic.New(1000, 800, render.Color{R: 0.98, G: 0.98, B: 0.98, A: 1})

	// Save as PNG
	err := core.SavePNG(fig, r, "axes_enhanced.png")
	if err != nil {
		fmt.Printf("Error saving PNG: %v\n", err)
		return
	}

	// Create a second example with logarithmic scales
	fig2 := core.NewFigure(1000, 800)
	ax2 := fig2.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.15, Y: 0.12},
		Max: geom.Pt{X: 0.95, Y: 0.88},
	})

	// Set logarithmic scales using convenience methods
	ax2.SetXLimLog(0.1, 1000, 10)
	ax2.SetYLimLog(1, 10000, 10)

	// Add grid for log plot
	ax2.AddXGrid()
	ax2.AddYGrid()

	// Generate exponential data
	nExp := 50
	xExp := make([]float64, nExp)
	yExp1 := make([]float64, nExp)
	yExp2 := make([]float64, nExp)

	for i := 0; i < nExp; i++ {
		t := float64(i) / float64(nExp-1)
		xExp[i] = 0.1 * math.Pow(10000, t)     // 0.1 to 1000
		yExp1[i] = 10 * math.Pow(xExp[i], 1.5) // power law
		yExp2[i] = 5 * math.Exp(0.01*xExp[i])  // exponential
	}

	// Power law line
	powerLine := &core.Line2D{
		XY:  make([]geom.Pt, nExp),
		W:   3.0,
		Col: render.Color{R: 0.8, G: 0.2, B: 0.2, A: 1}, // red
	}
	for i := 0; i < nExp; i++ {
		powerLine.XY[i] = geom.Pt{X: xExp[i], Y: yExp1[i]}
	}
	ax2.Add(powerLine)

	// Exponential line
	expLine := &core.Line2D{
		XY:     make([]geom.Pt, nExp),
		W:      3.0,
		Col:    render.Color{R: 0.2, G: 0.6, B: 0.8, A: 1}, // blue
		Dashes: []float64{10, 5},
	}
	for i := 0; i < nExp; i++ {
		expLine.XY[i] = geom.Pt{X: xExp[i], Y: yExp2[i]}
	}
	ax2.Add(expLine)

	// Save the logarithmic plot
	r2 := gobasic.New(1000, 800, render.Color{R: 1, G: 1, B: 1, A: 1})
	err = core.SavePNG(fig2, r2, "axes_logarithmic_enhanced.png")
	if err != nil {
		fmt.Printf("Error saving logarithmic PNG: %v\n", err)
		return
	}

	fmt.Println("Successfully created enhanced axis examples!")
	fmt.Println("- axes_enhanced.png: Multiple functions with grid lines and custom styling")
	fmt.Println("- axes_logarithmic_enhanced.png: Log-scale plot with power law and exponential data")
}
