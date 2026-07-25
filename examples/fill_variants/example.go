// Package fill_variants is a user-facing showcase that gathers the fill/area
// surface into one figure: fill_between two curves, fill_betweenx, stacked
// fills, and translucent overlapping areas. It closes the Phase 18.1
// "fill-variants" demo-breadth gap by promoting behavior that previously only
// existed as parity fixtures (fill_between, fill_stacked, plot_variants) into a
// single browsable gallery.
package fill_variants

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 840
	Height = 620
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)

	addFillBetweenPanel(fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.07, Y: 0.585}, Max: geom.Pt{X: 0.46, Y: 0.93}}))
	addFillBetweenXPanel(fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.56, Y: 0.585}, Max: geom.Pt{X: 0.96, Y: 0.93}}))
	addStackedPanel(fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.07, Y: 0.10}, Max: geom.Pt{X: 0.46, Y: 0.445}}))
	addAlphaOverlapPanel(fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.56, Y: 0.10}, Max: geom.Pt{X: 0.96, Y: 0.445}}))

	return fig
}

// addFillBetweenPanel mirrors the fill_between fixture.
func addFillBetweenPanel(ax *core.Axes) {
	ax.SetTitle("Fill Between")
	ax.SetXLim(0, 6.28)
	ax.SetYLim(-1.5, 1.5)

	n := 50
	x := make([]float64, n)
	y1 := make([]float64, n)
	y2 := make([]float64, n)
	for i := 0; i < n; i++ {
		t := 6.28 * float64(i) / float64(n-1)
		x[i] = t
		y1[i] = math.Sin(t)
		y2[i] = 0.8 * math.Cos(t)
	}
	fill := core.FillBetween(x, y1, y2, render.Color{R: 0.8, G: 0.3, B: 0.3, A: 0.6})
	fill.EdgeColor = render.Color{R: 0.5, G: 0.1, B: 0.1, A: 1.0}
	fill.EdgeWidth = 1.5
	ax.Add(fill)

	sineLine := &core.Line2D{XY: make([]geom.Pt, n), W: 2.0, Col: render.Color{R: 1, G: 0, B: 0, A: 1}}
	cosLine := &core.Line2D{XY: make([]geom.Pt, n), W: 2.0, Col: render.Color{R: 0, G: 0, B: 1, A: 1}}
	for i := 0; i < n; i++ {
		sineLine.XY[i] = geom.Pt{X: x[i], Y: y1[i]}
		cosLine.XY[i] = geom.Pt{X: x[i], Y: y2[i]}
	}
	ax.Add(sineLine)
	ax.Add(cosLine)
}

// addFillBetweenXPanel mirrors the plot_variants fill_betweenx panel.
func addFillBetweenXPanel(ax *core.Axes) {
	ax.SetTitle("Fill BetweenX")
	ax.SetXLim(0, 7)
	ax.SetYLim(0, 6)
	edgeWidth := 1.2
	ax.FillBetweenX(
		[]float64{0.4, 1.2, 2.0, 2.8, 3.6, 4.4, 5.2},
		[]float64{1.3, 2.1, 1.7, 2.8, 2.2, 3.1, 2.6},
		[]float64{3.4, 4.1, 4.8, 5.1, 5.6, 6.0, 6.3},
		core.FillOptions{
			Color:     &render.Color{R: 0.24, G: 0.68, B: 0.54, A: 0.72},
			EdgeColor: &render.Color{R: 0.12, G: 0.38, B: 0.28, A: 1},
			EdgeWidth: &edgeWidth,
		},
	)
}

// addStackedPanel mirrors the fill_stacked fixture.
func addStackedPanel(ax *core.Axes) {
	ax.SetTitle("Stacked Fills")
	ax.SetXLim(0, 8)
	ax.SetYLim(0, 8)

	x := []float64{1, 2, 3, 4, 5, 6, 7}
	zeros := make([]float64, len(x))
	layer1 := []float64{1, 1.5, 2, 1.8, 2.2, 1.9, 1.6}
	layer2 := make([]float64, len(layer1))
	layer3 := make([]float64, len(layer1))
	for i := range layer1 {
		layer2[i] = layer1[i] + 1.5 + 0.3*math.Sin(float64(i))
		layer3[i] = layer2[i] + 1.2 + 0.4*math.Cos(float64(i))
	}
	fill1 := core.FillBetween(x, zeros, layer1, render.Color{R: 0.8, G: 0.2, B: 0.2, A: 0.8})
	fill1.EdgeColor = render.Color{R: 0.5, G: 0, B: 0, A: 1}
	fill1.EdgeWidth = 1.0
	fill2 := core.FillBetween(x, layer1, layer2, render.Color{R: 0.2, G: 0.8, B: 0.2, A: 0.8})
	fill2.EdgeColor = render.Color{R: 0, G: 0.5, B: 0, A: 1}
	fill2.EdgeWidth = 1.0
	fill3 := core.FillBetween(x, layer2, layer3, render.Color{R: 0.2, G: 0.2, B: 0.8, A: 0.8})
	fill3.EdgeColor = render.Color{R: 0, G: 0, B: 0.5, A: 1}
	fill3.EdgeWidth = 1.0
	ax.Add(fill1)
	ax.Add(fill2)
	ax.Add(fill3)
}

// addAlphaOverlapPanel overlaps two translucent fill-to-baseline areas.
func addAlphaOverlapPanel(ax *core.Axes) {
	ax.SetTitle("Alpha Overlap")
	ax.SetXLim(0, 6)
	ax.SetYLim(0, 5)

	n := 50
	x := make([]float64, n)
	yA := make([]float64, n)
	yB := make([]float64, n)
	for i := 0; i < n; i++ {
		t := 6.0 * float64(i) / float64(n-1)
		x[i] = t
		yA[i] = 2.5 + 1.6*math.Sin(t)
		yB[i] = 2.5 + 1.6*math.Cos(t)
	}
	fillA := core.FillToBaseline(x, yA, 0, render.Color{R: 0.85, G: 0.25, B: 0.25, A: 0.45})
	fillB := core.FillToBaseline(x, yB, 0, render.Color{R: 0.20, G: 0.35, B: 0.85, A: 0.45})
	ax.Add(fillA)
	ax.Add(fillB)
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
