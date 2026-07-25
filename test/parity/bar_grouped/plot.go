package bar_grouped

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
	ax.SetTitle("Grouped Bars")
	ax.SetXLim(0, 7)
	ax.SetYLim(0, 10)

	bar1 := &core.Bar2D{
		X:           []float64{1.2, 2.2, 3.2, 4.2, 5.2},
		Heights:     []float64{3, 7, 2, 8, 5},
		Width:       0.35,
		Color:       render.Color{R: 0.8, G: 0.2, B: 0.2, A: 1},
		EdgeColor:   render.Color{R: 0.5, G: 0, B: 0, A: 1},
		EdgeWidth:   1.0,
		Baseline:    0,
		Orientation: core.BarVertical,
	}
	ax.Add(bar1)

	bar2 := &core.Bar2D{
		X:           []float64{1.8, 2.8, 3.8, 4.8, 5.8},
		Heights:     []float64{5, 4, 6, 3, 7},
		Width:       0.35,
		Color:       render.Color{R: 0.2, G: 0.8, B: 0.2, A: 1},
		EdgeColor:   render.Color{R: 0, G: 0.5, B: 0, A: 1},
		EdgeWidth:   1.0,
		Baseline:    0,
		Orientation: core.BarVertical,
	}
	ax.Add(bar2)
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
