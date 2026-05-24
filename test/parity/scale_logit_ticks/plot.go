package scale_logit_ticks

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	common "github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 720
	Height = 400
	DPI    = 100
)

// Plot builds a logit scale-default parity fixture.
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.12, Y: 0.18}, Max: geom.Pt{X: 0.92, Y: 0.88}})
	ax.SetTitle("Logit Scale Ticks")
	ax.SetXLabel("probability")
	ax.SetYLabel("score")
	common.AddReferenceYGrid(ax)

	x := []float64{0.001, 0.01, 0.05, 0.2, 0.5, 0.8, 0.95, 0.99, 0.999}
	y := []float64{0.12, 0.21, 0.34, 0.46, 0.55, 0.67, 0.78, 0.86, 0.91}
	color := render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1}
	width := 2.0
	ax.Plot(x, y, core.PlotOptions{Color: &color, LineWidth: &width})

	ax.SetXLim(0.001, 0.999)
	ax.SetYLim(0, 1)
	_ = ax.SetXScale("logit")
	ax.YAxis.Locator = core.FixedLocator{TicksList: []float64{0, 0.5, 1.0}}
	return fig
}

// Render is the AGG-rendered fixture image.
func Render() image.Image {
	return common.RenderFixtureFigure(Plot(), Width, Height)
}
