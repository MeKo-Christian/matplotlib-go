// Package sketch_xkcd renders a simple plot with Matplotlib's xkcd sketch mode
// enabled, exercising the global path.sketch perturbation across spines, ticks,
// and line artists.
package sketch_xkcd

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

const (
	Width  = 640
	Height = 360
)

// Plot builds the figure. xkcd mode is enabled globally via style.WithXkcd so
// every drawn path — spines, ticks, and the curves — receives the sketch wiggle.
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height, style.WithXkcd())
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.14},
		Max: geom.Pt{X: 0.94, Y: 0.88},
	})
	ax.SetXLim(0, 10)
	ax.SetYLim(-1.2, 1.2)

	// A smooth sine curve: the sketch filter turns it into a hand-drawn wiggle.
	const n = 200
	xs := make([]float64, n)
	ys := make([]float64, n)
	for i := range xs {
		x := float64(i) / float64(n-1) * 10
		xs[i] = x
		ys[i] = math.Sin(x)
	}
	ax.Plot(xs, ys, core.PlotOptions{
		Color:     colorPtr(render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1}),
		LineWidth: floatPtr(2),
	})

	// A straight reference line, to show the wiggle on an otherwise flat path.
	ax.Plot([]float64{0, 10}, []float64{0, 0}, core.PlotOptions{
		Color:     colorPtr(render.Color{R: 0.84, G: 0.15, B: 0.16, A: 1}),
		LineWidth: floatPtr(2),
	})

	return fig
}

// Render rasterizes the figure with the AGG backend (the parity reference).
func Render() image.Image {
	fig := Plot()
	r, err := agg.New(Width, Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	// xkcd mode renders the figure background as Matplotlib does: a non-
	// antialiased figure patch over a transparent canvas, so the sketch wiggle
	// perforates the border with fully-transparent notch pixels. Return the
	// buffer as straight-alpha NRGBA so those transparent pixels round-trip and
	// compare correctly (GetImage would mislabel them as premultiplied RGBA).
	return r.GetImageNRGBA()
}

func colorPtr(c render.Color) *render.Color { return &c }
func floatPtr(f float64) *float64           { return &f }
