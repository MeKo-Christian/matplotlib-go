package radar_basic

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/ticker"
	"github.com/cwbudde/matplotlib-go/transform"
)

const (
	Width  = 720
	Height = 720
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(720, 720)
	labels := []string{"Speed", "Power", "Range", "Handling", "Comfort"}
	ax, err := fig.AddRadarAxes(geom.Rect{Min: geom.Pt{X: 0.12, Y: 0.10}, Max: geom.Pt{X: 0.88, Y: 0.88}}, labels)
	if err != nil {
		panic(err)
	}
	if err := ax.SetThetaZeroLocation("N"); err != nil {
		panic(err)
	}
	ax.SetTitle("Radar Projection")
	ax.YScale = transform.NewLinear(0, 1)
	ax.YAxis.Locator = ticker.FixedLocator{TicksList: []float64{0.25, 0.5, 0.75, 1.0}}
	ax.YAxis.MinorLocator = nil
	ax.YAxis.Formatter = ticker.PercentFormatter{XMax: 1, Decimals: 0, DecimalsSet: true}

	thetaGrid := ax.AddGrid(core.AxisBottom)
	thetaGrid.Color = render.Color{R: 0.78, G: 0.80, B: 0.84, A: 1}
	thetaGrid.LineWidth = 0.8
	radiusGrid := ax.AddGrid(core.AxisLeft)
	radiusGrid.Color = render.Color{R: 0.80, G: 0.83, B: 0.88, A: 1}
	radiusGrid.LineWidth = 0.8

	angles := core.RadarAngles(len(labels))
	values := []float64{0.72, 0.88, 0.64, 0.79, 0.58}
	closedAngles := append(append([]float64(nil), angles...), angles[0])
	closedValues := append(append([]float64(nil), values...), values[0])
	lineColor := render.Color{R: 0.15, G: 0.35, B: 0.70, A: 1}
	fillColor := render.Color{R: 0.18, G: 0.50, B: 0.82, A: 0.22}
	lineWidth := 2.2
	ax.Fill(closedAngles, closedValues, core.FillOptions{Color: &fillColor})
	_, _ = ax.Plot(closedAngles, closedValues, core.PlotOptions{Color: &lineColor, LineWidth: &lineWidth, Label: "model A"})
	return fig
}

// Render is the AGG-rendered showcase image.
func Render() image.Image {
	fig := Plot()
	return common.RenderFixtureFigure(fig, Width, Height)
}
