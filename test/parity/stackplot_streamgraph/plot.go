// Package stackplot_streamgraph is a parity fixture exercising Matplotlib's
// stackplot weighted_wiggle baseline (the "streamgraph" layout). It uses the
// default Axes property cycle for layer colors, mirroring matplotlib's stackplot
// default coloring.
package stackplot_streamgraph

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

// streamData is shared between the Go figure and the matplotlib reference plot.
var streamData = [][]float64{
	{1, 2, 4, 6, 7, 6, 4, 3, 2, 2, 1, 1},
	{0.5, 1, 2, 3, 5, 7, 8, 7, 5, 3, 2, 1},
	{2, 2, 1, 1, 2, 3, 4, 6, 7, 6, 4, 2},
	{3, 2, 2, 1, 1, 1, 2, 3, 4, 5, 6, 5},
}

// Plot builds the figure (backend-agnostic) using the weighted_wiggle baseline.
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	ax.SetXLim(0, 11)
	ax.SetYLim(-13, 11)
	ax.SetTitle("Streamgraph (weighted_wiggle)")

	x := []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}
	ax.StackPlot(x, streamData, core.StackPlotOptions{
		BaselineMode: core.StackBaselineWeightedWiggle,
	})
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
