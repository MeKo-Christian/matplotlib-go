package image_heatmap

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 640
	Height = 360
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.15},
		Max: geom.Pt{X: 0.95, Y: 0.9},
	})
	ax.SetTitle("Image Heatmap")
	ax.SetXLim(0, 3)
	ax.SetYLim(0, 3)

	data := [][]float64{
		{0, 1, 2},
		{3, 4, 5},
		{6, 7, 8},
	}

	cmap := "viridis"
	nearest := "nearest"
	vmin, vmax := 0.0, 8.0
	extent := [4]float64{0, 3, 0, 3}
	ax.ImShow(data, core.ImShowOptions{
		Colormap:      optional.Of(cmap),
		VMin:          optional.Of(vmin),
		VMax:          optional.Of(vmax),
		Origin:        optional.Of(core.ImageOriginLower),
		Extent:        optional.Of(extent),
		Aspect:        optional.Of(core.AspectAuto),
		Interpolation: optional.Of(nearest),
	})
	return fig
}

// Render is the AGG-rendered showcase image.
func Render() image.Image {
	fig := Plot()
	r, err := agg.New(Width, Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.Image()
}
