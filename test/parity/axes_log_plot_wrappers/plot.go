package axes_log_plot_wrappers

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/ticker"
)

const (
	Width  = 960
	Height = 360
	DPI    = 100
)

// Plot builds focused parity coverage for the logarithmic plot wrappers.
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	color := render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1}
	lineWidth := 2.0
	opts := core.PlotOptions{Color: &color, LineWidth: &lineWidth}

	semilogX := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.07, Y: 0.18},
		Max: geom.Pt{X: 0.31, Y: 0.86},
	})
	semilogX.SetTitle("SemilogX")
	semilogX.SetXLabel("log x")
	semilogX.SetYLabel("linear y")
	semilogX.SemilogX(
		[]float64{1, 3, 10, 30, 100, 300, 1000},
		[]float64{0.5, 1.0, 1.8, 2.6, 3.2, 4.1, 4.6},
		opts,
	)
	semilogX.SetXLim(1, 1000)
	semilogX.SetYLim(0, 5)
	semilogX.YAxis.Locator = ticker.FixedLocator{TicksList: []float64{0, 2.5, 5}}
	common.AddReferenceXYGrid(semilogX)

	semilogY := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.39, Y: 0.18},
		Max: geom.Pt{X: 0.63, Y: 0.86},
	})
	semilogY.SetTitle("SemilogY")
	semilogY.SetXLabel("linear x")
	semilogY.SetYLabel("log y")
	semilogY.SemilogY(
		[]float64{0, 1, 2, 3, 4, 5, 6},
		[]float64{1, 3, 10, 30, 100, 300, 1000},
		opts,
	)
	semilogY.SetXLim(0, 6)
	semilogY.SetYLim(1, 1000)
	semilogY.XAxis.Locator = ticker.FixedLocator{TicksList: []float64{0, 3, 6}}
	common.AddReferenceXYGrid(semilogY)

	logLog := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.71, Y: 0.18},
		Max: geom.Pt{X: 0.95, Y: 0.86},
	})
	logLog.SetTitle("LogLog")
	logLog.SetXLabel("log x")
	logLog.SetYLabel("log y")
	logLog.LogLog(
		[]float64{1, 3, 10, 30, 100, 300, 1000},
		[]float64{1, 2, 7, 18, 70, 240, 900},
		opts,
	)
	logLog.SetXLim(1, 1000)
	logLog.SetYLim(1, 1000)
	common.AddReferenceXYGrid(logLog)

	return fig
}

// Render is the AGG-rendered fixture image.
func Render() image.Image {
	return common.RenderFixtureFigure(Plot(), Width, Height)
}
