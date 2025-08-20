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
	// Create a figure with dimensions 1000x700
	fig := core.NewFigure(1000, 700)

	// Add axes that take up most of the figure space
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	// Set up coordinate scales
	ax.XScale = transform.NewLinear(0, 10)
	ax.YScale = transform.NewLinear(-2, 8)

	// Generate some sample data
	n := 50
	x := make([]float64, n)
	for i := 0; i < n; i++ {
		x[i] = 10.0 * float64(i) / float64(n-1)
	}

	// Data series 1: Exponential decay with noise
	y1 := make([]float64, n)
	for i := 0; i < n; i++ {
		y1[i] = 4*math.Exp(-x[i]/3) + 0.2*math.Sin(x[i]*2) + 1
	}

	// Data series 2: Linear trend
	y2 := make([]float64, n)
	for i := 0; i < n; i++ {
		y2[i] = 0.3*x[i] + 0.5
	}

	// Data series 3: Oscillating function
	y3 := make([]float64, n)
	for i := 0; i < n; i++ {
		y3[i] = 2*math.Sin(x[i]*0.8) + 3
	}

	// Background fill - area under exponential decay
	backgroundFill := core.FillToBaseline(x, y1, 0,
		render.Color{R: 0.9, G: 0.9, B: 0.95, A: 0.6}) // very light blue
	ax.Add(backgroundFill)

	// Line plot 1: Exponential decay
	line1 := &core.Line2D{
		XY:  make([]geom.Pt, n),
		W:   3.0,
		Col: render.Color{R: 0.8, G: 0.2, B: 0.2, A: 1}, // red
	}
	for i := 0; i < n; i++ {
		line1.XY[i] = geom.Pt{X: x[i], Y: y1[i]}
	}
	ax.Add(line1)

	// Line plot 2: Linear trend with dashes
	line2 := &core.Line2D{
		XY:     make([]geom.Pt, n),
		W:      2.5,
		Col:    render.Color{R: 0.2, G: 0.6, B: 0.2, A: 1}, // green
		Dashes: []float64{5, 3},                            // dash pattern
	}
	for i := 0; i < n; i++ {
		line2.XY[i] = geom.Pt{X: x[i], Y: y2[i]}
	}
	ax.Add(line2)

	// Line plot 3: Oscillating function
	line3 := &core.Line2D{
		XY:  make([]geom.Pt, n),
		W:   2.0,
		Col: render.Color{R: 0.2, G: 0.2, B: 0.8, A: 1}, // blue
	}
	for i := 0; i < n; i++ {
		line3.XY[i] = geom.Pt{X: x[i], Y: y3[i]}
	}
	ax.Add(line3)

	// Scatter plot: Sample some points from the oscillating function
	scatterX := []float64{1, 3, 5, 7, 9}
	scatterY := make([]float64, len(scatterX))
	for i, xi := range scatterX {
		scatterY[i] = 2*math.Sin(xi*0.8) + 3
	}

	scatter := &core.Scatter2D{
		XY:     make([]geom.Pt, len(scatterX)),
		Size:   12.0,
		Color:  render.Color{R: 0.9, G: 0.5, B: 0.1, A: 1}, // orange
		Marker: core.MarkerDiamond,
	}
	for i := 0; i < len(scatterX); i++ {
		scatter.XY[i] = geom.Pt{X: scatterX[i], Y: scatterY[i]}
	}
	ax.Add(scatter)

	// Bar chart: Add some bars at specific x positions
	barX := []float64{0.5, 2.5, 4.5, 6.5, 8.5}
	barY := []float64{6, 5.5, 7, 6.5, 5.8}

	bars := &core.Bar2D{
		X:           barX,
		Y:           barY,
		BarWidth:    0.3,
		Color:       render.Color{R: 0.7, G: 0.3, B: 0.9, A: 0.7}, // purple
		Baseline:    4.5,                                          // bars start from y=4.5
		Orientation: core.BarVertical,
	}
	ax.Add(bars)

	// Create a GoBasic renderer with light gray background
	r := gobasic.New(1000, 700, render.Color{R: 0.98, G: 0.98, B: 0.98, A: 1})

	// Save as PNG
	err := core.SavePNG(fig, r, "multi_basic.png")
	if err != nil {
		fmt.Printf("Error saving PNG: %v\n", err)
		return
	}

	// Create a second example with automatic color cycling
	fig2 := core.NewFigure(1000, 700)
	ax2 := fig2.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	ax2.XScale = transform.NewLinear(0, 2*math.Pi)
	ax2.YScale = transform.NewLinear(-1.5, 1.5)

	// Generate multiple sine waves with different frequencies and phases
	nWaves := 5
	nPoints := 100
	xWave := make([]float64, nPoints)
	for i := 0; i < nPoints; i++ {
		xWave[i] = 2 * math.Pi * float64(i) / float64(nPoints-1)
	}

	// Default color palette (similar to matplotlib's default colors)
	colors := []render.Color{
		{R: 0.12, G: 0.47, B: 0.71, A: 1}, // blue
		{R: 1.00, G: 0.50, B: 0.05, A: 1}, // orange
		{R: 0.17, G: 0.63, B: 0.17, A: 1}, // green
		{R: 0.84, G: 0.15, B: 0.16, A: 1}, // red
		{R: 0.58, G: 0.40, B: 0.74, A: 1}, // purple
	}

	for wave := 0; wave < nWaves; wave++ {
		frequency := float64(wave + 1)
		phase := float64(wave) * math.Pi / 4

		yWave := make([]float64, nPoints)
		for i := 0; i < nPoints; i++ {
			yWave[i] = math.Sin(frequency*xWave[i] + phase)
		}

		// Create line for this wave
		lineWave := &core.Line2D{
			XY:  make([]geom.Pt, nPoints),
			W:   2.0,
			Col: colors[wave%len(colors)], // cycle through colors
		}
		for i := 0; i < nPoints; i++ {
			lineWave.XY[i] = geom.Pt{X: xWave[i], Y: yWave[i]}
		}
		ax2.Add(lineWave)

		// Add some scatter points for each wave
		scatterWave := &core.Scatter2D{
			XY: []geom.Pt{
				{X: math.Pi / 2, Y: math.Sin(frequency*math.Pi/2 + phase)},
				{X: math.Pi, Y: math.Sin(frequency*math.Pi + phase)},
				{X: 3 * math.Pi / 2, Y: math.Sin(frequency*3*math.Pi/2 + phase)},
			},
			Size:   8.0,
			Color:  colors[wave%len(colors)],
			Marker: core.MarkerCircle,
		}
		ax2.Add(scatterWave)
	}

	// Save the color cycling example
	r2 := gobasic.New(1000, 700, render.Color{R: 1, G: 1, B: 1, A: 1})
	err = core.SavePNG(fig2, r2, "multi_color_cycling.png")
	if err != nil {
		fmt.Printf("Error saving color cycling PNG: %v\n", err)
		return
	}

	fmt.Println("Successfully created multi-series examples!")
	fmt.Println("- multi_basic.png: Mixed plot types (lines, scatter, bars, fill)")
	fmt.Println("- multi_color_cycling.png: Multiple series with automatic color cycling")
}
