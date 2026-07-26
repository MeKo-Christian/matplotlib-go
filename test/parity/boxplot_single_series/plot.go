// Package boxplot_single_series is a focused parity fixture for Axes.BoxPlot.
package boxplot_single_series

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

// Plot builds one box plot directly from a single sample series.
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	ax.SetXLim(0, 2)
	ax.SetYLim(0, 10)
	ax.SetTitle("Single-series box plot")
	ax.SetXLabel("Series")
	ax.SetYLabel("Value")

	ax.BoxPlot([]float64{1.0, 1.4, 1.8, 2.2, 2.8, 3.5, 4.2, 4.8, 8.5}, core.BoxPlotOptions{})
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
	return r.Image()
}
