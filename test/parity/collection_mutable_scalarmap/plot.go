package collection_mutable_scalarmap

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/optional"
)

func Plot() *core.Figure {
	fig, ax := common.MeshFixtureFigure("Mutable ScalarMap Collection")
	initial := [][]float64{
		{-0.8, -0.4, 0.0, 0.4},
		{-0.6, -0.1, 0.3, 0.8},
		{-0.2, 0.2, 0.6, 1.0},
	}
	cmap := "viridis"
	mesh := ax.PColorMesh(initial, core.MeshOptions{
		XEdges:   []float64{0, 1, 2, 3, 4},
		YEdges:   []float64{0, 1, 2, 3},
		Shading:  core.MeshShadingFlat,
		Colormap: optional.Of(cmap),
	})
	if mesh != nil {
		fig.AddColorbar(ax, mesh, core.ColorbarOptions{Label: "updated"})
		_ = mesh.SetArray([]float64{
			1.00, 0.70, 0.35, 0.05,
			0.80, 0.45, 0.10, -0.20,
			0.55, 0.20, -0.15, -0.50,
		})
		mesh.SetColormap("plasma")
		_ = mesh.SetCLim(-0.5, 1.0)
	}
	ax.SetXLim(0, 4)
	ax.SetYLim(0, 3)
	return fig
}

func Render() image.Image {
	return common.RenderFixtureFigure(Plot(), 640, 360)
}
