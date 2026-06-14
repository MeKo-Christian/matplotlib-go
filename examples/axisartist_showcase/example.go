// Package axisartist_showcase is a showcase example. The body lived previously in
// examples/axisartist/showcase; this is its self-contained home.
package axisartist_showcase

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 980
	Height = 640
	DPI    = 100
)

// AxisArtistShowcase builds the same plot as
// test/matplotlib_ref/plots/axisartist_showcase.py.
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)

	host := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.08, Y: 0.14},
		Max: geom.Pt{X: 0.56, Y: 0.88},
	})
	host.SetTitle("AxisArtist / Parasite")
	host.SetXLabel("phase")
	host.SetYLabel("signal")
	host.SetXLim(-3.5, 3.5)
	host.SetYLim(-1.3, 1.3)
	grid := host.AddYGrid()
	grid.Color = render.Color{R: 0.78, G: 0.80, B: 0.84, A: 1}
	grid.LineWidth = 0.8

	x := make([]float64, 240)
	sine := make([]float64, len(x))
	cosScaled := make([]float64, len(x))
	for i := range x {
		x[i] = -3.5 + 7*float64(i)/float64(len(x)-1)
		sine[i] = math.Sin(x[i])
		cosScaled[i] = 55 + 35*math.Cos(x[i]*0.8)
	}

	hostColor := render.Color{R: 0.14, G: 0.34, B: 0.72, A: 1}
	hostWidth := 2.2
	host.Plot(x, sine, core.PlotOptions{
		Color:     &hostColor,
		LineWidth: &hostWidth,
		Label:     "sin(x)",
	})
	referenceColor := render.Color{R: 0.26, G: 0.26, B: 0.30, A: 1}
	referenceWidth := 1.4
	host.AxHLine(0, core.HLineOptions{
		Color:     &referenceColor,
		LineWidth: &referenceWidth,
		Dashes:    []float64{5 * 36.0 / DPI, 3 * 36.0 / DPI},
	})
	host.AxVLine(0, core.VLineOptions{
		Color:     &referenceColor,
		LineWidth: &referenceWidth,
		Dashes:    []float64{5 * 36.0 / DPI, 3 * 36.0 / DPI},
	})
	tickDirection := "inout"
	_ = host.TickParams(core.TickParams{Direction: &tickDirection})

	overlay := host.TwinX()
	if overlay != nil {
		overlay.SetYLim(0, 100)

		right := overlay.RightAxis()
		right.Color = render.Color{R: 0.74, G: 0.28, B: 0.18, A: 1}
		right.SetLineStyle(render.CapRound, render.JoinRound)

		overlayColor := render.Color{R: 0.74, G: 0.28, B: 0.18, A: 1}
		overlayWidth := 1.8
		overlay.Plot(x, cosScaled, core.PlotOptions{
			Color:     &overlayColor,
			LineWidth: &overlayWidth,
			Label:     "55 + 35 cos(0.8x)",
		})
	}

	host.Text(0.02, 0.98, "floating axes at x=0 / y=0\nparasite right scale", core.TextOptions{
		Coords:   core.Coords(core.CoordAxes),
		VAlign:   core.TextVAlignTop,
		FontSize: 10,
		BBox: &core.TextBBoxOptions{
			FaceColor: render.Color{R: 1, G: 1, B: 1, A: 1},
			EdgeColor: render.Color{R: 0.75, G: 0.75, B: 0.75, A: 1},
			Padding:   0.3,
		},
	})
	legend := host.AddLegend()
	legend.Location = core.LegendUpperCenter

	return fig
}

// Render is the AGG-rendered showcase image.
func Render() image.Image {
	fig := Plot()
	r, err := agg.New(Width, Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	r.SetResolution(DPI)
	core.DrawFigure(fig, r)
	return r.GetImage()
}
