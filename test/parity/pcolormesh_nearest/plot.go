package pcolormesh_nearest

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/optional"
)

func Render() image.Image {
	fig, ax := common.MeshFixtureFigure("nearest centers")
	cmap := "viridis"
	vmin, vmax := -0.75, 0.95
	ax.PColorMesh(common.MeshFixtureData(4, 5), core.MeshOptions{
		XEdges:   []float64{0.0, 0.8, 2.1, 3.5, 5.0},
		YEdges:   []float64{-1.0, 0.2, 1.4, 2.7},
		Shading:  core.MeshShadingNearest,
		Colormap: optional.Of(cmap),
		VMin:     optional.Of(vmin),
		VMax:     optional.Of(vmax),
	})
	ax.SetXLim(-0.4, 5.7)
	ax.SetYLim(-1.6, 3.35)
	return common.RenderFixtureFigure(fig, 640, 360)
}
