package locator_log_minor_threshold_labels

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	common "github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/transform"
)

const (
	Width  = 720
	Height = 420
	DPI    = 100
)

// Plot builds a log locator parity fixture with visible minor-grid behavior.
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	top := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.12, Y: 0.58}, Max: geom.Pt{X: 0.92, Y: 0.86}})
	bottom := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.12, Y: 0.16}, Max: geom.Pt{X: 0.92, Y: 0.44}})
	color := render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1}
	width := 2.0

	top.SetTitle("LogLocator Auto Minor Grid")
	top.SetXLabel("base 10")
	top.SetYLabel("value")
	top.Plot(
		[]float64{1, 10, 100, 1000},
		[]float64{0.18, 0.39, 0.67, 0.88},
		core.PlotOptions{Color: &color, LineWidth: &width},
	)
	top.SetYLim(0, 1)
	_ = top.SetXScale("log", transform.WithScaleBase(10))
	top.SetXLim(1, 1000)
	top.XAxis.MinorLocator = core.LogLocator{Base: 10, SubsMode: "auto"}
	top.YAxis.Locator = core.FixedLocator{TicksList: []float64{0, 0.5, 1.0}}
	xGrid := top.AddXGrid()
	xGrid.Minor = true
	xGrid.MinorLineWidth = 0.35
	xGrid.MinorColor = render.Color{R: 0.72, G: 0.76, B: 0.80, A: 0.55}
	xGrid.MinorDashes = []float64{1, 2}
	yGrid := top.AddYGrid()
	yGrid.Color = render.Color{R: 0.80, G: 0.80, B: 0.80, A: 1}
	yGrid.LineWidth = 0.5

	bottom.SetTitle("LogLocator Base 2")
	bottom.SetXLabel("base 2")
	bottom.SetYLabel("value")
	bottom.Plot(
		[]float64{1, 2, 4, 8, 16, 32, 64},
		[]float64{0.14, 0.26, 0.38, 0.52, 0.68, 0.80, 0.91},
		core.PlotOptions{Color: &color, LineWidth: &width},
	)
	bottom.SetYLim(0, 1)
	_ = bottom.SetXScale("log", transform.WithScaleBase(2))
	bottom.SetXLim(1, 64)
	bottom.YAxis.Locator = core.FixedLocator{TicksList: []float64{0, 0.5, 1.0}}
	common.AddReferenceYGrid(bottom)
	return fig
}

// Render is the AGG-rendered fixture image.
func Render() image.Image {
	return common.RenderFixtureFigure(Plot(), Width, Height)
}
