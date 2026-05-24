package asinh_norm_image

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
)

func Plot() *core.Figure {
	fig, ax := common.ColorNormFixtureFigure("AsinhNorm Image")
	cmap := "viridis"
	xmin, xmax := 0.0, 7.0
	ymin, ymax := 0.0, 5.0
	img := ax.Image(asinhNormFixtureData(5, 7), core.ImageOptions{
		Colormap: &cmap,
		Norm:     core.AsinhNorm{LinearWidth: 2, VMin: -80, VMax: 120},
		XMin:     &xmin,
		XMax:     &xmax,
		YMin:     &ymin,
		YMax:     &ymax,
		Origin:   core.ImageOriginLower,
	})
	if img != nil {
		fig.AddColorbar(ax, img, core.ColorbarOptions{Label: "asinh value"})
	}
	ax.SetXLim(0, 7)
	ax.SetYLim(0, 5)
	return fig
}

func Render() image.Image {
	return common.RenderFixtureFigure(Plot(), 640, 360)
}

func asinhNormFixtureData(rows, cols int) [][]float64 {
	data := make([][]float64, rows)
	for row := 0; row < rows; row++ {
		data[row] = make([]float64, cols)
		y := float64(row) / math.Max(1, float64(rows-1))
		for col := 0; col < cols; col++ {
			x := float64(col) / math.Max(1, float64(cols-1))
			data[row][col] = 160*(x-0.5) + 30*math.Sin((y-0.5)*math.Pi)
		}
	}
	return data
}
