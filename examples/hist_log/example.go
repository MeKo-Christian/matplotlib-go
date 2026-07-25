package hist_log

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
//
// It mirrors hist_basic but sets HistOptions.Log, so the count (y) axis renders
// on a logarithmic scale — matplotlib's hist(log=True). The log axis reveals the
// distribution tails that a linear count axis compresses to the baseline.
func Plot() *core.Figure {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.12},
		Max: geom.Pt{X: 0.95, Y: 0.90},
	})
	ax.SetTitle("Logarithmic Histogram")

	data := common.NormalData(42, 0, 500, 5.0, 1.5)

	blue := render.Color{R: 0.26, G: 0.53, B: 0.80, A: 0.8}
	black := render.Color{R: 0, G: 0, B: 0, A: 1}
	ew := 0.8
	_, _ = ax.Hist(data, core.HistOptions{
		Color:     &blue,
		EdgeColor: &black,
		EdgeWidth: &ew,
		Log:       true,
	})

	// Match the x margin of AutoScale(0.05); the log y axis uses a fixed
	// positive window so the reference plot can pin identical limits.
	ax.AutoScale(0.05)
	ax.SetYLim(1, 200)
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
