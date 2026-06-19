package stat_variants

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 840
	Height = 620
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(840, 620)
	grid := fig.Subplots(
		2,
		2,
		core.WithSubplotPadding(0.08, 0.97, 0.10, 0.93),
		core.WithSubplotSpacing(0.10, 0.14),
	)

	stackAx := grid[0][0]
	stackAx.SetTitle("StackPlot")
	stackAx.SetXLim(0, 5)
	stackAx.SetYLim(0, 7)
	stackAx.SetAxisBelow(true)
	stackAx.AddYGrid()
	stackAx.StackPlot(
		[]float64{0, 1, 2, 3, 4, 5},
		[][]float64{
			{1.0, 1.4, 1.3, 1.8, 1.6, 2.0},
			{0.8, 1.1, 1.4, 1.2, 1.6, 1.8},
			{0.5, 0.8, 1.0, 1.4, 1.1, 1.5},
		},
		core.StackPlotOptions{
			Colors: []render.Color{
				{R: 0.20, G: 0.55, B: 0.75, A: 0.76},
				{R: 0.90, G: 0.48, B: 0.18, A: 0.76},
				{R: 0.35, G: 0.66, B: 0.42, A: 0.76},
			},
		},
	)

	ecdfAx := grid[0][1]
	ecdfAx.SetTitle("ECDF")
	ecdfAx.SetXLim(0, 8)
	ecdfAx.SetYLim(0, 1.05)
	ecdfAx.SetAxisBelow(true)
	ecdfAx.AddYGrid()
	ecdfAx.ECDF(
		[]float64{1.2, 1.8, 2.0, 2.0, 3.1, 3.7, 4.3, 5.0, 5.8, 6.6, 7.0},
		core.ECDFOptions{
			Color:     &render.Color{R: 0.18, G: 0.36, B: 0.75, A: 1},
			LineWidth: common.FloatPtr(2),
			Compress:  true,
		},
	)

	cumulativeAx := grid[1][0]
	cumulativeAx.SetTitle("Cumulative Step Hist")
	cumulativeAx.SetXLim(0, 6)
	cumulativeAx.SetYLim(0, 1.05)
	cumulativeAx.SetAxisBelow(true)
	cumulativeAx.AddYGrid()
	cumulativeAx.Hist(
		[]float64{0.4, 0.7, 1.2, 1.4, 2.1, 2.6, 3.1, 3.2, 4.0, 4.8, 5.2},
		core.HistOptions{
			BinEdges:   []float64{0, 1, 2, 3, 4, 5, 6},
			Norm:       core.HistNormProbability,
			Cumulative: true,
			HistType:   core.HistTypeStepFilled,
			Color:      &render.Color{R: 0.42, G: 0.62, B: 0.90, A: 0.55},
			EdgeColor:  &render.Color{R: 0.12, G: 0.25, B: 0.55, A: 1},
			EdgeWidth:  common.FloatPtr(1.4),
		},
	)

	multiAx := grid[1][1]
	multiAx.SetTitle("Stacked Multi-Hist")
	multiAx.SetXLim(0, 6)
	multiAx.SetYLim(0, 6)
	multiAx.SetAxisBelow(true)
	multiAx.AddYGrid()
	multiAx.HistMulti(
		[][]float64{
			{0.3, 0.8, 1.2, 1.7, 2.6, 3.4, 4.1, 5.2},
			{0.5, 1.1, 1.9, 2.3, 2.8, 3.0, 3.7, 4.5, 5.0},
			{1.0, 1.6, 2.2, 2.9, 3.5, 4.2, 4.8, 5.4},
		},
		core.MultiHistOptions{
			BinEdges: []float64{0, 1, 2, 3, 4, 5, 6},
			Stacked:  true,
			Colors: []render.Color{
				{R: 0.22, G: 0.55, B: 0.70, A: 0.8},
				{R: 0.86, G: 0.42, B: 0.19, A: 0.8},
				{R: 0.36, G: 0.62, B: 0.36, A: 0.8},
			},
			EdgeColor: &render.Color{R: 0.10, G: 0.10, B: 0.10, A: 1},
			EdgeWidth: common.FloatPtr(0.7),
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
	return r.GetImage()
}
