// Package errorbar_capthick is a parity fixture exercising the errorbar
// capthick kwarg: the error-bar caps are drawn with a thicker line than the
// bars themselves.
package errorbar_capthick

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

// Shared with test/parity/errorbar_capthick/plot.py.
var (
	xs   = []float64{1, 2, 3, 4, 5, 6}
	ys   = []float64{1.8, 2.5, 2.2, 3.1, 2.8, 3.7}
	xErr = []float64{0.20, 0.25, 0.15, 0.22, 0.30, 0.18}
	yErr = []float64{0.28, 0.20, 0.35, 0.24, 0.30, 0.22}
)

// Plot builds the figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	ax.SetTitle("Error Bars (capthick)")
	ax.SetXLim(0, 7)
	ax.SetYLim(0, 6)

	black := render.Color{R: 0, G: 0, B: 0, A: 1}
	pointColor := render.Color{R: 0.17, G: 0.63, B: 0.17, A: 1}
	errorWidth := 1.2
	capSize := 6.0
	capThick := 3.0
	edgeWidth := 0.0
	pointSize := core.ScatterAreaFromRadius(4.5, style.Default.DPI)

	ax.Scatter(xs, ys, core.ScatterOptions{
		Color:     &pointColor,
		Size:      &pointSize,
		EdgeWidth: &edgeWidth,
	})
	ax.ErrorBar(xs, ys, xErr, yErr, core.ErrorBarOptions{
		Color:      &black,
		LineWidth:  &errorWidth,
		CapSize:    &capSize,
		CapThick:   &capThick,
		NoDataLine: true,
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
	return r.GetImage()
}
