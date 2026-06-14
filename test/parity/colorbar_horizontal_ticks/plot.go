package colorbar_horizontal_ticks

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
)

func Plot() *core.Figure {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.12, Y: 0.16}, Max: geom.Pt{X: 0.90, Y: 0.88}})
	ax.SetTitle("Horizontal Colorbar Ticks")
	ax.SetXLabel("x")
	ax.SetYLabel("y")

	cmap := "viridis"
	vmin, vmax := -1.0, 1.2
	mesh := ax.PColorMesh([][]float64{
		{-1.0, -0.5, 0.0, 0.5},
		{-0.6, -0.1, 0.4, 0.9},
		{-0.2, 0.3, 0.8, 1.2},
	}, core.MeshOptions{
		XEdges:   []float64{0, 1, 2, 3, 4},
		YEdges:   []float64{0, 1, 2, 3},
		Shading:  core.MeshShadingFlat,
		Colormap: &cmap,
		VMin:     &vmin,
		VMax:     &vmax,
	})
	if mesh != nil {
		fig.AddColorbar(ax, mesh, core.ColorbarOptions{
			Location: "bottom",
			Label:    "horizontal",
			Ticks:    []float64{-1, 0, 1},
		})
	}
	ax.SetXLim(0, 4)
	ax.SetYLim(0, 3)
	return fig
}

func Render() image.Image {
	return common.RenderFixtureFigure(Plot(), 640, 360)
}
