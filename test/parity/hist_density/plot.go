package hist_density

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
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
		Min: geom.Pt{X: 0.12, Y: 0.12},
		Max: geom.Pt{X: 0.95, Y: 0.90},
	})
	ax.SetTitle("Density Histogram")

	data := common.NormalData(42, 0, 500, 5.0, 1.5)

	green := render.Color{R: 0.20, G: 0.65, B: 0.30, A: 0.8}
	black := render.Color{R: 0, G: 0, B: 0, A: 1}
	ew := 0.8
	bins := 20
	_, _ = ax.Hist(data, core.HistOptions{
		Bins:      bins,
		Norm:      core.HistNormDensity,
		Color:     optional.Of(green),
		EdgeColor: optional.Of(black),
		EdgeWidth: optional.Of(ew),
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
	return r.Image()
}
