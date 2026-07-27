// Package scatter_gallery is a user-facing showcase that gathers the advanced
// scatter surface into one figure: colormapped scalar mapping, variable marker
// size, alpha blending of overlapping markers, and multiple marker families. It
// closes the "advanced-scatter" demo-breadth gap by promoting
// behavior that previously only existed as parity fixtures (scatter_advanced,
// scatter_marker_types) into a single browsable gallery. The large
// path-collection case stays represented by the dedicated large_scatter
// fixture.
package scatter_gallery

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

const (
	Width  = 840
	Height = 620
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)

	addColormappedPanel(fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.07, Y: 0.585}, Max: geom.Pt{X: 0.46, Y: 0.93}}))
	addVariableSizePanel(fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.56, Y: 0.585}, Max: geom.Pt{X: 0.96, Y: 0.93}}))
	addAlphaPanel(fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.07, Y: 0.10}, Max: geom.Pt{X: 0.46, Y: 0.445}}))
	addMarkerFamiliesPanel(fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.56, Y: 0.10}, Max: geom.Pt{X: 0.96, Y: 0.445}}))

	return fig
}

func radii(rs ...float64) []float64 {
	out := make([]float64, len(rs))
	for i, r := range rs {
		out[i] = core.ScatterAreaFromRadius(r, style.Default.DPI)
	}
	return out
}

// addColormappedPanel maps per-point scalar values through a colormap.
func addColormappedPanel(ax *core.Axes) {
	ax.SetTitle("Colormapped")
	ax.SetXLim(0, 6)
	ax.SetYLim(0, 5)
	edge := render.Color{R: 0.08, G: 0.08, B: 0.08, A: 1}
	edgeWidth := 1.5
	ax.Scatter(
		[]float64{0.8, 1.8, 2.8, 3.8, 4.8},
		[]float64{1.2, 3.2, 2.2, 3.8, 1.8},
		core.ScatterOptions{
			ScalarValues: []float64{-1.0, -0.2, 0.35, 0.8, 1.2},
			Colormap:     "viridis",
			Sizes:        radii(7, 10, 13, 16, 19),
			EdgeColor:    optional.Of(edge),
			EdgeWidth:    optional.Of(edgeWidth),
		},
	)
}

// addVariableSizePanel shows a single-color series with a per-point size array.
func addVariableSizePanel(ax *core.Axes) {
	ax.SetTitle("Variable Size")
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 9)
	col := render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1}
	edge := render.Color{R: 0.08, G: 0.08, B: 0.08, A: 1}
	edgeWidth := 1.0
	ax.Scatter(
		[]float64{1.5, 3.0, 4.5, 6.0, 7.5, 9.0},
		[]float64{5.0, 3.0, 6.0, 4.0, 7.0, 5.0},
		core.ScatterOptions{
			Color:     optional.Of(col),
			Sizes:     radii(5, 9, 13, 17, 21, 25),
			EdgeColor: optional.Of(edge),
			EdgeWidth: optional.Of(edgeWidth),
		},
	)
}

// addAlphaPanel overlaps two translucent clusters to show alpha blending.
func addAlphaPanel(ax *core.Axes) {
	ax.SetTitle("Alpha Blending")
	ax.SetXLim(0, 8)
	ax.SetYLim(0, 8)
	size := core.ScatterAreaFromRadius(28.0, style.Default.DPI)
	alpha := 0.45
	edgeWidth := 0.0

	red := render.Color{R: 0.85, G: 0.20, B: 0.20, A: 1}
	ax.Scatter(
		[]float64{2.5, 3.5, 3.0, 3.0},
		[]float64{4.0, 4.0, 4.8, 3.2},
		core.ScatterOptions{Color: optional.Of(red), Size: optional.Of(size), Alpha: optional.Of(alpha), EdgeWidth: optional.Of(edgeWidth)},
	)
	blue := render.Color{R: 0.20, G: 0.30, B: 0.85, A: 1}
	ax.Scatter(
		[]float64{4.5, 5.5, 5.0, 5.0},
		[]float64{4.0, 4.0, 4.8, 3.2},
		core.ScatterOptions{Color: optional.Of(blue), Size: optional.Of(size), Alpha: optional.Of(alpha), EdgeWidth: optional.Of(edgeWidth)},
	)
}

// addMarkerFamiliesPanel lays out several built-in marker families.
func addMarkerFamiliesPanel(ax *core.Axes) {
	ax.SetTitle("Marker Families")
	ax.SetXLim(0.5, 3.5)
	ax.SetYLim(0.5, 2.5)

	markers := []core.MarkerType{
		core.MarkerCircle, core.MarkerSquare, core.MarkerTriangleUp,
		core.MarkerDiamond, core.MarkerPentagon, core.MarkerStar,
	}
	palette := style.Default.Palette()
	edge := render.Color{R: 0.08, G: 0.08, B: 0.08, A: 1}
	size := core.ScatterAreaFromRadius(14.0, style.Default.DPI)
	edgeWidth := 1.2
	for i, m := range markers {
		marker := m
		x := float64(i%3) + 1
		y := float64(2 - i/3)
		col := palette[i%len(palette)]
		ax.Scatter(
			[]float64{x},
			[]float64{y},
			core.ScatterOptions{
				Color:     optional.Of(col),
				Size:      optional.Of(size),
				Marker:    optional.Of(marker),
				EdgeColor: optional.Of(edge),
				EdgeWidth: optional.Of(edgeWidth),
			},
		)
	}
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
