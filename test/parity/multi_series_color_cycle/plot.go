package multi_series_color_cycle

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
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
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.SetTitle("Color Cycle")
	ax.SetXLim(0, 2*math.Pi)
	ax.SetYLim(-1.2, 1.2)

	nPoints := 50
	x := make([]float64, nPoints)
	for i := 0; i < nPoints; i++ {
		x[i] = 2 * math.Pi * float64(i) / float64(nPoints-1)
	}

	tab10 := color.Tab10
	lineWidth := 2.0
	for seriesIdx, freq := range []int{1, 2, 3, 4} {
		y := make([]float64, nPoints)
		for j := 0; j < nPoints; j++ {
			y[j] = math.Sin(float64(freq) * x[j])
		}
		_, _ = ax.Plot(x, y, core.PlotOptions{
			Color:     &tab10[seriesIdx],
			LineWidth: &lineWidth,
		})
	}
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
