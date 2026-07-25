package scale_function_defaults

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/ticker"
	"github.com/cwbudde/matplotlib-go/transform"
)

const (
	Width  = 720
	Height = 480
	DPI    = 100
)

// Plot builds function and functionlog scale-default parity fixtures.
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	top := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.12, Y: 0.58}, Max: geom.Pt{X: 0.92, Y: 0.90}})
	bottom := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.12, Y: 0.15}, Max: geom.Pt{X: 0.92, Y: 0.47}})
	color := render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1}
	width := 2.0

	top.SetTitle("Function Scale Defaults")
	top.SetXLabel("sqrt-scaled x")
	top.SetYLabel("response")
	common.AddReferenceYGrid(top)
	x := []float64{0, 4, 16, 36, 64, 100}
	y := []float64{0.12, 0.25, 0.41, 0.58, 0.76, 0.90}
	_, _ = top.Plot(x, y, core.PlotOptions{Color: &color, LineWidth: &width})
	top.SetXLim(0, 100)
	top.SetYLim(0, 1)
	_ = top.SetXScale(
		"function",
		transform.WithScaleFunctions(
			func(v float64) float64 { return math.Sqrt(v) },
			func(v float64) (float64, bool) { return v * v, true },
		),
	)
	top.YAxis.Locator = ticker.FixedLocator{TicksList: []float64{0, 0.5, 1.0}}

	bottom.SetTitle("Functionlog Scale Defaults")
	bottom.SetXLabel("sqrt-scaled log x")
	bottom.SetYLabel("response")
	common.AddReferenceYGrid(bottom)
	x = []float64{1, 10, 100, 1000, 10000}
	y = []float64{0.12, 0.31, 0.52, 0.72, 0.90}
	_, _ = bottom.Plot(x, y, core.PlotOptions{Color: &color, LineWidth: &width})
	bottom.SetXLim(1, 10000)
	bottom.SetYLim(0, 1)
	_ = bottom.SetXScale(
		"functionlog",
		transform.WithScaleBase(10),
		transform.WithScaleFunctions(
			func(v float64) float64 { return math.Sqrt(v) },
			func(v float64) (float64, bool) { return v * v, true },
		),
	)
	bottom.YAxis.Locator = ticker.FixedLocator{TicksList: []float64{0, 0.5, 1.0}}

	return fig
}

// Render is the AGG-rendered fixture image.
func Render() image.Image {
	return common.RenderFixtureFigure(Plot(), Width, Height)
}
