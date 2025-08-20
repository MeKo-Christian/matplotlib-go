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
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	// Set up coordinate scales
	ax.XScale = transform.NewLinear(0, 4*math.Pi)
	ax.YScale = transform.NewLinear(-2, 3)

	// Generate data for sine and cosine curves
	n := 100
	x := make([]float64, n)
	y1 := make([]float64, n) // sine curve
	y2 := make([]float64, n) // cosine curve

	for i := 0; i < n; i++ {
		t := 4 * math.Pi * float64(i) / float64(n-1)
		x[i] = t
		y1[i] = math.Sin(t)
		y2[i] = math.Cos(t)
	}

	// Example 1: Fill between sine and cosine
	fillBetween := core.FillBetween(x, y1, y2,
		render.Color{R: 0.3, G: 0.7, B: 0.9, A: 0.6}) // semi-transparent blue
	ax.Add(fillBetween)

	// Example 2: Fill sine curve to baseline
	fillSine := core.FillToBaseline(x, y1, 0,
		render.Color{R: 1, G: 0.3, B: 0.3, A: 0.4}) // semi-transparent red
	ax.Add(fillSine)

	// Add the actual curves on top for reference
	sineLine := &core.Line2D{
		XY:  make([]geom.Pt, n),
		W:   2.0,
		Col: render.Color{R: 1, G: 0, B: 0, A: 1}, // red line
	}

	cosLine := &core.Line2D{
		XY:  make([]geom.Pt, n),
		W:   2.0,
		Col: render.Color{R: 0, G: 0, B: 1, A: 1}, // blue line
	}

	for i := 0; i < n; i++ {
		sineLine.XY[i] = geom.Pt{X: x[i], Y: y1[i]}
		cosLine.XY[i] = geom.Pt{X: x[i], Y: y2[i]}
	}

	ax.Add(sineLine)
	ax.Add(cosLine)

	// Create a GoBasic renderer with white background
	r := gobasic.New(800, 600, render.Color{R: 1, G: 1, B: 1, A: 1})

	// Save as PNG
	err := core.SavePNG(fig, r, "fill_basic.png")
	if err != nil {
		fmt.Printf("Error saving PNG: %v\n", err)
		return
	}

	// Create a second figure for stacked areas
	fig2 := core.NewFigure(800, 600)
	ax2 := fig2.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	ax2.XScale = transform.NewLinear(0, 10)
	ax2.YScale = transform.NewLinear(0, 10)

	// Create stacked area data
	xStack := []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	area1 := []float64{1, 1.2, 1.8, 2.1, 2.5, 3.0, 2.8, 2.3, 1.9, 1.5, 1.0}
	area2 := make([]float64, len(area1))
	area3 := make([]float64, len(area1))

	// Stack the areas
	for i := range area1 {
		area2[i] = area1[i] + 1.5 + 0.5*math.Sin(float64(i)*0.5)
		area3[i] = area2[i] + 2.0 + 0.8*math.Cos(float64(i)*0.3)
	}

	// Create stacked fills
	stack1 := core.FillToBaseline(xStack, area1, 0,
		render.Color{R: 0.8, G: 0.3, B: 0.3, A: 0.8}) // red
	stack2 := core.FillBetween(xStack, area1, area2,
		render.Color{R: 0.3, G: 0.8, B: 0.3, A: 0.8}) // green
	stack3 := core.FillBetween(xStack, area2, area3,
		render.Color{R: 0.3, G: 0.3, B: 0.8, A: 0.8}) // blue

	ax2.Add(stack1)
	ax2.Add(stack2)
	ax2.Add(stack3)

	// Save stacked areas
	r2 := gobasic.New(800, 600, render.Color{R: 1, G: 1, B: 1, A: 1})
	err = core.SavePNG(fig2, r2, "fill_stacked.png")
	if err != nil {
		fmt.Printf("Error saving stacked PNG: %v\n", err)
		return
	}

	fmt.Println("Successfully created fill examples!")
	fmt.Println("- fill_basic.png: Fill between sine/cosine and sine to baseline")
	fmt.Println("- fill_stacked.png: Stacked area chart")
}
