package twoslope_norm_image

import (
	"image"

	matcolor "github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
)

func Render() image.Image {
	fig, ax := common.ColorNormFixtureFigure("TwoSlopeNorm Diverging")
	cmap := "diverging fixture"
	matcolor.RegisterColormap(cmap, matcolor.NewColormap(cmap, []matcolor.ColorStop{
		{Pos: 0.0, Color: render.Color{R: 0.23, G: 0.30, B: 0.75, A: 1}},
		{Pos: 0.5, Color: render.Color{R: 0.86, G: 0.86, B: 0.86, A: 1}},
		{Pos: 1.0, Color: render.Color{R: 0.71, G: 0.02, B: 0.15, A: 1}},
	}))
	xmin, xmax := 0.0, 7.0
	ymin, ymax := 0.0, 5.0
	img := ax.Image(common.TwoSlopeFixtureData(5, 7), core.ImageOptions{
		Colormap: optional.Of(cmap),
		Norm:     core.TwoSlopeNorm{VMin: -3, VCenter: 0, VMax: 6},
		XMin:     optional.Of(xmin),
		XMax:     optional.Of(xmax),
		YMin:     optional.Of(ymin),
		YMax:     optional.Of(ymax),
		Origin:   core.ImageOriginLower,
	})
	if img != nil {
		fig.AddColorbar(ax, img, core.ColorbarOptions{Label: "anomaly"})
	}
	ax.SetXLim(0, 7)
	ax.SetYLim(0, 5)
	return common.RenderFixtureFigure(fig, 640, 360)
}
