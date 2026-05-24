package formatter_scalar_scientific_labels

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	common "github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 640
	Height = 360
	DPI    = 100
)

// Plot builds a formatter-focused scalar scientific-label parity fixture.
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.14, Y: 0.18}, Max: geom.Pt{X: 0.92, Y: 0.88}})
	ax.SetTitle("Scalar Scientific Formatter")
	ax.SetXLabel("value")
	ax.SetYLabel("score")
	common.AddReferenceYGrid(ax)

	x := []float64{-1200, 0, 1200}
	y := []float64{0.28, 0.62, 0.78}
	color := render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1}
	width := 2.0
	ax.Plot(x, y, core.PlotOptions{Color: &color, LineWidth: &width})

	ax.SetXLim(-1400, 1400)
	ax.SetYLim(0, 1)
	ax.XAxis.Locator = core.FixedLocator{TicksList: x}
	ax.XAxis.Formatter = core.ScalarFormatter{
		Prec:           1,
		UsePowerLimits: true,
		PowerLimits:    [2]int{0, 0},
		UseMathText:    true,
	}
	ax.YAxis.Locator = core.FixedLocator{TicksList: []float64{0, 0.5, 1.0}}
	return fig
}

// Render is the AGG-rendered fixture image.
func Render() image.Image {
	return common.RenderFixtureFigure(Plot(), Width, Height)
}
