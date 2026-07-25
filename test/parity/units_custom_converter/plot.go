package units_custom_converter

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 680
	Height = 380
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	common.RegisterTestDistanceUnits()
	fig := core.NewFigure(680, 380)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.10, Y: 0.18}, Max: geom.Pt{X: 0.94, Y: 0.88}})
	ax.SetTitle("Custom Distance Units")
	ax.SetXLabel("Distance")
	ax.SetYLabel("Pace")
	common.AddReferenceXYGrid(ax)
	distances := []common.TestDistanceKM{5, 10, 21.1, 30, 42.2}
	pace := []float64{6.4, 5.9, 5.3, 5.1, 5.4}
	brown := render.Color{R: 0.55, G: 0.34, B: 0.29, A: 1}
	green := render.Color{R: 0.17, G: 0.63, B: 0.17, A: 0.92}
	edge := render.Color{R: 0.09, G: 0.36, B: 0.09, A: 1}
	lineWidth := 1.4
	markerSize := core.ScatterAreaFromRadius(8.0, 100.0)
	if _, err := ax.Plot(distances, pace, core.PlotOptions{Color: &brown, LineWidth: &lineWidth}); err != nil {
		panic(err)
	}
	if _, err := ax.ScatterUnits(distances, pace, core.ScatterOptions{Color: &green, EdgeColor: &edge, Size: &markerSize}); err != nil {
		panic(err)
	}
	ax.AutoScale(0.08)
	return fig
}

// Render is the AGG-rendered showcase image.
func Render() image.Image {
	fig := Plot()
	return common.RenderFixtureFigure(fig, Width, Height)
}
