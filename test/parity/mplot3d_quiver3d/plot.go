package mplot3d_quiver3d

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/plot3d"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 720
	Height = 560
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(720, 560)
	ax, err := plot3d.AddAxes(fig, geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.16},
		Max: geom.Pt{X: 0.88, Y: 0.88},
	})
	if err != nil {
		panic(err)
	}

	const n = 4
	step := 2.0 / float64(n-1)
	size := n * n * n
	x := make([]float64, 0, size)
	y := make([]float64, 0, size)
	z := make([]float64, 0, size)
	u := make([]float64, 0, size)
	v := make([]float64, 0, size)
	w := make([]float64, 0, size)
	for j := 0; j < n; j++ {
		yv := -1 + float64(j)*step
		for i := 0; i < n; i++ {
			xv := -1 + float64(i)*step
			for k := 0; k < n; k++ {
				zv := -1 + float64(k)*step
				x = append(x, xv)
				y = append(y, yv)
				z = append(z, zv)
				u = append(u, (xv+yv)/5)
				v = append(v, (yv-xv)/5)
				w = append(w, 0)
			}
		}
	}
	ax.Quiver(x, y, z, u, v, w)
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
