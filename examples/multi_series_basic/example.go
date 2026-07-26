package multi_series_basic

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
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
	ax.SetTitle("Multi-Series Plot")
	ax.SetXLim(0, 8)
	ax.SetYLim(0, 6)

	tab10 := color.Tab10
	lineWidth := 2.0
	_, _ = ax.Plot(
		[]float64{1, 2, 3, 4, 5, 6},
		[]float64{1.5, 2.8, 2.2, 3.5, 3.8, 4.2},
		core.PlotOptions{
			Color:     optional.Of(tab10[0]),
			LineWidth: optional.Of(lineWidth),
		},
	)

	size := core.ScatterAreaFromRadius(8.0, style.Default.DPI)
	edgeWidth := 0.0
	ax.Scatter(
		[]float64{1.5, 2.5, 3.5, 4.5, 5.5},
		[]float64{2.2, 3.1, 2.9, 4.1, 4.5},
		core.ScatterOptions{
			Color:     optional.Of(tab10[1]),
			Size:      optional.Of(size),
			EdgeWidth: optional.Of(edgeWidth),
		},
	)

	width := 0.4
	_, _ = ax.Bar(
		[]float64{2, 3, 4, 5},
		[]float64{3.8, 2.5, 4.8, 3.2},
		core.BarOptions{
			Color: optional.Of(tab10[2]),
			Width: optional.Of(width),
		},
	)
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
