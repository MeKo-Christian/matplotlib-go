package main

import (
	"fmt"

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
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	// Set up coordinate scales
	ax.XScale = transform.NewLinear(0, 6)
	ax.YScale = transform.NewLinear(0, 10)

	// Example 1: Simple vertical bar chart
	categories := []float64{1, 2, 3, 4, 5}
	values := []float64{3, 7, 2, 9, 4}

	verticalBars := &core.Bar2D{
		X:           categories,
		Heights:     values,
		Width:       0.6,
		Color:       render.Color{R: 0.2, G: 0.6, B: 0.8, A: 1}, // blue
		Baseline:    0,
		Orientation: core.BarVertical,
	}
	ax.Add(verticalBars)

	// Create a GoBasic renderer with white background
	r := gobasic.New(800, 600, render.Color{R: 1, G: 1, B: 1, A: 1})

	// Save as PNG
	err := core.SavePNG(fig, r, "bar_basic.png")
	if err != nil {
		fmt.Printf("Error saving PNG: %v\n", err)
		return
	}

	// Create a second figure for horizontal bars
	fig2 := core.NewFigure(800, 600)
	ax2 := fig2.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	// Set up coordinate scales for horizontal bars
	ax2.XScale = transform.NewLinear(0, 10)
	ax2.YScale = transform.NewLinear(0, 6)

	// Example 2: Horizontal bar chart
	horizontalBars := &core.Bar2D{
		X:           categories,                                  // Y positions for horizontal bars
		Heights:     values,                                     // Bar lengths
		Width:       0.6,                                        // Bar thickness
		Color:       render.Color{R: 0.8, G: 0.4, B: 0.2, A: 1}, // orange
		Baseline:    0,
		Orientation: core.BarHorizontal,
	}
	ax2.Add(horizontalBars)

	// Save horizontal bars
	r2 := gobasic.New(800, 600, render.Color{R: 1, G: 1, B: 1, A: 1})
	err = core.SavePNG(fig2, r2, "bar_horizontal.png")
	if err != nil {
		fmt.Printf("Error saving horizontal bars PNG: %v\n", err)
		return
	}

	// Create a third figure for variable colors and widths
	fig3 := core.NewFigure(800, 600)
	ax3 := fig3.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	ax3.XScale = transform.NewLinear(0, 6)
	ax3.YScale = transform.NewLinear(0, 10)

	// Example 3: Variable colors and widths
	varColors := []render.Color{
		{R: 1, G: 0.2, B: 0.2, A: 1}, // red
		{R: 0.2, G: 1, B: 0.2, A: 1}, // green
		{R: 0.2, G: 0.2, B: 1, A: 1}, // blue
		{R: 1, G: 1, B: 0.2, A: 1},   // yellow
		{R: 1, G: 0.2, B: 1, A: 1},   // magenta
	}

	varWidths := []float64{0.4, 0.8, 0.3, 0.9, 0.5}

	variableBars := &core.Bar2D{
		X:           categories,
		Heights:     values,
		Widths:      varWidths,
		Colors:      varColors,
		Width:       0.6,                                        // fallback width
		Color:       render.Color{R: 0.5, G: 0.5, B: 0.5, A: 1}, // fallback color
		Baseline:    0,
		Orientation: core.BarVertical,
	}
	ax3.Add(variableBars)

	// Save variable bars
	r3 := gobasic.New(800, 600, render.Color{R: 1, G: 1, B: 1, A: 1})
	err = core.SavePNG(fig3, r3, "bar_variable.png")
	if err != nil {
		fmt.Printf("Error saving variable bars PNG: %v\n", err)
		return
	}

	fmt.Println("Successfully created bar chart examples!")
	fmt.Println("- bar_basic.png: Simple vertical bars")
	fmt.Println("- bar_horizontal.png: Horizontal bars")
	fmt.Println("- bar_variable.png: Variable colors and widths")
}
