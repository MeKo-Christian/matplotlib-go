package mplot3d_fill_between3d

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	mplcolor "github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
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
	ax, err := fig.AddAxes3D(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.16},
		Max: geom.Pt{X: 0.88, Y: 0.88},
	})
	if err != nil {
		panic(err)
	}

	const n = 50
	theta := make([]float64, n)
	x1 := make([]float64, n)
	y1 := make([]float64, n)
	z1 := make([]float64, n)
	x2 := make([]float64, n)
	y2 := make([]float64, n)
	z2 := make([]float64, n)
	for i := range n {
		t := 2 * math.Pi * float64(i) / float64(n-1)
		theta[i] = t
		x1[i] = math.Cos(t)
		y1[i] = math.Sin(t)
		z1[i] = float64(i) / float64(n-1)
		x2[i] = math.Cos(t + math.Pi)
		y2[i] = math.Sin(t + math.Pi)
		z2[i] = z1[i]
	}

	width := 2.0
	blue := mplcolor.Tab10[0]
	ax.Plot3D(x1, y1, z1, core.PlotOptions{LineWidth: &width, Color: &blue})
	ax.Plot3D(x2, y2, z2, core.PlotOptions{LineWidth: &width, Color: &blue})

	alpha := 0.5
	ax.FillBetween(x1, y1, z1, x2, y2, z2, core.FillBetween3DOptions{
		Alpha: &alpha,
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
