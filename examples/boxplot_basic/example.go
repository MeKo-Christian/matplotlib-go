// Package boxplot_basic is a showcase example: three vertical box plots with
// custom colors, whiskers, medians, caps, and outlier markers.
package boxplot_basic

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 640
	Height = 360
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.SetXLim(0, 4)
	ax.SetYLim(0, 10)
	ax.SetTitle("Box Plots")
	ax.SetXLabel("Group")
	ax.SetYLabel("Value")

	datasets := [][]float64{
		{0.9, 1.0, 1.1, 1.2, 1.3, 1.45, 1.5, 1.7, 1.8},
		{4.0, 4.2, 4.3, 4.5, 4.8, 5.0, 5.4, 5.8, 9.4},
		{2.0, 2.1, 2.1, 2.2, 2.3, 2.4, 2.4, 2.6, 3.8},
	}
	positions := []float64{1.0, 2.0, 3.0}
	colors := []render.Color{
		{R: 0.25, G: 0.55, B: 0.82, A: 1},
		{R: 0.80, G: 0.45, B: 0.20, A: 1},
		{R: 0.35, G: 0.70, B: 0.35, A: 1},
	}

	black := render.Color{R: 0, G: 0, B: 0, A: 1}
	alpha := 0.75
	boxWidth := 0.55
	edgeWidth := 1.2
	whiskerWidth := 1.2
	medianWidth := 1.8
	flierSize := 4.0
	showFliers := true
	manageTicks := false
	patchArtist := true
	ax.BoxPlots(datasets, core.BoxPlotsOptions{
		Positions:    positions,
		Width:        optional.Of(boxWidth),
		Colors:       colors,
		PatchArtist:  optional.Of(patchArtist),
		EdgeColor:    optional.Of(black),
		MedianColor:  optional.Of(black),
		WhiskerColor: optional.Of(black),
		CapColor:     optional.Of(black),
		FlierColor:   optional.Of(black),
		EdgeWidth:    optional.Of(edgeWidth),
		WhiskerWidth: optional.Of(whiskerWidth),
		MedianWidth:  optional.Of(medianWidth),
		FlierSize:    optional.Of(flierSize),
		Alpha:        optional.Of(alpha),
		ShowFliers:   optional.Of(showFliers),
		ManageTicks:  optional.Of(manageTicks),
	})

	grid := ax.AddYGrid()
	grid.Color = render.Color{R: 0.8, G: 0.8, B: 0.8, A: 1}
	grid.LineWidth = 0.5
	return fig
}

// Render is the AGG-rendered showcase image.
func Render() image.Image {
	fig := Plot()
	r, err := agg.New(Width, Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	r.SetResolution(DPI)
	core.DrawFigure(fig, r)
	return r.Image()
}
