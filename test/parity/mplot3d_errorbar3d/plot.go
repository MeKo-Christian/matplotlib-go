package mplot3d_errorbar3d

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

	x := []float64{-0.8, -0.25, 0.35, 0.9}
	y := []float64{0.1, 0.65, -0.25, 0.35}
	z := []float64{0.2, 0.85, 0.45, 1.1}
	xerr := []float64{0.12, 0.08, 0.16, 0.10}
	yerr := []float64{0.10, 0.14, 0.09, 0.13}
	zerr := []float64{0.18, 0.12, 0.16, 0.10}
	color := render.Color{R: 0.12156862745098039, G: 0.4666666666666667, B: 0.7058823529411765, A: 1}
	ax.ErrorBar3D(x, y, z, xerr, yerr, zerr, plot3d.ErrorBar3DOptions{Color: &color})
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
