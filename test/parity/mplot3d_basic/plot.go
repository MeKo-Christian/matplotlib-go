package mplot3d_basic

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	common "github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/plot3d"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 760
	Height = 560
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(760, 560)
	ax, err := plot3d.AddAxes(fig, geom.Rect{Min: geom.Pt{X: 0.12, Y: 0.14}, Max: geom.Pt{X: 0.88, Y: 0.88}})
	if err != nil {
		panic(err)
	}
	ax.SetTitle("3D Toolkit Scaffold")
	ax.SetXLabel("x")
	ax.SetYLabel("y")
	ax.SetView(30, -60)
	x := []float64{0, 1}
	y := []float64{0, 1}
	zGrid := [][]float64{{0, 1}, {1, 2}}
	gray := render.Color{R: 0.35, G: 0.35, B: 0.35, A: 1}
	cmap := "viridis"
	surfaceAlpha := 0.35
	barColor := render.Color{R: 1.0, G: 0.4980392156862745, B: 0.054901960784313725, A: 1}
	barAlpha := 0.7
	ax.Plot3D([]float64{0, 1}, []float64{0, 1}, []float64{0, 1})
	ax.Scatter3D([]float64{0.5, 0.7}, []float64{0.2, 0.9}, []float64{0.1, 0.3})
	ax.Wireframe(x, y, zGrid, core.PlotOptions{Color: &gray})
	ax.Surface(x, y, zGrid, core.PlotOptions{Alpha: &surfaceAlpha, Colormap: &cmap})
	ax.Contour(x, y, zGrid)
	ax.Bar3D([]float64{0.2}, []float64{0.3}, []float64{0.4}, []float64{0.2}, []float64{0.2}, []float64{0.3}, plot3d.Bar3DOptions{Color: &barColor, Alpha: &barAlpha})
	ax.Text3D(0.2, 0.8, 0.6, "demo point")
	return fig
}

// Render is the AGG-rendered showcase image.
func Render() image.Image {
	fig := Plot()
	return common.RenderFixtureFigure(fig, Width, Height)
}
