// Package pyplot_workflow demonstrates a compact stateful pyplot migration flow.
package pyplot_workflow

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/pyplot"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 1000
	Height = 680
)

// Plot builds a stateful pyplot figure using common migration wrappers.
func Plot() *core.Figure {
	must(pyplot.CloseAll())
	pyplot.RCDefaults()

	fig, axes := pyplot.Subplots(2, 2)
	fig.SizePx.X = Width
	fig.SizePx.Y = Height
	pyplot.Suptitle("Pyplot workflow")

	x := linspace(0, 2*math.Pi, 80)
	y := make([]float64, len(x))
	for i, v := range x {
		y[i] = math.Sin(v)
	}

	must(pyplot.SCA(axes[0][0]))
	pyplot.Plot(x, y, core.PlotOptions{Label: "sin(x)"})
	pyplot.XLabel("x")
	pyplot.YLabel("signal")
	pyplot.Title("line")
	pyplot.Legend()

	must(pyplot.SCA(axes[0][1]))
	pyplot.Scatter([]float64{0, 1, 2, 3, 4}, []float64{0.1, 1.3, 0.8, 1.9, 1.4}, core.ScatterOptions{Label: "samples"})
	pyplot.Annotate("peak", 3, 1.9)
	pyplot.Title("scatter")
	pyplot.Legend()

	must(pyplot.SCA(axes[1][0]))
	pyplot.Bar([]float64{0, 1, 2}, []float64{2.2, 3.0, 1.7}, core.BarOptions{Label: "counts"})
	pyplot.XLabel("group")
	pyplot.YLabel("count")
	pyplot.Title("bar")
	pyplot.Legend()

	must(pyplot.SCA(axes[1][1]))
	interp := "bilinear"
	img := pyplot.ImShow(heatmap(), core.ImShowOptions{Interpolation: &interp})
	pyplot.Title("imshow")
	pyplot.Colorbar(img)

	return fig
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

// Render returns an AGG-rendered preview image for tests and docs.
func Render() image.Image {
	fig := Plot()
	r, err := agg.New(Width, Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.Image()
}

func linspace(start, stop float64, n int) []float64 {
	out := make([]float64, n)
	if n == 1 {
		out[0] = start
		return out
	}
	step := (stop - start) / float64(n-1)
	for i := range out {
		out[i] = start + float64(i)*step
	}
	return out
}

func heatmap() [][]float64 {
	data := make([][]float64, 12)
	for y := range data {
		data[y] = make([]float64, 16)
		for x := range data[y] {
			fx := float64(x) / 15
			fy := float64(y) / 11
			data[y][x] = math.Sin(fx*math.Pi) * math.Cos(fy*math.Pi/2)
		}
	}
	return data
}
