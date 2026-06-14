package locator_linear_labels

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	common "github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 720
	Height = 540
	DPI    = 100
)

// Plot builds a linear locator family parity fixture.
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	axes := []*core.Axes{
		fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.12, Y: 0.70}, Max: geom.Pt{X: 0.92, Y: 0.92}}),
		fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.12, Y: 0.40}, Max: geom.Pt{X: 0.92, Y: 0.62}}),
		fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.12, Y: 0.10}, Max: geom.Pt{X: 0.92, Y: 0.32}}),
	}
	titles := []string{"Default Linear Locator", "LinearLocator(5)", "MultipleLocator(1.5)"}
	x := []float64{0, 1.5, 3, 4.5, 6}
	y := []float64{0.20, 0.38, 0.64, 0.78, 0.90}
	color := render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1}
	width := 2.0

	for i, ax := range axes {
		ax.SetTitle(titles[i])
		ax.SetXLabel("x")
		ax.SetYLabel("y")
		common.AddReferenceYGrid(ax)
		ax.Plot(x, y, core.PlotOptions{Color: &color, LineWidth: &width})
		ax.SetXLim(0, 6)
		ax.SetYLim(0, 1)
		ax.YAxis.Locator = core.FixedLocator{TicksList: []float64{0, 0.5, 1.0}}
	}
	axes[1].XAxis.Locator = core.LinearLocator{NumTicks: 5}
	axes[2].XAxis.Locator = core.MultipleLocator{Base: 1.5}
	return fig
}

// Render is the AGG-rendered fixture image.
func Render() image.Image {
	return common.RenderFixtureFigure(Plot(), Width, Height)
}
