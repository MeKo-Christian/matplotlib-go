package geo_lambert_axes

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/ticker"
)

const (
	Width  = 520
	Height = 520
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(520, 520)
	ax, err := fig.AddAxesProjection(geom.Rect{
		Min: geom.Pt{X: 0.08, Y: 0.10},
		Max: geom.Pt{X: 0.92, Y: 0.88},
	}, "lambert")
	if err != nil {
		panic(err)
	}
	ax.SetTitle("Lambert Projection")
	ax.SetXLabel("longitude")
	ax.SetYLabel("latitude")
	ax.XAxis.Locator = ticker.FixedLocator{TicksList: common.LambertLongitudeTicks()}
	ax.XAxis.Formatter = ticker.FuncFormatter(common.PlainDegreeFormat)

	gridColor := render.Color{R: 0.78, G: 0.80, B: 0.84, A: 1}
	lonGrid := ax.AddGrid(core.AxisBottom)
	lonGrid.Color = gridColor
	lonGrid.LineWidth = 0.8
	latGrid := ax.AddGrid(core.AxisLeft)
	latGrid.Color = gridColor
	latGrid.LineWidth = 0.8

	const n = 361
	lon := make([]float64, n)
	lat := make([]float64, n)
	for i := range n {
		t := float64(i) / float64(n-1)
		lon[i] = -math.Pi/2 + math.Pi*t
		lat[i] = 0.35 * math.Sin(3*lon[i])
	}

	lineColor := render.Color{R: 0.14, G: 0.34, B: 0.70, A: 1}
	lineWidth := 2.0
	ax.Plot(lon, lat, core.PlotOptions{Color: &lineColor, LineWidth: &lineWidth})
	return fig
}

// Render is the AGG-rendered showcase image.
func Render() image.Image {
	fig := Plot()
	return common.RenderFixtureFigure(fig, Width, Height)
}
