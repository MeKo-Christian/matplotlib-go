package pcolor_flat

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
)

func Render() image.Image {
	fig, ax := common.MeshFixtureFigure("pcolor flat")
	cmap := "plasma"
	edge := render.Color{R: 0.96, G: 0.96, B: 0.96, A: 1}
	width := 0.75
	ax.PColor(common.MeshFixtureData(4, 5), core.MeshOptions{
		XEdges:    []float64{0.0, 0.9, 2.0, 3.4, 4.1, 5.2},
		YEdges:    []float64{-0.2, 0.8, 1.6, 2.9, 4.0},
		Shading:   core.MeshShadingFlat,
		Colormap:  optional.Of(cmap),
		EdgeColor: optional.Of(edge),
		EdgeWidth: optional.Of(width),
	})
	ax.SetXLim(0, 5.2)
	ax.SetYLim(-0.2, 4.0)
	return common.RenderFixtureFigure(fig, 640, 360)
}
