package pcolormesh_gouraud

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/optional"
)

func Render() image.Image {
	fig, ax := common.MeshFixtureFigure("Gouraud mesh")
	cmap := "viridis"
	vmin, vmax := -0.85, 0.85
	mesh := ax.PColorMesh(common.MeshFixtureData(5, 6), core.MeshOptions{
		XEdges:   []float64{-2.5, -1.5, -0.4, 0.8, 1.7, 2.8},
		YEdges:   []float64{-1.8, -0.8, 0.1, 1.1, 2.0},
		Shading:  core.MeshShadingGouraud,
		Colormap: optional.Of(cmap),
		VMin:     optional.Of(vmin),
		VMax:     optional.Of(vmax),
	})
	if mesh != nil {
		fig.AddColorbar(ax, mesh, core.ColorbarOptions{Label: "value"})
	}
	ax.SetXLim(-2.5, 2.8)
	ax.SetYLim(-1.8, 2.0)
	return common.RenderFixtureFigure(fig, 640, 360)
}
