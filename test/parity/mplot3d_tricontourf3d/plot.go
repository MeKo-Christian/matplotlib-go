package mplot3d_tricontourf3d

import (
	"image"
	"math"

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

// fanMesh builds the same polar fan point cloud used by mplot3d_trisurf3d so the
// auto-Delaunay triangulation matches matplotlib's for an unstructured grid.
func fanMesh() (x, y, z []float64) {
	const (
		nRadii  = 8
		nAngles = 36
	)
	radii := make([]float64, nRadii)
	for i := range radii {
		radii[i] = 0.125 + (float64(i)/float64(nRadii-1))*(1.0-0.125)
	}
	angles := make([]float64, nAngles)
	for i := range angles {
		angles[i] = 2 * math.Pi * float64(i) / float64(nAngles)
	}

	x = make([]float64, 1+nRadii*nAngles)
	y = make([]float64, len(x))
	z = make([]float64, len(x))
	index := 1
	for _, angle := range angles {
		for _, radius := range radii {
			x[index] = radius * math.Cos(angle)
			y[index] = radius * math.Sin(angle)
			index++
		}
	}
	for i := range x {
		z[i] = math.Sin(-(x[i] * y[i]))
	}
	return x, y, z
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

	x, y, z := fanMesh()
	cmap := "CMRmap"
	vmin := -0.6
	vmax := 0.6
	ax.TriContourf(core.Triangulation{X: x, Y: y}, z, core.PlotOptions{
		Colormap: &cmap,
		Levels:   []float64{-0.6, -0.45, -0.3, -0.15, 0, 0.15, 0.3, 0.45, 0.6},
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
