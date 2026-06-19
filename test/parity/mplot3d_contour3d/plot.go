package mplot3d_contour3d

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
	Height = 560
	DPI    = 100
)

// testData mirrors mpl_toolkits.mplot3d.axes3d.get_test_data(delta): the axis
// values come from numpy.arange(-3, 3, delta), Z is a difference of two
// Gaussians evaluated on the unscaled grid, then the axes are scaled by 10 and
// Z by 500.
func testData(delta float64) (xs, ys []float64, z [][]float64) {
	n := int(math.Round(6.0 / delta))
	axis := make([]float64, n)
	for i := range axis {
		axis[i] = -3.0 + delta*float64(i)
	}
	xs = make([]float64, n)
	ys = make([]float64, n)
	z = make([][]float64, n)
	for yi := range n {
		z[yi] = make([]float64, n)
		for xi := range n {
			x := axis[xi]
			y := axis[yi]
			z1 := math.Exp(-(x*x+y*y)/2) / (2 * math.Pi)
			z2 := math.Exp(-((x-1)/1.5*((x-1)/1.5)+(y-1)/0.5*((y-1)/0.5))/2) /
				(2 * math.Pi * 0.5 * 1.5)
			z[yi][xi] = (z2 - z1) * 500
		}
	}
	for i := range n {
		xs[i] = axis[i] * 10
		ys[i] = axis[i] * 10
	}
	return xs, ys, z
}

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(720, 560)
	ax, err := fig.AddAxes3D(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.16},
		Max: geom.Pt{X: 0.88, Y: 0.88},
	})
	if err != nil {
		panic(err)
	}

	x, y, z := testData(0.25)
	cmap := "coolwarm"
	vmin := -50.0
	vmax := 75.0
	ax.Contour(x, y, z, core.PlotOptions{
		Colormap: &cmap,
		Levels:   []float64{-50, -25, 0, 25, 50, 75},
		VMin:     &vmin,
		VMax:     &vmax,
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
	return r.GetImage()
}
