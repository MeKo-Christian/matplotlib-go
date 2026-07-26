// Package lines_markers_gallery is a user-facing showcase that gathers the
// Line2D stroke and marker-styling surface into one figure: dash arrays, line
// joins and caps, a built-in marker grid (including open-fill markers), and a
// multi-series legend. It closes the Phase 18.1 "marker-grid" demo-breadth gap
// by promoting behavior that previously only existed as parity fixtures
// (dashes, joins_caps, scatter_marker_types) into a single browsable gallery.
package lines_markers_gallery

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/color"
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

	addDashPanel(fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.07, Y: 0.585}, Max: geom.Pt{X: 0.46, Y: 0.93}}))
	addJoinsCapsPanel(fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.56, Y: 0.585}, Max: geom.Pt{X: 0.96, Y: 0.93}}))
	addMarkerGridPanel(fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.07, Y: 0.10}, Max: geom.Pt{X: 0.46, Y: 0.445}}))
	addLegendPanel(fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.56, Y: 0.10}, Max: geom.Pt{X: 0.96, Y: 0.445}}))

	return fig
}

// addDashPanel mirrors the dashes fixture: four parallel lines with different
// dash arrays.
func addDashPanel(ax *core.Axes) {
	ax.SetTitle("Dash Patterns")
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 5)

	specs := []struct {
		y       float64
		pattern []float64
		color   render.Color
	}{
		{4, nil, render.Color{R: 0, G: 0, B: 0, A: 1}},
		{3, []float64{10, 4}, render.Color{R: 0.8, G: 0, B: 0, A: 1}},
		{2, []float64{6, 2, 2, 2}, render.Color{R: 0, G: 0.6, B: 0, A: 1}},
		{1, []float64{2, 2}, render.Color{R: 0, G: 0, B: 0.8, A: 1}},
	}
	for _, spec := range specs {
		line := &core.Line2D{
			XY:  []geom.Pt{{X: 1, Y: spec.y}, {X: 9, Y: spec.y}},
			W:   3.0,
			Col: spec.color,
		}
		if len(spec.pattern) > 0 {
			line.SetDashes(spec.pattern...)
		}
		ax.Add(line)
	}
}

// addJoinsCapsPanel mirrors the joins_caps fixture: a thick polyline showing
// line joins and a thick segment showing line caps.
func addJoinsCapsPanel(ax *core.Axes) {
	ax.SetTitle("Line Joins and Caps")
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 6)

	ax.Add(&core.Line2D{
		XY:  []geom.Pt{{X: 1, Y: 5}, {X: 3, Y: 5}, {X: 3, Y: 3}, {X: 5, Y: 3}},
		W:   8.0,
		Col: render.Color{R: 0.8, G: 0.2, B: 0.2, A: 1},
	})
	ax.Add(&core.Line2D{
		XY:  []geom.Pt{{X: 7, Y: 5}, {X: 9, Y: 5}},
		W:   8.0,
		Col: render.Color{R: 0.2, G: 0.2, B: 0.8, A: 1},
	})
}

// addMarkerGridPanel lays out a curated grid of built-in markers, including two
// open-fill markers, to show marker shapes plus marker edge/fill styling.
func addMarkerGridPanel(ax *core.Axes) {
	ax.SetTitle("Marker Grid + Fill Styles")
	ax.SetXLim(0.5, 6.5)
	ax.SetYLim(0.5, 2.5)

	type cell struct {
		style core.MarkerStyle
		open  bool
	}
	cells := []cell{
		{core.NewMarkerStyle(core.MarkerCircle), false},
		{core.NewMarkerStyle(core.MarkerSquare), false},
		{core.NewMarkerStyle(core.MarkerTriangleUp), false},
		{core.NewMarkerStyle(core.MarkerDiamond), false},
		{core.NewMarkerStyle(core.MarkerPentagon), false},
		{core.NewMarkerStyle(core.MarkerStar), false},
		{core.NewMarkerStyle(core.MarkerHexagon1), false},
		{core.NewMarkerStyle(core.MarkerOctagon), false},
		{core.NewMarkerStyle(core.MarkerFilledPlus), false},
		{core.NewMarkerStyle(core.MarkerFilledX), false},
		{core.NewMarkerStyle(core.MarkerThinDiamond), false},
		{core.MarkerStyle{Type: core.MarkerCircle, FillStyle: core.MarkerFillNone}, true},
	}

	palette := style.Default.Palette()
	edge := render.Color{R: 0.05, G: 0.05, B: 0.05, A: 1}
	for i, c := range cells {
		x := float64(i%6) + 1
		y := float64(2 - i/6)
		col := palette[i%len(palette)]
		markerEdge := edge
		if c.open {
			markerEdge = col
		}
		ax.Add(&core.Scatter2D{
			XY:          []geom.Pt{{X: x, Y: y}},
			Size:        core.ScatterAreaFromRadius(9.0, style.Default.DPI),
			Color:       col,
			EdgeColor:   markerEdge,
			EdgeWidth:   1.2,
			MarkerStyle: c.style,
			Alpha:       1.0,
		})
	}
}

// addLegendPanel draws several labeled series sharing one axes and a legend,
// demonstrating color cycling and multi-series legend behavior.
func addLegendPanel(ax *core.Axes) {
	ax.SetTitle("Multi-Series Legend")
	ax.SetXLim(0, 6)
	ax.SetYLim(0, 6)

	tab10 := color.Tab10
	lineWidth := 2.0
	x := []float64{0.5, 1.5, 2.5, 3.5, 4.5, 5.5}
	series := []struct {
		label string
		col   render.Color
		y     []float64
	}{
		{"rising", tab10[0], []float64{0.8, 1.6, 2.5, 3.3, 4.2, 5.0}},
		{"falling", tab10[1], []float64{5.2, 4.4, 3.7, 2.9, 2.1, 1.3}},
		{"wave", tab10[2], []float64{2.6, 3.6, 2.8, 3.8, 3.0, 4.0}},
	}
	for _, s := range series {
		_, _ = ax.Plot(x, s.y, core.PlotOptions{
			Color:     optional.Of(s.col),
			LineWidth: optional.Of(lineWidth),
			Label:     s.label,
		})
	}
	legend := ax.AddLegend()
	legend.Location = core.LegendUpperRight
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
