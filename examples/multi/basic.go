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
		Heights:     barY,
		Width:       0.3,
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

	// Create a second example with automatic color cycling using new convenience methods
	fig2 := core.NewFigure(1000, 700)
	ax2 := fig2.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	ax2.XScale = transform.NewLinear(0, 2*math.Pi)
	ax2.YScale = transform.NewLinear(-1.5, 1.5)

	// Generate multiple sine waves with different frequencies and phases using convenience methods
	nWaves := 5
	nPoints := 100
	xWave := make([]float64, nPoints)
	for i := 0; i < nPoints; i++ {
		xWave[i] = 2 * math.Pi * float64(i) / float64(nPoints-1)
	}

	for wave := 0; wave < nWaves; wave++ {
		frequency := float64(wave + 1)
		phase := float64(wave) * math.Pi / 4

		yWave := make([]float64, nPoints)
		for i := 0; i < nPoints; i++ {
			yWave[i] = math.Sin(frequency*xWave[i] + phase)
		}

		// Create line using new convenience method with automatic color cycling
		label := fmt.Sprintf("Wave %d", wave+1)
		ax2.Plot(xWave, yWave, core.PlotOptions{
			Label: label,
		})

		// Add some scatter points using convenience method
		scatterX := []float64{math.Pi / 2, math.Pi, 3 * math.Pi / 2}
		scatterY := make([]float64, len(scatterX))
		for i, x := range scatterX {
			scatterY[i] = math.Sin(frequency*x + phase)
		}

		ax2.Scatter(scatterX, scatterY, core.ScatterOptions{
			Label: label + " points",
		})
	}

	// Save the color cycling example
	r2 := gobasic.New(1000, 700, render.Color{R: 1, G: 1, B: 1, A: 1})
	err = core.SavePNG(fig2, r2, "multi_color_cycling.png")
	if err != nil {
		fmt.Printf("Error saving color cycling PNG: %v\n", err)
		return
	}

	// Create a third example showing convenience methods for different plot types
	fig3 := core.NewFigure(1000, 700)
	ax3 := fig3.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	ax3.XScale = transform.NewLinear(0, 10)
	ax3.YScale = transform.NewLinear(-2, 8)

	// Sample data
	xData := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9}
	
	// Line plot with automatic color cycling
	yLine := []float64{2, 3.5, 2.8, 4.1, 3.2, 4.8, 4.2, 5.1, 4.9}
	ax3.Plot(xData, yLine, core.PlotOptions{
		Label: "Trend line",
	})

	// Scatter plot with automatic color cycling
	xScatter := []float64{1.5, 2.5, 3.5, 4.5, 5.5, 6.5, 7.5, 8.5}
	yScatter := []float64{1.8, 2.9, 2.5, 3.8, 2.9, 4.5, 3.9, 4.8}
	marker := core.MarkerSquare
	ax3.Scatter(xScatter, yScatter, core.ScatterOptions{
		Label:  "Data points",
		Marker: &marker,
	})

	// Bar plot with automatic color cycling
	xBars := []float64{0.8, 1.8, 2.8, 3.8, 4.8}
	yBars := []float64{1.2, 1.8, 1.5, 2.1, 1.9}
	width := 0.3
	ax3.Bar(xBars, yBars, core.BarOptions{
		Label: "Baseline data",
		Width: &width,
	})

	// Fill area with automatic color cycling
	xFill := []float64{6, 7, 8, 9, 10}
	yFill := []float64{6, 7.2, 6.8, 7.5, 7.1}
	ax3.FillToBaselinePlot(xFill, yFill, core.FillOptions{
		Label: "Filled area",
	})

	// Save the convenience methods example
	r3 := gobasic.New(1000, 700, render.Color{R: 1, G: 1, B: 1, A: 1})
	err = core.SavePNG(fig3, r3, "multi_convenience_methods.png")
	if err != nil {
		fmt.Printf("Error saving convenience methods PNG: %v\n", err)
		return
	}

	fmt.Println("Successfully created multi-series examples!")
	fmt.Println("- multi_basic.png: Mixed plot types (manual approach)")
	fmt.Println("- multi_color_cycling.png: Multiple series with automatic color cycling")
	fmt.Println("- multi_convenience_methods.png: Different plot types using convenience methods")
}
