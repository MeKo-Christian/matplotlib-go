package imshow_clipped

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/optional"
)

func Plot() *core.Figure {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.12, Y: 0.16}, Max: geom.Pt{X: 0.92, Y: 0.88}})
	ax.SetTitle("Clipped Imshow")
	ax.SetXLabel("x")
	ax.SetYLabel("y")

	cmap := "viridis"
	nearest := "nearest"
	extent := [4]float64{0, 8, 0, 8}
	ax.ImShow(common.WaveImageData(8, 8), core.ImShowOptions{
		Colormap:      optional.Of(cmap),
		Extent:        optional.Of(extent),
		Origin:        optional.Of(core.ImageOriginLower),
		Aspect:        optional.Of("auto"),
		Interpolation: optional.Of(nearest),
	})
	ax.SetXLim(2, 6)
	ax.SetYLim(1, 7)

	return fig
}

func Render() image.Image {
	fig := Plot()
	return common.RenderImageFixture(fig, 640, 360)
}
