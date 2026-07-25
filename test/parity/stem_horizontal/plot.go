// Package stem_horizontal is a parity fixture exercising Matplotlib's stem
// orientation="horizontal" layout: locs run along y and heads extend along x
// from a vertical baseline.
package stem_horizontal

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 720
	Height = 420
	DPI    = 100
)

// Shared with test/parity/stem_horizontal/plot.py.
var (
	locs  = []float64{1, 2, 3, 4, 5, 6, 7}
	heads = []float64{0.9, 2.2, 1.6, 3.3, 2.4, 3.7, 2.1}
)

// Plot builds the figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.10, Y: 0.16}, Max: geom.Pt{X: 0.94, Y: 0.86}})
	ax.SetTitle("Horizontal Stem")
	ax.SetXLabel("Amplitude")
	ax.SetYLabel("Sample")
	ax.SetXLim(-0.2, 4.2)
	ax.SetYLim(0.5, 7.5)
	grid := ax.AddXGrid()
	grid.Color = render.Color{R: 0.8, G: 0.8, B: 0.8, A: 1}
	grid.LineWidth = 0.5

	stemColor := render.Color{R: 0.15, G: 0.42, B: 0.73, A: 1}
	baseline := 0.3
	markerSize := 7.0
	ax.Stem(locs, heads, core.StemOptions{
		Orientation:   "horizontal",
		Color:         optional.Of(stemColor),
		Baseline:      baseline,
		MarkerSize:    optional.Of(markerSize),
		BaselineColor: optional.Of(render.Color{R: 0.32, G: 0.32, B: 0.32, A: 1}),
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
