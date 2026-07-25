package gridspec_composition

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 960
	Height = 640
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(960, 640)
	common.ApplyMatplotlibGridSpecStyle(fig)

	outer := fig.GridSpec(
		2,
		2,
		core.WithGridSpecPadding(0.08, 0.96, 0.10, 0.92),
		core.WithGridSpecSpacing(0.06/(2+0.06), 0.28/(2+0.28)),
		core.WithGridSpecWidthRatios(2, 1),
	)

	mainAx := outer.Span(0, 0, 2, 1).AddAxes()
	common.ConfigureCompositionAxes(mainAx, "Main Span", []float64{0, 1, 2, 3, 4}, []float64{1.2, 2.8, 2.1, 3.6, 3.1}, render.Color{R: 0.15, G: 0.35, B: 0.72, A: 1})
	common.ConfigureCompositionTicks(mainAx, []float64{0, 1, 2, 3, 4}, []float64{1.0, 1.5, 2.0, 2.5, 3.0, 3.5}, "%.1f")

	nested := outer.Cell(0, 1).GridSpec(2, 1, core.WithGridSpecSpacing(0, 0.75/(2+0.75)))
	topRight := nested.Cell(0, 0).AddAxes()
	common.ConfigureCompositionAxes(topRight, "Nested Top", []float64{0, 1, 2, 3}, []float64{3.4, 2.6, 2.9, 1.8}, render.Color{R: 0.72, G: 0.32, B: 0.18, A: 1})
	common.ConfigureCompositionTicks(topRight, []float64{0, 1, 2, 3}, []float64{2, 3}, "%.0f")

	bottomRight := nested.Cell(1, 0).AddAxes(core.WithSharedX(topRight))
	common.ConfigureCompositionAxes(bottomRight, "Nested Bottom", []float64{0, 1, 2, 3}, []float64{1.0, 1.6, 1.3, 2.2}, render.Color{R: 0.18, G: 0.55, B: 0.34, A: 1})
	common.ConfigureCompositionTicks(bottomRight, []float64{0, 1, 2, 3}, []float64{1, 2}, "%.0f")

	sub := outer.Cell(1, 1).SubFigure()
	inset := sub.AddSubplot(1, 1, 1)
	common.ConfigureCompositionAxes(inset, "SubFigure", []float64{0, 1, 2, 3}, []float64{2.0, 2.4, 1.9, 2.7}, render.Color{R: 0.55, G: 0.22, B: 0.50, A: 1})
	common.ConfigureCompositionTicks(inset, []float64{0, 1, 2, 3}, []float64{2.0, 2.2, 2.4, 2.6}, "%.1f")
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
