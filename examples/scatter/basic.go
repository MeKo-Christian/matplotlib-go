package main

import (
	"fmt"
	"math"
	"math/rand"

	"matplotlib-go/backends/gobasic"
	"matplotlib-go/core"
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
	"matplotlib-go/transform"
)

func main() {
	// Create a figure with dimensions 640x480
	fig := core.NewFigure(640, 480)

	// Add axes that take up most of the figure space
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	// Set up coordinate scales
	ax.XScale = transform.NewLinear(-2, 2)
	ax.YScale = transform.NewLinear(-2, 2)

	// Example 1: Basic scatter with circles
	basicPoints := []geom.Pt{
		{X: -1.5, Y: 1.5},
		{X: -1.0, Y: 1.0},
		{X: -0.5, Y: 0.8},
		{X: 0.0, Y: 0.0},
		{X: 0.5, Y: -0.8},
		{X: 1.0, Y: -1.0},
		{X: 1.5, Y: -1.5},
	}

	scatter1 := &core.Scatter2D{
		XY:        basicPoints,
		Size:      8.0,
		Color:     render.Color{R: 1, G: 0, B: 0, A: 1}, // red
		EdgeColor: render.Color{R: 0.5, G: 0, B: 0, A: 1}, // dark red edge
		EdgeWidth: 1.5,
		Marker:    core.MarkerCircle,
		Alpha:     1.0,
	}
	ax.Add(scatter1)

	// Example 2: Different marker shapes
	markerTypes := []core.MarkerType{
		core.MarkerSquare, core.MarkerTriangle, core.MarkerDiamond,
		core.MarkerPlus, core.MarkerCross,
	}

	for i, markerType := range markerTypes {
		x := -1.5 + float64(i)*0.75
		scatter := &core.Scatter2D{
			XY:        []geom.Pt{{X: x, Y: -1.8}},
			Size:      10.0,
			Color:     render.Color{R: 0, G: 0, B: 1, A: 1}, // blue
			EdgeColor: render.Color{R: 0, G: 0, B: 0.5, A: 1}, // dark blue edge
			EdgeWidth: 2.0,
			Marker:    markerType,
			Alpha:     1.0,
		}
		ax.Add(scatter)
	}

	// Example 3: Variable sizes, colors, and transparency
	rng := rand.New(rand.NewSource(42)) // for reproducible results
	var variablePoints []geom.Pt
	var variableSizes []float64
	var variableColors []render.Color
	var edgeColors []render.Color

	for i := 0; i < 20; i++ {
		// Random points in a circle
		angle := 2 * math.Pi * float64(i) / 20
		radius := 0.3 + 0.5*rng.Float64()
		x := radius * math.Cos(angle)
		y := radius * math.Sin(angle)

		variablePoints = append(variablePoints, geom.Pt{X: x, Y: y})
		variableSizes = append(variableSizes, 3.0+10.0*rng.Float64())

		// Color based on angle (rainbow effect)
		hue := angle / (2 * math.Pi)
		r, g, b := hsvToRgb(hue, 0.8, 0.9)
		variableColors = append(variableColors, render.Color{R: r, G: g, B: b, A: 1})
		
		// Darker edge colors
		rEdge, gEdge, bEdge := hsvToRgb(hue, 1.0, 0.6)
		edgeColors = append(edgeColors, render.Color{R: rEdge, G: gEdge, B: bEdge, A: 1})
	}

	scatter3 := &core.Scatter2D{
		XY:         variablePoints,
		Sizes:      variableSizes,
		Colors:     variableColors,
		EdgeColors: edgeColors,
		EdgeWidth:  1.0,
		Alpha:      0.7, // Semi-transparent
		Marker:     core.MarkerCircle,
	}
	ax.Add(scatter3)

	// Create a GoBasic renderer with white background
	r := gobasic.New(640, 480, render.Color{R: 1, G: 1, B: 1, A: 1})

	// Save as PNG
	err := core.SavePNG(fig, r, "scatter_basic.png")
	if err != nil {
		fmt.Printf("Error saving PNG: %v\n", err)
		return
	}

	fmt.Println("Successfully created scatter_basic.png with scatter plots!")
	fmt.Println("- Red circles with dark red edges")
	fmt.Println("- Blue shapes showing different marker types with edges")
	fmt.Println("- Colorful semi-transparent circle with variable sizes, colors, and edge colors")
}

// hsvToRgb converts HSV color space to RGB
func hsvToRgb(h, s, v float64) (r, g, b float64) {
	c := v * s
	x := c * (1 - math.Abs(math.Mod(h*6, 2)-1))
	m := v - c

	switch {
	case h < 1.0/6:
		r, g, b = c, x, 0
	case h < 2.0/6:
		r, g, b = x, c, 0
	case h < 3.0/6:
		r, g, b = 0, c, x
	case h < 4.0/6:
		r, g, b = 0, x, c
	case h < 5.0/6:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	return r + m, g + m, b + m
}
