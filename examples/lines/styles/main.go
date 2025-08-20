// Package demonstrates line join and cap styles with matplotlib-go.
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
	ax.XScale = transform.NewLinear(0, 12)
	ax.YScale = transform.NewLinear(0, 8)

	// Create L-shaped paths to demonstrate different line joins
	createJoinDemo(ax)

	// Create straight lines to demonstrate different line caps
	createCapDemo(ax)

	// Create a GoBasic renderer with white background
	r := gobasic.New(800, 600, render.Color{R: 1, G: 1, B: 1, A: 1})

	// Render the figure
	core.DrawFigure(fig, r)

	// Save to PNG
	err := r.SavePNG("examples/lines/styles.png")
	if err != nil {
		log.Fatalf("Failed to save PNG: %v", err)
	}

	log.Println("Saved line styles demo to examples/lines/styles.png")
}

// createJoinDemo adds lines demonstrating different line join styles
func createJoinDemo(ax *core.Axes) {
	// Base L-shaped path
	basePath := []geom.Pt{
		{X: 1, Y: 6}, {X: 3, Y: 6}, {X: 3, Y: 4},
	}

	// Miter join (red)
	miterPath := make([]geom.Pt, len(basePath))
	copy(miterPath, basePath)
	miterLine := &core.Line2D{
		XY:  miterPath,
		W:   12.0,
		Col: render.Color{R: 0.8, G: 0.2, B: 0.2, A: 1},
	}
	ax.Add(miterLine)

	// Round join (green) - offset to the right
	roundPath := make([]geom.Pt, len(basePath))
	for i, pt := range basePath {
		roundPath[i] = geom.Pt{X: pt.X + 3, Y: pt.Y}
	}
	roundLine := &core.Line2D{
		XY:  roundPath,
		W:   12.0,
		Col: render.Color{R: 0.2, G: 0.8, B: 0.2, A: 1},
	}
	ax.Add(roundLine)

	// Bevel join (blue) - offset further right
	bevelPath := make([]geom.Pt, len(basePath))
	for i, pt := range basePath {
		bevelPath[i] = geom.Pt{X: pt.X + 6, Y: pt.Y}
	}
	bevelLine := &core.Line2D{
		XY:  bevelPath,
		W:   12.0,
		Col: render.Color{R: 0.2, G: 0.2, B: 0.8, A: 1},
	}
	ax.Add(bevelLine)
}

// createCapDemo adds lines demonstrating different line cap styles
func createCapDemo(ax *core.Axes) {
	// Butt cap (red)
	buttPath := []geom.Pt{
		{X: 1, Y: 2}, {X: 3, Y: 2},
	}
	buttLine := &core.Line2D{
		XY:  buttPath,
		W:   12.0,
		Col: render.Color{R: 0.8, G: 0.2, B: 0.2, A: 1},
	}
	ax.Add(buttLine)

	// Round cap (green)
	roundPath := []geom.Pt{
		{X: 4, Y: 2}, {X: 6, Y: 2},
	}
	roundLine := &core.Line2D{
		XY:  roundPath,
		W:   12.0,
		Col: render.Color{R: 0.2, G: 0.8, B: 0.2, A: 1},
	}
	ax.Add(roundLine)

	// Square cap (blue)
	squarePath := []geom.Pt{
		{X: 7, Y: 2}, {X: 9, Y: 2},
	}
	squareLine := &core.Line2D{
		XY:  squarePath,
		W:   12.0,
		Col: render.Color{R: 0.2, G: 0.2, B: 0.8, A: 1},
	}
	ax.Add(squareLine)
}
