package formatter_fixed_null_labels

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/ticker"
)

const (
	Width  = 640
	Height = 360
	DPI    = 100
)

// Plot builds a formatter-focused fixed/null label parity fixture.
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.12, Y: 0.18}, Max: geom.Pt{X: 0.92, Y: 0.88}})
	ax.SetTitle("Fixed + Null Formatters")
	ax.SetXLabel("stage")
	ax.SetYLabel("hidden values")
	common.AddReferenceYGrid(ax)

	x := []float64{0, 1, 2, 3}
	y := []float64{0.20, 0.62, 0.44, 0.82}
	color := render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1}
	width := 2.0
	_, _ = ax.Plot(x, y, core.PlotOptions{Color: &color, LineWidth: &width})

	ax.SetXLim(-0.25, 3.25)
	ax.SetYLim(0, 1)
	ax.XAxis.Locator = ticker.FixedLocator{TicksList: x}
	ax.XAxis.Formatter = ticker.FixedFormatter{Labels: []string{"draft", "review", "ship", ""}}
	ax.YAxis.Locator = ticker.FixedLocator{TicksList: []float64{0, 0.5, 1.0}}
	ax.YAxis.Formatter = ticker.NullFormatter{}
	return fig
}

// Render is the AGG-rendered fixture image.
func Render() image.Image {
	return common.RenderFixtureFigure(Plot(), Width, Height)
}
