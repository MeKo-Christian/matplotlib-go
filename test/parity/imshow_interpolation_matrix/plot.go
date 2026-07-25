package imshow_interpolation_matrix

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	figureWidth  = 800
	figureHeight = 480
)

var interpolationModes = []string{
	"nearest", "none", "bilinear", "bicubic", "hanning",
	"hamming", "lanczos", "spline16", "spline36", "kaiser",
	"quadric", "catrom", "gaussian", "bessel", "mitchell",
	"sinc", "blackman", "hermite", "antialiased", "auto",
}

func Plot() *core.Figure {
	fig := core.NewFigure(figureWidth, figureHeight)
	data := interpolationMatrixData()
	cmap := "viridis"
	vmin, vmax := 0.0, 1.0
	extent := [4]float64{0, float64(len(data[0])), 0, float64(len(data))}

	const (
		cols   = 5
		left   = 0.04
		right  = 0.98
		bottom = 0.06
		top    = 0.94
		hgap   = 0.024
		vgap   = 0.075
	)
	cellW := (right - left - hgap*float64(cols-1)) / cols
	cellH := (top - bottom - vgap*3) / 4

	for idx, mode := range interpolationModes {
		col := idx % cols
		row := idx / cols
		x0 := left + float64(col)*(cellW+hgap)
		y1 := top - float64(row)*(cellH+vgap)
		ax := fig.AddAxes(geom.Rect{
			Min: geom.Pt{X: x0, Y: y1 - cellH},
			Max: geom.Pt{X: x0 + cellW, Y: y1},
		})
		ax.SetTitle(mode)
		ax.XAxis.ShowTicks = false
		ax.XAxis.ShowLabels = false
		ax.YAxis.ShowTicks = false
		ax.YAxis.ShowLabels = false
		mode := mode
		ax.ImShow(data, core.ImShowOptions{
			Colormap:      optional.Of(cmap),
			VMin:          optional.Of(vmin),
			VMax:          optional.Of(vmax),
			Origin:        optional.Of(core.ImageOriginLower),
			Extent:        optional.Of(extent),
			Aspect:        optional.Of("auto"),
			Interpolation: optional.Of(mode),
		})
	}
	return fig
}

func Render() image.Image {
	fig := Plot()
	r, err := agg.New(figureWidth, figureHeight, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.Image()
}

func interpolationMatrixData() [][]float64 {
	const n = 16
	data := make([][]float64, n)
	for y := range data {
		row := make([]float64, n)
		for x := range row {
			checker := 0.0
			if (x+y)%2 == 0 {
				checker = 0.35
			}
			wave := 0.25 * math.Sin(float64(x)*0.95) * math.Cos(float64(y)*0.7)
			gradient := 0.35 * float64(x+y) / float64(2*(n-1))
			row[x] = clampUnit(0.20 + checker + wave + gradient)
		}
		data[y] = row
	}
	return data
}

func clampUnit(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
