package formatter_percent_labels

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	common "github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 640
	Height = 360
	DPI    = 100
)

// Plot builds a formatter-focused percent-label parity fixture.
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.12, Y: 0.18}, Max: geom.Pt{X: 0.92, Y: 0.88}})
	ax.SetTitle("Percent Formatter")
	ax.SetXLabel("share")
	ax.SetYLabel("score")
	common.AddReferenceYGrid(ax)

	x := []float64{-0.5, 0, 0.125, 0.5, 1.0}
	y := []float64{0.15, 0.30, 0.42, 0.68, 0.86}
	color := render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1}
	width := 2.0
	ax.Plot(x, y, core.PlotOptions{Color: &color, LineWidth: &width})

	ax.SetXLim(-0.5, 1.0)
	ax.SetYLim(0, 1.0)
	ax.XAxis.Locator = core.FixedLocator{TicksList: x}
	ax.XAxis.Formatter = core.PercentFormatter{XMax: 1, DisplayRange: 1.5}
	ax.YAxis.Locator = core.FixedLocator{TicksList: []float64{0, 0.5, 1.0}}
	return fig
}

// Render is the AGG-rendered fixture image.
func Render() image.Image {
	return common.RenderFixtureFigure(Plot(), Width, Height)
}
