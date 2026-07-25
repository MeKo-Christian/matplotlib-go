package bar_horizontal

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
	ax.SetTitle("Horizontal Bars")
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 6)

	y := []float64{1, 2, 3, 4, 5}
	widths := []float64{3, 7, 2, 8, 5}
	height := 0.6
	color := render.Color{R: 0.8, G: 0.4, B: 0.2, A: 1}
	orientation := core.BarHorizontal
	ax.Bar(y, widths, core.BarOptions{
		Width:       &height,
		Color:       &color,
		Orientation: &orientation,
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
