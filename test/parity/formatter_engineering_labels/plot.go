package formatter_engineering_labels

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	common "github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 720
	Height = 400
	DPI    = 100
)

// Plot builds a formatter-focused engineering-label parity fixture.
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	addPanel(
		fig, geom.Rect{Min: geom.Pt{X: 0.12, Y: 0.58}, Max: geom.Pt{X: 0.94, Y: 0.88}},
		"Micro Engineering Labels",
		[]float64{-2e-6, -1e-6, 0, 1e-6, 2e-6},
		core.EngFormatter{Unit: "V"},
	)
	addPanel(
		fig, geom.Rect{Min: geom.Pt{X: 0.12, Y: 0.16}, Max: geom.Pt{X: 0.94, Y: 0.46}},
		"Kilohertz Engineering Labels",
		[]float64{0, 1000, 1500, 2000},
		core.EngFormatter{Unit: "Hz", Places: 1},
	)
	return fig
}

func addPanel(fig *core.Figure, rect geom.Rect, title string, ticks []float64, formatter core.Formatter) {
	ax := fig.AddAxes(rect)
	ax.SetTitle(title)
	ax.SetXLabel("value")
	ax.SetYLabel("score")
	common.AddReferenceYGrid(ax)

	y := make([]float64, len(ticks))
	for i := range y {
		y[i] = 0.2 + 0.15*float64(i)
	}
	color := render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1}
	width := 2.0
	ax.Plot(ticks, y, core.PlotOptions{Color: &color, LineWidth: &width})
	ax.SetXLim(ticks[0], ticks[len(ticks)-1])
	ax.SetYLim(0, 1)
	ax.XAxis.Locator = core.FixedLocator{TicksList: ticks}
	ax.XAxis.Formatter = formatter
	ax.YAxis.Locator = core.FixedLocator{TicksList: []float64{0, 0.5, 1.0}}
}

// Render is the AGG-rendered fixture image.
func Render() image.Image {
	return common.RenderFixtureFigure(Plot(), Width, Height)
}
