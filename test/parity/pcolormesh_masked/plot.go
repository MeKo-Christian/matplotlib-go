package pcolormesh_masked

import (
	"image"

	matcolor "github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
)

func Render() image.Image {
	bad := render.Color{R: 0.62, G: 0.62, B: 0.62, A: 0.78}
	cmapName := "mesh fixture mask"
	matcolor.RegisterColormap(cmapName, matcolor.LookupColormap("viridis").WithBad(bad))

	fig, ax := common.MeshFixtureFigure("masked mesh")
	edge := render.Color{R: 0.98, G: 0.98, B: 0.98, A: 1}
	width := 0.7
	ax.PColorMesh(common.MeshFixtureData(4, 5), core.MeshOptions{
		XEdges:    []float64{0, 1, 2, 3, 4, 5},
		YEdges:    []float64{0, 1, 2, 3, 4},
		Colormap:  &cmapName,
		EdgeColor: &edge,
		EdgeWidth: &width,
		Mask: [][]bool{
			{false, true, false, false, false},
			{false, false, false, true, false},
			{true, false, false, false, false},
			{false, false, true, false, false},
		},
	})
	ax.SetXLim(0, 5)
	ax.SetYLim(0, 4)
	return common.RenderFixtureFigure(fig, 640, 360)
}
