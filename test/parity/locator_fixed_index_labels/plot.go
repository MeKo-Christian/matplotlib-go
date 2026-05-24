package locator_fixed_index_labels

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	common "github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 720
	Height = 420
	DPI    = 100
)

// Plot builds a fixed/index locator parity fixture.
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	top := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.12, Y: 0.58}, Max: geom.Pt{X: 0.92, Y: 0.86}})
	bottom := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.12, Y: 0.16}, Max: geom.Pt{X: 0.92, Y: 0.44}})
	color := render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1}
	width := 2.0

	top.SetTitle("FixedLocator Subsampling")
	top.SetXLabel("fixed positions")
	top.SetYLabel("value")
	common.AddReferenceYGrid(top)
	top.Plot(
		[]float64{-6, -4, -2, 0, 2, 4, 6},
		[]float64{0.18, 0.34, 0.47, 0.62, 0.74, 0.86, 0.92},
		core.PlotOptions{Color: &color, LineWidth: &width},
	)
	top.SetXLim(-6, 6)
	top.SetYLim(0, 1)
	top.XAxis.Locator = core.FixedLocator{TicksList: []float64{-6, -4, -2, 0, 2, 4, 6}, Nbins: 4}
	top.YAxis.Locator = core.FixedLocator{TicksList: []float64{0, 0.5, 1.0}}

	bottom.SetTitle("IndexLocator Base + Offset")
	bottom.SetXLabel("index")
	bottom.SetYLabel("value")
	common.AddReferenceYGrid(bottom)
	bottom.Plot(
		[]float64{0, 1, 2, 3, 4, 5, 6, 7, 8},
		[]float64{0.15, 0.28, 0.36, 0.51, 0.63, 0.70, 0.82, 0.88, 0.93},
		core.PlotOptions{Color: &color, LineWidth: &width},
	)
	bottom.SetXLim(0, 8)
	bottom.SetYLim(0, 1)
	bottom.XAxis.Locator = core.IndexLocator{Base: 2, Offset: 1}
	bottom.YAxis.Locator = core.FixedLocator{TicksList: []float64{0, 0.5, 1.0}}
	return fig
}

// Render is the AGG-rendered fixture image.
func Render() image.Image {
	return common.RenderFixtureFigure(Plot(), Width, Height)
}
