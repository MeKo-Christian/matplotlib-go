package fill_stacked

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/backends/agg"
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
	ax.SetTitle("Stacked Fills")
	ax.SetXLim(0, 8)
	ax.SetYLim(0, 8)

	x := []float64{1, 2, 3, 4, 5, 6, 7}
	zeros := make([]float64, len(x))
	layer1 := []float64{1, 1.5, 2, 1.8, 2.2, 1.9, 1.6}
	layer2 := make([]float64, len(layer1))
	layer3 := make([]float64, len(layer1))
	for i := range layer1 {
		layer2[i] = layer1[i] + 1.5 + 0.3*math.Sin(float64(i))
		layer3[i] = layer2[i] + 1.2 + 0.4*math.Cos(float64(i))
	}

	fill1 := core.FillBetween(x, zeros, layer1, render.Color{R: 0.8, G: 0.2, B: 0.2, A: 0.8})
	fill1.EdgeColor = render.Color{R: 0.5, G: 0, B: 0, A: 1}
	fill1.EdgeWidth = 1.0

	fill2 := core.FillBetween(x, layer1, layer2, render.Color{R: 0.2, G: 0.8, B: 0.2, A: 0.8})
	fill2.EdgeColor = render.Color{R: 0, G: 0.5, B: 0, A: 1}
	fill2.EdgeWidth = 1.0

	fill3 := core.FillBetween(x, layer2, layer3, render.Color{R: 0.2, G: 0.2, B: 0.8, A: 0.8})
	fill3.EdgeColor = render.Color{R: 0, G: 0, B: 0.5, A: 1}
	fill3.EdgeWidth = 1.0

	ax.Add(fill1)
	ax.Add(fill2)
	ax.Add(fill3)
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
