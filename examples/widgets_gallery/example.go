// Package widgets_gallery shows the widget and selector family in one figure.
package widgets_gallery

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

const (
	Width  = 1100
	Height = 760
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	return plot()
}

func plot(opts ...style.Option) *core.Figure {
	fig := core.NewFigure(Width, Height, opts...)
	fig.ConstrainedLayout()

	outer := fig.GridSpec(
		3,
		4,
		core.WithGridSpecPadding(0.06, 0.96, 0.08, 0.92),
		core.WithGridSpecSpacing(0.10, 0.22),
		core.WithGridSpecHeightRatios(6, 1.2, 1.6),
		core.WithGridSpecWidthRatios(1.1, 1.5, 1.1, 1.1),
	)

	mainAx := outer.Span(0, 0, 1, 3).AddAxes()
	mainAx.SetTitle("Widgets and selectors")
	mainAx.SetXLabel("phase")
	mainAx.SetYLabel("amplitude")
	mainAx.SetXLim(0, 2*math.Pi)
	mainAx.SetYLim(-1.35, 1.35)
	mainAx.AddYGrid()

	x := make([]float64, 220)
	signal := make([]float64, len(x))
	modulation := make([]float64, len(x))
	for i := range x {
		x[i] = 2 * math.Pi * float64(i) / float64(len(x)-1)
		signal[i] = math.Sin(x[i])
		modulation[i] = 0.55 * math.Cos(1.7*x[i])
	}
	blue := render.Color{R: 0.12, G: 0.34, B: 0.68, A: 1}
	orange := render.Color{R: 0.84, G: 0.35, B: 0.18, A: 1}
	lwSignal := 2.2
	lwMod := 1.7
	mainAx.Plot(x, signal, core.PlotOptions{Color: &blue, LineWidth: &lwSignal, Label: "signal"})
	mainAx.Plot(x, modulation, core.PlotOptions{Color: &orange, LineWidth: &lwMod, Label: "modulation"})
	mainAx.AddLegend()

	span := mainAx.SpanSelector("horizontal")
	span.SetSpan(0.7, 1.35)
	rect := mainAx.RectangleSelector()
	rect.SetBounds(geom.Pt{X: 2.25, Y: -0.95}, geom.Pt{X: 3.05, Y: -0.25})
	ellipse := mainAx.EllipseSelector()
	ellipse.SetBounds(geom.Pt{X: 3.55, Y: 0.25}, geom.Pt{X: 4.45, Y: 1.00})
	polygon := mainAx.PolygonSelector()
	polygon.AppendPoint(geom.Pt{X: 4.95, Y: -0.9})
	polygon.AppendPoint(geom.Pt{X: 5.75, Y: -0.75})
	polygon.AppendPoint(geom.Pt{X: 5.45, Y: -0.15})
	polygon.Close()
	lasso := mainAx.LassoSelector()
	lasso.Begin(geom.Pt{X: 1.2, Y: 0.75})
	for _, p := range []geom.Pt{
		{X: 1.45, Y: 1.00},
		{X: 1.85, Y: 0.92},
		{X: 2.05, Y: 0.63},
		{X: 1.75, Y: 0.42},
		{X: 1.34, Y: 0.50},
	} {
		lasso.AddPoint(p)
	}
	lasso.Finish()
	cursor := mainAx.Cursor()
	cursor.SetData(2.8, 0.35)

	auxAx := outer.Cell(0, 3).AddAxes(core.WithSharedX(mainAx))
	auxAx.SetTitle("shared cursor")
	auxAx.SetXLim(0, 2*math.Pi)
	auxAx.SetYLim(-1, 1)
	auxAx.Plot(x, modulation, core.PlotOptions{Color: &orange, LineWidth: &lwMod})
	auxAx.AddYGrid()
	multi := mainAx.MultiCursor(auxAx)
	multi.SetFigurePoint(geom.Pt{X: Width * 0.72, Y: Height * 0.32})

	pressed := true
	outer.Cell(1, 0).AddWidgetAxes().Button("Apply", core.ButtonOptions{Pressed: &pressed})
	outer.Cell(1, 1).AddWidgetAxes().Slider("gain", 0, 1, 0.68)
	outer.Cell(1, 2).AddWidgetAxes().RangeSlider("window", 0, 1, 0.22, 0.78)
	active := true
	outer.Cell(1, 3).AddWidgetAxes().TextBox("label", "phase scan", core.TextBoxOptions{Active: &active})
	outer.Span(2, 0, 1, 2).AddWidgetAxes().CheckButtons([]string{"signal", "modulation", "grid"}, []bool{true, true, false})
	outer.Span(2, 2, 1, 2).AddWidgetAxes().RadioButtons([]string{"blue", "amber", "mono"}, 1)

	return fig
}

// Render is the AGG-rendered showcase image.
func Render() image.Image {
	return RenderWithOptions()
}

// RenderWithOptions is the AGG-rendered showcase image with extra style options.
func RenderWithOptions(opts ...style.Option) image.Image {
	fig := plot(opts...)
	r, err := agg.New(Width, Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}
