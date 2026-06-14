package imshow_transformed

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
)

func Plot() *core.Figure {
	fig := core.NewFigure(420, 420)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.16, Y: 0.14}, Max: geom.Pt{X: 0.9, Y: 0.88}})
	ax.SetTitle("Transformed Imshow")
	ax.SetXLabel("x")
	ax.SetYLabel("y")
	ax.SetXLim(-1, 5)
	ax.SetYLim(-1, 5)
	_ = ax.SetAspect("equal")

	cmap := "magma"
	vmin := 0.0
	vmax := 1.0
	xmin := 0.0
	xmax := 4.0
	ymin := 0.0
	ymax := 4.0
	angle := 28.0
	bilinear := "bilinear"
	ax.Image(common.WaveImageData(6, 6), core.ImageOptions{
		Colormap:      &cmap,
		VMin:          &vmin,
		VMax:          &vmax,
		XMin:          &xmin,
		XMax:          &xmax,
		YMin:          &ymin,
		YMax:          &ymax,
		Origin:        core.ImageOriginLower,
		Angle:         &angle,
		Interpolation: &bilinear,
	})

	return fig
}

func Render() image.Image {
	fig := Plot()
	return common.RenderImageFixture(fig, 420, 420)
}
