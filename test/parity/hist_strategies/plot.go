package hist_strategies

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
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
		Min: geom.Pt{X: 0.12, Y: 0.12},
		Max: geom.Pt{X: 0.95, Y: 0.90},
	})
	ax.SetTitle("Histogram Strategies")

	data1 := common.NormalData(42, 0, 300, 4.0, 1.0)
	data2 := common.NormalData(7, 0, 300, 7.0, 1.2)

	blue := render.Color{R: 0.26, G: 0.53, B: 0.80, A: 0.6}
	orange := render.Color{R: 0.90, G: 0.50, B: 0.10, A: 0.6}
	black := render.Color{R: 0, G: 0, B: 0, A: 1}
	ew := 0.5
	bins := 15
	prob := core.HistNormProbability

	ax.Hist(data1, core.HistOptions{
		Bins:      bins,
		Norm:      prob,
		Color:     &blue,
		EdgeColor: &black,
		EdgeWidth: &ew,
	})
	ax.Hist(data2, core.HistOptions{
		Bins:      bins,
		Norm:      prob,
		Color:     &orange,
		EdgeColor: &black,
		EdgeWidth: &ew,
	})
	ax.AutoScale(0.05)
	_, yMax := ax.YScale.Domain()
	ax.SetYLim(0, yMax)
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
	return r.GetImage()
}
