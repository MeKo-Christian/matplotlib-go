// Package histogram_variants is a user-facing showcase that gathers the
// histogram surface into one figure: raw counts, a density-normalized
// histogram, a cumulative histogram, and two overlapping probability
// histograms. It closes the Phase 18.1 "histogram-variants" demo-breadth gap by
// promoting behavior that previously only existed as parity fixtures
// (hist_basic, hist_density, hist_strategies) into a single browsable gallery.
//
// Each panel uses explicit, identical x/y limits in the Go body and the
// matplotlib reference so the deterministic PCG-generated samples and bin edges
// are the only thing under test (panel autoscaling does not need to match).
package histogram_variants

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	common "github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 840
	Height = 620
	DPI    = 100
)

var black = render.Color{R: 0, G: 0, B: 0, A: 1}

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)

	addCountsPanel(fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.08, Y: 0.585}, Max: geom.Pt{X: 0.46, Y: 0.93}}))
	addDensityPanel(fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.57, Y: 0.585}, Max: geom.Pt{X: 0.96, Y: 0.93}}))
	addCumulativePanel(fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.08, Y: 0.10}, Max: geom.Pt{X: 0.46, Y: 0.445}}))
	addMultiplePanel(fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.57, Y: 0.10}, Max: geom.Pt{X: 0.96, Y: 0.445}}))

	return fig
}

// addCountsPanel draws a raw-count histogram.
func addCountsPanel(ax *core.Axes) {
	ax.SetTitle("Counts")
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 70)
	data := common.NormalData(1, 0, 400, 5.0, 1.5)
	color := render.Color{R: 0.26, G: 0.53, B: 0.80, A: 0.85}
	ew := 0.8
	bins := 18
	ax.Hist(data, core.HistOptions{Bins: bins, Color: &color, EdgeColor: &black, EdgeWidth: &ew})
}

// addDensityPanel mirrors the hist_density fixture content.
func addDensityPanel(ax *core.Axes) {
	ax.SetTitle("Density")
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 0.35)
	data := common.NormalData(42, 0, 500, 5.0, 1.5)
	color := render.Color{R: 0.20, G: 0.65, B: 0.30, A: 0.8}
	ew := 0.8
	bins := 20
	ax.Hist(data, core.HistOptions{Bins: bins, Norm: core.HistNormDensity, Color: &color, EdgeColor: &black, EdgeWidth: &ew})
}

// addCumulativePanel draws a cumulative-count histogram.
func addCumulativePanel(ax *core.Axes) {
	ax.SetTitle("Cumulative")
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 420)
	data := common.NormalData(1, 0, 400, 5.0, 1.5)
	color := render.Color{R: 0.55, G: 0.35, B: 0.75, A: 0.85}
	ew := 0.8
	bins := 18
	ax.Hist(data, core.HistOptions{Bins: bins, Cumulative: true, Color: &color, EdgeColor: &black, EdgeWidth: &ew})
}

// addMultiplePanel mirrors the hist_strategies fixture content.
func addMultiplePanel(ax *core.Axes) {
	ax.SetTitle("Multiple (Probability)")
	ax.SetXLim(0, 11)
	ax.SetYLim(0, 0.25)
	data1 := common.NormalData(42, 0, 300, 4.0, 1.0)
	data2 := common.NormalData(7, 0, 300, 7.0, 1.2)
	blue := render.Color{R: 0.26, G: 0.53, B: 0.80, A: 0.6}
	orange := render.Color{R: 0.90, G: 0.50, B: 0.10, A: 0.6}
	ew := 0.5
	bins := 15
	ax.Hist(data1, core.HistOptions{Bins: bins, Norm: core.HistNormProbability, Color: &blue, EdgeColor: &black, EdgeWidth: &ew})
	ax.Hist(data2, core.HistOptions{Bins: bins, Norm: core.HistNormProbability, Color: &orange, EdgeColor: &black, EdgeWidth: &ew})
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
