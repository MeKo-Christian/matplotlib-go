package line2d_semantics

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
)

func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.14},
		Max: geom.Pt{X: 0.94, Y: 0.88},
	})
	ax.SetTitle("Line2D Semantics")
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 5)

	mutated := ax.Plot(
		[]float64{0.6, 2.2, 3.8},
		[]float64{4.65, 4.65, 4.65},
		core.PlotOptions{
			Color:     colorPtr(render.Color{R: 0.10, G: 0.26, B: 0.78, A: 1}),
			LineWidth: floatPtr(3),
			Label:     "set_data",
		},
	)
	mutated.SetData(
		[]float64{0.6, 1.8, 3.0, 4.2},
		[]float64{4.55, 4.25, 4.65, 4.35},
	)

	invalid := &core.Line2D{
		XY: []geom.Pt{
			{X: 0.6, Y: 3.45},
			{X: 1.6, Y: 3.85},
			{X: math.NaN(), Y: 3.1},
			{X: 2.4, Y: 3.15},
			{X: 3.5, Y: 3.75},
			{X: math.Inf(1), Y: 3.4},
			{X: 4.3, Y: 3.25},
			{X: 5.2, Y: 3.65},
		},
		W:   3,
		Col: render.Color{R: 0.84, G: 0.15, B: 0.16, A: 1},
	}
	ax.Add(invalid)

	gap := &core.Line2D{
		XY: []geom.Pt{
			{X: 5.6, Y: 3.55},
			{X: 9.4, Y: 3.55},
		},
		W:   4,
		Col: render.Color{R: 0.08, G: 0.45, B: 0.18, A: 1},
	}
	gap.SetDashes(4, 2)
	gap.SetGapColor(render.Color{R: 0.92, G: 0.48, B: 0.10, A: 1})
	ax.Add(gap)

	x := []float64{0.7, 1.5, 2.3, 3.1, 3.9, 4.7, 5.5, 6.3, 7.1, 7.9, 8.7, 9.5}
	y1 := []float64{1.1, 1.45, 1.2, 1.65, 1.35, 1.75, 1.45, 1.85, 1.55, 1.95, 1.65, 2.05}
	markerA := core.MarkerCircle
	lineA := ax.Plot(x, y1, core.PlotOptions{
		Color:      colorPtr(render.Color{R: 0.58, G: 0.40, B: 0.74, A: 1}),
		LineWidth:  floatPtr(1.5),
		Marker:     &markerA,
		MarkerSize: floatPtr(8),
	})
	lineA.SetMarkEvery(core.StartStepMarkers(1, 3))

	y2 := []float64{0.55, 0.85, 0.65, 0.95, 0.75, 1.05, 0.85, 1.15, 0.95, 1.25, 1.05, 1.35}
	markerB := core.MarkerSquare
	lineB := ax.Plot(x, y2, core.PlotOptions{
		Color:      colorPtr(render.Color{R: 0.55, G: 0.34, B: 0.29, A: 1}),
		LineWidth:  floatPtr(1.5),
		Marker:     &markerB,
		MarkerSize: floatPtr(7),
	})
	lineB.SetMarkEvery(core.IndexedMarkers(0, 4, 8, 11))

	return fig
}

func Render() image.Image {
	fig := Plot()
	r, err := agg.New(Width, Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.Image()
}

func colorPtr(c render.Color) *render.Color { return &c }

func floatPtr(v float64) *float64 { return &v }
