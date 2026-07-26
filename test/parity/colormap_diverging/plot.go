package colormap_diverging

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/optional"
)

func Plot() *core.Figure {
	fig, ax := common.ColorNormFixtureFigure("Diverging Colormap")
	cmap := "RdBu"
	nearest := "nearest"
	xmin, xmax := 0.0, 9.0
	ymin, ymax := 0.0, 5.0
	vmin, vmax := -1.0, 1.0
	ax.Image(divergingData(5, 9), core.ImageOptions{
		Colormap:      optional.Of(cmap),
		VMin:          optional.Of(vmin),
		VMax:          optional.Of(vmax),
		XMin:          optional.Of(xmin),
		XMax:          optional.Of(xmax),
		YMin:          optional.Of(ymin),
		YMax:          optional.Of(ymax),
		Origin:        core.ImageOriginLower,
		Interpolation: optional.Of(nearest),
	})
	ax.SetXLim(0, 9)
	ax.SetYLim(0, 5)
	return fig
}

func Render() image.Image {
	fig := Plot()
	return common.RenderFixtureFigure(fig, 640, 360)
}

func divergingData(rows, cols int) [][]float64 {
	data := make([][]float64, rows)
	for row := range data {
		data[row] = make([]float64, cols)
		y := float64(row) / float64(max(1, rows-1))
		for col := range data[row] {
			x := float64(col) / float64(max(1, cols-1))
			data[row][col] = 2*x - 1 + 0.22*(y-0.5)
		}
	}
	return data
}
