package scatter_advanced

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
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
	ax.SetTitle("Advanced Scatter")
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 10)

	points := []geom.Pt{
		{X: 2, Y: 2},
		{X: 4, Y: 4},
		{X: 6, Y: 6},
		{X: 8, Y: 8},
		{X: 2, Y: 8},
		{X: 4, Y: 6},
		{X: 6, Y: 4},
		{X: 8, Y: 2},
	}
	radii := []float64{6, 10, 14, 18, 8, 12, 16, 20}
	sizes := make([]float64, len(radii))
	for i, radius := range radii {
		sizes[i] = core.ScatterAreaFromRadius(radius, style.Default.DPI)
	}
	fillColors := []render.Color{
		{R: 1, G: 0.5, B: 0.5, A: 1},
		{R: 0.5, G: 1, B: 0.5, A: 1},
		{R: 0.5, G: 0.5, B: 1, A: 1},
		{R: 1, G: 1, B: 0.5, A: 1},
		{R: 1, G: 0.5, B: 1, A: 1},
		{R: 0.5, G: 1, B: 1, A: 1},
		{R: 0.8, G: 0.8, B: 0.8, A: 1},
		{R: 0.3, G: 0.3, B: 0.3, A: 1},
	}
	edgeColors := []render.Color{
		{R: 0.5, G: 0, B: 0, A: 1},
		{R: 0, G: 0.5, B: 0, A: 1},
		{R: 0, G: 0, B: 0.5, A: 1},
		{R: 0.5, G: 0.5, B: 0, A: 1},
		{R: 0.5, G: 0, B: 0.5, A: 1},
		{R: 0, G: 0.5, B: 0.5, A: 1},
		{R: 0.4, G: 0.4, B: 0.4, A: 1},
		{R: 0, G: 0, B: 0, A: 1},
	}

	scatter := &core.Scatter2D{
		XY:         points,
		Sizes:      sizes,
		Colors:     fillColors,
		EdgeColors: edgeColors,
		EdgeWidth:  2.0,
		Alpha:      0.8,
		Marker:     core.MarkerCircle,
	}
	ax.Add(scatter)
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
