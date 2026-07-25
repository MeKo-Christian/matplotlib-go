package polar_axes

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 720
	Height = 720
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(720, 720)
	ax := fig.AddPolarAxes(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.10},
		Max: geom.Pt{X: 0.88, Y: 0.88},
	})
	ax.SetTitle("Polar Axes")
	ax.SetXLabel("theta")
	ax.SetYLabel("radius")
	ax.SetYLim(0, 1.1)

	thetaGrid := ax.AddGrid(core.AxisBottom)
	thetaGrid.Color = render.Color{R: 0.8, G: 0.82, B: 0.86, A: 1}
	thetaGrid.LineWidth = 0.9

	radiusGrid := ax.AddGrid(core.AxisLeft)
	radiusGrid.Color = render.Color{R: 0.82, G: 0.84, B: 0.88, A: 0.9}
	radiusGrid.LineWidth = 0.8

	const n = 720
	theta := make([]float64, n)
	radius := make([]float64, n)
	for i := range n {
		theta[i] = 2 * math.Pi * float64(i) / float64(n-1)
		radius[i] = 0.55 + 0.35*math.Cos(5*theta[i])
	}

	lineColor := render.Color{R: 0.16, G: 0.33, B: 0.73, A: 1}
	fillColor := render.Color{R: 0.36, G: 0.56, B: 0.92, A: 0.2}
	lineWidth := 2.2

	_, _ = ax.Plot(theta, radius, core.PlotOptions{
		Color:     &lineColor,
		LineWidth: &lineWidth,
		Label:     "r = 0.55 + 0.35 cos(5theta)",
	})
	ax.FillToBaseline(theta, radius, core.FillOptions{
		Color: &fillColor,
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
