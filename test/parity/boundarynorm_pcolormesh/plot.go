package boundarynorm_pcolormesh

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/optional"
)

func Render() image.Image {
	fig, ax := common.ColorNormFixtureFigure("BoundaryNorm PColorMesh")
	cmap := "viridis"
	mesh := ax.PColorMesh([][]float64{
		{0.2, 0.8, 1.2, 1.8},
		{2.2, 2.8, 3.2, 3.8},
		{0.5, 1.5, 2.5, 3.5},
	}, core.MeshOptions{
		XEdges:   []float64{0, 1, 2, 3, 4},
		YEdges:   []float64{0, 1, 2, 3},
		Colormap: optional.Of(cmap),
		Norm: core.BoundaryNorm{
			Boundaries: []float64{0, 1, 2, 3, 4},
			NColors:    256,
		},
	})
	if mesh != nil {
		fig.AddColorbar(ax, mesh, core.ColorbarOptions{Label: "band"})
	}
	ax.SetXLim(0, 4)
	ax.SetYLim(0, 3)
	return common.RenderFixtureFigure(fig, 640, 360)
}
