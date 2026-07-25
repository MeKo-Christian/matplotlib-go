package fill_basic

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

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.SetTitle("Fill to Baseline")
	ax.SetXLim(0, 10)
	ax.SetYLim(-1, 3)

	x := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9}
	zeros := make([]float64, len(x))
	y := []float64{0.5, 1.8, 2.3, 1.2, 2.8, 1.9, 2.1, 1.5, 0.8}
	faceColor := render.Color{R: 0.3, G: 0.7, B: 0.9, A: 0.7}
	edgeColor := render.Color{R: 0.1, G: 0.3, B: 0.5, A: 1.0}
	edgeWidth := 2.0
	_, _ = ax.FillBetween(x, zeros, y, core.FillOptions{
		Color:     &faceColor,
		EdgeColor: &edgeColor,
		EdgeWidth: &edgeWidth,
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
