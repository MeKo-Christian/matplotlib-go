package spine_positions

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
	Height = 400
	DPI    = 100
)

// Plot builds an axes-fraction/outward spine-position parity fixture.
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)

	centered := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.08, Y: 0.16},
		Max: geom.Pt{X: 0.47, Y: 0.84},
	})
	configureAxes(centered, "axes = 0.5")
	centered.XAxis.SetSpinePositionAxes(0.5)
	centered.YAxis.SetSpinePositionAxes(0.5)

	outward := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.58, Y: 0.16},
		Max: geom.Pt{X: 0.97, Y: 0.84},
	})
	configureAxes(outward, "outward = 10 pt")
	outward.XAxis.SetSpinePositionOutward(10)
	outward.YAxis.SetSpinePositionOutward(10)

	return fig
}

func configureAxes(ax *core.Axes, title string) {
	ax.ShowFrame = false
	ax.SetTitle(title)
	ax.SetXLabel("x")
	ax.SetYLabel("y")
	ax.SetXLim(-2, 2)
	ax.SetYLim(-2, 2)
	ticks := []float64{-2, -1, 0, 1, 2}
	ax.XAxis.Locator = ticker.FixedLocator{TicksList: ticks}
	ax.YAxis.Locator = ticker.FixedLocator{TicksList: ticks}

	color := render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1}
	width := 2.0
	_, _ = ax.Plot(
		[]float64{-2, -1, 0, 1, 2},
		[]float64{-1.4, -0.4, 0.1, 0.9, 1.6},
		core.PlotOptions{Color: &color, LineWidth: &width},
	)
}

// Render is the AGG-rendered fixture image.
func Render() image.Image {
	return common.RenderFixtureFigure(Plot(), Width, Height)
}
