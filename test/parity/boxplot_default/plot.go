// Package boxplot_default is a parity fixture exercising Matplotlib's default
// boxplot styling: patch_artist=False (unfilled boxes), C1 medians, black
// whiskers/caps, and unfilled-circle fliers. It is the counterpart to
// boxplot_basic, which uses patch_artist=True with explicit per-part styling.
package boxplot_default

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 640
	Height = 360
	DPI    = 100
)

// Plot builds the figure (backend-agnostic) using only default box styling.
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	ax.SetXLim(0, 4)
	ax.SetYLim(0, 10)
	ax.SetTitle("Box Plots (default styling)")
	ax.SetXLabel("Group")
	ax.SetYLabel("Value")

	datasets := [][]float64{
		{0.9, 1.0, 1.1, 1.2, 1.3, 1.45, 1.5, 1.7, 1.8},
		{4.0, 4.2, 4.3, 4.5, 4.8, 5.0, 5.4, 5.8, 9.4},
		{2.0, 2.1, 2.1, 2.2, 2.3, 2.4, 2.4, 2.6, 3.8},
	}
	positions := []float64{1.0, 2.0, 3.0}
	width := 0.55
	manageTicks := false

	ax.BoxPlots(datasets, core.BoxPlotsOptions{
		Positions:   positions,
		Width:       &width,
		ManageTicks: &manageTicks,
	})

	ax.SetAxisBelow(true)
	grid := ax.AddYGrid()
	grid.Color = render.Color{R: 0.8, G: 0.8, B: 0.8, A: 1}
	grid.LineWidth = 0.5
	return fig
}

// Render is the AGG-rendered parity image.
func Render() image.Image {
	fig := Plot()
	r, err := agg.New(Width, Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}
