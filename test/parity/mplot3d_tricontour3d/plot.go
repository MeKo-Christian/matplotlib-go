package mplot3d_tricontour3d

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

// tricontourMesh mirrors Matplotlib's mplot3d/tricontour3d.py gallery example.
func tricontourMesh() (core.Triangulation, []float64) {
	const (
		nAngles   = 48
		nRadii    = 8
		minRadius = 0.25
	)
	radii := make([]float64, nRadii)
	for i := range radii {
		radii[i] = minRadius + (float64(i)/float64(nRadii-1))*(0.95-minRadius)
	}

	x := make([]float64, nAngles*nRadii)
	y := make([]float64, len(x))
	z := make([]float64, len(x))
	index := 0
	for angleIdx := range nAngles {
		baseAngle := 2 * math.Pi * float64(angleIdx) / float64(nAngles)
		for radiusIdx, radius := range radii {
			angle := baseAngle
			if radiusIdx%2 == 1 {
				angle += math.Pi / nAngles
			}
			x[index] = radius * math.Cos(angle)
			y[index] = radius * math.Sin(angle)
			z[index] = math.Cos(radius) * math.Cos(3*angle)
			index++
		}
	}

	tri, err := core.NewTriangulation(x, y)
	if err != nil {
		panic(err)
	}
	tri.Mask = make([]bool, len(tri.Triangles))
	for i, triangle := range tri.Triangles {
		cx := (tri.X[triangle[0]] + tri.X[triangle[1]] + tri.X[triangle[2]]) / 3
		cy := (tri.Y[triangle[0]] + tri.Y[triangle[1]] + tri.Y[triangle[2]]) / 3
		tri.Mask[i] = math.Hypot(cx, cy) < minRadius
	}
	return tri, z
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

	tri, z := tricontourMesh()
	cmap := "CMRmap"
	ax.TriContour(tri, z, core.PlotOptions{
		Colormap: &cmap,
	})
	ax.SetView(45, -60)
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
