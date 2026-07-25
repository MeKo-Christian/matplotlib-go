package imshow_bilinear

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	imshowInterpFigureWidth  = 256
	imshowInterpFigureHeight = 256
)

func Render() image.Image {
	filter := "bilinear"
	const n = 32
	data := make([][]float64, n)
	for j := range data {
		row := make([]float64, n)
		for i := range row {
			// Diagonal-stripe pattern: enough variation that a smoothing
			// filter (bilinear/bicubic) produces visibly different output
			// than nearest-neighbour.
			if (i+j)%2 == 0 {
				row[i] = 1.0
			} else {
				row[i] = 0.0
			}
		}
		data[j] = row
	}

	fig := core.NewFigure(imshowInterpFigureWidth, imshowInterpFigureHeight)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0, Y: 0},
		Max: geom.Pt{X: 1, Y: 1},
	})

	cmap := "gray"
	vmin, vmax := 0.0, 1.0
	extent := [4]float64{0, float64(n), 0, float64(n)}
	ax.ImShow(data, core.ImShowOptions{
		Colormap:      optional.Of(cmap),
		VMin:          optional.Of(vmin),
		VMax:          optional.Of(vmax),
		Origin:        optional.Of(core.ImageOriginLower),
		Extent:        optional.Of(extent),
		Interpolation: optional.Of(filter),
	})

	r, err := agg.New(imshowInterpFigureWidth, imshowInterpFigureHeight, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.Image()
}
