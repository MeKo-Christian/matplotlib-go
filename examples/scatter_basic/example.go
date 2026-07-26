package scatter_basic

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

const (
	Width  = 640
	Height = 360
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.SetTitle("Basic Scatter")
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 10)

	size := core.ScatterAreaFromRadius(8.0, style.Default.DPI)
	edgeWidth := 0.0
	ax.Scatter(
		[]float64{2, 4, 6, 8, 3, 7},
		[]float64{3, 6, 4, 7, 8, 2},
		core.ScatterOptions{
			Color:     optional.Of(render.Color{R: 0.8, G: 0.2, B: 0.2, A: 1}),
			Size:      optional.Of(size),
			EdgeWidth: optional.Of(edgeWidth),
		},
	)
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
