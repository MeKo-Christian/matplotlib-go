// Package contour_styles renders a showcase of contour configuration options:
// monochrome contour lines with default negative_linestyles dashing and clabel
// format strings, plus a filled contour with extend="both" and cycled hatches.
package contour_styles

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 640
	Height = 360
	DPI    = 100

	gridN = 41
)

// dipoleField evaluates z = x * exp(-x^2 - y^2) on a regular grid over
// [-3, 3]^2, producing a symmetric field with negative and positive lobes.
func dipoleField() (xs, ys []float64, z [][]float64) {
	xs = make([]float64, gridN)
	ys = make([]float64, gridN)
	for i := range gridN {
		t := -3.0 + 6.0*float64(i)/float64(gridN-1)
		xs[i] = t
		ys[i] = t
	}
	z = make([][]float64, gridN)
	for j := range gridN {
		row := make([]float64, gridN)
		for i := range gridN {
			x, y := xs[i], ys[j]
			row[i] = x * math.Exp(-x*x-y*y)
		}
		z[j] = row
	}
	return xs, ys, z
}

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	xs, ys, z := dipoleField()

	black := render.Color{R: 0, G: 0, B: 0, A: 1}
	lineWidth := 1.5

	// Left: monochrome contour lines. Negative levels dash by default
	// (negative_linestyles), and clabel applies a "%.2f" format string with
	// inline labels.
	lineAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.08, Y: 0.12}, Max: geom.Pt{X: 0.46, Y: 0.90}})
	lineAx.SetTitle("Contour: negative dashing")
	lineAx.SetXLim(-3, 3)
	lineAx.SetYLim(-3, 3)
	cs := lineAx.Contour(z, core.ContourOptions{
		X:         xs,
		Y:         ys,
		Levels:    []float64{-0.3, -0.2, -0.1, 0.1, 0.2, 0.3},
		Color:     &black,
		LineWidth: &lineWidth,
	})
	noInline := false
	lineAx.Clabel(cs, core.ClabelOptions{
		FormatString: "%.2f",
		Inline:       &noInline,
		ManualPositions: []geom.Pt{
			{X: -0.7, Y: 0.6},
			{X: -0.7, Y: 1.21},
			{X: 0.7, Y: 0.6},
			{X: 0.7, Y: 1.21},
		},
	})

	// Right: filled contour with extend="both" (under/over bands) and hatches
	// cycled across the bands.
	fillAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.56, Y: 0.12}, Max: geom.Pt{X: 0.94, Y: 0.90}})
	fillAx.SetTitle("Contourf: extend + hatches")
	fillAx.SetXLim(-3, 3)
	fillAx.SetYLim(-3, 3)
	fillAx.Contourf(z, core.ContourOptions{
		X:       xs,
		Y:       ys,
		Levels:  []float64{-0.3, -0.2, -0.1, 0, 0.1, 0.2, 0.3},
		Extend:  "both",
		Hatches: []string{"//", "\\\\", "xx"},
	})

	return fig
}

// Render is the AGG-rendered showcase image.
func Render() image.Image {
	fig := Plot()
	r, err := agg.New(Width, Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.Image()
}
