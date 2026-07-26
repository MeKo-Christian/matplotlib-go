package lognorm_imshow

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/optional"
)

func Render() image.Image {
	fig, ax := common.ColorNormFixtureFigure("LogNorm Imshow")
	cmap := "magma"
	xmin, xmax := 0.0, 6.0
	ymin, ymax := 0.0, 5.0
	img := ax.Image(common.LogNormFixtureData(5, 6), core.ImageOptions{
		Colormap: optional.Of(cmap),
		Norm:     core.LogNorm{VMin: 1, VMax: 1000},
		XMin:     optional.Of(xmin),
		XMax:     optional.Of(xmax),
		YMin:     optional.Of(ymin),
		YMax:     optional.Of(ymax),
		Origin:   core.ImageOriginLower,
	})
	if img != nil {
		fig.AddColorbar(ax, img, core.ColorbarOptions{Label: "log value"})
	}
	ax.SetXLim(0, 6)
	ax.SetYLim(0, 5)
	return common.RenderFixtureFigure(fig, 640, 360)
}
