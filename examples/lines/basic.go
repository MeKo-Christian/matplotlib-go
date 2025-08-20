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
	// Create a figure with dimensions 640x360
	fig := core.NewFigure(640, 360)

	// Add axes that take up most of the figure space
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.15},
		Max: geom.Pt{X: 0.95, Y: 0.9},
	})

	// Set up coordinate scales
	ax.XScale = transform.NewLinear(0, 10)
	ax.YScale = transform.NewLinear(0, 1)

	// Create a line with some sample data (diagonal line)
	line := &core.Line2D{
		XY: []geom.Pt{
			{X: 0, Y: 0},
			{X: 1, Y: 0.2},
			{X: 3, Y: 0.9},
			{X: 6, Y: 0.4},
			{X: 10, Y: 0.8},
		},
		W:   2.0,
		Col: render.Color{R: 0, G: 0, B: 0, A: 1}, // black line
	}

	// Add the line to the axes
	ax.Add(line)

	// Create a GoBasic renderer with white background
	r := gobasic.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})

	// Save as PNG
	err := core.SavePNG(fig, r, "out.png")
	if err != nil {
		fmt.Printf("Error saving PNG: %v\n", err)
		return
	}

	fmt.Println("Successfully created out.png with a line plot!")
}
