// Package demonstrates dash patterns with matplotlib-go.
package main

import (
	"log"

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
	ax.XScale = transform.NewLinear(0, 10)
	ax.YScale = transform.NewLinear(0, 8)

	// Create horizontal lines with different dash patterns
	dashStyles := []struct {
		y      float64
		dashes []float64
		color  render.Color
		label  string
	}{
		{7, []float64{}, render.Color{R: 0, G: 0, B: 0, A: 1}, "Solid"},
		{6, []float64{8, 4}, render.Color{R: 0.8, G: 0, B: 0, A: 1}, "Basic dash"},
		{5, []float64{4, 2, 4, 2}, render.Color{R: 0, G: 0.6, B: 0, A: 1}, "Even dash"},
		{4, []float64{6, 2, 1, 2}, render.Color{R: 0, G: 0, B: 0.8, A: 1}, "Dash-dot"},
		{3, []float64{3, 1, 1, 1, 1, 1}, render.Color{R: 0.8, G: 0.4, B: 0, A: 1}, "Dash-dot-dot"},
		{2, []float64{1, 1}, render.Color{R: 0.6, G: 0, B: 0.6, A: 1}, "Dotted"},
		{1, []float64{10, 2, 2, 2, 2, 2}, render.Color{R: 0, G: 0.6, B: 0.6, A: 1}, "Long dash-dots"},
	}

	for _, style := range dashStyles {
		// Create horizontal line
		path := []geom.Pt{
			{X: 1, Y: style.y}, {X: 9, Y: style.y},
		}

		line := &core.Line2D{
			XY:     path,
			W:      4.0,
			Col:    style.color,
			Dashes: style.dashes,
		}
		ax.Add(line)
	}

	// Add a curved path to show dashes on curves
	curvePath := []geom.Pt{
		{X: 1, Y: 0.5},
		{X: 2, Y: 0.8},
		{X: 3, Y: 0.2},
		{X: 4, Y: 0.9},
		{X: 5, Y: 0.1},
		{X: 6, Y: 0.7},
		{X: 7, Y: 0.3},
		{X: 8, Y: 0.6},
		{X: 9, Y: 0.4},
	}

	curveLine := &core.Line2D{
		XY:     curvePath,
		W:      3.0,
		Col:    render.Color{R: 0.5, G: 0.5, B: 0.5, A: 1},
		Dashes: []float64{5, 3},
	}
	ax.Add(curveLine)

	// Create a GoBasic renderer with white background
	r := gobasic.New(800, 600, render.Color{R: 1, G: 1, B: 1, A: 1})

	// Render the figure
	core.DrawFigure(fig, r)

	// Save to PNG
	err := r.SavePNG("examples/lines/dash/dash_patterns.png")
	if err != nil {
		log.Fatalf("Failed to save PNG: %v", err)
	}

	log.Println("Saved dash patterns demo to examples/lines/dash/dash_patterns.png")
}
