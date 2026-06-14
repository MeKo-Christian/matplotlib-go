package quad_mesh

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 980
	Height = 620
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(980, 620)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.10, Y: 0.14}, Max: geom.Pt{X: 0.94, Y: 0.88}})
	ax.SetTitle("RendererAgg quad mesh")
	ax.SetXLim(0, 9)
	ax.SetYLim(0, 6)
	ax.XAxis.Locator = core.MultipleLocator{Base: 1}
	ax.YAxis.Locator = core.MultipleLocator{Base: 1}

	data := make([][]float64, 6)
	for y := range data {
		data[y] = make([]float64, 9)
		for x := range data[y] {
			data[y][x] = 0.45 + 0.38*math.Sin(float64(x)*0.7) + 0.22*math.Cos(float64(y)*1.1)
		}
	}
	cmap := "viridis"
	vmin, vmax := -0.15, 1.1
	edgeColor := render.Color{R: 0.96, G: 0.96, B: 0.96, A: 1}
	edgeWidth := 0.65
	ax.PColorMesh(data, core.MeshOptions{
		XEdges:    []float64{0, 1.1, 1.9, 3.0, 3.7, 4.9, 5.8, 6.7, 7.9, 9.0},
		YEdges:    []float64{0, 0.8, 1.7, 2.9, 3.6, 4.8, 6.0},
		Colormap:  &cmap,
		VMin:      &vmin,
		VMax:      &vmax,
		EdgeColor: &edgeColor,
		EdgeWidth: &edgeWidth,
		Label:     "quad mesh",
	})
	return fig
}

// Render is the AGG-rendered showcase image.
func Render() image.Image {
	fig := Plot()
	return common.RenderFixtureFigure(fig, Width, Height)
}
