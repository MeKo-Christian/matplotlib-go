package colormap_cyclic

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/optional"
)

func Render() image.Image {
	fig, ax := common.ColorNormFixtureFigure("Cyclic Colormap")
	cmap := "twilight"
	nearest := "nearest"
	xmin, xmax := 0.0, 12.0
	ymin, ymax := 0.0, 6.0
	vmin, vmax := 0.0, 1.0
	ax.Image(cyclicData(6, 12), core.ImageOptions{
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
	ax.SetXLim(0, 12)
	ax.SetYLim(0, 6)
	return common.RenderFixtureFigure(fig, 640, 360)
}

func cyclicData(rows, cols int) [][]float64 {
	data := make([][]float64, rows)
	cx := float64(cols-1) / 2
	cy := float64(rows-1) / 2
	for row := range data {
		data[row] = make([]float64, cols)
		for col := range data[row] {
			angle := math.Atan2(float64(row)-cy, float64(col)-cx)
			data[row][col] = (angle + math.Pi) / (2 * math.Pi)
		}
	}
	return data
}
