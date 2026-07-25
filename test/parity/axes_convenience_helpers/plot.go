package axes_convenience_helpers

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 840
	Height = 540
)

func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)

	boxAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.07, Y: 0.58}, Max: geom.Pt{X: 0.30, Y: 0.92}})
	boxAx.SetTitle("Bxp")
	boxAx.SetXLim(0.4, 2.6)
	boxAx.SetYLim(0, 6)
	meanA := 2.35
	meanB := 3.6
	boxAx.Bxp([]core.BxpStat{
		{Med: 2.2, Q1: 1.4, Q3: 3.1, Whislo: 0.8, Whishi: 4.1, Mean: &meanA, Fliers: []float64{5.0}, Label: "A"},
		{Med: 3.4, Q1: 2.5, Q3: 4.2, Whislo: 1.5, Whishi: 5.0, Mean: &meanB, Label: "B"},
	}, core.BxpOptions{
		ShowMeans: boolPtr(true),
		Color:     colorPtr(render.Color{R: 0.18, G: 0.36, B: 0.70, A: 1}),
	})

	violinAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.38, Y: 0.58}, Max: geom.Pt{X: 0.61, Y: 0.92}})
	violinAx.SetTitle("Violin")
	violinAx.SetXLim(0.4, 2.6)
	violinAx.SetYLim(0, 6)
	violinAx.Violin([]core.ViolinStat{
		{Coords: []float64{1, 2, 3, 4, 5}, Vals: []float64{0.2, 0.7, 1.0, 0.5, 0.15}, Mean: 3.0, Median: 3.0, Min: 1, Max: 5, Quantiles: []float64{2, 4}},
		{Coords: []float64{1, 2, 3, 4, 5}, Vals: []float64{0.15, 0.5, 1.0, 0.7, 0.2}, Mean: 3.2, Median: 3.3, Min: 1, Max: 5},
	}, core.ViolinStatsOptions{
		ShowMeans:   boolPtr(true),
		ShowMedians: boolPtr(true),
	})

	lineAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.69, Y: 0.58}, Max: geom.Pt{X: 0.96, Y: 0.92}})
	lineAx.SetTitle("H/VLines")
	lineAx.SetXLim(0, 5)
	lineAx.SetYLim(0, 5)
	lineColor := render.Color{R: 0.12, G: 0.30, B: 0.72, A: 1}
	lineAx.HLines([]float64{1, 2.5, 4}, []float64{0.5}, []float64{4.5}, core.LineCollectionOptions{
		Alpha:     optional.Of(1.0),
		Color:     optional.Of(lineColor),
		LineWidth: optional.Of(1.4),
	})
	lineAx.VLines([]float64{1, 2.5, 4}, []float64{0.6}, []float64{4.4}, core.LineCollectionOptions{
		Alpha:     optional.Of(1.0),
		Color:     optional.Of(render.Color{R: 0.75, G: 0.18, B: 0.16, A: 1}),
		LineWidth: optional.Of(1.4),
	})

	contourAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.18, Y: 0.08}, Max: geom.Pt{X: 0.82, Y: 0.43}})
	contourAx.SetTitle("Clabel")
	contourAx.SetXLim(0, 3)
	contourAx.SetYLim(0, 3)
	contourLineWidth := 1.0
	contours := contourAx.Contour([][]float64{
		{0, 1, 2, 3},
		{1, 2, 3, 4},
		{2, 3, 4, 5},
		{3, 4, 5, 6},
	}, core.ContourOptions{
		Levels:    []float64{2, 3, 4},
		Color:     colorPtr(render.Color{R: 0.13, G: 0.20, B: 0.35, A: 1}),
		LineWidth: &contourLineWidth,
	})
	fontSize := 9.0
	contourAx.Clabel(contours, core.ClabelOptions{
		Levels:   []float64{3},
		FontSize: &fontSize,
		Color:    colorPtr(render.Color{R: 0.05, G: 0.05, B: 0.05, A: 1}),
	})

	return fig
}

func Render() image.Image {
	r, err := agg.New(Width, Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(Plot(), r)
	return r.Image()
}

func boolPtr(v bool) *bool { return &v }

func colorPtr(v render.Color) *render.Color { return &v }
