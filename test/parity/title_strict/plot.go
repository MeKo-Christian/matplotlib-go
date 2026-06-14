package title_strict

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func Render() image.Image {
	const (
		w = 320
		h = 280
	)

	fig := core.NewFigure(w, h)
	titles := []string{
		"Histogram Strategies",
		"Fill to Baseline",
		"Dash Patterns",
		"Box Plots",
		"Text Labels",
	}
	rows := []geom.Rect{
		{Min: geom.Pt{X: 0.05, Y: 0.20}, Max: geom.Pt{X: 0.95, Y: 0.28}},
		{Min: geom.Pt{X: 0.05, Y: 0.36}, Max: geom.Pt{X: 0.95, Y: 0.44}},
		{Min: geom.Pt{X: 0.05, Y: 0.52}, Max: geom.Pt{X: 0.95, Y: 0.60}},
		{Min: geom.Pt{X: 0.05, Y: 0.68}, Max: geom.Pt{X: 0.95, Y: 0.76}},
		{Min: geom.Pt{X: 0.05, Y: 0.84}, Max: geom.Pt{X: 0.95, Y: 0.92}},
	}
	for i, title := range titles {
		ax := fig.AddAxes(rows[i])
		ax.SetTitle(title)
		ax.SetXLim(0, 1)
		ax.SetYLim(0, 1)
		ax.XAxis.ShowSpine = false
		ax.XAxis.ShowTicks = false
		ax.XAxis.ShowLabels = false
		ax.YAxis.ShowSpine = false
		ax.YAxis.ShowTicks = false
		ax.YAxis.ShowLabels = false
		ax.ShowFrame = false
	}

	r, err := agg.New(w, h, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}
