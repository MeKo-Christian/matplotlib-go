package mplot3d_bar2d_zdir

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 720
	Height = 560
	DPI    = 100
)

// Plot mirrors the mplot3d bars3d gallery example: 2D bar graphs projected onto
// the y=3, y=2, y=1, y=0 planes (zdir="y") with 80% opacity. Heights are fixed
// (rather than random) so Go and the matplotlib reference share identical data,
// and each layer uses a single base color since the Go plane-bar API does not
// take a per-bar color array.
func Plot() *core.Figure {
	fig := core.NewFigure(720, 560)
	ax, err := fig.AddAxes3D(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.16},
		Max: geom.Pt{X: 0.88, Y: 0.88},
	})
	if err != nil {
		panic(err)
	}

	xs := []float64{0, 1, 2, 3, 4, 5, 6, 7}
	heights := [][]float64{
		{0.9, 0.5, 0.7, 0.3, 0.8, 0.4, 0.6, 0.5},
		{0.4, 0.8, 0.5, 0.9, 0.3, 0.7, 0.5, 0.6},
		{0.7, 0.3, 0.9, 0.4, 0.6, 0.8, 0.4, 0.7},
		{0.5, 0.7, 0.4, 0.8, 0.5, 0.3, 0.9, 0.6},
	}
	// Matplotlib base colors r, g, b, y zipped with planes y=3, 2, 1, 0.
	colors := []render.Color{
		{R: 1, G: 0, B: 0, A: 1},       // r
		{R: 0, G: 0.5, B: 0, A: 1},     // g
		{R: 0, G: 0, B: 1, A: 1},       // b
		{R: 0.75, G: 0.75, B: 0, A: 1}, // y
	}
	planes := []float64{3, 2, 1, 0}
	alpha := 0.8
	for i, z := range planes {
		zc := z
		c := colors[i]
		ax.Bar(xs, heights[i], core.Bar3DPlaneOptions{
			Color: &c,
			Z:     &zc,
			ZDir:  "y",
			Alpha: &alpha,
		})
	}

	ax.SetXLabel("X")
	ax.SetYLabel("Y")
	ax.SetZLabel("Z")
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
