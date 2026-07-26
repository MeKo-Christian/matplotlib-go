package hist2d_weighted_density

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/optional"
)

func Render() image.Image {
	fig, ax := common.MeshFixtureFigure("hist2d weighted density")
	x, y, weights := common.Hist2DWeightedData()
	cmap := "magma"
	result := ax.Hist2D(x, y, core.Hist2DOptions{
		XBinEdges: []float64{-2, -1, 0, 1, 2, 3},
		YBinEdges: []float64{-1.5, -0.5, 0.5, 1.5, 2.5},
		Weights:   weights,
		Norm:      core.HistNormDensity,
		Colormap:  optional.Of(cmap),
	})
	if result != nil {
		fig.AddColorbar(ax, result.Mesh, core.ColorbarOptions{Label: "density"})
	}
	ax.SetXLim(-2, 3)
	ax.SetYLim(-1.5, 2.5)
	return common.RenderFixtureFigure(fig, 640, 360)
}
