package errorbar_basic

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
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
	ax.SetTitle("Error Bars")
	ax.SetXLim(0, 7)
	ax.SetYLim(0, 6)

	x := []float64{1, 2, 3, 4, 5, 6}
	y := []float64{1.8, 2.5, 2.2, 3.1, 2.8, 3.7}
	xErr := []float64{0.20, 0.25, 0.15, 0.22, 0.30, 0.18}
	yErr := []float64{0.28, 0.20, 0.35, 0.24, 0.30, 0.22}

	lineColor := render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1}
	black := render.Color{R: 0, G: 0, B: 0, A: 1}
	pointColor := render.Color{R: 0.17, G: 0.63, B: 0.17, A: 1}
	lineWidth := 2.0
	errorWidth := 1.2
	pointSize := core.ScatterAreaFromRadius(4.5, style.Default.DPI)
	edgeWidth := 0.0
	errorCap := 6.0

	_, _ = ax.Plot(x, y, core.PlotOptions{
		Color:     &lineColor,
		LineWidth: &lineWidth,
	})
	ax.Scatter(x, y, core.ScatterOptions{
		Color:     &pointColor,
		Size:      &pointSize,
		EdgeWidth: &edgeWidth,
	})
	_, _ = ax.ErrorBar(x, y, xErr, yErr, core.ErrorBarOptions{
		Color:      &black,
		LineWidth:  &errorWidth,
		CapSize:    &errorCap,
		NoDataLine: true,
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
