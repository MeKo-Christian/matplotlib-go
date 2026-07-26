package matshow_basic

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/optional"
)

func Render() image.Image {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.22, Y: 0.12}, Max: geom.Pt{X: 0.78, Y: 0.9}})
	ax.SetTitle("Matshow")
	cmap := "cividis"
	ax.MatShow([][]float64{
		{0.10, 0.20, 0.35, 0.55},
		{0.18, 0.32, 0.48, 0.70},
		{0.28, 0.46, 0.66, 0.86},
		{0.40, 0.58, 0.78, 0.96},
	}, core.MatShowOptions{Colormap: optional.Of(cmap)})

	return common.RenderImageFixture(fig, 640, 360)
}
