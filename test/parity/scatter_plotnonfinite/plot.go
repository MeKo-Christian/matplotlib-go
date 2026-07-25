// Package scatter_plotnonfinite is a parity fixture exercising the scatter
// plotnonfinite kwarg: points with non-finite scalar (color) values are kept
// and ride the colormap's "bad" color (transparent by default), while finite
// points map through viridis.
package scatter_plotnonfinite

import (
	"image"
	"math"

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

// scatterData builds the shared point/scalar arrays. Indices 3 and 8 carry a
// non-finite scalar value; index 6 carries a non-finite position.
func scatterData() (x, y, c []float64) {
	n := 12
	x = make([]float64, n)
	y = make([]float64, n)
	c = make([]float64, n)
	for i := 0; i < n; i++ {
		x[i] = 0.5 + 0.5*float64(i)
		y[i] = 3 + 2*math.Sin(float64(i)*0.6)
		c[i] = float64(i) / float64(n-1)
	}
	c[3] = math.NaN()
	c[8] = math.NaN()
	x[6] = math.NaN()
	return x, y, c
}

// Plot builds the figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.12, Y: 0.12}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	ax.SetTitle("Scatter (plotnonfinite)")
	ax.SetXLim(0, 7)
	ax.SetYLim(0, 6)

	x, y, c := scatterData()
	vmin, vmax := 0.0, 1.0
	size := core.ScatterAreaFromRadius(7, style.Default.DPI)
	edgeWidth := 0.0
	ax.Scatter(x, y, core.ScatterOptions{
		ScalarValues:  c,
		Colormap:      "viridis",
		VMin:          &vmin,
		VMax:          &vmax,
		Size:          &size,
		EdgeWidth:     &edgeWidth,
		PlotNonfinite: true,
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
