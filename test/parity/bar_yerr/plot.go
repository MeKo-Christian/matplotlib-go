package bar_yerr

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

// Plot builds the figure (backend-agnostic): a categorical bar chart with
// symmetric y error bars, mirroring matplotlib's bar(x, h, yerr=…, capsize=…).
func Plot() *core.Figure {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.SetTitle("Bars with Error Bars")
	ax.SetXLim(0, 6)
	ax.SetYLim(0, 12)

	x := []float64{1, 2, 3, 4, 5}
	heights := []float64{3, 7, 2, 8, 5}
	yErr := []float64{0.8, 1.2, 0.5, 1.5, 0.9}

	width := 0.6
	barColor := render.Color{R: 0.2, G: 0.6, B: 0.8, A: 1}
	black := render.Color{R: 0, G: 0, B: 0, A: 1}
	capSize := 5.0
	errWidth := 1.2

	ax.Bar(x, heights, core.BarOptions{
		Width:   &width,
		Color:   &barColor,
		YErr:    yErr,
		ECol:    &black,
		CapSize: &capSize,
		ErrorKw: &core.ErrorBarOptions{LineWidth: &errWidth},
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
