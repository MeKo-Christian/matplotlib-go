package colorbar_extendfrac

import (
	"image"

	matcolor "github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
)

func Render() image.Image {
	under := render.Color{R: 0.08, G: 0.16, B: 0.72, A: 1}
	over := render.Color{R: 0.78, G: 0.12, B: 0.08, A: 1}
	cmapName := "extendfrac fixture"
	matcolor.RegisterColormap(cmapName, matcolor.LookupColormap("viridis").Copy(cmapName).WithUnder(under).WithOver(over))

	fig, ax := common.ColorNormFixtureFigure("Colorbar ExtendFrac")
	vmin, vmax := 0.0, 1.0
	mesh := ax.PColorMesh([][]float64{
		{-0.35, 0.15, 0.35},
		{0.55, 0.85, 1.35},
	}, core.MeshOptions{
		XEdges:   []float64{0, 1, 2, 3},
		YEdges:   []float64{0, 1, 2},
		Colormap: optional.Of(cmapName),
		VMin:     optional.Of(vmin),
		VMax:     optional.Of(vmax),
	})
	if mesh != nil {
		fig.AddColorbar(ax, mesh, core.ColorbarOptions{
			Label:      "extendfrac",
			Extend:     "both",
			ExtendFrac: []float64{0.1, 0.3},
		})
	}
	ax.SetXLim(0, 3)
	ax.SetYLim(0, 2)
	return common.RenderFixtureFigure(fig, 640, 360)
}
