// Package bar_variants is a user-facing showcase that gathers the bar-chart
// surface into one figure: vertical categorical bars with tick labels,
// horizontal bars, grouped bars, and stacked bars with bar labels. It closes
// the Phase 18.1 "bar-variants" demo-breadth gap by promoting behavior that
// previously only existed as parity fixtures (bar_horizontal, bar_grouped,
// bar_basic_tick_labels, plot_variants) into a single browsable gallery.
package bar_variants

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/ticker"
)

const (
	Width  = 840
	Height = 620
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)

	addVerticalLabeledPanel(fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.07, Y: 0.585}, Max: geom.Pt{X: 0.46, Y: 0.93}}))
	addHorizontalPanel(fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.56, Y: 0.585}, Max: geom.Pt{X: 0.96, Y: 0.93}}))
	addGroupedPanel(fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.07, Y: 0.10}, Max: geom.Pt{X: 0.46, Y: 0.445}}))
	addStackedLabeledPanel(fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.56, Y: 0.10}, Max: geom.Pt{X: 0.96, Y: 0.445}}))

	return fig
}

// addVerticalLabeledPanel draws categorical vertical bars with custom tick
// labels.
func addVerticalLabeledPanel(ax *core.Axes) {
	ax.SetTitle("Vertical + Tick Labels")
	ax.SetXLim(0.4, 5.6)
	ax.SetYLim(0, 10)
	ax.XAxis.Locator = ticker.FixedLocator{TicksList: []float64{1, 2, 3, 4, 5}}
	ax.XAxis.Formatter = ticker.FixedFormatter{Labels: []string{"alpha", "beta", "gamma", "delta", "eps"}}

	color := render.Color{R: 0.20, G: 0.60, B: 0.80, A: 1}
	width := 0.6
	_, _ = ax.Bar(
		[]float64{1, 2, 3, 4, 5},
		[]float64{3, 7, 2, 8, 5},
		core.BarOptions{Color: optional.Of(color), Width: optional.Of(width)},
	)
}

// addHorizontalPanel mirrors the bar_horizontal fixture.
func addHorizontalPanel(ax *core.Axes) {
	ax.SetTitle("Horizontal Bars")
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 6)
	color := render.Color{R: 0.8, G: 0.4, B: 0.2, A: 1}
	height := 0.6
	orientation := core.BarHorizontal
	_, _ = ax.Bar(
		[]float64{1, 2, 3, 4, 5},
		[]float64{3, 7, 2, 8, 5},
		core.BarOptions{Width: optional.Of(height), Color: optional.Of(color), Orientation: optional.Of(orientation)},
	)
}

// addGroupedPanel mirrors the bar_grouped fixture.
func addGroupedPanel(ax *core.Axes) {
	ax.SetTitle("Grouped Bars")
	ax.SetXLim(0, 7)
	ax.SetYLim(0, 10)
	ax.Add(&core.Bar2D{
		X:           []float64{1.2, 2.2, 3.2, 4.2, 5.2},
		Heights:     []float64{3, 7, 2, 8, 5},
		Width:       0.35,
		Color:       render.Color{R: 0.8, G: 0.2, B: 0.2, A: 1},
		EdgeColor:   render.Color{R: 0.5, G: 0, B: 0, A: 1},
		EdgeWidth:   1.0,
		Baseline:    0,
		Orientation: core.BarVertical,
	})
	ax.Add(&core.Bar2D{
		X:           []float64{1.8, 2.8, 3.8, 4.8, 5.8},
		Heights:     []float64{5, 4, 6, 3, 7},
		Width:       0.35,
		Color:       render.Color{R: 0.2, G: 0.8, B: 0.2, A: 1},
		EdgeColor:   render.Color{R: 0, G: 0.5, B: 0, A: 1},
		EdgeWidth:   1.0,
		Baseline:    0,
		Orientation: core.BarVertical,
	})
}

// addStackedLabeledPanel mirrors the plot_variants stacked-bar panel.
func addStackedLabeledPanel(ax *core.Axes) {
	ax.SetTitle("Stacked + Bar Labels")
	ax.SetXLim(0.4, 4.6)
	ax.SetYLim(0, 7.6)
	ax.AddYGrid()
	x := []float64{1, 2, 3, 4}
	base := []float64{0, 0, 0, 0}
	seriesA := []float64{1.4, 2.2, 1.8, 2.5}
	seriesB := []float64{2.1, 1.6, 2.4, 1.7}
	bottom, _ := ax.Bar(x, seriesA, core.BarOptions{
		Baselines: base,
		Color:     optional.Of(render.Color{R: 0.16, G: 0.59, B: 0.49, A: 1}),
	})
	top, _ := ax.Bar(x, seriesB, core.BarOptions{
		Baselines: seriesA,
		Color:     optional.Of(render.Color{R: 0.88, G: 0.47, B: 0.16, A: 1}),
	})
	ax.BarLabel(bottom, []string{"A1", "A2", "A3", "A4"}, core.BarLabelOptions{
		Position: "center",
		Color:    render.Color{R: 1, G: 1, B: 1, A: 1},
		FontSize: 10,
	})
	ax.BarLabel(top, nil, core.BarLabelOptions{
		Format:   "%.1f",
		Color:    render.Color{R: 0.20, G: 0.20, B: 0.20, A: 1},
		FontSize: 10,
		Padding:  4,
	})
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
