package axes_secondary_y_twiny

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/ticker"
)

const (
	Width  = 760
	Height = 400
	DPI    = 100
)

// Plot builds focused parity coverage for TwinY and SecondaryYAxis.
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	blue := render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1}
	orange := render.Color{R: 1.0, G: 0.50, B: 0.05, A: 1}
	lineWidth := 2.0

	host := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.09, Y: 0.17},
		Max: geom.Pt{X: 0.45, Y: 0.82},
	})
	host.SetTitle("TwinY")
	host.SetXLabel("bottom x")
	host.SetYLabel("shared y")
	host.Plot(
		[]float64{0, 1, 2, 3, 4},
		[]float64{0, 1, 2, 3, 4},
		core.PlotOptions{Color: &blue, LineWidth: &lineWidth},
	)
	host.SetXLim(0, 4)
	host.SetYLim(0, 4)
	host.XAxis.Locator = ticker.FixedLocator{TicksList: []float64{0, 2, 4}}
	host.YAxis.Locator = ticker.FixedLocator{TicksList: []float64{0, 2, 4}}
	common.AddReferenceYGrid(host)

	twin := host.TwinY()
	if twin == nil {
		panic("TwinY returned nil")
	}
	twin.Plot(
		[]float64{10, 20, 30, 40, 50},
		[]float64{4, 3, 2, 1, 0},
		core.PlotOptions{Color: &orange, LineWidth: &lineWidth},
	)
	twin.SetXLim(10, 50)
	twin.XAxisTop.Locator = ticker.FixedLocator{TicksList: []float64{10, 30, 50}}
	twin.SetXLabel("top x")
	if err := twin.SetXLabelPosition("top"); err != nil {
		panic(err)
	}

	primary := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.57, Y: 0.17},
		Max: geom.Pt{X: 0.88, Y: 0.82},
	})
	primary.SetTitle("SecondaryYAxis")
	primary.SetXLabel("sample")
	primary.SetYLabel("Celsius")
	primary.Plot(
		[]float64{0, 1, 2, 3, 4},
		[]float64{0, 20, 40, 60, 100},
		core.PlotOptions{Color: &blue, LineWidth: &lineWidth},
	)
	primary.SetXLim(0, 4)
	primary.SetYLim(0, 100)
	primary.XAxis.Locator = ticker.FixedLocator{TicksList: []float64{0, 2, 4}}
	primary.YAxis.Locator = ticker.FixedLocator{TicksList: []float64{0, 50, 100}}
	common.AddReferenceYGrid(primary)

	secondary, err := primary.SecondaryYAxis(
		core.AxisRight,
		func(celsius float64) float64 { return celsius*1.8 + 32 },
		func(fahrenheit float64) (float64, bool) { return (fahrenheit - 32) / 1.8, true },
	)
	if err != nil {
		panic(err)
	}
	secondary.YAxisRight.Locator = ticker.FixedLocator{TicksList: []float64{32, 122, 212}}
	secondary.SetYLabel("Fahrenheit")
	if err := secondary.SetYLabelPosition("right"); err != nil {
		panic(err)
	}

	return fig
}

// Render is the AGG-rendered fixture image.
func Render() image.Image {
	return common.RenderFixtureFigure(Plot(), Width, Height)
}
