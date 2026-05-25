package scale_symlog_ticks

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
	Height = 400
	DPI    = 100
)

// Plot builds a symlog scale-default parity fixture.
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.12, Y: 0.18}, Max: geom.Pt{X: 0.92, Y: 0.88}})
	ax.SetTitle("Symlog Scale Ticks")
	ax.SetXLabel("signed value")
	ax.SetYLabel("response")
	common.AddReferenceYGrid(ax)

	x := []float64{-1000, -100, -10, -1, 0, 1, 10, 100, 1000}
	y := []float64{0.14, 0.23, 0.34, 0.46, 0.52, 0.59, 0.69, 0.80, 0.88}
	color := render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1}
	width := 2.0
	ax.Plot(x, y, core.PlotOptions{Color: &color, LineWidth: &width})

	ax.SetXLim(-1000, 1000)
	ax.SetYLim(0, 1)
	_ = ax.SetXScale(
		"symlog",
		transform.WithScaleBase(10),
		transform.WithScaleLinThresh(1),
	)
	ax.YAxis.Locator = core.FixedLocator{TicksList: []float64{0, 0.5, 1.0}}
	return fig
}

// Render is the AGG-rendered fixture image.
func Render() image.Image {
	return common.RenderFixtureFigure(Plot(), Width, Height)
}
