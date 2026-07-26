package mplot3d_scatter3d

import (
	"image"
	"math/rand/v2"

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

	rng := rand.New(rand.NewPCG(19680801, 0))
	uniform := func(low, high float64, n int) []float64 {
		values := make([]float64, n)
		for i := range values {
			values[i] = low + rng.Float64()*(high-low)
		}
		return values
	}
	const n = 100
	x := uniform(23, 32, n)
	y := uniform(0, 100, n)
	z := uniform(-50, -25, n)
	ax.Scatter3D(x, y, z, core.ScatterOptions{})
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
