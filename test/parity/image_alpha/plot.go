package image_alpha

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
)

func Render() image.Image {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.12, Y: 0.16}, Max: geom.Pt{X: 0.92, Y: 0.88}})
	ax.SetTitle("Image Alpha")
	ax.SetXLabel("column")
	ax.SetYLabel("row")
	ax.SetXLim(0, 6)
	ax.SetYLim(0, 6)

	lineColor := render.Color{R: 0.08, G: 0.08, B: 0.10, A: 1}
	lineWidth := 2.0
	_, _ = ax.Plot([]float64{0, 6}, []float64{0, 6}, core.PlotOptions{
		Color:     optional.Of(lineColor),
		LineWidth: optional.Of(lineWidth),
	})
	_, _ = ax.Plot([]float64{0, 6}, []float64{6, 0}, core.PlotOptions{
		Color:     optional.Of(lineColor),
		LineWidth: optional.Of(lineWidth),
	})

	cmap := "plasma"
	alpha := 0.45
	vmin := 0.0
	vmax := 1.0
	xmin := 0.0
	xmax := 6.0
	ymin := 0.0
	ymax := 6.0
	bilinear := "bilinear"
	ax.Image(common.WaveImageData(6, 6), core.ImageOptions{
		Colormap:      optional.Of(cmap),
		VMin:          optional.Of(vmin),
		VMax:          optional.Of(vmax),
		Alpha:         optional.Of(alpha),
		XMin:          optional.Of(xmin),
		XMax:          optional.Of(xmax),
		YMin:          optional.Of(ymin),
		YMax:          optional.Of(ymax),
		Origin:        core.ImageOriginLower,
		Interpolation: optional.Of(bilinear),
	})

	return common.RenderImageFixture(fig, 640, 360)
}
