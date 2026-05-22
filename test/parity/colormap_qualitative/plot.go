package colormap_qualitative

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	common "github.com/cwbudde/matplotlib-go/internal/parityutil"
)

func Render() image.Image {
	fig, ax := common.ColorNormFixtureFigure("Qualitative Colormap")
	cmap := "tab20"
	nearest := "nearest"
	xmin, xmax := 0.0, 10.0
	ymin, ymax := 0.0, 2.0
	vmin, vmax := 0.0, 19.0
	ax.Image(qualitativeData(), core.ImageOptions{
		Colormap:      &cmap,
		VMin:          &vmin,
		VMax:          &vmax,
		XMin:          &xmin,
		XMax:          &xmax,
		YMin:          &ymin,
		YMax:          &ymax,
		Origin:        core.ImageOriginLower,
		Interpolation: &nearest,
	})
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 2)
	return common.RenderFixtureFigure(fig, 640, 360)
}

func qualitativeData() [][]float64 {
	return [][]float64{
		{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		{10, 11, 12, 13, 14, 15, 16, 17, 18, 19},
	}
}
