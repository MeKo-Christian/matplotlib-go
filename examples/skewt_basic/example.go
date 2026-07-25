package skewt_basic

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/ticker"
)

const (
	Width  = 720
	Height = 640
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(720, 640)
	ax, err := fig.AddSkewXAxes(geom.Rect{Min: geom.Pt{X: 0.16, Y: 0.14}, Max: geom.Pt{X: 0.88, Y: 0.88}})
	if err != nil {
		panic(err)
	}
	ax.SetTitle("Skew-T Style Projection")
	ax.SetXLabel("Temperature (deg C)")
	ax.SetYLabel("Pressure (hPa)")
	if err := ax.SetYScale("log"); err != nil {
		panic(err)
	}
	ax.SetXLim(-70, 35)
	ax.SetYLim(1050, 180)
	ax.XAxis.Locator = ticker.MultipleLocator{Base: 10}
	ax.XAxis.MinorLocator = ticker.MultipleLocator{Base: 5}
	if ax.XAxisTop != nil {
		ax.XAxisTop.Locator = ax.XAxis.Locator
		ax.XAxisTop.MinorLocator = ax.XAxis.MinorLocator
	}
	ax.YAxis.Locator = ticker.FixedLocator{TicksList: []float64{100, 200, 300, 500, 700, 850, 1000}}
	ax.YAxis.MinorLocator = ticker.LogLocator{Base: 10, Minor: true, Subs: []float64{2, 3, 4, 5, 6, 7, 8, 9}}
	ax.YAxis.Formatter = ticker.ScalarFormatter{Prec: 0}
	ax.YAxis.MinorFormatter = ticker.NullFormatter{}
	gridColor := render.Color{R: 0.82, G: 0.84, B: 0.88, A: 1}
	xGrid := ax.AddGrid(core.AxisBottom)
	xGrid.Color = gridColor
	xGrid.LineWidth = 0.8
	yGrid := ax.AddGrid(core.AxisLeft)
	yGrid.Color = gridColor
	yGrid.LineWidth = 0.8
	pressure := []float64{1000, 925, 850, 700, 600, 500, 400, 300, 250, 200}
	temperature := []float64{24, 20, 15, 5, -4, -14, -28, -43, -51, -58}
	dewpoint := []float64{18, 14, 8, -4, -14, -25, -38, -50, -57, -64}
	tempColor := render.Color{R: 0.78, G: 0.13, B: 0.16, A: 1}
	dewColor := render.Color{R: 0.05, G: 0.48, B: 0.28, A: 1}
	width := 2.4
	ax.Plot(temperature, pressure, core.PlotOptions{Color: &tempColor, LineWidth: &width, Label: "temperature"})
	ax.Plot(dewpoint, pressure, core.PlotOptions{Color: &dewColor, LineWidth: &width, Label: "dewpoint"})
	ax.AddLegend()
	return fig
}

// Render is the AGG-rendered showcase image.
func Render() image.Image {
	fig := Plot()
	return common.RenderFixtureFigure(fig, Width, Height)
}
